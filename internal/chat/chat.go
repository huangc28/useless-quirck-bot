package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"interview-chatbot/internal/clirunner"
)

const chatSystemPrompt = "Reply in concise Traditional Chinese. Answer the Telegram user directly. Do not edit files, run commands, or browse."

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
	command string
	timeout time.Duration
}

func NewLiveResponder(command string, timeout time.Duration) *LiveResponder {
	return &LiveResponder{
		command: strings.TrimSpace(command),
		timeout: timeout,
	}
}

func (r *LiveResponder) Respond(ctx context.Context, question string) (string, error) {
	stdout, err := clirunner.Run(ctx, r.command, []byte(chatPrompt(question)), r.timeout)
	if err != nil {
		return "", err
	}
	reply := strings.TrimSpace(string(stdout))
	if reply == "" {
		return "", errors.New("empty chat response")
	}
	return reply, nil
}

func chatPrompt(question string) string {
	return fmt.Sprintf("%s\n\nUser message:\n%s\n", chatSystemPrompt, strings.TrimSpace(question))
}
