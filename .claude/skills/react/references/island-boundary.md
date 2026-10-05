# The island boundary — the React side

Everything an island knows arrives across one boundary, from the host's
build render to the React root that hydrates it in the browser. **The host
framework defines that boundary** — what may cross it and how. This file is
what React code does on its side of it, and what to do about state two
islands both need.

Sources: Context7 `/reactjs/react.dev` (`useContext` — the closest
provider above in the tree; `useSyncExternalStore`; `useId`), read
2026-10-05. Catalog entries are named where applied.

## Props

What may cross into an island, and how, is the host framework's definition.
On the React side, **design by contract** in its boundary form: state what
crosses, and let the type say it.

- **Export a typed props interface beside the component**, so a type check
  of the call site can hold it to the interface.

```tsx
// src/components/ApiFilter.tsx
export interface ApiFilterProps {
  entries: ReadonlyArray<{ slug: string; title: string; group: string }>;
  initialGroup?: string;
}

export default function ApiFilter({ entries, initialGroup }: ApiFilterProps) {
  // …
}
```

**Pass the narrowest shape the island reads** — **interface segregation**
applied to props. The page maps its content down to the fields the island
uses; it does not hand over whole records, because a wide prop couples the
island to the content schema. The catalog's limit holds: a small props object
whose fields are one cohesive record is not shattered into a dozen scalars.

## Context does not cross islands

Each island is a separate React root, and context is scoped to the tree
below its provider — so a provider in one island is invisible to every
other. The answers, in order of preference:

1. **One island.** Two controls that share state and sit near each other are
   one component tree. Put the shared state in their closest common parent.
   This is the answer most of the time, and it is the KISS one.
2. **State that already has a home.** A filter whose state belongs in the
   URL (shareable, survives reload) writes a query parameter and the list
   reads it; a preference belongs in storage. Each island reads the home
   directly — through `useSyncExternalStore`, per
   [`hydration.md`](hydration.md) — and neither knows the other exists.
   The home is then a contract, and is treated as one — see
   [The URL as a contract](#the-url-as-a-contract) below.
3. **Two separated islands sharing *live* state with no natural home — an
   open decision.** Settle it the day the need exists, not before
   (**YAGNI**: a store installed in advance is an extensibility seam with no
   registrant). The real options:

   | Option | When it fits | Cost |
   | --- | --- | --- |
   | **No store — URL or storage only** (option 2, stretched) | the shared state is something a reader would bookmark or keep across visits: a filter, a selected platform, a preference | every change goes through the URL or storage and its parser; awkward for high-frequency or transient state (a hover, an open panel) |
   | **A cross-island store** — a small external store module both islands import, read through `useSyncExternalStore` or the store's own React binding | transient state two distant islands must see live, with no reason to appear in the URL | a dependency (or a hand-written store with its own subscription contract), vetted when chosen; one more module both islands couple to |
   | **Merge the islands; React context within one root** | the two controls can be one tree — one island wrapping both, with the shared state in a context or the common parent | the merged island hydrates as one unit, so markup between the two controls becomes part of it and ships with it |

   With no detected pattern there is no consensus to report: the first fits
   most of what a documentation site shares, the third is option 1 reached
   from the other side, and the second is for what neither covers.

Not on the list: a module-level variable mutated by two islands (no
subscription, so the second island never re-renders), and a hand-rolled
`CustomEvent` bus (a store without the store's guarantees).

## The URL as a contract

A query parameter two islands share is not private state. Readers bookmark
it, paste it into issues and link to it from other sites, so the moment one
ships it is **published**, and it outlives the islands that read it.
**Information hiding** names this exactly in its own limit: some information
*is* the contract, and hiding it behind a leaky wrapper only obscures the
agreement — publish it explicitly and version it instead. The islands still
hide everything else from each other; the parameter is the one thing they
deliberately share.

**One declaration.** The parameter's name, its allowed values and its
default live in **one module in `src/lib/`** — say
`src/lib/reference-filter-param.ts` — which exports the type, a `parse`
function and a `serialize` function. Every island that reads or writes the
parameter imports that module; none touches `URLSearchParams` for that key
directly. **Single responsibility**, within its own limit: parsing and
serializing are one decision (the parameter's shape), so they stay in one
module rather than being split below the concept. Because the module is in
`src/lib/`, it is node-tested like the rest of the site's logic.

**Versioned by evolving additively.** A new value is added; an existing value
keeps parsing for as long as old links exist — which on a public
documentation site is indefinitely. Renaming a value or the key is a breaking
change, and the module carries the old spelling as an alias that parses to
the new one, rather than breaking every link that used it.

**A trust boundary — validated on read, every time.** The query string is
input anyone can type. **Design by contract** draws the line in its own
limit: at a trust boundary "caller's fault" never means skip the check, so
`parse` validates against the closed set of values and returns a typed result
— a valid value, or absent, or invalid with the raw string as its reason —
never a raw string the islands then trust.

Behaviour for each case is defined once, beside the parser:

- **Absent** — the default state. This is also what the build rendered,
  since a static page has no query string, so it is the
  `getServerSnapshot` value and the first frame matches.
- **Unknown or malformed** — the island renders the default state **and
  says so** ("No group called *xyz* — showing all entries"). It never
  silently coerces to the default, which would tell the reader their link
  worked when it did not, and it never writes the bad value back to the URL.
- **The raw value is display text only.** It is rendered as a React text
  child, which escapes it; it never reaches `dangerouslySetInnerHTML`, an
  `href`, or a selector.

## Many instances of one island

The same component used several times on a page hydrates as several roots
with independent state — two code-sample switchers do not stay in sync
unless their state has a shared home (option 2 above). IDs inside the island
come from `useId`, so two instances never collide on a `for`/`id` pair.
