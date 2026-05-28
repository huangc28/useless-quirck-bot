# Roadmap

## Current Milestone: MVP Demo

- [ ] **Phase 1: Website Lookup Chatbot MVP** — Telegram-based chatbot demo for website lookup requests.

### Phase 1: Website Lookup Chatbot MVP

**Goal:** A Telegram chat user can send general chat, public lookup, or browser lookup messages; lookup questions such as `請問義美小泡芙多少錢` receive source-grounded answers through the backend cache and agent lookup flow.

**Depends on:** None

**Mode:** mvp

**Success Criteria**:
1. Telegram webhook path can receive a user message and return a chat reply.
2. Backend uses an LLM intent router to classify messages as `general_chat`, `public_lookup`, or `browser_lookup`.
3. General chat messages receive a normal chatbot answer without website lookup.
4. Public lookup messages use public-source lookup and do not open Chrome MCP by default.
5. Browser lookup messages use Chrome MCP/browser automation through a structured contract.
6. Backend checks a SQLite TTL cache before invoking lookup workers for lookup modes.
7. Lookup results include answer text, source evidence, and query time.
8. Auth-required website responses tell the user/operator that a browser auth session must be prepared.
9. Demo documentation explains how to run the `請問義美小泡芙多少錢` scenario.

**Plans:** TBD after spec and discussion.
