package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const chatSystemPrompt = "Reply in concise Traditional Chinese. You can answer general chat and mention that the bot can help search website data."

type Responder interface {
	Respond(ctx context.Context, question string) (string, error)
}

type MockResponder struct{}

func (MockResponder) Respond(_ context.Context, question string) (string, error) {
	if strings.TrimSpace(question) == "" {
		return "請輸入想詢問的內容。", nil
	}
	if strings.TrimSpace(question) == "你好" {
		return "你好，我可以協助查找網站資料。你可以試試：請問義美小泡芙多少錢", nil
	}
	return "我可以協助一般對話，也可以查找網站資料。", nil
}

type LiveResponder struct {
	apiKey     string
	model      string
	endpoint   string
	httpClient *http.Client
}

func NewLiveResponder(apiKey, model, endpoint string, httpClient *http.Client) *LiveResponder {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &LiveResponder{
		apiKey:     apiKey,
		model:      model,
		endpoint:   endpoint,
		httpClient: httpClient,
	}
}

func (r *LiveResponder) Respond(ctx context.Context, question string) (string, error) {
	if strings.TrimSpace(r.apiKey) == "" {
		return "", errors.New("missing API key")
	}
	if strings.TrimSpace(r.model) == "" {
		return "", errors.New("missing model")
	}
	endpoint := r.endpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}

	payload, err := json.Marshal(map[string]any{
		"model": r.model,
		"messages": []map[string]string{
			{"role": "system", "content": chatSystemPrompt},
			{"role": "user", "content": question},
		},
		"temperature": 0.2,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("chat request failed: %s", resp.Status)
	}
	return parseReply(data)
}

func parseReply(data []byte) (string, error) {
	var direct struct {
		Reply string `json:"reply"`
	}
	if err := json.Unmarshal(data, &direct); err == nil && strings.TrimSpace(direct.Reply) != "" {
		return strings.TrimSpace(direct.Reply), nil
	}

	var chat struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &chat); err == nil && len(chat.Choices) > 0 {
		reply := strings.TrimSpace(chat.Choices[0].Message.Content)
		if reply != "" {
			return reply, nil
		}
	}

	var response struct {
		OutputText string `json:"output_text"`
	}
	if err := json.Unmarshal(data, &response); err == nil && strings.TrimSpace(response.OutputText) != "" {
		return strings.TrimSpace(response.OutputText), nil
	}

	return "", errors.New("invalid chat response")
}
