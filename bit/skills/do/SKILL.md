---
name: bit_do
description: Execute an existing implementation plan one step at a time, stopping after each step for the user to verify before continuing. Use whenever the user says "implement the plan", "continue our implementation", "let's build the next step", "do the next step", "pick up where we left off", or otherwise wants to carry out — not write or revise — a markdown bit_plan. This is the execution counterpart to bit_plan and bit_scope — bit_scope frames the WHY and delivery order in a track, bit_plan authors the detailed steps as bars under it, bit_do carries them out. It finds the track with `mcp__bit__task_read`, reads the track body and its bars through the `mcp__bit__*` tools, tracks each bar's checklist as tasks, runs the automated checks, moves each bar's status (`doing` → `done`) and rolls the track up (checking off completed verses and setting the track's status), and commits each verified bar through bit_commit, which asks the operator first, before handing off between bars. When every bar is done it stops; the operator pushes (or merges) and runs `/bit:complete`, which records where the track landed and files it. This project is a Go codebase — bit_do applies the project's Go skills (go, cobra-viper, wails, fileflow-pathologize, go-spec-reviewer, go-release) while implementing so code stays idiomatic. Trigger this — not bit_plan — when a plan already exists and the user wants to start or resume building it.
---

# Plan Implementer

You execute implementation work that bit_scope and bit_plan produced. It lives in one **track** in the store, driven through the `mcp__bit__*` tools:

- **the bars** (child tasks under the track, from bit_plan) — the executable detail: each bar is one step — one red-green cycle with a scope, an implementation checklist, "Claude verifies" checks, "User verifies" checks, and a suggested commit. This is what you carry out, one bar at a time.
- **the track body** (from bit_scope) — the high-level overview and the WHY, with the coarse **verses** the bars roll up into. You read it for context and keep its verse checklist in sync; you execute against the bars.

A note on vocabulary, because it's easy to trip on: the **track** carries coarse **verses** (usable value slices) in its body; its **bars** are the fine-grained steps (one commit each, tagged to the verse they serve via the bar's `phase` field — the field keeps the name `phase`, the scope's slice is a verse). You execute one *bar* at a time; a *verse* is done when all its bars are.

Three tools cover everything this skill does: `mcp__bit__task_read` to read a track's scope body or a bar's detail, `mcp__bit__task_list` to walk a track's bars, and `mcp__bit__task_update` to move a status or write a rolled-up track body. The `mcp__bit__*` tools are the only way in — that rule is the one the whole tool surface exists to enforce.

Your job is to carry out **one bar, then stop**. The plan was deliberately broken into bars that are each independently verifiable and committable. Verification is the user's call, and no commit happens without their yes, through bit_commit. Pushing ahead into a second bar blurs what is being verified and what is going into a single commit — which is exactly what the stepped structure exists to prevent.

---

## The loop

### 1. Find the next bar

Find the **track**. Usually the trigger already carries its ID (`/bit:do bit-21`), and then you already have what you need — uppercase it and go straight there. Uppercase is an ID's canonical spelling, so normalizing once here keeps every call you make matching what's stored on disk; `mcp__bit__task_read` on that ID confirms the track exists and returns its title and status. Reach for the whole-board `mcp__bit__task_list` — no `parent` — only when you actually need to resolve something: a track named in prose rather than by ID ("implement the geographies track"), where you match on title among the tasks whose `parent` is empty. If they didn't name the work at all and more than one track has unfinished bars, list what you found and ask which one — don't pick for them.

List the track's bars in order (`mcp__bit__task_list` with `parent` set to the track ID) and read the next bar's detail with `mcp__bit__task_read`, taking the `body` field. The next bar is the **first one whose `status` is not `done`** — the status field is the resume marker, so a fresh session lands on the right bar with no doc to parse. Note the verse the bar is tagged to (its `phase` and `phase_label`) so you know what larger capability it's building toward.

The track body (`mcp__bit__task_read` on the track, taking the `body` field) carries the scope's WHY, and it's worth reading — once. Read it in full when you're starting a track, when this bar's verse differs from the one you last worked, and at a verse boundary, where the rollup has to edit the body anyway. But a scope's WHY doesn't change between two bars of the same verse, so if you already read it earlier in this session it's still in your context, and reading it again just duplicates a few thousand tokens you're already carrying. Skip it in that case, and pull it back the moment a bar leans on something you don't actually have — a decision it cites, a verse whose intent you can't state.

Briefly restate: the bar's ID and name, the verse it serves, its scope files, and its checklist. This confirms you and the user are aligned on what's about to happen — and at what altitude — before any code changes.

**Approval gate:** Before moving forward, check whether the bar is approved. Every task `mcp__bit__task_list` and `mcp__bit__task_read` return carries an `approved` boolean. If it's `false`, stop and tell the user:

> "BIT-X.N is not approved — approve it first in the TUI (`bp ui`)."

Don't mark the bar `doing` or touch any code until the gate is cleared. A bar whose approval was revoked by an edit is just as blocked as one that was never approved — `approved` is either `true` or it isn't. Note that there is no approve tool: approval is the operator's act, so clearing the gate is something only they can do.

### 2. Mark the bar in progress, load its checklist

Move the bar to `doing` (`mcp__bit__task_update` with `status` set to `doing`) and roll the track up: if the track isn't already `doing`, set it the same way. Now the board reflects that this step is active.

Then put each implementation-checklist item from the bar into your harness's session task list (the TaskCreate tool, if your harness has one), one task per item. Mark them in-progress and completed as you work. This keeps the bar's sub-tasks visible to the user and stops you from dropping or merging them. Only load the *current* bar's items — not the whole track. (This is harness bookkeeping, separate from the bars tracked in the store — skip it if no such tool is available.)

### 3. Implement the bar

Do only this bar's work:
- Check the bar body for a `**Needs real data:**` note before writing anything. If present, run the mechanism it names — or ask the user to run it or hand you the result — and use that real artifact for the test data. Writing the test first against an invented value defeats the point of this bar having flagged it.
- Touch only the scope files the step names. Read them and their adjacent code before editing, so you extend the existing pattern rather than inventing a new one.
- Follow the plan's intent on tests (TDD where it says so; YAGNI — don't add behavior or tests the step didn't ask for).
- No "while we're in here" cleanup. If you notice something out of scope, mention it for a later step; don't fix it now.

**This is a Go project — apply the project's Go skills while writing code, not just at review time.** Consult them proactively as they become relevant to the step, rather than writing plain Go first and retrofitting idioms after:
- **go** — idiomatic Go for any `.go` file you touch: package design, error handling, interfaces, concurrency, testing patterns. Applies to every step that writes or edits Go code.
- **cobra-viper** — when the step adds or changes CLI commands, subcommands, flags, or configuration binding.
- **wails** — when the step touches a desktop/webview frontend-to-Go bridge.
- **fileflow-pathologize** — when the step moves, copies, renames, or generates files on disk, or builds a path from untrusted input.
- **go-spec-reviewer** — if the step's plan detail reads more like a spec than settled code (rare mid-plan, but check before implementing a step that introduces a new subsystem).
- **go-release** — only for steps that touch `go.mod` versioning, tags, or exported API surface of a published module.

Not every step needs every skill — pick the ones the step's scope files actually call for.

### 4. Run the "Claude verifies" checks

Run the deterministic checks the step lists under **Claude verifies** — whatever commands the plan specifies (test suite, linter, build, count assertion, etc.). Some checks are long-running or ones the user prefers to trigger themselves — present those and ask before running, rather than assuming. Report what passed and what didn't, with the actual output, not a summary that hides a failure.

If an automated check fails, fix it within this bar's scope and re-run before stopping. A bar isn't ready for the user until its own checks pass.

### 5. Close out the bar

How you close out depends on whether the bar has anything left for a *human* to judge — that's exactly what its **User verifies** items are.

**If the bar has User verifies items**, those are real judgment calls the automated checks can't settle — does the API feel right, is this safe to ship, does the output make sense for real data. Present them as a checklist for the user to work through, show the commit message bit_commit will use (don't wait to be asked — it's part of what "done" means, not a follow-up question), then stop and hand control back. Leave the bar `doing`; it isn't done until the user has looked. When they confirm, run the **Verified good** close-out below, which commits.

**If the bar has no User verifies items**, there's nothing for a human to judge — the passing "Claude verifies" checks *are* the verification, and waiting for a "looks good" that carries no new information just burns a round-trip. So run the **Verified good** close-out now, inline: commit through bit_commit, roll the track up, and prompt the compaction point. This is optimistic, not unsupervised — bit_commit shows the files and the message, and nothing is committed until the operator says yes. If they spot a problem later, they say so and you **unwind**: the commit stays. Set the bar back to `doing` (`mcp__bit__task_update` with `status` set to `doing`), reverse any verse checkoff you made, and treat it as **Not as expected**. The call returns `approved`, and on this path it comes back `true` — walking a bar back to `doing` isn't a change to what was reviewed, so the bar stays approved and can resume without a second blessing. The fix lands as a follow-up commit through bit_commit, which overwrites the bar's `commit` and reuses its subject.

Either way, two lines hold firm: never commit except through bit_commit, and do **not** start the next bar. The next bar is the user's call.

The user often follows up with small cleanup on the step you just implemented — a tweak, a rename, "actually make this a table test," fixing something the checks didn't catch. Handle those in place, without treating them as a new step. If the bar is already committed, every such reply ends by offering a follow-up commit through bit_commit. If it isn't committed yet, the reply ends with the message bit_commit will use (refined if the change affects what it should say). The point is that the user never has to ask for it — it should be the last thing they see once the step's code is in a state they could commit, however many small back-and-forths it took to get there.

When one of those follow-ups was a *correction* — the tweak fixed something the plan should have settled and didn't — add one line offering to record it: "want me to capture that as a feedback note?" Yes or no, and `/bit:feedback` writes it. Never write the note unasked; the user is the judge of what counts as feedback, and the offer only saves them having to remember it exists.

---

## Closing out

### Verified good

This is the close-out procedure step 5 points to — run it inline for a bar with no **User verifies** items, or once the user confirms a bar that had them.

1. **Commit through bit_commit.** It asks the operator, commits, and records `commit`, `branch` and `done` on the bar in one `task_update`. If the commit is declined, or a hook failure is unresolved, the bar stays `doing`: stop here, with no rollup. With nothing to commit, or no repo, bit_commit marks the bar `done` after the operator's OK. The status field is the resume marker: a fresh session continues at the first bar that isn't `done`, with no doc to parse.
2. **Roll the track up.** This is skill logic run through the tools (nothing cascades for you):
   - Re-list the bars: `mcp__bit__task_list` with `parent` set to the track ID.
   - **Verse checkoff:** if this bar was the *last* one tagged to its verse — every bar with that `phase` is now `done` — check off that verse in the track body: find its `- [ ] Verse N` line and change `[ ]` to `[x]` (bit_scope keeps the checkbox and `Verse N` on the same line, so it's a one-line toggle). Read the body, edit that line, write it back.
   - **Track status:** none started (all `todo`) → `todo`; anything else → `doing`. Note what's deliberately *absent*: even when every bar is now `done`, you do **not** set the track `done` here. A finished-looking track stays `doing` until `/bit:complete` runs after the push. See **Track sign-off** below.
   - Apply both in one call so the track moves once: `mcp__bit__task_update` on the track — pass `body` only if the verse checkoff changed it, `status` only if the status changed, since an omitted field is left unchanged. **If neither changed, there's nothing to roll up — skip the call.** (This is the common mid-verse case: finishing a bar when its verse isn't complete yet and the track is already `doing`.)

   Keeping the track's verse checklist and status current lets a reader see delivered value at a glance from one `mcp__bit__task_read` on the track — and the track and its bars never disagree about what's done.
3. **Compaction point.** Tell the user this is a clean place to `/compact`, since the bar is done, verified, and committed. You can't run `/compact` yourself — it's a user command — so prompt them, then continue when they say so. If this was the **last** bar — the rollup shows every bar `done` — there's no next bar to continue to; point them at **Track sign-off** instead.

### Track sign-off

A track is completed once its work has landed, not when its last bar is done. So when you close out a bar and the rollup shows **every** bar is now `done`, finish that bar's own close-out as normal (verified, committed), then tell the operator the track is ready. The next step is theirs: push (or merge) the work, then run `/bit:complete <track>`, which checks that it landed, records the landing commit and files the track as completed.

The track stays `doing` until then. Don't set it `done` and don't file it yourself. If the operator isn't ready — wants more testing, or spots something — nothing changes, and `/bit:complete` waits until they are.

### Not as expected

Don't thrash or silently retry the same approach. Figure out which of three problems it is — ask the user if it isn't obvious:

1. **The scope is wrong.** The bar did what it said, but the *direction* is off — the verse isn't delivering the value we expected, or the delivery order is wrong. This is bigger than one bar. Stop and hand back to **bit_scope** to rethink the track's shape; that will usually mean re-planning the affected verses (their bars) with bit_plan afterward.
2. **The plan is wrong.** The scope is sound, but this bar's detail was off. Stop implementing and hand back to **bit_plan** to revise the bar. Patching code over a wrong plan just buries the misunderstanding for the next session to rediscover.
3. **The plan is right, the implementation is wrong.** The intent was correct but the code doesn't deliver it. The user will often fix this directly. Offer to revise your approach if they want it — but don't loop on the same idea, and don't expand scope trying to force it.

Cases 1 and 2 are plan gaps by definition — the plan was silent, or wrong, about something the work turned out to require. So before control leaves for bit_scope or bit_plan, offer to record it in one line, answered yes or no; `/bit:feedback` writes the note. The timing is the whole point: the hand-back is what overwrites the broken plan, so this is the last moment the evidence of what went wrong still exists.

In all cases, leave the bar **not `done`** so it stays the next bar to resume. Its `doing` status is fine and accurate — work started, not verified. If you'd already auto-marked it `done` (the no-User-verifies path) and the user then flagged a problem, set it back to `doing` and reverse any verse checkoff you made — that's the unwind step 5 mentions. A committed bar keeps its commit, and its fix is a follow-up commit through bit_commit.

---

## What this skill does not do

- **Author or redesign the scope or plan** — that's bit_scope and bit_plan. If there are no bars yet, or the bars or the track's shape need rethinking, switch to the right authoring skill.
- **Commit on its own.** Every commit goes through bit_commit, which asks first. A declined commit leaves the bar `doing`.
- **Push.** bit_do never pushes. bot-dev does, after a permitted commit.
- **Run multiple bars unattended** — one bar per cycle, every time.
- **Declare a track done on its own** — finishing the last bar makes the track *ready*; the operator then pushes and runs `/bit:complete`, which marks it `done` and files it as completed.
- **Go around the tools** — the `mcp__bit__*` tools are the only way in; every status move and body change goes through them.
- **Compact on its own** — the user runs `/compact`; you mark the boundary.
