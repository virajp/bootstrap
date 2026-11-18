# U5 — Docs

- **Read first:** the DOCS FALSIFIED lines that U1 to U3 returned, then the
  passages of "Facts the survey established" in `index.md`.
- **Lazy-load:** `docs/setup/index.md` (the Mac adoption prompt).

## Ruling

- **D10:** "uv becomes a required member of the Mac base pack: uv = { required =
  true } in macos/packs.toml [...]. [...] the Mac mise tool gets depends =
  [\"cli\", \"python\", \"uv\"]: then both are in the init stage."
  `docs/setup/index.md`, `docs/setup/tools.md`
  and `docs/design/packs.md` say that uv is required on the Mac;
  `docs/design/mac.md` "Stages and exit 4" says that `mise` depends on `cli`,
  `python` and `uv`.
- **D1, D2, D3, D10:** `docs/setup/tools.md` gives the `.py` task files of
  each Mac tool, the Mac `helpers.py` and its import preamble, and that each
  Mac task brings uv and Python 3.14 through mise.
- **D5:** the docs name no shell Mac helpers. A Mac needs no hand step: the
  stale rule of the Mac sync removes the old copies, so no migrate item.
- `docs/design/mac.md` (about lines 67-68) and CLAUDE.md: the entry task
  `updateall:configs` is Python; step 1 says "directly, with uv on PATH" in
  place of "under bash" (D11). `docs/setup/tools.md` names the ruff plugin
  of the Mac dprint config (D13).

## Edits

1. Run `vwf:docs-sync` over the branch delta, and apply its findings, every
   DOCS FALSIFIED line, and the passages of the survey.

## Verification

- `mise x -- mise run code:format` passes.
- `grep -rnE '_scripts/helpers([^.]|$)' docs/setup docs/design/mac.md` finds
  no passage that names the shell Mac helpers.

## Guardrails

- Touch nothing outside the owned paths.
- Do not edit a plan folder in `docs/plans/`.
- Write each changed passage in ASD-STE100.

## Commit

`docs: the Mac tasks in Python`
