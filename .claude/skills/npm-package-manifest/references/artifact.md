# The artifact — one tarball, shaped by the manifest

## One package, one tarball

The published package is the `cli` member. A workspace root must never be
what gets published. The guard for that is npm's refusal to publish a
`"private": true` manifest (`EPRIVATE`), which the workspace bundle may
apply to the root manifest it owns. The tarball is built from `cli/` by
`pnpm pack`. That is the step that turns `workspace:` and `catalog:`
specifiers into real ranges and applies `publishConfig` overrides. Anything
in the manifest that only works inside the workspace must be resolved by
then.

## `bin`: exactly one entry, a built file, a shebang

- **One entry.** `npx @virajp.dev/bootstrap` works out which executable to
  run from the package. With a single `bin` entry it runs that entry. With
  several, it runs the one whose name matches the package's unscoped name
  (`bootstrap`). Otherwise it fails. Name the single entry `bootstrap`, so
  that adding a second bin later can never break `npx`.
- **Point it at built output**, never at a TypeScript source. Consumers
  have no TypeScript toolchain.
- **The file starts with `#!/usr/bin/env node`.** npm links the bin file
  onto `PATH` (and writes a `.cmd` shim on Windows). Without the shebang the
  script starts without Node.
- The file in `bin` is always included in the tarball, even if `files`
  forgets it. Do not rely on that: list the build directory in `files` so
  the bin's own imports ship too.

## `files`: an allowlist, the build output only

List the build output directory and nothing else. npm documents its
selection rules:

- `package.json`, `README`, `LICENSE` and the `bin` files are always
  included.
- `.git`, `.npmrc`, `node_modules` and lockfiles are always dropped.
- A root `.npmignore` does not override `files`.

**pnpm, which packs this tarball, documents its rules only partly.** Its
pack docs cover the `--dry-run` listing and the `.env` warning. Its stated
order of `files`, then `.npmignore`, then `.gitignore` is documented for
`pnpm deploy`, not for `pack`. Do not assume npm's always-include and
always-exclude lists hold for pnpm. The listing check in the hygiene
reference is what confirms the actual contents.

Do not ship sources, tests, fixtures, tsconfig or the repo config files the
CLI generates for other repos (those are templates the code embeds or reads
from the build output, not loose files). Each one costs install time on
every `npx` cold start and adds surface for no consumer benefit.

Source maps are a decision, not a default. They help users file readable
stack traces and they show your sources. Ship them only if they map to the
public repository anyway.

## `exports`: closed, only `./package.json`

The user decided on 2026-10-06 that the import surface is closed:

```json
{
  "exports": {
    "./package.json": "./package.json"
  }
}
```

Once `exports` is defined, Node encapsulates every subpath it does not list.
A deep import such as `@virajp.dev/bootstrap/dist/x.js` throws
`ERR_PACKAGE_PATH_NOT_EXPORTED`, and there is no `.` entry, so the package
root cannot be imported either. That leaves the `bootstrap` bin as the only
contract. The encapsulation is not absolute: an import by absolute file
path still loads. But it is the resolution contract that tools and semver
reasoning rely on.

**Why closed now:**

- A published import surface is contract-shaped.
- Adding `exports` to a package that people already deep-import is a
  breaking change. Node's own guidance is to list every previously
  supported subpath when introducing it.
- Opening a subpath later is additive.

Starting closed keeps every later change additive. (Principles: YAGNI's
when-not-to-apply on contract-shaped decisions; information hiding,
through a narrow published interface.)

## `publishConfig`

- `access: "public"`. A scoped package is private by default, and the
  first publish must say otherwise. Keeping it in the manifest means it
  never depends on someone remembering a flag.
- pnpm can also override `bin`, `main`, `exports`, `engines` and similar
  fields at pack time. Use that only where the workspace and the published
  manifest really differ. Otherwise keep one spelling.

## `engines`

State the Node range the CLI supports. It is advisory: npm warns and
installs anyway unless the user set `engine-strict`. So it documents the
contract but does not enforce it.

## Dependencies: what the consumer's `npx` installs

Whatever the published manifest declares as runtime dependencies is
installed on every first `npx` run. That list is the cold-start cost and
the consumer-side supply-chain surface. Which dependencies go where, and
whether the build bundles them, belongs to the language and workspace
bundles. This target sets one test: the published version must install
cleanly from the public registry into an empty directory and run. The
install smoke test checks exactly that.

## Version

`version` in `cli/package.json` is the release record. It equals the
release tag's semver and is changed only as part of cutting a release. A
version is never reused.
