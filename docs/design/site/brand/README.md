# bootstrap brand mark

A geometric lowercase `b` cut from an accent rounded square: the sibling of
virajp.dev's V mark (same 64-unit container, radius 8, glyph in `surface-page`,
one stroke carrying the family fade). Here the stem fades upward in three flat
opacity steps (40%, 70%, 100%).

## Variants

| Variant       | Path                              | Where it goes                                                                                                                                                              |
| ------------- | --------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Mark          | `docs/design/site/brand/logo.svg` | The site header at 22px, before the wordmark `bootstrap` set as live Geist Mono text in `text-body`. The favicon source: every favicon and app icon is rasterized from it. |
| Single colour | `docs/design/site/brand/mono.svg` | Anywhere the accent cannot be used (one-colour print, monochrome badges). It takes the surrounding text colour (`currentColor`); its stem has no fade.                     |

## Clear space

One quarter of the mark's height on every side (16 units on the 64-unit grid).

## Minimum size

16px for both variants.

## Rules

- Never recolour the mark's square; `mono.svg` is the only single-colour form.
- Never rotate, stretch, add a shadow or add a glow.
- Never place the mark on an `accent` surface: the mark already carries its own
  accent tile (part of `logo.svg`), and it would vanish into an accent
  background.
- The faded stem appears only in `logo.svg`.
- The wordmark is always live text beside the mark, never merged into it.
