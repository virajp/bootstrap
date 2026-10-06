# U1 — Scaffold the pnpm workspace and the cli package

- **Wave:** 1
- **Depends on:** —
- **Owns:** `pnpm-workspace.yaml`, `package.json`, `.npmrc`,
  `tsconfig.base.json`, `cli/package.json`, `cli/tsconfig.json`,
  `cli/vitest.config.ts`, `cli/tsdown.config.ts`, `cli/src/bin.ts`,
  `cli/src/cli.ts`, `cli/test/cli.test.ts`, `.config/mise/tasks/p/cli/build`,
  `.config/mise/tasks/p/cli/test`, `.config/mise/tasks/p/cli/check`,
  `pnpm-lock.yaml`
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/cli.test.ts` — running the root command with
  `--version` prints the `cli/package.json` version and succeeds; running it
  with no arguments prints help and exits 2. Fails because no package exists.
- **Read first:** every owned file that exists (`package.json`,
  `pnpm-lock.yaml`), `.claude/stackgen/templates/typescript-effect-cli.md`,
  `.claude/stackgen/templates/pnpm-workspace.md`,
  `.claude/stackgen/templates/npm-package.md`.
- **Lazy-load:** `.config/mise/tasks/p/site/build` (the shape of an existing
  per-member task), `.config/mise/tasks/_scripts/helpers`.

## Ruling

> | 1 | Effect major | Effect v4: `effect` 4.0.1 and `@effect/platform-node`
> 4.0.1; the CLI from `effect/cli` | `@effect/cli` 0.77 on Effect v3 | U1 | | 2
> | Build tool | tsdown builds `cli/dist` (earlier product decision) | plain
> `tsc` emit | U1 |

New dependencies this unit adds (and no others): `effect`,
`@effect/platform-node`, `liquidjs`, `yaml`; dev `typescript`, `vitest`,
`@effect/vitest`, `tsdown`, `@types/node`. Use Context7 for every library API
before writing code with it.

## Edits

1. **`pnpm-workspace.yaml`** — new. Members listed by name: `cli`. A
   `minimumReleaseAge` cooldown per the pnpm convention. No `allowBuilds` entry
   unless an install proves one is needed (then name the package and say why in
   `DECIDED:`).
2. **`.npmrc`** — new, exactly `ignore-scripts=true` and `fund=false`.
3. **`package.json`** (root) — make it the private workspace root:
   `"private": true`, `packageManager` pinned to the pnpm the repo's mise config
   runs, no `main`, no publishable fields; scripts only if a task needs them.
4. **`tsconfig.base.json`** — new shared base: `strict`, `ESNext`,
   `moduleResolution: bundler`, `verbatimModuleSyntax`,
   `noUncheckedIndexedAccess`.
5. **`cli/package.json`** — `name: @virajp.dev/bootstrap`, `version: 0.0.1`,
   `type: module`, exactly one `bin` entry `bootstrap` → `dist/bin.mjs`, a
   `files` allowlist of `dist` only, `exports` closed to `./package.json`
   (earlier decision), `engines.node` matching the mise-pinned Node major, the
   dependencies above.
6. **`cli/tsconfig.json`** — extends the base; `@/*` → `src/*`.
7. **`cli/vitest.config.ts`** — Vitest with the `@/` alias and v8 coverage.
8. **`cli/tsdown.config.ts`** — one entry `src/bin.ts` → `dist/bin.mjs`, ESM,
   `#!/usr/bin/env node` banner, the `@/` alias resolved.
9. **`cli/src/cli.ts`** — the root command `bootstrap` built with `effect/cli`,
   no subcommands yet; `--version` from the package version; no command → help
   on stdout and exit 2 (design-system Terminal UX: "The bare command prints
   help and exits `2`").
10. **`cli/src/bin.ts`** — the process boundary: provides the Node layers and
    runs the command with `NodeRuntime.runMain`; maps the command's exit to the
    process exit code without calling `process.exit` elsewhere.
11. **`.config/mise/tasks/p/cli/{build,test,check}`** — new task files in the
    style of `p/site/*`: `build` runs tsdown for `cli`, `test` runs Vitest for
    `cli`, `check` runs `tsc --noEmit` for `cli`. Executable, with a shebang.
12. **`pnpm-lock.yaml`** — written by `pnpm install` after the edits above.

## Verification

- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:test` exits 0 and runs `cli/test/cli.test.ts`.
- `mise x -- mise run p:cli:build` writes `cli/dist/bin.mjs`; its first line is
  `#!/usr/bin/env node`.
- `node cli/dist/bin.mjs --version` prints `0.0.1`.
- `MISE_ENV=dev mise run code:all` exits 0.

## Guardrails

- Do not touch `.config/mise/tasks/p/i/*` (parked) or any other `.config/` file.
- Do not add any dependency not named above.
- `ignore-scripts=true` stays; never run an install with scripts enabled.
- Delete with `rm`, never `git rm`.

## Commit

`feat: scaffold the cli package in a pnpm workspace`
