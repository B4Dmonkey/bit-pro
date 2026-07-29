---
id: BIT-17.6
title: One sentence becomes a well-formed note
status: doing
phase: 2
phase_label: The skill composes it
---
## **Verse 2**

`/bit:feedback` turns "that's wrong, do X instead" into a well-formed note, so recording one costs a
sentence rather than a writing exercise. Skill text, not code — no red-green cycle, so the checks
are greps plus a real session, the same shape BIT-16.5 used for a skill port.

## Scope
- `bit/skills/feedback/SKILL.md` — new. The directory name is what namespaces it, so `feedback/`
  gives `/bit:feedback`; frontmatter `name:` is `bit_feedback`, matching how `do/` declares
  `bit_do`.

Frontmatter: `name` and a `description` that names the trigger phrases the user actually says — "capture
that", "record that", "note that for the retro", "that's wrong, do X instead" — and states plainly
that it records an observation and does **not** evaluate, classify, or fix anything.

The body covers:

- **Read the contract first** — run `bp instructions`, as every bit skill does. Every write into
  `.bit/` goes through `bp`; never hand-write a file under `.bit/feedback/`.
- **The trigger question**: *did I make a decision the plan didn't make for me?* Sharper than "note
  any problems," and it converges with the goal — every unspecified decision is either a plan gap or
  something too trivial to plan, and a plan that leaves nothing to decide is one a small model can
  one-shot.
- **Find the track.** The note keys to the track, not the bar. Resolve it from the session — the
  work in flight — or with `bp task list` when it is ambiguous; ask rather than guess, because a
  note filed under the wrong track is a note retro cannot use.
- **What the note says**, in three parts and nothing else: where it happened (cite the bar as prose
  — "happened at BIT-11.4"), what the plan said, and what the work actually turned out to require.
- **Observations only.** No cause, no blame, no proposed fix, no lesson. Right after being
  corrected, attribution skews toward blaming the artifact ("the plan was unclear") over the model's
  own choice ("I didn't read the file I was told to read"); facts are cheap and reliable in-flight,
  judgment is not. Classifying a note is retro's job.
- **Keep it short.** A few sentences. Capture has to be cheap enough to happen mid-run, and a note
  that takes ten minutes to write will not get written at the moment it is worth writing.
- **Write it** by composing the body in a file and passing `-d "$(cat note.md)"`, then report the
  path the command prints back to the user.
- **What this skill does not do**: evaluate notes, revise the scope or plan, or fix the code. If the
  correction also means the plan was wrong, that is a separate hand-back to bit_plan — recording the
  note does not repair anything, and conflating the two is what turned the old retro skill into
  reconstruction instead of evidence.

## Claude verifies
- [ ] `just test` and `just lint` — no Go changed, but the tree must stay green
- [ ] `claude plugin validate ./bit` exits 0; the missing-`author` warning is pre-existing and is
      not a failure
- [ ] `grep -n 'bp instructions' bit/skills/feedback/SKILL.md` matches — the new skill points at the
      contract by command, never at a path
- [ ] `find bit/skills -name SKILL.md` lists exactly five files

## User verifies
- [ ] Get the skill in front of a live session. The reliable route is the one BIT-16.3 proved: commit,
      push, then `claude plugin marketplace update bit-pro` and restart. `claude --plugin-dir ./bit`
      is the faster loop, but the installed `bit@bit-pro` plugin may shadow or collide with a
      second plugin of the same name — if `/bit:feedback` does not appear, use the push route rather
      than debugging it.
- [ ] `/bit:feedback` is offered alongside `/bit:scope`, `/bit:plan`, `/bit:do`, `/bit:check`.
- [ ] Whole slice: hand it **one sentence** about a correction that really happened, then open the
      file it wrote. It contains the track, the bar it happened at, what the plan said, and what the
      work required — four things you did not type. That gap between one sentence in and a
      well-formed note out is the whole capability.

## Commit (user)
`feat(plugin): add the bit_feedback skill`