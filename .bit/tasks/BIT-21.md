---
id: BIT-21
title: Task IDs are case-sensitive, silently corrupting .bit/ state
status: todo
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
in use in the conversation, which is how IDs normally travel — the skills are this CLI's
primary caller, not a human at a keyboard, so wrong-case input is the routine path rather than
a rare slip.

## Summary

Make a task ID's case irrelevant to correctness. Resolve an incoming ID against the tasks that
actually exist, so every downstream lookup, guard, and file write uses the spelling the task
file already carries. Then close the two write paths that trusted their caller's spelling —
the feedback note write and the next-ID scan — so neither can destroy or duplicate a record
even if a non-canonical ID ever reaches them again.

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
        │
        ▼
  two tasks answer to BIT-20 · the old track's orphaned bars and
  feedback notes silently re-attach to the unrelated new one
```

Observed, not theorised — same unfinished bar, only the argument's case differs:

```
$ bp task complete BIT-1   → Error: cannot relocate: unfinished bars BIT-1.1   (exit 1)
$ bp task complete bit-1   →                                                   (exit 0)
```

## Decisions

- **A wrong-case ID is accepted, not rejected.** `bp task complete bit-20` finds `BIT-20` and
  does the work. Case carries no meaning in these IDs, so refusing one would spend an error on
  a difference that isn't ambiguous, and the caller getting it wrong is an agent that retries
  rather than a human who learns. Genuine typos still fail: an ID matching no existing task is
  an error either way.
- **Canonical means the spelling on the task file — never a transform.** There is no rule that
  produces the right case. `bp init` stores whatever prefix is typed (`cmd/init.go` trims
  whitespace and nothing else), and a sibling project runs on a lowercase prefix with IDs like
  `foo-6.1`. Uppercasing would corrupt that project exactly the way lowercasing corrupted this
  repo. So an incoming ID is matched case-insensitively against the tasks that exist, and the
  matched file's own ID is what every later step uses.
- **The root cause is one comparison, not three bugs.** IDs are read canonically from each
  file's YAML frontmatter but compared against the argument as typed, case-sensitively. Every
  symptom traces back to that mismatch.
- **The fix lives in `task/`, not at the CLI boundary.** Seven commands take an ID (`read`,
  `update`, `complete`, `delete`, `move`, `feedback add`, and `create --parent`) and every one
  hands it to a `Store` method, so the boundary is seven places and a convention each future
  command has to remember — and forgetting it fails silently, which is the failure this track
  exists to kill. `Store` is the one choke point they all pass through, and `Path` is the
  single function that turns an ID into a filename.
- **The two damaged write paths get fixed independently of the resolve.** A note write that
  overwrites, and an ID scan that trusts filename case, are defects on their own terms — the
  CLI documents create-only notes and permanent ID reservation, and delivers neither. Fixing
  only the lookup would leave both promises still untrue for any other bad input.
- **No recovery or repair work is in scope.** The real `.bit/` was probed for wrong-case
  filenames, duplicate IDs, and orphaned bars, and is clean — the one damaged track was
  repaired by hand when the bug was found. Nothing to migrate, so no doctor command.
- **Case-insensitive filesystems are the dangerous environment, and the target one.** macOS
  and Windows silently merge `bit-1.md` and `BIT-1.md` into one file, which is what turns the
  bug into data loss. On Linux the same command fails differently. The fix must be correct on
  both, and the tests must not assume either.

## Verses

- [ ] Verse 1 — An operator who names a task in any case gets the CLI's documented behaviour
  rather than silent corruption: the unfinished-bars guard holds, and any file the command
  writes lands under the task's real name.
  Touches: the lookup and path helpers in `task/store.go` (`Load`, `children`, `Path`) — where
  to look to verify.

- [ ] Verse 2 — Recording a feedback note can never destroy one already recorded, whatever ID
  reached it. The guarantee the CLI already advertises becomes true, which matters because
  capture fires right after something has gone wrong and the note is unreproducible.
  Touches: the note write in `task/` (`AddNote`) and its sequence numbering.

- [ ] Verse 3 — An ID that has been reserved by completing or archiving a task can never be
  handed to a new one, so a task's name stays a stable reference for commit messages, feedback
  notes, and older scopes that cite it.
  Touches: the next-ID scan in `task/` (`NextID` / `NextChildID`).