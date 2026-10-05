# npm registry — pick & trade

## What it is for

A tool people install or run on their own machine, not one you host. The
public npm registry is where `npx` looks by default. A consumer types
`npx @virajp.dev/bootstrap` with no registry config, no account and no
token. Being reachable that way is why this target was picked.

## What it costs

- **Nothing in money, for a public package.** A free organization or user
  scope publishes unlimited public packages. Private packages need a paid
  plan.
- **Immutability.** A published `name@version` can never be reused, even
  after an unpublish. Every mistake ships as a new version, so the dry run
  and the pack listing matter more than on a target you can redeploy.
- **Exposure is total.** Publishing is the exposure. There is no network
  layer to hide behind, so the package's access level is the only control.
  A scoped package is private by default, and the first publish must say
  public explicitly.
- **The consumer's machine is the runtime.** Their Node version, their
  operating system and their network decide whether the tool works. The
  manifest's `engines` range is only advisory: npm warns and does not refuse
  unless the user set `engine-strict`.

## The alternative it beats here

GitHub Packages also serves npm packages. But installing from it needs an
authenticated `.npmrc` with a personal access token, even for a public
package. That breaks the one-command `npx` path, which is this product's
whole distribution story. It also makes every consumer mint and store a
credential for a read that needs none. Use it only if every consumer is
already inside one GitHub organization. (Principle: least privilege, where
no actor holds a credential its task does not need.)

## When it stops being the answer

- **Consumers do not have Node.** A tool for people with no Node toolchain
  is better shipped as a standalone binary through a platform package
  manager or release downloads. Wrapping a Node runtime in it is machinery
  for a different product.
- **The package must not be public.** Then it is a scoped, restricted
  package on a paid organization, or a private registry. The access setting
  becomes the whole control, and `npx` needs credentials again.
- **It is really a library.** A library published outside the workspace
  uses this same target, but its contract is an `exports` map rather than a
  `bin`. This package is not a library, and the user has decided to say so
  in the manifest: the import surface is **closed**. The `exports` map
  exposes only `./package.json`, so the `bootstrap` bin is the only
  contract, and nothing can be deep-imported (see the `npm-package-manifest`
  skill's artifact reference). The reasoning is about what can change later:
  - A published import surface is contract-shaped.
  - Closing it after consumers depend on it is a breaking change.
  - Opening it later is additive.

  So it starts closed. (Principles: YAGNI's when-not-to-apply on
  contract-shaped decisions; information hiding, through a narrow published
  interface.)
