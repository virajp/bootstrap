# U2 — The dotfiles migration guide

- **Read first:** `docs/design/setup-sync.md`, `macos/examples/macos.toml`,
  `docs/sync/migrate.md` (the form of a migration doc),
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/dotfiles/.config/mise/config.toml`,
  and the file list of
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/dotfiles/`.
- **Lazy-load:** the usage header of `sync-engine/setup-sync` (the flags),
  the entry task of `macos/tools/macos-setup/`, the mise docs of the
  `[dotfiles]` section and the `dotfiles.root` setting through Context7
  (`resolve-library-id` for mise, then `query-docs`).

## Ruling

From the design rulings of 2026-11-04 (verbatim):

"Friends' forks (user, 2026-11-04): [...] `migrate-dotfiles.md` moves
anyone's Mac (the user's too) from the symlink model to the repo model:
find the fork's changes (`git merge-base` with `virajp/macos-setup`, diff
`dotfiles/`), classify each (tools/packs tables, `[tools.<name>]` values,
`[overrides.mise.*]`, `[mise.*]` in `macos.toml`, files no tool ships —
open), write the new user-config file/folder, make the repo private,
verify with the scratch-`HOME` comparison, `--dry-run`, first real run.
The migration never deletes `dotfiles/`; each owner removes it.
The migration removes NO symlink (user, 2026-11-04): links into
`dotfiles/` (e.g. `~/.config/mise`, `~/.zshrc`, `~/.ssh/config`) stay, so
the engine writes through them and the generated files land in the
private `macos-setup` repo as a backup. Only linked paths are backed up;
a path that is not a link (e.g. `~/.config/fish`) is a plain file.
86 paths exist in both `dotfiles/` and `macos-setup/tools/`; the engine
overwrites them through the links. `migrate-dotfiles.md` is the one-time
place for deletions: it deletes the unwanted files in the owner's repo
(git keeps them for recovery), for example the 19 old topic fragments
`dotfiles/.config/mise/conf.d/*.toml` (else mise loads old and new
config together, and old `[dotfiles]` entries keep making links). The
comparison step runs before the deletions.
A `dotfiles/` file that no tool ships (13 in the user's repo today) is
classified with the owner: obsolete → delete; personal → stays in
`dotfiles/` (user: no new `files/` folder) and gets a `[mise.dotfiles]`
entry in `macos.toml` (generated into `zz-macos`, linked by
`mise bootstrap`); generic → a proposal for a new tool in
`bootstrap/macos/tools/`, the bootstrap owner decides. [...]"

"Import (user, 2026-11-04): [...] Comparison: (1) render into a scratch
`HOME` with `--no-bootstrap` and a test `macos.toml` with the real
selection and answers; (2) plain files outside `.config/mise/conf.d/`:
byte compare after header removal; (3) mise config: merge all `conf.d`
files of each model, compare section by section (`[tools]`,
`[bootstrap.*]`, `[shell_alias]`, `[env]`, `[dotfiles]`, `[settings]`); (4)
each difference is a bug or an intended change, listed for the user's
approval before the first real run; (5) identity files with the real
answers (live renders come from `[vars]` of `identity.toml`). The real-Mac
`--dry-run` stays the last check."

"Values row, item 1: the Mac values live in the `macos-setup` repo, made
private, at its root (not iCloud). [...] A new Mac runs
`mise x gh@latest -- gh auth login` in `setup` before the clone. [...]"

From this plan:

D6: "A personal file stays in dotfiles/ and gets a [mise.dotfiles] entry in
macos.toml. The live model sets the mise setting dotfiles.root to the
dotfiles/ folder of the macos-setup clone; the guide tells the owner to keep
that setting in [mise.settings] of macos.toml when a personal file uses a
source path relative to it. The unit checks the [dotfiles] section and
dotfiles.root in the mise docs with Context7 before it writes the note."

D4: "On a Mac that has mise and the macos-setup clone, the first run passes
the variable inline: MACOS_SETUP_CONFIG_PATH=<path of the macos-setup clone>
mise run updateall:configs. Later runs get it from the generated mise env. A
new Mac uses the parked change of the macos-setup setup script; the doc says
so and names the parked step. Each command is one that the landed code
supports."

D5: "An AI agent that works with the owner of the Mac, as docs/sync/index.md
does for a repository. Each owner decision is one question at a time. Each
step names the command and its pass condition."

D1: "No. The plan writes docs only; no runnable line changes."

D7: "docs/setup/index.md, docs/setup/migrate-dotfiles.md and
docs/setup/tools.md. The later "Mac docs" ruling wins over the older
placement bullet, which names sync.md and tool-reference.md."

D8: "The friend skips the git merge-base step: the diff of the scratch-HOME
comparison is then the list of changes. With access, the agent uses
merge-base first."

D9: "mise x gh@latest -- gh repo edit <owner>/<repo> --visibility private
--accept-visibility-change-consequences. It names the repo and runs through
mise."

D2: "No. The 16 plans form one linear chain through requires, because most of
them edit sync-engine/repo-sync or sync-engine/setup-sync, and parallel plans
conflict there. This plan requires 2026-11-04-docs-design-split."

## Edits

1. **`docs/setup/migrate-dotfiles.md`** (new) — for an AI agent that works
   with the owner (D5). It moves a Mac from the symlink model of
   `dotfiles/` to `macos.toml`. The steps, in this order:
   1. **Find the fork's changes.** In the owner's `macos-setup` clone, add
      `virajp/macos-setup` as a remote when it is missing, run
      `git merge-base HEAD <remote>/main`, and diff `dotfiles/` from that
      commit to `HEAD`. The user's own repo has no fork changes; it starts
      at step 2 with all of `dotfiles/`. A friend without read access to
      `virajp/macos-setup` skips this step: the diff of the scratch-`HOME`
      comparison (step 6) is then the list of changes (D8).
   2. **Classify each change** with the owner, one at a time: a pack or a
      tool to select (`[packs.<name>]`, `[tools.<name>]`); a value of a tool
      (`[tools.<name>]` keys); a change to generated mise config
      (`[overrides.mise.*]`); an own mise addition (`[mise.*]` of
      `macos.toml`, generated into `zz-macos`); a file that no tool ships.
   3. **Files that no tool ships** (13 in the user's repo): ask the owner for
      each one: obsolete → delete in step 9; personal → it stays in
      `dotfiles/` and gets a `[mise.dotfiles]` entry in `macos.toml`, with
      the `dotfiles.root` note of D6; generic → a proposal for a new tool in
      `bootstrap/macos/tools/`, which the bootstrap owner decides (no new
      `files/` folder).
   4. **Write `macos.toml`** at the root of the clone, with
      `macos/examples/macos.toml` as the model, and the answers of
      `identity.toml` `[vars]` as `[tools.<name>]` values.
   5. **Make the repo private** on GitHub. Run
      `mise x gh@latest -- gh auth login` first, then
      `mise x gh@latest -- gh repo edit <owner>/<repo> --visibility private
      --accept-visibility-change-consequences`
      (D9).
   6. **The scratch-`HOME` comparison**, the 5 steps of the ruling: render
      into a scratch `HOME` with `--no-bootstrap` and the new `macos.toml`;
      byte compare the plain files after the header removal; merge and
      compare the mise config section by section (`[tools]`,
      `[bootstrap.*]`, `[shell_alias]`, `[env]`, `[dotfiles]`,
      `[settings]`); list each difference for the owner, as a bug or an
      intended change; compare the identity files with the real answers.
      The render command passes the variable inline, as on a first run:
      `MACOS_SETUP_CONFIG_PATH=<path of the macos-setup clone> mise run
      updateall:configs --no-bootstrap`
      with `HOME` set to the scratch folder (D4).
   7. **Remove no symlink.** Explain why: the links into `dotfiles/`
      (`~/.config/mise`, `~/.zshrc`, `~/.ssh/config`) stay, the engine
      writes through them, and the private repo keeps the generated files
      as a backup. A path that is not a link is a plain file and has no
      backup.
   8. **`--dry-run`** of the Mac sync on the real `HOME`, with
      `MACOS_SETUP_CONFIG_PATH=<path of the macos-setup clone>` inline (D4),
      read with the
      owner.
   9. **Delete the unwanted files** in the owner's repo, after the
      comparison, as one commit: the 19 old topic fragments
      `dotfiles/.config/mise/conf.d/*.toml` (else mise loads the old and
      the new config, and the old `[dotfiles]` entries keep making links),
      and the obsolete files of step 3. Git keeps them for recovery.
   10. **The first real run**, then commit `macos.toml` and the files that
       the sync wrote through the links.
   11. **What the owner does later:** the migration deletes nothing else,
       and never deletes `dotfiles/`; the owner decides.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- dprint check docs/setup` passes.
- `grep -nF -e <pattern> docs/setup/migrate-dotfiles.md` finds a line for
  each of these patterns: `git merge-base`, `gh auth login`,
  `--accept-visibility-change-consequences`, `MACOS_SETUP_CONFIG_PATH=`,
  `--dry-run`,
  `--no-bootstrap`, `dotfiles.root`, `mise.dotfiles`.
- `grep -niE 'rm .*dotfiles/?$|unlink|rm -rf .*\.config/mise$'
  docs/setup/migrate-dotfiles.md` finds no command that removes
  `dotfiles/` or a symlink.
- The comparison step comes before the deletion step in the file.

## Guardrails

- Touch only `docs/setup/migrate-dotfiles.md`. U1 owns the other files of
  `docs/setup/`; link to them by their path.
- Do not edit `MS`; read it only.
- Write all prose in ASD-STE100 Simplified Technical English. Write
  identifiers in code spans.
- Delete with `rm`, never `git rm`.

## Commit

`docs: add the dotfiles migration guide for a Mac` — written by the
orchestrator after the unit gate and the wave gate, not by the unit.
