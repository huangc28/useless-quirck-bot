---
phase: 01
slug: website-lookup-chatbot-mvp
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-05-28
---

# Phase 01 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go test |
| **Config file** | `go.mod` |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~10 seconds after dependencies are available |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test ./...`
- **Before `$gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds for mock/test mode

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 1 | Telegram ingress | — | Bot token is read from env, not committed | unit | `go test ./...` | ❌ W0 | ⬜ pending |
| 01-01-02 | 01 | 1 | LLM router | — | Invalid router output cannot trigger lookup | unit | `go test ./...` | ❌ W0 | ⬜ pending |
| 01-01-03 | 01 | 1 | SQLite cache | — | Cache stores no secrets; stale values are labeled | unit | `go test ./...` | ❌ W0 | ⬜ pending |
| 01-01-04 | 01 | 1 | Lookup worker contract | — | Subprocess timeout prevents unbounded execution | unit | `go test ./...` | ❌ W0 | ⬜ pending |
| 01-01-05 | 01 | 1 | Orchestration | — | `general_chat` bypasses lookup/cache | unit | `go test ./...` | ❌ W0 | ⬜ pending |
| 01-01-06 | 01 | 1 | Documentation | — | Secrets documented as env vars only | grep/manual | `go test ./...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `go.mod` — module initialized.
- [ ] Unit tests for router validation and fallback behavior.
- [ ] Unit tests for cache key, TTL, stale-if-error, and force refresh behavior.
- [ ] Unit tests for lookup worker subprocess contract using mock command.
- [ ] Unit tests for Telegram webhook parsing and reply payload generation.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Live Telegram demo | Demo runtime shape | Requires bot token and reachable webhook URL | Send `你好`, `請問義美小泡芙多少錢`, and `請問 momo 義美小泡芙多少錢` to the configured bot and verify mode-specific replies. |
| Live browser lookup | Browser lookup | Requires Chrome MCP/browser profile and live target site behavior | Run the browser lookup demo message and verify source-grounded result or graceful auth/status failure. |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s in mock/test mode
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending

