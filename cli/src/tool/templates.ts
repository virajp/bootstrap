/**
 * The embedded templates: `cli/templates/<tool>/<target path>` → template text.
 *
 * `import.meta.glob` with `?raw` is expanded at build time — by Vite under Vitest and by
 * rolldown under tsdown — so the texts are bundled into the JS and nothing reads the file
 * system at run time. `exhaustive` makes the glob include dot directories such as `.config/`.
 */

declare global {
  interface ImportMeta {
    glob<T>(
      pattern: string,
      options: { readonly query: "?raw"; readonly import: "default"; readonly eager: true; readonly exhaustive: true },
    ): Record<string, T>;
  }
}

/** Keys glob results (`<root><tool>/<target path>`) by target path. */
export const templatesFrom = (
  root: string,
  modules: Readonly<Record<string, string>>,
): ReadonlyMap<string, string> =>
  new Map(
    Object.entries(modules).map(([key, text]) => [key.slice(root.length).replace(/^[^/]+\//, ""), text]),
  );

export const templates: ReadonlyMap<string, string> = templatesFrom(
  "../../templates/",
  import.meta.glob<string>("../../templates/**", {
    query: "?raw",
    import: "default",
    eager: true,
    exhaustive: true,
  }),
);
