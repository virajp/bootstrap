# Testing an island

The React layer on top of two references that already own most of testing:
the language pack's Vitest reference owns the runner — config, `_testUtils/`,
coverage, how suites run — and the host framework's component owns what is
and is not unit-tested on its own files. This file owns what an island
adds.

Sources, read 2026-10-05: Context7 `/testing-library/testing-library-docs`
— React Testing Library's setup page (cleanup and Vitest globals), its API
page (`render` options `container`, `hydrate`, `onRecoverableError`,
`onCaughtError`), the `ByRole` query page, and the `user-event` intro
(`userEvent.setup()`); `/vitest-dev/vitest` — the environment guide
(`@vitest-environment`); `/reactjs/react.dev` — `renderToString` and
`hydrateRoot`. The environment override cites the TypeScript pack's Vitest
reference and this composition's bundle, which are rulings, not API
sources.

## Environment: jsdom per file, node by default — an override

**This component overrides the language pack's Web UI row.** That row, in
the Vitest reference's *Variations by what the code needs* table, adds
`environment: "jsdom"` and `globals: true` project-wide for component tests.
On this site that row is the wrong default, and the composition says so:
this composition's bundle ruling for testing is **node for `src/lib/` and anything
the build calls, jsdom only for the islands a repo actually writes**. Where
the two disagree, the bundle's ruling is the more specific one and wins; the
rest of the Web UI row — the coverage include — stands.

So the vitest config keeps the node default and sets **neither**
`environment` nor `globals`, and each island's test file opts in with
Vitest's control comment as its first line:

```tsx
// @vitest-environment jsdom
```

That keeps the split visible per file, with no second config.

## Cleanup

`globals: true` came bundled with the Web UI row's jsdom default, and this
component drops both: with jsdom per file there is no reason to inject test
globals into every node suite. That has one consequence, documented by
Testing Library: its automatic unmount-after-each-test registers itself only
when the runner's globals are on. With them off, cleanup is explicit, in
`src/_testUtils/setup.ts` — the language pack's `setupFiles` entry:

```ts
import { cleanup } from "@testing-library/react";
import { afterEach } from "vitest";

afterEach(() => {
  cleanup();
});
```

Without it, every render accumulates in one document, and a query in the
third test finds the second test's button. The hook also runs in node
suites. Testing Library documents the hook, not its behaviour where no
`document` exists, so that it is harmless there is our reasoning rather than
a documented guarantee — if a node suite fails on the import, move the hook
into a setup file only the island tests load. A repo that turns `globals: true`
on anyway gets Testing Library's automatic cleanup and should drop this hook
rather than run both.

## Writing the test

**Assert what a reader perceives, drive it the way a reader would.**

- **Query by role and accessible name first** (the `ByRole` query page) — `getByRole("button",
  { name: "Copy" })`, `getByRole("searchbox")`, `getByRole("heading",
  { level: 2 })`. A role query that cannot find the control is usually an
  accessibility defect, not a test problem: the test is checking the same
  tree a screen reader reads. `getByLabelText` for form fields with no
  implicit role; `getByTestId` last, and only for something with no
  accessible identity.
- **Interact through `user-event`**, created once per test with
  `userEvent.setup()` (the `user-event` intro) and awaited: it fires the full sequence a real
  interaction does (focus, key events, input), where a bare `fireEvent`
  fires one event.
- **Props are fixtures of the boundary contract.** Build them the way the
  host page does — the narrow, typed shape from
  [`island-boundary.md`](island-boundary.md) — so the test exercises the
  shape the real page passes.
- **The failure paths are tests too**: the fallback a boundary renders, and
  the inline message a failing handler shows ([`errors.md`](errors.md)).
  **Explicit error semantics** — assert the reader-facing message, not merely
  that something threw.

```tsx
// @vitest-environment jsdom
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect, test } from "vitest";
import ApiFilter from "@/components/ApiFilter";

const entries = [
  { slug: "init", title: "init", group: "setup" },
  { slug: "deploy", title: "deploy", group: "release" },
];

test("filtering by group hides entries outside it", async () => {
  const user = userEvent.setup();
  render(<ApiFilter entries={entries} />);

  await user.selectOptions(screen.getByRole("combobox", { name: "Group" }), "release");

  expect(screen.getByRole("link", { name: "deploy" })).toBeTruthy();
  expect(screen.queryByRole("link", { name: "init" })).toBeNull();
});
```

## The hydration test

For any island that reads the browser — an external store, a two-pass
render, `use(browser())` — one test proves the first client frame matches
the build: render to a string, hydrate it, and assert React reported no
recoverable error. Testing Library's `render` takes `container`,
`hydrate: true` and an `onRecoverableError` callback (its API page; the
callback behaves as `onRecoverableError` on React's `createRoot`), and
`renderToString` produces the build-side HTML:

```tsx
// @vitest-environment jsdom
import { render } from "@testing-library/react";
import { renderToString } from "react-dom/server";
import { expect, test, vi } from "vitest";
import VersionSwitcher from "@/components/VersionSwitcher";

test("hydrates without a mismatch", () => {
  const props = { versions: ["1.x", "2.x"], current: "2.x" };
  const container = document.createElement("div");
  container.innerHTML = renderToString(<VersionSwitcher {...props} />);
  document.body.appendChild(container);

  const onRecoverableError = vi.fn();
  render(<VersionSwitcher {...props} />, { container, hydrate: true, onRecoverableError });

  expect(onRecoverableError).not.toHaveBeenCalled();
});
```

**What it catches, and a limit we reason about.** It catches nondeterminism
(a date, a random id, a counter-based id), a `getServerSnapshot` that
disagrees with the build, and a two-pass render done in render instead of in
an Effect. By our reasoning — not a documented statement — it is unlikely to
catch a `typeof window` branch: the test runs under jsdom, which defines a
`window`, so both renders probably take the same branch. Treat that case as
one for review against [`hydration.md`](hydration.md) rather than for this
test.

## Coverage

The baseline's include for a web UI already names `src/components/**`; the
islands live there and count, and this component adds no exclusion.
Exclusions for the host's own file types are the host framework's
component's; the threshold is the repo's.

## What is not tested here

Whether an island hydrates at the right moment, and how it looks on the
built page — those are build and end-to-end concerns, the host framework's
component's. Visual regressions are the stylesheet component's gate, if it
has one.
