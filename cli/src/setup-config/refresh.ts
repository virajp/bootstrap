/** Setup-config invariants 2 and 4–7: the refresh of `files` and the rewrite rule. */
import {
  type SetupConfig,
  setupConfigPath,
} from "@/setup-config/schema";
import { toolsByName } from "@/tool/catalog";

/**
 * Full (init, add, tui apply): the rendered set. Remove: the recorded paths minus those of the
 * removed tools, plus those of the tools Repair on read added back.
 */
export type RefreshMode =
  | { readonly mode: "full"; }
  | {
    readonly mode: "remove";
    readonly removed: ReadonlyArray<string>;
    readonly added: ReadonlyArray<string>;
  };

/** The paths of catalog tools, named by tool names the caller has already validated. */
const pathsOf = (names: ReadonlyArray<string>) =>
  names.flatMap(name => toolsByName.get(name)!.files.map(file => file.path));

/**
 * The next config with `files` refreshed, and the recorded paths not rendered but on disk
 * (`orphaned`). A path neither rendered nor on disk leaves; the setup config never stays.
 */
export const refreshFiles = (
  config: SetupConfig,
  mode: RefreshMode,
  rendered: ReadonlySet<string>,
  present: ReadonlySet<string>,
): { readonly config: SetupConfig; readonly orphaned: ReadonlyArray<string>; } => {
  const candidates = mode.mode === "full"
    ? new Set([...config.files, ...rendered])
    : new Set([...config.files, ...pathsOf(mode.added)]).difference(
      new Set(pathsOf(mode.removed)),
    );
  const files = [...candidates]
    .filter(path => path !== setupConfigPath && (rendered.has(path) || present.has(path)))
    .toSorted();
  return {
    config: { ...config, files },
    orphaned: files.filter(path => !rendered.has(path)),
  };
};

/** The command that may rewrite the setup config. */
export type RewriteMode = "init" | "tui-apply" | "add" | "remove";

/**
 * Invariant 7: init and the tui apply rewrite when the render differs from the recorded text;
 * add and remove only on a run with a tool change or a repair.
 */
export const rewrites = (
  mode: RewriteMode,
  recorded: string,
  next: string,
  toolChangeOrRepair: boolean,
): boolean =>
  mode === "init" || mode === "tui-apply" ? next !== recorded : toolChangeOrRepair;
