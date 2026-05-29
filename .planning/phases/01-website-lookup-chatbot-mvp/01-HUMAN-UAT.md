---
status: partial
phase: 01-website-lookup-chatbot-mvp
source: [01-VERIFICATION.md]
started: 2026-05-29T11:30:27Z
updated: 2026-05-29T11:30:27Z
---

# Phase 01 Human UAT

## Current Test

Awaiting human testing for live external integrations.

## Tests

### 1. Live Telegram Webhook Round Trip

expected: A real Telegram message to the configured bot reaches `POST /telegram/webhook` and sends exactly one reply to the same chat.
result: pending

### 2. Live Public Lookup Agent

expected: With `LOOKUP_WORKER_MODE=live` and a real lookup agent command, `請問義美小泡芙多少錢` returns current answer text, source evidence, and query time.
result: pending

### 3. Live Browser Lookup With Chrome MCP/Profile

expected: `請問 momo 義美小泡芙多少錢` uses the browser lookup worker/Chrome MCP profile and returns evidence or the safe auth-session preparation reply.
result: pending

## Summary

total: 3
passed: 0
issues: 0
pending: 3
skipped: 0
blocked: 0

## Gaps

None reported yet.
