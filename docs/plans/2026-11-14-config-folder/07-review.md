# U7 — Review: the engine, the tool scripts, the examples and the tests

## Scope

U1 to U6, the whole branch delta since the branch base. The engine is Go code
that ships, and the tool scripts are bash that runs in every consumer repo
(D14). Focus: the old-form detection and the migration (no data loss of a
moved key or comment; an already-migrated repo is a no-op), the schema
(`check` refuses unknown keys), the `config` verb's exit codes (0 for any valid
key, set or default; 1 for a read failure; 2 for an invalid key; D39), the
releases rule, the back-link write into another repo,
and every `bootstrap config get` call in `repo/tools/` (an absent key must not
break a task).

### Rulings review, round 2

These rulings override any earlier text of this file that they contradict.

- **D39** — D21 wins: 0 for any valid key (set or default), 1 for a read
  failure, 2 for an invalid key.

### Rulings review, round 4

These rulings override any earlier text of this file that they contradict. D20
no longer has a version check (D84).

- **D91** — cli.md and the review focus also name D24's exit 2 when
  .config/bootstrap/ is absent.
