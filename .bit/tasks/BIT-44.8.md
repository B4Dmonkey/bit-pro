---
id: BIT-44.8
title: bot routes new planning work to bit:ruler
status: todo
approved: true
phase: 3
phase_label: Ruler
---
## **Verse 3**

Operators who start in `bit:bot` need to find the ruler. bot's routing table sends new planning work straight to `bit:scope`, which skips analyze.

## Scope
- `bit/agents/bot.md`: in the routing table, add a row for starting new work that has no track yet, pointing the operator to relaunch as `claude --agent bit:ruler` (bot can't switch agents itself). Keep the `bit:scope` row for refining an existing track.

## Method
- [ ] Edit the routing table and any prose right below it that lists the routes.

## Claude verifies
- [ ] `claude plugin validate ./bit`

## User verifies
- [ ] Run `claude --plugin-dir ./bit --agent bit:ruler` and describe a small real change. Check the following:
  - A stub track appears in `bp task list`.
  - `.bit/research/<id>/index.md` gets written.
  - A scope body is written, and the session stops for your approval before planning.
  - Asking a question at the gate triggers another analyze pass before the scope is revised.
  - Approving runs bit:plan and then stops without starting bit:do.
- [ ] In `claude --plugin-dir ./bit --agent bit:bot`, say "I want to plan a new feature". bot points you to `claude --agent bit:ruler`. **Whole slice:** describing work gets you analyze → scope → gate → plan without running each skill by hand.

## Commit (user)
`feat(bit): route new planning work to bit:ruler`