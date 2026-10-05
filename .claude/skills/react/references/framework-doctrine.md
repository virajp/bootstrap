# React — framework doctrine

The one artifact the `language-bundle` bar owes for this framework component:
what React is for in this composition, when a React component should exist,
how one is composed, and the seams with the components beside it.

**Every opinion below traces to a source**, in the bar's precedence order:
(1) the repo's settled pattern — there is none yet, so nothing here is
overridden by detection; (2) React's documented recommendation — Context7
`/reactjs/react.dev` and `/react/react`, read 2026-10-05; (3) a catalog
entry, named where it is applied.

**This is React doctrine, not a pairing.** The host framework owns the
build, the routes, the page markup and the island boundary, and defines what
crosses it. Neighbours are named here only by role.

## What React is for here

**Interactivity the host's markup cannot give, one island at a time.** Each
hydrated component is an **island** — its own React root, hydrated
independently of every other island on the page. React is not the site's
component model: pages, layouts, content rendering and the head are the
host's.

## When a React component should exist

**Only as an island, or as a component an island renders.** A second
component model for static markup costs the reader: every file in
`src/components/` now has to be known as one model or the other. **KISS**
(simplicity measured at the reader) settles it — one model for static
markup, the host's. The catalog's own limit applies: static markup *inside*
an island's tree is React, because that is the island's idiom.

**Is it an island at all?** A documentation site's honest island list is
short: search-as-you-type over a prebuilt index, a filterable reference
table, a switcher that rewrites several code blocks, a playground. **YAGNI**
governs the default — a page starts with none, and an island is added for a
stated interaction, not to keep options open.

## Effects are an escape hatch

React's own rule: an Effect synchronizes a component with an **external
system** — a subscription, a browser API, a network request. Two things are
never Effects:

- **Transforming data for rendering.** Compute it in the render body. A
  `useState` + `useEffect` pair that copies one value into another is a
  render behind, and a state that can disagree with its source.
- **Responding to a user action.** That is the event handler's job.

An Effect that fetches ignores a stale response in its cleanup (an `ignore`
flag set in the returned function), or a fast second request is overwritten
by a slow first one. On a static site most data is a build input passed as a
prop, so a fetching Effect is itself a smell.

## Composition inside an island

- **Lift state to the closest common parent** in the island; pass it down,
  pass callbacks up.
- **Start with one modest component; split when a second concern starts
  changing.** **Single responsibility** names the eventual split — a
  component that filters *and* fetches *and* formats has three reasons to
  change — but its own limit applies: don't split below the concept, and a
  small island that is one coherent thing stays one file until its history
  shows two concerns moving independently.
- **Keys are stable identities** from the data (a slug, a path) — never an
  index for a list that can reorder or filter.
- **Keep components pure.** Same props, same output; side effects in
  handlers and Effects only. Purity is also what makes
  [`hydration.md`](hydration.md) achievable.

## Memoization and the React Compiler

React's compiler automates memoization; whether and how it is enabled is a
build setting, not React code. For a handful
of small islands there is no measured re-render cost to remove, so neither
hand-written `useMemo` / `useCallback` / `memo` nor the compiler is added in
advance. **KISS**, within its own limit: a performance measure may justify
the extra machinery, with the measurement beside it.

## Styling

Styling is the `stylesheet` component's.

## The `.tsx` compiler keys

**These are this component's.** No other component in the composition
carries them — the `tsconfig` component's doctrine covers the base, the `@/`
alias and the build config, and says nothing about JSX:

```jsonc
{
  "compilerOptions": {
    "jsx": "react-jsx",
    "jsxImportSource": "react"
  }
}
```

They go in the project's `tsconfig.json`, never the shared base, because they
are a property of this project's UI library and not of every TypeScript
package a workspace might hold. Without them `.tsx` is not typed against
React's automatic JSX runtime. This component lands no config file: the two
keys are added in the same change that adds React.

**Versions.** The current line at generation time was React 19.3; the
lockfile decides the exact version.

## The Rules of Hooks gate

React's documented enforcement of the Rules of Hooks and Effect dependencies
is **`eslint-plugin-react-hooks`, with its flat-config `recommended`
preset**, which also surfaces the React Compiler's diagnostics whether or not
the compiler is on.

**Today that check lands nowhere.** The `toolchain-gate/eslint` component
runs the house linter, `@askviraj/linter`: zero-config, bundling ESLint and
its plugins for TypeScript, JSON, CSS, HTML, Markdown, YAML, TOML and the
host framework's file type — no React plugin. This component lands no lint config. So a hook called
inside a condition type-checks, builds and passes every gate, and fails at
runtime when the condition flips.

**This is an open decision**, settled by the person who owns the lint gate:

| Option | What it takes | Consequence |
| --- | --- | --- |
| **(a) Add the plugin to the house linter upstream** | a release of `@askviraj/linter` bundling `eslint-plugin-react-hooks`, applied to `.tsx`/`.jsx` | every repo on the house linter gets the check with no local config, and the one-place-the-linter-is-configured rule holds; it waits on an upstream release, and repos without React carry an inert plugin |
| **(b) A project-local config adding the plugin** | either a `configs` entry in `.config/linter.yaml` — documented for rule overrides scoped by `files`, **unverified** for loading a plugin the linter does not bundle — or a separate flat `eslint.config.*` run beside the house linter | the check lands in this repo now; the second form is a second lint configuration and a second run, which the eslint component's one-place rule argues against |
| **(c) No lint gate; rely on review** | reviewers checking hooks against React's rules | nothing to install or maintain; a violation is caught only if a reviewer sees it, and the Effect-dependency check is effectively gone |

(a) is the option that keeps the gate's own structure intact; (b) is the
fastest; (c) is only honest if it is written down as the choice.

## The seam with the components beside it

- **TypeScript baseline** — every rule applies to `.tsx`. The island's
  props interface is exported beside the component.
- **The host framework** — owns whether an island exists on a page, its
  hydration directive, what crosses the boundary, the build, and React's
  registration with it.
- **`stylesheet`** — owns styling.
- **`toolchain-gate/eslint`** — owns the lint gate; whether it carries the
  hooks rules is the open decision above.
- **`toolchain-gate/tsconfig`** — owns the compiler config files; this
  component owns only the two JSX keys.

Neighbours are named by role. Cross-framework integration judgment is not
this file's.
