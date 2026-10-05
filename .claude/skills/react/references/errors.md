# Errors in an island

How a failure inside an island travels, where it is decided, and what the
reader sees. The baseline's error-semantics reference owns how a failure is
typed and carried in TypeScript; this file owns what React adds.

Sources: Context7 `/reactjs/react.dev` (`Component` — error boundaries;
`createRoot` and `hydrateRoot` — root error options and default console
logging), read 2026-10-05. Catalog: **explicit error semantics**.

## The blast radius is one island

By default, a render error that nothing catches makes React **remove its
root's UI from the screen**. On a single-page app that is the whole
application; here the root is one island, so the static page around it — the
documentation the reader came for — survives. That containment is free, and
it is the property to protect: nothing in this file widens it.

## Where an error is decided

| Failure | Caught by | Decided where |
| --- | --- | --- |
| Throws during render or in a hook, in the browser | the nearest error boundary | the boundary's fallback |
| Throws during render **at build** | nothing — boundaries do not catch server rendering | by default, expected to fail the build; depends on the host's server-render call; inside a `<Suspense>` subtree it may render the fallback instead |
| Throws in an event handler | nothing React provides | the handler itself |
| Rejects in async code (`fetch`, timers) | nothing React provides — except inside `startTransition` | the code that awaited it |
| Throws inside the boundary itself | the next boundary up, or none | — |

Read the second row as a hedge, not a guarantee. Where the host's
server-render call does fail the build — unverified for this host — a broken
island is a **red build** rather than a blank region on a published page,
and that is worth keeping: don't wrap build-time rendering in a `try` just
to keep the build green. Where the throw sits inside a `<Suspense>` subtree,
the build may instead ship that boundary's fallback, so the fallback has to
be acceptable as published content.

## Error boundaries

A boundary is a **class component** — `static getDerivedStateFromError`
switches to the fallback; `componentDidCatch`, when present, reports. React
still has no hook equivalent. Write **one** in the repo, taking the fallback
as a prop, and reuse it. With no telemetry sink declared it needs no
`componentDidCatch` at all — React's default logs what it catches (see
[Reporting](#reporting)):

```tsx
// src/components/IslandBoundary.tsx
import { Component, type ReactNode } from "react";

interface Props { fallback: ReactNode; children: ReactNode }

export class IslandBoundary extends Component<Props, { failed: boolean }> {
  state = { failed: false };

  static getDerivedStateFromError() {
    return { failed: true };
  }

  render() {
    return this.state.failed ? this.props.fallback : this.props.children;
  }
}
```

**Granularity follows the reader**, per React's own guidance: a boundary
goes where an error message makes sense, not around every component. On this
site that is usually **the island's top**, with a fallback that says what is
unavailable and what still works ("Search is unavailable — browse the
reference index"). An island small enough that losing it loses nothing (a
cosmetic enhancement) needs none: the default removal is already the right
outcome. **Explicit error semantics**: the fallback is the outermost edge,
so it speaks the reader's vocabulary — an actionable sentence, not a stack.

## Handlers and async code

Boundaries see none of it, so the handler owns its failure:

- **Catch at the altitude that can decide** — usually the handler — and put
  the outcome into state the island renders: an inline message, a retry
  control. A failure that is only logged is a control that silently did
  nothing.
- **The causal chain still goes somewhere.** React never sees a handler's
  failure, so its default logging does not either: the `catch` that renders
  the reader's message also passes the original error to `console.error`.
  The reader gets the sentence; the console gets the cause.
- **No empty `catch`, no `catch → default`.** A search that fails and shows
  "no results" has told the reader something false. Show the failure as a
  failure.
- **Async work in a transition** (`startTransition` from `useTransition`) is
  the one case a boundary does catch; use it where the boundary's fallback
  is genuinely the right answer for the async failure, not as a way to avoid
  writing the handler's branch.

## Reporting

**Where the causal chain goes.** Explicit error semantics asks for two
destinations at the outermost edge: an actionable message for the reader,
and the full cause somewhere a maintainer can read it. The reader's half is
the fallback and the handler's inline message above. The maintainer's half,
on this site, is **the browser console** — because the detected stack
declares **no telemetry sink**, and the console is the outermost place the
cause can land without one. That is a stated limit of a static docs site:
a failure is observable to whoever opens the console on that page, not
collected anywhere.

**React writes there by default.** React's default is to log every error to
the console — caught by a boundary, uncaught, and recoverable (which
includes a hydration mismatch). The root-level options `onCaughtError`,
`onUncaughtError` and `onRecoverableError` on `createRoot` / `hydrateRoot`
replace that default, and they are options of the call that **creates the
root**, which here is the host framework's. A host that sets them would
replace the console default; whether this host does is unverified. This
component sets none of them: with no sink to send to, **React's defaults are
the intended behaviour**, and the points below assume they are what runs:

- **Render failures** reach the console through React's default, with the
  component stack. The boundary adds no `componentDidCatch` just to log
  again.
- **Handler and async failures** reach it only through the handler's own
  `console.error`, per the rule above — React never sees them.
- **Hydration mismatches** are logged by the same default in the browser,
  and caught earlier by the hydration test ([`testing.md`](testing.md)),
  before they ship.

**When the product declares a sink**, the destination changes and the shape
does not: the boundary gains a `componentDidCatch` that calls the repo's one
reporting function, named by role and never a vendor SDK directly, and
every handler's `catch` calls the same function instead of `console.error`.
The boundary and the handlers are the reporting points this component
owns.
