---
name: bit_do
description: Execute an existing implementation plan one step at a time, stopping after each step for the user to verify before continuing. Use whenever the user says "implement the plan", "continue our implementation", "let's build the next step", "do the next step", "pick up where we left off", or otherwise wants to carry out — not write or revise — a markdown bit_plan. This is the execution counterpart to bit_plan and bit_scope: bit_scope frames the WHY and delivery order, bit_plan authors the detailed steps, bit_do carries them out. It reads both the plan and its partner scope, tracks each step's checklist as tasks, runs the plan's automated checks, marks steps done in the plan and rolls completed phases up into the scope, and hands off to the user for verification and commit between steps. This project is a Go codebase — bit_do applies the project's Go skills (go, cobra-viper, wails, fileflow-pathologize, go-spec-reviewer, go-release) while implementing so code stays idiomatic. Trigger this — not bit_plan — when a plan already exists and the user wants to start or resume building it.
---

# Plan Implementer

You execute implementation work that bit_scope and bit_plan produced. It lives across **two documents**:

- **the plan** (`<feature>-plan.md`, from bit_plan) — the executable detail: a series of **steps**, each one red-green cycle with a scope, an implementation checklist, "Claude verifies" checks, "User verifies" checks, and a suggested commit. This is what you carry out.
- **the scope** (`<feature>-scope.md`, from bit_scope) — the high-level overview and the WHY. Its coarse **phases** are what the plan's steps roll up into. You read it for context and keep it in sync; you execute against the plan.

A note on vocabulary, because it's easy to trip on: the **scope** has coarse **phases** (usable value slices); the **plan** breaks each phase into fine-grained **steps** (one commit each, tagged to their phase). You execute one *step* at a time; a *phase* is done when all its steps are.

Your job is to carry out **one step, then stop**. The plan was deliberately broken into steps that are each independently verifiable and committable. Verification is the user's call, and so is the commit. Pushing ahead into a second step blurs what is being verified and what is going into a single commit — which is exactly what the stepped structure exists to prevent.

---

## The loop

### 1. Find the next step

Read **both** documents — the plan and its partner scope (paired by name, `<feature>-plan.md` / `<feature>-scope.md`, or via the link in the plan's Context). The user usually names the work when they trigger this ("implement the geographies plan"); use that. If they didn't, and more than one plan exists, list what you found and ask which one — don't pick for them.

The next step is the first one in the plan **not** marked `✅ Done`. Note which scope phase it's tagged to, so you know what larger capability it's building toward — and read that phase's WHY in the scope for context.

Briefly restate: the step's name, the scope phase it serves, its scope files, and its checklist. This confirms you and the user are aligned on what's about to happen — and at what altitude — before any code changes.

### 2. Load the checklist into tasks

Put each implementation-checklist item from the step into the task list (TaskCreate), one task per item. Mark them in-progress and completed as you work. This keeps the step's sub-tasks visible to the user and stops you from dropping or merging them. Only load the *current* step's items — not the whole plan.

### 3. Implement the step

Do only this step's work:
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

If an automated check fails, fix it within this step's scope and re-run before stopping. A step isn't ready for the user until its own checks pass.

### 5. Stop for user verification

Present the step's **User verifies** items as a checklist for the user to work through. Then state the step's suggested commit message — don't wait to be asked for it; it's part of what "done with this step" means, not a follow-up question. Then stop:
- Do **not** commit.
- Do **not** start the next step.
- Do **not** compact yet.

Hand control back and wait.

The user often follows up with small cleanup on the step you just implemented — a tweak, a rename, "actually make this a table test," fixing something the checks didn't catch. Handle those in place, without treating them as a new step. But every such reply still ends with the commit message (refined if the change affects what it should say), the same way the original stop did. The point is that the user never has to ask for it — it should be the last thing they see once the step's code is in a state they could commit, however many small back-and-forths it took to get there.

---

## After the user verifies

### Verified good

1. **Mark the step done in the plan.** Check its implementation-checklist boxes (`- [x]`) and add a status line under the step name:
   `**Status:** ✅ Done — verified <YYYY-MM-DD>`
   This is the resume marker: a fresh session continues at the first step without it.
2. **Roll up to the scope.** Check whether this was the *last* step tagged to its scope phase. If every step for that phase is now done, check off the phase in the scope (`- [x] Phase N — …`); if steps remain, leave it unchecked. Keeping the scope current lets a reader see delivered value at a glance without reading the plan — and the two documents never disagree about what's done.
3. **Suggest the commit.** Offer the step's commit message (refined if the work diverged from it). The user commits — you never run the commit yourself. If checking off a scope phase, fold that scope edit into the same commit so the docs move with the code.
4. **Compaction point.** Tell the user this is a clean place to `/compact` before the next step, since the step is done, verified, and committed. You can't run `/compact` yourself — it's a user command — so prompt them, then continue to the next step when they say so.

### Not as expected

Don't thrash or silently retry the same approach. Figure out which of three problems it is — ask the user if it isn't obvious:

1. **The scope is wrong.** The step did what it said, but the *direction* is off — the phase isn't delivering the value we expected, or the delivery order is wrong. This is bigger than one step. Stop and hand back to **bit_scope** to rethink the shape; that will usually mean re-planning the affected phases with bit_plan afterward.
2. **The plan is wrong.** The scope is sound, but this step's detail was off. Stop implementing and hand back to **bit_plan** to revise the step. Patching code over a wrong plan just buries the misunderstanding for the next session to rediscover.
3. **The plan is right, the implementation is wrong.** The intent was correct but the code doesn't deliver it. The user will often fix this directly. Offer to revise your approach if they want it — but don't loop on the same idea, and don't expand scope trying to force it.

In all cases, leave the step **unmarked** so it stays the next step to resume.

---

## What this skill does not do

- **Author or redesign the scope or plan** — that's bit_scope and bit_plan. If there's no plan yet, or the plan's steps or the scope's shape need rethinking, switch to the right authoring skill.
- **Commit** — always the user's action; you suggest the message.
- **Run multiple steps unattended** — one step per cycle, every time.
- **Compact on its own** — the user runs `/compact`; you mark the boundary.
