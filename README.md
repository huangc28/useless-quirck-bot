# Interview Chatbot

## Overview

Telegram chatbot MVP for website data lookup. The backend receives a Telegram update, acknowledges it quickly, routes the message as `general_chat`, `public_lookup`, or `browser_lookup`, uses local Codex CLI for live routing/chat, checks SQLite cache for lookup modes, calls a local lookup worker when needed, and replies with answer text, source evidence, and query time.

## Architecture

Flow: Telegram webhook -> quick HTTP acknowledgement -> background router/app processing -> general chat or lookup path -> SQLite cache -> local lookup worker -> Telegram reply.

Key packages:
- `internal/router` classifies messages and supports mock/local-Codex adapters.
- `internal/chat` handles general chat without lookup/cache and supports mock/local-Codex responders.
- `internal/cache` stores lookup results with TTL and stale-if-error behavior.
- `internal/worker` invokes a local CLI worker through stdin/stdout JSON.
- `internal/liveworker` powers the repo-owned live lookup worker command.
- `internal/app` orchestrates routing, cache, worker, and reply formatting.
- `internal/telegram` parses webhook payloads and sends Telegram messages.

## Environment

Copy `.env.example` to `.env` for local runs. The server loads `.env` on startup; environment variables already set in the shell take precedence.

```sh
SERVER_ADDR=:8080
TELEGRAM_BOT_TOKEN=
TELEGRAM_WEBHOOK_SECRET=
LLM_MODE=mock
CODEX_ROUTER_CMD='codex exec --sandbox read-only --output-schema schemas/router_result.schema.json -'
CODEX_CHAT_CMD='codex exec --sandbox read-only -'
CODEX_TIMEOUT=60s
LOOKUP_WORKER_MODE=mock
LOOKUP_WORKER_CMD=
LIVE_LOOKUP_CODEX_CMD='codex exec --output-schema schemas/lookup_response.schema.json -'
LIVE_LOOKUP_PUBLIC_PROMPT=prompts/public_lookup.md
LIVE_LOOKUP_BROWSER_PROMPT=prompts/browser_lookup.md
LIVE_LOOKUP_TIMEOUT=120s
CACHE_PATH=data/cache.db
PUBLIC_LOOKUP_TIMEOUT=130s
BROWSER_LOOKUP_TIMEOUT=130s
```

`LLM_MODE=mock|live` controls router/general-chat adapters. In `live`, the Go backend calls local `codex exec` through `CODEX_ROUTER_CMD` and `CODEX_CHAT_CMD`; no `OPENAI_API_KEY` is required by the Go process. `LOOKUP_WORKER_MODE=mock|live` controls whether an empty worker command defaults to `./scripts/mock_lookup_worker.sh`. In live lookup mode, set `LOOKUP_WORKER_CMD='go run ./cmd/lookup-worker'`; that command reads the `LIVE_LOOKUP_*` variables and uses the repo-owned prompt files.

## Run Locally

```sh
GOCACHE="$PWD/.gocache" go test ./...
GOCACHE="$PWD/.gocache" go run ./cmd/server
```

The server listens on `SERVER_ADDR` and exposes `POST /telegram/webhook` plus `GET /healthz`.

## Telegram Webhook

Configure the Telegram bot webhook to your HTTPS endpoint that forwards to:

```text
POST /telegram/webhook
```

The handler parses `update_id`, `message.chat.id`, and `message.text`, acknowledges valid text updates immediately, then processes routing/lookup and sends `sendMessage` in the background. Duplicate deliveries with the same `update_id` are ignored so Telegram retries do not produce duplicate visible replies.
When `TELEGRAM_WEBHOOK_SECRET` is set, requests must include Telegram's `X-Telegram-Bot-Api-Secret-Token` header with the same value.

## Mock Fallback

Default mock mode is runnable without LLM or browser credentials:

```sh
LLM_MODE=mock LOOKUP_WORKER_MODE=mock GOCACHE="$PWD/.gocache" go run ./cmd/server
```

Mock router fixtures cover general chat, public lookup, browser lookup, URL lookup, auth-required, no-result, and worker error paths through tests and `scripts/mock_lookup_worker.sh`.

## Live Codex Chat/Router

Use this mode to test Telegram -> Go backend -> local Codex CLI -> Telegram without enabling live website lookup:

```sh
LLM_MODE=live \
LOOKUP_WORKER_MODE=mock \
CODEX_ROUTER_CMD='codex exec --sandbox read-only --output-schema schemas/router_result.schema.json -' \
CODEX_CHAT_CMD='codex exec --sandbox read-only -' \
GOCACHE="$PWD/.gocache" \
go run ./cmd/server
```

The local `codex` CLI must already be authenticated and runnable.

## Live Lookup Worker

Set `LOOKUP_WORKER_MODE=live` and point `LOOKUP_WORKER_CMD` to the repo-owned worker command:

```sh
LOOKUP_WORKER_MODE=live
LOOKUP_WORKER_CMD='go run ./cmd/lookup-worker'
LIVE_LOOKUP_CODEX_CMD='codex exec --output-schema schemas/lookup_response.schema.json -'
```

The worker reads one lookup request JSON object from stdin, branches internally between Public Lookup and Browser Lookup, reads `prompts/public_lookup.md` or `prompts/browser_lookup.md`, and writes one lookup response JSON object to stdout. Invalid request JSON, unknown modes, Codex/tool failures, invalid output, and ungrounded Browser Lookup success become `status:"error"`. Ungrounded Public Lookup success becomes `status:"no_result"` rather than a guessed price.

Browser lookup requires Chrome MCP/browser profile setup. If a target site requires login, prepare the browser session in the local browser profile used by the worker before retrying the Telegram request.

## Live UAT Readiness

Before running the final human UAT, check that the live-demo environment is configured without printing secret values:

```sh
sh scripts/live_uat_check.sh --check-env
```

The script verifies the required live-mode environment and prints the three Telegram messages to send for the demo.
It also rejects `scripts/mock_lookup_worker.sh` for live UAT; use `LOOKUP_WORKER_CMD='go run ./cmd/lookup-worker'` so no live demo path silently substitutes mock success.

## Demo Script

1. `你好`
2. `請問義美小泡芙多少錢`
3. `請問 momo 義美小泡芙多少錢`

## Cache Behavior

Lookup modes use SQLite TTL cache. Cache keys prefer router `cache_hint` metadata such as lookup type, target, site, and URL; fallback keys use mode plus normalized question. Price lookups use a 30 minute TTL, general lookup results use 60 minutes, `no_result` uses 5 minutes, and `error` is not cached. Requests containing `重新查`, `更新`, `refresh`, or `不要快取` bypass cache.

If a cached entry is expired and refresh fails, the bot returns the stale result with explicit expired labeling instead of pretending the refresh succeeded.

## Auth Wall Behavior

For `auth_required`, the bot asks the operator to prepare the browser session used by Chrome MCP/browser automation. Users should not send third-party credentials in Telegram.

## Verification

```sh
GOCACHE="$PWD/.gocache" go test ./...
sh scripts/live_uat_check.sh --check-env
rg "go run ./cmd/lookup-worker|prompts/public_lookup.md|prompts/browser_lookup.md|Duplicate deliveries|Live UAT Readiness" README.md
```
