import { readFileSync } from "node:fs";
import { defineConfig } from "tsdown";

export default defineConfig({
  entry: { bin: "src/bin.ts" },
  format: "esm",
  platform: "node",
  fixedExtension: true,
  banner: "#!/usr/bin/env node",
  plugins: [
    {
      // Embeds `<file>?raw` imports (the templates glob) as their text, as Vite does under Vitest.
      name: "raw",
      load: {
        filter: { id: /\?raw$/ },
        handler: (id) => `export default ${JSON.stringify(readFileSync(id.slice(0, -"?raw".length), "utf8"))};`,
      },
    },
  ],
});
