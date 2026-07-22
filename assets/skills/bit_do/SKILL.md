---
name: bit_do
description: Execute an existing implementation plan one step at a time, stopping after each step for the user to verify before continuing. Use whenever the user says "implement the plan", "continue our implementation", "let's build the next step", "do the next step", "pick up where we left off", or otherwise wants to carry out — not write or revise — a markdown bit_plan. This is the execution counterpart to bit_plan and bit_scope: bit_scope frames the WHY and delivery order in a track, bit_plan authors the detailed steps as bars under it, bit_do carries them out. It finds the track in `.bit/`, reads the track body and its bars through the `bit` CLI, tracks each bar's checklist as tasks, runs the automated checks, moves each bar's status (`doing` → `done`) and rolls the track up (checking off completed verses and setting the track's status), and hands off to the user for verification and commit between bars. This project is a Go codebase — bit_do applies the project's Go skills (go, cobra-viper, wails, fileflow-pathologize, go-spec-reviewer, go-release) while implementing so code stays idiomatic. Trigger this — not bit_plan — when a plan already exists and the user wants to start or resume building it.
---

# Plan Implementer

You execute implementation work that bit_scope and bit_plan produced. It lives in one **track** in `.bit/`, driven through the `bit` CLI:

- **the bars** (child tasks under the track, from bit_plan) — the executable detail: each bar is one step — one red-green cycle with a scope, an implementation checklist, "Claude verifies" checks, "User verifies" checks, and a suggested commit. This is what you carry out, one bar at a time.
- **the track body** (from bit_scope) — the high-level overview and the WHY, with the coarse **verses** the bars roll up into. You read it for context and keep its verse checklist in sync; you execute against the bars.

A note on vocabulary, because it's easy to trip on: the **track** carries coarse **verses** (usable value slices) in its body; its **bars** are the fine-grained steps (one commit each, tagged to the verse they serve via the `--phase` flag — the flag keeps the name `phase`, the scope's slice is a verse). You execute one *bar* at a time; a *verse* is done when all its bars are.

**Before you drive the CLI, read `.claude/bit-cli.md`** — the shared command contract (find the track, list its bars, read a bar body, set status, roll the track up). Every write goes through `bit`; never hand-edit `.bit/tasks/*.md`.

Your job is to carry out **one bar, then stop**. The plan was deliberately broken into bars that are each independently verifiable and committable. Verification is the user's call, and so is the commit. Pushing ahead into a second bar blurs what is being verified and what is going into a single commit — which is exactly what the stepped structure exists to prevent.

---

## The loop

### 1. Find the next bar

Find the **track**. The user usually names the work when they trigger this ("implement the geographies track"); resolve it to a track ID with `bit task list` (tracks are the rows whose ID has no dot — match on title). If they didn't name it and more than one track has unfinished bars, list what you found and ask which one — don't pick for them.

Read the track body for context (`bit task read <track> --body`), then list its bars in order (`bit task list --parent <track>`). The next bar is the **first one whose status is not `done`** — the status field is the resume marker, so a fresh session lands on the right bar with no doc to parse. Read that bar's body (`bit task read <bar> --body`) for its detail, and note the verse it's tagged to (shown as `phase N — label` in the list output) so you know what larger capability it's building toward.

Briefly restate: the bar's ID and name, the verse it serves, its scope files, and its checklist. This confirms you and the user are aligned on what's about to happen — and at what altitude — before any code changes.

### 2. Mark the bar in progress, load its checklist

Move the bar to `doing` (`bit task update <bar> -s doing`) and roll the track up: if the track isn't already `doing`, set it (`bit task update <track> -s doing`). Now the board reflects that this step is active.

Then put each implementation-checklist item from the bar into your harness's session task list (the TaskCreate tool, if your harness has one), one task per item. Mark them in-progress and completed as you work. This keeps the bar's sub-tasks visible to the user and stops you from dropping or merging them. Only load the *current* bar's items — not the whole track. (This is harness bookkeeping, separate from the bit tasks themselves — skip it if no such tool is available.)

### 3. Implement the bar

Do only this bar's work:
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

**If the bar has User verifies items**, those are real judgment calls the automated checks can't settle — does the API feel right, is this safe to ship, does the output make sense for real data. Present them as a checklist for the user to work through, state the bar's suggested commit message (don't wait to be asked — it's part of what "done" means, not a follow-up question), then stop and hand control back. Leave the bar `doing`; it isn't done until the user has looked. When they confirm, run the **Verified good** close-out below.

**If the bar has no User verifies items**, there's nothing for a human to judge — the passing "Claude verifies" checks *are* the verification, and waiting for a "looks good" that carries no new information just burns a round-trip. So run the **Verified good** close-out now, inline: mark the bar `done`, roll the track up, state the commit message, and prompt the compaction point. This is optimistic, not unsupervised — the user still reads the diff when they commit. If they spot a problem, they say so and you **unwind**: set the bar back to `doing` (`bit task update <bar> -s doing`), reverse any verse checkoff you made, and treat it as **Not as expected**. The done state is cheap to undo, and marking it now keeps the `.bit/tasks/*.md` status change in the tree for the *same* commit as the code, instead of lagging into the next bar's.

Either way, two lines hold firm: do **not** commit, and do **not** start the next bar. Both are the user's call.

The user often follows up with small cleanup on the step you just implemented — a tweak, a rename, "actually make this a table test," fixing something the checks didn't catch. Handle those in place, without treating them as a new step. But every such reply still ends with the commit message (refined if the change affects what it should say), the same way the original close-out did. The point is that the user never has to ask for it — it should be the last thing they see once the step's code is in a state they could commit, however many small back-and-forths it took to get there.

---

## Closing out

### Verified good

This is the close-out procedure step 5 points to — run it inline for a bar with no **User verifies** items, or once the user confirms a bar that had them.

1. **Mark the bar done.** `bit task update <bar> -s done`. The status field is the resume marker: a fresh session continues at the first bar that isn't `done`, with no doc to parse. (You don't need to tick the checklist boxes inside the bar body — the status field supersedes them.)
2. **Roll the track up.** This is skill logic run through the CLI (the tool doesn't cascade for you):
   - Re-list the bars: `bit task list --parent <track>`.
   - **Verse checkoff:** if this bar was the *last* one tagged to its verse — every bar with that `--phase` is now `done` — check off that verse in the track body: find its `- [ ] Verse N` line and change `[ ]` to `[x]` (bit_scope keeps the checkbox and `Verse N` on the same line, so it's a one-line toggle). Read the body, edit that line, write it back.
   - **Track status:** all bars `done` → track `done`; none started (all `todo`) → `todo`; anything in between → `doing`.
   - Apply both in one call so the track moves once: `bit task update <track> -d "<edited body>" -s <status>` — pass `-d` only if the verse checkoff changed the body, `-s` only if the status changed. **If neither changed, there's nothing to roll up — skip the call.** (This is the common mid-verse case: finishing a bar when its verse isn't complete yet and the track is already `doing`.)

   Keeping the track's verse checklist and status current lets a reader see delivered value at a glance from `task read <track>` — and the track and its bars never disagree about what's done.
3. **Suggest the commit.** Offer the bar's commit message (refined if the work diverged from it). The user commits — you never run the commit yourself. The `.bit/tasks/*.md` changes from steps 1–2 are part of the working tree, so they go into the same commit as the code — mention that.
4. **Compaction point.** Tell the user this is a clean place to `/compact` before the next bar, since it's done, verified, and committed. You can't run `/compact` yourself — it's a user command — so prompt them, then continue to the next bar when they say so.

### Not as expected

Don't thrash or silently retry the same approach. Figure out which of three problems it is — ask the user if it isn't obvious:

1. **The scope is wrong.** The bar did what it said, but the *direction* is off — the verse isn't delivering the value we expected, or the delivery order is wrong. This is bigger than one bar. Stop and hand back to **bit_scope** to rethink the track's shape; that will usually mean re-planning the affected verses (their bars) with bit_plan afterward.
2. **The plan is wrong.** The scope is sound, but this bar's detail was off. Stop implementing and hand back to **bit_plan** to revise the bar. Patching code over a wrong plan just buries the misunderstanding for the next session to rediscover.
3. **The plan is right, the implementation is wrong.** The intent was correct but the code doesn't deliver it. The user will often fix this directly. Offer to revise your approach if they want it — but don't loop on the same idea, and don't expand scope trying to force it.

In all cases, leave the bar **not `done`** so it stays the next bar to resume. Its `doing` status is fine and accurate — work started, not verified. If you'd already auto-marked it `done` (the no-User-verifies path) and the user then flagged a problem, set it back to `doing` and reverse any verse checkoff you made — that's the unwind step 5 mentions.

---

## What this skill does not do

- **Author or redesign the scope or plan** — that's bit_scope and bit_plan. If there are no bars yet, or the bars or the track's shape need rethinking, switch to the right authoring skill.
- **Commit** — always the user's action; you suggest the message.
- **Run multiple bars unattended** — one bar per cycle, every time.
- **Hand-edit `.bit/tasks/*.md`** — every status move and body change goes through `bit`.
- **Compact on its own** — the user runs `/compact`; you mark the boundary.
