package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"interview-chatbot/internal/app"
	"interview-chatbot/internal/chat"
	"interview-chatbot/internal/config"
	"interview-chatbot/internal/router"
)

func TestTelegramWebhookRejectsBadSecret(t *testing.T) {
	bot := testBot()
	sender := &fakeSender{}
	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(`{"message":{"chat":{"id":123},"text":"你好"}}`))
	rec := httptest.NewRecorder()

	telegramWebhook(bot, sender, "secret")(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rec.Code)
	}
	if sender.calls != 0 {
		t.Fatal("unauthorized request should not send Telegram message")
	}
}

func TestTelegramWebhookAcceptsMatchingSecret(t *testing.T) {
	bot := testBot()
	sender := &fakeSender{}
	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(`{"message":{"chat":{"id":123},"text":"你好"}}`))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "secret")
	rec := httptest.NewRecorder()

	telegramWebhook(bot, sender, "secret")(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if sender.calls != 1 {
		t.Fatalf("sender calls = %d", sender.calls)
	}
	if sender.chatID != 123 {
		t.Fatalf("chatID = %d", sender.chatID)
	}
}

func testBot() *app.App {
	return &app.App{
		Router: router.MockClient{Threshold: 0.65},
		Chat:   chat.MockResponder{},
		Config: config.Config{RouterConfidenceThreshold: 0.65},
	}
}

type fakeSender struct {
	calls  int
	chatID int64
	text   string
}

func (s *fakeSender) SendMessage(_ context.Context, chatID int64, text string) error {
	s.calls++
	s.chatID = chatID
	s.text = text
	return nil
}
