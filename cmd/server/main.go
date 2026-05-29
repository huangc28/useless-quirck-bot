package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"interview-chatbot/internal/app"
	"interview-chatbot/internal/cache"
	"interview-chatbot/internal/chat"
	"interview-chatbot/internal/config"
	"interview-chatbot/internal/router"
	"interview-chatbot/internal/telegram"
	"interview-chatbot/internal/worker"
)

func main() {
	cfg := config.LoadFromEnv()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cacheStore, err := cache.Open(ctx, cfg.CachePath)
	if err != nil {
		log.Fatalf("initialize cache: %v", err)
	}
	defer cacheStore.Close()

	lookupCommand := cfg.LookupWorkerCommand
	if cfg.LookupWorkerMode == "mock" && strings.TrimSpace(lookupCommand) == "" {
		lookupCommand = "./scripts/mock_lookup_worker.sh"
	}

	bot := &app.App{
		Router: routerClient(cfg),
		Chat:   chatResponder(cfg),
		Cache:  cacheStore,
		Worker: worker.NewCLIClient(lookupCommand),
		Config: cfg,
	}
	sender := telegram.NewHTTPSender(cfg.TelegramBotToken, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /telegram/webhook", telegramWebhook(bot, sender, cfg.TelegramWebhookSecret))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	server := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("server listening on %s with LLM_MODE=%s LOOKUP_WORKER_MODE=%s", cfg.ServerAddr, cfg.LLMMode, cfg.LookupWorkerMode)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}

func telegramWebhook(bot *app.App, sender telegram.Sender, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if secret != "" && r.Header.Get("X-Telegram-Bot-Api-Secret-Token") != secret {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		chatID, text, ok, err := telegram.ParseUpdate(r.Body)
		if err != nil {
			http.Error(w, "invalid telegram update", http.StatusBadRequest)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		reply, err := bot.HandleMessage(r.Context(), text)
		if err != nil {
			http.Error(w, "message handling failed", http.StatusInternalServerError)
			return
		}
		if err := sender.SendMessage(r.Context(), chatID, reply); err != nil {
			http.Error(w, "telegram send failed", http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	}
}

func routerClient(cfg config.Config) router.Client {
	if cfg.LLMMode == "live" {
		return router.NewLiveClient(cfg.OpenAIAPIKey, cfg.OpenAIModel, "", nil, cfg.RouterConfidenceThreshold)
	}
	return router.MockClient{Threshold: cfg.RouterConfidenceThreshold}
}

func chatResponder(cfg config.Config) chat.Responder {
	if cfg.LLMMode == "live" {
		return chat.NewLiveResponder(cfg.OpenAIAPIKey, cfg.OpenAIModel, "", nil)
	}
	return chat.MockResponder{}
}
