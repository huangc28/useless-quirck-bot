# Phase 1: Website Lookup Chatbot MVP - Research

**Created:** 2026-05-28
**Purpose:** Implementation research for planning Phase 1.

## RESEARCH COMPLETE

## Planning-Relevant Findings

### Telegram Bot API

- Telegram's Bot API is HTTP-based. Bot requests use `https://api.telegram.org/bot<token>/<METHOD_NAME>`, and `sendMessage` requires a `chat_id` plus text payload. Source: https://core.telegram.org/bots/api
- Webhooks are a good fit for this MVP because the Go backend can expose one endpoint, parse Telegram update payloads, and reply through `sendMessage`.
- For local demos, the implementation should support a configurable public webhook URL, because Telegram webhooks require a reachable HTTPS endpoint unless using a local Bot API server.

### LLM Router And Structured JSON

- The router should use a structured-output API or a strict post-parse validator so the backend only accepts the allowed modes: `general_chat`, `public_lookup`, and `browser_lookup`.
- OpenAI's Structured Outputs guidance supports JSON Schema-based output constraints with strict validation modes. Source: https://platform.openai.com/docs/guides/structured-outputs
- Even with structured output, the backend should validate the mode enum, confidence range, required fields, and `cache_hint` shape before execution. Router failure must not invoke lookup.

### Go Runtime Shape

- Go 1.25.4 is available locally.
- The backend can use standard `net/http` for the webhook and Telegram API client calls. A custom lightweight client is enough for Phase 1; a Telegram SDK is optional, not required.
- Use dependency interfaces around LLM routing, lookup worker execution, cache, and Telegram sending. This makes the all-live runtime possible while preserving mock fallback for tests/reviewer validation.

### SQLite Cache

- A CGo-free SQLite driver exists for Go through `modernc.org/sqlite`. Source: https://pkg.go.dev/modernc.org/sqlite
- SQLite is suitable for a local demo cache because it is file-based, easy to reset, and supports TTL metadata with simple SQL.
- Use WAL mode and a busy timeout during database initialization. Keep `data/cache.db` out of git.

### Lookup Worker Boundary

- Local `codex` and `claude` commands are available on this machine. The Phase 1 worker should not hard-code one provider as the only path; make the command configurable.
- The CLI worker boundary should be stable JSON over stdin/stdout. Stderr can carry logs, but stdout must contain only the response JSON.
- The backend should pass a timeout context to the subprocess: `20s` for `public_lookup`, `60s` for `browser_lookup`.
- Worker response statuses should be `ok`, `no_result`, `auth_required`, and `error`. Backend reply formatting should branch on status.

### Chrome MCP / Browser Lookup

- Chrome DevTools MCP is an official Chrome DevTools project for giving AI agents access to a live Chrome browser. Sources: https://developer.chrome.com/blog/chrome-devtools-mcp and https://github.com/ChromeDevTools/chrome-devtools-mcp
- Browser lookup should be reserved for `browser_lookup`; public lookup should not open Chrome MCP by default.
- Auth-required handling should be explicit: if a site blocks data behind login, return `auth_required` and ask the operator to prepare the browser profile/session. Do not ask for credentials in Telegram.

## Recommended Plan Shape

1. Create the Go backend skeleton, configuration, and test harness.
2. Implement Telegram webhook parsing and reply sending behind interfaces.
3. Implement the LLM router contract with mock and live modes.
4. Implement SQLite cache with canonical `cache_hint` keys, TTL, stale-if-error, and force refresh detection.
5. Implement lookup worker subprocess contract with mock and live CLI modes.
6. Wire orchestration: Telegram message -> router -> general chat or lookup cache -> worker -> reply.
7. Add demo docs covering live setup, mock fallback, and the three locked demo messages.

## Risks And Mitigations

- **Live demo flakiness:** Telegram webhooks, LLM calls, public web lookup, Chrome MCP, and commerce sites can fail independently. Mitigate with mock fallback and graceful status replies.
- **Router overreach:** LLM could produce plausible but unsafe routing output. Mitigate with strict schema validation, enum checks, confidence threshold, and no lookup on invalid output.
- **Cache poisoning/stale facts:** Volatile price data can go stale. Mitigate with short TTLs, observed timestamps, source evidence, force refresh, and stale-if-error labeling.
- **Worker JSON contamination:** Agent CLI logs can pollute stdout. Mitigate by requiring stdout-only JSON and sending logs to stderr.
- **Credential safety:** Never accept third-party site credentials in Telegram. Auth session preparation is an operator/browser-profile action only.

## Validation Architecture

### Test Fixtures

- `你好` -> `general_chat`; no cache lookup; no worker invocation.
- `請問義美小泡芙多少錢` -> `public_lookup`; cacheable; public lookup worker path.
- `請問 momo 義美小泡芙多少錢` -> `browser_lookup`; cacheable; browser lookup worker path.
- `請幫我查 https://example.com` -> `browser_lookup`; URL preserved in `cache_hint`.
- Low-confidence router output -> clarification reply; no worker invocation.
- Invalid router JSON -> safe fallback; no worker invocation.
- `auth_required` worker response -> safe browser-session preparation reply.
- Expired cache + worker error -> stale result with expired/timestamp labeling.
- Natural language refresh -> bypass fresh cache and update cache on success.

### Verification Commands

- `go test ./...` should cover router validation, orchestration, cache behavior, worker subprocess adapter with mock command, and Telegram handler behavior.
- Manual live demo should exercise the three locked messages in Telegram.

### Evidence Expected From Implementation

- README documents env vars, live modes, mock modes, webhook setup, and demo script.
- `.gitignore` excludes runtime cache DB and secret env files.
- Tests prove mock fallback can validate backend flow without live credentials.

