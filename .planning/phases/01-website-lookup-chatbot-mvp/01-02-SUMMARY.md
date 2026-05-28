---
phase: 01-website-lookup-chatbot-mvp
plan: "02"
subsystem: ai-routing
tags: [go, router, llm, chat]
requires:
  - phase: 01-01
    provides: shared contracts and config defaults
provides:
  - Intent router interface with mock and live implementations
  - Router validation decisions for executable, clarify, and fallback paths
  - General chat responder interface with mock and live implementations
affects: [app, telegram, cache, worker]
tech-stack:
  added: []
  patterns: [interface-boundaries, fake-roundtripper-tests, fail-closed-routing]
key-files:
  created:
    - internal/router/router.go
    - internal/router/router_test.go
    - internal/chat/chat.go
    - internal/chat/chat_test.go
  modified: []
key-decisions:
  - "Router only classifies; it does not answer lookup questions."
  - "General chat is isolated from lookup cache and worker packages."
patterns-established:
  - "Live LLM adapters accept API key/model/endpoint/http.Client and can be tested without network listeners."
  - "Router validation converts invalid or low-confidence output into non-executable decisions."
requirements-completed: ["REQ-001"]
duration: 4min
completed: 2026-05-28
---

# Phase 01: Website Lookup Chatbot MVP Plan 02 Summary

**Intent router and general-chat responder with deterministic mock fixtures and live HTTP adapters**

## Performance

- **Duration:** 4 min
- **Started:** 2026-05-28T15:38:08Z
- **Completed:** 2026-05-28T15:42:03Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments

- Added router classification for `general_chat`, `public_lookup`, `browser_lookup`, and URL-specific browser lookup.
- Added live router adapter that requests structured routing JSON and refuses prose answers.
- Added mock/live general chat responders without importing cache or worker packages.

## Task Commits

1. **Task 1: Implement router validation and mock router** - `182d95e`
2. **Task 2: Add live-router adapter behind the same interface** - `182d95e`
3. **Task 3: Implement general chat responder with mock and live modes** - `4989e14`

## Files Created/Modified

- `internal/router/router.go` - Router interface, mock routing, validation decisions, and live HTTP adapter.
- `internal/router/router_test.go` - Locked fixture routing, low-confidence fallback, invalid JSON, and live adapter tests.
- `internal/chat/chat.go` - General chat responder interface, mock responder, and live HTTP adapter.
- `internal/chat/chat_test.go` - Mock greeting and live adapter tests.

## Decisions Made

None - followed plan as specified.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Avoid localhost listeners in sandboxed tests**
- **Found during:** Task 2 verification
- **Issue:** `httptest.NewServer` failed because the sandbox does not permit binding localhost ports.
- **Fix:** Replaced listener-based tests with fake `http.RoundTripper` clients.
- **Files modified:** `internal/router/router_test.go`, `internal/chat/chat_test.go`
- **Verification:** `GOCACHE=/Users/huangchihan/develop/gamania/interview-chatbot/.gocache go test ./...`
- **Committed in:** `182d95e`, `4989e14`

---

**Total deviations:** 1 auto-fixed (Rule 3)
**Impact on plan:** Test transport changed only; live adapter production behavior remains HTTP based.

## Issues Encountered

None beyond the sandbox listener restriction documented above.

## User Setup Required

None - no external service configuration required for mock mode.

## Next Phase Readiness

The app orchestrator can use `router.Client` and `chat.Responder` interfaces without knowing whether they are mock or live.

## Self-Check: PASSED

- `GOCACHE=/Users/huangchihan/develop/gamania/interview-chatbot/.gocache go test ./...` passed.
- Router fixtures classify the locked demo messages correctly.
- Invalid router output cannot produce an executable route.
- `internal/chat` and `internal/router` do not import lookup cache or worker packages.

---
*Phase: 01-website-lookup-chatbot-mvp*
*Completed: 2026-05-28*
