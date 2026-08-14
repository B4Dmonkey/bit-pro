---
id: BIT-21
title: Task IDs are case-sensitive, silently corrupting .bit/ state
status: done
---
## Why

A task ID typed in the wrong case silently corrupts `.bit/` state instead of being rejected.
`bp instructions` makes three guarantees to anyone driving the CLI — `complete` "refuses a
track that still has an unfinished bar, and there is no override"; a feedback note "can never
damage one already recorded"; and a relocated task's ID "is never re-minted onto a different
task." All three were observed breaking from a single lowercase argument, with **exit code 0
and no output**. The operator's only signal that anything went wrong is noticing the damage
later by eye.

This is worse than an ordinary typo bug because the failure is silent and the artifacts it
damages are the ones that cannot be reconstructed: a feedback note captures a moment that has
passed, and a re-minted ID makes two different tasks answer to the same name across commit
messages, notes, and every prior scope that cited it.

It was hit for real while completing BIT-20 in this repo: an agent ran `bp task complete
bit-20`, and the track filed itself as `.bit/completed/bit-20.md` while all ten of its bars
stayed behind in `.bit/tasks/`. The wrong case came from the agent echoing the lowercase form
in use in the conversation, which is how IDs normally travel.

A fourth symptom was measured while planning this track, and it is the most destructive of
them: `bp task create --parent bit-1` overwrote the existing bar `BIT-1.1`, leaving the file
under its original name with a different task's contents inside it. The bar was not moved or
renamed — it was destroyed, exit code 0, no output.

## Summary

Make uppercase the one spelling a task ID ever has. Normalize the existing projects on disk
once, then normalize at every boundary the CLI reads from — the ID a caller passes, the ID
stored in a task's frontmatter, and the prefix in `config.toml` — so no later input, hand-edit,
or freshly-initialised project can put a lowercase ID back into circulation.

## Visual aid

One lowercase argument, three broken guarantees:

```
bp task complete bit-20
        │
        │  children("bit-20") compares "bit-20" against "BIT-20"
        │  read from each file's frontmatter → case-sensitive == → 0 matches
        ▼
  ┌─────────────────────────────────────────────────────────────┐
  │ 1. GUARD BYPASSED    unfinished-bars check reads that empty  │
  │                      child list → nothing to object to       │
  │                      → files an incomplete track, exit 0     │
  │                                                              │
  │ 2. FILE MISNAMED     write path built from the raw argument  │
  │                      → completed/bit-20.md, not BIT-20.md    │
  │                                                              │
  │ 3. ID RE-MINTED      next-ID scan doesn't recognise          │
  │                      bit-20.md as BIT-20 → hands BIT-20 to   │
  │                      the next new task                       │
  └─────────────────────────────────────────────────────────────┘
```

Observed, not theorised — same unfinished bar, only the argument's case differs:

```
$ bp task complete BIT-1   → Error: cannot relocate: unfinished bars BIT-1.1   (exit 1)
$ bp task complete bit-1   →                                                   (exit 0)
```

And the fourth symptom, measured the same way. A case-insensitive filesystem resolves the
wrong-case write onto the existing file and keeps that file's original name, so the damage
leaves no trace in a directory listing:

```
$ bp task create "sneaky" --parent bit-1     → bit-1.1     (exit 0)
$ ls .bit/tasks/                             → BIT-1.md  BIT-1.1.md
$ head -3 .bit/tasks/BIT-1.1.md              → id: bit-1.1
                                                title: sneaky
```

## Decisions

- **Uppercase is the canonical spelling, and it is a transform, not a lookup.** Every ID is
  uppercase everywhere it is stored: filename, `id:` frontmatter, `order:` entries, and the
  `config.toml` prefix. A caller may type any case; the tool uppercases it and proceeds. This
  reverses an earlier decision that canonical meant "whatever spelling the file already
  carries" — a transform is simpler, has no ambiguity to resolve, and needs no directory scan.

- **Uppercase over lowercase, knowing which project pays.** Measured: this repo and the
  marketplace clone are uppercase; the `evus` client project is lowercase (105 files, zero
  uppercase `id:` fields). Lowercase would have left `evus` untouched and migrated bit-pro;
  uppercase does the reverse. Uppercase was chosen deliberately with that cost accepted, since
  the operator runs all three projects and the migration is a one-time task.

- **Normalize on read as well as on write.** Writing uppercase is not enough on its own — a
  hand-edited `config.toml` or task file would put a lowercase ID straight back into
  circulation. So the prefix is uppercased both when `init` stores it and when it is read
  back, and IDs are uppercased on the way in rather than trusted.

- **`bp init` normalizes the prefix.** A user typing `foo` gets `FOO` stored. Without this, a
  project created after the migration would be born lowercase and reproduce the bug in a
  project no migration script ever saw.

- **Migration is a one-time operator task, not a product feature.** A committed `update/`
  directory holds a bash script plus instructions for Claude; the operator names the
  directories to normalize and runs it. No `doctor` command, no auto-repair on startup, no
  detection of un-migrated projects — the operator is the only person running this tool and
  knows which three directories exist.

- **The migration covers five ID carriers, not four.** Filename, `id:` frontmatter, `order:`
  lists, the `config.toml` prefix, and — added after measurement — **feedback note filenames**
  (`.bit/feedback/<TRACK>-NNN.md`). The fifth was nearly missed: with the first four migrated
  and `feedback/` left alone, `bp feedback add BIT-1` was observed overwriting the contents of
  an existing `bit-1-001.md` while leaving its filename untouched, destroying the note. A
  four-carrier migration would therefore have *created* the data-loss condition this track
  exists to prevent. Body prose and git commit messages that cite `BIT-11.4` are deliberately
  left alone: nothing in the code parses them, so a stale citation is cosmetic.

- **Case-only renames go through `git mv --force`.** Measured: a plain two-step `mv`
  (`bit-1.md` → tmp → `BIT-1.md`) leaves `git status --porcelain` completely empty on a
  case-insensitive filesystem — git never sees the rename, so the change cannot be committed or
  reviewed. `git mv --force` reports `R  bit-1.md -> BIT-1.md`. The script uses it wherever the
  target is a git working tree.

- **The migration script is invoked on project roots, and is itself tested in bash.**
  `update/normalize.sh <project-dir>...` takes the directories that *contain* `.bit/`, not the
  `.bit/` paths themselves, so the operator never types the implementation detail; a directory
  with no `.bit/` is an error rather than a silent skip. `update/README.md` holds the
  Claude-facing instructions and is read on demand rather than auto-loaded. `update/normalize_test.sh`
  is the harness — bash, not Go, keeping the one-time migration self-contained and uncoupled
  from the product's test suite.

- **Migration lands before the code change.** Once every project is uppercase, the *existing*
  case-sensitive code is already correct — so the migration leaves every project working
  whether or not the code change has shipped. The code change is then purely about preventing
  recurrence.

- **The fix lives in `task/`, not at the CLI boundary.** Seven commands take an ID and every
  one hands it to a `Store` method. `Store` is the one choke point they all pass through, and
  `Path` is the single function that turns an ID into a filename — normalizing there covers
  every command, including ones not yet written. The corollary is the part worth stating, because
  planning first got it wrong: a command has to *go through* `Store` rather than filter
  `List()` output itself. Measured — `bp task list --parent` compared the raw flag value
  against `t.ID` inside `cmd/task_list.go`, never reaching disk and so never reaching a
  normalizing path, and so returned zero bars with exit 0 for `--parent bit-21`. It is the
  only such filter in `cmd/` and `tui/`, and it closes by exporting a `Store.Children(parent)`
  wrapper over the private `children` — which already normalizes — not by adding ID handling
  to `cmd/`.

- **Normalization is worth building even though Claude drives the CLI.** The original hold on
  this track asked whether IDs arriving via tool calling rather than hand-composed shell
  commands would remove the exposure. Settled: normalize anyway. It is a few lines, and it
  covers the case the tool-calling argument does not — a human hand-editing `.bit/` or typing
  a command directly.

- **A wrong-case ID is accepted, not rejected.** `bp task complete bit-20` finds `BIT-20` and
  does the work. Case carries no meaning in these IDs, so refusing one would spend an error on
  a difference that is not ambiguous. Genuine typos still fail: an ID matching no existing task
  is an error either way.

## Verses

- [x] Verse 1 — The operator can point a script at any set of `.bit/` project roots and get
  them back fully uppercase and internally consistent, so every project on this machine agrees
  on one spelling. Delivered on its own: after this lands, the existing code is correct for
  every project even before Verse 2 exists.
  Touches: a new `update/` directory at the repo root — `normalize.sh`, its bash harness
  `normalize_test.sh`, and `README.md` for Claude — where to look to verify.

- [x] Verse 2 — A wrong-case ID can no longer put lowercase state back into a normalized
  project, whoever or whatever typed it: IDs are uppercased on the way in, the `config.toml`
  prefix is uppercased both when stored and when read, and a newly initialised project is born
  uppercase. The four symptoms above stop being reachable and the guarantees in `bp
  instructions` become true again.
  Touches: the ID and path helpers in `task/store.go` (`Path`, `Load`, `Save`, `children`,
  `NextID` / `NextChildID`, the order helpers), the note write in `task/feedback.go`, the
  prefix in `task/config.go`, and `cmd/init.go` — where to look to verify.