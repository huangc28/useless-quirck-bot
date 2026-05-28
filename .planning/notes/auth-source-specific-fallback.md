---
title: "Auth is a source-specific fallback"
date: "2026-05-28"
context: "Exploration of the take-home interview chatbot requirement"
---

# Auth is a source-specific fallback

## Decision

Website authentication is not part of the default user flow.

The normal flow should query public website data and return results directly to the chat. Authentication is only needed when the selected source blocks the requested information behind a login session.

## Interpretation

The prompt says the system may let a website log in first. The best interpretation is:

- A browser automation profile can be prepared with an authenticated session.
- The lookup agent can reuse that session when querying sites that require login.
- The Telegram, Line, or Google Chat user should not be asked to send third-party website credentials through chat.

## Demo Behavior

- `請問義美小泡芙多少錢` should query public sources without auth.
- `請問 momo 義美小泡芙多少錢` should query momo directly if public results are available.
- If a selected site requires login, the bot should explain that the site needs an authenticated browser session and ask the operator to prepare it.

## Rationale

This keeps the MVP reliable and avoids unsafe credential handling while still addressing the prompt's hint about pre-authenticated websites.

