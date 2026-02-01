import fs from "node:fs";
import path from "node:path";

/**
 * Script to replace all absolute resource paths (e.g., "/icon.png", "/_app/...")
 * with relative paths ("./icon.png", "./_app/...") throughout the build output.
 *
 * It solves the following SvelteKit issue:
 * https://github.com/sveltejs/kit/issues/9569
 *
 * It works in two phases:
 * 1. Scans the 'build' directory to create a list of all existing file paths.
 * 2. Iterates over each text file (.html, .js, .css) and replaces occurrences
 *    of those file paths with their relative equivalent.
 */

/**
 * Read a directory recursively and return all file paths
 *
 * @param filePath The directory path to read
 * @returns A list of all file paths (absolute) under the given directory, recursively
 */
const readDirRecursive = async (filePath: string): Promise<string[]> => {
  const dir = await fs.promises.readdir(filePath);
  const files = await Promise.all(
    dir.map(async (relativePath) => {
      const absolutePath = path.join(filePath, relativePath);
      const stat = await fs.promises.lstat(absolutePath);
      return stat.isDirectory() ? readDirRecursive(absolutePath) : absolutePath;
    }),
  );
  return files.flat();
};

/**
 * Main patching function
 */
const patchFiles = async () => {
  try {
    const buildDir = path.resolve("./build");
    console.log(`Starting patch process in: ${buildDir}`);

    // 1. Scan the build directory to get all file paths
    const allFilePaths = await readDirRecursive(buildDir);
    const targetPaths = allFilePaths
      .map((filePath) => {
        return path.relative(buildDir, filePath).replace(/\\/g, "/");
      })
      .map((p) => `/${p}`);
    console.log(`Found ${targetPaths.length} potential asset paths to patch`);

    // 2. Iterate over each text file and replace absolute paths with relative ones
    for (const file of allFilePaths) {
      const hasDesiredExtension =
        file.endsWith(".js") ||
        file.endsWith(".html") ||
        file.endsWith(".map") ||
        file.endsWith(".css");

      if (!hasDesiredExtension) {
        continue;
      }

      let data = await fs.promises.readFile(file, "utf8");
      let patched = false;

      for (const absolutePath of targetPaths) {
        const relativePath = `.${absolutePath}`;

        // Replace occurrences with double quotes: "/icon.png" -> "./icon.png"
        const doubleQuoteSearch = `"${absolutePath}"`;
        if (data.includes(doubleQuoteSearch)) {
          data = data.replaceAll(doubleQuoteSearch, `"${relativePath}"`);
          patched = true;
        }

        // Replace occurrences with single quotes: '/icon.png' -> './icon.png'
        const singleQuoteSearch = `'${absolutePath}'`;
        if (data.includes(singleQuoteSearch)) {
          data = data.replaceAll(singleQuoteSearch, `'${relativePath}'`);
          patched = true;
        }

        // Replace occurrences with backticks: `/icon.png` -> `./icon.png`
        const backtickSearch = `\`${absolutePath}\``;
        if (data.includes(backtickSearch)) {
          data = data.replaceAll(backtickSearch, `\`${relativePath}\``);
          patched = true;
        }
      }

      if (patched) {
        await fs.promises.writeFile(file, data, "utf8");
        console.log(
          `Patched relative assets in: ${path.relative(process.cwd(), file)}`,
        );
      }
    }
  } catch (error) {
    console.error("An error occurred during the patching process:", error);
    process.exit(1);
  }
};

patchFiles();
