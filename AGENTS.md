# AGENTS.md

Bookie: a gRPC service (book server) with a Go BFF HTTP client, being migrated toward MySQL persistence. Workspace layout and commands are discoverable (`makefile`, `src/`, `protos/`); run `make --help` style targets rather than guessing.

## Task tracker (source of truth: GitHub Issues)

- One issue = one actionable task = one small PR. Epics group related tickets and carry the decision record; check the epic/parent and linked docs before implementing — decisions live in the epic first, code follows.
- Issues include a WHAT / WHY / DoD with acceptance criteria. Implement against the ACs; do not invent scope. Discover adjacent work → comment it on the issue as follow-up tickets, never widen the branch.
- Branch naming: `BOOK-<issue number>` (e.g. `BOOK-19`). Commit messages: `[BOOK-<nn>] (feat|fix|chore|refactor|docs|test) Title` with did-bullets.

## Implementing an issue

Prefer the `implement-task` skill (`.agents/skills/implement-task`) — it runs the full gated workflow (git gate → docs-approval gate → verify subagent → human review). Even without invoking it, these gates are **repo invariants**:

1. **Clean tree first.** Uncommitted or untracked changes → halt, ask the human.
2. **Docs before code.** If an ADR, ER/C4, or security diagram is warranted by the task, draft it, present, and get explicit human approval before implementation. See the `domain-modeling` skill for ADR conventions.
3. **Never commit without the human's review.** Prepare the commit message, show it, wait for explicit go. Committing or pushing on your own initiative is the standing failure this rule closes.
4. **Verify before handoff.** `make test` and `make lint` (or the issue's DoD commands) must pass; run a verification pass over the diff against the issue's ACs before presenting.

PR conventions (titles, bodies): `.agents/skills/pr-conventions`. Finish with that skill.

## Code style: comments & readability

Write code that speaks for itself. Comments are a last resort, not a habit.

- **Avoid inline comments.** If code needs a `// explain this` line, the fix is usually better naming, smaller functions, or clearer structure — apply that first. Do not narrate *what* the code does; a competent Go developer can already read it.
- **Self-explanatory names, always.** Variables, methods, types, and fixtures must say what they mean in their name (`uuid.NewV7().String() > idStr = gen2()`). No cryptic abbreviations, no names that force the reader to check the type.
- **No god methods.** One method = one job. Anything that needs paragraph comments to explain its steps is doing too much — split it or simplify it.
- **Docblocks: only what is necessary.** Per Go convention, every exported symbol carries a doc comment — but other than that, comment only what a reader *cannot* infer from the code: rationale ("this must stay frozen for test determinism"), constraints, protocol/contract facts, and non-obvious trade-offs. If a comment restates the next line of code, delete it. Think like a senior SWE: assume the reader is skilled, explain only the why, and prefer refactoring over explaining.
- Rule of thumb: the *what* lives in the code; the *why* is the only part allowed in a comment — and even then, only when the why isn't obvious from context.

## Codebase decisions (don't re-litigate; change via ADR)

- **UUIDs:** RFC 9562 UUIDv7 via Go 1.27 **stdlib** `uuid` package (import `uuid`), not `github.com/google/uuid`. IDs are generated in Go and travel as canonical lowercase strings; MySQL will store them as `BINARY(16)` in RFC byte order (repo decision in epic #27).
- **Data layer:** sqlc (plain SQL codegen, no ORM runtime), goose migrations, repository interface seam around the books store.
- **Mocks:** mockgen via `go generate` / `make gen/mocks`.
