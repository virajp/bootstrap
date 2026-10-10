/** The Tool category catalog (1.0) of docs/blueprint/entities/tool-category/index.md: read-only data, no I/O. */

export interface ToolCategory {
  readonly name: string;
  /** How many tools of the category a repository can select at once. */
  readonly max: "one" | "many";
  readonly purpose: string;
}

/** The categories in display order. */
export const categories: ReadonlyArray<ToolCategory> = [
  { name: "tool-manager", max: "one", purpose: "Installs the tools and runs the task library" },
  { name: "version-control", max: "one", purpose: "Ignore rules and git settings" },
  { name: "git-hooks", max: "one", purpose: "Runs the gates before each commit" },
  { name: "editor", max: "many", purpose: "Editor defaults for the repository" },
  { name: "formatter", max: "many", purpose: "Formats the repository's files" },
  { name: "linter", max: "many", purpose: "Finds errors in code" },
  { name: "secret-scanner", max: "one", purpose: "Finds secrets in code and commits" },
  { name: "vulnerability-scanner", max: "many", purpose: "Finds known vulnerabilities in dependencies" },
  { name: "secrets-manager", max: "one", purpose: "Gives the repository's secrets to its tasks" },
  { name: "ai-agent", max: "many", purpose: "Settings for an AI coding agent" },
  { name: "knowledge-graph", max: "one", purpose: "Builds a knowledge graph of the code" },
  { name: "agent-memory", max: "one", purpose: "Memory store for AI agents" },
  { name: "forge", max: "many", purpose: "Pull request and issue templates of the code host" },
];

export const categoriesByName: ReadonlyMap<string, ToolCategory> = new Map(
  categories.map(category => [category.name, category]),
);
