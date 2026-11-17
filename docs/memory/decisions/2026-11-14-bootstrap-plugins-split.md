DECISION 2026-11-14 ★5 | repo:bootstrap, repo:claude-plugins,
skill:git-workflow,
skill:architecture, plugin:stackgen | bootstrap owns every deterministic concern
(tools, repo config, tasks, git actions, LSP binaries, plugin installs);
claude-plugins owns every judgment concern; they meet at one versioned JSON
contract on the bootstrap CLI | alt rejected: clone-and-read files, prompt-only
handoff, vwf owns topology, stackgen packs

# The split between bootstrap and claude-plugins

The user approved this design on 2026-11-14 in three review rounds. The
reviewed design, with its diagrams, is the artifact
<https://claude.ai/artifact/BZKUcn2vbr1msAk7GLgsuE>. This record has the same
text in the `bootstrap` repo and in the `claude-plugins` repo.

## Principle

- If a script can do it with the same result every time, bootstrap owns it. If
  it needs a model, the plugins own it.
- Four layers: machine (bootstrap `macos/`), repo shape (bootstrap `repo/`),
  stack knowledge (stackgen), judgment (vwf). A higher layer calls a lower
  layer. A lower layer never names a higher layer.

## The contract (D1, D2, CT2)

- The plugins call the installed `bootstrap` CLI. Each command has `--json`
  and a contract version. Contract v1 ships inside bootstrap 4.0.0 (amended
  2026-11-14, see below).
- Commands: `contract`, `describe` (tools, members, LSP, ports), `config`
  (`get`, `set`, `unset`, `show`), `catalog`, `add` and `remove` (tool,
  plugin), `init --values`, `check`, `lsp`.
- Task ports that exist in every shaped repo (a port with no work is a no-op
  that exits 0): `code:format`, `code:lint`, `code:sec`, `code:precommit`,
  `git:merge:<b>`, `git:worktree:{create,setup,delete}`, `code:lsp:check`,
  `setup`, `test`, `test:e2e`, `test:golden`, `p:<id>:deploy`.
- The plugins never read or write a file in `.config/bootstrap/` directly.
- `docs/sync/vwf.md` stays as the human doc of the contract.
- bootstrap is pinned to the latest signed release. The CLI reads the catalog
  of that release, never the newest `main` (D9).

## Config layout (OW1, CF1, RT1)

`.config/bootstrap/` holds files that only the bootstrap CLI writes:

| File                 | Holds                                                         |
| -------------------- | ------------------------------------------------------------- |
| `repo.toml`          | Repo-level facts that any tool or skill reads                 |
| `tools.toml`         | Tool selection and tool values (was `.config/bootstrap.toml`) |
| `releases.toml`      | Pending release levels (was `.config/releases.yaml`)          |
| `repo-generated.tsv` | The sync manifest (was `.config/repo-generated.tsv`)          |

`bootstrap config` reads and writes them. The first segment of a key names the
file, for example `repo.branches.main.landing` or `releases.vwf`. Reads take
`--format json|yaml|toml`.

`repo.toml` keys (snake case):

- Identity: `repo_name`, `product_name` (replaces vwf `product.name`),
  `repo_url`.
- Git: `core_branches = [...]` at the root. `[branches.<b>]` has one key,
  `landing = "pr" | "direct-merge" | "direct-commit"`, plus `merge_from` and
  `merge_never`. `direct-merge` allows a merge without a PR but no direct
  commit. `direct-commit` allows both.
- Commits: `[commits]` with `style` and `scopes`.
- Engine: `environments` (mise environments), `bootstrap_commit`,
  `[worktree].files`. `[overrides.mise.*]` stays in `tools.toml`.
- Topology: `topology`, `topology_reason`, `[members.<name>]` with `folder`,
  `type` (`submodule | folder | sibling`), `url`, `label`,
  `worktree_install`. A member has a `[home]` back-link, so each repo finds
  the others. `linkage` goes away.
- Deploy: `release_env` and `[deploy.<env>]` with one URL for each project.

## vwf.yaml format 27 (VY1, key decisions)

- Stays: `config_format`, `blueprint_format`, `blueprint`, `ui`,
  `integrations`, `stacks`, `addons`, `enforcement`, `pipeline`, `design`,
  `docs_sync`, `after_landing`, and in each `projects.<p>`: `repo`, `design`,
  `cicd`, `stylesheet`, `coverage_target`, `harness.health`.
- `projects.<p>.stack` becomes a slim decision record: `frameworks`,
  `dependencies`, `services` (was `backing_template`), `deploy` (was
  `deploy_template`), `os`, `note`. `languages` and `package_manager` are read
  from bootstrap.
- Moves to `repo.toml`: `product`, `topology`, `topology_reason`, `linkage`,
  `members`, `environments`, `production_env`.
- Removed: `repo.stack`, `harness`, every `*_template` key.

## Ownership (OW2)

- Git: the `git-workflow` skill is the one place for git decisions (branch,
  timing, consent, commit message, plan bookkeeping). Any skill calls it.
  bootstrap's `git:*` tasks do the actions (direct merge or PR with `gh` or
  `glab`; worktree create, setup, delete). A person can run the same tasks by
  hand (D4, D5, GW1).
- Stack: packs go. `/vwf:architecture` decides with `bootstrap catalog`.
  stackgen stays without packs: dependencies in the project manifests,
  reputation checks, conventions and skills in `.claude/`, LSP registration
  (FL2, ST1).
- Tools per project: each repo lists its projects in `[projects.<name>]` of
  `repo.toml`; each tool in `tools.toml` has a `projects` list of those names.
  `bootstrap detect` proposes both, and `bootstrap check` verifies them (ST2).
- LSP: each language or runtime tool in bootstrap pins its LSP server and has
  an `[lsp]` block. `bootstrap lsp` lists the servers; `code:lsp:check`
  checks them. stackgen writes the Claude LSP entries in `vwf-tools` and
  installs nothing (D6, LS1).
- Plugins: `setup:ai` installs a listed plugin at user scope only when it is
  not there. When a plugin leaves the list, it is removed from the project
  scope only. A bootstrap tool named `virajp-plugins` carries the vwf
  `.gitignore` lines and plugin entries. The engine and the shared files name
  no vwf file (D7, D8).
- Drift: `bootstrap check` reports repo drift. `/vwf:doctor` calls it and adds
  the judgment checks (D10).

## Phases

| Phase | Repo           | Content                                                                                     |
| ----- | -------------- | ------------------------------------------------------------------------------------------- |
| P0    | claude-plugins | Fix the drift against today's bootstrap key names, paths and verbs                          |
| R1    | bootstrap      | Config folder, `config` verb, contract v1, task ports (4.0.0), as three plans R1a, R1b, R1c |
| R2    | bootstrap      | `git:*` tasks, LSP blocks and check, `virajp-plugins` tool, `setup:ai`                      |
| —     | bootstrap      | Release 4.0.0, with the user's consent, after R1a, R1b, R1c and R2                          |
| R3    | claude-plugins | Plugins use the contract; `git-workflow` over `git:*`                                       |
| R4    | claude-plugins | `vwf.yaml` format 27 and its migration                                                      |
| R5    | claude-plugins | Stack without packs                                                                         |
| R6    | both           | One release in each repo; each plan carries its own docs                                    |

A vwf plan runs only in its own repo, and `requires` cannot name a plan in the
other repo. R3 starts only after the bootstrap 4.0.0 release. The user
confirms that release by hand.

## Amendments 2026-11-14 (R1a planning)

The planning of R1a (`docs/plans/2026-11-14-config-folder` of bootstrap)
changed these points. They win over the text above.

- **The release (D1, reverses D2).** bootstrap 3.0.0 (the selection model) and
  3.1.0 (the TUI) ship as planned. R1 moves the values files, a breaking
  change, so it ships as the MAJOR release `bootstrap@4.0.0`, not inside
  3.0.0. R3 in claude-plugins waits for 4.0.0, not 3.0.0.
- **R1 is three plans (D2).** R1a: the `.config/bootstrap/` folder, the
  `repo.toml` schema, the `init` migration and the `config` verb. R1b:
  `contract`, `describe`, `catalog`, `check --json`, and `--json` on `add`,
  `remove` and `init`; it requires R1a. R1c: the task ports as no-op
  defaults; it requires R1a.
- **One tag (D92).** `bootstrap@4.0.0` is tagged once, after R1a, R1b, R1c and
  R2 all land. Until then the user tests with a local binary. R1b, R1c and R2
  are not 4.x minor releases. R3 waits for 4.0.0.
- **`settings` (D3).** The old `bootstrap config`, which printed the settings
  of the engine, is now `bootstrap settings`. `bootstrap config
  get|set|unset|show`
  is the verb of the values.
- **The overrides (D5).** `.config/bootstrap.overrides.toml` merges into
  `tools.toml`, under `[overrides.tools.*]`. A key of `repo.toml` has no
  override.
- **`releases.toml` (D10).** R1a defines the schema of `releases.toml` and the
  `config` keys `releases.<project>`, but `init` does not move
  `.config/releases.yaml`. R3 switches vwf and the release tasks to
  `bootstrap config`, and its migration moves the file.

## Amendments 2026-11-14 (R1b planning)

The planning of R1b (`docs/plans/2026-11-14-contract-commands` of bootstrap)
changed these points. They win over the text above.

- **One tag (D16, supersedes R1a D69).** R1a, R1b, R1c and R2 land before one
  `bootstrap@4.0.0` tag. Until then the user tests with a local binary. R1b,
  R1c and R2 are not 4.1, 4.2 or 4.3 releases. It restores R1a D36, and it
  agrees with D92 above.
- **The envelope (D1).** Each `--json` command prints one object on stdout:
  `{"contract": 1, "command", "ok", "code", "data", "errors", "warnings"}`.
  Each human line goes to stderr. The exit code is `code`. `questions --json`
  also prints the envelope, with its list in `data` (D2).
- **The version (D7).** The contract version is one integer, `1` in R1b. A new
  field, command or enum value keeps it. A removed or renamed field, or a
  changed type or meaning, raises it. `bootstrap contract` lists the versions
  that the CLI supports.
- **No LSP field yet (D10).** `describe` reports `repo`, `tools`, `members` and
  `ports`. R2 adds the `lsp` field; by D7 that keeps contract `1`.
- **One fetch rule (D6, reverses R1a D23 and D52).** `config set` and
  `config unset` of a `tools.*` key, `catalog` and `describe` use the cached
  clone when its last fetch is 30 minutes old or less, else fetch. When the
  fetch fails, they use a cached clone less than 7 days old, with a warning,
  else they stop with code 5 (D5).

## Amendments 2026-11-14 (R1c planning)

The planning of R1c (`docs/plans/2026-11-14-task-ports` of bootstrap) changed
these points. They win over the text above.

- **The ports of R1c (D1).** R1c makes `test`, `test:e2e` and `test:golden`
  exist, as group dispatchers of the `mise` tool. R2 makes
  `git:merge:<branch>`, `git:worktree:create|setup|delete` and
  `code:lsp:check` with real work. `code:format`, `code:lint`, `code:sec`,
  `code:precommit` and `setup` already exist in every shaped repo through the
  required base tools. No release comes between R1c and R2.
- **No default for a patterned port (D5).** R1c ships no `p:<id>:deploy`. A
  patterned port has no default: it exists for an id only when a task for that
  id exists, which `describe` reports. The rule "every port exists in every
  shaped repo" holds for the plain ports only. Per-project ids in bootstrap go
  with R5.

## Amendments 2026-11-14 (R2 planning)

The planning of R2a (`docs/plans/2026-11-14-git-tasks` of bootstrap) changed
these points. They win over the text above.

- **R2 is three plans (D9).** R2a: the `git:*` tasks. R2b: the LSP blocks,
  `bootstrap lsp` and `code:lsp:check`. R2c: the `virajp-plugins` tool and a
  declarative `setup:ai`. Each requires the one before it. Backlog item B28 is
  a piece of R2b, and R2c finishes it.
- **No aliases (D1).** R2a renames `code:merge`, `code:merge:<b>`,
  `setup:worktree` and `code:worktrees` outright. The sync removes the old task
  files through the manifest. One `bootstrap@4.0.0` tag covers R1a to R2, so no
  release ships between. This reverses the text that kept the old names as
  aliases for one release.
- **One merge task per core branch (D2).** The `git` tool ships
  `git:merge:main` and `git:merge:develop`, each only when its branch is in
  `repo.core_branches`. `<branch>` is always the destination. The source is
  optional: for `main`, `develop` when it is a core branch, else the current
  branch; for `develop`, the current branch. The merge library and its landing
  do not change, and mise's `code:merge` goes away.
- **Raw git for the worktrees (D3).** `git:worktree:create <branch>
  [--from <base>]` runs `git worktree add` in `.worktrees/<branch>` of the main
  checkout, then `mise trust` and `git:worktree:setup` there.
  `git:worktree:delete <branch>` runs `git worktree remove`, and deletes the
  branch only after it proves that a core branch has it. Not through
  worktrunk.
- **The lister (D4).** `code:worktrees` becomes `git:worktree:list`, and the
  shell alias `worktrees` runs it. Its content does not change.

## Amendments 2026-11-14 (R2b planning)

The planning of R2b (`docs/plans/2026-11-14-lsp` of bootstrap) changed these
points. They win over the text above.

- **Every language tool (D1).** kotlin, swift, flutter, node, go, python, rust
  and java each have an `[lsp.<server>]` block: `kotlin-lsp`,
  `sourcekit-lsp`, `dart-lsp`, `typescript-lsp`, `gopls`, `pyrefly`,
  `rust-analyzer` and `jdtls`. A new pin is dev-only. swift's server comes
  from the host, and dart's comes with flutter, so neither has its own pin.
  bun and android get no block. node alone gets no server: its pins and
  `typescript-lsp` need a `package.json` or a `tsconfig.json` (amended
  2026-11-16, D35 to D40).
- **Claude-ready output (D4).** `bootstrap lsp --json` prints
  `data.servers`, one verbatim entry of Claude's `lspServers` for each server
  (`command` `mise`, `args` `["x", "--", <command>, <args>…]`,
  `extensionToLanguage`, `startupTimeout`), so stackgen copies each entry
  without a change. `data.checks` gives the real command and the source of
  each server. `describe --json` has the same object in its `lsp` field.
- **What `code:lsp:check` tests (D5).** It tests that each listed server's
  command resolves in the dev env: `MISE_ENV=dev mise which <command>` for a
  pinned server, `xcrun --find <command>` for a host server. It starts no
  server. A missing server prints one line with its fix and exits 1.

## Amendments 2026-11-14 (R2c planning)

The planning of R2c (`docs/plans/2026-11-14-plugins-setup-ai` of bootstrap)
changed these points. They win over the text above.

- **Contributed merge items (D1, D7, D11).** A tool adds items to a
  merge-mode file of another tool with
  `[[contributes_items.<owner>."<target>".item]]`, in the item format of the
  owner's spec; the owner stays the one owner of the file. The items go after
  the owner's own items, in the order of the contributing tools' names, with
  no duplicates, and the owner's override table applies to them. `check`
  refuses (code 2) an owner that the catalog does not have and a target that
  the owner does not declare as merge-mode. `virajp-plugins` carries the vwf
  `.gitignore` lines this way, and its plugin entry in
  `[contributes.claude.plugins.user]`.
- **The record (D3, D5).** `setup:ai` keeps the project plugins that it
  installed in `tools.claude.plugins.installed`. A recorded plugin that leaves
  `[tools.claude.plugins.project]` is uninstalled at project scope and dropped
  from the record. A plugin installed by hand, and user scope, are never
  touched.
- **Update only on a flag (D4, amended by D39).** A plain `setup:ai` installs
  what is missing and does the D3 removal, and nothing else; a second run
  changes nothing. `mise run setup:ai --upgrade` also updates the marketplaces
  and the listed plugins. The flag is `--upgrade`, the same as
  `setup --upgrade`, which runs `setup:ai --upgrade`; a plain `setup` does not
  run `setup:ai`.

## Amendment 2026-11-15 (R5 brainstorm)

The user stopped the first R5 interview: "packs go" was read as "bundles go
too". A brainstorm on 2026-11-15 replaced the R5 line of the Ownership section
("Stack: packs go") and the "every `*_template` key" removal of the format 27
section with these rulings. Background:
`claude-plugins:
docs/memory/gaps/2026-11-15-r5-stack-without-packs-resolved.md`.

- **Bundles stay (B1).** A bundle is the curated unit a consumer picks. It is
  never mandatory. It lists bootstrap `tools` and knowledge `packs`.
- **Repo bundles (B2).** A user who picks no shipped bundle composes one with
  architecture: bootstrap tools, shipped packs and generated packs. stackgen
  saves it as `.claude/stackgen/bundles/<slug>.md` in the repo.
- **Tool knowledge (B3).** The 18 `language`, `package-manager` and
  `toolchain-gate` packs stop being packs. Their conventions, skills and hook
  stay in stackgen, keyed by the bootstrap tool name, and apply when a project
  links that tool. bootstrap ships no model prose. `ux-gate` and the
  `docs/blueprint/` references move to vwf.
- **Knowledge packs (B4).** The other packs stay, without the data bootstrap
  owns (`lsp_servers`, `mise_tool`, `binaries`, `package_manager`,
  `lockfile`). Their task files ship as `NN-<name>` parts.
- **Tools at setup (B5).** `/vwf:setup`'s materialize pass runs
  `bootstrap add tool` for each missing bundle tool and `bootstrap detect
  --add`, on consent.
- **LSP (B6).** stackgen copies `bootstrap lsp --json` into `vwf-tools`.
- **Design adapters (B7).** The `design-tool` kind goes. stackgen keeps the
  adapters outside the packs and installs them into `vwf-tools`: one skill per
  role at the fixed names, plus `references/<tool>.md` for each tool a repo
  picks.
- **Addons (B8).** The `addon` kind goes. An addon is a Claude plugin that
  `bootstrap add plugin` installs; `addons:` stays in `vwf.yaml`.
- **Templates stay (B9).** `stack.template` and `repo.stack.template` name a
  bundle, shipped or repo. No `vwf.yaml` format 29.
- **Existing repos (B10).** `stackgen-sync` converts tool-pack rows of
  `lock.yaml` to tool-knowledge rows and moves pack task files to `NN-` parts.
- **The generator stays (B11)** for each stack item no pack covers.
- **Doctor (B12).** The language check examines that each tool of the bundle
  is linked to the project.
- **Two plans (B13).** R5a: stackgen and the repo tooling. R5b: vwf.

The three test tasks of the stackgen packs (android `test/e2e`, compose
`test/golden`, swiftui `test/golden`) move to the `android` and `swift` tools
in `docs/plans/2026-11-15-test-parts` (R2e; claude-plugins R5a D8).
