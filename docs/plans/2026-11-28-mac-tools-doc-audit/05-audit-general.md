# U5 — Audit the general sections and the checklist of the old Notes

- **Read first:** `docs/setup/tools.md` from the top to the heading
  `## The tools` ("How a tool works" and "The packs"), then `macos/packs.toml`.
- **Lazy-load:** each file that a claim below cites, `docs/design/mac.md`,
  `docs/design/tools.md`, `docs/design/mise.md`, `docs/design/fnox.md`. The
  old text, only to read a claim in its context:
  `git -C ../macos-setup show a4ec825^:docs/tool-reference.md`, lines 71-201
  (`../macos-setup` is the sibling clone of the main checkout of bootstrap; in
  a worktree, go to the main checkout first with
  `git rev-parse --path-format=absolute --git-common-dir`). The checklist
  below is complete without it.

## Ruling

> D1: Audit docs/setup/tools.md against the current macos/tools/. The Notes of
> the deleted macos-setup docs/tool-reference.md (lines 71-201 at commit
> a4ec825^) are a checklist, not a source: each claim that is still true must
> be in the doc, and a false claim is dropped.

> D2: All the ### <tool> sections of docs/setup/tools.md (110 on 2026-11-28),
> plus its sections 'How a tool works' and 'The packs'. Each section is
> compared with tools.toml, files/, tasks/ and scripts/ of its tool folder,
> and 'The packs' with macos/packs.toml.

> D3: Each fact goes in the ### section of the tool that owns it. A rule for
> many tools goes in 'How a tool works', as a paragraph before the item legend
> ('Each section below gives these items for one tool:').

> D4: It requires docs/plans/2026-11-27-tool-skills, which requires
> 2026-11-27-tool-selection and 2026-11-27-mac-backup-target. The audit then
> examines the final state of macos/tools/ after the three plans. Each claim
> of the survey is examined again at run time, because those plans can change
> it.

> D5: Audit in parallel, then apply once: U1 to U5 (wave 1) each write one
> findings file in docs/plans/2026-11-28-mac-tools-doc-audit/audit/ and edit
> nothing else; U6 (wave 2) applies the five files to docs/setup/tools.md.

> D7: Each finding gives the section, the kind (wrong, missing, extra,
> order), the current text of the doc (or 'none'), the evidence as
> file:line, and the full correct text that U6 writes. The evidence is in
> macos/tools/ or macos/packs.toml; for a finding of How a tool works it can
> also be in docs/design/ or cli/. A finding with no evidence is not
> written. A missing section and a section with no tool folder are findings
> too. The kind order covers a section out of sort order and an item out of
> the order of the legend inside one section. A new text of more than one
> line (a whole section, an item with a sub-list) goes in a fenced markdown
> block, indented under its finding, in place of the quoted new text.

> D8: New text follows the existing doc: the items of the legend in 'How a
> tool works' (Packs, Stage, Depends, Files, Tasks, Scripts, Contributes to,
> Defaults, Mise sections, Question), in that order, and the plain STE prose
> of the doc (ASD-STE100: short sentences, simple present, active voice). A
> 'why' rule is one or two sentences in the prose paragraph under the ###
> heading, after the first sentence. No secret value is written: a secret is
> named by its variable name only.

> D9: No unit changes a file of macos/tools/. An audit unit lists each stale
> comment that it finds in the '## Stale comments' section of its audit file.
> U6 does not apply them.

> D10: No. U1 to U5 read docs/setup/tools.md, macos/tools/,
> macos/packs.toml, the docs of docs/design/ and, for How a tool works,
> cli/; they write only their own audit file.

> D11: teams. An audit unit that finds a fact that belongs to another part can
> send it to the unit of that part. Talk is a help only: each unit file stands
> alone.

## The checklist

The survey of 2026-11-28 found these claims of the old Notes true, and not
stated in a bootstrap doc. Paths are below `macos/tools/` unless they start
with `docs/`. The line numbers are from 2026-11-28 and can move (D4).

| ID  | Claim                                                                                                                                                                                                                                                                            | Evidence on 2026-11-28                                                                                                                                                                | Target section                  |
| --- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------- |
| C1  | `[mise.bootstrap.packages]` of `tools.toml` lists the Homebrew formulae (`brew:`), casks (`brew-cask:`) and Mac App Store apps (`mas:<adam id>`); a versioned dev tool goes in `[mise.tools]`                                                                                    | `developer/tools.toml:8-9`, `claude-cli/tools.toml:11-12`; `docs/design/cli.md:33-35`                                                                                                 | How a tool works                |
| C2  | mise installs the Homebrew packages itself, into `/opt/homebrew`; `brew list` and `brew upgrade` still see them                                                                                                                                                                  | `homebrew/tasks/upgrade-brew.py:53-55`                                                                                                                                                | How a tool works, or `homebrew` |
| C3  | Casks go to `~/Applications` (`MISE_BREW_CASK_OPT_APPDIR` of the mise config, `HOMEBREW_CASK_OPTS` of homebrew); `1password` uses `appdir=/Applications`                                                                                                                         | `mise/files/.config/mise/config.toml:43-45`, `homebrew/tools.toml:15`, `1password/tools.toml:23`                                                                                      | `mise`, `homebrew`, `1password` |
| C4  | A tool declares a third-party tap in `[mise.bootstrap.brew.taps]`; the formula of a tap installs only when the tap publishes `api/formula/<name>.json`                                                                                                                           | `claude-status/tools.toml:9-13` (the api rule has no evidence in the files)                                                                                                           | `claude-status`                 |
| C5  | Each tool with a `mas:` package depends on `mas`                                                                                                                                                                                                                                 | each `mas:` tool, for example `developer/tools.toml:4`                                                                                                                                | How a tool works                |
| C6  | A tool holds one app or package, unless one depends on the other; the exceptions are `fonts`, `zsh` (zsh and zinit) and `mise`                                                                                                                                                   | `zsh/tools.toml:13-15`, `fonts/tools.toml`, `mise/tools.toml`                                                                                                                         | How a tool works                |
| C7  | The shells load the drop-in files of `~/.config/bash/`, `~/.config/zsh/` and `~/.config/fish/conf.d/`, sorted by name                                                                                                                                                            | `bash/files/.bashrc:9-14`, `zsh/files/.zshrc:11-14`                                                                                                                                   | How a tool works                |
| C8  | mise activates after `brew shellenv` (`01a` before `01b`, `05` before `06`), because mise prepends to `PATH`; a later change of `PATH` makes a tool of both (`jq`, `yq`) resolve to the Homebrew copy                                                                            | the drop-in names in `homebrew/tools.toml:30-41` and `mise/tools.toml:43-59`; `mise/files/.config/zsh/01b-mise.zsh.tmpl:9` (not the comment of `05-homebrew.fish:16`, which is stale) | `mise`                          |
| C9  | An interactive shell gets a full `mise activate`; another shell gets `--shims`                                                                                                                                                                                                   | `mise/files/.config/fish/conf.d/06-mise.fish.tmpl:17-23`, `01b-mise.zsh.tmpl:12-18`, `01b-mise.bash.tmpl:16-22`                                                                       | `mise`                          |
| C10 | `06-mise.fish` sets `MISE_FISH_AUTO_ACTIVATE=0`, so the `vendor_conf.d` snippet of the Homebrew mise formula does not activate mise again                                                                                                                                        | `06-mise.fish.tmpl:11-15`                                                                                                                                                             | `mise`                          |
| C11 | `PATH` is in the shell config (`02-path.fish`, `zsh/00-bootstrap.zsh`, `bash/00-bootstrap.bash`, `.zprofile`), not in `[env] _.path` of mise, because mise puts `_.path` before its tool paths                                                                                   | `fish/files/.config/fish/conf.d/02-path.fish:14-18`, `zsh/files/.config/zsh/00-bootstrap.zsh:25-26`, `bash/files/.config/bash/00-bootstrap.bash:26-27`, `mise/files/.zprofile:3-4`    | `fish`, `zsh`, `bash`           |
| C12 | `02-path.fish` uses `fish_add_path --global`: the universal scope stays in `fish_variables`, so a removed line keeps its path; `set --erase --universal fish_user_paths` clears an old one                                                                                       | `02-path.fish:20-22,29-30`                                                                                                                                                            | `fish`                          |
| C13 | The login shell, `/etc/shells` and Touch ID for sudo are mise declarations, because they are outside `$HOME`                                                                                                                                                                     | `fish/tools.toml:16-18,27-32`, `touchid-sudo/tools.toml:11-17`                                                                                                                        | `fish`, `touchid-sudo`          |
| C14 | The init stage and exit 4 of the first setup                                                                                                                                                                                                                                     | `docs/design/mac.md:271` and below, "The first setup and exit 4" (already stated: examine only)                                                                                       | —                               |
| C15 | fnox reads the secrets from the macOS Keychain; `env = false` exports no secret to the shell; `if_missing = "error"`                                                                                                                                                             | `fnox/files/.config/fnox/config.toml:5-13`                                                                                                                                            | `fnox`                          |
| C16 | The `brew` alias runs `fnox … exec -P brew -- brew`; the profile `brew` gives `GITHUB_API_TOKEN` and `HOMEBREW_GITHUB_API_TOKEN` (names only)                                                                                                                                    | `homebrew/tools.toml:26-27`, `fnox/files/.config/fnox/config.toml:19-27`                                                                                                              | `homebrew`, `fnox`              |
| C17 | `updateall` runs `upgrade:brew` under `fnox exec --profile brew`; another `mise bootstrap` run is not authenticated                                                                                                                                                              | `mise/tasks/updateall.py:117-126`                                                                                                                                                     | `homebrew` or `mise`            |
| C18 | The second profile of fnox is `fnox`, with `AGE_KEY` (name only)                                                                                                                                                                                                                 | `fnox/files/.config/fnox/config.toml:29-32`, `fnox/tools.toml:2`                                                                                                                      | `fnox`                          |
| C19 | `projects_dir` of `cli` (default `~/Projects`) sets `PROJECTS_DIR`, the `gitdir:` rules of git, and the trusted paths of mise                                                                                                                                                    | `cli/tools.toml:10-19`, `git/files/.config/git/identity.tmpl:5,18,28`, `06-mise.fish.tmpl:3,9`                                                                                        | `cli`, `git`                    |
| C20 | The git `config` includes `identity`, which holds `[user]` and includes `github.config` and `gitlab.config` by `gitdir:` and by `hasconfig:remote.*.url`                                                                                                                         | `git/files/.config/git/config.tmpl:5-6`, `identity.tmpl:6-29`                                                                                                                         | `git`                           |
| C21 | `push.default` is `current`, and the push aliases push only the current branch                                                                                                                                                                                                   | `config.tmpl:18-19`, `git/tools.toml:43`, `config.tmpl:75`                                                                                                                            | `git`                           |
| C22 | `.ssh/config` includes `~/.ssh/config.local` (hosts of one Mac, not synced) and `config.identity`; `StrictHostKeyChecking accept-new`                                                                                                                                            | `ssh/files/.ssh/config.tmpl:13-19`, `config.identity.tmpl:5`                                                                                                                          | `ssh`                           |
| C23 | The merge spec of `settings.json` removes `deniedMcpServers` and `env.CONTEXT7_RUNNER`                                                                                                                                                                                           | `claude/files/.claude/settings.json.merge.toml:66-75`                                                                                                                                 | `claude`                        |
| C24 | A key of `settings.json` that the merge spec does not name stays as Claude Code wrote it; on a new Mac, set it again in Claude Code                                                                                                                                              | `settings.json.merge.toml` (whole file); `docs/setup/tools.md:226-229` (partly stated)                                                                                                | `claude`                        |
| C25 | `cc`, `ccf`, `cco`, `ccs` and `cch` run `func:claude` with the default model, Fable, Opus, Sonnet and Haiku; `func:claude` takes `name`, `-n/--new` and `-m/--model`, and starts `claude` with `--name <folder>[-<name>]` (and `--resume` unless `--new`) and `--remote-control` | `claude-cli/tools.toml:14-20`, `claude-cli/tasks/func-claude.py:13-17,57-85`                                                                                                          | `claude-cli`                    |
| C26 | The steps of `updateall`, in order                                                                                                                                                                                                                                               | `mise/tasks/updateall.py:76-164`                                                                                                                                                      | `mise`                          |
| C27 | Once a week (the stamp `/tmp/mise-updateall-weekly.stamp`; `--force` runs it now), `updateall` applies the macOS defaults again and runs `upgrade:power`                                                                                                                         | `updateall.py:14,27-32,57-62,147-155`                                                                                                                                                 | `mise`                          |
| C28 | A step of an opt-in tool, for example `tailscale:daemon`, runs only when its task exists                                                                                                                                                                                         | `mise/tasks/helpers.py:182-204`, `updateall.py:139-164`                                                                                                                               | `mise`                          |
| C29 | `tailscaled` runs as root and mise has no root services, so `tailscale:daemon` starts it with `sudo brew services` when it does not run; `upgrade:tailscale`, which `upgrade:brew` calls, gives the root-owned keg back before an upgrade                                        | `tailscale/tasks/tailscale-daemon.py:13-33`, `upgrade-tailscale.py:77-90`, `homebrew/tasks/upgrade-brew.py:57-61`                                                                     | `tailscale`                     |
| C30 | mise never restarts an app, so the `post-defaults` hook of `macos-defaults` runs the `killall`s, `chflags nohidden ~/Library` and writes `NewWindowTargetPath`                                                                                                                   | `macos-defaults/tools.toml:123-133`                                                                                                                                                   | `macos-defaults`                |
| C31 | `dock_orientation` is `bottom`, `left` or `right`; the detection proposes the current position, and another value stops the render                                                                                                                                               | `macos-defaults/tools.toml:19-22,79`                                                                                                                                                  | `macos-defaults`                |
| C32 | `upgrade:brew` runs `mise bootstrap packages prune`, which removes the packages that no tool declares                                                                                                                                                                            | `homebrew/tasks/upgrade-brew.py:63-68`                                                                                                                                                | `homebrew`                      |

The survey found these claims of the old Notes false on 2026-11-28. They are
not written: the bundle tools with `[defaults]` package lists and
`conf.d/<tool>.toml.tmpl`; the `macos-setup` and `devtools` tools; the
context7 profile; "set them in `<tool>.toml`" (the values are in
`macos.toml`); `CLAUDE.md` as a seed; the 19 enforced keys of
`settings.json`; `remove_items`; "only `--model` is parsed"; `updateall` as a
task of `macos-setup` with `--no-bootstrap`.

> D15: A claim with a part that has no evidence in the files is stated with
> the proven part only, and its line names the dropped part with the reason
> (for example C2, C4, C12). A claim that a bootstrap doc already states is
> stated — already in <doc> <section>, and that doc can be another doc than
> docs/setup/tools.md (for example C14, docs/design/mac.md). A claim that
> the doc states in part (for example C24) is stated with new text for the
> rest only, and its line names the section that already holds the first
> part.

> D16: Keep an item when the files support it. Each item kind that occurs in
> more than one section, and the forms Question `<key>`: and Contributes to
> `<tool>`:, gets a line in the legend of How a tool works: U5 finds these
> kinds with a grep of the whole doc and writes the legend lines. The form
> Question `<key>`: takes the place of Question, and Contributes to
> `<tool>`: the place of Contributes to; each other new kind goes after Mise
> sections and before Question, in the order that U5 gives in the legend.
> For a kind that occurs in one section only, the part unit of that section
> decides: keep it when it names a real part of the tool (for example Manual
> step:), else move its text into the prose paragraph of the section, as a
> finding of kind wrong. An item that the files do not support is a finding
> of kind wrong or extra.

## Edits

1. **`docs/plans/2026-11-28-mac-tools-doc-audit/audit/general.md`** — create
   it, in two parts:
   - **The general sections.** Compare "How a tool works" with
     `docs/design/tools.md`, `docs/design/mac.md` and the folders of
     `macos/tools/`, and "The packs" with `macos/packs.toml`: the count of
     packs, each pack, each required and other member, in the order of the
     file. Write each difference as a finding in the form below.
     Also apply D16: list the item kinds of the `### <tool>` sections that the
     legend does not name (a grep of the lines that start with `- **` in
     `docs/setup/tools.md`). For each kind that occurs in more than one
     section, and for the two forms with a key or a tool name (the Question
     item with a key, and the Contributes to item with a tool name), write a
     `missing` finding with the legend line.
   - **The checklist.** For each claim C1 to C32, examine it again against the
     current files (D4). Give it one state:
     - `stated` — it is true; give its target section, the evidence, and the
       full text that U6 writes there. When the claim is already in the doc,
       write `stated — already in <section>` and no new text.
     - `dropped` — it is false now, or it has no evidence in the files; give
       the reason.
     - A claim with a part that has no evidence is ``stated`` with the proven
       part, and its line adds ``dropped part: <part> — <reason>`` (D15). A
       claim
       that another doc already states is `stated — already in <doc>
       <section>` (D15).

   Write the file in this form:

   ```markdown
   # Audit — the general sections and the checklist of the old Notes

   ## Findings

   ### How a tool works

   - G<n> <kind: wrong | missing | extra | order> — doc: "<current text, or
     none>" — evidence: <path:line> — new: "<the full text that U6 writes>"

   ### The packs

   - G<n> …

   ## Checklist

   - C<n> stated — section: `<tool>` or How a tool works — evidence:
     <path:line> — new: "<the full text>"
   - C<n> dropped — <the reason>

   ## Stale comments

   - <path:line> — <what the comment says> — <the fact, with its path:line>
   ```

   Write "none" under a heading with no item.

## Verification

- `mise x -- mise run code:format --all` passes.
- Each of C1 to C32 has exactly one line under "Checklist", with the state
  `stated` or `dropped`.
- `git status --porcelain` shows no change outside
  `docs/plans/2026-11-28-mac-tools-doc-audit/audit/`, and this unit changed
  only its own audit file.

## Guardrails

- Do not edit `docs/setup/tools.md`, any file of `macos/tools/`, or the audit
  file of another unit.
- A checklist claim with a target section of part A to D is still yours: you
  write its text. Send the claim id to the unit of that part by a message, so
  it does not write the same text.
- Write no secret value. A secret is named by its variable name only.
- Delete with `rm`, never `git rm`.

## Commit

`docs: audit the general Mac tool sections and the old Notes`
