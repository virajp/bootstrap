import { defineConfig } from "tsdown";

export default defineConfig({
  entry: { bin: "src/bin.ts" },
  format: "esm",
  platform: "node",
  fixedExtension: true,
  banner: "#!/usr/bin/env node",
});
