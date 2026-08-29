---
name: bit_retro
description: Reads across a project's `.bit/feedback/*.md` notes — the evidence bit_feedback records — to diagnose what the bit_scope/bit_plan/bit_do/bit_check/bit_feedback process itself failed to settle, and turns any recurring, pipeline-level pattern into a portable, client-agnostic proposal that bit_learn can act on. Use this whenever the user says "bit_retro", "retro", "retrospective", "how did this cycle go", "what went wrong", or after a bit_check (or several bit_feedback notes) and wants to close the loop — also reach for it when a track, or the whole project, has accumulated feedback notes worth reading together, even if the user just says "let's look back at this" or "what should we learn from this." This is the mechanism that makes the pipeline self-improving, and it's deliberately safe to run inside a client or otherwise confidential project — every proposal it writes is already generalized before it's written, and applying it is a separate skill (bit_learn) that runs somewhere else entirely.
---

# Pipeline Retrospective

bit_feedback records evidence, one note at a time, on purpose without judgment — its own words: "no cause, no category, no lesson learned... reading across notes is a separate cycle." This skill is that separate cycle. You read every note that bears on the question at hand, find what actually recurs, and decide whether it's a lesson about *this project* (already handled, nothing to carry forward) or a lesson about the *bit-\* pipeline itself* — something worth fixing so every project that uses it benefits.

You produce one artifact: a proposals file, written only after the user has walked through and accepted each proposal with you. Nothing here edits a skill, changes the CLI, or leaves this project on its own — that's a second skill, bit_learn, and it runs somewhere else (see *Handoff* at the end).

---

## Context reality

This often runs after `/compact`, sometimes long after the cycle it's reviewing. Don't invent a narrative. Reconstruct from durable signals:

- `.bit/feedback/*.md` — the notes themselves, see below.
- The track and bar bodies: `mcp__bit__task_read`, whose `body` field *is* the prose.
- `git log --oneline` for what actually shipped.
- The user, for anything the documents don't answer — they remember the experience, you remember the artifacts.

If you're not sure something happened the way you're about to say it did, say so and ask.

## Finding the evidence

Pull track and bar context with `mcp__bit__task_list` and `mcp__bit__task_read`. Note the asymmetry up front: you *read* through the tool surface, but you do not *write* through it here — this skill's own output (below) is a plain file you write directly, because the surface carries no retro tool. That asymmetry doesn't have to be taken on faith anymore: the tool list is enumerable, so a tool that doesn't exist is visibly absent rather than something you have to be warned about.

Feedback notes have no tool of their own either — they're plain files. List `.bit/feedback/*.md` directly and read each one. A note's own prose names its track and cites its bar, so no separate lookup is required to place it; cross-reference the track with `mcp__bit__task_read` only when you need to check a note's claim against the track's *current* Decisions or Verses (a track may have been rescoped since the note was written).

Scope the review before diving in: one track, several, or the whole project ("the whole album"). If the user didn't say and more than one track has notes, list what you found and ask — don't guess which ones matter.

## Reading across notes for the pattern

The question for the whole cycle is:

> What did scope, plan, do, check, or feedback fail to settle that it should have, and would settling it up front have prevented this note *and everything shaped like it*?

Not: was this one instance bad. A retro that walks through notes one at a time and restates each one hasn't done its job — group the notes that share a shape, cite their IDs together, and name the shared failure once. Some cycles turn up one real pattern from twelve notes; that's success, not thin coverage.

**Filter: pipeline-relevant, or already handled locally?** Most notes describe something that got fixed right there in the project and needs nothing further — a missing project-specific convention, a one-off wrong assumption about *this* codebase, a local tooling gap. That's not a retro finding; it's just history. A note is a retro *candidate* only when the gap is in the pipeline itself: something bit_scope, bit_plan, bit_do, bit_check, or bit_feedback should have asked, checked, or enforced, regardless of which project it's running in.

The test: **could this exact gap recur in a completely different project using bit-pro?** If yes, it's a candidate. If the answer depends on this project's own stack, conventions, or data, it stays local — don't write a proposal just to have something to show for the pass.

## Considering the mechanism

For each real pipeline-level pattern, the question isn't only *which skill's prose should say this* — it's *would prose actually catch this reliably, or does it need to be caught mechanically*. Skill prose only works if the model remembers to apply it, every time, including three follow-ups deep into a conversation. Some failures are better addressed by a Claude Code plugin hook or a bundled MCP server than by one more sentence in a SKILL.md.

Reason from what's actually true, not from what would be convenient to be true. Current ground truth (verified against Anthropic's plugin/hooks documentation — treat this table as the source, don't invent event names or fields beyond it):

- Plugins can ship **hooks** (`hooks/hooks.json`, or inline in the plugin manifest) and **MCP servers** (`.mcp.json` or `mcpServers` in the manifest — stdio only: a bundled local process or a remote command, not arbitrary HTTP).
- Hook events that exist today: session lifecycle (`SessionStart`, `Setup`, `SessionEnd`); per-turn (`UserPromptSubmit`, `UserPromptExpansion`, `Stop`, `StopFailure`); tool execution (`PreToolUse`, `PostToolUse`, `PostToolUseFailure`, `PostToolBatch`, `PermissionRequest`, `PermissionDenied`); subagents (`SubagentStart`, `SubagentStop`, `TeammateIdle`, `TaskCreated`, `TaskCompleted`); file/config (`FileChanged`, `ConfigChange`, `InstructionsLoaded`, `CwdChanged`, `DirectoryAdded`); compaction/MCP (`PreCompact`, `PostCompact`, `Elicitation`, `ElicitationResult`); messages/worktrees (`MessageDisplay`, `WorktreeCreate`, `WorktreeRemove`, `Notification`).
- `PreToolUse` can block, allow, or modify a tool call before it runs. `PostToolUse`/`PostToolBatch` can modify output or inject context after tool calls resolve, before the next model turn — this is the one that fits "make sure X reliably happens on every reply," since it fires every turn regardless of how the reply was shaped. `Stop` can block Claude from stopping and inject `additionalContext`, but **its input carries no field distinguishing why Claude is stopping** — task genuinely finished, versus the user asking for an early handoff to a different step. Two independently configured Stop hooks can't tell each other apart on that basis. Don't propose a Stop-hook fix for "distinguish these two cases" — no such mechanism currently exists; say that plainly instead of inventing one.
- MCP is for live or queryable external state, not for "remind the model to do something." A pattern that's really about reliability of a reminder is a hook question (or, if genuinely unenforceable today, an open question) — reaching for MCP there is solving the wrong layer.

Worked examples, so the shape is concrete:
- *"Every reply after a step's close-out should end with the commit message, even a few small follow-ups later — right now it depends on the model remembering, and it doesn't always."* → candidate: a `PostToolBatch` hook that checks whether the reply looks like a close-out and reminds if the commit message is missing. Grounded, because `PostToolBatch` actually fires every turn.
- *"A downstream project's own Stop hook forces code changes even after the user asked to stop and hand off to planning."* → **not** a proposal to write, because no hook-level mechanism currently distinguishes the two Stop cases. The honest proposal here is "no good mechanism yet" — note it as an open question, don't fabricate a fix to close it out.

If you're not sure a mechanism does what you think it does, say so in the proposal rather than asserting it — the same standing rule that applies to any tool or config syntax.

## Writing proposals

Only after the user has seen and accepted a proposal does it go in the file — this mirrors the old retro's accept/reject habit, and for the same reason: a proposal that goes to another project unreviewed is a proposal nobody vetted for confidentiality. Walk through candidates conversationally first.

A proposal that only names a category ("verify citations," "check the mocks") isn't finished — it tells bit_learn *that* something should change without saying what to actually do about it, which just moves the vagueness downstream instead of resolving it. Two fields carry the weight here, and neither is optional:

- **Root cause chain** is not one line restating the symptom in different words — "the citation was wrong because it came from recalled knowledge instead of the installed source" describes what happened, not why the process let it happen. Ask **why** and answer it; then ask why *that* is true, and answer again; keep going until the answer names an actual mechanism — a check that doesn't exist anywhere in the relevant skill, a rule that exists but nothing enforces, an assumption the process bakes in without ever stating it. Show the chain itself, numbered, not just its last link — a reader needs to see the reasoning to judge whether it actually holds, not just trust the conclusion.

  **There's no fixed number of whys.** Some problems bottom out at an actual mechanism on the first question; others take five or more. Don't treat 3 (or any count) as a target to hit or a minimum to clear — the depth is however many it actually takes, and stop the moment an answer is concrete enough for **bit_learn** to act on directly: "nothing reads both and cross-checks them" is a good stopping point *because it names a missing mechanism*, not because it's the third why. Padding a chain past that point manufactures depth instead of finding it; stopping short of it leaves a symptom wearing the label.

  As you chase separate proposals down their own chains, watch for chains that bottom out at the *same* mechanism — that's not duplication, it's a deeper, more general finding wearing several instances, and it's worth naming once rather than as separate proposals that each patch a symptom.

- **Concrete change** is the literal instruction, written exactly as it would land — a drop-in paragraph or checklist item for the specific section of the specific skill it belongs in (name the section, not just the skill), or the specific behavior a hook/CLI feature needs to implement. Not "bit_plan should verify citations" — "add to bit_plan's References guidance: before citing a third-party library's file:line, read the actually-installed source in this repo's module cache, not general docs or recalled API shape." It should follow directly from the *last* link of the root cause chain, not from the symptom.

Each accepted proposal, in `.bit/retro/<track-or-album>-proposals.md`:

```markdown
## Proposal: <short, generic title>

**Problem:** <what actually recurred, worded generically — the shape, not the instance>
**Root cause chain:**
1. Why did this happen? <answer>
2. Why is *that* true? <answer>
3. Why is *that* true? <answer>
[continue until the answer names an actual, missing-or-unenforced mechanism — not another
restatement of the symptom]
**Seen in:** <track>-<NNN>, <track>-<MMM> [local reference only, for this project's own
audit trail — these IDs mean nothing outside it and bit_learn doesn't need them]
**Suggested mechanism:** skill: bit_<name>, section "<the specific heading it changes>"
  | hook: <EventName> — <the specific behavior it needs to implement>
  | mcp | bp CLI feature — <the specific behavior it needs to implement> | unclear
**Concrete change:** <the actual drop-in instruction — literal paragraph or checklist item
text for a skill section, or a specific behavior spec for a hook/CLI feature, following
directly from the last link of the chain. No project names, no business-domain nouns, no
schema/API specifics from this project — but no category labels standing in for the real
text either.>
```

If a pattern can't be stated that way without losing what makes it true, that's usually a sign it isn't actually pipeline-level (back to the filter above) — don't force an abstraction that guts the point just to have an entry. And if the chain won't bottom out at an actual mechanism — only at "the model should have been more careful" — that's a sign this isn't a process gap at all, and forcing a **Concrete change** onto it will just produce a rule nothing can enforce; say so and leave it as an open question instead.

Report the file's path back to the user when you're done — that's the handle they carry elsewhere.

---

## Avoiding staleness

A first retro on a project with a real backlog of notes is high-signal. Repeat passes risk diminishing returns:

- **Check for existing proposals first.** Don't re-surface a pattern already written to a `.bit/retro/*-proposals.md` file, accepted or rejected — if rejected, the user already made that call.
- **It's fine to come back empty.** "Nothing here recurs, and what recurred before is already covered" is a complete retro. Don't manufacture a pattern to justify the pass.
- **Raise the bar each time.** The second and third retro on the same project should need more evidence to justify a new proposal than the first one did.

## Tone

Be specific — cite the actual note IDs, the actual quoted text, the actual pattern. No generalities standing in for evidence. If a whole pass turns up nothing pipeline-level, say that plainly and stop; don't pad it out.

---

## What this skill does not do

- **Edit a skill, the CLI, or anything else in bit-pro** — that's bit_learn, and it runs in bit-pro itself, not here.
- **Move the proposals file anywhere** — carrying it from this project to wherever bit_learn runs is the user's own action, deliberately manual; this skill has no opinion on how.
- **Write feedback notes** — that's bit_feedback; this skill only reads what's already been written.
- **Assert a mechanism it isn't sure exists** — an unverified hook event, field, or CLI flag doesn't belong in a proposal; an honest "unclear" does.
- **Restate individual notes as if that were the finding** — the finding is the pattern across them, or that there isn't one.
