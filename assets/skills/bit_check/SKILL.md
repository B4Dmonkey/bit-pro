---
name: bit_check
description: Post-implementation audit that reviews completed work against the plan, PR feedback, and code quality signals. Produces a structured findings list for the retro skill and routes code-change items back through bit_scope → bit_plan → bit_do. Use this skill whenever the user says "bit_check", "check the work", "audit the implementation", "review what we did", or after completing a bit_do cycle and wanting to verify everything landed cleanly. Can be run multiple times — re-running after fixes updates the findings list.
---

# Post-Implementation Audit

You audit completed work and produce a structured findings list. You are the quality gate between "implementation done" and "ready for retro." Your job is to surface what was missed, what needs fixing, and what went well — then route each finding to the right place.

You do **not** implement code. When something needs a code change, you capture it as a finding and route it to bit_scope for the normal pipeline (scope → plan → do).

**Before you drive the CLI, read `.claude/bit-cli.md`** — the shared command contract. The scope and plan now live as a **track** and its **bars** in `.bit/`, not as paired root files; you find and read them through `bp`.

---

## Inputs

You need:
1. **The track** — the scope-and-plan, found by ID or title. The user usually names the work; resolve it with `bp task list` (tracks are the rows whose ID has no dot). Its body holds the scope + phases; its bars (`bp task list --parent <track>`) are the plan steps.
2. **The PR** (optional) — if one exists on the current branch, use it for checks 4-5

There's no `<feature>-scope.md` / `<feature>-plan.md` filename pairing to hunt for anymore — the track *is* the pairing. If more than one track has recent work and it's unclear which to audit, list what you found and ask.

---

## Order of checks

Run these in order. Check 1 is a gate — if the work is genuinely incomplete, stop there.

### 1. Plan completion (gate)

List the track's bars (`bp task list --parent <track>`). For every bar:
- Verify its status is `done`
- Cross-reference `git log --oneline` to confirm a commit exists that corresponds to each bar's suggested message (read the bar body with `bp task read <bar> --body` for it; fuzzy match on the track ID and key words — don't require exact wording)

**If genuinely incomplete** (bars are not `done` with no corresponding work in git): stop, report which bars are unfinished, and inform the user. They likely triggered the skill early.

**If the work landed but the status is stale** (commits exist but a bar isn't `done`): set it with `bp task update <bar> -s done` and record this as a finding ("status cleanup — marked bar N done").

### 2. Scope/plan sync

The scope phases live as a checklist in the **track body**; the bars carry the phase they serve (`--phase`). Reconcile them:
- Every phase whose bars are all `done` should be checked off (`- [x] Phase N`) in the track body
- Every phase with an unfinished bar should be unchecked
- If they're out of sync, fix the track body (`bp task update <track> -d "…"`) and record a finding

### 3. Commit message hygiene

Check each commit on the current branch (vs main) for:
- References the ticket ID (e.g., `PTC-2823`)
- Follows conventional format (`feat(...)`, `test(...)`, `fix(...)`, etc.)

Minor deviations are just findings, not blockers.

### 4. CI/PR checks (skip if no PR)

Run `gh pr checks` (or equivalent). Report:
- All passing → no finding
- Any failing → finding with the check name and status. No deep debugging — just surface it.

### 5. PR comments (skip if no PR)

Fetch all PR comments (including bot comments — bots sometimes catch real bugs). For each actionable comment, categorize:

| Category | What it means | Severity |
|----------|---------------|----------|
| **Bug introduced** | Reviewer or bot identified a defect | high |
| **Change requested** | Reviewer wants something different before merge | high |
| **Change suggested** | Reviewer proposes an improvement, not blocking | medium |
| **Nitpick/style** | Small style or preference note | low |

Include all categories in the findings — nitpicks get escalated to the user for a judgment call (they may scope it or push back).

Skip: resolved threads where the fix is already committed, pure "LGTM" comments, and non-actionable praise.

### 6. Refactor signals (light pass)

Scan the files touched by this branch's commits for:
- TODO/FIXME comments left behind
- Duplicated code patterns introduced (same block appearing 2+ times)
- Functions that grew past ~50 lines during this work

Flag as "refactor opportunity" findings. No deep architectural analysis — this is a surface scan.

---

## What went well

After the checks, note things that worked:
- Steps that landed without rework
- TDD catching bugs before they shipped
- Clean PR with no comments
- Any other signal that the process worked

Keep it brief — a few bullet points. This feeds the retro skill with positive signal about what to keep doing.

---

## Output format

Write `<track-id>-check.md` in the repo root (e.g. `BIT-7-check.md`). The check report is a transient audit artifact the retro skill consumes — it is *not* tracked work, so it stays a plain file rather than a bp task:

```markdown
# Check: <track title>

Track: <track-id>
PR: [#N](link) (or "no PR")
Date: YYYY-MM-DD

## Summary
- Total findings: N
- Needs code work: N (→ bit_scope)
- Retro only: N
- Blocking: Y/N

## Findings

### 1. [category] Short description
- **Severity:** bug | requested-change | suggested-change | nitpick | missed-step | doc-cleanup | commit-hygiene | refactor-opportunity
- **Source:** plan-audit | scope-sync | commit-check | ci-check | pr-comment (by @reviewer) | pr-comment (bot: name) | code-scan
- **Detail:** What happened / what's wrong
- **Action:** scope-it | note-for-retro | already-fixed
- **Status:** open | resolved

### 2. ...

## What went well
- ...

## Handoff

### Route to bit_scope → bit_plan → bit_do
These findings need code changes:
- Finding #N: <summary>
- ...

**Next steps:**
1. Run `bit_scope` in refine mode to add fix/cleanup phase(s)
2. Run `bit_plan` to create TDD steps for the new phase(s)
3. Run `bit_do` to execute

### Route to retro
- Finding #N: <summary>
- ...
- What went well: <bullets>
```

If there are no findings that need code work, omit the "Route to bit_scope" section. If all findings are resolved (on re-run), say "All clear — nothing remaining."

---

## Re-run behavior

When you find an existing `<track-id>-check.md`:
1. Read it
2. Re-run all checks
3. Mark previously-open findings as `resolved` if the issue no longer exists
4. Add any new findings
5. Update the summary counts
6. Report: "N findings resolved since last run, M remaining, K new"

---

## Routing logic

**Needs code work → bit_scope:**
- Bug introduced
- Change requested
- Change suggested (if user decides to do it)
- Refactor opportunity (if user decides to do it)

**Retro only:**
- Doc cleanup (already fixed during this check)
- Scope/plan sync fix (already fixed during this check)
- Commit hygiene notes
- CI failures (user likely already handled)
- Nitpicks (after user makes the call — if rejected, it's retro-only)
- What went well

**User decides:**
- Nitpicks — present them and ask: "scope it or reject?"
- Suggested changes — present them and ask: "scope it or skip?"

For items the user needs to decide on, present them at the end and wait for their call before finalizing the routing in the check file.

---

## How to gather context

1. Find the track (by user direction, or `bp task list`) and its bars (`bp task list --parent <track>`); read the track body and each bar body via `bp task read <id> --body`
2. `git log --oneline main..HEAD` for commits on this branch
3. `gh pr view --json number,url,state` to find the PR (if any)
4. `gh pr checks` for CI status (if PR exists)
5. `gh pr view --comments` or `gh api` for PR comments (if PR exists)
6. `git diff --stat main..HEAD` for files touched, then read those files for refactor signals
