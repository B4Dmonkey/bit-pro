---
name: bit_learn
description: Reads a proposals file that bit_retro produced elsewhere — often carried over by hand from a different, sometimes confidential, project — and for each proposal either drafts the matching bit-pro skill edit via skill-creator, hands a real code/CLI change off to bit_scope to be planned properly, or says plainly that neither fits and asks the user what to do. Use whenever the user says "bit_learn", "learn", "apply these proposals", "here's what retro found", hands you a `.bit/retro/*-proposals.md` file or its contents from elsewhere, or asks what to do with retro or feedback output. This only ever runs inside bit-pro itself, since it's the only place the skills and the `bp` CLI a proposal might change actually live — never inside the project the proposals came from.
---

# Proposal Triage

bit_retro runs elsewhere — in whatever project produced the feedback — and hands you a file of generalized, already-reviewed proposals. Your job starts where that ends: for each proposal, decide what actually has to happen in *this* repo (bit-pro) to make it real, or decide that nothing should, yet.

You never edit a skill directly and you never implement a CLI feature inline. You route: to skill-creator for a skill-text change, to bit_scope for anything that's really a code or CLI change, or back to the user when it's neither.

---

## Finding the input

Ask for the proposals file if the user didn't already hand you a path or its contents. Read it in full before acting on anything in it — proposals can reference each other (the same recurring pattern sometimes shows up as more than one angle on the same fix).

Handle proposals **one at a time**, the same reason bit_do handles one bar at a time: bulk-applying a batch blurs what's being reviewed and what's landing in which change.

## Sanity-check before triaging

A proposal's `Suggested mechanism` reflects bit-pro's state *when bit_retro wrote it* — not necessarily now. Proposals can sit around before they're carried over, and bit-pro's skills or CLI may have already moved since. Before acting on any proposal:

- If it names a target skill, read that skill's current text. The gap it describes may already be covered — say so and skip it rather than re-applying a fix that's no longer needed.
- If it names a CLI feature, check whether `bp`/`bit` already does it.

Only proceed with proposals that are still real gaps.

## The three buckets

For each still-real proposal, decide which one it is — don't force a fit if it's genuinely unclear.

**1. Skill-text change.** The proposal's `Concrete change` describes something a bit-\* skill's prose should say (or say differently), and its `Suggested mechanism` names a target skill. Identify the exact target under `bit/skills/` — use that `skill: bit_<name>` hint as a starting point, but confirm it's the right one; a pattern about "the plan inventing an unverified value" might belong in bit_plan's prose even if the hint pointed at bit_do. Invoke **skill-creator** (via the Skill tool) to draft the edit, giving it the proposal's Problem, Root cause chain, and Concrete change plus the target file. Present the draft to the user as an explicit accept/reject before it's considered applied — a proposal already reviewed once (in the project it came from) still gets reviewed again here, because this is where it actually changes shared behavior. Once accepted, validate it:

  ```bash
  SC=~/.claude/plugins/cache/claude-plugins-official/skill-creator/unknown/skills/skill-creator
  uv run --quiet --with pyyaml python "$SC/scripts/quick_validate.py" bit/skills/<name>/
  claude plugin validate ./bit
  ```

  The two disagree on one known point: `quick_validate.py` flags every `bit_*` name as not-kebab-case, which is noise (the plugin namespaces slash commands by directory, not by the `name:` field) — `claude plugin validate ./bit` passing is what actually matters there. For body/prose changes, that structural check alone doesn't prove the new instruction fires — lean on skill-creator's own judgment about whether an eval pass is warranted for this change.

**2. Code or CLI change.** The proposal is really about something bit-pro itself needs to *build* — a `bp` feature, a plugin-bundled hook or MCP server bit-pro should ship, anything that isn't just prose. Don't implement it here. Hand off to **bit_scope**, using the proposal's `Problem` and `Root cause chain` as the new feature's WHY, so scope → plan → do runs the normal way. Tell the user you're doing this and why it isn't a direct edit.

**3. Neither, or genuinely unclear.** Some proposals won't cleanly be either — say so plainly, quote the proposal, and ask the user how they want to handle it rather than guessing which bucket it belongs in. Forcing an awkward fit here is worse than admitting the proposal needs a human call.

## Presenting decisions

Work through proposals as a sequence of accept/reject/discuss decisions, the same shape bit_retro uses. A one-line restatement of the proposal's title isn't enough — the user is often seeing this proposal for the first time here (it may have been accepted in a different project, by a different session, a while ago), and they can't judge accept/reject/discuss without actually understanding what recurred and why. Surface the substance, not just a label:

> **Proposal:** [the `Problem`, in plain language — what actually recurred]
> **Why it happened:** [the `Root cause chain`'s bottom line — the actual missing-or-unenforced mechanism it names, not a restatement of the symptom]
> **The fix:** [the `Concrete change` — or its gist if long, but don't compress away what it would actually make the skill/CLI do differently]
> **Bucket:** skill-text change (bit_\<name\>) | code/CLI change → bit_scope | unclear
> **Accept / Reject / Discuss?**

Only act once the user has answered. A rejected proposal just stops here — nothing to record, nothing to carry further.

---

## What this skill does not do

- **Diagnose patterns from raw feedback notes** — that's bit_retro; by the time a proposal reaches here it's already been generalized and reviewed once.
- **Edit a skill directly** — always through skill-creator, never a hand-edit to a `bit/skills/*/SKILL.md`.
- **Implement a CLI or plugin feature inline** — always handed to bit_scope, never built in this skill's own pass.
- **Trust a proposal's suggested mechanism blindly** — always sanity-checked against bit-pro's current state first.
- **Bulk-apply proposals without review** — one at a time, same as bit_do's one-bar-at-a-time discipline.
