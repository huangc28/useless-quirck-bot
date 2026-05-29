# Interview Chatbot

A Telegram-based website lookup chatbot. The domain centers on routing chat messages into ordinary chat, public-source lookup, or browser-backed website lookup, then returning source-grounded answers.

## Language

**Live Lookup Worker**:
A repo-owned CLI wrapper that receives one lookup request JSON object from the Go backend and returns one strict lookup response JSON object. The backend calls one stable worker entrypoint; the worker branches internally based on whether the request is public lookup or browser lookup.
_Avoid_: ad hoc lookup command, raw Codex command, separate backend worker commands per lookup mode

**Public Lookup**:
A lookup for current public information when the user does not require a specific website or URL. Public lookup must use source evidence and must not open browser automation by default.
_Avoid_: general search, model-memory answer, browser-backed lookup

**Codex-Backed Public Lookup**:
A public lookup performed by the Live Lookup Worker through Codex CLI, constrained to source-grounded answers. It is acceptable for the MVP only when unsupported answers become no-result responses rather than guesses.
_Avoid_: deterministic search API, hard-coded scraper, model-memory price

**Browser Lookup**:
A lookup for a specific website, platform, or URL that may require browser automation or a prepared browser session. Browser lookup is only selected when the user names a site, platform, or URL.
_Avoid_: public lookup, arbitrary browsing

**Browser Session**:
The browser state used by browser lookup to inspect a named website or URL. A browser session may already exist and should be reused when available; otherwise one can be prepared before the lookup continues.
_Avoid_: per-request fresh browser, chat-provided credentials

**Auth-Required Response**:
A lookup outcome indicating the selected source needs a prepared browser session. It is not a request for the chat user to send third-party credentials.
_Avoid_: ask for password, login flow

**Source Evidence**:
The proof attached to a successful lookup answer. It must include a source name, URL, observed time, and enough page context to explain where the answer came from.
_Avoid_: unsupported answer, model guess, source-free price

**No-Result Response**:
A lookup outcome meaning the worker could not find reliable source evidence for the requested fact. It is preferred over guessing when the requested information is volatile, such as product price.
_Avoid_: guessed answer, unsupported fallback

**Live Lookup Failure**:
A failure of the live lookup path whose cause is unknown or operational, such as invalid worker output, tool failure, or timeout. It must be reported as a temporary lookup failure, not replaced with mock success.
_Avoid_: fake success, silent mock fallback

## Example Dialogue

Developer: "The user asks `請問義美小泡芙多少錢`; is that browser lookup?"

Domain expert: "No. That is public lookup unless they name a site or URL."

Developer: "The user asks `請問 momo 義美小泡芙多少錢`; should the Live Lookup Worker use browser tooling?"

Domain expert: "Yes. That is browser lookup because the user named momo."
