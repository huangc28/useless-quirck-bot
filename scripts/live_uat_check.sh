#!/bin/sh
set -eu

if [ "${1:-}" != "--check-env" ]; then
  echo "usage: sh scripts/live_uat_check.sh --check-env" >&2
  exit 2
fi

missing=0

require_non_empty() {
  name="$1"
  value="${2:-}"
  if [ -z "$value" ]; then
    echo "missing $name"
    missing=1
  fi
}

require_equals() {
  name="$1"
  value="${2:-}"
  want="$3"
  if [ "$value" != "$want" ]; then
    echo "missing $name=$want"
    missing=1
  fi
}

require_non_empty "TELEGRAM_BOT_TOKEN" "${TELEGRAM_BOT_TOKEN:-}"
require_non_empty "PUBLIC_BASE_URL" "${PUBLIC_BASE_URL:-}"
require_equals "LLM_MODE" "${LLM_MODE:-}" "live"
require_equals "LOOKUP_WORKER_MODE" "${LOOKUP_WORKER_MODE:-}" "live"
require_non_empty "LOOKUP_WORKER_CMD" "${LOOKUP_WORKER_CMD:-}"

case "${LOOKUP_WORKER_CMD:-}" in
  *mock_lookup_worker.sh*)
    echo "LOOKUP_WORKER_CMD must use the repo-owned live worker: go run ./cmd/lookup-worker"
    missing=1
    ;;
esac

if [ "$missing" -ne 0 ]; then
  echo "Live UAT readiness: missing required environment"
  exit 1
fi

base="${PUBLIC_BASE_URL%/}"

echo "Live UAT readiness: ok"
echo "Webhook endpoint: POST $base/telegram/webhook"
echo "Run the server, register the Telegram webhook, then send these messages from Telegram:"
echo "1. 你好"
echo "2. 請問義美小泡芙多少錢"
echo "3. 請問 momo 義美小泡芙多少錢"
echo "Record whether each reply includes exactly one visible bot message, source evidence when applicable, and observed/query time."
