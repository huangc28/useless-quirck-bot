# Interview Chatbot

給面試官快速啟動的 Telegram live demo。

Bot：[@useless_quirk_bot](https://t.me/useless_quirk_bot)

這個 bot 會把 Telegram 訊息分成一般對話、公開資料查詢、指定網站查詢，並要求查詢答案附上來源證據。

## 不想讀文件？交給 AI agent 架

Nowadays no one reads docs. If you use Codex, Claude Code, Cursor, or another local coding agent, paste this prompt and let it set up the demo:

```text
Set up this Telegram live demo for me.

Goal:
- Run the Go server locally on :8080.
- Expose it with ngrok.
- Configure Telegram webhook for @useless_quirk_bot.
- Verify the bot can answer the demo messages.

Use these settings:
- Bot token: 8803375396:AAGGsS_dTK1rZKXbY8OSSDUSUz6yLqLJlEQ
- Live mode only: LLM_MODE=live and LOOKUP_WORKER_MODE=live
- Lookup worker command: go run ./cmd/lookup-worker

Steps:
1. Check that `go`, `codex`, and `ngrok` are available.
2. Check Codex with: `codex exec 'reply with ok'`.
3. Start `ngrok http 8080` and capture the HTTPS forwarding URL.
4. Create `.env` using that URL as PUBLIC_BASE_URL and the bot token above.
5. Run `GOCACHE="$PWD/.gocache" go test ./...`.
6. Start `GOCACHE="$PWD/.gocache" go run ./cmd/server`.
7. Set the Telegram webhook to `${PUBLIC_BASE_URL}/telegram/webhook`.
8. Tell me to test @useless_quirk_bot with:
   - 你好
   - 義美小泡芙多少錢?
   - 請問 momo 義美小泡芙多少錢

If anything fails, stop and show the exact command and error.
Do not commit changes unless I explicitly ask.
```

## 需要先準備

- Go 1.25+
- `codex` CLI 已登入
- `ngrok`

先確認 Codex CLI 可以跑：

```sh
codex exec 'reply with ok'
```

## 1. 開 ngrok

開一個 terminal：

```sh
ngrok http 8080
```

複製 ngrok 顯示的 HTTPS URL，例如：

```sh
https://example.ngrok-free.app
```

## 2. 建立 `.env`

開第二個 terminal，在 repo 根目錄執行。把 `PUBLIC_BASE_URL` 換成你的 ngrok HTTPS URL。

```sh
export PUBLIC_BASE_URL="https://example.ngrok-free.app"

cat > .env <<EOF
SERVER_ADDR=:8080
PUBLIC_BASE_URL=$PUBLIC_BASE_URL
TELEGRAM_BOT_TOKEN=8803375396:AAGGsS_dTK1rZKXbY8OSSDUSUz6yLqLJlEQ
TELEGRAM_WEBHOOK_SECRET=

LLM_MODE=live
CODEX_ROUTER_CMD='codex exec --sandbox read-only --output-schema schemas/router_result.schema.json -'
CODEX_CHAT_CMD='codex exec --sandbox read-only -'
CODEX_TIMEOUT=60s

LOOKUP_WORKER_MODE=live
LOOKUP_WORKER_CMD='go run ./cmd/lookup-worker'
LIVE_LOOKUP_CODEX_CMD='codex exec --output-schema schemas/lookup_response.schema.json -'
LIVE_LOOKUP_PUBLIC_PROMPT=prompts/public_lookup.md
LIVE_LOOKUP_BROWSER_PROMPT=prompts/browser_lookup.md
LIVE_LOOKUP_TIMEOUT=120s

CACHE_PATH=data/cache.db
ROUTER_CONFIDENCE_THRESHOLD=0.65
PUBLIC_LOOKUP_TIMEOUT=130s
BROWSER_LOOKUP_TIMEOUT=130s
EOF
```

## 3. 啟動 server

```sh
GOCACHE="$PWD/.gocache" go test ./...
GOCACHE="$PWD/.gocache" go run ./cmd/server
```

健康檢查：

```sh
curl http://localhost:8080/healthz
```

應該看到：

```text
ok
```

## 4. 設定 Telegram webhook

再開第三個 terminal：

```sh
source .env
curl -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook" \
  -d "url=${PUBLIC_BASE_URL}/telegram/webhook"
```

確認 webhook：

```sh
curl "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/getWebhookInfo"
```

## 5. 在 Telegram 測試

打開 [@useless_quirk_bot](https://t.me/useless_quirk_bot)，依序傳：

```text
你好
義美小泡芙多少錢?
請問 momo 義美小泡芙多少錢
```

預期結果：

- `你好`：快速回覆一般對話。
- `義美小泡芙多少錢?`：live public lookup，會回覆價格、來源 URL、查詢時間。這題可能需要 20 到 120 秒。
- `請問 momo 義美小泡芙多少錢`：指定網站 lookup。若本機瀏覽器/session 沒準備好，回覆需要準備 browser session 是合理結果。

## 常用排查

查看 server log：

```sh
GOCACHE="$PWD/.gocache" go run ./cmd/server
```

重新設定 webhook：

```sh
source .env
curl "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/deleteWebhook"
curl -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook" \
  -d "url=${PUBLIC_BASE_URL}/telegram/webhook"
```

直接測 live lookup worker：

```sh
printf '%s\n' '{"mode":"public_lookup","question":"義美小泡芙多少錢?","normalized_question":"義美小泡芙多少錢?","cache_hint":{"lookup_type":"price","target":"義美小泡芙"},"force_refresh":false}' \
  | go run ./cmd/lookup-worker
```

## Mock fallback

如果只是要確認 Telegram webhook plumbing，不想跑 live Codex lookup：

```sh
LLM_MODE=mock \
LOOKUP_WORKER_MODE=mock \
GOCACHE="$PWD/.gocache" \
go run ./cmd/server
```

這只驗證流程，不代表 live lookup 品質。
