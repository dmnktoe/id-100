#!/usr/bin/env node

/**
 * Build script for ID-100
 * - Bundles and minifies CSS and JS
 * - Adds content hashing for cache busting
 * - Generates manifest.json for asset paths
 */

const { execSync } = require("child_process");
const fs = require("fs");
const path = require("path");
const crypto = require("crypto");

const DIST_DIR = path.join(__dirname, "../web/static/dist");
const WEB_STATIC_DIR = path.join(__dirname, "../web/static");

// Ensure dist directory exists
if (!fs.existsSync(DIST_DIR)) {
  fs.mkdirSync(DIST_DIR, { recursive: true });
}

// Clean previous builds
const files = fs.readdirSync(DIST_DIR);
files.forEach((file) => {
  if (file.startsWith("main.") || file.startsWith("bundle.")) {
    fs.unlinkSync(path.join(DIST_DIR, file));
  }
});

console.log("🏗️  Building assets...\n");

// Build JS first (no CSS output expected now)
console.log("📦 Building JS...");
const isProduction = process.argv.includes("--production");
const esbuildCmd = isProduction
  ? "esbuild src/main.ts --bundle --minify --sourcemap --outfile=web/static/dist/main.tmp.js"
  : "esbuild src/main.ts --bundle --sourcemap --outfile=web/static/dist/main.tmp.js";

execSync(esbuildCmd, { stdio: "inherit" });

// Build CSS after JS (so it won't be overwritten)
console.log("📦 Building CSS...");
execSync("npx postcss src/styles.css -o web/static/dist/main.tmp.css", {
  stdio: "inherit",
});

// Generate content hashes
function getFileHash(filePath) {
  const content = fs.readFileSync(filePath);
  return crypto
    .createHash("sha256")
    .update(content)
    .digest("hex")
    .substring(0, 8);
}

// Move CSS sourcemap if exists and update sourceMappingURL before hashing
const cssSourceMapPath = path.join(DIST_DIR, "main.tmp.css.map");
let cssMapFileName = null;
if (fs.existsSync(cssSourceMapPath)) {
  const cssMapHash = getFileHash(cssSourceMapPath);
  cssMapFileName = `main.${cssMapHash}.css.map`;
  fs.renameSync(cssSourceMapPath, path.join(DIST_DIR, cssMapFileName));

  const cssTmpPath = path.join(DIST_DIR, "main.tmp.css");
  const cssContent = fs.readFileSync(cssTmpPath, "utf8");
  const updatedCssContent = cssContent.replace(
    /\/\*# sourceMappingURL=main\.tmp\.css\.map \*\//,
    `/*# sourceMappingURL=${cssMapFileName} */`,
  );
  fs.writeFileSync(cssTmpPath, updatedCssContent);
}

// Move JS sourcemap if exists and update sourceMappingURL before hashing
const sourceMapPath = path.join(DIST_DIR, "main.tmp.js.map");
let jsMapFileName = null;
if (fs.existsSync(sourceMapPath)) {
  const jsMapHash = getFileHash(sourceMapPath);
  jsMapFileName = `main.${jsMapHash}.js.map`;
  fs.renameSync(sourceMapPath, path.join(DIST_DIR, jsMapFileName));

  const jsTmpPath = path.join(DIST_DIR, "main.tmp.js");
  const jsContent = fs.readFileSync(jsTmpPath, "utf8");
  const updatedJsContent = jsContent.replace(
    /\/\/# sourceMappingURL=main\.tmp\.js\.map/,
    `//# sourceMappingURL=${jsMapFileName}`,
  );
  fs.writeFileSync(jsTmpPath, updatedJsContent);
}

const cssHash = getFileHash(path.join(DIST_DIR, "main.tmp.css"));
const jsHash = getFileHash(path.join(DIST_DIR, "main.tmp.js"));

// Rename files with hashes
const cssFileName = `main.${cssHash}.css`;
const jsFileName = `main.${jsHash}.js`;

fs.renameSync(
  path.join(DIST_DIR, "main.tmp.css"),
  path.join(DIST_DIR, cssFileName),
);

fs.renameSync(
  path.join(DIST_DIR, "main.tmp.js"),
  path.join(DIST_DIR, jsFileName),
);

// Generate manifest.json
const manifest = {
  "main.css": `/static/dist/${cssFileName}`,
  "main.js": `/static/dist/${jsFileName}`,
};

fs.writeFileSync(
  path.join(WEB_STATIC_DIR, "manifest.json"),
  JSON.stringify(manifest, null, 2),
);

console.log("\n✅ Build complete!");
console.log(`   CSS: ${cssFileName}`);
console.log(`   JS:  ${jsFileName}`);
console.log(`   Manifest: web/static/manifest.json\n`);
