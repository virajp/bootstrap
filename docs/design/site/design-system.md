# Design System: bootstrap site

The documentation site for `@virajp.dev/bootstrap`. A sibling of virajp.dev:
same family, adapted from a one-page portfolio to long-form technical docs.

**Design read:** documentation site for developers deciding whether to adopt a
repo-tooling CLI, in virajp.dev's quiet, crafted dark language, with terminal
and code samples as first-class content. Dials: variance 4, motion 3, density 5.

## 1. Visual Theme & Atmosphere

Dark only. A warm near-black canvas, never pure black, with one desaturated red
accent. Quiet, crafted, matter-of-fact: it reads like a well-kept README, not a
launch page. Hierarchy comes from hairline rules, spacing and color, not boxes
or shadows. The only "special effect" on the whole site is a single ambient
glow behind the landing hero; docs pages have none.

Copy voice: developer to developer, second person in docs ("run", "you"),
present tense, sentence case everywhere except mono uppercase eyebrows. No
marketing-speak, no emoji.

Accessibility standard: **WCAG 2.2 AA** across the whole site. Every value
below is checked against it; where a token cannot meet it at small sizes, its
use is restricted (see `text-faint`).

## 2. Color Palette & Roles

One theme (dark). One accent hue. Status colors exist only inside terminal
samples and callouts, where they mirror the CLI's own role colors.

| Role | Value | Use | Contrast on `surface-page` |
| --- | --- | --- | --- |
| `surface-page` | `#0d0c0b` | Page canvas | n/a |
| `surface-card` | `#141311` | Raised rows, sidebar hover, table header row | n/a |
| `surface-code` | `#171513` | Code and terminal blocks, inline code | n/a |
| `surface-raised` | `#1b1917` | Search panel, copy-button hover | n/a |
| `border-default` | `#24211e` | Hairline rules between rows and sections, table rules | non-text |
| `border-strong` | `#332f2a` | Emphasised dividers, code-block outline, tag underline (decorative only) | 1.47:1, decorative |
| `border-control` | `#6f6a63` | The visible edge of a control: search input and secondary button outlines | 3.65:1 (3.46:1 on `surface-card`) |
| `text-body` | `#efece7` | Body copy and headings | 16.6:1 |
| `text-muted` | `#a8a39b` | Secondary text, sidebar items, captions, eyebrows, metadata | 7.8:1 |
| `text-faint` | `#6f6a63` | **Large text (24px+) and non-text UI only**: a display headline's second line | 3.65:1 (fails 4.5:1 small) |
| `accent` | `#d9665f` | Links in running copy, primary button fill, active sidebar marker, focus ring, terminal `error` role | 5.6:1 |
| `accent-hover` | `#e8827a` | Hover state of every accent use | 7.4:1 |
| `accent-dim` | accent at 14% | Active sidebar item background, row hover spotlight | non-text |
| `accent-glow` | accent at 28% | The landing hero's ambient glow only | non-text |
| `accent-ink` | `#0d0c0b` | Text on an accent fill | 5.6:1 on accent |
| `status-success` | `#8fbf8a` | Terminal `success` role (created, unchanged); "tip" callout marker | 9.3:1 |
| `status-warning` | `#d9a95c` | Terminal `warning` role (mise not at the expected path, file left in place); "caution" callout marker | 9.1:1 |

Rules:

- **One accent.** `accent` is the only brand hue. `status-success` and
  `status-warning` never appear outside terminal samples and callout markers.
- **Error is the accent.** The terminal `error` role uses `accent`, always with
  a glyph and the word `error:`, so meaning never rests on color alone.
- **`text-faint` is large-only.** Never on body-sized text, sidebar items,
  eyebrows, captions or code; those use `text-muted`. (virajp.dev's eyebrow
  violated this; this site does not repeat it.)
- No gradients on text, no second accent hue, no pure `#000` or `#fff`.

## 3. Typography Rules

Families: **Geist** for display and body, **Geist Mono** for code, terminal
samples, commands, flags, file paths, eyebrows and tags. Self-hosted, never
fetched from a third-party font CDN.

| Step | Size / line height | Weight | Use |
| --- | --- | --- | --- |
| `xs` | 12px / 1.4 | 500 | Eyebrows (mono uppercase), tags |
| `sm` | 13px / 1.5 | 400 | Sidebar items, captions, table cells, code blocks |
| `md` | 16px / 1.65 | 400 | Body copy in docs (default) |
| `lg` | 18px / 1.5 | 500 | Lead paragraph, h4 |
| `xl` | 22px / 1.3 | 600 | h3 |
| `2xl` | 28px / 1.2 | 600 | h2 (doc section) |
| `3xl` | 40px / 1.1 | 600 | h1 (doc page title) |
| `display` | `clamp(40px, 6vw, 64px)` / 1.0 | 600 | Landing hero headline only |

- Body copy is 16px (virajp.dev's 15px is too small for long reading); measure
  capped at 70ch.
- Display weights track `-0.04em`; headings `-0.02em`; body `0`. Mono tracks
  `0`, except uppercase mono eyebrows at `0.12em`.
- Hierarchy comes from weight and color before size. A landing headline may set
  its second line in `text-faint` (display size, so large text).
- Numerals in mono contexts are tabular (exit codes, version strings, counts).
- Inline code: Geist Mono at 0.9em on `surface-code`, 2px radius, no color
  change.
- Banned: serif faces, italic display headlines, all-caps outside eyebrows,
  font sizes below 12px.

## 4. Component Stylings

All components are flat: no shadows anywhere on the site. Radius scale: 2px
(inline code), 4px (buttons, code blocks, callouts, inputs),
12px (search panel only). No pills.

- **Buttons.** `primary`: `accent` fill, `accent-ink` text, at most one per
  view (the landing hero's "Get started"). `secondary`: transparent with a 1px
  `border-control` outline, `text-body`. `ghost`: text-only in `text-muted`.
  Hover shifts color only; press scales to 98%; focus shows the ring. Labels
  never wrap at desktop.
- **Text links.** `accent` in running copy, underlined on hover; external
  links add the `↗` glyph after the text. Sidebar and table-of-contents links
  are not accent (see Navigation).
- **Code block.** `surface-code`, 1px `border-strong` outline, 4px radius,
  `sm` mono, horizontal scroll (never wrap code). A filename or language label
  in `text-muted` mono `xs` at the top-left; a copy button (ghost, top-right)
  that confirms with "Copied" in `text-muted` for 1.5s. No line numbers by
  default.
- **Terminal sample.** A code block whose first line is a `$ ` prompt in
  `text-muted`; output lines may use the four terminal roles: `status-success`,
  `status-warning`, `accent` (error) and `text-body` bold (emphasis), mirroring
  the CLI's own colors. Every colored line also carries its glyph or label
  (`✓`, `!`, `error:`).
- **Callout.** A 2px left rule plus a mono `xs` label, on `surface-card`, 4px
  radius: `note` (rule `border-strong`, label `text-muted`), `tip` (rule and
  label `status-success`), `caution` (rule and label `status-warning`). No
  icons beyond the label, no filled backgrounds in a status hue.
- **Table.** For flags, exit codes and report fields. Header row on
  `surface-card` in `text-muted` `sm` weight 500; body rows divided by
  `border-default` hairlines only (no vertical rules, no zebra stripes); code in
  cells as inline code.
- **Navigation.** Sidebar: section names as mono `xs` uppercase eyebrows in
  `text-muted`, page links in `sm` `text-muted`; the current page in
  `text-body` on `accent-dim` with a 2px `accent` left marker; hover shifts to
  `text-body`. "On this page": `sm` `text-muted` links, the section in view in
  `text-body`. Previous / next page links close every doc page as two
  `secondary`-style rows. No icons in navigation. The mobile menu (below
  900px) moves focus into itself on open, closes on Escape, and returns focus
  to the "Menu" button.
- **Tags.** Mono `xs`, `text-muted`, underlined by a 1px `border-strong`
  hairline, no background, no radius. Used for "since v1.0" style markers.
- **Inputs (search).** The one input on the site: `surface-card`, 1px
  `border-control`, 4px radius, placeholder in `text-muted`, an accessible label
  (placeholder is never the label), a `/` shortcut hint as a tag. Results open in a
  `surface-raised` panel (12px radius, 1px `border-strong`); results are rows,
  matched text in `text-body` weight 600. Keyboard: `/` (or focusing the
  input) opens the panel; Up and Down move through results while focus stays in
  the input; Enter opens the highlighted result; Escape closes the panel and
  returns focus to where it was; the result count is announced to assistive
  tech.
- **Loaders.** Search shows "Searching…" in `text-muted`; no spinners or
  skeletons anywhere else (the site is static).
- **Empty state.** Search with no results: one line naming the query and a
  link to the docs index. 404 page: one sentence, a link home, a link to search.

Interaction states, everywhere: hover (color shift, 120ms), focus-visible (a
2px `surface-page` gap then a 2px `accent` ring, outside the element), active
(press 98% on buttons only), disabled (not used: the site has no disabled
controls). Every interactive target is at least 24x24px (WCAG 2.2 target
size); keyboard reaches everything in reading order, with no traps.

## 5. Layout Principles

- **Spacing scale:** 4, 8, 12, 16, 24, 32, 48, 64, 96, 144 (px). 24px is the
  gutter.
- **Docs layout (three columns at desktop):** left sidebar 240px (section
  navigation), content column max 720px (70ch), right "On this page" table of
  contents 200px. Max page width 1280px, centered.
- **Breakpoints:** below 1200px the table of contents is hidden; below 900px
  the sidebar becomes a menu behind a "Menu" button in the header; below 640px
  tables scroll horizontally inside their own container.
- **Header:** 56px, `surface-page` with a `border-default` bottom hairline.
  Wordmark left, search center-right, GitHub `↗` link right. One line always.
- **Landing page:** single column, 1040px container (the virajp.dev width). An
  asymmetric hero (headline and one terminal sample, 1.25fr / 1fr), then
  sections separated by full-width hairline rules, 64px apart. At most one
  eyebrow per three sections.
- **Docs rhythm:** h2 sections open 48px below the previous block with a
  hairline above; paragraphs 16px apart; code blocks and callouts 24px above
  and below.
- No card grids, no three-equal feature columns, no bento.

## 6. Motion & Interaction

- One easing curve for everything: `cubic-bezier(0.16, 1, 0.3, 1)`.
- Durations: `fast` 120ms (hover color), `base` 240ms (sidebar menu, search
  panel open), `reveal` 600ms (landing entry only).
- Landing only: hero and first section rise 12px and fade in once on load,
  60ms stagger. Docs pages do not animate content.
- Hover: color and background shifts only. Press: buttons scale to 98%.
- Animate only transform and opacity (hover color is the one exception). No
  scroll-linked animation, no parallax, no perpetual loops.
- `prefers-reduced-motion: reduce` disables the entry reveal, the stagger and
  the press scale; state changes become instant.

## 7. Anti-Patterns

- No light mode and no theme toggle (dark only, by decision).
- No second accent hue; status colors never leave terminal samples and callout
  markers.
- No `text-faint` below 24px.
- No `border-strong` as the only edge of a control; controls use `border-control`.
- No shadows, no glow outside the landing hero, no glassmorphism.
- No card grids, bento, three-equal feature columns or pill shapes.
- No div-built fake terminals or screenshots: terminal samples are real text in
  the code-block component, copied from real CLI output.
- No em-dashes in site copy; no emoji; no marketing-speak ("seamless",
  "supercharge", "next-gen").
- No section-number eyebrows, no version labels in the hero, no scroll cues.
- No wrapped code lines, no syntax themes outside this palette.
- No color as the only carrier of meaning (status always has a glyph or label).
