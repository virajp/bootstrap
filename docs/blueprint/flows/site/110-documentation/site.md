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
| 110a | doc-page | every address of the Page set in [index](./index.md#page-set), under `/docs/` | n/a — static page | default; mobile menu open and closed; scripting off; narrow view with scripting off; copy confirmed; clipboard unavailable; search panel "Searching…", no results, and search unavailable (the panel shows "Search is not available."; the rest of the page works), per the design-system [Component Behaviors](../../../design-system.md#component-behaviors) "Inputs (search)", "Loading" and "Empty and error". Loading: n/a — static. Empty: n/a — every page has content. Error: search unavailable (above); an unknown address shows `100b` not-found. | Sidebar link; "On this page" link; previous and next links; copy command; Search; open and close the mobile menu | n/a — the search input is not a form that validates |

### `110a` — `doc-page` components

| Component | Rules |
| --------- | ----- |
| Header | The site header of `100a` in [Home](../100-home/site.md), plus, on a narrow view only, the button that opens the mobile menu. Always visible. With scripting off, that button is hidden, never inert ([reliability](../../../conventions.md#reliability)). |
| Search input | Behaves per the design-system "Inputs (search)" and searches the full text of every docs page. Each result shows the page title, the matching section heading, and a short text snippet containing the match. Hidden when scripting is off ([reliability](../../../conventions.md#reliability)). |
| Sidebar navigation | Lists the pages of the Page set in [index](./index.md#page-set), in that order, under the section eyebrows START, COMMANDS, GUIDES and REFERENCE ([Navigation](../../../design-system.md#component-behaviors)). The current page is marked. Each entry opens its page. On a narrow view it is the mobile menu: focus moves into it on open, it closes on Escape, on the menu button (a toggle), on a tap outside the menu, or when a page link is selected; on close, focus returns to its button. Works with scripting off; on a narrow view with scripting off, it is shown inline above the page content. |
| Page content | Heading, prose, code blocks, terminal samples, tables and callouts of the page. Always visible. Each section has a heading; heading levels do not skip. |
| Code block | Per the design-system [Code block](../../../design-system.md#component-behaviors); the text stays selectable. Holds a ghost copy button. |
| Copy button (ghost) | Activating it places exactly the text of its code block on the clipboard and shows "Copied" for 1.5 seconds. If the clipboard is unavailable, no confirmation shows and the command stays selectable. Hidden when scripting is off, never inert. |
| Terminal sample | Per the design-system [Terminal sample](../../../design-system.md#component-behaviors). Static; not interactive. |
| Table | Per the design-system [Table](../../../design-system.md#component-behaviors). Used for flags, exit codes and JSON result fields. |
| Callout | Per the design-system [Callout](../../../design-system.md#component-behaviors), with the kind `note`, `tip` or `caution`. Shown only where a page has one. |
| "On this page" list | Lists the page's level-2 and level-3 headings; level-3 entries are indented under their level-2 section. Each link moves to its section; the section in view is emphasised. Hidden on a page with no level-2 or level-3 headings. Works with scripting off, without the in-view emphasis. |
| Previous and next links | Placed at the end of every page, after the content, and follow the sidebar order. Each shows the neighbour page title. The first page has no previous link. The last page has no next link. |
| Footer | The site footer of `100a` in [Home](../100-home/site.md). |

### `110a` — `doc-page` metadata

| Field       | Value |
| ----------- | ----- |
| title       | `<page title> \| <site name>` (Site name per [web metadata](../../../conventions.md#web-metadata)), where the page title is the Page column of the [Page set](./index.md#page-set) (for example "bootstrap init \| bootstrap") |
| description | The page's own one-sentence description: the Description column of the [Page set](./index.md#page-set) (required; also used in `/llms.txt`) |
| index       | yes |
| image       | default — the social preview card in [Brand assets](../../../design-system.md#brand-assets) |

Product-wide values come from [web metadata](../../../conventions.md#web-metadata).
