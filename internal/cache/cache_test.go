package cache

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"interview-chatbot/internal/contracts"
)

func TestKeyFromRouterUsesCacheHint(t *testing.T) {
	result := contracts.RouterResult{
		Mode:               contracts.ModePublicLookup,
		NormalizedQuestion: "泡芙多少錢",
		CacheHint:          &contracts.CacheHint{LookupType: "price", Target: "義美小泡芙"},
	}

	got := KeyFromRouter(result)
	want := "public_lookup:price:義美小泡芙"
	if got != want {
		t.Fatalf("KeyFromRouter = %q, want %q", got, want)
	}
}

func TestStoreFreshStaleMiss(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t)
	defer store.Close()
	now := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	result := priceRouterResult()
	response := contracts.LookupResponse{
		Status:     contracts.StatusOK,
		Answer:     "約 NT$59",
		ObservedAt: now.Format(time.RFC3339),
		Evidence:   []contracts.Evidence{{Title: "Store", URL: "https://example.test", Source: "example"}},
	}
	key := KeyFromRouter(result)

	miss, err := store.Get(ctx, key, now)
	if err != nil {
		t.Fatalf("Get miss: %v", err)
	}
	if miss.State != StateMiss {
		t.Fatalf("State = %s, want miss", miss.State)
	}

	if err := store.Put(ctx, key, result, response, 30*time.Minute, now); err != nil {
		t.Fatalf("Put: %v", err)
	}
	fresh, err := store.Get(ctx, key, now.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("Get fresh: %v", err)
	}
	if fresh.State != StateFresh {
		t.Fatalf("State = %s, want fresh", fresh.State)
	}
	if fresh.Response.Answer != "約 NT$59" {
		t.Fatalf("Answer = %q", fresh.Response.Answer)
	}

	stale, err := store.Get(ctx, key, now.Add(31*time.Minute))
	if err != nil {
		t.Fatalf("Get stale: %v", err)
	}
	if stale.State != StateStale {
		t.Fatalf("State = %s, want stale", stale.State)
	}
}

func TestTTLFor(t *testing.T) {
	result := priceRouterResult()
	if got := TTLFor(result, contracts.LookupResponse{Status: contracts.StatusOK}); got != 30*time.Minute {
		t.Fatalf("price ttl = %s", got)
	}
	if got := TTLFor(result, contracts.LookupResponse{Status: contracts.StatusNoResult}); got != 5*time.Minute {
		t.Fatalf("no_result ttl = %s", got)
	}
	if got := TTLFor(result, contracts.LookupResponse{Status: contracts.StatusAuthRequired}); got != 0 {
		t.Fatalf("auth_required ttl = %s", got)
	}
	if got := TTLFor(result, contracts.LookupResponse{Status: contracts.StatusError}); got != 0 {
		t.Fatalf("error ttl = %s", got)
	}
}

func TestShouldForceRefresh(t *testing.T) {
	for _, question := range []string{"請重新查一次", "請更新", "refresh please", "不要快取"} {
		if !ShouldForceRefresh(question, contracts.RouterResult{}) {
			t.Fatalf("expected force refresh for %q", question)
		}
	}
	if !ShouldForceRefresh("正常問題", contracts.RouterResult{ForceRefresh: true}) {
		t.Fatal("expected router force refresh to win")
	}
}

func openTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(context.Background(), filepath.Join(t.TempDir(), "cache.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return store
}

func priceRouterResult() contracts.RouterResult {
	return contracts.RouterResult{
		Mode:               contracts.ModePublicLookup,
		NormalizedQuestion: "請問義美小泡芙多少錢",
		CacheHint:          &contracts.CacheHint{LookupType: "price", Target: "義美小泡芙"},
	}
}
