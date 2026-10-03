---
paths:
  - "**/*_test.go"
---

# Go tests

- **One test function per unit under test.** The unit is a function, method or command:
  `TestOpen`, `TestAddCmd`, `TestResolve`. Behaviours are cases inside it, never part of
  the function name. `TestOpen_MigratesAFreshDatabase` is wrong; `TestOpen` with a
  `migrates a fresh database` case is right.
- **Prefer a table test.** Use one when the cases share setup and assertions and differ
  only in data: a `[]struct{ name string; ... }` looped with `t.Run(tt.name, ...)`.
- **Otherwise use subtests.** When setup or assertions differ per case, write
  `t.Run("migrates a fresh database", func(t *testing.T) { ... })` inside the unit's test.
- **Case names are short lowercase phrases** describing the behaviour. Plans name a test by
  its `go test -run` path: `TestOpen/migrates a fresh database`.
- **New cases join the unit's existing test.** If `TestOpen` exists, add a row or subtest
  to it rather than a new top-level function.
- **This applies to existing tests too.** Flat-named tests (`TestListCmd_PrintsProjectsByCode`)
  are wrong. A change that touches a test file converts every test in that file to this
  shape first, as a pure restructure: the same cases and assertions, still green, before
  any new test is written. Test files the change doesn't touch are left alone.
- **Removals get no new tests.** Code that's only being deleted is deleted, and the
  existing tests that break are fixed or removed. Tests are for new or modified behaviour.
- **No comments in tests**, including arrange/act/assert markers.
