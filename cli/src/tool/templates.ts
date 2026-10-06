/**
 * The embedded templates: `<tool>/<target path>` (their path under `cli/templates/`) → template text.
 * The tool folder stays in the key, so a template only serves its own tool's paths.
 *
 * `import.meta.glob` with `?raw` is expanded at build time — by Vite under Vitest and by
 * rolldown under tsdown — so the texts are bundled into the JS and nothing reads the file
 * system at run time. `exhaustive` makes the glob include dot directories such as `.config/`;
 * the negated pattern keeps Finder junk (`.DS_Store`) out of the map and the bundle.
 */

declare global {
  interface ImportMeta {
    glob<T>(
      pattern: string | ReadonlyArray<string>,
      options: {
        readonly query: "?raw";
        readonly import: "default";
        readonly eager: true;
        readonly exhaustive: true;
      },
    ): Record<string, T>;
  }
}

/** Keys glob results (`<root><tool>/<target path>`) by `<tool>/<target path>`. */
export const templatesFrom = (
  root: string,
  modules: Readonly<Record<string, string>>,
): ReadonlyMap<string, string> =>
  new Map(
    Object.entries(modules).map((
      [key, text],
    ) => [key.slice(root.length), text]),
  );

export const templates: ReadonlyMap<string, string> = templatesFrom(
  "../../templates/",
  import.meta.glob<string>([
    "../../templates/**",
    "!../../templates/**/.DS_Store",
  ], {
    query: "?raw",
    import: "default",
    eager: true,
    exhaustive: true,
  }),
);
