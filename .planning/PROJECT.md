# Interview Chatbot

## Purpose

Build a take-home demo chatbot that can receive a natural-language request in a chat window, look up information from websites, and return a concise answer with sources.

## Current Direction

The first demo target is the interview prompt example: `請問義美小泡芙多少錢`.

The system should not be limited to price lookup as a product concept. Price lookup is the first proof path for a broader website lookup chatbot.

## Key Decisions

- Use one chat platform for the MVP, with Telegram preferred for low-friction webhook setup.
- Keep the backend thin: receive messages, decide whether a website lookup is needed, check cache, call a lookup agent, and return the response.
- Use a Chrome MCP/browser automation agent for website lookup instead of hard-coding complex source-specific logic in the backend.
- Treat authentication as a fallback when a website blocks requested data behind login.
- Use SQLite for runtime TTL cache.
- Use `.llm-wiki/` only for durable project knowledge, not volatile lookup results.

