package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"interview-chatbot/internal/app"
	"interview-chatbot/internal/chat"
	"interview-chatbot/internal/config"
	"interview-chatbot/internal/contracts"
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
	if calls := sender.callCount(); calls != 0 {
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
	chatID, _ := sender.waitCall(t)
	if calls := sender.callCount(); calls != 1 {
		t.Fatalf("sender calls = %d", calls)
	}
	if chatID != 123 {
		t.Fatalf("chatID = %d", chatID)
	}
}

func TestTelegramWebhookAcknowledgesBeforeMessageProcessingCompletes(t *testing.T) {
	release := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })

	bot := testBot()
	bot.Router = blockingRouter{
		release: release,
		result:  contractsRouterResult("你好"),
	}
	sender := &fakeSender{}
	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(`{"update_id":42,"message":{"chat":{"id":123},"text":"你好"}}`))
	rec := httptest.NewRecorder()
	done := make(chan struct{})

	go func() {
		telegramWebhook(bot, sender, "")(rec, req)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		releaseOnce.Do(func() { close(release) })
		<-done
		t.Fatal("webhook waited for message processing before acknowledging Telegram")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestTelegramWebhookIgnoresDuplicateUpdateID(t *testing.T) {
	bot := testBot()
	sender := &fakeSender{}
	handler := telegramWebhook(bot, sender, "")

	req1 := httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(`{"update_id":42,"message":{"chat":{"id":123},"text":"你好"}}`))
	rec1 := httptest.NewRecorder()
	handler(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first status = %d", rec1.Code)
	}
	sender.waitCall(t)

	req2 := httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(`{"update_id":42,"message":{"chat":{"id":123},"text":"你好"}}`))
	rec2 := httptest.NewRecorder()
	handler(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("second status = %d", rec2.Code)
	}
	sender.waitNoCall(t)
	if calls := sender.callCount(); calls != 1 {
		t.Fatalf("sender calls = %d", calls)
	}
}

func TestTelegramWebhookIgnoresNonTextUpdateWithoutProcessing(t *testing.T) {
	bot := testBot()
	router := &countingRouter{result: contractsRouterResult("你好")}
	bot.Router = router
	sender := &fakeSender{}
	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(`{"update_id":43,"message":{"chat":{"id":123}}}`))
	rec := httptest.NewRecorder()

	telegramWebhook(bot, sender, "")(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if router.callCount() != 0 {
		t.Fatal("non-text update should not invoke message processing")
	}
	if calls := sender.callCount(); calls != 0 {
		t.Fatalf("sender calls = %d", calls)
	}
}

func TestTelegramWebhookAcknowledgesWhenBackgroundSendFails(t *testing.T) {
	bot := testBot()
	sender := &fakeSender{err: errors.New("telegram unavailable")}
	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(`{"update_id":44,"message":{"chat":{"id":123},"text":"你好"}}`))
	rec := httptest.NewRecorder()

	telegramWebhook(bot, sender, "")(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	sender.waitCall(t)
	if calls := sender.callCount(); calls != 1 {
		t.Fatalf("sender calls = %d", calls)
	}
}

func TestLoadRuntimeEnvReadsDotEnv(t *testing.T) {
	chdirTemp(t)
	unsetEnv(t, "LLM_MODE")
	if err := os.WriteFile(".env", []byte("LLM_MODE=live\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	loadRuntimeEnv()

	if got := os.Getenv("LLM_MODE"); got != "live" {
		t.Fatalf("LLM_MODE = %q", got)
	}
}

func TestLoadRuntimeEnvDoesNotOverrideExistingEnv(t *testing.T) {
	chdirTemp(t)
	t.Setenv("LLM_MODE", "mock")
	if err := os.WriteFile(".env", []byte("LLM_MODE=live\n"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	loadRuntimeEnv()

	if got := os.Getenv("LLM_MODE"); got != "mock" {
		t.Fatalf("LLM_MODE = %q", got)
	}
}

func testBot() *app.App {
	return &app.App{
		Router: router.MockClient{Threshold: 0.65},
		Chat:   chat.MockResponder{},
		Config: config.Config{RouterConfidenceThreshold: 0.65},
	}
}

type blockingRouter struct {
	release <-chan struct{}
	result  contracts.RouterResult
}

func (r blockingRouter) Route(context.Context, string) (contracts.RouterResult, error) {
	<-r.release
	return r.result, nil
}

type countingRouter struct {
	mu     sync.Mutex
	calls  int
	result contracts.RouterResult
}

func (r *countingRouter) Route(context.Context, string) (contracts.RouterResult, error) {
	r.mu.Lock()
	r.calls++
	r.mu.Unlock()
	return r.result, nil
}

func (r *countingRouter) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

func contractsRouterResult(text string) contracts.RouterResult {
	return contracts.RouterResult{
		Mode:               contracts.ModeGeneralChat,
		Confidence:         0.9,
		NormalizedQuestion: text,
	}
}

type fakeSender struct {
	mu     sync.Mutex
	calls  int
	chatID int64
	text   string
	sent   chan struct{}
	err    error
}

func (s *fakeSender) SendMessage(_ context.Context, chatID int64, text string) error {
	ch := s.sentChan()
	s.mu.Lock()
	s.calls++
	s.chatID = chatID
	s.text = text
	s.mu.Unlock()
	select {
	case ch <- struct{}{}:
	default:
	}
	return s.err
}

func (s *fakeSender) waitCall(t *testing.T) (int64, string) {
	t.Helper()
	ch := s.sentChan()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Telegram send")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.chatID, s.text
}

func (s *fakeSender) waitNoCall(t *testing.T) {
	t.Helper()
	ch := s.sentChan()
	select {
	case <-ch:
		t.Fatal("unexpected Telegram send")
	case <-time.After(100 * time.Millisecond):
	}
}

func (s *fakeSender) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func (s *fakeSender) sentChan() chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sent == nil {
		s.sent = make(chan struct{}, 10)
	}
	return s.sent
}

func chdirTemp(t *testing.T) {
	t.Helper()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWD); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	oldValue, ok := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
	t.Cleanup(func() {
		if ok {
			if err := os.Setenv(key, oldValue); err != nil {
				t.Fatalf("restore %s: %v", key, err)
			}
			return
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("clear %s: %v", key, err)
		}
	})
}
