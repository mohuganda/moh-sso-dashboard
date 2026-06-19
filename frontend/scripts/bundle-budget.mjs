/* global process */
import { existsSync, readdirSync, statSync } from "node:fs";
import { join } from "node:path";
import { finish, root } from "./workspace-lib.mjs";

const failures = [];
const budgets = {
  shellEntryJs: Number(process.env.BUNDLE_BUDGET_SHELL_JS_KB ?? 8500),
  shellCss: Number(process.env.BUNDLE_BUDGET_SHELL_CSS_KB ?? 1200),
  appChunkJs: Number(process.env.BUNDLE_BUDGET_APP_CHUNK_JS_KB ?? 10000),
};

function sizeKb(path) {
  return statSync(path).size / 1024;
}

function assertBudget(label, path, maxKb) {
  const actualKb = sizeKb(path);
  if (actualKb > maxKb) {
    failures.push(`${label} is ${actualKb.toFixed(1)} KiB, above ${maxKb} KiB: ${path}`);
  }
}

const assetsDir = join(root, "dist", "assets");
if (!existsSync(assetsDir)) {
  failures.push("dist/assets is missing; run build:docker first");
} else {
  for (const entry of readdirSync(assetsDir)) {
    const path = join(assetsDir, entry);
    if (/^index-.+\.js$/.test(entry)) {
      assertBudget("Shell entry JS", path, budgets.shellEntryJs);
    }
    if (/^index-.+\.css$/.test(entry)) {
      assertBudget("Shell CSS", path, budgets.shellCss);
    }
  }
}

const mfDir = join(root, "dist", "mf");
if (!existsSync(mfDir)) {
  failures.push("dist/mf is missing; run build:docker first");
} else {
  for (const app of readdirSync(mfDir)) {
    const appDir = join(mfDir, app);
    if (!statSync(appDir).isDirectory()) {
      continue;
    }

    for (const entry of readdirSync(appDir)) {
      if (entry.endsWith(".js")) {
        assertBudget(`${app} app chunk`, join(appDir, entry), budgets.appChunkJs);
      }
    }
  }
}

finish("Bundle budget check", failures);
