# Hydration — the same first frame

Every island on this site is rendered twice: once to HTML at build, and once
in the browser when the host hydrates it. The browser's first render
**hydrates** the build's HTML — React attaches to the existing markup rather
than replacing it — and that only works if both renders produce the same
output.

Sources: Context7 `/reactjs/react.dev` (`hydrateRoot`, `useSyncExternalStore`,
`use`, `browser`, `useId`, `useEffect`) and `/react/react` (the 19.3
changelog), read 2026-10-05.

## Why a mismatch is a bug

The reader looks at the build's HTML before the island's JavaScript loads —
with a lazy hydration trigger, possibly for the whole visit. React's docs are blunt:
React recovers from some mismatches, but they are to be fixed like any other
bug. At best hydration is slower; at worst event handlers attach to the
wrong elements. And a build-time snapshot is a *fixed* snapshot: whatever the
build machine saw — its clock, its locale, the absence of `window` — is what
every reader's first frame must also produce.

## What may not be read during render

The documented causes, plus the ones a static build makes inevitable:

- `typeof window !== "undefined"` branches — the build takes one branch, the
  browser the other, by construction.
- Browser-only APIs: `window`, `document`, `localStorage`, `matchMedia`,
  `navigator`.
- Values that differ per render: `Date.now()`, `new Date()` formatted for
  display, `Math.random()`, `crypto.randomUUID()`.
- Locale-dependent formatting with no fixed locale — the build machine's
  locale is not the reader's. Pass the locale explicitly, or format at build
  and pass the string.

## The four tools, by situation

**A browser value that can change — use an external store.**
`useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot)` is React's
API for exactly this. `getServerSnapshot` runs at build *and* during
hydration, so returning the build's value there is what makes the first
frame match; React then re-renders with `getSnapshot`'s real value.

```tsx
import { useSyncExternalStore } from "react";

const query = "(prefers-reduced-motion: reduce)";

function subscribe(onChange: () => void) {
  const media = window.matchMedia(query);
  media.addEventListener("change", onChange);
  return () => media.removeEventListener("change", onChange);
}

export function usePrefersReducedMotion(): boolean {
  return useSyncExternalStore(
    subscribe,
    () => window.matchMedia(query).matches,
    () => false, // the build's answer, and hydration's
  );
}
```

`getSnapshot` must return the same value while the store is unchanged — a
fresh object each call re-renders forever. Return a primitive, or cache.

**A value that is simply different in the browser — two passes.** Render the
build's value, then switch in an Effect. React documents the cost: the
component renders twice, and a visible change right after hydration can feel
jarring on a slow connection. Use it for small things (a relative date), not
for whole subtrees.

**A subtree that cannot render at build at all — `use(browser())`.** React
19.3 added `browser()` to `react-dom`: `use(browser())` inside a
`<Suspense>` boundary suspends at build, so the boundary's **fallback** is
what the HTML carries, and renders normally in the browser — without
reporting a mismatch. Two hard rules from the docs: it must sit inside a
`<Suspense>` boundary, or the build render fails; and the fallback is the
reader's first frame, so it is real content (a skeleton the size of the
result, or a sentence saying what loads), never an empty box.

**An attribute that cannot match — `suppressHydrationWarning`.** One
element, one level deep, for a value like a rendered timestamp. It silences
the warning and **does not patch the text** — the build's value stays until
the next render. An escape hatch, not a pattern; more than one in an island
means the island wants one of the three tools above.

## IDs come from `useId`

Labels, `aria-describedby`, `aria-controls` — every id that ties two
elements together comes from `useId`, which produces the same id at build and
at hydration and a distinct one per instance. A module-level counter or a
random id does neither. (React 19.2 changed the generated format to use
underscores rather than colons; nothing should parse it.)

## The pre-hydration frame is UI

Between the HTML arriving and the host hydrating the island, the island is
visible and inert. With a lazy hydration trigger that window can be the whole
time a reader spends above it. Design for it — this is the accessibility guardrail the
catalog's YAGNI and KISS entries both refuse to trade away:

- Prefer markup that **works without JavaScript** and is enhanced by it — a
  real link, a `<form>` that submits to a real URL, a `<details>` the island
  takes over.
- Where that is impossible, **do not present a control that does nothing.**
  Render it disabled, or render the static result instead of the control,
  until the island is live (a two-pass `isHydrated` flag is the honest use of
  that pattern).

## Testing it

An island that reads anything from the browser gets a hydration test: render
it to a string, put that in a jsdom container, hydrate, and assert React
reported no recoverable error. The shape is in [`testing.md`](testing.md).
