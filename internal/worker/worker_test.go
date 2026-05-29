package worker

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"interview-chatbot/internal/contracts"
)

func TestCLIClientLookupOK(t *testing.T) {
	client := NewCLIClient(mockWorkerPath())
	got, err := client.Lookup(context.Background(), contracts.LookupRequest{
		Mode:               contracts.ModePublicLookup,
		Question:           "請問義美小泡芙多少錢",
		NormalizedQuestion: "請問義美小泡芙多少錢",
		CacheHint:          &contracts.CacheHint{LookupType: "price", Target: "義美小泡芙"},
	}, time.Second)
	if err != nil {
		t.Fatalf("Lookup returned error: %v", err)
	}
	if got.Status != contracts.StatusOK {
		t.Fatalf("Status = %s", got.Status)
	}
	if len(got.Evidence) == 0 {
		t.Fatal("expected evidence")
	}
}

func TestCLIClientLookupAuthRequired(t *testing.T) {
	client := NewCLIClient(mockWorkerPath())
	got, err := client.Lookup(context.Background(), contracts.LookupRequest{
		Mode:               contracts.ModeBrowserLookup,
		Question:           "auth_required momo",
		NormalizedQuestion: "auth_required momo",
	}, time.Second)
	if err != nil {
		t.Fatalf("Lookup returned error: %v", err)
	}
	if got.Status != contracts.StatusAuthRequired {
		t.Fatalf("Status = %s", got.Status)
	}
}

func TestCLIClientInvalidJSON(t *testing.T) {
	script := writeScript(t, "bad-json.sh", "#!/bin/sh\necho not-json\n")
	client := NewCLIClient(script)
	_, err := client.Lookup(context.Background(), contracts.LookupRequest{
		Mode:               contracts.ModePublicLookup,
		Question:           "test",
		NormalizedQuestion: "test",
	}, time.Second)
	if err == nil || !strings.Contains(err.Error(), "invalid lookup worker JSON") {
		t.Fatalf("expected invalid JSON error, got %v", err)
	}
}

func TestCLIClientTimeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell sleep script is Unix-specific")
	}
	script := writeScript(t, "sleep.sh", "#!/bin/sh\nsleep 2\necho '{\"status\":\"ok\"}'\n")
	client := NewCLIClient(script)
	_, err := client.Lookup(context.Background(), contracts.LookupRequest{
		Mode:               contracts.ModePublicLookup,
		Question:           "test",
		NormalizedQuestion: "test",
	}, 10*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func writeScript(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func mockWorkerPath() string {
	return filepath.Join("..", "..", "scripts", "mock_lookup_worker.sh")
}
