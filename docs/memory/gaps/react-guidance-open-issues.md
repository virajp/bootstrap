G:react-guidance 2026-10-06 site astro-ssg — reviewer issues accepted at landing

# React guidance: open issues (accepted 2026-10-06)

The generated `framework/react` doctrine (`.claude/skills/react/`) was landed
after six reviewer rounds with these issues knowingly left open by the user.
Fix them during the website slice (Documentation) or upstream in the stackgen
plugin.

- **Host claims without a source** — `hydration.md`, `island-boundary.md` and
  `conventions.md` state how the host renders islands ("rendered twice", "each
  island its own root", "lazy hydration trigger") without citing the host's
  docs; false for the host's client-only mode.
- **Rules of Hooks** — React documents `eslint-plugin-react-hooks`
  (`recommended`) as its check, so "no lint gate" should not be an equal option;
  only where the check lands is open. The house linter `@askviraj/linter`
  bundles no React plugin today.
- **`.tsx` compiler keys** — the `jsx` / `jsxImportSource` doctrine lives in a
  skill scoped to `**/*.tsx`, so editing `tsconfig.json` never surfaces it.

Upstream gap (stackgen): the cross-framework integration judgment (kinds.md
topic 1) belongs to `language/typescript` standards and no shipped pack carries
it — recorded UNCOVERED in `.claude/stackgen/citations/react.yaml`.
