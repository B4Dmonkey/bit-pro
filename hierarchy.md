# Hierarchy and naming

The vocabulary `bit` uses to talk about work. This is **presentation only** — it's how
the docs, the CLI's output, and the TUI name things. The code keeps saying `task`,
`scope`, and `step`.

This is a live proposal, not settled doctrine. It's here so future work has one agreed
set of words instead of inventing them twice. Expect it to move until the shape stops
changing; eventually these meanings become documentation reachable from `bit --help`.

## The map

| Word      | What it is                        | In the code      | Addressable? |
|-----------|-----------------------------------|------------------|--------------|
| **album** | the project — one `.bit/` directory | the `Store` root | no — it's the container |
| **track** | one deliverable: a scope, and the work under it | a task with no dot in its ID (`BIT-2`) | yes |
| **verse** | a phase — a coarse group of bars inside a track | a label on a step | no — see below |
| **bar**   | one step: the smallest unit of work, one commit | a task with a dot in its ID (`BIT-2.5`) | yes |

## How it fits together

```
.bit/                album   the project
├── BIT-1.md         track   CLI Bootstrap
├── BIT-1.1.md       bar     phase: 1
├── BIT-2.md         track   Task Management (CRUD)
├── BIT-2.1.md       bar     phase: 1 — init wizard + create
├── BIT-2.6.md       bar     phase: 2 — list & read
└── BIT-2.13.md      bar     phase: 4 — delete
```

A bar's ID names its track: `BIT-2.5` is the 5th bar of `BIT-2`. The parent is readable
straight from the ID — no index, no lookup, and `ls` shows the tree.

## Two things worth knowing

**A verse isn't a record.** Phases group bars, but you never open a phase — you approve a
whole track, or you review one bar at a time. So a verse is a label a bar carries, not a
file on disk. "The plan for BIT-2" isn't a thing either; it's just BIT-2's bars.

**A dot is what makes a bar a bar.** There's no type field. No dot means track, a dot
means bar. If a third kind of thing ever needs to exist, that's when a real field earns
its place.
