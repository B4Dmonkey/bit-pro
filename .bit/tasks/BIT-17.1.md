---
id: BIT-17.1
title: A correction lands as a file on disk
status: done
phase: 1
phase_label: A durable record
---
## **Verse 1**

`bp feedback add` writes its first note to disk. The sequence is hardcoded to `001` — that is the
minimum one test can demand, and the second note in the next bar is what forces a real scan.

## Scope
- `task/feedback.go` — new. `feedbackSubdir` const, `feedbackDir()`, `notePath(track string, seq
  int)`, and `AddNote(track, body string) (string, error)`, which `MkdirAll`s `.bit/feedback/`,
  writes the body at `fileMode`, and returns the path. Build the path with `pathologize.Join` the
  way `Store.Path` does — the track ID arrives from a CLI arg, so it is untrusted input to a path.
  Hardcode seq 1 here.
- `cmd/feedback.go` — new. `newFeedbackCmd()`, the `feedback` group with nothing but `add` under
  it, mirroring `cmd/task.go`.
- `cmd/feedback_add.go` — new. `newFeedbackAddCmd()`, mirroring `cmd/task_create.go`: `Use: "add
  <track>"`, `Args: cobra.ExactArgs(1)`, a `-d`/`--description` flag, `task.New(bitDir)`, and the
  result printed with `fmt.Fprintln(cmd.OutOrStdout(), …)`.
- `cmd/feedback_add_test.go` — new.
- `cmd/root.go` — register `newFeedbackCmd()` beside `newTaskCmd()`.

The command prints the note's **path**, not a bare ID. Capture is create-only, so no later command
ever takes a note ID as a handle — the only thing a caller does with the output is tell the user
where the note went, and a path is what answers that.

No frontmatter in the note file. The filename carries the track, the bar is cited in the prose as
data, and the scope rules out a cause field — nothing is left for a header to hold.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestFeedbackAddCmd_WritesFirstNote`
     - **Behavior:** a correction described on the command line becomes a file under
       `.bit/feedback/` whose name keys it to its track — the durable record the Why says repair
       currently destroys.
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Ship the bit plugin", "...")`; then
       `mustRun(t, "feedback", "add", "BIT-1", "-d", note)` where `note` is realistic prose, e.g.
       "Happened at BIT-1.9.\n\nThe plan said: fall back to `plugin install` when `plugin update`
       fails.\nThe work required: deciding whether the fallback also runs `marketplace add`, which
       the plan did not settle."
     - **Assertions:** `os.ReadFile(".bit/feedback/BIT-1-001.md")` returns exactly `note`; the
       command's stdout is `.bit/feedback/BIT-1-001.md\n`.
     - **Boundary:** note count under this track == 1 — the lower bound, and the only count a
       hardcoded sequence can serve.
   - [ ] Confirm fails: `unknown command "feedback" for "bp"`. If it instead fails on an unknown
     `-d` flag then the group registered but the subcommand did not, which is a different problem.

2. **Implement (GREEN):**
   - [ ] `task/feedback.go` with `AddNote` writing seq 1.
   - [ ] `cmd/feedback.go`, `cmd/feedback_add.go`, and the `root.go` registration.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(feedback): write a note to .bit/feedback/`