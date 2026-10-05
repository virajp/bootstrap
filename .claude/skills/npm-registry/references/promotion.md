# npm registry — promotion & release

## There is nothing to promote, so the version is the record

A registry has no staging host and no production host. There is one
registry, and a published version is what every consumer gets. The
guarantee that promotion gives elsewhere, that the tested artifact is the
shipped one, comes from **one tarball** here. The release packs it once,
lists it, smoke-tests what the registry serves back, and never rebuilds
between the dry run and the upload.

**Staging, if the product ever needs it**, is a prerelease version on the
same line (`1.4.0-rc.1`) under a non-`latest` dist-tag. npm refuses to
publish a prerelease without an explicit `--tag`, so a prerelease cannot
take over `latest` by accident. The release intent for this product is a
**single release once the feature set is done**, so there is no prerelease
channel. Do not build one until a release needs it (principle: YAGNI).

## The release is the repo's task

The release is wrapped in the repo's task library, so the same command
runs locally and in CI. The pipeline calls the task, never `npm publish`
directly.

This repo groups project tasks under `p:` (`.config/mise/tasks/p/`). It
already has an installer pair, `p:i:release` and `p:i:publish`, which by
the user's decision become `p:cli:release` and `p:cli:publish`:

- `p:cli:release` tags what `main` already carries. The CI system owns the
  pipeline it starts.
- `p:cli:publish` is the publish step. It has a `--dry-run` flag.

stackgen's release-trigger contract would spell the name
`<project>:release:<environment>`. The repo's existing names take
precedence over the contract's suggestion. **Confirm the tasks exist under
their new names before wiring anything to them.** A guessed task name fails
at publish time.

### The publish task must be rewritten, not just renamed

The existing `p:i:publish` cannot serve as the trusted-publishing release,
for three reasons:

- It exits unless `npm whoami` succeeds. Under OIDC there is no login
  session to ask about, because npm exchanges the token inside
  `npm publish` itself. The same gate also makes the dry run fail without
  credentials.
- It runs `npm publish` on the directory, which would ship `workspace:` and
  `catalog:` specifiers unresolved.
- It skips none of the checks below.

**Open item for `/vwf:plan`:** rewrite `p:cli:publish` to the behaviour
below. Until that lands, CI must not be wired to it.

Required behaviour, in order:

1. **Check the version match.** Read the version in `cli/package.json` and
   refuse to continue unless it equals the semver part of the release tag
   the pipeline passes in. The tag's grammar is the CI system's to parse.
   This step only compares the two numbers.
2. **If already published, verify instead of publishing.** Ask the
   registry whether `@virajp.dev/bootstrap@<version>` exists. If it does,
   skip steps 3 to 5, but **still run the smoke test (step 6)** against the
   published version. The re-run then passes or fails on that, so it
   reports the original outcome rather than an unconditional success. A
   404 for a name that has never been published means "not published". Any
   other registry error fails the task, because an unknown answer is not a
   "no". (Principle: idempotency, where a replay returns the original
   outcome.)
3. **Pack once.** Run `pnpm pack` in `cli/` into a scratch directory. pnpm
   rewrites `workspace:` and `catalog:` specifiers to real version ranges
   and applies `publishConfig` overrides as it packs.
4. **List the tarball.** Print its file list. This is the hygiene check
   the `npm-package-manifest` skill describes.
5. **Publish that tarball** with `npm publish <tarball>`. npm is used here
   because the registry documents trusted publishing for npm (CLI 11.5.1 or
   later, on Node 22.14.0 or later). Since pnpm 11, `pnpm publish` no longer
   calls npm, and its docs do not say whether it does the OIDC exchange.
   pnpm's own documented fallback is the same pack-then-`npm publish`
   pairing. Publishing a tarball runs no lifecycle scripts, which is right:
   the build already ran.
   **A race lands on the same path.** If a concurrent run published first,
   `npm publish` refuses with "You cannot publish over the previously
   published versions". Treat that refusal like step 2's "already
   published": go to the smoke test, do not hard-fail. (Principle:
   idempotency.)
6. **Smoke-test what consumers will get** (below). This runs on every
   non-dry-run path, including the already-published ones.

**Credentials depend on the mode. There is no blanket login gate.**

- **CI mode.** The job can request a GitHub OIDC token. GitHub's OIDC
  reference documents that the runner sets `ACTIONS_ID_TOKEN_REQUEST_URL`
  and `ACTIONS_ID_TOKEN_REQUEST_TOKEN` when the job has `id-token: write`,
  so the task uses `ACTIONS_ID_TOKEN_REQUEST_URL` being set as the
  signal. The task runs no credential
  pre-check, and `npm publish` performs the trusted-publisher exchange
  itself.
- **Break-glass mode.** This is a local, non-dry-run publish with no OIDC
  token available. It is the only mode that requires an interactive
  `npm login` session, and it publishes with a 2FA prompt. See
  [Break-glass](#break-glass-the-maintainers-interactive-publish).
- **`--dry-run`.** Requires no credentials in either mode.

**`--dry-run` runs steps 1 to 5 with `npm publish --dry-run`.** npm still
packs, applies the prerelease-tag and version-exists rules and prints the
tarball. Without credentials it warns instead of failing. So a contributor
with no registry access can dry-run a release, and that is what the flag is
for. The dry run skips the smoke test because there is nothing published to
install.

Check the npm version inside the task, not in a workflow comment. If the
Node release pinned in the repo's toolchain bundles an npm below 11.5.1,
publishing fails with an authentication error that looks like a credentials
problem. Fix it in the toolchain config, not with a global install in the
pipeline.

## Trusted publishing: the one credential is none

Publishing uses the npm trusted publisher for this package: the GitHub
repository, the **workflow filename** (exactly, with `.yml`,
case-sensitive) and, optionally, a GitHub environment. The job exchanges a
short-lived OIDC token at publish time, so there is no npm token to store,
rotate or leak. Provenance is attested automatically for a public package
built in a **public repository** on a **GitHub-hosted runner**:

- A private repository gets no provenance, even for a public package.
- Self-hosted runners are not supported at all.

(Principle: least privilege, where the narrowest credential is no stored
credential.)

The id-token permission, the job layout and the filename staying stable
belong to the CI system. This target sets one requirement on the
pipeline: the workflow file that runs the publish must be the one named in
the trusted-publisher config. If another workflow calls it as a reusable
workflow, the registry checks the caller's name instead and matches nothing.

**Bind the trusted publisher to a GitHub environment** if the pipeline
uses one for releases. Then a workflow that never entered the environment
cannot mint a publish token, even with the right filename.

## The first publish: a one-time bootstrap

**A trusted publisher can only be configured on a package that already
exists.** So `@virajp.dev/bootstrap` cannot go out on its first publish
through OIDC. The bootstrap, once:

1. The maintainer publishes a **placeholder version** by hand, with account
   2FA, using `--access public` because a scoped package is private by
   default. The placeholder is a version with no real content and a README
   that says the release is coming. No access token is created for it.
   The maintainer logs out of npm afterwards, so the login session does not
   stay behind either.
2. Configure the trusted publisher on the package (on npmjs.com, or with
   `npm trust github`, which needs npm 11.15.0 or later and account 2FA).
3. In the package's publishing-access settings, **require 2FA and disallow
   tokens**. This blocks traditional tokens and leaves the OIDC path open.
4. After the real release ships, **deprecate the placeholder** so nobody
   pins it.

**This is a decision for the maintainer, not a default to apply
silently.** The alternative is to hand-publish the real first release.
That avoids a placeholder, but the one version every early consumer runs
then has no provenance. That conflicts with the release intent, so the
placeholder is the recommendation.

**Set `publishConfig.access` to `public` in the manifest** so access never
depends on someone remembering the flag. npm reads `publishConfig` at
publish time, and pnpm keeps it in the packed manifest.

## Staged publishing: available, not adopted

npm can stage a publish (`npm stage publish`). The workflow uploads, and
nothing goes live until a maintainer approves it with 2FA. pnpm 11.3 and
later supports it too. A trusted publisher can allow staged publishing
alongside direct publishing, or instead of it.

**What it defends against: a compromised pipeline.** Examples are a
malicious step, a hijacked third-party action, or a poisoned dependency in
the release job. Any of these can mint the OIDC token and publish under the
package's name, with valid provenance, because the trusted publisher trusts
the workflow, not the person. Staging splits that authority. CI can only
propose a version, and only the 2FA holder can make it public. That is
least privilege applied to the publish itself.

**Why the risk is accepted for now:**

- There is one maintainer. That person is both the approver and the person
  who pushes the release tag, so the second key would sit in the same hands
  as the first.
- One release is planned. The exposure window is a single publish, and the
  maintainer watches it run, followed by the install smoke test.
- The trigger is narrow. The publish runs only from the release workflow on
  a release tag, and the CI system's tag rules restrict who can push one.
- The residual risk is real: a compromised CI step during that one run
  could publish a malicious version with valid provenance. Accept it as a
  decision, not as an oversight.

**Revisit triggers.** Adopt staged publishing when any of these becomes
true:

- a second person can push release tags or edit the release workflow
- releases become routine rather than a watched single event
- the release job gains third-party actions or dependencies beyond the
  repo's toolchain
- a supply-chain incident hits an action or package the release job uses

## Break-glass: the maintainer's interactive publish

When the CI path is broken (the trusted-publisher config, the runner, or
the workflow), a release that cannot wait goes out by hand. It uses the
rewritten `p:cli:publish` in its break-glass mode: run locally, with no
OIDC token, after an interactive `npm login`. This keeps the one local
fallback the repo already has (`p:i:publish` checks for a login session),
but runs it through the same pack-then-publish steps. Its own controls:

- **Holder.** Only the package's maintainer, who holds the account 2FA.
  Nobody else is a package owner.
- **Authentication.** The "require 2FA and disallow tokens" setting stays
  on. npm still allows an interactive publish under it, as long as the
  maintainer answers a 2FA prompt for that publish. No token is created to
  make the fallback easier.
- **The same task.** Use the same dry run, pack listing and version checks,
  and run the install smoke test afterwards.
- **Afterwards.** Log out of npm so the login session does not outlive the
  emergency. Record why CI was bypassed. Fix the CI path so the next release
  goes back through the trusted publisher.
- **The cost, stated.** A hand-published version has **no provenance**.
  Every use of this path leaves one version that consumers cannot trace to
  a build, which is why it is break-glass and not a second release path.

(Principle: least privilege, whose when-not-to-apply says break-glass must
exist, with named holders, its own controls and an audit trail.)

## The install smoke test is the last step

It is the only check that tests what consumers actually resolve, rather
than what the workspace built:

1. **Wait for the version to be visible.** `npm view` fetches once and
   never retries, so poll it, with a time limit, until
   `@virajp.dev/bootstrap@<version>` answers.
2. In an **empty temporary directory**, outside the workspace so no local
   copy can shadow it, run `npx --yes @virajp.dev/bootstrap@<version>`
   with a harmless entry point such as the version flag. Fail unless it
   prints the version that was just published. Pin the version: an unpinned
   `npx` can resolve whatever `latest` is.
3. Install the same version into that directory and run
   `npm audit signatures --json --include-attestations` there. The command
   verifies the registry signatures and attestations of the installed
   packages. With these flags its JSON output has a `verified` array, which
   holds the attestation bundles for each verified package. Fail unless
   `@virajp.dev/bootstrap@<version>` has an attestation bundle in that
   array. The npm docs do not say how to tell a provenance attestation from
   the registry's publish attestation in this output. So treat this check
   as "an attestation is present and verifies", and use the package page's
   provenance check mark as the human confirmation that it is provenance.
   A **break-glass** version legitimately has no provenance, so in that mode
   this step runs signature verification only and reports the missing
   provenance as expected.

A failure here does not unpublish anything, and it cannot. It marks the
release broken, and the fix ships as the next patch. If the broken version
must not be run, deprecate it.

## Rollback is supersede, never unpublish

An unpublished `name@version` can never be reused. Unpublishing every
version locks the name for 24 hours. After 72 hours, unpublishing is
limited anyway. The remedy for a bad version is the next patch, plus
`npm deprecate` on the bad one so installs warn. The release task never
calls unpublish.

## The pipeline is not written here

Releases follow **vwf's delivery-pipeline contract**: tag-triggered,
branch-validated, and tested before release (the tagged project and its
dependents). The **CI system** implements it under stackgen's
release-trigger contract, which sets the tag shape. That contract's shape
for this project is `cli-prod-v<semver>`. If the product records a
different tag shape (for example a bare `v<semver>`), the record wins, and
this target is indifferent to which. It requires only that the tag's
semver equals the manifest version (step 1).
