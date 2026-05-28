# Phase 1: Website Lookup Chatbot MVP - Context

**Gathered:** 2026-05-28
**Status:** Ready for planning

<domain>
## Phase Boundary

Build the first shippable take-home demo path for a Telegram-based chatbot that routes every incoming message through an LLM intent router, answers general chat directly, uses public-source lookup for current public information, uses Chrome MCP/browser automation for site-specific lookup, caches lookup results in SQLite, and replies with source-grounded output.

</domain>

<spec_lock>
## Requirements (locked via SPEC.md)

**8 requirements are locked.** See `01-SPEC.md` for full requirements, boundaries, and acceptance criteria.

Downstream agents MUST read `01-SPEC.md` before planning or implementing. Requirements are not duplicated here.

**In scope (from SPEC.md):**
- Telegram as the single chat platform for Phase 1.
- Go backend for webhook handling and orchestration.
- LLM intent router with structured JSON output for `general_chat`, `public_lookup`, and `browser_lookup`.
- Public lookup worker contract for public-source current information.
- Browser lookup worker contract for Chrome MCP/browser automation.
- SQLite TTL cache for lookup modes.
- Auth-required status handling for browser lookup.
- Demo documentation for the interview prompt scenario.

**Out of scope (from SPEC.md):**
- Line, Google Chat, Vyin Chat, or other chat platforms — Phase 1 proves one chat integration first.
- Production deployment and public hosting — local/demo execution is enough for the take-home MVP.
- Dedicated source-specific scrapers for momo, PChome, or commerce sites — lookup behavior is delegated to agent workers.
- Multi-user third-party website credential management — credentials must not be requested through chat.
- Full arbitrary autonomous web browsing — browser lookup is constrained to the user's lookup request and worker contract.
- `.llm-wiki/` runtime caching — volatile lookup results belong in SQLite.

</spec_lock>

<decisions>
## Implementation Decisions

### LLM Router Contract
- **D-01:** Use `Minimal + cache_hint` router output. The router classifies and provides cache-key metadata, but it does not generate execution plans or answers.
- **D-02:** Router modes are limited to `general_chat`, `public_lookup`, and `browser_lookup`.
- **D-03:** Router output includes `mode`, `confidence`, `reason`, `normalized_question`, and optional `cache_hint`.
- **D-04:** `cache_hint` includes `lookup_type`, `target`, `site`, and `url` when confidently extractable.
- **D-05:** If router confidence is below `0.65`, the bot asks a clarification question and does not invoke lookup.
- **D-06:** If router execution fails, returns invalid JSON, or returns an unknown mode, the backend uses a safe non-lookup fallback response.
- **D-07:** `normalized_question` uses light normalization only: trim whitespace, remove command prefix, and preserve the original language, website names, and URLs.

### Lookup Worker Contract
- **D-08:** Use one lookup worker interface for both `public_lookup` and `browser_lookup`; the request `mode` selects behavior.
- **D-09:** Go invokes the lookup worker as a local CLI subprocess. Request JSON is passed on stdin and response JSON is read from stdout.
- **D-10:** Worker responses must be strict structured JSON with `status`, `answer`, `evidence`, `observed_at`, and `error`.
- **D-11:** Supported worker statuses are `ok`, `no_result`, `auth_required`, and `error`.
- **D-12:** Worker timeouts are mode-specific: `public_lookup` uses `20s`; `browser_lookup` uses `60s`.

### Cache Policy
- **D-13:** Cache keys prioritize canonical metadata from `cache_hint` for higher hit rate.
- **D-14:** Cache-key fallback is `mode + normalized_question` when `cache_hint` is missing or insufficient.
- **D-15:** TTL policy is by lookup type/status: `price` is `30m`, general public lookup is `60m`, `no_result` is `5m`, `auth_required` is not cached or cached for at most `1m`, and `error` is not cached.
- **D-16:** Use stale-if-error: if an expired cache entry exists and refresh fails, return the stale result with timestamp and expired labeling.
- **D-17:** Support natural-language force refresh, including wording such as `重新查`, `更新`, `refresh`, and `不要快取`.
- **D-18:** `general_chat` bypasses the lookup cache.

### Demo Runtime Shape
- **D-19:** The primary Phase 1 demo is all live: live Telegram, live LLM router, live lookup worker, and live Telegram reply.
- **D-20:** Keep mock fallback modes for reviewer/local verification even though the primary demo is live.
- **D-21:** The locked demo script includes three messages: `你好`, `請問義美小泡芙多少錢`, and `請問 momo 義美小泡芙多少錢`.
- **D-22:** Live lookup failures return graceful status to Telegram. The system must not pretend success or silently substitute a mock answer.

### Folded Todos
- **Verify browser agent can look up public commerce prices:** Folded into Phase 1 planning as a verification concern. The plan must include a way to verify that lookup can retrieve source-grounded product price data for `請問義美小泡芙多少錢`, while still preserving mock fallback for local/reviewer validation.

### the agent's Discretion
- Choose exact Go package layout, library choices, and test file organization as long as the locked contracts above remain intact.
- Choose exact mock payload contents as long as they exercise `general_chat`, `public_lookup`, `browser_lookup`, cache hit, stale-if-error, and `auth_required` behavior.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Locked Phase Definition
- `.planning/phases/01-website-lookup-chatbot-mvp/01-SPEC.md` — Locked requirements, boundaries, constraints, and acceptance criteria.
- `.planning/ROADMAP.md` — Phase goal, success criteria, and phase boundary.
- `.planning/REQUIREMENTS.md` — Project-level website lookup chatbot requirement.
- `.planning/PROJECT.md` — Project purpose and top-level design direction.

### Prior Decisions
- `.planning/notes/auth-source-specific-fallback.md` — Auth is fallback for source-specific browser sessions, not the default chat flow.
- `.planning/notes/runtime-cache-and-project-wiki.md` — SQLite is the runtime cache; `.llm-wiki/` is durable project knowledge only.

### Folded Phase Todo
- `.planning/todos/pending/verify-browser-agent-public-commerce-prices.md` — Verification concern for source-grounded commerce price lookup.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- None. The repository currently has planning artifacts only.

### Established Patterns
- None from application code. Phase 1 is greenfield and should establish simple Go backend, worker contract, and cache patterns.

### Integration Points
- New code will need to create the Telegram webhook entrypoint, LLM routing boundary, lookup worker subprocess boundary, SQLite cache boundary, and documentation from scratch.

</code_context>

<specifics>
## Specific Ideas

- The required public lookup demo message is `請問義美小泡芙多少錢`.
- The browser lookup demo message is `請問 momo 義美小泡芙多少錢`.
- The general chat demo message is `你好`.
- The backend should be thin: route, cache, call workers, format replies.
- Chrome MCP/browser automation is used for `browser_lookup`, not every lookup.
- Public lookup must use source evidence, not model memory alone.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 01-Website Lookup Chatbot MVP*
*Context gathered: 2026-05-28*
