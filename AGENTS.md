# AGENTS.md

## Purpose
Experiment with a minimal agentic loop in Go, with a clean path to a future TUI.

## Project Structure
- `README.md`: overview and usage
- `AGENTS.md`: agent instructions and conventions
- `cmd/` (optional): CLI entrypoints if the project grows
- `internal/` (optional): core loop, tools, and prompts

## Build, Test, Run
- Build: `go build ./...`
- Test: `go test ./...`
- Run: `go run ./cmd/minibrain`

## Coding Style
- Keep the core loop small and explicit (goal → action → execute → observe → repeat).
- Prefer standard library; introduce deps only when they clearly simplify the TUI path.
- Keep files small and focused; split helpers instead of “v2” copies.
- Add short comments only for non-obvious logic.

## Safety & Hygiene
- Do not commit secrets or real credentials; use obvious placeholders in examples.
- Avoid editing generated files (if/when they exist).

## Git Workflow
- Use conventional commits (e.g., `feat:`, `fix:`, `chore:`).
- Do not add co-authors to commit messages.
- Keep git simple; group related changes only.
- Before committing: reformat code, ensure consistent style, add tests for new code, run tests, and verify it compiles.
