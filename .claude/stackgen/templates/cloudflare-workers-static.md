---
slug: cloudflare-workers-static
axis: deploy
kind: cloud-provider
components:
  - cloud-provider/cloudflare@0.1.0
  - cloud-service/workers-static-assets@0.1.3
platforms: []
languages: []
language_facts: {}
optional_languages: []
frameworks: []
dependencies: []
capabilities: []
artifact: static-assets
harness:
  health:
    task: n/a
    mechanism: Two HTTP probes of the deployed origin, because a static host either serves files or does not — there is no process to be unhealthy and therefore no readiness endpoint to write. `/` returns 200 with the built index, and a path that certainly does not exist returns 404; the second is the one worth having, because it is what proves `not_found_handling` actually landed rather than defaulting to a bare edge 404
  e2e_staging:
    task: n/a
    mechanism: A preview deployment — a second Worker name, or a version uploaded rather than deployed — and the suite targets that URL. Which of the two, and where the URL comes from, is the product's decision to make and record; this component states that pre-production is a separate URL rather than a separate configuration of the same one
  local_stack:
    task: n/a
    mechanism: '`wrangler dev` serves the same directory through the same asset routing rules, so `not_found_handling` and the redirect and header files are exercised locally. Its fidelity trap is the edge itself — the custom domain, TLS and the CDN cache do not exist on a laptop, so every routing failure that lives in DNS or the route pattern is invisible until a deployed environment sees it'
materialized: "2026-10-05"
---
# Deploy — Cloudflare Workers Static Assets

A **built directory of files served from the edge**: a site, a docs build,
a single-page app — anything whose whole deployable is what a build step
left on disk. The Worker carries no script at all; the platform serves the
directory, and the deploy is the upload of that directory under one Worker
name.

**The composition is the provider plus one service**, which is what a
Cloud-Bundle is. The provider component carries what spans services — the
account and role model, the token scoping rule, what does and does not
exist locally. The service component carries this one service and **cites**
that rule rather than restating it.

## What this bundle decides that no component decides alone

**The artifact is a directory of files, and there is no promotion by
digest.** A container bundle ships an image it can move from staging to
production untouched; this one has no such handle — the upload is the
release. So the guarantee has to come from the build instead: the same
commit must produce the same directory, which makes a **reproducible
build** this pipeline's real job rather than a nicety. One Worker name per
environment, each fed its own build of the same commit.

**The release runs behind `p:<project>:deploy`, and this bundle ships no
workflow.** The task is the only thing that knows a Cloudflare Worker is on
the other end, which is what keeps the target swappable; the CI system
pinned on the project's `cicd` axis decides what fires it, behind
stackgen's release-trigger contract. Naming the task and writing the
workflow are different jobs, and only the first one is stackgen's.

**Credentials arrive from the environment, never from the config file.**
`wrangler.jsonc` is committed and describes the deployment; the account and
the token that authorize it come from the secrets provider at deploy time.
A config file that carries either is a config file that cannot be read in
review.

**A Worker script fronting these assets is not here.** No `main`, no
`run_worker_first`, no assets binding — the moment code runs in front of
the directory, the deployable stops being a directory and the reproducible
build stops being the whole story. That is the sibling bundle,
[Cloudflare Workers SSR](cloudflare-workers-ssr.md), and the three
Cloudflare deploy bundles are alternatives rather than layers: a `main`
that fronts a container image is
[Cloudflare Containers](cloudflare-containers.md). Among the Astro
project bundles, `astro-ssg` and `astro-csr` pair **here** — every
response decided at build time — while `astro-ssr` and `astro-hybrid`
pair there. Which Cloudflare services stackgen offers, and which are
planned or declined, is the provider component's to state — see the
`cloud-provider/cloudflare` component's conventions, in this
composition's template.

**The seam with [Cloudflare Zero Trust Access](cloudflare-zero-trust.md).**
That bundle produces no artifact and "composes with a hosting pin rather
than replacing one" — this is a hosting pin it composes with. A `site` that
must not be publicly reachable pins **both**: this one decides how the
files get served, that one decides who is allowed to reach them. Since
`config_format` 16 made `deploy_template` a list, pinning two is
representable, and pairing them is vwf's job.

Full judgment: the components' own skills and their references.

# Cloudflare — conventions

The provider half of the Cloud-Bundle: what holds across every Cloudflare
service the product uses, carried once so no service component restates it.

**The coverage here is bounded, and saying where the boundary falls is
part of the component.** A menu that comes back short without explaining
itself is indistinguishable from a broken one, so what is offered and
what is declined are stated below rather than implied. Do not fill a gap
from general Cloudflare knowledge: a service this component has not
written doctrine for is a service it does not offer.

**Offered.** Three deploy targets — **Workers Static Assets**, **Workers
SSR** (a Worker with a script in front of its own assets) and
**Containers** (a container image running beside a Worker) — plus **Zero
Trust Access** for the private plane, the storage and data services —
**Workers KV**, **R2** (including R2 Data Catalog and R2 SQL), **D1**,
**Hyperdrive**, **Vectorize**, **Pipelines** and **Analytics Engine** —
compute and orchestration — **Durable Objects**, **Workflows** and
**Queues** — AI: **Workers AI**, **AI Gateway**, **AI Search** and
**Browser Rendering** — and media, messaging and secrets: **Images**,
**Realtime**, **Email Service** and **Secrets Store**. Each is its own
service component and its own bundle; they are pinned side by side, not
chosen between, with one exception — Containers is pinned *instead of*
Workers SSR, never beside it, and the `cloudflare-containers` bundle
carries the reasoning.

**The Agents SDK is offered too, and it is the one that is not a
service.** It is an npm framework that compiles to a Durable Object, so
it ships as a framework component on the project axis —
`framework/cloudflare-agents`, reached through the
`typescript-cloudflare-agents` language bundle — and a reader looking for
it under `cloud-service/` will not find it there. A project that is an
agent pins that language bundle on its project axis and the
`cloudflare-durable-objects` bundle on its backing axis; the object is
what the agent runs as.

**Nothing is planned-but-missing any more.** The developer-platform
coverage this component set out to carry is complete: every Cloudflare
surface a repo composes its stack from is either offered above or
declined below. A service not named on either list is out of scope by
decision, never by omission — so a product that needs one has a gap to
name and a decision to reopen, not an oversight to route around.

**The runtime secrets store and the repo's secrets provider are two
different things, and they coexist.** **Secrets Store** is the
account-level store a deployed Worker or Container reads through a
binding in staging and production; `capability-provider/fnox`, on the
capability axis, is what holds a developer's and CI's secrets on the
way in. A repo pins both, for different environments, and neither
replaces the other. Which clause of the secrets contract each satisfies
is the `secrets-store` component's doctrine to state, not this one's.

**Declined, and they are not coming.** Pages is superseded by Workers
Static Assets in Cloudflare's own guidance; Workers Sites is deprecated
in Wrangler v4; Stream and Turnstile were offered and declined.
Account-level products — WAF, DNS, Tunnels, Zaraz, Logpush, Workers for
Platforms and their kind — are configuration of an account rather than
components of a repo's stack, so they have no place in this model at
all.

**Cloudflare hosts what it can serve from the edge and fronts everything
else.** At the scope offered here it hosts three shapes — a built
directory of files, on Workers Static Assets; that directory with a script
in front of it, on Workers SSR; and a container image beside a Worker, on
Containers — and for anything with a running process of a kind none of
those can hold, it fronts what runs on another cloud. That second half
inverts the usual reading of a Cloud-Bundle and is the single fact most
likely to be got wrong: a service or fullstack project none of the three
fits pins its hosting elsewhere and pairs the private plane with it, which
is vwf's job, and any cloud's own deploy bundle composes with it.

**The account is the unit of blast radius, and the roles are broader than
they look.** Grants are account-scoped, so a role handed out to edit one
application reaches every application in the account. `Cloudflare Access`
— which edits Access applications, policies and Tunnels — is the narrow
grant for the private plane; `Cloudflare Zero Trust` is administrator
over every Zero Trust product and is not the same request. Which
permission each other service needs is that service component's
identity-shape reference to state, not this one's. Automation uses an
account-owned API token scoped to what it touches, never the Global API
Key, which is unscoped and carries the account.

**Billing follows the population, not the traffic.** A seat is consumed per
user allowed through, and it is freed by removing the user from the seat
rather than by their access expiring on its own. That shape is what makes
this cheap for an operator plane and the wrong answer for anything
customer-facing — the same conclusion the scoping rule reaches from the
other direction, and the reason offboarding is a billing event as well as a
security one.

**The private plane does not exist locally and must not be simulated.**
Cloudflare ships no emulator for the identity-aware proxy, and a local
stand-in would prove only that the stand-in works. Local runs reach the
project directly and inject the identity assertion as a fake through the
same seam the project already verifies in production.

Full judgment: the `cloudflare` skill and its references. The services
this provider carries are the `cloud-service` components named in the
offered list above, each under its own slug — `zero-trust-access`,
`workers-static-assets`, `workers-ssr`, `containers`, `kv`, `r2`, `d1`,
`hyperdrive`, `vectorize`, `pipelines`, `analytics-engine`,
`durable-objects`, `workflows`, `queues`, `workers-ai`, `ai-gateway`,
`ai-search`, `browser-rendering`, `images`, `realtime`, `email-service`
and `secrets-store` — plus one framework component,
`framework/cloudflare-agents`.

# Cloudflare Workers Static Assets — conventions

An **assets-only Worker**: the build output directory is the whole
deployment. There is no `main`, so no script runs; the edge matches a
request against the uploaded file set and serves it. `wrangler deploy`
uploads the directory and that is the release.

**This is a hosting pin, and it produces an artifact.** Unlike
`zero-trust-access`, which fronts something that runs elsewhere, this is
where the project actually runs — which makes it the entry a `site`
project pins on the `deploy` axis. The two compose: a static site behind
the identity-aware proxy is both pins on the same axis, and neither
replaces the other.

## What this component writes

**`wrangler.jsonc` at the repo root**, not under `.config/`. Wrangler
discovers its configuration by walking up from the working directory to a
`wrangler.jsonc` / `wrangler.toml`, and it has no ambient way to be told
otherwise — the alternative is `--config .config/wrangler.jsonc` on every
invocation any caller might ever type, which is a flag someone eventually
forgets and then deploys from a config that does not exist. The root
allowlist in stackgen's output charter admits the file for exactly that
reason; being on the list makes it landable, not standard, and the three
Cloudflare deploy packs ship one
— this pack, `workers-ssr` and `containers`.

**`.config/mise/tasks/p/<project-id>/deploy`**, an overlay in the
project's own task group. It ships as `p/_project/deploy` — a marked
directory name, not a task — and the command that pins this stack renames
the directory to the project's registry id. Until it is renamed the task
is inert rather than wrong: mise ignores a task directory whose name
starts with an underscore, which is the same rule that keeps `_scripts/`
out of `mise tasks`.

**Two marked positions in `wrangler.jsonc`**, both filled by the pinning
command and neither guessable by the pack: the Worker `name`, and the
`routes[]` entry that binds it to a custom domain. Everything else ships
with a real value, because everything else is this component's judgment
rather than the repo's identity.

## The values `wrangler.jsonc` ships

- **`$schema`** is relative to the file, so a repo that installs wrangler
  inside a sub-project points it at that project's `node_modules`.
- **`name`** is account-unique and lowercase — letters, digits and dashes —
  and derived from the project rather than the domain, so a repo that moves
  domains keeps its Worker. The shipped `PLACEHOLDER` is deliberately
  invalid, so an unfilled slot fails at the first deploy instead of
  publishing a Worker nobody meant to create.
- **`compatibility_date`** is a date, not a version: pinning it stops a
  future runtime change from altering an already-shipped deployment. Move
  it deliberately and read the changelog for the span skipped.
- **`assets.directory`** is relative to the file; a repo whose site is a
  sub-project points at that project's output (`./site/dist`).
- **`not_found_handling: "404-page"`** serves `404.html` with a 404 status,
  the right answer for a multi-page site. A client-routed single-page app
  wants `"single-page-application"`, which serves `index.html` with a 200
  so the router can take the path; picked for a multi-page site, it makes
  every typo a 200 and drops the site from search results.
- **`routes`** ships commented out. Filling it means uncommenting it with
  the hostname the site answers on; the zone must already be on the same
  account, and `custom_domain` is what makes wrangler create and manage the
  DNS record. A repo with no custom domain deletes the block, and the
  Worker answers at `<name>.<account-subdomain>.workers.dev` — a complete
  deployment for a preview surface or an internal tool.

## Credentials

`wrangler` reads **`CLOUDFLARE_API_TOKEN`** and
**`CLOUDFLARE_ACCOUNT_ID`** from the environment. They are account-wide
values shared across every repo that deploys to the account, so the
secrets convention names them **`GLB_CLOUDFLARE_API_TOKEN`** and
**`GLB_CLOUDFLARE_ACCOUNT_ID`** in the secrets provider, per stackgen's
secrets contract, and the provider supplies them to the process under the
names wrangler expects. They never
appear in `wrangler.jsonc`, and the deploy task refuses to start without
them rather than letting wrangler fail with an auth trace that reads like
a network problem.

`wrangler login` is the interactive alternative and is a developer's
convenience only. It stores an OAuth grant on one laptop; CI has no
browser and no laptop, and a pipeline that depends on someone's grant is
one that breaks when they leave.

**`--dry-run` needs no credentials.** It neither authenticates nor
uploads, so the deploy task skips the credential check for it — requiring
them would stop a contributor without account access from ever validating
the config, which is the one thing the flag exists for.

## The pipeline

**The task CI must run is `p:<project-id>:deploy`.** The workflow that
calls it is the repo's own — a pack states the task name and never writes
the workflow, which is stackgen's output charter's fence. Nothing here
decides the trigger either; that belongs to the CI
system pinned on the project's `cicd` axis.

The task does not build. It runs `p:<project-id>:build` when that task
exists and otherwise assumes the output directory is already built,
because what produces the directory is the framework's business and not
this component's.

## The artifact contract

**A directory of files** — `./dist` by default, overridable in one place
when a framework disagrees. **That default is not this pack's guess — it is
the framework pack's stated fact**, under the heading `## Build output` in
the project bundle's framework component, which says where the build writes
and that a deploy target may rely on the path. `framework/astro` is the
specimen that states it today; any framework pack stating the same fact
under the same heading pairs here the same way, and one that states a
different path is one whose `assets.directory` differs by that much.

- **Fingerprinted assets are immutable.** Where the framework hashes
  content into the filename, those paths get a long `max-age` with
  `immutable`; the entry HTML does not, or a deploy is invisible to every
  browser that already has it.
- **`404.html` at the directory root** is what `not_found_handling:
  "404-page"` serves, so the build has to emit one. A missing file turns
  every unknown path into a bare edge 404 with no branding and no
  navigation, and nothing reports it.
- **One deploy is distinguishable from the next** by the uploaded file
  set, which is what makes the rollback path a version rather than a
  rebuild.

## What is explicitly not here

**No Worker script.** No `main`, no `assets.binding`, no
`run_worker_first`. Those are the shape where code fronts the files —
server-side rendering, an API route beside the site, an auth check at the
edge — and that shape is **`cloud-service/workers-ssr`**, a separate pack
and a separate pin. The two are alternatives rather than layers: a
deployment either has a `main` or it does not.

**No other Cloudflare service is this component's to speak for.** Which
Cloudflare services stackgen offers, and which are planned or declined, is
the provider component's to state — see the `cloud-provider/cloudflare`
component's conventions, in this composition's template.

**No wrangler pin.** Wrangler is a development dependency of the project
that deploys, declared in that project's language manifest — and a
manifest is outside the config tier's fence. The deploy task calls the
manifest's wrangler through the package manager rather than a globally
pinned binary, so the version CI runs is the version the lockfile
records.

Full judgment: the `workers-static-assets` skill and its references. The
provider-wide doctrine it cites — the account model, the role grants,
seat-shaped billing, the private plane — is the `cloudflare` skill's. The
shape with a script in front of the files is
`cloud-service/workers-ssr`.
