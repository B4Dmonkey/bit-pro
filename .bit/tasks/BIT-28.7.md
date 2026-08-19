---
id: BIT-28.7
title: The disabled store makes stopped a third state
status: todo
approved: true
phase: 2
phase_label: lifecycle
---
## **Verse 2**

A label in launchd's disabled store reports `stopped`, even though `launchctl list` would say the
same thing it says for a job that simply is not running. Contradicts BIT-28.6, which collapses the
two — and collapsing them hides the one fact an operator needs when queued bars are not moving.

## Scope
- `launchd/status.go` — read `launchctl print-disabled gui/$UID` before consulting `list`
- `launchd/status_test.go`, `cmd/status_test.go` — the new tests

## References
- `automation-notes.md` (repo root, untracked) — "Daemons on macOS" for why `disable` is the durable
  half of a stop and `print-disabled` is how the store is read back. The parse rule itself is a
  Decision on this track's scope and supersedes the note's open question.

## TDD cycle

1. **Write test (RED):** `cmd/status_test.go`
   - [ ] `TestStatusCmd_ReportsStoppedFromTheDisabledStore` (table-driven)
     - **Behavior:** a deliberate stop is durable and survives a reboot, so "the operator stopped
       this and it will not come back at login" is a different answer from "it happens not to be
       running right now".
     - **Setup:** extend the fake `launchd.Runner` to answer two subcommands — `print-disabled gui/<uid>`
       and `list <label>` — where `<uid>` is `os.Getuid()`. Three rows, all with `list` returning a
       dict containing `"PID" = 4242;` and code `0`, varying only the store: (a) the store contains
       `"com.github.b4dmonkey.bit-pro" => disabled`; (b) it contains
       `"com.github.b4dmonkey.bit-pro" => enabled`; (c) it contains other labels but not this one.
     - **Assertions:** (a) output is exactly `"stopped\n"` and `list` is never called; (b) and (c)
       are exactly `"running (pid 4242)\n"`. The recorded `print-disabled` call names
       `gui/` + `os.Getuid()`.
     - **Boundary:** all three shapes the label's entry can take in the store — absent, `=> enabled`,
       `=> disabled`. This is the measured trap: `launchctl enable` flips the entry to `=> enabled`
       rather than removing it, so the label stays in the store forever after the first `bp stop`,
       and matching on the key's presence would report `stopped` for the rest of the machine's life.
       Row (b) is the row that fails if the implementation matches on the key.
   - [ ] Confirm fails: row (a) fails with output `"running (pid 4242)\n"`, want `"stopped\n"` —
         BIT-28.6 never consults the disabled store at all.

2. **Implement (GREEN):**
   - [ ] `launchd/status.go`: before the `list` call, run
         `run(ctx, "launchctl", "print-disabled", "gui/"+strconv.Itoa(os.Getuid()))`. Propagate a
         non-nil `err`; ignore a non-zero `code` and fall through (an empty store is not a failure).
   - [ ] `launchd/status.go`: match the output against a package-level regexp built as
         `regexp.QuoteMeta` of a quoted `Label` followed by the Go raw string `\s*=>\s*disabled`.
         `QuoteMeta` is what keeps the label's dots from matching any character. On a match return
         `StateStopped, 0, nil` without calling `list`.
   - [ ] `launchd/status.go`: leave the existing `list` path untouched for every other case.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic. The real store is exercised on BIT-28.12.

## Commit (user)
`feat(status): report stopped from launchd's disabled store`