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
| 100a | home | `/` | n/a — static page | default only. Loading: n/a — the page itself is static. Empty: n/a — fixed content. Error: search unavailable — when the search index cannot load, the search panel shows exactly "Search is not available." and the rest of the page works. The search panel's states ("Searching…" and no results) are per the design-system [Component Behaviors](../../../design-system.md#component-behaviors) "Inputs (search)", "Loading" and "Empty and error". | Read the docs; Install; Source ↗ (hero and footer); Docs, Search, Source ↗ in the header; three docs rows (the Add or remove row has two links) | n/a — the search input is not a form that validates |
| 100b | not-found | any unknown path | n/a — static page | default only. Loading: n/a — static. Empty: n/a — fixed content. Error: this screen is the error state; an unknown path is served with HTTP status 404 and this screen. | Link home; link to search | n/a |

### `100a` — `home` components

| Component | Rules |
| --------- | ----- |
| Header | Brand mark with the live wordmark "bootstrap", then navigation: "Docs" link to `/docs/` (the introduction of the [documentation flow](../110-documentation/index.md)), Search input, "Source ↗" external link. Always visible. No mobile-menu button. |
| Search input | Behaves per the design-system Component Behaviors "Inputs (search)": results open in the search panel under the input, over the documentation pages. Hidden when scripting is off ([reliability](../../../conventions.md#reliability)). |
| Eyebrow (text) | "REPOSITORY SETUP". Always visible; sits above the heading. |
| Heading | "One setup for every repository." |
| Description (text) | One line; the product default description from [web metadata](../../../conventions.md#web-metadata). |
| Command display ([component behaviors](../../../design-system.md#component-behaviors)) | Shows exactly `bootstrap init`. |
| "Install" (text link) | Sits under the Command display. Activating it opens `/docs/start/quick-start/`, the Quick start page of the [documentation flow](../110-documentation/index.md). Always visible. |
| "Read the docs" (primary button) | Activating it opens `/docs/`, the introduction of the documentation flow, [110](../110-documentation/index.md). |
| "Source ↗" (ghost external link) | Opens the source repository (pinned in [web metadata](../../../conventions.md#web-metadata)); external. |
| Terminal sample | Shows a first run of `bootstrap init`, exactly these lines, matching the interactive output of the [setup-repository flow](../../cli/110-setup-repository/index.md). Static; not interactive.<br>`$ bootstrap init`<br>`? repository › virajp/app` — muted prompt marker<br>`? commit scopes › api, web`<br>`? merge into develop › direct`<br>`? merge into main › pr`<br>`? tools › 13 selected`<br>`? write 62 files › yes`<br>`✓ created  62 files` — success role with its glyph<br>(blank line)<br>``commit the changes, then run `MISE_ENV=dev mise run setup:all` `` — emphasis on the command<br>`(exit 0)` — muted |
| Rows (three) | Each row links to its command's page in the [documentation flow](../110-documentation/index.md): Set up to `/docs/commands/init/`, Add or remove to both `/docs/commands/add/` (`bootstrap add`) and `/docs/commands/remove/` (`bootstrap remove`), Select tools to `/docs/commands/tui/`. Content: "Set up: `bootstrap init` writes the formatters, commit gates, scanners and tasks into a repository; a re-run brings it up to date."; "Add or remove: `bootstrap add` and `bootstrap remove` change a repository's tools, and git keeps every earlier version."; "Select tools: `bootstrap tui` shows every tool by category in one view." |
| Footer | Shows "MIT License" and a "Source ↗" external link to the source repository. |

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
| Header | The same site header as `100a`. |
| Message (text) | Exactly "This page does not exist.", per the design-system Empty and error pattern. |
| Home link | Activating it opens `100a`. Always visible. |
| Search link | Activating it puts focus on the header Search input. With scripting off, the link and the input are hidden ([reliability](../../../conventions.md#reliability)); the home link remains. |

### `100b` — `not-found` metadata

| Field       | Value |
| ----------- | ----- |
| title       | Page not found \| bootstrap |
| description | The product default description, per [web metadata](../../../conventions.md#web-metadata) |
| index       | no |
| image       | default — the social preview card in [Brand assets](../../../design-system.md#brand-assets) |
