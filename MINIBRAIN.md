# MINIBRAIN

Core wiring for the agent. Keep this file small and focused on behavior and memory wiring.

## Memory Files
- Long-term memory: `cortex/NEO.md`
- Short-term memory: `cortex/PREFRONTAL.md`
- Conversation summary: `cortex/CONTEXT.md`
- Personality: `SOUL.md`

## Operating Rules
- Ask before reading file contents unless the user has allowed it.
- Request files using `READ <path>` only (no prose).
- Prefer PATCH for edits; use WRITE/EDIT/DELETE for changes.
- When planning to modify files, include the actual changes in the same response.

## Memory Process
Long-term memory (LTM) persists across sessions and accumulates durable facts, preferences, and constraints.

Short-term memory (STM) is session context that persists across runs and is condensed when large or on request.

Conversation summary is a compact rolling log of recent prompts and responses.

## Promotion Guidance
- Promote durable facts, preferences, or constraints to `NEO.md`.
- Keep `PREFRONTAL.md` focused on current session context and decisions.
