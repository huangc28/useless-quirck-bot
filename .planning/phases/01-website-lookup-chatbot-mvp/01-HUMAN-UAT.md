---
status: partial
phase: 01-website-lookup-chatbot-mvp
source: [01-VERIFICATION.md]
started: 2026-05-29T11:30:27Z
updated: 2026-05-29T17:59:48Z
---

# Phase 01 Human UAT

## Current Test

 Automated implementation for async Telegram handling and the repo-owned Live Lookup Worker is ready. Awaiting human testing for live external integrations.

## Tests

### 1. Live Telegram Webhook Round Trip

expected: A real Telegram message to the configured bot reaches `POST /telegram/webhook` and sends exactly one reply to the same chat.
result: passed
evidence: User confirmed that typing `hello` in Telegram produces a bot reply. Reply content still needs separate diagnosis.

### 2. Live Public Lookup Agent

expected: With `LOOKUP_WORKER_MODE=live` and a real lookup agent command, `請問義美小泡芙多少錢` returns current answer text, source evidence, and query time.
result: pending
evidence: Automated worker checks pass for Public Lookup prompt dispatch, source-free no-result behavior, price caution text, strict Source Evidence validation, invalid output, tool failure, and timeout. Live Telegram/Codex run still needs human observation.

### 3. Live Browser Lookup With Chrome MCP/Profile

expected: `請問 momo 義美小泡芙多少錢` uses the browser lookup worker/Chrome MCP profile and returns evidence or the safe auth-session preparation reply.
result: pending
evidence: Automated worker checks pass for Browser Lookup prompt dispatch, named-site request preservation, grounded ok response, auth-required response, and source-free failure behavior. Live Chrome MCP/browser-session run still needs human observation.

## Readiness Check

Before human UAT, run:

```sh
sh scripts/live_uat_check.sh --check-env
```

The check verifies required live-mode environment variables without printing secret values, then prints the three Telegram messages to send.

latest result: passed locally after setting `LOOKUP_WORKER_CMD='go run ./cmd/lookup-worker'` in `.env`. Remaining gate is sending the listed messages through the real Telegram bot and recording the observed replies.

## Summary

total: 3
passed: 1
issues: 0
pending: 2
skipped: 0
blocked: 0

## Gaps

Telegram round-trip delivery previously worked. Async dedupe and live worker contract checks are automated; final Public Lookup and Browser Lookup UAT still requires a real Telegram bot, live Codex/tooling, and human judgment that returned Source Evidence is current and relevant.
