package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"interview-chatbot/internal/contracts"
)

func TestRunLookupWorkerUsesEnvironmentConfig(t *testing.T) {
	dir := t.TempDir()
	publicPrompt := writeFile(t, dir, "public.md", "PUBLIC PROMPT")
	browserPrompt := writeFile(t, dir, "browser.md", "BROWSER PROMPT")
	capturePath := filepath.Join(dir, "codex-input.txt")
	codex := writeScript(t, dir, "codex.sh", `#!/bin/sh
cat > "$1"
printf '{"status":"ok","answer":"ok","observed_at":"2026-05-29T12:00:00Z","evidence":[{"source":"example","url":"https://example.test","snippet":"page context","observed_at":"2026-05-29T12:00:00Z"}]}'`)

	t.Setenv("LIVE_LOOKUP_CODEX_CMD", codex+" "+capturePath)
	t.Setenv("LIVE_LOOKUP_PUBLIC_PROMPT", publicPrompt)
	t.Setenv("LIVE_LOOKUP_BROWSER_PROMPT", browserPrompt)
	t.Setenv("LIVE_LOOKUP_TIMEOUT", "5s")

	var output bytes.Buffer
	err := runLookupWorker(context.Background(), strings.NewReader(`{"mode":"public_lookup","question":"q","normalized_question":"q"}`), &output)
	if err != nil {
		t.Fatalf("runLookupWorker returned error: %v", err)
	}

	var response contracts.LookupResponse
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("decode output: %v: %s", err, output.String())
	}
	if response.Status != contracts.StatusOK {
		t.Fatalf("Status = %s", response.Status)
	}

	captured, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatalf("read capture: %v", err)
	}
	if !strings.Contains(string(captured), "PUBLIC PROMPT") {
		t.Fatalf("codex input missing public prompt: %s", captured)
	}
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func writeScript(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}
