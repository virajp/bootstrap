# U4 — Docs

- **Read first:** the DOCS FALSIFIED lines that U1 and U2 returned, and
  `docs/design/cli.md` (`detect`, `check`, `config`, `describe`).

## Ruling

- **D13:** update `docs/design/`, `README.md`, `CLAUDE.md` and `docs/sync/`
  where they describe `repo.toml`, `tools.toml`, `detect`, `check` or
  `describe`.
- In `docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md`, the ST2
  line says: each repo lists its projects in `[projects.<name>]`; each tool
  has a `projects` list of those names; `bootstrap detect` proposes both;
  `bootstrap check` verifies them.
- **D20:** in each `repo/tools/*/scripts/detect` that says bootstrap detect
  runs it in the root, change that comment only: it runs in the root, and in
  each project folder. Change no code line of a script.

## Verification

- `mise x -- mise run code:format` passes.

## Guardrails

- Touch nothing outside the owned paths.

## Commit

`docs: projects and tool links in bootstrap`
