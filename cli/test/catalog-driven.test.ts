import { describe, expect, it } from "@effect/vitest";
import { readdirSync, readFileSync } from "node:fs";
import { join, relative } from "node:path";
import { tools } from "@/tool/catalog";
import { categories } from "@/tool-category/catalog";

/** Row 13 of the plan: catalog-driven code holds no tool or category name as a literal. */
const src = join(import.meta.dirname, "..", "src");
const folders = ["tool", "tool-category", "setup-config"];
const catalogs = new Set(["tool/catalog.ts", "tool-category/catalog.ts"]);
const names = new Set([...tools.map(tool => tool.name), ...categories.map(category => category.name)]);

/** Every quoted string literal: single, double or backtick quotes. */
const literal = /(["'`])((?:\\.|(?!\1)[^\\])*)\1/g;

const scanned = folders.flatMap(folder =>
  readdirSync(join(src, folder), { recursive: true, encoding: "utf8" })
    .filter(file => file.endsWith(".ts"))
    .map(file => relative(src, join(src, folder, file)))
    .filter(file => !catalogs.has(file))
);

describe("catalog-driven code", () => {
  it("scans the source of every catalog-driven folder", () => {
    expect(scanned.length).toBeGreaterThan(0);
  });

  it("holds no tool name or category name as a quoted literal outside the catalogs", () => {
    const hits = scanned.flatMap(file =>
      [...readFileSync(join(src, file), "utf8").matchAll(literal)]
        .filter(match => names.has(match[2]!))
        .map(match => `${file}: ${match[0]}`)
    );
    expect(hits).toEqual([]);
  });
});
