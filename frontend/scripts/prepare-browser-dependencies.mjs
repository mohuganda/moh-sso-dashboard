/* global console */
// Test-only dependency mirror. Production artifacts and import maps stay unchanged.
import { mkdir, readFile, rm, writeFile } from "node:fs/promises";
import { createRequire } from "node:module";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { build } from "esbuild";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const require = createRequire(import.meta.url);
const directory = join(root, "node_modules/.cache/browser-dependencies");
const importMap = JSON.parse(await readFile(join(root, "dist/import-map.json"), "utf8"));
const dependencies = Object.entries(importMap.imports)
  .filter(([, url]) => url.startsWith("https://esm.sh/"))
  .map(([name]) => name);
const shared = new Set(dependencies);
await rm(directory, { recursive: true, force: true });
await mkdir(directory, { recursive: true });
const imports = {};
for (const [index, name] of dependencies.entries()) {
  const module = require(name);
  const names = Object.keys(module).filter((key) => key !== "__esModule");
  if ((name === "react" || name.startsWith("react-dom")) && !names.includes("default")) {
    names.push("default");
  }
  const filename = `dependency-${index}.js`;
  const result = await build({
    absWorkingDir: root,
    stdin: {
      contents: `export { ${names.join(", ")} } from ${JSON.stringify(name)};`,
      resolveDir: root,
    },
    outfile: join(directory, filename),
    write: false,
    metafile: true,
    bundle: true,
    format: "esm",
    platform: "browser",
    target: "es2022",
    minify: true,
    define: { "process.env.NODE_ENV": '"production"' },
    plugins: [
      {
        name: "shared-browser-instances",
        setup(builder) {
          builder.onResolve({ filter: /^[^./]/ }, ({ path }) =>
            path !== name && shared.has(path) ? { path, external: true } : undefined,
          );
        },
      },
    ],
  });
  // CommonJS peers (notably react-dom/client) use require even in an ESM
  // bundle. Bridge only their actual external requires to the same ESM instances.
  const requires = [
    ...new Set(
      Object.values(result.metafile.outputs)
        .flatMap((output) => output.imports)
        .filter((entry) => entry.external && entry.kind === "require-call")
        .map((entry) => entry.path),
    ),
  ];
  const bridge = requires.length
    ? requires
        .map((peer, index) => `import * as shared${index} from ${JSON.stringify(peer)};`)
        .join("\n") +
      `\nconst require = (name) => { switch(name) { ${requires
        .map(
          (peer, index) =>
            `case ${JSON.stringify(peer)}: return Object.assign({__esModule: true}, shared${index});`,
        )
        .join(" ")} default: throw new Error("Unexpected browser require: " + name); } };\n`
    : "";
  await writeFile(join(directory, filename), bridge + result.outputFiles[0].text);
  imports[name] = `/__browser_dependencies__/${filename}`;
}
await writeFile(join(directory, "import-map.json"), JSON.stringify({ imports }));
console.log(
  `Prepared ${dependencies.length} real browser dependencies from the installed lockfile.`,
);
