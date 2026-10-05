---
name: react
version: 0.1.0
category: development
description: >-
  React as an island runtime inside a host web framework — when a React
  component should exist at all, the React side of the island boundary
  (typed narrow props, no context across roots, shared state through a
  published contract), hydration that renders the same first frame as the
  build, Effects only for external systems, errors contained per root, and
  testing an island. Layers on the TypeScript baseline; the host framework
  defines the island boundary itself.
  Auto-applies when editing a .tsx or .jsx file.
license: MIT
user-invocable: false
allowed-tools: Read Grep Glob Edit Write Bash
paths:
  - "**/*.tsx"
  - "**/*.jsx"
---

# React

Layers on the TypeScript baseline — read that skill's standards first; this
adds to them and replaces none of them. **The host framework defines the
island boundary — whether a component hydrates, when, and what crosses into
it. This skill governs the React side of the island.**

**Each island is its own React root on a static page.** No root layout, no
client router, no site-wide provider.

| Doing | Read |
| --- | --- |
| Deciding whether a React component should exist; components, Effects, compiler keys, the lint gate | [Framework doctrine](references/framework-doctrine.md) |
| Props arriving from the host; state shared between islands | [The island boundary](references/island-boundary.md) |
| Anything reading the browser — `window`, storage, media queries, dates, ids | [Hydration](references/hydration.md) |
| A control that can fail — error boundaries, handlers, reporting | [Errors](references/errors.md) |
| Writing or wiring an island's tests | [Testing](references/testing.md) |
