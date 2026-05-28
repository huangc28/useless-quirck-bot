package config

import (
	"testing"
	"time"
)

func TestLoadFromEnvDefaults(t *testing.T) {
	clearConfigEnv(t)

	cfg := LoadFromEnv()

	if cfg.ServerAddr != ":8080" {
		t.Fatalf("ServerAddr = %q", cfg.ServerAddr)
	}
	if cfg.LLMMode != "mock" {
		t.Fatalf("LLMMode = %q", cfg.LLMMode)
	}
	if cfg.LookupWorkerMode != "mock" {
		t.Fatalf("LookupWorkerMode = %q", cfg.LookupWorkerMode)
	}
	if cfg.CachePath != "data/cache.db" {
		t.Fatalf("CachePath = %q", cfg.CachePath)
	}
	if cfg.RouterConfidenceThreshold != 0.65 {
		t.Fatalf("RouterConfidenceThreshold = %f", cfg.RouterConfidenceThreshold)
	}
	if cfg.PublicLookupTimeout != 20*time.Second {
		t.Fatalf("PublicLookupTimeout = %s", cfg.PublicLookupTimeout)
	}
	if cfg.BrowserLookupTimeout != 60*time.Second {
		t.Fatalf("BrowserLookupTimeout = %s", cfg.BrowserLookupTimeout)
	}
}

func TestLoadFromEnvOverrides(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("SERVER_ADDR", ":9090")
	t.Setenv("PUBLIC_BASE_URL", "https://bot.example.test")
	t.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "secret")
	t.Setenv("LLM_MODE", "live")
	t.Setenv("OPENAI_API_KEY", "test-key")
	t.Setenv("OPENAI_MODEL", "test-model")
	t.Setenv("LOOKUP_WORKER_MODE", "live")
	t.Setenv("LOOKUP_WORKER_CMD", "codex exec")
	t.Setenv("CACHE_PATH", "data/test.db")
	t.Setenv("ROUTER_CONFIDENCE_THRESHOLD", "0.8")
	t.Setenv("PUBLIC_LOOKUP_TIMEOUT", "25s")
	t.Setenv("BROWSER_LOOKUP_TIMEOUT", "75s")

	cfg := LoadFromEnv()

	if cfg.ServerAddr != ":9090" {
		t.Fatalf("ServerAddr = %q", cfg.ServerAddr)
	}
	if cfg.PublicBaseURL != "https://bot.example.test" {
		t.Fatalf("PublicBaseURL = %q", cfg.PublicBaseURL)
	}
	if cfg.TelegramBotToken != "test-token" {
		t.Fatalf("TelegramBotToken = %q", cfg.TelegramBotToken)
	}
	if cfg.TelegramWebhookSecret != "secret" {
		t.Fatalf("TelegramWebhookSecret = %q", cfg.TelegramWebhookSecret)
	}
	if cfg.LLMMode != "live" {
		t.Fatalf("LLMMode = %q", cfg.LLMMode)
	}
	if cfg.OpenAIAPIKey != "test-key" {
		t.Fatalf("OpenAIAPIKey = %q", cfg.OpenAIAPIKey)
	}
	if cfg.OpenAIModel != "test-model" {
		t.Fatalf("OpenAIModel = %q", cfg.OpenAIModel)
	}
	if cfg.LookupWorkerMode != "live" {
		t.Fatalf("LookupWorkerMode = %q", cfg.LookupWorkerMode)
	}
	if cfg.LookupWorkerCommand != "codex exec" {
		t.Fatalf("LookupWorkerCommand = %q", cfg.LookupWorkerCommand)
	}
	if cfg.CachePath != "data/test.db" {
		t.Fatalf("CachePath = %q", cfg.CachePath)
	}
	if cfg.RouterConfidenceThreshold != 0.8 {
		t.Fatalf("RouterConfidenceThreshold = %f", cfg.RouterConfidenceThreshold)
	}
	if cfg.PublicLookupTimeout != 25*time.Second {
		t.Fatalf("PublicLookupTimeout = %s", cfg.PublicLookupTimeout)
	}
	if cfg.BrowserLookupTimeout != 75*time.Second {
		t.Fatalf("BrowserLookupTimeout = %s", cfg.BrowserLookupTimeout)
	}
}

func TestLoadFromEnvInvalidDurationsFallback(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("PUBLIC_LOOKUP_TIMEOUT", "bad")
	t.Setenv("BROWSER_LOOKUP_TIMEOUT", "also-bad")

	cfg := LoadFromEnv()

	if cfg.PublicLookupTimeout != 20*time.Second {
		t.Fatalf("PublicLookupTimeout = %s", cfg.PublicLookupTimeout)
	}
	if cfg.BrowserLookupTimeout != 60*time.Second {
		t.Fatalf("BrowserLookupTimeout = %s", cfg.BrowserLookupTimeout)
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"SERVER_ADDR",
		"PUBLIC_BASE_URL",
		"TELEGRAM_BOT_TOKEN",
		"TELEGRAM_WEBHOOK_SECRET",
		"LLM_MODE",
		"OPENAI_API_KEY",
		"OPENAI_MODEL",
		"LOOKUP_WORKER_MODE",
		"LOOKUP_WORKER_CMD",
		"CACHE_PATH",
		"ROUTER_CONFIDENCE_THRESHOLD",
		"PUBLIC_LOOKUP_TIMEOUT",
		"BROWSER_LOOKUP_TIMEOUT",
	} {
		t.Setenv(key, "")
	}
}
