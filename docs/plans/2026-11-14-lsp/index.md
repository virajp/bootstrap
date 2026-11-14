# Plan — R2b: the LSP blocks, bootstrap lsp and code:lsp:check (2026-11-14)

## Goal

After this plan lands, each language tool of bootstrap declares its LSP server
in an `[lsp]` block and pins it in the dev env; `bootstrap lsp` lists the
servers of a repo in Claude's `lspServers` shape; `describe` reports them; and
`code:lsp:check` tests that each one resolves. It is the second of the three
R2 plans of `docs/memory/decisions/2026-11-14-bootstrap-plugins-split.md`, and
one piece of backlog item B28 (identify the LSP servers).

## Facts the survey established

- bootstrap pins one LSP server today: `kotlin-lsp`, a dev-only
  `http:kotlin-lsp` pin of the kotlin tool (`repo/tools/kotlin/tools.toml`,
  about lines 23-30). The python tool pins `pipx:pyrefly` as a lint gate.
  Swift comes from the host; dart comes inside flutter.
- The mise registry (2026-11-14) holds `rust-analyzer`
  (`aqua:rust-lang/rust-analyzer`) and none of gopls, jdtls,
  typescript-language-server, pyright, pyrefly, sourcekit-lsp or kotlin-lsp.
- stackgen (claude-plugins) knows four servers, each a verbatim `lspServers`
  entry in a pack, launched with `mise x --`: `kotlin-lsp`
  (`language/kotlin/pack.yaml`), `sourcekit-lsp` (`language/swift`),
  `dart-lsp` (`app-framework/flutter`) and `typescript-lsp`
  (`language/typescript`, through `pnpm dlx`). The skill instructions write
  them into vwf-tools; no script does.
- The engine: `Manifest` in `cli/internal/source/manifest.go` (about 48-83),
  `repoLines` with the `[formatter]` parse as the model for a small typed table
  (about 210-219); no allowlist of top-level keys. Rules in
  `cli/internal/source/check.go` (about 100-161).
- R1b D10 leaves an `lsp` field of `describe` to R2 (additive, D7 of R1b); R1b
  D3 names the port `code:lsp:check`.
- No LSP check exists in either repo. claude-plugins parks it in
  `docs/memory/gaps/2026-11-13-execute-lsp-check-parked.md` and in vwf
  doctor's `checks.mjs`.
- Gates: `code:format`, `code:lint`, `test:go`, `test`, `code:sec`. Commit
  types: the ten of `.config/git-conventional-commits.yaml`, no scopes.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this plan, its
owned paths plus "touch nothing outside this list", the facts section, the
plan's shared files, and the return block below. A unit never bumps a version,
never runs a generator, never edits a doc, never adds a dependency the plan
does not list, never commits. A unit deletes with plain `rm`, never `git rm` —
it stages nothing. A unit never writes a lock file — the file a package
manager or a toolchain manager writes to pin what it resolved — with Write,
Edit, `sed` or a script: a lock file changes only as the output of the CLI
that owns it, the command the stack's package-manager conventions name, and
the unit keeps what that CLI wrote. A unit whose manifest change moves a lock
file lists that lock file in its `owns` beside the manifest.

A unit returns exactly this block and nothing else — no file contents, no diff:

    CHANGED: <path> — <one line>            (one per file)
    DECIDED: <what> — <why>                 (choices made inside scope, or none)
    DOCS FALSIFIED: <path> — <passage>      (reported, never edited; or none)
    QUESTION Q<n>: <the question, readable alone>
    OPTIONS Q<n>: <option 1, the recommendation> — <consequence> | <option 2> — <consequence> [| <option 3> | <option 4>]
    WHY Q<n>: <path:line or plan section> — <why the plan does not decide it>

The three QUESTION lines repeat per question, `n` counting from 1 in the
block, with two to four options; a block with no question carries the one
line `QUESTION: none`. A question is a point the plan does not decide and the
unit cannot settle inside its scope; the unit never proceeds on an
assumption.

A unit completes every edit that does not depend on an answer, never builds on
a part an answer can change, then returns all its open questions in one block.
Its edits stay uncommitted in the worktree. The orchestrator asks the user,
records each answer in `decisions`, and resumes the unit with the question ids
and answers; the resumed unit applies them and returns the whole block again.

## Out of scope

- R2c: the `virajp-plugins` tool and a declarative `setup:ai` (the other piece
  of B28).
- claude-plugins: stackgen and vwf doctor and execute (parked below).
- An LSP initialize test of each server (D5).
- LSP servers on a Linux dev host (D26): dev hosts are macOS only, and CI does
  not run `code:lsp:check`.
- This repo's own `.config/` files, except `.config/mise/tasks/test/`, which
  holds bootstrap's own tests.
- A version bump or a tag (R1b D16).

## Parked

- B28: install a Claude marketplace and its plugins — planned next as R2c,
  which finishes B28.
- A TypeScript server for a bun-only repo (D18): bun gets no `[lsp]` block in
  R2b; a later plan decides its server name and pin.
- stackgen must read `bootstrap lsp --json` instead of the packs'
  `lsp_servers`, and vwf doctor and execute can use `code:lsp:check`:
  `docs/memory/gaps/2026-11-14-plugins-use-bootstrap-lsp-parked.md`.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches —
whichever kind the plan is:

/vwf:execute docs/plans/2026-11-14-lsp

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
