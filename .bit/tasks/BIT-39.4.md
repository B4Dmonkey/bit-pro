---
id: BIT-39.4
title: A punctuated title forces real slugging
status: todo
phase: 2
phase_label: Bar runs unattended
---
## **Verse 2**

A second dispatch test whose track title carries punctuation and mixed case contradicts any
name that could be produced without real slugging — `"Dispatch — the daemon works queued bars
unattended"` cannot be reached by returning the previous bar's `"ACME-1-a-track"`.

## Scope
- `claude/dispatch.go` — `slug` handles runs of non-alphanumerics, including multi-byte ones.
- `claude/dispatch_test.go` — new file: the derivation table, tested directly now that a
  `Tick`-level assertion has forced the function into existence.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestWorktreeName` in `claude/dispatch_test.go`, table-driven
     - **Behavior:** the derived identifier is a single safe token — usable as a git branch
       suffix, a directory name under `.claude/worktrees/`, and a `claude agents` row name — for
       any track title an operator actually writes.
     - **Setup:** cases as `{trackID, title, want}`:
       `{"ACME-1", "a track", "ACME-1-a-track"}`;
       `{"BIT-39", "Dispatch — the daemon works queued bars unattended", "BIT-39-dispatch-the-daemon-works-queued-bars-unattended"}`;
       `{"BIT-7", "bp init registers the MCP server", "BIT-7-bp-init-registers-the-mcp-server"}`;
       `{"BIT-8", "  spaced  out  ", "BIT-8-spaced-out"}`;
       `{"BIT-9", "slash/and.dot", "BIT-9-slash-and-dot"}`.
     - **Assertions:** `WorktreeName(trackID, title) == want` for every case.
     - **Boundary:** the em-dash case is a **multi-byte non-alphanumeric flanked by spaces** — the
       upper end of the "run of separators" range, where a naive per-byte replace yields
       `dispatch----the` rather than `dispatch-the`. The `"  spaced  out  "` case is the other
       bound: separators at both edges, which must be trimmed rather than kept.
   - [ ] Confirm fails: the em-dash and edge-space cases return extra `-` runs. The first case
     already passes — that is expected, it is the one the previous bar's assertion pinned.

2. **Implement (GREEN):**
   - [ ] Rewrite `slug` to scan runes: append `unicode.ToLower(r)` when
     `unicode.IsLetter(r) || unicode.IsDigit(r)`, otherwise remember that a separator is pending
     and emit a single `-` only before the next kept rune. Emitting on the next kept rune rather
     than on the separator is what makes both collapsing and trailing-trim fall out with no
     `strings.Trim` pass.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic. The table is the whole check.

## Commit (user)
`feat(claude): collapse separator runs in the worktree name`