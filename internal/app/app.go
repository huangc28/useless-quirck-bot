package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"interview-chatbot/internal/cache"
	"interview-chatbot/internal/chat"
	"interview-chatbot/internal/config"
	"interview-chatbot/internal/contracts"
	"interview-chatbot/internal/router"
	"interview-chatbot/internal/worker"
)

type CacheStore interface {
	Get(ctx context.Context, key string, now time.Time) (cache.Entry, error)
	Put(ctx context.Context, key string, result contracts.RouterResult, response contracts.LookupResponse, ttl time.Duration, now time.Time) error
}

type App struct {
	Router    router.Client
	Chat      chat.Responder
	Cache     CacheStore
	Worker    worker.Client
	Clock     func() time.Time
	Config    config.Config
	Threshold float64
}

func (a *App) HandleMessage(ctx context.Context, text string) (reply string, err error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "請輸入想詢問的內容。", nil
	}
	if a.Router == nil {
		return safeFallback(), nil
	}

	result, err := a.Router.Route(ctx, text)
	if err != nil {
		return safeFallback(), nil
	}
	result, decision := router.ValidateResult(result, a.threshold())
	switch decision.Kind {
	case router.DecisionClarify, router.DecisionFallback:
		return decision.Message, nil
	}

	if result.Mode == contracts.ModeGeneralChat {
		if a.Chat == nil {
			return safeFallback(), nil
		}
		reply, err := a.Chat.Respond(ctx, text)
		if err != nil {
			return safeFallback(), nil
		}
		return reply, nil
	}

	return a.handleLookup(ctx, text, result)
}

func (a *App) handleLookup(ctx context.Context, text string, result contracts.RouterResult) (string, error) {
	if a.Worker == nil {
		return lookupFailureReply(), nil
	}
	now := a.now()
	key := cache.KeyFromRouter(result)
	forceRefresh := cache.ShouldForceRefresh(text, result)

	var stale cache.Entry
	if a.Cache != nil && !forceRefresh {
		entry, err := a.Cache.Get(ctx, key, now)
		if err == nil {
			switch entry.State {
			case cache.StateFresh:
				if entry.Response.Status == contracts.StatusAuthRequired {
					return authRequiredReply(entry.Response), nil
				}
				return formatLookupReply(entry.Response, true, false), nil
			case cache.StateStale:
				stale = entry
			}
		}
	}

	response, err := a.Worker.Lookup(ctx, contracts.LookupRequest{
		Mode:               result.Mode,
		Question:           text,
		NormalizedQuestion: result.NormalizedQuestion,
		CacheHint:          result.CacheHint,
		ForceRefresh:       forceRefresh,
	}, a.lookupTimeout(result.Mode))
	if err != nil {
		if stale.State == cache.StateStale {
			return staleFallbackReply(stale), nil
		}
		return lookupFailureReply(), nil
	}

	if response.Status == contracts.StatusError {
		if stale.State == cache.StateStale {
			return staleFallbackReply(stale), nil
		}
		return lookupFailureReply(), nil
	}

	if response.Status == contracts.StatusAuthRequired {
		return authRequiredReply(response), nil
	}

	a.cacheResponse(ctx, key, result, response, now)
	return formatLookupReply(response, false, false), nil
}

func (a *App) cacheResponse(ctx context.Context, key string, result contracts.RouterResult, response contracts.LookupResponse, now time.Time) {
	if a.Cache == nil {
		return
	}
	ttl := cache.TTLFor(result, response)
	if ttl <= 0 {
		return
	}
	_ = a.Cache.Put(ctx, key, result, response, ttl, now)
}

func (a *App) now() time.Time {
	if a.Clock != nil {
		return a.Clock()
	}
	return time.Now().UTC()
}

func (a *App) threshold() float64 {
	if a.Threshold > 0 {
		return a.Threshold
	}
	if a.Config.RouterConfidenceThreshold > 0 {
		return a.Config.RouterConfidenceThreshold
	}
	return config.DefaultRouterConfidenceThreshold
}

func (a *App) lookupTimeout(mode string) time.Duration {
	switch mode {
	case contracts.ModeBrowserLookup:
		if a.Config.BrowserLookupTimeout > 0 {
			return a.Config.BrowserLookupTimeout
		}
		return config.DefaultBrowserLookupTimeout
	default:
		if a.Config.PublicLookupTimeout > 0 {
			return a.Config.PublicLookupTimeout
		}
		return config.DefaultPublicLookupTimeout
	}
}

func formatLookupReply(response contracts.LookupResponse, fromCache bool, stale bool) string {
	var b strings.Builder
	switch {
	case stale:
		b.WriteString("快取已過期，重新查詢暫時失敗；以下提供舊資料。\n")
	case fromCache:
		b.WriteString("快取結果：\n")
	}

	if response.Status == contracts.StatusNoResult {
		if response.Answer != "" {
			b.WriteString(response.Answer)
		} else {
			b.WriteString("找不到符合條件的資料。")
		}
	} else {
		b.WriteString(response.Answer)
	}
	if len(response.Evidence) > 0 {
		ev := response.Evidence[0]
		b.WriteString("\n來源：")
		source := firstNonEmpty(ev.Source, ev.Title)
		if source != "" {
			b.WriteString(source)
		}
		if ev.Title != "" && ev.Title != source {
			b.WriteString(" - ")
			b.WriteString(ev.Title)
		}
		if ev.URL != "" {
			b.WriteString(" ")
			b.WriteString(ev.URL)
		}
	}
	observedAt := response.ObservedAt
	if observedAt == "" && len(response.Evidence) > 0 {
		observedAt = response.Evidence[0].ObservedAt
	}
	if observedAt != "" {
		b.WriteString("\n查詢時間：")
		b.WriteString(observedAt)
	}
	return strings.TrimSpace(b.String())
}

func staleFallbackReply(entry cache.Entry) string {
	if entry.Response.Status == contracts.StatusAuthRequired {
		return authRequiredReply(entry.Response)
	}
	return formatLookupReply(entry.Response, true, true)
}

func authRequiredReply(response contracts.LookupResponse) string {
	detail := strings.TrimSpace(response.Error)
	if detail == "" {
		detail = "目標網站需要登入狀態"
	}
	return fmt.Sprintf("這個查詢需要先準備本機瀏覽器登入工作階段（%s）。請在 Chrome MCP 使用的 browser profile 完成登入後再重試；do not send third-party credentials in Telegram.", detail)
}

func lookupFailureReply() string {
	return "查詢暫時失敗，請稍後再試；沒有使用 mock 結果假裝成功。"
}

func safeFallback() string {
	return "我暫時無法處理這則訊息，請稍後再試或換個方式描述。"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
