---
id: BIT-3.9
title: The import proves the model
status: done
phase: 4
phase_label: The project's own plans are in the project
---
## Step 9 (Phase 4 — The project's own plans are in the project) — The import proves the model

The acceptance test. No code — this is the scope's finish line, run against real,
backtick-heavy content rather than a fixture. Driven by the prompt in the scope's
**The import prompt (Phase 4)** section; execute that prompt rather than paraphrasing it.

This work imports itself. `plan-hierarchy-scope.md` becomes BIT-3 and this plan becomes
its nine bars — including this step, which is the one running the import. Its status is
whatever it is at the moment you write it, and that's fine: the record is a snapshot, not
a live view. It's also the sharpest test of the model, because BIT-3 is the only track
whose plan was written after the dotted-ID design existed.

**Scope:**
- `.bit/tasks/` — new records only

**Checklist:**

- [ ] `just build` first — the import drives `./bin/bit`, not `go run`, and `bit` is not
      on PATH
- [ ] Import in delivery order: `cli-bootstrap-plan.md` under BIT-1, `task-crud-plan.md`
      under BIT-2, then `plan-hierarchy-scope.md` as a new track (BIT-3, via a plain
      `task create` — it has no parent) with `plan-hierarchy-plan.md`'s nine steps as its
      bars
- [ ] Fold each plan's preamble (Context, How this plan works) into its parent track's body
- [ ] One child per `## Step N`, in order, so dotted IDs match step numbers; title from
      the step name, `--phase` / `--phase-label` from the scope phase it's tagged to,
      status from its `**Status:**` line, body verbatim
- [ ] Verify by reading each back and diffing against the source — not by eyeballing.
      `$(cat file)` strips trailing newlines; use the sentinel
      (`body=$(cat f; printf 'X'); body=${body%X}`)

**Claude verifies:**
- [ ] `go test -count=1 ./...` — `-count=1` matters here: the cache doesn't notice that
      `.bit/` changed on disk
- [ ] `just lint`
- [ ] `./bin/bit task list` shows BIT-3, BIT-2, and BIT-1 each heading their own bars,
      grouped and phase-labelled
- [ ] `ls .bit/tasks/ | grep -c '^BIT-2\.'` returns 13 and `'^BIT-3\.'` returns 9 — the
      counts are checkable, so check them rather than eyeballing the listing

**User verifies:**
- [ ] `./bin/bit task read BIT-2` — does folding the plan preamble into the track body
      muddy the approve/disapprove view? The scope names this as the open question, and
      this is the read that answers it. If it reads badly, the preamble is droppable.
- [ ] `phase` + `phase_label` earn their place against 13 real bars, or one of them doesn't

**Commit (user):** `feat(bit): import the project's own plans as tasks`