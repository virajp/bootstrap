/** Repair on read of docs/blueprint/entities/setup-config/index.md. */
import type { ValidSetupConfig } from "@/setup-config/validity";
import { unremovableTools } from "@/tool/catalog";

/** A valid setup config after Repair on read, the tools it added back and their warnings. */
export interface Repaired {
  readonly config: ValidSetupConfig;
  readonly added: ReadonlyArray<string>;
  readonly warnings: ReadonlyArray<string>;
}

/**
 * Adds back every unremovable tool `values.tools` misses. Validity has already refused a
 * `max: one` category that holds another tool in its place, so each one is a repair.
 */
export const repair = (config: ValidSetupConfig): Repaired => {
  const added = unremovableTools
    .map(tool => tool.name)
    .filter(name => !config.values.tools.includes(name));
  return {
    config: {
      ...config,
      values: { ...config.values, tools: [...config.values.tools, ...added] },
    } as ValidSetupConfig,
    added,
    warnings: added.map(name => `added back unremovable tool \`${name}\``),
  };
};
