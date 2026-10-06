---
type: vwf-flow
title: Home
description: An outside developer opens the site, sees what bootstrap does, copies the install command, and moves on to the docs or the source.
status: reviewed
implementation: none
owner: [site]
---

# Flow: Home

## Purpose

A visitor who opens the home address learns in one view what bootstrap does and
can start at once: copy the install command, read the docs, or open the source.
The flow exists so an outside developer can decide to adopt the product without
asking anyone.

Serves: [Outside adoption](../../../product.md#goal-outside-adoption)

## Platforms

| Platform | File | Notes |
| -------- | ---- | ----- |
| site | [site](./site.md) | The only surface; a single static page plus its not-found page. |

## Trigger & Actors

| Actor | May trigger | Authorization | Audit-recorded |
| ----- | ----------- | ------------- | -------------- |
| Outside developer | Opens `https://bootstrap.virajp.dev/` | none — public page | no |
| Repository owner | Opens `https://bootstrap.virajp.dev/` | none — public page | no |
| Search or link-preview robot | Opens `https://bootstrap.virajp.dev/`; reads the page metadata only | none — public page | no |

## Steps

1. Site serves the home page — N/A: the page reads no product data
2. Visitor copies the install command with the copy button; the button shows
   "Copied" for 1.5 seconds ([Code block](../../../design-system.md))
3. Visitor follows "Read the docs" to the
   [documentation flow](../110-documentation/index.md), or "Source ↗" to the
   source repository (external)

## Guarantees

| Step / group | Consistency | On failure | Idempotency | Load & latency |
| ------------ | ----------- | ---------- | ----------- | -------------- |
| all | atomic — the page is static and holds no state | Clipboard unavailable: no confirmation; the command stays visible and selectable as text. Unknown address: the `not-found` screen. | n/a — every step is safe to repeat | default — per [reliability](../../../conventions.md#reliability) |

## Diagram

```mermaid
sequenceDiagram
    actor V as Visitor
    participant S as Site
    participant C as Clipboard
    participant D as Documentation flow
    participant R as Source repository
    V->>S: open home address
    S-->>V: home page
    V->>C: copy install command
    alt clipboard available
        C-->>V: "Copied" for 1.5 s
    else clipboard unavailable
        C-->>V: no confirmation; command stays selectable
    end
    alt read the docs
        V->>D: Read the docs
    else view the source
        V->>R: Source ↗
    end
    alt unknown address
        V->>S: open unknown path
        S-->>V: not-found screen
    end
```

## Background Jobs

N/A — static pages, no processing.

## Acceptance

- Given the home address, when a visitor opens it, then the home page renders
  every block of screen `100a` in order.
- Given the home page, when the visitor activates the copy button, then the
  clipboard holds exactly `pnpx @virajp.dev/bootstrap@latest init` and the
  button shows "Copied" for 1.5 seconds.
- Given the clipboard is unavailable, when the visitor activates the copy
  button, then no confirmation shows and the command stays visible and
  selectable as text.
- Given an unknown path, when a visitor opens it, then the `not-found` screen
  (`100b`) shows.
- Given the home page, when a robot reads it, then its metadata matches the
  Metadata block of `100a`.
- Given the `not-found` page, when a robot reads it, then the page is not
  indexable.
- Given the keyboard only, when a visitor uses the home page, then the copy
  button, the buttons and the links are reachable and show a visible focus,
  to WCAG 2.2 AA per the design-system
  [Accessibility Standard](../../../design-system.md#accessibility-standard).
- Given client scripting is off, when a visitor opens the home page, then the
  content and the links work, and the copy button and the search input are
  hidden, never inert ([reliability](../../../conventions.md#reliability)).
- Abuse case: n/a — the page is public, read-only and accepts no input that
  changes state.

## References

- [design-system](../../../design-system.md) — Code block, Empty and error,
  Brand assets, Accessibility Standard
- [web metadata](../../../conventions.md#web-metadata),
  [theming](../../../conventions.md#theming) (dark only),
  [reliability](../../../conventions.md#reliability),
  [observability](../../../conventions.md#observability) (no analytics)
- API surface: N/A — static site, no service project
- Entities: N/A — the page reads no product data
