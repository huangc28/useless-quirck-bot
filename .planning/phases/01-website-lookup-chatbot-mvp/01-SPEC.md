# Phase 1: Website Lookup Chatbot MVP — Specification

**Created:** 2026-05-28
**Ambiguity score:** 0.18 (gate: <= 0.20)
**Requirements:** 8 locked

## Goal

A Telegram chat user can send general chat, public lookup, or browser lookup messages, and lookup messages receive source-grounded answers through a Go backend, LLM intent router, SQLite TTL cache, and agent lookup flow.

## Background

The repository currently has planning artifacts only. No Go backend, Telegram webhook, intent router, SQLite cache, lookup worker contract, Chrome MCP integration, or demo documentation exists. The take-home prompt asks for a chatbot that can search website data from a chat window and return results; `請問義美小泡芙多少錢` is the required demo query.

## Requirements

1. **Telegram message ingress**: The system accepts Telegram webhook updates and returns a reply to the originating chat.
   - Current: No backend or Telegram webhook handler exists.
   - Target: A Go backend exposes a Telegram webhook endpoint, parses incoming text messages, and sends replies through the Telegram Bot API.
   - Acceptance: A test webhook payload containing a text message results in exactly one outbound Telegram reply payload addressed to the source chat ID.

2. **LLM intent routing**: Every incoming text message is classified into `general_chat`, `public_lookup`, or `browser_lookup` before execution.
   - Current: No message router exists.
   - Target: The backend calls an LLM intent router that returns validated structured JSON with one of the three allowed modes.
   - Acceptance: Fixtures classify `你好` as `general_chat`, `請問義美小泡芙多少錢` as `public_lookup`, `請問 momo 義美小泡芙多少錢` as `browser_lookup`, and `請幫我查 https://example.com` as `browser_lookup`.

3. **General chat response**: General chat messages receive a normal chatbot answer without website lookup or cache lookup.
   - Current: No general chat flow exists.
   - Target: `general_chat` mode routes to an LLM response path and does not invoke lookup workers or SQLite lookup cache.
   - Acceptance: A `general_chat` fixture returns a reply and records no lookup worker invocation.

4. **Public lookup response**: Public lookup messages use public-source lookup and do not open Chrome MCP by default.
   - Current: No public lookup worker contract exists.
   - Target: `public_lookup` mode invokes a public lookup worker that retrieves source-grounded current information and returns structured evidence.
   - Acceptance: The sample query `請問義美小泡芙多少錢` produces an answer containing source evidence and query time through the public lookup path.

5. **Browser lookup response**: Browser lookup messages use Chrome MCP/browser automation through a structured worker contract.
   - Current: No browser lookup worker contract or Chrome MCP invocation path exists.
   - Target: `browser_lookup` mode invokes a browser lookup worker with the original question and returns structured status, answer, source evidence, and query time.
   - Acceptance: A fixture for `請問 momo 義美小泡芙多少錢` invokes the browser lookup path, not the public lookup path.

6. **Auth wall fallback**: Auth-required browser lookup results are reported clearly to the chat user/operator.
   - Current: No auth-required status exists.
   - Target: Browser lookup worker output supports an `auth_required` status with site/context; the backend converts it into a reply asking for the browser profile/session to be prepared.
   - Acceptance: A worker response with `status: "auth_required"` produces a Telegram reply that does not ask for credentials in chat and clearly says a browser session must be prepared.

7. **SQLite TTL cache for lookup modes**: Lookup results are cached for `public_lookup` and `browser_lookup` modes only.
   - Current: No runtime cache exists.
   - Target: The backend checks SQLite before invoking lookup workers, stores successful lookup results with expiration metadata, and bypasses cache for `general_chat`.
   - Acceptance: Repeating the same lookup fixture before expiration returns the cached answer without a second worker invocation; repeating a general chat fixture does not use the lookup cache.

8. **Demo documentation**: The repository documents how to run and demonstrate the MVP.
   - Current: No README or demo instructions exist.
   - Target: Documentation explains setup, environment variables, Telegram webhook flow, lookup modes, cache behavior, auth fallback, and the required demo query.
   - Acceptance: A reviewer can follow the documentation to identify how `請問義美小泡芙多少錢`, `請問 momo 義美小泡芙多少錢`, and `你好` are expected to flow through the system.

## Boundaries

**In scope:**
- Telegram as the single chat platform for Phase 1.
- Go backend for webhook handling and orchestration.
- LLM intent router with structured JSON output for `general_chat`, `public_lookup`, and `browser_lookup`.
- Public lookup worker contract for public-source current information.
- Browser lookup worker contract for Chrome MCP/browser automation.
- SQLite TTL cache for lookup modes.
- Auth-required status handling for browser lookup.
- Demo documentation for the interview prompt scenario.

**Out of scope:**
- Line, Google Chat, Vyin Chat, or other chat platforms — Phase 1 proves one chat integration first.
- Production deployment and public hosting — local/demo execution is enough for the take-home MVP.
- Dedicated source-specific scrapers for momo, PChome, or commerce sites — lookup behavior is delegated to agent workers.
- Multi-user third-party website credential management — credentials must not be requested through chat.
- Full arbitrary autonomous web browsing — browser lookup is constrained to the user's lookup request and worker contract.
- `.llm-wiki/` runtime caching — volatile lookup results belong in SQLite.

## Constraints

- Router output must be validated before execution; unknown modes must not invoke tools.
- Lookup workers must return structured output rather than prose-only responses.
- Public lookup must not answer volatile facts from model memory alone; it must include source evidence.
- Chrome MCP/browser automation is used for `browser_lookup`, not for every lookup.
- Cache entries must include expiration metadata and observed/query time.
- Secrets such as Telegram bot tokens and LLM API keys must be provided through environment variables or local configuration, not committed files.

## Acceptance Criteria

- [ ] Telegram webhook payloads can be parsed and answered with the correct chat ID.
- [ ] LLM router fixtures classify `你好`, `請問義美小泡芙多少錢`, `請問 momo 義美小泡芙多少錢`, and URL lookup examples into the expected modes.
- [ ] `general_chat` returns a reply without invoking lookup workers or lookup cache.
- [ ] `public_lookup` invokes public lookup, returns source evidence, and does not open Chrome MCP by default.
- [ ] `browser_lookup` invokes browser lookup/Chrome MCP contract for site-specific or URL-specific requests.
- [ ] `auth_required` worker output produces a safe reply asking for browser session preparation, not chat credential submission.
- [ ] Repeated lookup requests hit SQLite cache before TTL expiration.
- [ ] Documentation covers setup and the three demo flows: general chat, public lookup, and browser lookup.

## Ambiguity Report

| Dimension          | Score | Min   | Status | Notes |
|--------------------|-------|-------|--------|-------|
| Goal Clarity       | 0.88  | 0.75  | ✓      | Phase goal names the platform, modes, cache, and lookup flow. |
| Boundary Clarity   | 0.84  | 0.70  | ✓      | Explicit single-platform MVP and out-of-scope list. |
| Constraint Clarity | 0.74  | 0.65  | ✓      | Key constraints locked for routing, lookup evidence, Chrome MCP use, cache, and secrets. |
| Acceptance Criteria| 0.76  | 0.70  | ✓      | Fixtures and flow checks define pass/fail behavior. |
| **Ambiguity**      | 0.18  | <=0.20| ✓      | Gate passed. |

Status: ✓ = met minimum, ⚠ = below minimum (planner treats as assumption)

## Interview Log

| Round | Perspective | Question summary | Decision locked |
|-------|-------------|------------------|-----------------|
| 1 | Researcher | Is the objective only price lookup? | No. The product is website data lookup through chat; price lookup is the first demo path. |
| 1 | Researcher | When should Chrome MCP be used? | Chrome MCP is only for browser lookup, such as specified websites, URLs, or auth/session needs. |
| 2 | Simplifier | Should the backend hard-code website/source routing? | No. Keep backend thin and use LLM intent routing plus agent workers. |
| 2 | Simplifier | What should happen to generic chat? | Generic messages use `general_chat` and get normal LLM replies without lookup/cache. |
| 3 | Boundary Keeper | How are lookup modes chosen? | LLM intent router classifies each message into `general_chat`, `public_lookup`, or `browser_lookup`. |
| 3 | Boundary Keeper | What is out of scope for Phase 1? | Multi-platform chat, production deployment, dedicated scrapers, credential management, arbitrary autonomous browsing, and `.llm-wiki` runtime cache. |

---

*Phase: 01-website-lookup-chatbot-mvp*
*Spec created: 2026-05-28*
*Next step: $gsd-discuss-phase 1 — implementation decisions (how to build what's specified above)*
