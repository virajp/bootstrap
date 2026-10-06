# U2 — Repository context: root, origin, interactivity

- **Wave:** 2
- **Depends on:** U1
- **Owns:** `cli/src/repo/**`, `cli/test/repo/**`
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/repo/repo.test.ts` — (a) from a subdirectory the
  root resolves to the repository root; outside a repository the result is the
  typed "not a git repository" error; (b) `origin` URLs `git@<host>:o/r.git`,
  `https://<host>/o/r.git`, `https://<host>/a/b/c.git`, `ssh://…` forms the flow
  accepts and URLs with and without `.git` give `o/r` or `a/b/c`; no `origin`
  and unreadable URLs give no default and no error; (c) interactive is true only
  when stdin and stdout are terminals and `--json` is not given. Fails because
  the module does not exist.
- **Read first:** flow 110 step 1 and step 3 (the `--repo` default rule),
  `docs/blueprint/conventions.md` `#errors` (root from any subdirectory,
  interactive definition).
- **Lazy-load:** the Effect process-spawning and terminal service docs via
  Context7.

## Ruling

| # | Decision   | Ruling                                                                                                                   | Rejected              | Unit |
| - | ---------- | ------------------------------------------------------------------------------------------------------------------------ | --------------------- | ---- |
| 1 | Git access | "Run the git binary": `git rev-parse --show-toplevel` and `git remote get-url origin` through the Effect process service | parse `.git` directly | U2   |

## Edits

1. **`cli/src/repo/root.ts`** — the repository root via
   `git rev-parse
   --show-toplevel`; a failure is the typed error with the
   exact text "not a git repository — run `git init` first".
2. **`cli/src/repo/origin.ts`** — read `origin` via `git remote get-url origin`;
   parse ssh `<user>@<host>:<path>(.git)` and https
   `https://<host>/<path>(.git)` on any host into the repo path (one or more
   owner segments, then the name). Anything else gives "no default", never an
   error.
3. **`cli/src/repo/interactive.ts`** — interactive = stdin and stdout are TTYs
   and `--json` is not given.
4. **`cli/test/repo/**`** — the test-first cases, run against a stub process
   layer and against real temp repositories where needed.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.

## Guardrails

- No `node:child_process` import; use the platform process service.
- Do not touch `cli/src/apply/**`, `cli/src/report/**`, `cli/src/commands/**`.

## Commit

`feat: resolve the repository root and origin`
