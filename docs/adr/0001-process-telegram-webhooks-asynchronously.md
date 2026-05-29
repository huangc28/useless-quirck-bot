# Process Telegram Webhooks Asynchronously

For the live demo, Telegram webhook handling should acknowledge valid updates quickly and continue router/lookup/reply work in the background. Browser lookup can be slow because it may involve Codex, Chrome DevTools MCP, and external websites; keeping all work inside the webhook HTTP request risks Telegram retries and duplicate user-visible behavior. This MVP uses lightweight in-process async handling plus in-memory recent `update_id` deduplication rather than a durable queue.
