---
title: "Runtime cache uses SQLite; project wiki stores durable knowledge"
date: "2026-05-28"
context: "Exploration of cache design for the website lookup chatbot"
---

# Runtime cache uses SQLite; project wiki stores durable knowledge

## Decision

Use SQLite as the runtime TTL cache for website lookup results. Do not use `.llm-wiki/` as the cache for volatile website data such as prices.

## Runtime Cache

SQLite should store structured lookup results keyed by task type, normalized target, source set, auth profile, and freshness policy.

Example cache dimensions:

- task type: `product_price_lookup`
- target: `義美小泡芙`
- sources: `momo,pchome,coupang`
- auth context: public or named browser profile
- freshness: short TTL for price data

The cache should support stale-if-error behavior: if a fresh lookup fails, the backend may return a stale cached result with clear timestamp labeling.

## Project Wiki Role

`.llm-wiki/` should store durable project knowledge, such as:

- architecture decisions
- supported query types
- source strategy
- browser automation limitations
- demo operation notes

It should not store:

- live price results
- full chat transcripts
- active task state
- credentials or session details

## Rationale

Price and website lookup data expire quickly. SQLite gives the backend predictable TTL behavior and simple operational control. `.llm-wiki/` remains useful as a long-lived knowledge layer without becoming a mutable runtime datastore.

