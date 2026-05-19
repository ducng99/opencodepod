import { build } from "bun";
import { copyFileSync, mkdirSync } from "node:fs";
import { $ } from "bun";

const watch = process.argv.includes("--watch");

async function doBuild() {
  mkdirSync("dist", { recursive: true });

  // Build JS/TSX with Bun
  await build({
    entrypoints: ["./src/main.tsx"],
    outdir: "./dist",
    target: "browser",
    minify: !watch,
    sourcemap: watch ? "inline" : "none",
  });

  // Build CSS with Tailwind v4
  const twArgs = ["-i", "./src/index.css", "-o", "./dist/index.css"];
  if (!watch) twArgs.push("--minify");

  const proc = Bun.spawn({
    cmd: ["bunx", "tailwindcss", ...twArgs],
    stdout: "inherit",
    stderr: "inherit",
    env: { ...process.env },
  });
  await proc.exited;
  if (proc.exitCode !== 0) {
    process.exit(proc.exitCode ?? 1);
  }

  // Copy HTML shell
  copyFileSync("index.html", "dist/index.html");

  console.log("Build complete!");
}

if (watch) {
  console.log("Watching for changes... (not implemented, run manually)");
  await doBuild();
} else {
  await doBuild();
}
