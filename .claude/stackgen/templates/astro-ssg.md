---
slug: astro-ssg
axis: project
kind: language-bundle
components:
  - language/typescript@0.3.1
  - package-manager/pnpm@0.6.1
  - toolchain-gate/tsconfig@0.2.2
  - toolchain-gate/eslint@0.3.4
  - framework/astro@0.5.1
  - framework/react@generated
platforms:
  - site
languages:
  - typescript
  - javascript
language_facts:
  typescript:
    lsp: typescript-language-server, run through mise; declared by this pack as `typescript-lsp` and provisioned via the generated local plugin
    mise_tool: node
    manifest: package.json
  javascript:
    lsp: typescript-language-server, run through mise; the same `typescript-lsp` server serves both tokens
    mise_tool: node
    manifest: package.json
optional_languages: []
frameworks:
  - astro
  - react
dependencies: []
capabilities: []
artifact: n/a
package_manager: pnpm
harness: n/a
materialized: "2026-10-06"
---
# site — Astro (SSG)

A **content surface built once and served as files**: a marketing site, a
documentation build, a changelog, a landing page. Every route is rendered at
build time, so nothing runs per request and there is no server to operate — the
deployable is a directory.

**A static site publishes no API.** It may call someone else's, from the
browser, at the cost of that call being public; a project that owns an API
contract is `fullstack`, and one that must read the request before it can
answer is [`astro-hybrid`](astro-hybrid.md) or [`astro-ssr`](astro-ssr.md)
instead. An app whose state lives in the browser behind one shell page is
[`astro-csr`](astro-csr.md). This bundle is the entry a `site` project's
architecture round preselects; its three siblings are picked deliberately.

**What Astro is, and how its modes differ, is the `framework/astro` pack's
doctrine** — this bundle does not restate it. What the bundle pins is the mode:
`output: "static"` with **no adapter**, which is the pack's `ssg.md`.

This doc covers the **project axis** only; backing services and deploy target
are their own axes.

## Stack

- **Framework**: Astro on `output: "static"`, the default. No adapter is
  installed, and installing one is the signal that this is the wrong bundle: an
  adapter exists to render on demand, and nothing here does.
- **Routing and content**: file routes under `src/pages/`, and **content
  collections** for anything that is a body of documents — the schema on a
  collection is what turns a typo in frontmatter into a build failure rather
  than a blank page. The pack's `content-and-routing.md` carries the shape.
- **React is present, and mostly unused.** `@astrojs/react` ships in the bundle
  so an island is a decision rather than a migration, but **a page with no
  island ships no JavaScript** — that is the whole reason to be on this mode,
  and reaching for a component where markup would do gives it away for nothing.
  Hydrate with the narrowest directive the interaction survives.
- **Layout**: `src/pages/` (routes), `src/content/` (collections),
  `src/components/`, `src/layouts/`, `src/lib/` (pure helpers the build calls).
- **Config**: values are read at **build** time, so a "secret" in this project
  is a build input and anything reaching the browser is public. Set the site's
  canonical origin in the Astro config — sitemaps and canonical URLs derive
  from it, and without it a static build silently emits none.

## Build output and deploy

The build leaves **a directory of files at `./dist`** — the `## Build output`
fact in the `framework/astro` pack's conventions, which is what a deploy pack's
asset directory cites rather than guessing.

**Pair with `cloudflare-workers-static`**, the bundle this mode was built for:
its deployable is exactly that directory, no script runs in front of it, and
its reproducible-build rule is what stands in for promotion by digest. Any host
that serves a directory works the same way; what the pairing decides is the
trailing-slash and not-found behaviour, and those must agree with what the
Astro config declares.

## Testing

**Vitest on the node environment** for `src/lib/` and anything the build calls
— that is where a static site's logic actually lives. **jsdom + Testing Library
only for the islands a repo actually writes**, since a bundle-wide jsdom
default buys nothing on a site that ships no client JavaScript. Scope the
coverage include to `lib/` and `components/`, excluding `.astro` shells, and
let the repo set the threshold. A build that fails is a test too: a broken
content-collection schema or a dead route is caught by `astro check` in the
repo's gate, not by a unit test.

# TypeScript — conventions

The Node/TypeScript baseline. New code is TypeScript with `strict` on;
JavaScript files get the same standards minus the type-level rules and are
migration candidates, never an excuse to relax them.

**Errors are values at the boundary and exceptions in the middle.** One mapping
home turns internal failures into the product's coded responses — never
scattered `try`/`catch` that each invent their own shape.

**`async`/`await` throughout; never block the event loop.** CPU-bound work moves
off the main thread rather than being awaited around.

**Tests are Vitest**, colocated, with the shared config and v8 coverage.

**The `@/` path alias, no deep relative chains, and a clean→check→build
pipeline.** Barrels are for public surfaces only.

**Config is read once at the composition root** — names-not-values, catalogued
in `docs/blueprint/environment.md`, never `process.env` reads scattered through
the code.

**Telemetry is OTLP.** The product emits it and never imports a vendor SDK.

Full judgment: the `typescript` skill's references.

# pnpm — conventions

pnpm is the only package manager. A repo with two lockfiles has two dependency
graphs and resolves differently depending on who ran what.

**The lockfile is committed and authoritative.** CI installs frozen and fails on
drift rather than resolving something new — an install that can resolve
differently in CI than locally is not a gate.

**A publish cooldown guards the supply chain**, so neither a routine install nor
an automated update adopts a release published minutes ago.

**In a workspace, internal dependencies are linked, not versioned**, and shared
versions live in a catalog so one bump moves every package.

**Two settings ship as `.npmrc` at the repo root**, which is the one path the
manager reads them from: `ignore-scripts=true`, so an install never executes a
dependency's install-time code, and `fund=false`, so it never prints a banner
over what it did. A dependency that genuinely has to build is allowed by name
in `pnpm-workspace.yaml` (`allowBuilds`, or `onlyBuiltDependencies` before
pnpm 10.26) — the exception is a reviewable line, not a switch.
Beside it, the pack's template
`templates/.config/mise/conf.d/pnpm/mise.dev.toml`, rendered into the repo by
`stackgen:tool-config`, aliases `npx` to `pnpm dlx` in dev, so a one-off
package runs through this manager's store, resolver and registry settings
rather than another tool's.

**An agent's `npm`/`npx` command is rewritten before it runs.** This pack
ships `hooks/npm-normalize.sh`, which lands at `.claude/hooks/npm-normalize.sh`
and — once its `hooks.yaml` entry is accepted into `.claude/settings.json` —
resolves the repo's manager from its lockfile and rewrites the command to it.
Declining the settings entry leaves the script landed and inert, which is
safe: the hook only rewrites a command that was going to run the wrong manager
anyway. It allows exactly two managers, pnpm and bun (`npx` → `pnpm dlx` or
`bunx`, `npm ci` → `<pm> install --frozen-lockfile`, any other `npm` → `<pm>`,
flags after `npx` kept verbatim), and resolves which one by walking up from the
working directory: a lockfile first — `bun.lock`/`bun.lockb` or
`pnpm-lock.yaml`, the ground truth, since bun reuses npm's `workspaces` field
and nothing else tells them apart — then `package_manager: bun` in
`.config/vwf.yaml`, for a project scaffolded but not yet installed, then pnpm,
because the hook fires in every repo, including ones that never heard of vwf.
Its `sed` stays BSD-compatible: no `\s`, no `\b`.

## The task library this pack owns

This pack ships a `config/.config/mise/tasks/` tree — the `code/format/pnpm`
subtask and the five `setup/deps/<verb>/pnpm` subtasks — landing at the
repo's own `.config/mise/tasks/` behind the materializer's config consent
line. Every file is named for the pack, so no other component writes the same
path: the repo's `code:format:all` and `setup:deps:<verb>:all`, which
tool-config renders, call each subtask by name.

**`code:format:pnpm` sorts every `package.json`, and only that.** dprint runs
beside it as the universal `code:format:dprint` subtask, so this file carries
no formatter step of its own. The sorter, `npm:sort-package-json`, is a mise
pin in the dev environment only, in the same `mise.dev.toml` template —
resolved by its mise path, and the step is skipped where it is not installed,
as in CI.

**The subtask takes an optional file list, and the empty case is the whole
tree.** That is the whole pre-commit story for this pack: it ships **no
fragment**, because the gate config's `format` and `lint` hooks call
`code:format:all` and `code:lint:all` with the staged files. The sorter
narrows to the `package.json` files it is given. Linting is the house linter's,
through the universal `code:lint:house`, which runs the whole tree either way —
its rules are cross-file — and every exclusion it needs lives in
`.config/linter.yaml`, which `stackgen:tool-config` lands.

**The `setup/deps/*` verbs are `install`, `outdated`, `audit`, `upgrade` and
`cleanup` — all five slots.** `install` is `pnpm install --recursive`, because a
workspace install that stops at the root leaves the repo half resolved.
`cleanup` deletes `dist`, `node_modules` and `*.tsbuildinfo`, then prunes the
store — a store left behind makes the next install look clean when it is
replaying — and deliberately leaves the lockfile alone: the lockfile is an input
a human reviews, and moving it forward is `upgrade`'s job. The optional verbs
are **probed by name**, so a missing file is itself the answer: a manager that
ships no `upgrade` has no such verb, not a choice still pending.

`install --frozen` is the contract's name for "the lockfile is the input, not
the output" — what a fresh worktree and CI want, turning a stale lockfile into
a failure rather than a silent rewrite. `audit` is advisory and never a gate:
`pnpm audit` reads a registry feed that moves without any lockfile change, and
the blocking supply-chain check is `code:sec`, which runs pinned tools.
`outdated` swallows its exit status, since `pnpm outdated` fails whenever it
finds anything — a healthy repo's normal state. `upgrade` updates pnpm itself
first, because a resolver a major version behind writes a lockfile the current
one then rewrites, and passes `--latest` on purpose: the ranges say what still
works, and this task is where a person decides something newer should — the
diff is the review surface. `cleanup` prunes the store's `.pnpm` link farm
rather than deleting it, which would re-download every unchanged package. The
verbs print no header of their own; `setup:deps:all` frames each.

`code:format:pnpm`'s sorter pair is the inverse of dprint's — sorting is its
default, `--check` its read-only mode — and the sorter, like the house linter,
runs by its mise path so a package in `node_modules/.bin` cannot shadow the
pin.

Full judgment: the `pnpm` skill's references.

# tsconfig — conventions

**`strict` is on, everywhere, and is not negotiated per project.** The
type-level rules in the TypeScript baseline assume it; without it they are
suggestions.

**One shared base config, extended per project.** A per-project config that
restates the base has already drifted from it.

**The `@/` path alias** replaces deep relative chains, and the build resolves it
the same way the editor does.

**A separate emit variant for builds**, so type checking and emitting are
distinct operations — `tsc --noEmit` is the checker, and nothing about a check
should depend on output settings.

## What this pack writes

No file. This pack's conventions and skill guide the agent when it writes the
`tsconfig*.json` files, which are per-project and belong to the project —
written where the project is, not laid down from here.

Full judgment: the `tsconfig` skill.

# ESLint — conventions

The **correctness** gate for TypeScript and JavaScript. Topic 10 of the language
bundle, deliberately not a repo gate: a linter meaningful for exactly one
toolchain belongs to that toolchain's bundle, or a polyglot repo acquires one
per language.

**Flat config only.**

**Zero formatting rules.** The formatter owns layout — a rule a formatter can
satisfy must never be able to fail a lint run. The dprint config
`stackgen:tool-config` lands is the other half of that split.

**Overrides are scoped by `files` glob**, never disabled globally. A rule turned
off everywhere because one file could not satisfy it is a rule the repo no
longer has.

**One lint command, wired through the task library**, so local and CI run the
identical gate.

## What this pack writes

No file: its skill and these conventions. ESLint runs only inside the house
linter, `@askviraj/linter`, which `stackgen:tool-config` pins in every repo
and runs through the universal `code:lint:house` subtask — so this pack ships
no `code/lint/eslint` subtask and **no pre-commit fragment**: the gate
config's `lint` hook calls `mise run code:lint:all --fix`, which runs
`code:lint:house` with every other lint subtask. The house linter reads the
whole tree whatever list it is given — its rules are cross-file.

`.config/linter.yaml`, the linter's own config, is **not** this pack's either:
`stackgen:tool-config` lands it with the other gate configs, because the house
linter reads it in every repo, not on this stack alone. It lands **empty of
overrides** — the linter is zero-config without it, so the file exists to give
a misfiring default one obvious place to be answered — with an `ignores:` list
of the generated trees the stack packs produce. The `eslint` skill still guides
every edit to it.

**A disable comment goes on its own line above the offending one**, line style,
so the decision is visible, with its reason written by hand.

Full judgment: the `eslint` skill.

# Astro — conventions

Astro layers **on top of** the TypeScript baseline rather than replacing it:
the baseline's rules — `strict`, the `@/` alias, one mapping home for errors,
config read once at the composition root, Vitest — apply to every `.ts` and
`.tsx` file the site holds. Astro adds the `.astro` component file, the file
router, content collections, and the build.

**Astro owns the build.** No other bundler config competes with it: a Vite
plugin a project needs goes inside Astro's `vite` block, never in a second
Vite config beside it.

## The four modes

Astro has exactly **two** `output` values. The four project shapes below are
decisions over those two values plus two more: whether an adapter is present,
and whether the pages are content or an application.

| Mode   | `output`   | Adapter  | Renders                                                    |
| ------ | ---------- | -------- | ---------------------------------------------------------- |
| SSG    | `static`   | none     | every route prerendered at build time                      |
| Hybrid | `static`   | required | prerendered by default; `prerender = false` opts a route in |
| SSR    | `server`   | required | on demand by default; `prerender = true` opts a route out  |
| CSR    | `static`   | none     | one prerendered shell; the app is a `client:only` island   |

**When each is the answer.** Nothing per-request: SSG. A content site with a
few request-time pages: Hybrid. Anything reading a request, a session or a
datastore on most routes: SSR. An application whose state lives in the browser
and whose pages are not content: CSR.

**Two honest notes.** `output: 'hybrid'` was a config value until Astro 5,
which removed it and merged its behaviour into `static` — material written
before that release names a value the current config rejects. And **CSR is a
shape, not a mode**: it is `static` output with one page, so nothing in the
config distinguishes it. What distinguishes it is the routing — the client
router owns the URL, and Astro's file router serves one shell to every path.

`prerender` is a per-route `export` and takes a literal `true` or `false`;
Astro 5 removed support for a computed value.

## Build output

**The build writes `./dist`** — Astro's `outDir` default — and a deploy pack
may rely on that path. It is the contract between the project axis and the
deploy axis: what a deploy target uploads, and where it looks for it.

A repo that changes `outDir` has changed that contract and must change its
deploy configuration in the same commit. Nothing detects the mismatch: a
deploy target pointed at a directory that no longer exists uploads nothing and
reports success.

What lands there differs by mode — files only for SSG and CSR; files plus the
adapter's server entry for Hybrid and SSR. Details, and the rule that a
post-build step writes **into** `dist/` rather than beside it, are the
`build-output` reference.

## Islands

**Zero JavaScript by default.** An `.astro` component renders to HTML at build
or request time and ships no client bundle. A UI-library component ships one
only when it carries a `client:*` directive, and each directive is a
deliberate cost paid for interactivity that HTML cannot give.

React is available in every Astro bundle this plugin offers, and is used only
where interactivity demands it. **A page with no island ships no JavaScript**
— that is the property the framework is chosen for, and a project that
hydrates every component has given it up without noticing.

## Four config facts, and the reasons for them

Measured against a proven static Astro site, not asserted:

- **`site` must be set.** Sitemap generation and canonical URLs are built from
  it. Without it a static build silently emits no sitemap and no canonical —
  no error, no warning, just an absent file nobody looks for. It is the one
  value the `## Head` section below cannot do without, and the head is the
  reason it is not optional.
- **`trailingSlash` is chosen to match the host**, not to taste. A static host
  that redirects the bare form to the slashed one (the common default) makes
  `"always"` the shape where an internal link never takes a redirect; a host
  with the opposite default inverts it. Choosing it independently of the host
  costs one redirect per navigation.
- **A Content-Security-Policy with no inline allowance forces two settings.**
  `style-src 'self'` admits no inline `<style>`, and Astro's default inlines a
  stylesheet under Vite's size limit — so `build.inlineStylesheets: "never"`.
  `script-src 'self'` needs the same on the script side, which is
  `vite.build.assetsInlineLimit: 0`. Set one and not the other and the build
  passes while the page fails in the browser.
- **A search index is built over `dist/` after the build**, as a second step
  in the build task, writing into the output directory. It is not an Astro
  integration and does not run at request time.

## Head

**One layout owns the head**, and every page passes it values — the title,
the one-sentence description, whether the page is offered to search, and
which picture a shared link shows. Those four are the per-screen contract the
blueprint pins; the layout declares them as props and decides none of them.
A second layout emitting its own tags is the failure this rule exists for: a
page whose description was written into its own markup is invisible to
everything that would check it.

Which pages get the full set is not Astro's decision. A site-shaped project
ships all of it; an application-shaped one ships the icons, the manifest and
the title unconditionally, and the search-facing half only when the project
declares that it means to be found. What the set contains — the canonical,
the icon links, the OpenGraph and twitter tags, the locale, the theme colour,
the JSON-LD with every `<` escaped — is the web-head contract, stated for
every framework rather than for this one.

Astro's own half of it:

- **The canonical is built, never typed.** `new URL(Astro.url.pathname,
  Astro.site)` is the whole of it, which is why `site` is mandatory above.
  Set `trailingSlash` before the first page ships: the canonical carries
  whichever form the config produces, and changing it later republishes every
  URL the product has.
- **The sitemap is the sitemap integration's**, and its `filter` is where a
  page pinned as not-indexed is excluded. The `<meta name="robots">` on the
  page is the other half, and neither alone is enough — a noindex page listed
  in the sitemap is an invitation followed by a refusal.
- **Static files live under `public/`**: the SVG mark, the rasterized icon
  set, the ICO, the web app manifest, `robots.txt` naming the sitemap by
  absolute URL, and the social-preview image. They carry the product's own
  name, colours and art, so the workflow writes them from the contract rather
  than a pack landing them.
- **The JSON-LD is inlined deliberately.** Serialize each block, escape every
  `<`, and emit it with `is:inline` and `set:html` — Astro must not process a
  script it did not author, and the escape is what keeps a product name
  containing an angle bracket from closing the element.
- **The icons task** is one of the pack's two landed files: it rasterizes the
  whole set from `public/brand/favicon.svg` with one-off tools and lands under
  the project's own task group. It is run by hand when the mark changes, never
  in a gate and never in the build.

Three of the task's choices are not obvious from its code:

- **It looks in two places only** — `<id>/public/brand/favicon.svg` for a
  monorepo member, `public/brand/favicon.svg` for a single-package repo.
  Searching wider would write an icon set into whichever directory happened
  to match, and a wrong favicon is found by a user rather than a build.
- **The touch icon is squared per `<rect>` tag.** iOS masks the corners and
  paints transparency black, so the tile's rounding is zeroed rather than
  matched. On an `<ellipse>`, `rx` and `ry` *are* the shape, and zeroing
  them renders a blank icon; a minified mark holds every shape on one line,
  so the edit walks the tags in node rather than using a line-scoped `sed`.
- **The ICO is built through the png-to-ico library, not its bin.** The bin
  passes only its first argument on, and the single-path branch upscales
  that one PNG into every layer; `pnpm dlx` does not help, since the dlx
  directory is off node's resolution path. The library takes an array, so
  the package is installed into the task's temp directory — never the repo
  — and driven by a one-line script. `--density` renders the SVG large
  enough that every size is a downscale.

Depth — the tag order this settles on, the escaping, and how `trailingSlash`
interacts with the canonical — is the `head` reference.

## What this pack writes

| File                                  | Is                                                                      |
| ------------------------------------- | ----------------------------------------------------------------------- |
| `.config/mise/tasks/p/_project/icons` | the icons task above, renamed into the project's task group as it lands |

## What this component does not decide

The UI kit; **how styles are authored** — that is the project's own
`stylesheet` pin, a separate axis, and two projects on Astro routinely answer
it differently; the deploy target (the deploy-axis pin decides it, and the
adapter follows from it); the value of `site`; the words in the head; the
content model's schemas. It decides how Astro is used, never whether Astro is
the answer.

Full judgment: the `astro` skill's references.

# React — conventions

React layers **on top of** the TypeScript baseline: every rule there applies
to every `.tsx` file. In this composition React is **hosted**: a host web
framework owns the build, the routes, the page markup and the **island
boundary** — whether a component hydrates, when, and what may cross into it.
The host framework defines that boundary. This component governs the React
side of it.

**React here is an island runtime, not the application.** Each island is its
own React root on a page that is otherwise static HTML. A root layout, a
client router or a site-wide provider does not apply.

## When a React component exists

**Static markup is the host's.** A React component exists because it is an
island, or because an island renders it. An island is justified by an
interaction markup cannot give — live filtering, a stateful widget, a control
whose result depends on input. For a documentation site that list is short,
and the honest default for a new page is zero islands.

## The React side of the island boundary

What may cross into an island, and how, is defined by the host framework. On
the React side:

- **Export a typed props interface** beside the component, so a type check
  of the call site holds it to the interface.
- **Pass the narrowest data the island reads** — a wide prop couples the
  island to whatever schema it was cut from.

## One root per island, and what that means for state

**React context does not cross roots.** Two islands are two trees; a
provider in one is invisible to the other. Three answers, in order:

1. **Make it one island** — one tree, state in the closest common parent.
2. **Keep the state where it already lives** — the URL or a browser API —
   and read it per island through `useSyncExternalStore`. A shared query
   parameter is a **published contract**: declared once in a `src/lib/`
   module that owns its parse and serialize, evolved additively because
   readers keep links, and validated on every read — an unknown value renders
   the default state and says so, never silently.
3. **Two separated islands that share live state are an open decision**,
   settled the day that need exists and not before — the options and when
   each fits are in the `react` skill's island-boundary reference: no store
   (URL or storage only), a cross-island store, or merging the islands so
   React context in one root carries it.

## Hydration renders the same first frame

The browser shows the build's HTML before the island's JavaScript arrives,
and React's first client render **must produce the same output**. A mismatch
is a bug: at best a slower hydration, at worst handlers attached to the
wrong elements.

**Never read in render**: `window`, `document`, `localStorage`,
`matchMedia`, `navigator`, `Date.now()`, `Math.random()`, or a
`typeof window` check. Instead:

- **A browser value that changes** — `useSyncExternalStore` with a
  `getServerSnapshot` returning the value the build rendered.
- **A value that is only different in the browser** — render the build's
  value, switch in an Effect, accept one extra render.
- **A subtree that cannot render at build at all** — `<Suspense>` around
  `use(browser())` (React 19.3), so its fallback is the build HTML.
- **IDs** for labels and ARIA relations come from `useId`.

`suppressHydrationWarning` is a one-element escape hatch; it silences and
patches nothing.

**The pre-hydration frame is real UI.** A reader can see and press an island
before it is live. Its build HTML either works without JavaScript or does
not present itself as a control that does nothing.

## Effects are for external systems only

Derive values during render; handle user actions in event handlers. An
Effect synchronizes with something outside React — a subscription, a browser
API, a fetch — and every one that fetches ignores a stale response in its
cleanup.

## Errors stay inside their island

An uncaught render error makes React remove its root's UI — here, one
island, while the static page around it survives. That is the property to
keep:

- **Wrap an island whose failure would leave a broken control** in an error
  boundary whose fallback says what is unavailable. The boundary is a class
  component (React has no hook for it), written once in the repo and reused.
- **Error boundaries do not catch event handlers or async code.** A handler
  that can fail catches its own failure, shows it in the island's state, and
  passes the original error to `console.error`; an empty `catch` is the
  failure this rule exists for.
- **The cause goes to the browser console.** No telemetry sink is declared,
  so the console is the outermost place the causal chain can land. React's
  default is to log every caught, uncaught and recoverable error there; a
  host that sets `onCaughtError`, `onUncaughtError` or `onRecoverableError`
  on the root it creates would replace that default — unverified for this
  host.
  When the product declares a sink, the boundary's `componentDidCatch` and
  each handler call the repo's one reporting function.
- **A render that throws at build is expected to fail the build, by
  default.** Error boundaries do not catch during server rendering, but
  whether the throw stops the build depends on the host's server-render
  call (unverified for this host), and a throw inside a `<Suspense>`
  subtree may render its fallback instead. Prefer the failing build where
  the host gives one.

## Styling

Styling is the `stylesheet` component's.

## Compiler keys and versions

**The `.tsx` compiler settings are this component's**, since no other
component in the composition carries them: the project's `tsconfig.json`
(not the shared base) sets `"jsx": "react-jsx"` and
`"jsxImportSource": "react"`. This component lands no config file; the two
keys are added in the same change that adds React.

The current line at generation time was React 19 (19.3); the lockfile
decides the exact version. Dependencies: `react` and `react-dom`;
`@types/react` and `@types/react-dom` as dev dependencies.

## The Rules of Hooks gate

React's documented enforcement of the Rules of Hooks and Effect dependencies
is `eslint-plugin-react-hooks` with its flat-config `recommended` preset.
**Today that check lands nowhere**: the `toolchain-gate/eslint` component
runs the zero-config house linter, whose bundled plugins do not include it,
and this component lands no lint config. A conditionally called hook
type-checks, builds and passes every gate. **This is an open decision** —
the options and their consequences are in the `react` skill's framework
doctrine reference: add the plugin to the house linter upstream, add a
project-local config with it, or rely on review.

## Testing

Islands are tested in **jsdom with Testing Library**, opted in per file so
the node environment stays the default for `src/lib/`. This **overrides**
the TypeScript baseline's Web UI testing row (project-wide jsdom plus
`globals: true`) in favour of this composition's bundle ruling, so Testing
Library's cleanup is registered explicitly in the test setup file. Depth:
the `react` skill's testing reference.

## What this component does not decide

Whether a page has an island, its hydration directive, what may cross into
it, and how React is registered with the build — the host framework's.
Styling — the `stylesheet` component's. Lint configuration — the
`toolchain-gate/eslint` component's. The rest of `tsconfig` — only the two
JSX keys are this component's. The test runner's config and coverage
threshold — the TypeScript baseline's.

Full judgment: the `react` skill's references.
