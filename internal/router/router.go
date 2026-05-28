package router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"interview-chatbot/internal/contracts"
)

const DecisionExecutable = "executable"
const DecisionClarify = "clarify"
const DecisionFallback = "fallback"

const routerSystemPrompt = `Classify the user's message into exactly one mode: general_chat, public_lookup, or browser_lookup.
Return only JSON matching the router schema fields: mode, confidence, reason, normalized_question, optional cache_hint, optional force_refresh.
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
	if result.Confidence < threshold {
		return result, RouteDecision{
			Kind:    DecisionClarify,
			Message: "我不太確定你想查什麼資料，可以再補充網站或商品名稱嗎？",
		}
	}
	if result.NormalizedQuestion == "" {
		result.NormalizedQuestion = NormalizeQuestion(result.Reason)
	}
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
	apiKey     string
	model      string
	endpoint   string
	httpClient *http.Client
	threshold  float64
}

func NewLiveClient(apiKey, model, endpoint string, httpClient *http.Client, threshold float64) *LiveClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &LiveClient{
		apiKey:     apiKey,
		model:      model,
		endpoint:   endpoint,
		httpClient: httpClient,
		threshold:  thresholdOrDefault(threshold),
	}
}

func (c *LiveClient) Route(ctx context.Context, message string) (contracts.RouterResult, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return contracts.RouterResult{}, errors.New("missing API key")
	}
	if strings.TrimSpace(c.model) == "" {
		return contracts.RouterResult{}, errors.New("missing model")
	}
	endpoint := c.endpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}

	body := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": routerSystemPrompt},
			{"role": "user", "content": message},
		},
		"temperature": 0,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return contracts.RouterResult{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return contracts.RouterResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return contracts.RouterResult{}, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return contracts.RouterResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return contracts.RouterResult{}, fmt.Errorf("router request failed: %s", resp.Status)
	}

	result, err := parseRouterResult(data)
	if err != nil {
		return contracts.RouterResult{}, err
	}
	if result.NormalizedQuestion == "" {
		result.NormalizedQuestion = NormalizeQuestion(message)
	}
	validated, decision := ValidateResult(result, c.threshold)
	if decision.Kind != DecisionExecutable {
		return validated, errors.New(decision.Message)
	}
	return validated, nil
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
