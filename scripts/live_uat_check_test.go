package scripts

import (
	"os/exec"
	"strings"
	"testing"
)

func TestLiveUATCheckReportsMissingEnvironmentWithoutSecrets(t *testing.T) {
	cmd := exec.Command("sh", "live_uat_check.sh", "--check-env")
	cmd.Env = []string{"PATH=/usr/bin:/bin"}

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected missing environment to fail")
	}
	text := string(output)
	for _, want := range []string{
		"missing TELEGRAM_BOT_TOKEN",
		"missing PUBLIC_BASE_URL",
		"missing LLM_MODE=live",
		"missing LOOKUP_WORKER_MODE=live",
		"missing LOOKUP_WORKER_CMD",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "token") && strings.Contains(text, "123456") {
		t.Fatalf("output appeared to leak a secret value:\n%s", text)
	}
}

func TestLiveUATCheckPrintsDemoChecklistWhenReady(t *testing.T) {
	cmd := exec.Command("sh", "live_uat_check.sh", "--check-env")
	cmd.Env = []string{
		"PATH=/usr/bin:/bin",
		"TELEGRAM_BOT_TOKEN=123456:secret-token",
		"PUBLIC_BASE_URL=https://bot.example.test",
		"LLM_MODE=live",
		"LOOKUP_WORKER_MODE=live",
		"LOOKUP_WORKER_CMD=go run ./cmd/lookup-worker",
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected ready environment to pass: %v\n%s", err, output)
	}
	text := string(output)
	for _, want := range []string{
		"Live UAT readiness: ok",
		"POST https://bot.example.test/telegram/webhook",
		"你好",
		"請問義美小泡芙多少錢",
		"請問 momo 義美小泡芙多少錢",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "123456:secret-token") {
		t.Fatalf("output leaked token:\n%s", text)
	}
}

func TestLiveUATCheckRejectsMockLookupWorker(t *testing.T) {
	cmd := exec.Command("sh", "live_uat_check.sh", "--check-env")
	cmd.Env = []string{
		"PATH=/usr/bin:/bin",
		"TELEGRAM_BOT_TOKEN=123456:secret-token",
		"PUBLIC_BASE_URL=https://bot.example.test",
		"LLM_MODE=live",
		"LOOKUP_WORKER_MODE=live",
		"LOOKUP_WORKER_CMD=./scripts/mock_lookup_worker.sh",
	}

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected mock lookup worker to fail live UAT readiness")
	}
	text := string(output)
	for _, want := range []string{
		"LOOKUP_WORKER_CMD must use the repo-owned live worker",
		"go run ./cmd/lookup-worker",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "123456:secret-token") {
		t.Fatalf("output leaked token:\n%s", text)
	}
}
