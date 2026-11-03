# U1 — The Mac adoption prompt and the Mac tool reference

- **Read first:** `docs/sync/index.md` (the repo prompt, the model of the
  form), `docs/design/setup-sync.md`, `macos/examples/macos.toml`,
  `macos/packs.toml`, the entry task files of `macos/tools/macos-setup/`,
  the usage header of `sync-engine/setup-sync`, and
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/docs/tool-reference.md`.
- **Lazy-load:** each `macos/tools/<tool>/tools.toml` (for `tools.md`),
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/docs/setup.md`,
  `docs/design/packs.md`, `docs/design/tools.md`.

## Ruling

From the design rulings of 2026-11-04 (verbatim):

"Friends' forks (user, 2026-11-04): several friends derived their
`macos-setup` from the user's `dotfiles/` model and will move to the new
format. The new model exists to let others adopt without drift.
Mac docs (accepted): `docs/setup/index.md` (adoption + update prompt),
`docs/setup/migrate-dotfiles.md`, `docs/setup/tools.md` (today
`docs/tool-reference.md`; questions stay in `tools.toml`),
`macos/examples/macos.toml`. [...]"

"Values row, item 1: the Mac values live in the `macos-setup` repo, made
private, at its root (not iCloud). The iCloud config folder never
existed (the `tools/` model is not live), so no data moves. A new Mac
runs `mise x gh@latest -- gh auth login` in `setup` before the clone. The
engine reads a folder of that clone; `MACOS_SETUP_CONFIG_PATH` stays the
way to point at it."

"Values row, item 6: `bootstrap_commit` on both sides (Mac: in
`macos.toml`, replaces `macos_setup_commit`), same write rules. Each Mac
sync with a new commit changes `macos.toml` in the macos-setup clone; the
user commits it (`ops: sync bootstrap`). The Values row is closed."

"Values row, item 3: the shared `tools.toml` schema accepts `[defaults]`,
`[questions]` (`question`, `detect`, `example`; the guide asks, the
engine only validates) and `[contributes.<tool>]` on both sides. [...]"

From this plan:

D3: "From macos/tools/<tool>/tools.toml and macos/packs.toml: the
description, the packs that hold it and whether it is a default there, the
stage, depends, each file with its mode, template, format, requires and
remove keys, the contributions, the [mise.*] sections by name, and the key of
each question with a link to its [questions.<key>] table. The question text
stays in tools.toml only."

D4: "On a Mac that has mise and the macos-setup clone, the first run passes
the variable inline: MACOS_SETUP_CONFIG_PATH=<path of the macos-setup clone>
mise run updateall:configs. Later runs get it from the generated mise env. A
new Mac uses the parked change of the macos-setup setup script; the doc says
so and names the parked step. Each command is one that the landed code
supports."

D5: "An AI agent that works with the owner of the Mac, as docs/sync/index.md
does for a repository. Each owner decision is one question at a time. Each
step names the command and its pass condition."

D2: "No. The 16 plans form one linear chain through requires, because most of
them edit sync-engine/repo-sync or sync-engine/setup-sync, and parallel plans
conflict there. This plan requires 2026-11-04-docs-design-split."

D1: "No. The plan writes docs only; no runnable line changes."

D7: "docs/setup/index.md, docs/setup/migrate-dotfiles.md and
docs/setup/tools.md. The later "Mac docs" ruling wins over the older
placement bullet, which names sync.md and tool-reference.md."

D10: "A level-3 heading: ### `<tool>`."

## Edits

1. **`docs/setup/index.md`** (new) — the Mac adoption prompt for an AI agent
   (D5), in the form of `docs/sync/index.md`:
   - **Requirements:** a Mac, the private `macos-setup` clone of the owner
     (made private, `gh auth login` first), mise.
   - **Which path:** a Mac with no `macos.toml` → first adoption; a Mac
     with `macos.toml` → update; a Mac on the old `dotfiles/` model → send
     the agent to `docs/setup/migrate-dotfiles.md` first.
   - **First adoption, steps:** (1) select packs and tools with the owner:
     the 14 packs of `macos/packs.toml`, the `default = true` members, and
     single tools from `docs/setup/tools.md`; pack `base` is always added;
     (2) ask each `[questions.<key>]` of each selected tool, one at a time,
     and offer the result of its `detect` command as the proposal; (3)
     write `macos.toml` at the root of the clone, with
     `macos/examples/macos.toml`
     as the model (`[packs.<name>]`, `[tools.<name>]` with the answers,
     `[overrides.mise.*]`, `[mise.*]` for own additions, no `env` key, no
     `[overrides.mise_env.*]`); (4) answer the question `config_path` of
     `macos-setup`
     with the path of the clone; (5) run the Mac sync with `--dry-run`, with
     the variable inline on the first run:
     `MACOS_SETUP_CONFIG_PATH=<path of the macos-setup clone> mise run
     updateall:configs --dry-run`,
     and read the diff with the owner; (6) the real run, the same command
     without `--dry-run`; later runs get the variable from the generated
     mise env (D4); on a first run, exit 4 means: open a new terminal (the login
     shell changed) and run again; (7) commit `macos.toml` in the clone.
     A new Mac with no mise and no clone uses the parked change of the
     `setup` script of `macos-setup` (clone bootstrap over HTTPS,
     `gh auth login`, then the Mac sync); the doc says so (D4).
   - **Update to a newer `bootstrap_commit`:** read `bootstrap_commit` of
     `macos.toml`, run `git log <commit>..main -- macos sync-engine docs/setup`
     in a bootstrap clone, answer each new question, change `macos.toml`
     where a tool or a pack changed, `--dry-run`, the real run, then commit
     `macos.toml` as `ops: sync bootstrap`.
   - **Rules:** the agent never deletes `dotfiles/` and removes no symlink;
     the agent edits no file that the sync writes (it has the header); each
     owner decision is a question.
   - Links to `docs/setup/tools.md`, `docs/setup/migrate-dotfiles.md`,
     `docs/design/setup-sync.md`.
2. **`docs/setup/tools.md`** (new) — the reference of the 92 Mac tools, from
   `MS/docs/tool-reference.md` updated to the new format (D3):
   - An intro: where the tools are (`macos/tools/<tool>/`), how a tool is
     selected (`[packs.<name>]`, `[tools.<name>]` of `macos.toml`), how a
     value is set (`[tools.<name>]` keys), and a link to
     `docs/design/tools.md` for the schema.
   - A packs table: the 14 packs, the `default = true` members and the
     other members, from `macos/packs.toml`. Pack `base` is always added.
   - One section for each tool, in alphabetical order, under a level-3
     heading ``### `<tool>` `` (D10), with the items of
     D3, read from its `tools.toml`. A question shows its key and the link
     `macos/tools/<tool>/tools.toml` only.
   - Remove each old item: `tools = [...]`, `<tool>.toml` files, "a tool is
     in at most one pack", the iCloud config folder, the old fragments
     under `.config/mise/conf.d/`, `generated.txt`.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- dprint check docs/setup` passes.
- ``docs/setup/tools.md`` has one section for each folder of ``macos/tools/``:
  ``grep -c '^### ``' docs/setup/tools.md``equals``command ls macos/tools | wc
  -l`
  (92).
- `grep -nE 'iCloud|macos-setup\.toml|generated\.txt|tools = \['
  docs/setup/index.md docs/setup/tools.md`
  finds nothing, except a line that names an old item as gone.
- `grep -c 'question = ' docs/setup/tools.md` is 0 (no question text).
- `grep -nF 'MACOS_SETUP_CONFIG_PATH=' docs/setup/index.md` finds the
  first-run command.
- Each command that `docs/setup/index.md` gives exists in the landed code
  (the task name in `macos/tools/macos-setup/`, the flags in
  `sync-engine/setup-sync`).

## Guardrails

- Touch only `docs/setup/index.md` and `docs/setup/tools.md`. U2 owns
  `docs/setup/migrate-dotfiles.md`; link to it by its path.
- Do not edit `MS`; read it only.
- Write all prose in ASD-STE100 Simplified Technical English. Write
  identifiers in code spans.
- Delete with `rm`, never `git rm`.

## Commit

`docs: add the Mac adoption prompt and the Mac tool reference` — written by
the orchestrator after the unit gate and the wave gate, not by the unit.
