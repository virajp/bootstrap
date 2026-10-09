import type { Values } from "@/setup-config/schema";
import type { Tool } from "@/tool/catalog";
import { templates as embedded } from "@/tool/templates";
import {
  Data,
  Effect,
} from "effect";
import { Liquid } from "liquidjs";

/** The setup values every template reads: the setup-config `values` type itself. */
export type RenderValues = typeof Values.Type;

/** A catalog path of a selected tool has no embedded template. */
export class TemplateMissing
  extends Data.TaggedError("TemplateMissing")<{ readonly path: string; }>
{}

/** A template failed to render, e.g. it read an undefined variable. */
export class RenderError extends Data.TaggedError("RenderError")<
  { readonly path: string; readonly cause: unknown; }
> {}

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
  strictFilters: true,
});

/** The setup values plus the name derived from `repo`: its last segment. */
const scope = (values: RenderValues) => ({
  ...values,
  repo_name: values.repo.split("/").at(-1),
});

/** Renders one target path of a tool from the template in that tool's folder. */
const renderPath = (
  templates: ReadonlyMap<string, string>,
  context: object,
  tool: Tool,
  path: string,
): Effect.Effect<readonly [string, string], TemplateMissing | RenderError> => {
  const text = templates.get(`${tool.name}/${path}`);
  return text === undefined
    ? Effect.fail(new TemplateMissing({ path }))
    : Effect.try({
      try: () =>
        [
          path,
          engine.parseAndRenderSync(
            text.replaceAll("<%=", outputLeft),
            context,
          ) as string,
        ] as const,
      catch: cause => new RenderError({ path, cause }),
    });
};

/** Renders every target path of the selected tools; returns target path → content. */
export const render = (
  selected: Iterable<Tool>,
  values: RenderValues,
  templates: ReadonlyMap<string, string> = embedded,
): Effect.Effect<
  ReadonlyMap<string, string>,
  TemplateMissing | RenderError
> => {
  const context = scope(values);
  return Effect
    .forEach(
      [...selected].flatMap(tool =>
        tool.files.map(file => [tool, file.path] as const)
      ),
      ([tool, path]) => renderPath(templates, context, tool, path),
    )
    .pipe(Effect.map(entries => new Map(entries)));
};

/** A rendered file: its content and whether the writer must make it executable. */
export interface RenderedFile {
  readonly content: string;
  readonly executable: boolean;
}

/** The embed keeps text only, so the mode is a rule: every mise task file is executable, nothing else. */
const isExecutable = (path: string) => path.startsWith(".config/mise/tasks/");

/** Like `render`, but each target path carries whether the file is executable. */
export const renderFiles = (
  selected: Iterable<Tool>,
  values: RenderValues,
  templates: ReadonlyMap<string, string> = embedded,
): Effect.Effect<
  ReadonlyMap<string, RenderedFile>,
  TemplateMissing | RenderError
> =>
  render(selected, values, templates).pipe(
    Effect.map(rendered =>
      new Map(
        [...rendered].map((
          [path, content],
        ) => [path, { content, executable: isExecutable(path) }]),
      )
    ),
  );
