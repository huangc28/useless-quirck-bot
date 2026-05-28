# Requirements

## REQ-001: Website Lookup Chatbot MVP

The system must provide a chat-based interface that accepts natural-language requests to look up information from websites and returns a concise, source-grounded answer in the same chat.

### Scope

- The chat platform may be Telegram, Line, Google Chat, or another webhook-capable chat system.
- The initial implementation should use one chat platform for the demo, with Telegram as the preferred low-friction option.
- The backend should receive chat messages, classify the intended response mode, orchestrate lookup execution when needed, and return answers.
- The demo path must support the sample query: `請問義美小泡芙多少錢`.
- Price lookup is the first supported task type, but the architecture should not be hard-coded as a price-only bot.
- The backend should not maintain complex source registry or site-specific scraping logic in the MVP.
- An LLM intent router should classify incoming messages into `general_chat`, `public_lookup`, or `browser_lookup`.
- Website search behavior should be delegated to a Chrome MCP/browser automation lookup agent.

### Expected Behavior

- General chat messages should receive normal chatbot replies without website lookup.
- If the user asks for current public information without specifying a website, the system should use public lookup and summarize source-grounded findings.
- If the user specifies a website, such as `momo 義美小泡芙多少錢` or provides a URL, the system should use browser lookup and preserve that instruction.
- Answers should include the retrieved facts, source names or links, and query time when available.
- If a target site requires login, the system should ask for the browser profile/session to be prepared instead of treating login as the default path.
- Runtime lookup results should be cached with a freshness policy to improve response time.

### Non-Goals For MVP

- Full open-ended web browsing across arbitrary sites.
- Multi-user credential management for third-party shopping sites.
- Production-grade crawling, anti-bot bypassing, or background indexing.
- Using `.llm-wiki/` as a runtime cache for volatile lookup results.
- Building dedicated scrapers for momo, PChome, or other commerce sites during the first MVP.
- Hard-coded keyword-only intent routing as the main classification mechanism.
