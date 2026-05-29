---
phase: 01-website-lookup-chatbot-mvp
verified: 2026-05-29T11:27:05Z
status: human_needed
score: 41/41 must-haves verified
overrides_applied: 0
human_verification:
  - test: "Live Telegram webhook round trip"
    expected: "A real Telegram message to the configured bot reaches POST /telegram/webhook and sends exactly one reply to the same chat."
    why_human: "Requires a Telegram bot token, registered HTTPS webhook URL, and real Telegram delivery."
  - test: "Live public lookup agent for `請問義美小泡芙多少錢`"
    expected: "With LOOKUP_WORKER_MODE=live and a real lookup agent command, the reply contains current answer text, source evidence, and query time."
    why_human: "Requires external web search/agent execution and judgment that evidence is current and relevant."
  - test: "Live browser lookup with Chrome MCP/profile"
    expected: "`請問 momo 義美小泡芙多少錢` uses the browser lookup worker/Chrome MCP profile and returns evidence or the safe auth-session preparation reply."
    why_human: "Requires local Chrome MCP/browser profile setup and, for auth walls, prepared browser session state."
---

# Phase 1: Website Lookup Chatbot MVP Verification Report

**Phase Goal:** A Telegram chat user can send general chat, public lookup, or browser lookup messages; lookup questions such as `請問義美小泡芙多少錢` receive source-grounded answers through the backend cache and agent lookup flow.  
**Verified:** 2026-05-29T11:27:05Z  
**Status:** human_needed  
**Re-verification:** No - initial verification

## Goal Achievement

All codebase-verifiable roadmap, SPEC, and PLAN must-haves are implemented and wired. The status is `human_needed`, not `passed`, because the remaining validation is live external integration: Telegram delivery, real public lookup agent output, and Chrome MCP/browser profile behavior.

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | ROADMAP: Telegram webhook path can receive a user message and return a chat reply. | VERIFIED | `cmd/server/main.go:48` registers `POST /telegram/webhook`; `cmd/server/main.go:81`, `cmd/server/main.go:91`, and `cmd/server/main.go:96` parse, handle, and send the reply; `cmd/server/main_test.go:32` verifies matching-secret webhook sends once to chat `123`. |
| 2 | ROADMAP: Backend uses an LLM intent router to classify messages as `general_chat`, `public_lookup`, or `browser_lookup`. | VERIFIED | `internal/router/router.go:19` defines router decisions; `internal/router/router.go:26` defines the router interface; `internal/router/router.go:75` and `internal/router/router.go:114` provide mock/live clients; `internal/router/router_test.go:23` through `internal/router/router_test.go:26` cover all required fixture modes. |
| 3 | ROADMAP: General chat messages receive a normal chatbot answer without website lookup. | VERIFIED | `internal/app/app.go:51` through `internal/app/app.go:60` routes `general_chat` only to `Chat.Respond`; `internal/app/app_test.go:16`, `internal/app/app_test.go:35`, and `internal/app/app_test.go:38` verify reply, zero worker calls, and zero cache calls. |
| 4 | ROADMAP: Public lookup messages use public-source lookup and do not open Chrome MCP by default. | VERIFIED | `internal/router/router.go:102` classifies the sample price query as `public_lookup`; `internal/app/app.go:89` through `internal/app/app.go:95` passes the mode to the worker contract; Chrome MCP is documented only for browser lookup in `README.md:75`. |
| 5 | ROADMAP: Browser lookup messages use Chrome MCP/browser automation through a structured contract. | VERIFIED | `internal/router/router.go:94` and `internal/router/router.go:98` classify URL/site-specific requests as `browser_lookup`; `internal/worker/worker.go:17` through `internal/worker/worker.go:18` define the structured CLI contract; `README.md:65` through `README.md:75` documents live worker and Chrome MCP/profile setup. |
| 6 | ROADMAP: Backend checks a SQLite TTL cache before invoking lookup workers for lookup modes. | VERIFIED | `internal/app/app.go:75` through `internal/app/app.go:83` checks cache before `internal/app/app.go:89` invokes the worker; `internal/cache/cache.go:70`, `internal/cache/cache.go:110`, `internal/cache/cache.go:160`, and `internal/cache/cache.go:209` implement SQLite table, get, put, and TTL policy. |
| 7 | ROADMAP: Lookup results include answer text, source evidence, and query time. | VERIFIED | `internal/worker/worker.go:82` through `internal/worker/worker.go:89` rejects `ok` responses without answer, observed time, and usable source evidence; `internal/app/app.go:161` through `internal/app/app.go:199` formats answer/source/time into replies; `internal/app/app_test.go:42` and `internal/app/app_test.go:51` assert these fields. |
| 8 | ROADMAP: Auth-required website responses tell the user/operator that a browser auth session must be prepared. | VERIFIED | `internal/app/app.go:213` through `internal/app/app.go:219` renders browser-session preparation text and tells users not to send third-party credentials; `internal/app/app_test.go:96` and `internal/app/app_test.go:107` verify the safe reply. |
| 9 | ROADMAP: Demo documentation explains how to run the `請問義美小泡芙多少錢` scenario. | VERIFIED | `README.md:44`, `README.md:55`, `README.md:65`, `README.md:77` through `README.md:83`, and `README.md:89` document webhook, mock fallback, live worker, demo messages, cache, and auth wall behavior. |
| 10 | SPEC AC1: Telegram webhook payloads can be parsed and answered with the correct chat ID. | VERIFIED | `internal/telegram/telegram.go:30` parses updates; `internal/telegram/telegram.go:41` and `internal/telegram/telegram.go:62` define/send messages; `internal/telegram/telegram_test.go:12` and `cmd/server/main_test.go:32` verify parsing and reply routing. |
| 11 | SPEC AC2: Router fixtures classify `你好`, `請問義美小泡芙多少錢`, `請問 momo 義美小泡芙多少錢`, and URL lookup examples. | VERIFIED | `internal/router/router_test.go:14` plus `internal/router/router_test.go:23` through `internal/router/router_test.go:26` assert all four locked fixtures and expected modes. |
| 12 | SPEC AC3: `general_chat` returns a reply without invoking lookup workers or lookup cache. | VERIFIED | `internal/app/app_test.go:16`, `internal/app/app_test.go:35`, and `internal/app/app_test.go:38` explicitly verify reply, zero worker calls, and zero cache calls. |
| 13 | SPEC AC4: `public_lookup` invokes public lookup, returns source evidence, and does not open Chrome MCP by default. | VERIFIED | `internal/router/router.go:102` creates `public_lookup`; `internal/app/app.go:89` through `internal/app/app.go:95` invokes worker using the mode; `scripts/mock_lookup_worker.sh:21` returns public evidence; no Chrome path is used except browser documentation in `README.md:75`. |
| 14 | SPEC AC5: `browser_lookup` invokes browser lookup/Chrome MCP contract for site-specific or URL-specific requests. | VERIFIED | `internal/router/router.go:94` and `internal/router/router.go:98` produce browser lookup for URLs/site names; `internal/router/router_test.go:25` and `internal/router/router_test.go:26` cover momo and URL fixtures; `scripts/mock_lookup_worker.sh:17` through `scripts/mock_lookup_worker.sh:18` handles `browser_lookup`. |
| 15 | SPEC AC6: `auth_required` worker output produces a safe reply asking for browser session preparation, not chat credential submission. | VERIFIED | `internal/app/app.go:213` through `internal/app/app.go:219` contains the safe auth wall reply; `internal/app/app_test.go:96` and `internal/app/app_test.go:107` assert `browser profile` and `do not send third-party credentials`. |
| 16 | SPEC AC7: Repeated lookup requests hit SQLite cache before TTL expiration, while general chat bypasses lookup cache. | VERIFIED | `internal/cache/cache_test.go:26` verifies fresh/stale/miss states; `internal/app/app_test.go:61` and `internal/app/app_test.go:74` verify cache hit skips worker; `internal/app/app_test.go:16`, `internal/app/app_test.go:35`, and `internal/app/app_test.go:38` verify general chat bypasses cache. |
| 17 | SPEC AC8: Documentation covers setup and the three demo flows: general chat, public lookup, and browser lookup. | VERIFIED | `README.md:23` through `README.md:36` cover environment/run setup; `README.md:77` through `README.md:81` list `你好`, `請問義美小泡芙多少錢`, and `請問 momo 義美小泡芙多少錢`; `README.md:83` and `README.md:89` cover cache/auth behavior. |

**Score:** 41/41 must-haves verified

### Plan Frontmatter Must-Haves

| Plan | Truth | Status | Evidence |
|------|-------|--------|----------|
| 01-01 | Router contract types include mode, confidence, reason, normalized question, and optional cache hint. | VERIFIED | `internal/contracts/contracts.go:17` and `internal/contracts/contracts.go:24`; schema fields in `schemas/router_result.schema.json:6` through `schemas/router_result.schema.json:33`. |
| 01-01 | Lookup worker contract types include mode-specific requests and strict status-bearing responses. | VERIFIED | `internal/contracts/contracts.go:33`, `internal/contracts/contracts.go:49`, `internal/contracts/contracts.go:66`; `schemas/lookup_request.schema.json:8` and `schemas/lookup_response.schema.json:10`. |
| 01-01 | Config exposes router confidence threshold default `0.65`. | VERIFIED | `internal/config/config.go:13`, `internal/config/config.go:46`; test at `internal/config/config_test.go:25`. |
| 01-01 | Config exposes public/browser lookup timeout defaults of `20s` and `60s`. | VERIFIED | `internal/config/config.go:15`, `internal/config/config.go:16`, `internal/config/config.go:47`, `internal/config/config.go:48`; tests at `internal/config/config_test.go:28` and `internal/config/config_test.go:31`. |
| 01-01 | Config supports live-primary runtime and mock fallback modes. | VERIFIED | `internal/config/config.go:10`, `internal/config/config.go:11`, `internal/config/config.go:40`, `internal/config/config.go:43`; server selection in `cmd/server/main.go:105` and `cmd/server/main.go:112`. |
| 01-02 | Router returns minimal cache-hint fields and never an execution plan. | VERIFIED | `internal/router/router.go:21` through `internal/router/router.go:24` prompts for schema fields and says not to answer lookup questions; `internal/contracts/contracts.go:24` has no execution-plan field. |
| 01-02 | Confidence below `0.65` returns clarification behavior. | VERIFIED | `internal/router/router.go:58` through `internal/router/router.go:64`; tests at `internal/router/router_test.go:56` and `internal/app/app_test.go:196`. |
| 01-02 | Router errors and invalid JSON do not invoke lookup. | VERIFIED | `internal/app/app.go:41` through `internal/app/app.go:49` safe-fallbacks router errors/fallback decisions; invalid JSON test at `internal/router/router_test.go:131`. |
| 01-02 | Normalized question preserves original language, site names, and URLs. | VERIFIED | `internal/router/router.go:35` through `internal/router/router.go:43`; test at `internal/router/router_test.go:48`. |
| 01-02 | `general_chat` bypasses lookup cache. | VERIFIED | `internal/app/app.go:51` through `internal/app/app.go:60`; `internal/app/app_test.go:35` and `internal/app/app_test.go:38`. |
| 01-03 | One local CLI worker interface receives JSON on stdin and returns JSON on stdout. | VERIFIED | `internal/worker/worker.go:17` through `internal/worker/worker.go:18`, `internal/worker/worker.go:56` through `internal/worker/worker.go:58`, and `internal/worker/worker.go:72`. |
| 01-03 | Worker response validation handles `ok`, `no_result`, `auth_required`, and `error`. | VERIFIED | `internal/contracts/contracts.go:12` through `internal/contracts/contracts.go:15`, `internal/worker/worker.go:82` through `internal/worker/worker.go:89`, and `scripts/mock_lookup_worker.sh:8` through `scripts/mock_lookup_worker.sh:21`. |
| 01-03 | Worker subprocess timeouts are `20s` for public lookup and `60s` for browser lookup. | VERIFIED | Defaults in `internal/config/config.go:15` and `internal/config/config.go:16`; selection in `internal/app/app.go:146` through `internal/app/app.go:157`; subprocess timeout in `internal/worker/worker.go:49` through `internal/worker/worker.go:50`. |
| 01-03 | Cache key uses `cache_hint` when available and mode plus normalized question fallback. | VERIFIED | `internal/cache/cache.go:86` through `internal/cache/cache.go:107`; test at `internal/cache/cache_test.go:7`. |
| 01-03 | Cache TTL, stale-if-error, and natural-language force refresh behavior exist. | VERIFIED | TTL in `internal/cache/cache.go:209` through `internal/cache/cache.go:225`; stale fallback in `internal/app/app.go:97` and `internal/app/app.go:104`; force refresh in `internal/cache/cache.go:228`. |
| 01-03 | `general_chat` bypasses lookup cache. | VERIFIED | Same app branch evidence as 01-02; test at `internal/app/app_test.go:16`. |
| 01-04 | Primary demo path supports live Telegram, live LLM router, live lookup worker, and Telegram reply. | VERIFIED | `cmd/server/main.go:39`, `cmd/server/main.go:40`, `cmd/server/main.go:48`, `cmd/server/main.go:91`, `cmd/server/main.go:96`, `cmd/server/main.go:105`, and `cmd/server/main.go:112`. |
| 01-04 | Mock fallback mode remains documented and runnable. | VERIFIED | `cmd/server/main.go:34` through `cmd/server/main.go:35`; `README.md:55` through `README.md:63`; executable script confirmed with `ls -l scripts/mock_lookup_worker.sh`. |
| 01-04 | README documents demo messages `你好`, `請問義美小泡芙多少錢`, and `請問 momo 義美小泡芙多少錢`. | VERIFIED | `README.md:77` through `README.md:81`. |
| 01-04 | Live failures produce graceful status and never silently fake success. | VERIFIED | `internal/app/app.go:97` through `internal/app/app.go:107` and `internal/app/app.go:222`; tests at `internal/app/app_test.go:161` and `internal/app/app_test.go:184`. |
| 01-04 | Telegram webhook payloads can be parsed and answered with the correct chat ID. | VERIFIED | `cmd/server/main.go:81` through `cmd/server/main.go:96`; tests at `cmd/server/main_test.go:32` and `internal/telegram/telegram_test.go:12`. |
| 01-04 | `general_chat` returns a reply without invoking lookup workers or lookup cache. | VERIFIED | `internal/app/app.go:51` through `internal/app/app.go:60`; test at `internal/app/app_test.go:16`. |
| 01-04 | Public and browser lookup route through the correct worker/cache behavior. | VERIFIED | `internal/app/app.go:75` through `internal/app/app.go:95`; router fixtures at `internal/router/router_test.go:24` and `internal/router/router_test.go:25`; mock worker modes at `scripts/mock_lookup_worker.sh:17` and `scripts/mock_lookup_worker.sh:21`. |
| 01-04 | `auth_required` produces a browser-session preparation reply, not credential collection. | VERIFIED | `internal/app/app.go:213` through `internal/app/app.go:219`; test at `internal/app/app_test.go:96`; docs at `README.md:89` through `README.md:91`. |

### Deferred Items

None. The milestone roadmap has no later phase that explicitly defers any Phase 1 must-have.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `go.mod` | Go module | VERIFIED | Declares `module interview-chatbot`; `go test -count=1 ./...` passed. |
| `schemas/router_result.schema.json` | Router schema | VERIFIED | Enumerates `general_chat`, `public_lookup`, `browser_lookup` and required router fields. |
| `schemas/lookup_request.schema.json` | Worker request schema | VERIFIED | Restricts lookup request modes to `public_lookup` and `browser_lookup`. |
| `schemas/lookup_response.schema.json` | Worker response schema | VERIFIED | Enumerates `ok`, `no_result`, `auth_required`, and `error`. |
| `internal/contracts/contracts.go` | Shared contracts | VERIFIED | Defines modes, statuses, cache hint, router result, lookup request/response, evidence, and validation helpers. |
| `internal/config/config.go` | Runtime config | VERIFIED | Loads env defaults for modes, cache path, threshold, and timeouts. |
| `internal/router/router.go` | Intent routing | VERIFIED | Mock/live router clients, normalization, validation decisions, and strict JSON parsing. |
| `internal/chat/chat.go` | General chat | VERIFIED | Mock/live responders behind `Responder`, no cache/worker dependency. |
| `internal/cache/cache.go` | SQLite TTL cache | VERIFIED | SQLite table, WAL/busy timeout, keying, fresh/stale/miss, put, TTL, and force refresh helpers. |
| `internal/worker/worker.go` | Agent worker boundary | VERIFIED | CLI subprocess contract, JSON stdin/stdout, timeout, response validation, and command parsing. |
| `scripts/mock_lookup_worker.sh` | Mock worker fallback | VERIFIED | Executable and returns strict JSON for public/browser/auth/error/no-result cases. |
| `internal/telegram/telegram.go` | Telegram ingress/sender | VERIFIED | Parses update subset and posts `sendMessage` payloads. |
| `internal/app/app.go` | Orchestration | VERIFIED | Routes router decisions through chat/cache/worker and formats replies. |
| `cmd/server/main.go` | HTTP server entrypoint | VERIFIED | Loads config, opens cache, selects live/mock clients, registers webhook, and sends replies. |
| `README.md` | Demo documentation | VERIFIED | Covers setup, webhook, mock/live worker, demo script, cache, auth wall, and verification. |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `cmd/server/main.go` | `internal/telegram.ParseUpdate` | `telegram.ParseUpdate(r.Body)` | WIRED | `cmd/server/main.go:81`. |
| `cmd/server/main.go` | `internal/app.App.HandleMessage` | `bot.HandleMessage(r.Context(), text)` | WIRED | `cmd/server/main.go:91`. |
| `cmd/server/main.go` | Telegram sender | `sender.SendMessage(r.Context(), chatID, reply)` | WIRED | `cmd/server/main.go:96`. |
| `internal/app/app.go` | Router | `a.Router.Route(ctx, text)` plus `router.ValidateResult` | WIRED | `internal/app/app.go:41` and `internal/app/app.go:45`. |
| `internal/app/app.go` | General chat responder | `a.Chat.Respond(ctx, text)` | WIRED | `internal/app/app.go:55`. |
| `internal/app/app.go` | Cache before worker | `a.Cache.Get(ctx, key, now)` before `a.Worker.Lookup` | WIRED | `internal/app/app.go:75` before `internal/app/app.go:89`. |
| `internal/app/app.go` | Worker contract | `a.Worker.Lookup(ctx, contracts.LookupRequest{...}, timeout)` | WIRED | `internal/app/app.go:89` through `internal/app/app.go:95`. |
| `internal/app/app.go` | Cache write | `a.Cache.Put(ctx, key, result, response, ttl, now)` | WIRED | `internal/app/app.go:126`. |
| `internal/worker/worker.go` | OS subprocess agent | `exec.CommandContext` with JSON request on stdin and JSON stdout parse | WIRED | `internal/worker/worker.go:56`, `internal/worker/worker.go:58`, `internal/worker/worker.go:72`. |
| `internal/cache/cache.go` | SQLite runtime cache | SQL table/get/put | WIRED | `internal/cache/cache.go:70`, `internal/cache/cache.go:110`, `internal/cache/cache.go:160`. |
| `README.md` | Live agent setup | `LOOKUP_WORKER_CMD` and Chrome MCP/profile instructions | WIRED | `README.md:65` through `README.md:75`. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| `cmd/server/main.go` | `chatID`, `text`, `reply` | HTTP request body -> `telegram.ParseUpdate` -> `bot.HandleMessage` -> `sender.SendMessage` | Yes, from webhook request and app response | FLOWING |
| `internal/app/app.go` | `result`, `entry`, `response` | Router client, SQLite cache, CLI worker | Yes, app uses routed mode, cache state, and worker output; no hardcoded success on failure | FLOWING |
| `internal/router/router.go` | `RouterResult` | Mock fixture logic or live HTTP JSON response | Yes, mock is deterministic fallback; live parses LLM response and app validates | FLOWING |
| `internal/cache/cache.go` | `Entry.Response` | SQLite `lookup_cache` rows populated by `Put` | Yes, DB row fields and evidence JSON are decoded into response | FLOWING |
| `internal/worker/worker.go` | `LookupResponse` | Configured CLI subprocess stdout | Yes, subprocess output is parsed as exactly one JSON response and validated | FLOWING |
| `internal/chat/chat.go` | General chat reply | Mock responder or live HTTP response | Yes, app only uses this branch for `general_chat` | FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Full Go test suite | `GOCACHE=/Users/huangchihan/develop/gamania/interview-chatbot/.gocache go test ./...` | All packages passed. | PASS |
| Fresh uncached tests | `GOCACHE=/Users/huangchihan/develop/gamania/interview-chatbot/.gocache go test -count=1 ./...` | All packages passed; slowest package `internal/worker` 5.385s. | PASS |
| Static analysis | `GOCACHE=/Users/huangchihan/develop/gamania/interview-chatbot/.gocache go vet ./...` | Exit 0. | PASS |
| Server build | `GOCACHE=/Users/huangchihan/develop/gamania/interview-chatbot/.gocache go build ./cmd/server` | Exit 0; generated ignored binary removed after check. | PASS |
| Mock public worker | `printf ... public_lookup ... | scripts/mock_lookup_worker.sh` | Returned `status:"ok"`, public answer, source URL, and `observed_at`. | PASS |
| Mock browser worker | `printf ... browser_lookup ... | scripts/mock_lookup_worker.sh` | Returned `status:"ok"`, momo answer, source URL, and `observed_at`. | PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| REQ-001 | 01-01, 01-02, 01-03, 01-04 | Chat-based website lookup MVP with source-grounded answers. | SATISFIED | Telegram ingress, routing, cache, worker, and source/time reply formatting are wired across `cmd/server/main.go`, `internal/app/app.go`, `internal/cache/cache.go`, `internal/worker/worker.go`, and `README.md`. |
| SPEC-R1 | SPEC | Telegram message ingress. | SATISFIED | Covered by Observable Truths 1 and 10. |
| SPEC-R2 | SPEC | LLM intent routing. | SATISFIED | Covered by Observable Truths 2 and 11. |
| SPEC-R3 | SPEC | General chat response. | SATISFIED | Covered by Observable Truths 3 and 12. |
| SPEC-R4 | SPEC | Public lookup response. | SATISFIED | Covered by Observable Truths 4 and 13. |
| SPEC-R5 | SPEC | Browser lookup response. | SATISFIED | Covered by Observable Truths 5 and 14. |
| SPEC-R6 | SPEC | Auth wall fallback. | SATISFIED | Covered by Observable Truths 8 and 15. |
| SPEC-R7 | SPEC | SQLite TTL cache for lookup modes. | SATISFIED | Covered by Observable Truths 6 and 16. |
| SPEC-R8 | SPEC | Demo documentation. | SATISFIED | Covered by Observable Truths 9 and 17. |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| None | - | - | - | Scans found no TODO/FIXME/placeholder implementations, no empty returns, and no console-log-only behavior. Mock/fake references are intentional tests or documented fallback modes, not hidden stubs. |

### Human Verification Required

### 1. Live Telegram Webhook Round Trip

**Test:** Run the server with a real `TELEGRAM_BOT_TOKEN`, registered HTTPS webhook URL, and optional `TELEGRAM_WEBHOOK_SECRET`; send `你好` from Telegram.  
**Expected:** The server receives the update and sends exactly one reply to the same chat ID.  
**Why human:** Requires Telegram infrastructure and a reachable HTTPS webhook.

### 2. Live Public Lookup Agent

**Test:** Configure `LOOKUP_WORKER_MODE=live` and `LOOKUP_WORKER_CMD` to a real source-grounded lookup agent; send `請問義美小泡芙多少錢`.  
**Expected:** Reply includes current answer text, source evidence, and query time.  
**Why human:** Requires external web lookup and relevance/currentness judgment.

### 3. Live Browser Lookup / Chrome MCP

**Test:** Configure the live worker with Chrome MCP/browser profile access; send `請問 momo 義美小泡芙多少錢`.  
**Expected:** Browser lookup path is used and returns source evidence, or returns the safe browser-session preparation message if auth is required.  
**Why human:** Requires local browser profile/session state and Chrome MCP setup.

### Gaps Summary

No codebase gaps found. The implementation is substantive, wired, and covered by automated tests. The only remaining gate is human UAT for live external integrations.

---

_Verified: 2026-05-29T11:27:05Z_  
_Verifier: the agent (gsd-verifier)_
