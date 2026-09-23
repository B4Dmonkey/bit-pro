---
name: bit_analyze
description: Do the deep codebase research a track needs before it is scoped or re-scoped, and leave the findings as research notes under `.bit/research/<track>/` through the `mcp__bit__research_*` tools. Use whenever the user says "analyze", "/bit:analyze BIT-N", "research this track", "deep dive before scoping", "dig into the code for this", "validate the scope's assumptions", or when bit:ruler dispatches a research pass with open questions. It fans out explorer subagents across the code areas the track's verses touch, checks the assumptions the scope rests on, confirms or corrects each "Touches" pointer, and surfaces unknowns that light research missed. It writes one note per topic plus an `index` topic that summarizes the findings and links to the others. It does not edit the track body, write the scope, or plan bars. Shaping the scope is bit_scope's job, and it reads these notes afterward. Reach for this, not bit_scope, when the work needs evidence from the code before anyone decides its shape.
---

# Track Analysis

You do the deep research that bit_scope deliberately skips and bit_plan says it isn't for. A scope written on light research rests on assumptions nobody checked: that an API returns what we think, that a file owns what its name suggests, that a pattern exists to copy. Those assumptions surface later as surprises in bit_do or as plan-to-scope hand-backs, and by then they're expensive. Your job is to check them now, against the code, and write down what you found so the next agent doesn't have to rediscover it.

What you produce is **research notes**, not a document for the operator. The notes are an agent scratchpad. Their readers are bit_scope, bit_plan, bit:ruler, and later analyze passes, which open the `index` topic first and load only the topics they need. So the format is loose. What matters is that each note is true, specific (paths, function names, what the code actually does), and findable from the index.

Three tools cover everything this skill does:
- `mcp__bit__task_read` reads the track (its body holds the scope's WHY, verses, and "Touches" pointers).
- `mcp__bit__research_read` with only `track` lists the topic names that already exist. With `topic`, it returns that topic's body.
- `mcp__bit__research_write` writes one topic. Writing a topic that exists replaces it.

Research lives under `.bit/research/<track>/`, but you never read or edit those files directly. The tools are the only way in, which keeps a stray write from landing in the wrong track or outside it.

## Input

A track ID (uppercase it: `bit-44` → `BIT-44`), plus any **open questions** the caller hands over. bit:ruler passes the questions the operator raised at the scope gate. A user running `/bit:analyze` directly may pass none. In that case, the scope's own Risks, unknowns, and "Touches" pointers are the questions.

## 1. Read what exists

1. `mcp__bit__task_read` on the track. Take the `body`. Note each verse's intent and "Touches" pointers, every assumption a Decision leans on, and every open risk or unknown.
2. `mcp__bit__research_read` with only the track. No topics (an empty result) means this is the first pass. Otherwise, read `index` and the topics the open questions touch.

Earlier notes tell you where someone looked. They are not settled truth. If a topic bears on an open question, **re-check it against the code and rewrite it**. Don't carry its conclusion forward unchecked, because the code may have moved and the earlier pass may have been wrong. That is usually why the question is open again. Topics no question touches can stay as they are.

## 2. Fan out explorers

Split the work by code area: one explorer subagent per area the verses touch, plus one per open question that doesn't map to a single area. Run them in parallel, since they're independent. Give each one:
- the area or question, and the verse it serves;
- the specific assumptions to check ("the scope assumes `task/feedback.go` builds paths with pathologize. Confirm and show the function");
- the "Touches" pointers to confirm or correct;
- a request to report facts with file paths and line references, and to flag anything surprising: a missing piece, a second place that does the same thing, a constraint the scope didn't mention.

Explorers report back to you. You write the notes yourself, because the tools and the index structure are yours to keep consistent.

## 3. Write one topic per area

Call `mcp__bit__research_write` once per topic. A topic name is a short plain word or phrase: `store`, `mcp-server`, `feedback-pattern`. It can't contain `/`, `\`, or `..`, because a topic is a single note, not a path. A useful note says:
- **What was checked**: the assumption or question, in one line.
- **What the code does**: paths, function names, short excerpts where the exact text matters.
- **Verdict**: confirmed, corrected (and what's actually true), or still unknown (and what would settle it).
- **Pointer corrections**: a "Touches" entry that's wrong or missing, and what it should say.

Keep it to what a later agent needs in order to act, not a transcript of the exploration.

## 4. Write `index` last

Once every topic is written, write the `index` topic: a short summary of the findings, and a link to each topic (`[store](store.md)`), each with a one-line reason to open it. Put the findings that change the scope first: assumptions that turned out false, new unknowns, and corrected pointers. Readers open the index first and decide from it what to load, so it has to say what's in each topic, not just list the names.

Write it last because it summarizes the topics. Writing it first means summarizing notes that don't exist yet.

## 5. Report back

Tell the caller, in a few lines, what changed: assumptions confirmed, assumptions corrected, new unknowns, and pointers fixed, each naming the topic that holds the evidence. That is the whole hand-off. bit_scope or the operator decides what to do with it.

## What this skill does not do

- **Edit the track body.** No Decisions, no Risks, no verse changes. The scope belongs to bit_scope, which reads your notes and cites them. If a finding means the scope is wrong, say so in the report and let bit_scope change it.
- **Plan bars or write code.** Research informs those steps. It doesn't do them.
- **Read or write `.bit/research/` directly.** The `mcp__bit__research_*` tools are the only way in.
- **Keep history.** Rewriting a topic replaces it, and there's no delete. The notes are a scratchpad, so the current version is the only one that matters.
