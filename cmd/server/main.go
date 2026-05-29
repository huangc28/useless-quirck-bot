package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"interview-chatbot/internal/app"
	"interview-chatbot/internal/cache"
	"interview-chatbot/internal/chat"
	"interview-chatbot/internal/config"
	"interview-chatbot/internal/router"
	"interview-chatbot/internal/telegram"
	"interview-chatbot/internal/worker"
)

func main() {
	loadRuntimeEnv()
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

func loadRuntimeEnv() {
	_ = godotenv.Load()
}

func telegramWebhook(bot *app.App, sender telegram.Sender, secret string) http.HandlerFunc {
	deduper := newUpdateDeduper(1024)
	return func(w http.ResponseWriter, r *http.Request) {
		if secret != "" && r.Header.Get("X-Telegram-Bot-Api-Secret-Token") != secret {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		updateID, chatID, text, ok, err := telegram.ParseUpdate(r.Body)
		if err != nil {
			http.Error(w, "invalid telegram update", http.StatusBadRequest)
			return
		}
		if !ok {
			log.Print("telegram webhook ignored: no text message")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if !deduper.Remember(updateID) {
			log.Printf("telegram webhook duplicate ignored update_id=%d chat_id=%d", updateID, chatID)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok\n"))
			return
		}
		log.Printf("telegram webhook received update_id=%d chat_id=%d text_len=%d", updateID, chatID, len([]rune(text)))
		go processTelegramMessage(context.Background(), bot, sender, chatID, text)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	}
}

type updateDeduper struct {
	mu    sync.Mutex
	seen  map[int64]struct{}
	order []int64
	limit int
}

func newUpdateDeduper(limit int) *updateDeduper {
	if limit <= 0 {
		limit = 1024
	}
	return &updateDeduper{seen: make(map[int64]struct{}), limit: limit}
}

func (d *updateDeduper) Remember(updateID int64) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.seen[updateID]; ok {
		return false
	}
	d.seen[updateID] = struct{}{}
	d.order = append(d.order, updateID)
	if len(d.order) > d.limit {
		oldest := d.order[0]
		d.order = d.order[1:]
		delete(d.seen, oldest)
	}
	return true
}

func processTelegramMessage(ctx context.Context, bot *app.App, sender telegram.Sender, chatID int64, text string) {
	reply, err := bot.HandleMessage(ctx, text)
	if err != nil {
		log.Printf("telegram message handling failed chat_id=%d: %v", chatID, err)
		return
	}
	if err := sender.SendMessage(ctx, chatID, reply); err != nil {
		log.Printf("telegram send failed chat_id=%d: %v", chatID, err)
		return
	}
	log.Printf("telegram webhook replied chat_id=%d reply_len=%d", chatID, len([]rune(reply)))
}

func routerClient(cfg config.Config) router.Client {
	if cfg.LLMMode == "live" {
		return router.NewLiveClient(cfg.CodexRouterCommand, cfg.CodexTimeout, cfg.RouterConfidenceThreshold)
	}
	return router.MockClient{Threshold: cfg.RouterConfidenceThreshold}
}

func chatResponder(cfg config.Config) chat.Responder {
	if cfg.LLMMode == "live" {
		return chat.NewLiveResponder(cfg.CodexChatCommand, cfg.CodexTimeout)
	}
	return chat.MockResponder{}
}
