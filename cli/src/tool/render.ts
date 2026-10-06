import { Data, Effect } from "effect";
import { Liquid } from "liquidjs";
import type { Tool } from "@/tool/catalog";
import { templates as embedded } from "@/tool/templates";

/** The setup values every template reads (setup-config `values`). */
export interface RenderValues {
  readonly repo: string;
  readonly commit_scopes: ReadonlyArray<string>;
  readonly merge_model: { readonly develop: "direct" | "pr"; readonly main: "direct" | "pr" };
  readonly tools: ReadonlyArray<string>;
}

/** A catalog path of a selected tool has no embedded template. */
export class TemplateMissing extends Data.TaggedError("TemplateMissing")<{ readonly path: string }> {}

/** A template failed to render, e.g. it read an undefined variable. */
export class RenderError extends Data.TaggedError("RenderError")<{ readonly path: string; readonly cause: unknown }> {}

/**
 * LiquidJS matches the tag delimiter before the output delimiter, so `<%=` would read as a tag
 * `<%` named `=`. Templates are authored with `<%=`; the renderer swaps it for this same-length
 * internal delimiter (never present in a template) before parsing.
 */
const outputLeft = "<\u0001=";

const engine = new Liquid({
  tagDelimiterLeft: "<%",
  tagDelimiterRight: "%>",
  outputDelimiterLeft: outputLeft,
  outputDelimiterRight: "%>",
  strictVariables: true,
});

/** The setup values plus the names derived from `repo`: its last segment and the segments before it. */
const scope = (values: RenderValues) => {
  const segments = values.repo.split("/");
  return { ...values, repo_name: segments.at(-1), repo_owner: segments.slice(0, -1).join("/") };
};

/** Renders one target path from its template. */
const renderPath = (
  templates: ReadonlyMap<string, string>,
  context: object,
  path: string,
): Effect.Effect<readonly [string, string], TemplateMissing | RenderError> => {
  const text = templates.get(path);
  return text === undefined
    ? Effect.fail(new TemplateMissing({ path }))
    : Effect.try({
      try: () => [path, engine.parseAndRenderSync(text.replaceAll("<%=", outputLeft), context) as string] as const,
      catch: (cause) => new RenderError({ path, cause }),
    });
};

/** Renders every target path of the selected tools; returns target path → content. */
export const render = (
  selected: Iterable<Tool>,
  values: RenderValues,
  templates: ReadonlyMap<string, string> = embedded,
): Effect.Effect<ReadonlyMap<string, string>, TemplateMissing | RenderError> => {
  const context = scope(values);
  return Effect.forEach(
    [...selected].flatMap((tool) => tool.files),
    (path) => renderPath(templates, context, path),
  ).pipe(Effect.map((entries) => new Map(entries)));
};
