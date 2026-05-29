---
phase: 01-website-lookup-chatbot-mvp
plan: "03"
subsystem: lookup-runtime
tags: [go, sqlite, cache, subprocess, worker]
requires:
  - phase: 01-01
    provides: shared contracts and config defaults
provides:
  - SQLite lookup cache with fresh, stale, and miss states
  - TTL and force-refresh policy helpers
  - CLI lookup worker client using stdin/stdout JSON
  - Deterministic mock lookup worker script
affects: [app, telegram, demo]
tech-stack:
  added: [modernc.org/sqlite]
  patterns: [ttl-cache, stale-if-error, cli-json-worker]
key-files:
  created:
    - internal/cache/cache.go
    - internal/cache/cache_test.go
    - internal/worker/worker.go
    - internal/worker/worker_test.go
    - scripts/mock_lookup_worker.sh
    - go.sum
  modified:
    - go.mod
key-decisions:
  - "Cache keys prefer cache_hint metadata and fall back to mode plus normalized question."
  - "Lookup workers are invoked as local CLI subprocesses with strict JSON stdout."
patterns-established:
  - "Cache reads return explicit states: miss, fresh, stale."
  - "Worker errors preserve stderr and invalid JSON fails closed."
requirements-completed: ["REQ-001"]
duration: 10min
completed: 2026-05-29
---

# Phase 01: Website Lookup Chatbot MVP Plan 03 Summary

**SQLite lookup cache and CLI worker boundary with deterministic mock lookup responses**

## Performance

- **Duration:** 10 min
- **Started:** 2026-05-28T15:42:03Z
- **Completed:** 2026-05-29T10:36:12Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments

- Added SQLite cache initialization, canonical cache keys, fresh/stale/miss lookup behavior, and upsert storage.
- Added TTL and force-refresh helpers for price, no-result, auth-required, error, and natural-language refresh cases.
- Added CLI worker client and executable mock worker covering ok, no_result, auth_required, and error scenarios.

## Task Commits

1. **Task 1: Implement SQLite TTL cache** - `835a18f`
2. **Task 2: Implement TTL and force-refresh policy helpers** - `835a18f`
3. **Task 3: Implement local CLI lookup worker client** - `256fd7d`

## Files Created/Modified

- `internal/cache/cache.go` - SQLite cache store, keying, TTL, stale, and force-refresh helpers.
- `internal/cache/cache_test.go` - Cache key, fresh/stale/miss, TTL, and force-refresh tests.
- `internal/worker/worker.go` - CLI worker client using `exec.CommandContext`.
- `internal/worker/worker_test.go` - Worker ok, auth_required, invalid JSON, and timeout tests.
- `scripts/mock_lookup_worker.sh` - Deterministic strict-JSON mock lookup worker.
- `go.mod` / `go.sum` - `modernc.org/sqlite` dependency.

## Decisions Made

None - followed plan as specified.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Resolve SQLite dependency through approved network escalation**
- **Found during:** Task 1 verification
- **Issue:** The sandbox could not resolve `modernc.org/sqlite` through `proxy.golang.org`.
- **Fix:** Retried `go get modernc.org/sqlite` with user-approved escalated network access.
- **Files modified:** `go.mod`, `go.sum`
- **Verification:** `GOCACHE=/Users/huangchihan/develop/gamania/interview-chatbot/.gocache go test ./...`
- **Committed in:** `835a18f`

**2. [Rule 3 - Blocking] Fix mock worker test path**
- **Found during:** Task 3 verification
- **Issue:** Package tests run from `internal/worker`, so `./scripts/mock_lookup_worker.sh` was not found.
- **Fix:** Tests now reference `../../scripts/mock_lookup_worker.sh`; runtime defaults can still use `./scripts/mock_lookup_worker.sh` from repo root.
- **Files modified:** `internal/worker/worker_test.go`
- **Verification:** `GOCACHE=/Users/huangchihan/develop/gamania/interview-chatbot/.gocache go test ./...`
- **Committed in:** `256fd7d`

---

**Total deviations:** 2 auto-fixed (Rule 3)
**Impact on plan:** Dependency and test-path fixes only; planned runtime behavior is intact.

## Issues Encountered

None beyond dependency download and package-relative test path handling.

## User Setup Required

None for mock mode. Live lookup requires configuring `LOOKUP_WORKER_CMD` in the final app setup.

## Next Phase Readiness

The app orchestrator can now check lookup cache state, call worker subprocesses, and use mock worker responses for local demo verification.

## Self-Check: PASSED

- `GOCACHE=/Users/huangchihan/develop/gamania/interview-chatbot/.gocache go test ./...` passed.
- Cache tests cover fresh hit, stale hit, miss, no_result TTL, and force refresh detection.
- Worker tests cover ok, auth_required, invalid JSON, and timeout behavior.
- `scripts/mock_lookup_worker.sh` returns strict JSON on stdout.

---
*Phase: 01-website-lookup-chatbot-mvp*
*Completed: 2026-05-29*
