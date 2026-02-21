# minibrain

Minibrain is a minimal AI helper with a distinct personality, built around a small agentic loop in Go.

## Goals
- Keep the core loop tiny and understandable
- Make it easy to iterate on prompting and actions
- Leave a clean path to a TUI later

## Installation
From source (recommended):
```bash
git clone https://github.com/chrishannah/minibrain.git
cd minibrain
go install ./cmd/minibrain
```

Ensure your `GOBIN` (or `GOPATH/bin`) is on your `PATH`:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Or install directly:
```bash
go install github.com/chrishannah/minibrain/cmd/minibrain@latest
```

## Usage (TUI default)
```bash
export OPENAI_API_KEY=...
# optional: export OPENAI_MODEL=gpt-4.1
# optional: export MINIBRAIN_HOME=~/.minibrain
# optional: export MINIBRAIN_ALLOW_READ=1   # allow reading mentioned files by default
# optional: export MINIBRAIN_ALLOW_WRITE=1  # allow applying writes/deletes by default

minibrain
```

## Usage (CLI)
```bash
export OPENAI_API_KEY=...
# optional: export OPENAI_MODEL=gpt-4.1
# optional: export MINIBRAIN_HOME=~/.minibrain
# optional: export MINIBRAIN_ALLOW_READ=1
# optional: export MINIBRAIN_ALLOW_WRITE=1

minibrain -cli "I want you to build X" @docs/plan.md
```

## Brain Structure (User Home)
Default location: `~/.minibrain` (override with `MINIBRAIN_HOME`).

- `MINIBRAIN.md`: core config/initial prompt glue
- `SOUL.md`: personality traits and operating style
- `cortex/NEO.md`: long-term memory (durable facts, constraints, preferences)
- `cortex/PREFRONTAL.md`: short-term memory (session context, condensed when large)
- `cortex/CONTEXT.md`: rolling conversation summary
- `config.json`: user-level config (supports `openai_api_key`, `model`)

On startup, missing files are created automatically. Repo defaults are used only if present; otherwise built-in defaults are used.

## File Reading Approval
File contents are only read when the user approves.
- In TUI: when a prompt includes `@file`, approve with `/yes` (session) or `/always` (persist), or deny with `/no` (session).
- In CLI: set `MINIBRAIN_ALLOW_READ=1` to allow reading.
- Persistent decisions are stored in `.minibrain/config.json` at the project root.

## Write/Delete Confirmation
Writes and deletes require approval unless allowed:
- `/apply` apply changes and allow writes for this session
- `/apply-always` always apply changes (persist)
- `/deny` deny writes for this session
- `/deny-always` always deny changes (persist)
- CLI: set `MINIBRAIN_ALLOW_WRITE=1` to auto-apply
Patches (`PATCH`) follow the same approval flow.

## TUI Commands
- `/help` show commands
- `/clear` clear short-term memory
- `/condense` condense short-term memory
- `/retry` retry last prompt
- `/model` show or set model
- `/usage` show memory and token usage
- `/actions` toggle action log

## TUI Behavior
- Messages are left-aligned; prompts are prefixed with `>` and use a secondary color.
- "Thinking/plan" lines are rendered in a secondary color when detected.
- Conversation text is constrained to ~80% of the width.
- Streaming responses are rendered as they arrive.
- Status bar includes an estimated context token budget.

## Structure
- `cmd/minibrain/`: CLI + TUI entrypoint
- `internal/agent/`: core loop, memory, mentions, writes
- `internal/llm/`: OpenAI API integration

## Behavior (v0)
- Loads long-term memory from `cortex/NEO.md`.
- Loads core config from `MINIBRAIN.md` and personality from `SOUL.md`.
- Provides a relevant file shortlist to the model (based on prompt).
- Includes recent short-term memory context and a rolling conversation summary.
- Loads file contents only when explicitly mentioned and approved.
- Short-term memory persists across runs and is condensed when large or on request.
- Calls OpenAI Responses API, then optionally writes files if the model emits `WRITE`, `EDIT`, `DELETE`, or `PATCH` instructions.

## License
Apache-2.0. See `LICENSE` and `NOTICE`.
