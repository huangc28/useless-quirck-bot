---
title: "Verify browser agent can look up public commerce prices"
date: "2026-05-28"
priority: "high"
---

# Verify browser agent can look up public commerce prices

## Goal

Confirm that the browser lookup path can retrieve source-grounded product price information from public commerce websites for the demo query `請問義美小泡芙多少錢`.

## Sites To Try

- momo
- PChome
- 酷澎

## Checks

- The agent can search for `義美小泡芙`.
- The agent can extract product title, price, URL, and observed time.
- The agent can distinguish public results from login-required pages.
- The agent can return structured JSON suitable for the Go backend.
- The answer composer can turn the structured results into a short chat reply.

## Success Criteria

- At least one public source returns a usable price result.
- Login-required behavior is detected and reported clearly if encountered.
- The lookup can run from a persistent browser profile without requiring user credentials in chat.

