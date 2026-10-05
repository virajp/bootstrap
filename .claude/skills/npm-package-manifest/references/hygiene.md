# Hygiene — what must never be inside the tarball

This is a correctness concern, not housekeeping. A published file is
public, it is copied by mirrors and caches, and the version that carried it
can never be reused. There is no redeploy that removes it.

## The two failures

**A credential reaching the registry.** npm's always-excluded list covers
`.npmrc`, but **not `.env`** or any other local secret file. With no
`files` field, or with a deny-list `.npmignore`, the variant nobody listed
gets published. The only remedy is rotating the credential.

**Workspace state shadowing the package's own.** A specifier that only
resolves inside the workspace (`workspace:*`, `catalog:`) must never reach
the published manifest. Packing with pnpm rewrites them. Publishing the
directory with npm does not. A manifest that still holds one installs on
nobody's machine but yours.

## Allowlist, not deny-list

The `files` allowlist is the control. Because it is deny-by-default, a new
secret file is excluded automatically. A path that should ship but is
missing is caught **before publish**, by the listing check below.

**For credentials, this is least privilege.** A `.env`, key or token in
the tarball gives every reader, on every mirror, forever, whatever that
credential can do, until it is rotated. Ask the blast-radius question of
every path in the pack listing that could hold one. The allowlist's
deny-by-default is what keeps that answer at zero.

**For everything else** (sources, tests, fixtures), the reason is the
artifact itself, not authority. Each extra file adds to the download every
first `npx` run pays, and serves no consumer. Widen the allowlist on demand:
the listing check flags a missing path, and someone adds it deliberately.

Prefer no `.npmignore` at all. If there is one, remember that at the
package root it does **not** override `files`, but in a subdirectory it
does. That mix is what makes a pattern quietly fail to match.

## Verify it rather than trusting it

**The listing check runs before publish, in the dry-run path too.** Get
the tarball's file list from `pnpm pack --dry-run` (pnpm 10.26 and later)
or from the packed tarball. Compare it against an expected list, and fail
on any difference. The expected list holds:

- the `bin` target and the build output it imports
- `package.json`, `README` and `LICENSE`

Also check that:

- the packed `package.json` has no `workspace:` or `catalog:` specifier
- there is no `.env*`, key or certificate file, `.config/` directory or
  test fixture

pnpm 12.8 and later also warns when a `.env` file is in the tarball
without being listed in `files`. Treat that warning as a failure, not a
notice.

**The smoke test runs after publish and catches different things.** It
installs the published version from the registry and runs it. It catches
an entry point that cannot start, a missing runtime file the entry point
imports, and a dependency that will not install. It does not catch an
extra file that should not have shipped, or a secret in the tarball. By the
time it runs, those are already public, which is why the listing check
comes first.

**The repo's secret scanner does not replace this check.** It scans
tracked files. A file that is correctly gitignored can still be sitting in
`cli/` when the pack runs. With a `files` allowlist, the allowlist decides
what ships, not the ignore files.
