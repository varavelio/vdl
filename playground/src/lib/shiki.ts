import { createHighlighter } from "shiki";
import type { BundledLanguage } from "shiki";
import type { BundledTheme, Highlighter } from "shiki";

// Languages that are used in the playground snippets
const langs: BundledLanguage[] = [
  "bash",
  "c",
  "clojure",
  "csharp",
  "dart",
  "elixir",
  "go",
  "http",
  "java",
  "javascript",
  "json",
  "julia",
  "kotlin",
  "lua",
  "matlab",
  "objective-c",
  "ocaml",
  "perl",
  "php",
  "powershell",
  "python",
  "r",
  "ruby",
  "rust",
  "swift",
  "typescript",
  "ts",
  "yaml",
];

/**
 * Returns the provided language if it's supported, otherwise falls back to plain text.
 * @param lang - The language identifier to check
 * @returns The original language if supported, or "text" as fallback
 */
export const getOrFallbackLanguage = (lang: string): string => {
  const supportedLangs = ["vdl", ...langs];
  if (supportedLangs.includes(lang)) return lang;
  return "text"; // https://shiki.matsu.io/languages#plain-text
};

export const vdlSyntaxUrl = "./_app/vdl/vscode/vdl.tmLanguage.json";
export const lightTheme: BundledTheme = "github-light";
export const darkTheme: BundledTheme = "github-dark";

let highlighterInstance: Highlighter | null = null;
let highlighterPromise: Promise<Highlighter> | null = null;

/**
 * Returns a Shiki highlighter instance with VDL and bundled languages support.
 * This function implements a singleton pattern, returning an existing instance
 * or promise if available, otherwise creating a new highlighter.
 *
 * The highlighter is configured with both light and dark GitHub themes,
 * and includes VDL syntax highlighting loaded from a remote source in
 * addition to the bundled languages.
 *
 * @returns {Promise<Highlighter>} A promise that resolves to a Shiki highlighter instance
 */
export const getHighlighter = async (): Promise<Highlighter> => {
  if (highlighterInstance) return highlighterInstance;
  if (highlighterPromise) return highlighterPromise;

  highlighterPromise = (async () => {
    const vdlSyntax = await fetch(vdlSyntaxUrl);
    const vdlSyntaxJson = await vdlSyntax.json();
    vdlSyntaxJson.name = "vdl";

    highlighterInstance = await createHighlighter({
      langs: [vdlSyntaxJson, ...langs],
      themes: [lightTheme, darkTheme],
    });

    return highlighterInstance;
  })();

  return highlighterPromise;
};

/**
 * Preloads the Shiki highlighter during application startup.
 *
 * This function is intended to be called during the app's splash screen
 * to prevent lag when syntax highlighting is first needed. Since loading
 * the highlighter and all language definitions is resource-intensive,
 * initializing it early improves the user experience once the app is running.
 */
export const initializeShiki = async () => {
  await getHighlighter();
};
