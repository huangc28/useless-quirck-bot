package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"interview-chatbot/internal/cache"
	"interview-chatbot/internal/chat"
	"interview-chatbot/internal/config"
	"interview-chatbot/internal/contracts"
)

func TestHandleMessageGeneralChat(t *testing.T) {
	general_chat := "general_chat"
	_ = general_chat
	app := testApp()
	app.Router = fakeRouter{result: contracts.RouterResult{Mode: contracts.ModeGeneralChat, Confidence: 0.9, NormalizedQuestion: "你好"}}
	responder := &fakeChat{reply: "你好"}
	app.Chat = responder

	reply, err := app.HandleMessage(context.Background(), "你好")
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if reply != "你好" {
		t.Fatalf("reply = %q", reply)
	}
	if responder.calls != 1 {
		t.Fatalf("chat calls = %d", responder.calls)
	}
	if app.Worker.(*fakeWorker).calls != 0 {
		t.Fatal("general_chat should not call worker")
	}
	if app.Cache.(*fakeCache).gets != 0 {
		t.Fatal("general_chat should not call cache")
	}
}

func TestHandleMessageFreshWorkerOKIncludesEvidenceAndObservedAt(t *testing.T) {
	app := testApp()
	app.Router = fakeRouter{result: priceResult()}
	app.Worker = &fakeWorker{response: okResponse()}

	reply, err := app.HandleMessage(context.Background(), "請問義美小泡芙多少錢")
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	for _, want := range []string{"約 NT$59", "mock-public", "https://example.test/puff", "observed_at", "2026-05-28T12:00:00Z"} {
		if !strings.Contains(reply, want) {
			t.Fatalf("reply missing %q: %s", want, reply)
		}
	}
	if app.Cache.(*fakeCache).puts != 1 {
		t.Fatalf("cache puts = %d", app.Cache.(*fakeCache).puts)
	}
}

func TestHandleMessageCacheHitSkipsWorker(t *testing.T) {
	app := testApp()
	app.Router = fakeRouter{result: priceResult()}
	app.Cache = &fakeCache{entry: cache.Entry{State: cache.StateFresh, Response: okResponse()}}

	reply, err := app.HandleMessage(context.Background(), "請問義美小泡芙多少錢")
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if !strings.Contains(reply, "快取結果") {
		t.Fatalf("reply = %q", reply)
	}
	if app.Worker.(*fakeWorker).calls != 0 {
		t.Fatal("cache hit should skip worker")
	}
}

func TestHandleMessageForceRefreshBypassesCache(t *testing.T) {
	app := testApp()
	app.Router = fakeRouter{result: priceResult()}
	app.Cache = &fakeCache{entry: cache.Entry{State: cache.StateFresh, Response: okResponse()}}
	app.Worker = &fakeWorker{response: okResponse()}

	_, err := app.HandleMessage(context.Background(), "不要快取，請問義美小泡芙多少錢")
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if app.Cache.(*fakeCache).gets != 0 {
		t.Fatal("force refresh should skip cache get")
	}
	if app.Worker.(*fakeWorker).calls != 1 {
		t.Fatal("force refresh should call worker")
	}
}

func TestHandleMessageAuthRequired(t *testing.T) {
	auth_required := "auth_required"
	_ = auth_required
	app := testApp()
	app.Router = fakeRouter{result: browserResult()}
	app.Worker = &fakeWorker{response: contracts.LookupResponse{Status: contracts.StatusAuthRequired, Error: "browser session required"}}

	reply, err := app.HandleMessage(context.Background(), "請問 momo 義美小泡芙多少錢")
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if !strings.Contains(reply, "browser profile") || !strings.Contains(reply, "do not send third-party credentials") {
		t.Fatalf("unsafe auth reply: %s", reply)
	}
}

func TestHandleMessageCachedAuthRequired(t *testing.T) {
	app := testApp()
	app.Router = fakeRouter{result: browserResult()}
	app.Cache = &fakeCache{entry: cache.Entry{State: cache.StateFresh, Response: contracts.LookupResponse{Status: contracts.StatusAuthRequired}}}

	reply, err := app.HandleMessage(context.Background(), "請問 momo 義美小泡芙多少錢")
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if !strings.Contains(reply, "browser profile") {
		t.Fatalf("reply = %q", reply)
	}
	if app.Worker.(*fakeWorker).calls != 0 {
		t.Fatal("fresh auth_required cache should skip worker")
	}
}

func TestHandleMessageStaleAuthRequiredIfWorkerError(t *testing.T) {
	app := testApp()
	app.Router = fakeRouter{result: browserResult()}
	app.Cache = &fakeCache{entry: cache.Entry{State: cache.StateStale, Response: contracts.LookupResponse{Status: contracts.StatusAuthRequired}}}
	app.Worker = &fakeWorker{err: errors.New("worker_error")}

	reply, err := app.HandleMessage(context.Background(), "請問 momo 義美小泡芙多少錢")
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if !strings.Contains(reply, "browser profile") {
		t.Fatalf("reply = %q", reply)
	}
}

func TestHandleMessageStaleIfWorkerError(t *testing.T) {
	stale := "stale"
	_ = stale
	app := testApp()
	app.Router = fakeRouter{result: priceResult()}
	app.Cache = &fakeCache{entry: cache.Entry{State: cache.StateStale, Response: okResponse()}}
	app.Worker = &fakeWorker{err: errors.New("worker_error")}

	reply, err := app.HandleMessage(context.Background(), "請問義美小泡芙多少錢")
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if !strings.Contains(reply, "快取已過期") {
		t.Fatalf("reply = %q", reply)
	}
}

func TestHandleMessageWorkerErrorNoStaleIsGraceful(t *testing.T) {
	worker_error := "worker_error"
	_ = worker_error
	app := testApp()
	app.Router = fakeRouter{result: priceResult()}
	app.Worker = &fakeWorker{response: contracts.LookupResponse{Status: contracts.StatusError, Error: "worker_error"}}

	reply, err := app.HandleMessage(context.Background(), "請問義美小泡芙多少錢")
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if !strings.Contains(reply, "查詢暫時失敗") || !strings.Contains(reply, "沒有使用 mock") {
		t.Fatalf("reply = %q", reply)
	}
	if app.Cache.(*fakeCache).puts != 0 {
		t.Fatal("worker error should not be cached")
	}
}

func TestHandleMessageInvalidJSONWorkerError(t *testing.T) {
	invalidJSON := "invalid JSON"
	_ = invalidJSON
	app := testApp()
	app.Router = fakeRouter{result: priceResult()}
	app.Worker = &fakeWorker{err: errors.New("invalid JSON")}

	reply, err := app.HandleMessage(context.Background(), "請問義美小泡芙多少錢")
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if !strings.Contains(reply, "查詢暫時失敗") {
		t.Fatalf("reply = %q", reply)
	}
}

func TestHandleMessageLowConfidenceClarifies(t *testing.T) {
	app := testApp()
	app.Router = fakeRouter{result: contracts.RouterResult{Mode: contracts.ModePublicLookup, Confidence: 0.2, NormalizedQuestion: "泡芙"}}

	reply, err := app.HandleMessage(context.Background(), "泡芙")
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if !strings.Contains(reply, "不太確定") {
		t.Fatalf("reply = %q", reply)
	}
}

func testApp() *App {
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	return &App{
		Router: fakeRouter{result: priceResult()},
		Chat:   chat.MockResponder{},
		Cache:  &fakeCache{},
		Worker: &fakeWorker{response: okResponse()},
		Clock:  func() time.Time { return now },
		Config: config.Config{
			RouterConfidenceThreshold: 0.65,
			PublicLookupTimeout:       20 * time.Second,
			BrowserLookupTimeout:      60 * time.Second,
		},
	}
}

type fakeRouter struct {
	result contracts.RouterResult
	err    error
}

func (r fakeRouter) Route(context.Context, string) (contracts.RouterResult, error) {
	return r.result, r.err
}

type fakeChat struct {
	reply string
	calls int
}

func (c *fakeChat) Respond(context.Context, string) (string, error) {
	c.calls++
	return c.reply, nil
}

type fakeCache struct {
	entry cache.Entry
	gets  int
	puts  int
}

func (c *fakeCache) Get(context.Context, string, time.Time) (cache.Entry, error) {
	c.gets++
	if c.entry.State == "" {
		return cache.Entry{State: cache.StateMiss}, nil
	}
	return c.entry, nil
}

func (c *fakeCache) Put(context.Context, string, contracts.RouterResult, contracts.LookupResponse, time.Duration, time.Time) error {
	c.puts++
	return nil
}

type fakeWorker struct {
	response contracts.LookupResponse
	err      error
	calls    int
}

func (w *fakeWorker) Lookup(context.Context, contracts.LookupRequest, time.Duration) (contracts.LookupResponse, error) {
	w.calls++
	return w.response, w.err
}

func priceResult() contracts.RouterResult {
	return contracts.RouterResult{
		Mode:               contracts.ModePublicLookup,
		Confidence:         0.9,
		NormalizedQuestion: "請問義美小泡芙多少錢",
		CacheHint:          &contracts.CacheHint{LookupType: "price", Target: "義美小泡芙"},
	}
}

func browserResult() contracts.RouterResult {
	return contracts.RouterResult{
		Mode:               contracts.ModeBrowserLookup,
		Confidence:         0.9,
		NormalizedQuestion: "請問 momo 義美小泡芙多少錢",
		CacheHint:          &contracts.CacheHint{LookupType: "price", Target: "義美小泡芙", Site: "momo"},
	}
}

func okResponse() contracts.LookupResponse {
	return contracts.LookupResponse{
		Status:     contracts.StatusOK,
		Answer:     "約 NT$59",
		ObservedAt: "observed_at: 2026-05-28T12:00:00Z",
		Evidence: []contracts.Evidence{{
			Title:      "商品頁",
			URL:        "https://example.test/puff",
			Source:     "mock-public",
			ObservedAt: "2026-05-28T12:00:00Z",
		}},
	}
}
