import type { Values } from "@/setup-config/schema";
import {
  type Tool,
  type ToolFile,
  tools,
} from "@/tool/catalog";
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

/**
 * Of the selected tools that have origin hosts, in catalog order: the one whose hosts hold
 * `originHost`, else the first; its name, or "" with none (tool row 21, read by the templates).
 */
const originTool = (selected: ReadonlyArray<Tool>, originHost?: string) => {
  const hosted = selected.filter(tool => tool.originHosts !== undefined);
  return (
    hosted.find(tool =>
      originHost !== undefined && tool.originHosts!.includes(originHost)
    )
      ?? hosted[0]
  )
    ?.name ?? "";
};

/**
 * The setup values, the name derived from `repo` (its last segment), the selected tool names
 * and the origin tool: the variables every template reads.
 */
const scope = (
  values: RenderValues,
  selected: ReadonlyArray<Tool>,
  originHost?: string,
) => ({
  ...values,
  repo_name: values.repo.split("/").at(-1),
  selection: selected.map(tool => tool.name),
  origin_tool: originTool(selected, originHost),
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

/**
 * Renders every target path of the selected tools, in the order given; returns target path →
 * content. `originHost` is the host of the git remote `origin`, if any.
 */
export const render = (
  selected: Iterable<Tool>,
  values: RenderValues,
  templates: ReadonlyMap<string, string> = embedded,
  originHost?: string,
): Effect.Effect<
  ReadonlyMap<string, string>,
  TemplateMissing | RenderError
> => {
  const chosen = [...selected];
  const context = scope(values, chosen, originHost);
  return Effect
    .forEach(
      chosen.flatMap(tool =>
        tool.files.map(file => [tool, file.path] as const)
      ),
      ([tool, path]) => renderPath(templates, context, tool, path),
    )
    .pipe(Effect.map(entries => new Map(entries)));
};

/** The render input: the setup values, the selected tool names and the `origin` host, if any. */
export interface RenderInput {
  readonly values: RenderValues;
  readonly selection: Iterable<string>;
  readonly originHost?: string | undefined;
}

/** A rendered file: its path, its content, and how the writer must write it. */
export interface RenderedFile extends ToolFile {
  readonly content: string;
  readonly executable: boolean;
}

/** The embed keeps text only, so the mode is a rule: every mise task file is executable, nothing else. */
const isExecutable = (path: string) => path.startsWith(".config/mise/tasks/");

/** Renders the files of the selected tools, in catalog order; target path → rendered file. */
export const renderFiles = (
  input: RenderInput,
  templates: ReadonlyMap<string, string> = embedded,
): Effect.Effect<
  ReadonlyMap<string, RenderedFile>,
  TemplateMissing | RenderError
> => {
  const names = new Set(input.selection);
  const selected = tools.filter(tool => names.has(tool.name));
  const files = new Map(
    selected.flatMap(tool =>
      tool.files.map(file => [file.path, file] as const)
    ),
  );
  return render(selected, input.values, templates, input.originHost).pipe(
    Effect.map(rendered =>
      new Map(
        [...rendered].map(([path, content]) => [path, {
          ...files.get(path)!,
          content,
          executable: isExecutable(path),
        }]),
      )
    ),
  );
};
