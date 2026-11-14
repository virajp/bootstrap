# U8 — Docs

- **Read first:** the branch delta, then each passage named below.
- **Lazy-load:** `${CLAUDE_PLUGIN_ROOT}/skills/docs-sync/SKILL.md`.

## Ruling

D1, verbatim — a reversal of the design record's D2:

> 4.0.0 after 3.1.0 (Recommended): 3.0.0 (selection model) and 3.1.0 (TUI)
> ship as planned. R1 moves the values files, a breaking change, so it is
> MAJOR: bootstrap@4.0.0. R3 in claude-plugins waits for 4.0.0, not 3.0.0. The
> design record is amended.

D3, D5, D9, D10 and D11 are quoted in `02-migration-config-verb.md`; D7 in
`03-members.md`; D8 and D15 in `01-folder-schema.md`.

### Rulings review, round 1

These rulings override any earlier text of this file that they contradict.

- **D34** — Flat, as today: repo/examples/<x>/repo.toml and
  repo/examples/<x>/tools.toml. U6 (test/repo:2780) and U8 (CLAUDE.md:59) follow
  it.
- **D36** — U8 may fix every 3.0.0 reference in the record, the phase table and
  D2's text. bootstrap@4.0.0 is tagged after R1a, R1b, R1c and R2 land, because
  R2 renames tasks to git:*, also breaking. R3 waits for 4.0.0.

### Rulings review, round 2

These rulings override any earlier text of this file that they contradict.

- **D39** — D21 wins: 0 for any valid key (set or default), 1 for a read
  failure, 2 for an invalid key.

### Rulings review, round 3

These rulings override any earlier text of this file that they contradict.

- **D69** — bootstrap@4.0.0 is tagged at once after R1a lands, as the selection
  model does for 3.0.0. R1b, R1c and R2 become minor releases (4.1, 4.2, 4.3);
  R2 keeps the old task names as aliases for one release, as the design says, so
  it does not break. R3 waits for the release that carries the contract. D36 and
  D40 are superseded on the tag.

### Rulings review, round 4

These rulings override any earlier text of this file that they contradict. D20
no longer has a version check (D84).

- **D91** — cli.md and the review focus also name D24's exit 2 when
  .config/bootstrap/ is absent.

## Edits

1. Run `vwf:docs-sync` over the branch delta, and apply its findings plus every
   `DOCS FALSIFIED:` line of the earlier units.
2. **`docs/design/bootstrap-config.md`**: the title and every section become the
   folder (`repo.toml`, `tools.toml`, `releases.toml`, `repo-generated.tsv`):
   "Summary of the keys" (per file, D15), "`bootstrap_commit`", "Old forms"
   (the folder migration), "The location of the file" (`BOOTSTRAP_VALUES` names
   the folder, D12), the schema rule (D8).
3. **`docs/design/cli.md`**: the command list and usage (`config
   get|set|unset|show`,
   `settings`), "The config file" (now `settings`), "The engine variables",
   "The manifest and the stale rule", "The values of a tool",
   "`bootstrap_commit`",
   "The output of a run", "The exit codes" (code 2 for the old file and for an
   invalid key, code 0 with the default for `config get` of an unset key, D21
   and D39).
4. **`docs/design/repo.md`**: "The values file and the selection", "The branch
   model and the merge tasks" (`core_branches`, `landing`), members (siblings,
   `[home]`), "`bootstrap check`", "The manifest".
5. **`docs/design/index.md`, `mise.md`, `packs.md`, `mac.md`**: each mention of
   the old path.
6. **`docs/sync/migrate.md`**: a new migrate item after the selection model's
   item, for the folder migration (run `bootstrap init`; the key moves of
   D15 and D11; the overrides merge; `releases.yaml` stays until R3).
7. **`docs/sync/vwf.md`**: the plugin contract names `bootstrap config` for
   every read and write of a repo value, and the new key names
   (`[branches.<b>].landing`, `[commits].scopes`, `[members.*]`). It names no
   file path inside `.config/bootstrap/` as something a plugin reads.
8. **`docs/sync/index.md`, `docs/sync/tools/*.md`, `docs/setup/`**: each mention
   of the old path or the old `config` command.
9. **`README.md`, `CLAUDE.md`**: each mention of the old path, the old `config`
   command and the members keys.
10. **`docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md`**: add a
    section "Amendments 2026-11-14 (R1a planning)" that records D1 (4.0.0
    after 3.1.0; R3 waits for 4.0.0), D3 (the old `config` command becomes
    `settings`), D5 (the overrides merge into `tools.toml`), D10
    (`releases.yaml` moves in R3), D2 (R1 is three plans) and D69
    (`bootstrap@4.0.0` is tagged at once after R1a; R1b, R1c and R2 ship as
    4.x minors, R2 with alias task names; R3 waits for the release that
    carries the contract). Fix the phase table and every 3.0.0 reference to
    match (D36). Do not rewrite the rest of the record.

## Verification

- `mise x -- mise run code:format` passes.
- `grep -rn 'bootstrap.toml\|bootstrap.overrides.toml' README.md CLAUDE.md
  docs/design docs/sync docs/setup` finds only the old-form and migration
  passages that name the old file on purpose.

## Guardrails

- Touch nothing outside the owned paths. `docs/plans/` is not this unit's.
- Delete with `rm`, never `git rm`.

## Commit

`docs: the .config/bootstrap/ folder, the config verb and members`
