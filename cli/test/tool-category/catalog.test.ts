import { describe, expect, it } from "@effect/vitest";
import { tools } from "@/tool/catalog";
import { categories, categoriesByName } from "@/tool-category/catalog";

/** The Catalog (1.0) table of docs/blueprint/entities/tool-category/index.md, in display order. */
const expected = [
  ["tool-manager", "one", "Installs the tools and runs the task library"],
  ["version-control", "one", "Ignore rules and git settings"],
  ["git-hooks", "one", "Runs the gates before each commit"],
  ["editor", "many", "Editor defaults for the repository"],
  ["formatter", "many", "Formats the repository's files"],
  ["linter", "many", "Finds errors in code"],
  ["secret-scanner", "one", "Finds secrets in code and commits"],
  ["vulnerability-scanner", "many", "Finds known vulnerabilities in dependencies"],
  ["secrets-manager", "one", "Gives the repository's secrets to its tasks"],
  ["ai-agent", "many", "Settings for an AI coding agent"],
  ["knowledge-graph", "one", "Builds a knowledge graph of the code"],
  ["agent-memory", "one", "Memory store for AI agents"],
  ["forge", "many", "Pull request and issue templates of the code host"],
] as const;

describe("tool category catalog", () => {
  it("holds the 13 categories in display order with their max and purpose", () => {
    expect(categories.map(c => [c.name, c.max, c.purpose])).toEqual(expected);
  });

  it("looks a category up by name", () => {
    for (const category of categories) {
      expect(categoriesByName.get(category.name)).toBe(category);
    }
    expect(categoriesByName.get("nope")).toBeUndefined();
  });

  it("holds at least one tool in every category (invariant 6)", () => {
    for (const category of categories) {
      expect(tools.some(tool => tool.category === category.name), category.name)
        .toBe(true);
    }
  });

  it("holds at most one default-on and one unremovable tool per max-one category (invariant 5)", () => {
    for (const category of categories.filter(c => c.max === "one")) {
      const own = tools.filter(tool => tool.category === category.name);
      expect(own.filter(tool => tool.default === "on").length, category.name)
        .toBeLessThanOrEqual(1);
      expect(own.filter(tool => !tool.removable).length, category.name)
        .toBeLessThanOrEqual(1);
    }
  });
});
