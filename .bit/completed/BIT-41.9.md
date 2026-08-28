---
id: BIT-41.9
title: Ordering is numeric, and only behind fires
status: done
approved: true
phase: 5
phase_label: stale notice
---
## **Verse 5**

Contradiction: BIT-41.8 fires the notice whenever the two versions merely *differ*, which is wrong
in both directions. A machine whose install is *ahead* of the marketplace clone gets told to
update to an older version, and `0.9.0 → 0.10.0` is a case where a lexical comparison orders the
versions backwards. The scope's decision is that the notice fires when the install is **behind**
and stays silent otherwise, so this bar forces a real numeric comparison.

## Scope
- `cmd/root.go` — a `behind(installed, latest string) bool`; `execute` consults it instead of `!=`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestBehind` (table-driven)
     - **Behavior:** the notice fires only when the installed version is strictly below the latest,
       compared component by component as numbers, and never on a version string that is not a
       three-part number.
     - **Setup:** table of `(installed, latest, want)` pairs:
       `("0.1.0", "0.2.0", true)`; `("0.1.0", "0.1.0", false)`; `("0.2.0", "0.1.0", false)`;
       `("0.9.0", "0.10.0", true)`; `("0.10.0", "0.9.0", false)`; `("1.0.0", "0.9.0", false)`;
       `("4ebbe7cd5eff", "0.1.0", false)`; `("0.1.0", "", false)`.
     - **Assertions:** `behind(installed, latest) == want` for every row.
     - **Boundary:** the ordering relation across all three of its states — below, equal, above —
       at the minor position and again at the major; `0.9.0` vs `0.10.0` is the two-digit component
       where numeric and lexical ordering disagree; `4ebbe7cd5eff` is the real pre-versioning
       install string still recorded in `installed_plugins.json` today, the unparseable lower bound;
       `""` is the empty upper operand.
   - [ ] Confirm fails: `undefined: behind`. With `!=` substituted, the `("0.2.0", "0.1.0")`,
         `("0.10.0", "0.9.0")`, `("1.0.0", "0.9.0")` and `("4ebbe7cd5eff", "0.1.0")` rows fail; with a
         plain string `<`, the `("0.9.0", "0.10.0")` row fails.

2. **Implement (GREEN):**
   - [ ] `behind(installed, latest string) bool`: split both on `.`, require exactly three parts
         each, `strconv.Atoi` every part, and return false if any parse fails — an unparseable
         version means silence, per the never-block, never-wrong-fire decision. Compare major, then
         minor, then patch; return true only on the first strictly-less component.
   - [ ] `execute` calls `behind(installed, latest)` in place of `installed != latest`.

## Claude verifies
- [ ] `just test` passes, including BIT-41.8's two tests unchanged.
- [ ] `just lint` passes.

## User verifies
- [ ] none — deterministic.

## Commit (user)
`fix(version): compare versions numerically, notify only when behind`