# Phase 1: Website Lookup Chatbot MVP - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-28
**Phase:** 01-Website Lookup Chatbot MVP
**Areas discussed:** LLM Router Contract, Lookup Worker Contract, Cache Policy, Demo Runtime Shape

---

## LLM Router Contract

| Question | Options Considered | Selected |
|----------|--------------------|----------|
| Router output thickness | Minimal JSON; Task JSON; Rich Plan JSON | Minimal JSON plus `cache_hint` |
| Low confidence handling | Ask clarification; Default to `general_chat`; Default to `public_lookup` | Ask clarification below `0.65` |
| Router failure handling | Safe fallback; Retry once then clarify; Fail closed | Safe fallback |
| Cache key support | Mode + normalized question; Mode + extracted target/cache_hint; Raw message only | Canonical key from `cache_hint`, fallback to mode + normalized question |
| Normalization behavior | Light normalization; LLM rewrite to search query; No normalization | Light normalization |

**Notes:** Router should classify and provide cache metadata only. It must not produce full execution plans or final answers.

---

## Lookup Worker Contract

| Question | Options Considered | Selected |
|----------|--------------------|----------|
| Public/browser worker interface | Same interface; Separate workers; One agent decides internally | Same interface |
| Worker result strictness | Strict structured result; Answer + raw text; Freeform response | Strict structured result |
| Worker timeout | Mode-specific timeout; One timeout for all; No strict timeout | `public_lookup` 20s, `browser_lookup` 60s |
| Worker invocation mechanism | Local CLI worker; Local HTTP worker; Direct library call | Local CLI worker via stdin/stdout JSON |

**Notes:** The worker is the boundary between Go orchestration and local Codex/Claude-style lookup capability.

---

## Cache Policy

| Question | Options Considered | Selected |
|----------|--------------------|----------|
| Cache key strategy | Mode + normalized question; Mode + extracted target/cache_hint; Raw message only | High-hit-rate canonical key from `cache_hint` |
| TTL policy | Short TTL by lookup type; One TTL for all lookup; No TTL | Short TTL by lookup type |
| Expired cache behavior | Stale-if-error; Hard expire; Return stale immediately and refresh in background | Stale-if-error |
| Force refresh behavior | Natural language refresh; Command only; No force refresh | Natural language refresh |

**Notes:** Cache uses canonical hints when available, but must fall back safely when extraction is not confident.

---

## Demo Runtime Shape

| Question | Options Considered | Selected |
|----------|--------------------|----------|
| Test/demo layering | Mock for tests, live for demo; All live; All mock | All live for primary demo |
| Mock fallback mode | Keep fallback mock mode; No mock fallback; Mock only in tests | Keep fallback mock mode |
| Live demo script | Three locked messages; Add URL lookup too; Only product price examples | Three locked messages |
| Live failure behavior | Graceful failure with evidence/status; Fallback to mock answer; Generic error | Graceful failure with evidence/status |

**Notes:** The primary demo should show the real path, but mock fallback remains available for reviewer/local verification.

---

## the agent's Discretion

- Exact Go package structure.
- Exact Go libraries.
- Exact mock payload contents.
- Test organization, provided all locked behaviors are covered.

## Deferred Ideas

None.

