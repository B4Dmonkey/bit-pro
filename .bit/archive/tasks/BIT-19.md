---
id: BIT-19
title: Stop-hook reminder for uncommitted bar close-outs
status: todo
---
## Why
Bar close-outs in bit_do end with a stated commit message, but that rule lives only in
prose re-derived from memory each reply — a few small follow-up replies deep, the model
sometimes drops it. The two facts that would make this checkable already exist separately
(a bar's `doing`-with-uncommitted-work state via `bp`/`git`, and the reply's own text) but
nothing currently cross-checks them.

## Summary
Ship a `Stop` hook in bit-pro's plugin that, when a bar looks closed-out-but-uncommitted
and the reply doesn't contain a commit-message-shaped block, injects a reminder via
`additionalContext` rather than blocking the turn.

## Risks & unknowns
- **Unknown:** `bit_do` never defines a fixed, mechanically-checkable shape for "the
  suggested commit message" — right now it's freeform prose, so a hook has nothing
  reliable to grep for.
  **Resolve by:** decide a minimal literal marker `bit_do` must use going forward (e.g. a
  fenced block introduced by a fixed line like "Suggested commit:"), so the hook's check
  and `bit_do`'s prose agree on the same shape.
  **De-risk before planning?** Yes — this decision shapes both the hook's detection logic
  and a companion edit to `bit_do`'s close-out section; get it settled before bit_plan
  writes steps for either.

- **Unknown:** false positives vs. false negatives — a bar can legitimately sit `doing`
  for reasons unrelated to the current reply (e.g. genuinely awaiting a separate
  User-verifies confirmation from turns ago), so "any doing bar" is too broad a trigger.
  **Resolve by:** decide the tolerance now rather than mid-plan — this feature exists to
  catch a forgotten reminder, so an occasional unnecessary nudge is cheap; silently never
  reminding is the actual failure mode being fixed. Bias detection toward reminding too
  often over too rarely.
  **De-risk before planning?** No — this is a stance to state as a Decision, not something
  that needs a spike to resolve.

## Decisions
- **Use the `Stop` hook event, not `PostToolBatch`.** `PostToolBatch` fires before the
  model's reply exists and never receives reply text; `Stop` fires after the reply, receives
  `last_assistant_message`, and supports non-blocking `additionalContext` — the exact shape
  of "remind, don't block."
- **Detect via a combination of git and `bp`, not transcript parsing.** The hook shells out
  to `git status --porcelain` for uncommitted changes and `bp task list` for any bar in
  `doing` status; both are already-exposed, script-friendly signals, cheaper and more
  reliable than parsing the hook's own conversation transcript.
- **Reminders are non-blocking.** The hook only ever injects `additionalContext`; it never
  returns a blocking `decision`. A hook that hard-blocks the turn on a heuristic this fuzzy
  would be worse than the problem it fixes.

## Verses
- [ ] Verse 1 — A `Stop` hook fires and can tell, mechanically, whether a bar close-out
  looks uncommitted: given a repo with uncommitted changes and a bar in `doing`, the hook
  detects the condition (observable via the plugin's hook debug log).
  Touches: `bit/hooks/hooks.json` (new), a new hook script under `bit/hooks/`.
- [ ] Verse 2 — The hook actually reminds when the reply is missing the commit-message
  marker: given the condition from Verse 1 plus a reply lacking the fixed marker, the
  hook injects a visible reminder into the next turn.
  Touches: same hook script; `bit/skills/do/SKILL.md`'s close-out section (adds the fixed
  marker convention decided above).
- [ ] Verse 3 — The hook stays quiet on a real close-out: given a reply that *does* contain
  the marker, or a repo with no uncommitted changes, the hook injects nothing.
  Touches: same hook script.