---
id: BIT-50.16
title: Two bars that share a subject can't both claim one squash line
status: todo
phase: 4
phase_label: squash-landed tracks
---
## **Verse 4**

The scope says a shared generic subject can't match, because each bar needs its own matching line. With BIT-50.15's rule, two bars both subjected `chore(bit): tidy` both match a squash that lists the line once. That test forces each squash's lines to be claimed once.

## Scope
- `landing/landing.go`:
  - Bars are matched in bar order. Each squash keeps a count of each listed subject, plus one for its stripped `Subject`. A match decrements the count, and a bar skips a squash whose count for its subject is zero and moves on to later squashes.
  - A bar left unmatched keeps its `Pushed` or `Local` class.
- Test file: `landing/landing_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestCheck/two bars with one shared subject claim one line each`
     - **Behavior:** a generic subject listed once can't land two bars.
     - **Setup:** on `feat`, commit `t1 := r.Commit("chore(bit): tidy")` and `t2 := r.Commit("chore(bit): tidy")`, then push `feat`. On `main`, squash with body `"* chore(bit): tidy"` once, as `"Tidy (#8)"`, then push. Bars `{t1}` and `{t2}`.
     - **Assertions:** bar 1 is `Squash`, bar 2 is `Pushed`, and `Verdict == Partly`.
     - **Boundary:** a subject listed fewer times than bars carry it.
   - [ ] Confirm fails: both bars are `Squash`.

2. **Implement (GREEN):**
   - [ ] Per-squash line counts.

3. **More tests (RED → GREEN):**
   - [ ] `TestCheck/a subject listed twice lands both bars`: the same setup with the line listed twice gives both bars `Squash`. *Boundary:* the count equals the bars.
   - [ ] `TestCheck/a bar passed over moves to a later squash`: a later squash `"Tidy again (#9)"` lists the line once, so bar 2 lands there and `Landing` = `#9`. *Boundary:* the claim spills to a later squash.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit
`feat(bit): each squash line lands at most one bar`