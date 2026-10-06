---
type: vwf-flow-platform
title: Home — site
description: The home page and the not-found page of the public site.
status: reviewed
platform: site
implementation: none
owner: [site]
---

# Home — site

Flow contract: [Home](./index.md)

## Screens → site

| Code | Screen | Route | Reads (operationId) | States (loading/error/empty) | Actions | Form validation |
| ---- | ------ | ----- | ------------------- | ---------------------------- | ------- | --------------- |
| 100a | home | `/` | n/a — static page | default; copy confirmed ("Copied" for 1.5 s); clipboard unavailable (no confirmation). Loading: n/a — the page itself is static. Empty: n/a — fixed content. Error: n/a — static page, no request can fail. The search panel's states ("Searching…" and no results) are per the design-system [Component Behaviors](../../../design-system.md) "Inputs (search)", "Loading" and "Empty and error". | Copy install command; Read the docs; Source ↗; Docs, Search, Source ↗ in the header; three docs rows | n/a — the search input is not a form that validates |
| 100b | not-found | any unknown path | n/a — static page | default only. Loading: n/a — static. Empty: n/a — fixed content. Error: this screen is the error state. | Link home; link to search | n/a |

### `100a` — `home` components

| Component | Rules |
| --------- | ----- |
| Header | Brand mark with the live wordmark "bootstrap", then navigation: "Docs" link to `/docs/` (the introduction of the [documentation flow](../110-documentation/index.md)), Search input, "Source ↗" external link. Always visible. No mobile-menu button: home has no sidebar. |
| Search input | Behaves per the design-system Component Behaviors "Inputs (search)": results open in the search panel under the input, over the documentation pages. Hidden when scripting is off ([reliability](../../../conventions.md#reliability)). |
| Eyebrow (mono uppercase text) | "REPOSITORY SETUP". Always visible; sits above the heading. |
| Heading | "One setup for every repository." |
| Description (text) | One line; the product default description from [web metadata](../../../conventions.md#web-metadata). |
| Code block | Shows `pnpx @virajp.dev/bootstrap@latest init`; the text stays selectable. Holds a ghost copy button. |
| Copy button (ghost) | Activating it places exactly `pnpx @virajp.dev/bootstrap@latest init` on the clipboard and shows "Copied" for 1.5 seconds. If the clipboard is unavailable, no confirmation shows and the command stays visible and selectable. Hidden when scripting is off, never inert. |
| "Read the docs" (primary button) | Activating it opens `/docs/`, the introduction of the documentation flow, [110](../110-documentation/index.md). |
| "Source ↗" (ghost external link) | Opens the source repository; external. |
| Terminal sample | Shows `bootstrap check` reporting drift, exactly these lines. Static; not interactive.<br>`$ pnpx @virajp.dev/bootstrap@latest check`<br>`⚠ modified   .gitignore             (git)` — warning role with its glyph<br>`✓ unchanged  49 files` — success role with its glyph<br>(blank line)<br>``drift found in 1 file — run `bootstrap update` `` — emphasis on the command<br>`(exit 1)` — muted |
| Rows (three, rules not cards) | Each row links to its command's page in the [documentation flow](../110-documentation/index.md): Set up to `/docs/commands/init/`, Check to `/docs/commands/check/`, Update to `/docs/commands/update/`. Content: "Set up: `bootstrap init` writes the formatters, commit gates, scanners and tasks into a repository."; "Check: `bootstrap check` reports drift as a JSON report that an AI agent can act on."; "Update: `bootstrap update` brings a repository back to the shared source and keeps your own edits." |
| Footer | Shows "MIT License" and a "Source ↗" external link. |

### `100a` — `home` metadata

| Field       | Value |
| ----------- | ----- |
| title       | bootstrap \| one setup for every repository |
| description | The product default description, per [web metadata](../../../conventions.md#web-metadata) |
| index       | yes |
| image       | default — the social preview card in [Brand assets](../../../design-system.md#brand-assets) |

### `100b` — `not-found` components

| Component | Rules |
| --------- | ----- |
| Header | The same site header as `100a`: brand mark, live wordmark "bootstrap", "Docs" link to `/docs/`, Search input, "Source ↗" external link. The Search input behaves as on `100a`. |
| Message (text) | One sentence that states the page does not exist, per the design-system Empty and error pattern. |
| Home link | Activating it opens `100a`. Always visible. |
| Search link | Activating it puts focus on the header Search input. With scripting off, the link and the input are hidden ([reliability](../../../conventions.md#reliability)); the home link remains. |

### `100b` — `not-found` metadata

| Field       | Value |
| ----------- | ----- |
| title       | Page not found \| bootstrap |
| description | The product default description, per [web metadata](../../../conventions.md#web-metadata) |
| index       | no |
| image       | default — the social preview card in [Brand assets](../../../design-system.md#brand-assets) |
