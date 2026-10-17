---
name: implement-task
description: "Implement one GitHub issue as a single small PR, gated on clean git state, a REVIEW APPROVAL for architectural docs (ADRs/diagrams), a verification subagent, and human review of the prepared commit message before anything is committed. Use this skill when the user says '/implement-task' (with an issue number or link), 'implement task', 'pick up the issue/ticket', or asks to work an issue from this repo's GitHub tracker end-to-end. Do not auto-invoke for anything else."
disable-model-invocation: true
---

# Implement task workflow

You are turning one tracker issue into one small PR whose diff is small enough that the user can review it in minutes. Never commit. Hand control back at every gate below.

## Inputs

The user passes an issue number or URL (e.g. `/implement-task 19` or the GitHub link). If ambiguous, ask.

## Step 1 — Git gate (blocking, do this first)

1. Run `git status --porcelain`.
2. If there are ANY uncommitted changes or untracked files, **halt immediately** and tell the user: commit or stash first. Suggest the repo's commit conventions (`.agents/skills/pr-conventions`) but do not commit for them.
3. If clean, note the current branch, then `git fetch origin`, switch to the default branch (`git symbolic-ref refs/remotes/origin/HEAD`), and `git pull --ff-only`. If there are unpushed commits on the old branch, mention them to the user in one line but continue.
4. Create the work branch: `BOOK-<issue number>` (e.g. `BOOK-19`). If it already exists locally, ask.

Why blocking and first: a dirty worktree makes every later diff unreviewable by the user, which defeats the whole workflow.

## Step 2 — Read the issue

`gh issue view <number>` including comments. State the title and one-sentence summary before starting. If the issue lacks acceptance criteria or they're ambiguous, stop and ask the user — do not invent scope.

## Step 3 — Docs gate (blocking approval, before implementation)

Decide now whether this issue requires **architectural documentation** — an ADR, ER diagram, C4/mermaid diagram, or a security diagram. Signal the record first, then the code: implementation inside an unrecorded design creates a diff and a document that drift at the same time, and reordering them later never happens in practice.

1. **Adjudicate up front** from the issue text and one-line it: state whether a doc is needed and (if so) which kind. When the issue itself was produced by a planning process that recorded decisions (check the epic/parent issues and linked docs), reference or extend those instead of duplicating.
2. If needed: draft the doc(s) at the repo's existing ADR/docs location (see the `domain-modeling` skill for ADR conventions; if the repo has none yet, ask where they go).
3. **Halt and present the draft to the user. Proceed to implementation only after explicit approval.** An agent cannot approve its own design record: ADRs capture human decisions, and a wrong ADR outlives a wrong diff.
4. If no docs are needed, say so in one line and move on.

**Mid-flight rule:** if a doc need only surfaces during implementation (an unforeseen design choice with real trade-offs), stop at that exact point, draft, present, get approval, and only then touch the code.

## Step 4 — Implement against acceptance criteria

- Work through the issue's acceptance criteria as the spec; one PR's worth of scope only. If you discover adjacent work, note it on the issue as a comment listing follow-ups — never widen the branch.
- Follow existing code conventions; no speculative abstractions.
- **Code hygiene (see AGENTS.md "Code style: comments & readability" for the full rules):** no inline comments; self-explanatory names for every variable, method, and type; no god methods (one method = one job); doc comments only where necessary and only expressing the *why* — rationale, constraints, or trade-offs a reader can't infer from the code. If you feel the urge to comment, refactor first (better names, smaller functions); a comment that restates the next line of code must be deleted before handoff.
- Verify continuously: run the repo's health commands from the issue's DoD/verification section (e.g. `make test`, `make lint`). All must pass before you proceed.
- Do NOT editor-commit. (Why: Step 6 handoff is manual review first, and the user only approves a prepared commit message.)

## Step 5 — Verification subagent

Spawn a general subagent to independently verify, not re-implement. Give it: the issue body (AC list), the repo path, and `git diff <default-branch>...HEAD`. Ask it to check, and report as a numbered findings list (each: severity + evidence):

1. Every acceptance criterion demonstrably met in the diff (cite file:line)
2. Nothing implemented beyond issue scope
3. Health commands pass, new/changed behavior covered by tests
4. Convention adherence — including the AGENTS.md code-style rules (no inline comments, self-explanatory names, no god methods) — and no obvious bugs
5. If this branch adds design docs: implementation matches what the doc records (drift either direction is a finding)

Fix findings (if cheap and in-scope); anything arguable goes back to the user instead. Re-run the verification agent once after fixes if it flagged anything material.

## Step 6 — Human review handoff

Present: files changed, a short summary per file, the verification agent's verdict, and this exact commit-message format (filled in, but **not committed**):

```
[BOOK-<nn>] (feat|fix|chore|refactor|docs|test) Short imperative title
- did 1
- did 2
```

Bullet the actually-delivered acceptance criteria, each ≤ one line. Wait for the user's review. On approval, commit on `BOOK-<nn>` only if the user explicitly says so, then **propose running the `create-pr` skill** to finish the PR (that skill handles branch push and PR creation).
