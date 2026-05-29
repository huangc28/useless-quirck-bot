package config

import (
	"os"
	"strconv"
	"time"
)

const DefaultServerAddr = ":8080"
const DefaultLLMMode = "mock"
const DefaultCodexRouterCommand = "codex exec --sandbox read-only --output-schema schemas/router_result.schema.json -"
const DefaultCodexChatCommand = "codex exec --sandbox read-only -"
const DefaultLookupWorkerMode = "mock"
const DefaultCachePath = "data/cache.db"
const DefaultRouterConfidenceThreshold = 0.65

var DefaultCodexTimeout = 60 * time.Second
var DefaultPublicLookupTimeout = 130 * time.Second
var DefaultBrowserLookupTimeout = 130 * time.Second

type Config struct {
	ServerAddr                string
	PublicBaseURL             string
	TelegramBotToken          string
	TelegramWebhookSecret     string
	LLMMode                   string
	CodexRouterCommand        string
	CodexChatCommand          string
	CodexTimeout              time.Duration
	LookupWorkerMode          string
	LookupWorkerCommand       string
	CachePath                 string
	RouterConfidenceThreshold float64
	PublicLookupTimeout       time.Duration
	BrowserLookupTimeout      time.Duration
}

func LoadFromEnv() Config {
	return Config{
		ServerAddr:                env("SERVER_ADDR", DefaultServerAddr),
		PublicBaseURL:             os.Getenv("PUBLIC_BASE_URL"),
		TelegramBotToken:          os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramWebhookSecret:     os.Getenv("TELEGRAM_WEBHOOK_SECRET"),
		LLMMode:                   env("LLM_MODE", DefaultLLMMode),
		CodexRouterCommand:        env("CODEX_ROUTER_CMD", DefaultCodexRouterCommand),
		CodexChatCommand:          env("CODEX_CHAT_CMD", DefaultCodexChatCommand),
		CodexTimeout:              envDuration("CODEX_TIMEOUT", DefaultCodexTimeout),
		LookupWorkerMode:          env("LOOKUP_WORKER_MODE", DefaultLookupWorkerMode),
		LookupWorkerCommand:       os.Getenv("LOOKUP_WORKER_CMD"),
		CachePath:                 env("CACHE_PATH", DefaultCachePath),
		RouterConfidenceThreshold: envFloat("ROUTER_CONFIDENCE_THRESHOLD", DefaultRouterConfidenceThreshold),
		PublicLookupTimeout:       envDuration("PUBLIC_LOOKUP_TIMEOUT", DefaultPublicLookupTimeout),
		BrowserLookupTimeout:      envDuration("BROWSER_LOOKUP_TIMEOUT", DefaultBrowserLookupTimeout),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envFloat(key string, fallback float64) float64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return value
}
