# U1 — The engine: projects, tool links, detect, check, describe

- **Read first:** `cli/internal/core/detect.go` whole,
  `cli/internal/scripts/scripts.go` (`Job`, and `cmd.Dir = r.Root` at about
  line 159), `cli/internal/repo/rules.go` (the `git ls-files ...
  --exclude-standard` read at about line 405), and the code that R1a and R1b
  added: the `internal/schema` table, the `config` verb, `describe`,
  `detect --json` and `check`.
- **Lazy-load:** decisions D15 and D86 of
  `docs/plans/2026-11-14-config-folder/index.json`, and D7, D10 and D30 of
  `docs/plans/2026-11-14-contract-commands/index.json` (or their archived
  paths).

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`.

- **D3, D7:** `repo.toml` gets `[projects.<name>]` with one key, `folder`
  (repo-relative; `.` is the root). A name is lowercase kebab case.
  `tools.toml` gets `[tools.<t>].projects = [<name>...]`, always valid, as
  `env` is (R1a D86). A name that is not a `[projects.<name>]` of the same
  repo, or `projects` on a `root_only` tool, is code 2.
- **D4, D6, D8:** detect finds the projects:
  1. The workspace manifests: `pnpm-workspace.yaml` (`packages`),
     `aube-workspace.yaml` (`packages`, D24), `go.work`
     (`use`), `Cargo.toml` `[workspace] members`, `pyproject.toml`
     `[tool.uv.workspace] members`, globs expanded.
  2. The folder walk: a folder where the detect script of a tool with
     `category = "languages"` gives `match`. Skip `.git`, `node_modules`,
     `vendor` and the gitignored paths. A folder inside a found project is not
     a new project.
  3. With no project found, the root is one project named `repo_name`.
  4. A name is the folder basename; when two are the same, the path with `/`
     replaced by `-`.
- **D28, D29:** in step 1, drop a member equal to `.`; apply the pnpm and
  aube `!` patterns and the Cargo and uv `exclude` lists; keep only folders
  that exist; apply the nesting and gitignore rules of step 2 to these
  folders too.
- **D27:** a match in the walk or in a project folder counts only when one of
  its evidence paths sits in that folder itself (the path has no `/`).
- **D26:** detect never links a `root_only` tool, and check skips it.
- **D15:** the walk skips the root when step 1 or the walk below the root
  found a project. The root is a project only by step 3.
- **D16:** a name is normalized to lowercase kebab case (`My_App` →
  `my-app`); an empty result skips the folder with a warning.
- **D17:** a folder that `repo.toml` lists is never proposed again (matched
  by folder). A new folder whose name is already used takes the path form of
  step 4.
- **D18:** the walk stops 4 levels below the root, as `hasFile` does
  (`cli/internal/repo/rules.go:401`).
- **D4:** detect runs each detect script in the root and in each project
  folder (a new `Job` field for the folder; the default stays the root). A
  `match` in a project folder links the tool to that project.
- **D19:** the per-folder runs cover the listed projects and the projects
  proposed in the same run; `--add` writes both together.
- **D14:** a tool that matches at the root and in no project folder links to
  the projects that the workspace manifest in its evidence names (pnpm: the
  `packages` of `pnpm-workspace.yaml` or `aube-workspace.yaml`). With no
  workspace manifest in its evidence, it links to the root project; with no
  root project, it gets no link (D23). This root fallback applies only to
  tools of the categories `languages` and `package-managers` (D25). A tool
  that matches in a project folder links by D4, whatever its category.
- **D9:** `detect --json` data adds `projects` (the new tables) and `links`
  (tool → the project names to add). `detect --add` writes both, on consent.
  A tool that matches only in a project folder is offered too. `init` and the
  TUI write no projects.
- **D10:** `check`: a listed project whose folder does not exist is a finding
  with code 2; a link whose detect script no longer gives `match` in that
  folder is a warning. A project with a missing folder gets only the code 2
  finding, no link warning (D21). A D14 link is tested the way detect made
  it: the tool still matches at the root, and the workspace manifest in its
  evidence still names the folder; the script is not run in the project
  folder (D22). As `tools.toml` stores only names, a link passes when either
  test holds: the script matches in the folder (with D27), or the D22 test
  (only for tools of the categories `languages` and `package-managers`); check
  warns only when both fail (D31).
- **D30:** a project with the folder `.` gets D4 links only from tools of the
  categories `languages` and `package-managers`.
- **D11:** `describe`: each tool in `data.tools` gets `category` from the
  catalog. The contract stays 1.
- **D5, D12:** `config set repo.projects.<name> '<JSON table>' --json` works;
  `config unset repo.projects.<name>` is code 2 while a tool lists the name,
  and the error names those tools.

## Edits

1. The schema table, the tools.toml key rule, and the config unset guard.
2. Detect: the project pass, the per-folder runs, the `projects` and `links`
   data, and the `--add` writes.
3. Check and describe.
4. The Go tests: one case for each ruling above.

## Verification

- `mise x -- mise run test:go` passes.
- `mise x -- mise run code:format` passes.

## Guardrails

- Touch nothing outside `cli/`.
- Add no Go module dependency; parse TOML and YAML with the libraries that
  `cli/go.mod` already has.
- Delete with `rm`, never `git rm`.

## Commit

`feat: projects in repo.toml, a projects list per tool, found by detect`
