---
type: vwf-design-system
title: Design System
description: Product-wide UX/visual contract every blueprint screen references.
status: reviewed # draft | reviewed | stable
derived: false
timestamp: 2026-10-06
---

# Design System

**Stored import.** The design system is authored in the design tool's canvas
(`docs/design/site/design-system.md` and `docs/design/site/brand/`) and imported
here, so it is authoritative until someone changes the canvas. Hand-edits to this
doc are drift: change the canvas and re-run `/vwf:design-system`. The Terminal UX
section is the exception: it is elicited in text, because a terminal has no
canvas.

## Brand & Mood

A dark-only documentation site that reads like a well-kept README, not a launch
page. Warm near-black canvas (never pure black) with one desaturated red accent.
Quiet, crafted, matter-of-fact: hierarchy comes from hairline rules, spacing and
color, never boxes or shadows. The only special effect on the whole site is one
ambient glow behind the landing hero. Voice: developer to developer, second
person in docs, present tense, sentence case except mono uppercase eyebrows; no
marketing-speak, no emoji. Keywords: quiet, crafted, restrained, technical.

## Color Tokens

Dark only. No light mode and no theme toggle are in scope.

| Token (role) | Dark | Usage | Contrast pairing |
| --- | --- | --- | --- |
| `surface-page` | `#0d0c0b` | Page canvas | n/a |
| `surface-card` | `#141311` | Raised rows, sidebar hover, table header row, callouts | n/a |
| `surface-code` | `#171513` | Code and terminal blocks, inline code | n/a |
| `surface-raised` | `#1b1917` | Search panel, copy-button hover | n/a |
| `border-default` | `#24211e` | Hairline rules between rows and sections, table rules | non-text |
| `border-strong` | `#332f2a` | Emphasised dividers, code-block outline, tag underline (decorative only) | 1.47:1 on `surface-page`, decorative only |
| `border-control` | `#6f6a63` | The visible edge of a control: search input and secondary button outlines | 3.65:1 on `surface-page`, 3.46:1 on `surface-card` |
| `text-body` | `#efece7` | Body copy and headings | 16.6:1 on `surface-page` |
| `text-muted` | `#a8a39b` | Secondary text, sidebar items, captions, eyebrows, metadata | 7.8:1 on `surface-page` |
| `text-faint` | `#6f6a63` | Large text (24px and up) and non-text UI only | 3.65:1 on `surface-page`: fails 4.5:1, passes 3:1 |
| `accent` | `#d9665f` | Links in running copy, primary button fill, active sidebar marker, focus ring, terminal `error` role | 5.6:1 on `surface-page` |
| `accent-hover` | `#e8827a` | Hover state of every accent use | 7.4:1 on `surface-page` |
| `accent-dim` | accent at 14% opacity | Active sidebar item background, hover spotlight | non-text |
| `accent-glow` | accent at 28% opacity | The landing hero's ambient glow only | non-text |
| `accent-ink` | `#0d0c0b` | Text on an accent fill | 5.6:1 on `accent` |
| `status-success` | `#8fbf8a` | Terminal `success` role; "tip" callout marker | 9.3:1 on `surface-page` |
| `status-warning` | `#d9a95c` | Terminal `warning` role; "caution" callout marker | 9.1:1 on `surface-page` |

Contrast on the other surfaces (every text and status pairing in use):

| Foreground | `surface-card` | `surface-code` | `surface-raised` | `accent-dim` over page |
| --- | --- | --- | --- | --- |
| `text-body` | 15.8:1 | 15.5:1 | 14.9:1 | 14.3:1 |
| `text-muted` | 7.4:1 | 7.3:1 | 7.0:1 | 6.7:1 |
| `accent` | 5.3:1 | 5.2:1 | 5.0:1 | 4.8:1 |
| `status-success` | 8.8:1 | 8.7:1 | 8.3:1 | 8.0:1 |
| `status-warning` | 8.7:1 | 8.5:1 | 8.2:1 | 7.8:1 |

- **One accent.** `accent` is the only brand hue; `status-success` and
  `status-warning` never appear outside terminal samples and callout markers.
- **Error is the accent.** The terminal `error` role uses `accent`, always with a
  glyph and the word `error:`, so meaning never rests on color alone.
- **`text-faint` is large-only.** Never on body-sized text, sidebar items,
  eyebrows, captions or code; those use `text-muted`.
- No gradients on text, no second accent hue, no pure `#000` or `#fff`.

## Typography

Families: **Geist** for display and body, **Geist Mono** for code, terminal
samples, commands, flags, file paths, eyebrows and tags.

| Step | Size / Line height | Weight | Use |
| --- | --- | --- | --- |
| `xs` | 12px / 1.4 | 500 | Eyebrows (mono uppercase), tags |
| `sm` | 13px / 1.5 | 400 | Sidebar items, captions, table cells, code blocks |
| `md` | 16px / 1.65 | 400 | Body copy in docs (default) |
| `lg` | 18px / 1.5 | 500 | Lead paragraph, h4 |
| `xl` | 22px / 1.3 | 600 | h3 |
| `2xl` | 28px / 1.2 | 600 | h2 (doc section) |
| `3xl` | 40px / 1.1 | 600 | h1 (doc page title) |
| `display` | fluid 40px to 64px / 1.0 | 600 | Landing hero headline only |

- Body measure is capped at 70ch.
- Tracking: display `-0.04em`, headings `-0.02em`, body `0`, mono `0`; uppercase
  mono eyebrows `0.12em`.
- Hierarchy comes from weight and color before size. A landing headline may set
  its second line in `text-faint` (display size, so large text).
- Numerals in mono contexts are tabular (exit codes, version strings, counts).
- Inline code: mono at 0.9em on `surface-code`, 2px radius, no color change.
- Banned: serif faces, italic display headlines, all-caps outside eyebrows,
  sizes below 12px.

## Spacing & Layout

- Spacing scale: 4, 8, 12, 16, 24, 32, 48, 64, 96, 144 (px); 24px is the gutter.
- Grid / container: docs pages are three columns at desktop: a 240px section
  sidebar, a content column of at most 720px (70ch) and a 200px "On this page"
  column, within a 1280px maximum page width, centered. The landing page is a
  single 1040px column with an asymmetric hero (1.25fr / 1fr).
- Breakpoints (width to behavioral intent):
  - below 1200px: the "On this page" column is hidden.
  - below 900px: the sidebar becomes a menu behind a "Menu" button in the header.
  - below 640px: tables scroll horizontally inside their own container.
- Header: 56px, one line always: wordmark left, search, GitHub link right, with a
  `border-default` bottom hairline.
- Rhythm: h2 sections open 48px below the previous block with a hairline above;
  paragraphs sit 16px apart; code blocks and callouts have 24px above and below;
  landing sections are separated by full-width hairlines 64px apart.
- Radius scale: 2px (inline code), 4px (buttons, code blocks,
  callouts, inputs), 12px (search panel only). No pills.
- Elevation scale: flat. There are no shadows anywhere on the site.

## Motion

- Duration tokens: `fast` 120ms (hover color), `base` 240ms (menu, search panel),
  `reveal` 600ms (landing entry only).
- Easing tokens: one curve for everything, a strong decelerating ease-out with
  control points `0.16, 1, 0.3, 1`.
- Principles:
  - Landing only: the hero and first section rise 12px and fade in once on load,
    60ms stagger. Docs pages do not animate content.
  - Hover shifts color and background only; press scales buttons to 98%.
  - Animate only transform and opacity (hover color is the one exception). No
    scroll-linked animation, no parallax, no perpetual loops.
- Reduced-motion behavior: when the reader's operating system asks for reduced
  motion, the entry reveal, the stagger and the press scale are off; state
  changes become instant.

## Brand assets

- Favicon source mark: the brand mark (`docs/design/site/brand/logo.svg`), one square vector
  drawing on a rounded accent tile; the whole favicon set and app icons are
  rasterized from it.
- Social preview: a flat 1280x640 card showing the mark centered on the page
  canvas with the product name in mono beside it, no photograph and no terminal
  screenshot. Alt text: the product name followed by its one-line description.
- Theme colour: the `surface-page` Color Token role.

## Accessibility Standard

- Conformance target: **WCAG 2.2 AA** across the whole site.
- Contrast: text at least 4.5:1; large text (24px, or 18.66px bold) and non-text
  UI at least 3:1; the `text-faint` rule under Color Tokens exists to hold this
  line. Status color never carries meaning alone; it always has a glyph or label.
- Keyboard: every interactive element is reachable and operable in reading
  order with no traps. Visible focus is a 2px `surface-page` gap then a 2px
  `accent` ring, drawn outside the element.
- Focus management: the mobile menu moves focus into itself on open, closes on
  Escape and returns focus to its button; the search panel keeps focus in its
  input and returns focus on Escape (see Component Behaviors).
- Non-text contrast: every control's identifying edge is at least 3:1
  (`border-control`); `border-strong` is decorative only.
- Targets: every interactive target is at least 24x24px.
- Semantics: heading order is preserved with no skipped levels; every link has
  discernible text (never "click here"); the external-link glyph supplements,
  never replaces, link text; the search input has an accessible label (the
  placeholder is never the label).
- Motion and media: honor reduced motion per the Motion section; no autoplaying
  audio or video.

## Component Behaviors

- **Buttons**: `primary` (accent fill, at most one per view), `secondary`
  (1px `border-control` outline) and `ghost` (text only, `text-muted`). Hover
  shifts color only; press scales to 98%; labels never wrap at desktop.
- **Text links**: `accent` in running copy, underlined on hover; external links
  add the `↗` glyph.
- **Code block**: `surface-code`, 1px `border-strong` outline, 4px radius, `sm`
  mono, horizontal scroll and never wrapped; a filename or language label at the
  top left; a ghost copy button at the top right that confirms with "Copied" for
  1.5 seconds. No line numbers by default.
- **Terminal sample**: a code block whose first line is a `$ ` prompt in
  `text-muted`; output lines may use `status-success`, `status-warning`,
  `accent` (error) and bold `text-body` (emphasis), mirroring the CLI's own
  roles. Every colored line also carries its glyph or label.
- **Callout**: a 2px left rule plus a mono `xs` label on `surface-card`: `note`
  (neutral), `tip` (`status-success`), `caution` (`status-warning`). No icons, no
  filled status backgrounds.
- **Table**: for flags, exit codes and report fields; header row on
  `surface-card` in `text-muted`; rows divided by `border-default` hairlines
  only, with no vertical rules or zebra stripes.
- **Tags**: mono `xs`, `text-muted`, underlined by a 1px `border-strong`
  hairline, no background, no radius.
- **Navigation**: sidebar section names are mono uppercase eyebrows in
  `text-muted`; page links `sm` `text-muted`; the current page in `text-body` on
  `accent-dim` with a 2px `accent` left marker; "On this page" links in
  `text-muted` with the section in view in `text-body`; previous and next links
  close every doc page. No icons in navigation.
- **Inputs (search)**: the one input on the site; `surface-card`, 1px
  `border-control`, 4px radius, a `/` shortcut hint. Results open in a
  `surface-raised` panel as rows with the match in bold `text-body`. Keyboard:
  `/` (or focusing the input) opens the panel; Up and Down move through results
  while focus stays in the input; Enter opens the highlighted result; Escape
  closes the panel and returns focus to where it was; the result count is
  announced to assistive tech.
- **Loading**: search shows "Searching…" in `text-muted`; no spinners or
  skeletons elsewhere, since the site is static.
- **Empty and error**: search with no results shows one line naming the query
  and a link to the docs index; the not-found page is one sentence with a link
  home and a link to search.
- **Overlays and feedback**: none in scope beyond the search panel and the copy
  confirmation.
- Interaction states everywhere: hover (color shift), focus-visible (the ring),
  active (press scale on buttons only); no disabled controls exist.

## Brand

- Logo source: `docs/design/site/brand/logo.svg`
- Variants:

| Variant | Path | Use |
| --- | --- | --- |
| Mark | `docs/design/site/brand/logo.svg` | Site header at 22px before the wordmark (live mono text in `text-body`); the favicon source |
| Single colour | `docs/design/site/brand/mono.svg` | Where the accent cannot be used (one-colour print, monochrome badges); takes the surrounding text color; no stem fade |

- Clear space: one quarter of the mark's height on every side.
- Minimum sizes (per variant): 16px for both variants.
- Rules (what never happens to the mark):
  - Never recolour the mark's square; `mono.svg` is the only single-colour form.
  - Never rotate, stretch, add a shadow or add a glow.
  - Never place the mark on an `accent` surface.
  - The faded stem appears only in `logo.svg`.
  - The wordmark is always live text beside the mark, never merged into it.

## Terminal UX

The `cli` project is a shipped command-line tool. Elicited in text on
2026-10-06.

- Output formatting: human-readable text by default. Every command takes
  `--json`, which prints exactly one JSON document to stdout and nothing else.
  Results go to stdout; progress, warnings and errors go to stderr. `--quiet`
  shows errors only; `--verbose` shows each per-file decision.
- Color semantics: color carries meaning only. `success` (created, unchanged) maps
  to `status-success`; `warning` (drift found, kept files) to `status-warning`;
  `error` to `accent`; `emphasis` (paths, command names) to bold `text-body`.
  Roles map to the terminal theme's palette by role name, never fixed hex. Color
  is off when `NO_COLOR` is set, when the output is not a terminal, with
  `--no-color`, and always under `--json`.
- Progress conventions: an animated spinner per step in an interactive terminal;
  when piped, in CI or non-interactive, one stable line per finished step on
  stderr instead.
- Errors & exit codes: `0` success with no drift; `1` drift found (`check`) or
  changes declined; `2` usage error (prints short usage); `3` failure
  (unreadable config, template error, write failed). Every error states what
  happened, why, and the exact next command; no stack trace unless `--verbose`.
- Help & naming conventions: commands are single verbs (`init`, `check`,
  `update`, `add <group>`). Every flag has a long form; short forms only for the
  most common (`-q`, `-v`, `-y`). Booleans negate with `--no-<flag>`. Help is a
  one-line purpose, usage, flags, then one or two real examples. The bare
  command prints help and exits `2`.

## Anti-Patterns

- No light mode and no theme toggle (dark only, by decision).
- No second accent hue; status colors never leave terminal samples and callout
  markers.
- No `text-faint` below 24px.
- No `border-strong` as the only edge of a control; controls use `border-control`.
- No shadows, no glow outside the landing hero, no glassmorphism.
- No card grids, bento, three-equal feature columns or pill shapes.
- No fake terminals or screenshots built from layout boxes: terminal samples are
  real text in the code-block component, copied from real CLI output.
- No em-dashes in site copy, no emoji, no marketing-speak.
- No section-number eyebrows, no version labels in the hero, no scroll cues.
- No wrapped code lines.
- No color as the only carrier of meaning.

## Open Questions

- [ ] 2026-10-06 - The social preview card (the mark on the page canvas) does not
      exist as a file yet; produce it with the site's first build.
- [ ] 2026-10-06 - Syntax highlighting for code blocks is not specified beyond
      "inside this palette"; decide the token-to-role mapping when the first code
      samples are written.
