# U5 — Docs

- **Read first:** the DOCS FALSIFIED lines that U1 to U3 returned,
  `docs/design/tools.md` (the `[tasks]` table), and the claude-plugins copy of
  the split record:
  `/Users/virajpatel/Projects/github.com/virajp/claude-plugins/docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md`,
  section "Amendment 2026-11-15 (R5 brainstorm)".

## Ruling

- `docs/design/tools.md`: the `replaces` key of a `[tasks]` entry (D4, D5).
- `docs/design/` and `docs/sync/`: the three parts, their environment values
  and their "does not apply" rules (D1-D3, D6); `docs/sync/migrate.md` gets
  one line: `bootstrap init` removes the old stackgen copies of these tasks.
- `README.md` and `CLAUDE.md` where they list what a tool ships.
- `docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md`: add the
  section "Amendment 2026-11-15 (R5 brainstorm)" as the claude-plugins copy
  has it, word for word, except its background path, written as
  `claude-plugins:
  docs/memory/gaps/2026-11-15-r5-stack-without-packs-resolved.md`
  (D27); then one line under it: the three test tasks move to
  the `android` and `swift` tools in this plan (R2e; claude-plugins R5a D8).
- `docs/memory/gaps/2026-11-14-pack-test-parts-parked.md` and
  `2026-11-14-r5-test-parts-owner-parked.md`: rename each to `-resolved.md`
  with `mv`, and change the head line's `parked:` part to
  `resolved 2026-11-15 by docs/plans/2026-11-15-test-parts`. **D18:** the
  Related line of the second record names the first by its `-resolved` name.

## Verification

- `mise x -- mise run code:format` passes.

## Guardrails

- Touch nothing outside the owned paths.

## Commit

`docs: the tool test parts, the replaces key, and the R5 amendment`
