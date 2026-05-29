package main

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"interview-chatbot/internal/liveworker"
)

const defaultLiveLookupCodexCommand = "codex exec --output-schema schemas/lookup_response.schema.json -"
const defaultPublicLookupPrompt = "prompts/public_lookup.md"
const defaultBrowserLookupPrompt = "prompts/browser_lookup.md"

var defaultLiveLookupTimeout = 120 * time.Second

func main() {
	if err := runLookupWorker(context.Background(), os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func runLookupWorker(ctx context.Context, input io.Reader, output io.Writer) error {
	return liveworker.Run(ctx, input, output, liveworker.Config{
		CodexCommand:      env("LIVE_LOOKUP_CODEX_CMD", defaultLiveLookupCodexCommand),
		PublicPromptPath:  env("LIVE_LOOKUP_PUBLIC_PROMPT", defaultPublicLookupPrompt),
		BrowserPromptPath: env("LIVE_LOOKUP_BROWSER_PROMPT", defaultBrowserLookupPrompt),
		Timeout:           envDuration("LIVE_LOOKUP_TIMEOUT", defaultLiveLookupTimeout),
	})
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
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
