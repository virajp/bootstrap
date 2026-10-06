---
type: vwf-flow-platform
title: Documentation — site
description: The documentation page of the public site, one screen used by every docs page.
status: reviewed
platform: site
implementation: none
owner: [site]
---

# Documentation — site

Flow contract: [Documentation](./index.md)

## Screens → site

| Code | Screen | Route | Reads (operationId) | States (loading/error/empty) | Actions | Form validation |
| ---- | ------ | ----- | ------------------- | ---------------------------- | ------- | --------------- |
| 110a | doc-page | every address of the Page set in [index](./index.md#page-set), under `/docs/` | n/a — static page | default; mobile menu open and closed; scripting off (copy buttons and Search input hidden); narrow view with scripting off (mobile-menu button hidden, sidebar navigation shown inline above the page content); copy confirmed ("Copied" for 1.5 s); clipboard unavailable (no confirmation). The search panel's states ("Searching…" and no results) are per the design-system [Component Behaviors](../../../design-system.md#component-behaviors) "Inputs (search)", "Loading" and "Empty and error". Loading: n/a — static. Empty: n/a — every page has content. Error: n/a — static; an unknown address shows `100b` not-found. | Sidebar link; "On this page" link; previous and next links; copy command; Search; open and close the mobile menu | n/a — the search input is not a form that validates |

### `110a` — `doc-page` components

| Component | Rules |
| --------- | ----- |
| Header | The site header of `100a` in [Home](../100-home/site.md) (brand mark, live wordmark "bootstrap", "Docs" link, Search input, "Source ↗" external link), plus, on a narrow view only, the button that opens the mobile menu. Always visible. With scripting off, that button is hidden, never inert ([reliability](../../../conventions.md#reliability)). |
| Search input | Behaves per the design-system "Inputs (search)" and searches every docs page. Hidden when scripting is off ([reliability](../../../conventions.md#reliability)). |
| Sidebar navigation | Lists the pages of the Page set in [index](./index.md#page-set), in that order, under the section eyebrows START, COMMANDS, GUIDES and REFERENCE (design-system "Navigation"). The current page is marked. Each entry opens its page. On a narrow view it is the mobile menu: focus moves into it on open, it closes on Escape, on the menu button (a toggle), on a tap outside the menu, or when a page link is selected; on close, focus returns to its button. Works with scripting off; on a narrow view with scripting off, it is shown inline above the page content. |
| Page content | Heading, prose, code blocks, terminal samples, tables and callouts of the page. Always visible. Each section has a heading; heading levels do not skip. |
| Code block | Per the design-system "Code block"; the text stays selectable. Holds a ghost copy button. |
| Copy button (ghost) | Activating it places exactly the text of its code block on the clipboard and shows "Copied" for 1.5 seconds. If the clipboard is unavailable, no confirmation shows and the command stays selectable. Hidden when scripting is off, never inert. |
| Terminal sample | Per the design-system "Terminal sample". Static; not interactive. |
| Table | Per the design-system "Table". Used for flags, exit codes and report fields. |
| Callout | Per the design-system "Callout", with the kind `note`, `tip` or `caution`. Shown only where a page has one. |
| "On this page" list | Lists the sections of the page; each link moves to its section; the section in view is emphasised. Hidden on a page with no sections. Works with scripting off, without the in-view emphasis. |
| Previous and next links | Placed at the end of every page, after the content, and follow the sidebar order. Each shows the neighbour page title. The first page has no previous link. The last page has no next link. |
| Footer | Shows "MIT License" and a "Source ↗" external link, as on `100a`. |

### `110a` — `doc-page` metadata

| Field       | Value |
| ----------- | ----- |
| title       | `<page title> \| <site name>` (Site name per [web metadata](../../../conventions.md#web-metadata)), where the page title is the Page column of the [Page set](./index.md#page-set) (for example "bootstrap check \| bootstrap") |
| description | The page's own one-sentence description: the Description column of the [Page set](./index.md#page-set) (required; also used in `/llms.txt`) |
| index       | yes |
| image       | default — the social preview card in [Brand assets](../../../design-system.md#brand-assets) |

Product-wide values come from [web metadata](../../../conventions.md#web-metadata).
