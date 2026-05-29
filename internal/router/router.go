package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"interview-chatbot/internal/clirunner"
	"interview-chatbot/internal/contracts"
)

const DecisionExecutable = "executable"
const DecisionClarify = "clarify"
const DecisionFallback = "fallback"

const routerSystemPrompt = `Classify the user's message into exactly one mode: general_chat, public_lookup, or browser_lookup.
Use general_chat for greetings, thanks, and ordinary conversation.
Use public_lookup when the user asks for public information but does not require a specific website or URL.
Use browser_lookup when the user explicitly names a website, platform, or URL to inspect.
Return only JSON matching the router schema fields: mode, confidence, reason, normalized_question, cache_hint, force_refresh.
Always include cache_hint. Use empty strings for cache_hint fields that do not apply. Use force_refresh false unless the user explicitly asks to bypass cache.
For price lookups, set cache_hint.lookup_type to "price" and cache_hint.target to the normalized product name when possible.
You must not answer lookup questions.`

type Client interface {
	Route(ctx context.Context, message string) (contracts.RouterResult, error)
}

type RouteDecision struct {
	Kind    string
	Message string
}

func NormalizeQuestion(message string) string {
	normalized := strings.TrimSpace(message)
	for _, prefix := range []string{"/ask", "/price", "/lookup"} {
		if strings.HasPrefix(normalized, prefix) {
			normalized = strings.TrimSpace(strings.TrimPrefix(normalized, prefix))
			break
		}
	}
	return normalized
}

func ValidateResult(result contracts.RouterResult, threshold float64) (contracts.RouterResult, RouteDecision) {
	if err := contracts.ValidateMode(result.Mode); err != nil {
		return contracts.RouterResult{}, RouteDecision{
			Kind:    DecisionFallback,
			Message: "我暫時無法判斷這個請求，請換個方式再問一次。",
		}
	}
	if result.Confidence < 0 || result.Confidence > 1 {
		return contracts.RouterResult{}, RouteDecision{
			Kind:    DecisionFallback,
			Message: "我暫時無法判斷這個請求，請換個方式再問一次。",
		}
	}
	if result.Confidence < threshold {
		return result, RouteDecision{
			Kind:    DecisionClarify,
			Message: "我不太確定你想查什麼資料，可以再補充網站或商品名稱嗎？",
		}
	}
	result.NormalizedQuestion = strings.TrimSpace(result.NormalizedQuestion)
	if result.NormalizedQuestion == "" {
		return contracts.RouterResult{}, RouteDecision{
			Kind:    DecisionFallback,
			Message: "我暫時無法處理空白訊息。",
		}
	}
	return result, RouteDecision{Kind: DecisionExecutable}
}

type MockClient struct {
	Threshold float64
}

func (c MockClient) Route(_ context.Context, message string) (contracts.RouterResult, error) {
	normalized := NormalizeQuestion(message)
	lower := strings.ToLower(normalized)
	result := contracts.RouterResult{
		Mode:               contracts.ModeGeneralChat,
		Confidence:         0.9,
		Reason:             "general chat",
		NormalizedQuestion: normalized,
	}

	switch {
	case normalized == "你好":
		result.Mode = contracts.ModeGeneralChat
		result.Reason = "greeting"
	case strings.Contains(lower, "http://") || strings.Contains(lower, "https://"):
		result.Mode = contracts.ModeBrowserLookup
		result.Reason = "url-specific lookup"
		result.CacheHint = &contracts.CacheHint{URL: firstURL(normalized)}
	case strings.Contains(lower, "momo") && strings.Contains(normalized, "義美小泡芙"):
		result.Mode = contracts.ModeBrowserLookup
		result.Reason = "site-specific price lookup"
		result.CacheHint = &contracts.CacheHint{LookupType: "price", Target: "義美小泡芙", Site: "momo"}
	case strings.Contains(normalized, "義美小泡芙") && strings.Contains(normalized, "多少錢"):
		result.Mode = contracts.ModePublicLookup
		result.Reason = "public price lookup"
		result.CacheHint = &contracts.CacheHint{LookupType: "price", Target: "義美小泡芙"}
	}

	validated, decision := ValidateResult(result, thresholdOrDefault(c.Threshold))
	if decision.Kind != DecisionExecutable {
		return validated, errors.New(decision.Message)
	}
	return validated, nil
}

type LiveClient struct {
	command string
	timeout time.Duration
}

func NewLiveClient(command string, timeout time.Duration, _ float64) *LiveClient {
	return &LiveClient{
		command: strings.TrimSpace(command),
		timeout: timeout,
	}
}

func (c *LiveClient) Route(ctx context.Context, message string) (contracts.RouterResult, error) {
	stdout, err := clirunner.Run(ctx, c.command, []byte(routerPrompt(message)), c.timeout)
	if err != nil {
		return contracts.RouterResult{}, err
	}
	result, err := parseRouterResult(stdout)
	if err != nil {
		return contracts.RouterResult{}, err
	}
	return result, nil
}

func routerPrompt(message string) string {
	return fmt.Sprintf("%s\n\nUser message:\n%s\n", routerSystemPrompt, strings.TrimSpace(message))
}

func parseRouterResult(data []byte) (contracts.RouterResult, error) {
	var direct contracts.RouterResult
	if err := json.Unmarshal(data, &direct); err == nil && direct.Mode != "" {
		return direct, nil
	}

	var chat struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &chat); err == nil && len(chat.Choices) > 0 {
		var result contracts.RouterResult
		if err := json.Unmarshal([]byte(chat.Choices[0].Message.Content), &result); err != nil {
			return contracts.RouterResult{}, err
		}
		return result, nil
	}

	var response struct {
		OutputText string `json:"output_text"`
	}
	if err := json.Unmarshal(data, &response); err == nil && response.OutputText != "" {
		var result contracts.RouterResult
		if err := json.Unmarshal([]byte(response.OutputText), &result); err != nil {
			return contracts.RouterResult{}, err
		}
		return result, nil
	}

	return contracts.RouterResult{}, errors.New("invalid router JSON")
}

func firstURL(text string) string {
	for _, field := range strings.Fields(text) {
		if strings.HasPrefix(field, "http://") || strings.HasPrefix(field, "https://") {
			if parsed, err := url.Parse(field); err == nil && parsed.Host != "" {
				return field
			}
		}
	}
	return ""
}

func thresholdOrDefault(threshold float64) float64 {
	if threshold <= 0 {
		return 0.65
	}
	return threshold
}
