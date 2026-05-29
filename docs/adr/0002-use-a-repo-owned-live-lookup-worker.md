# Use a Repo-Owned Live Lookup Worker

The live lookup path should be implemented as a repo-owned CLI wrapper, not as an ad hoc `codex exec` command embedded directly in `.env`. The Go backend calls one stable worker entrypoint, the worker branches internally by lookup mode, and mode-specific prompts live in reviewable prompt files. This keeps the backend thin while making the live public lookup and browser lookup behavior testable, documented, and adjustable.
