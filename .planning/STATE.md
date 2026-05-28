# Project State

## Current Status

**Current Phase:** 1
**Current Phase Name:** Website Lookup Chatbot MVP
**Status:** Spec in progress

## Accumulated Context

### Roadmap Evolution

- 2026-05-28: Phase 1 created from exploration of the interview prompt. Scope is a Telegram-based website lookup chatbot MVP using SQLite cache and a Chrome MCP/browser lookup agent.

### Decisions

- Website lookup is the product objective; product price lookup is the first demo path.
- Backend stays thin and uses an LLM intent router to choose `general_chat`, `public_lookup`, or `browser_lookup`.
- Chrome MCP/browser automation is used for browser lookup, not for every public lookup.
- Authentication is source-specific fallback, not the default path.
- Runtime cache uses SQLite; `.llm-wiki/` is reserved for durable project knowledge.
