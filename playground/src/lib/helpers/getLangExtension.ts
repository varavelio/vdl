/**
 * Returns the file extension associated with a given programming language.
 * If the language is not recognized, returns the provided fallback extension.
 *
 * @param {string} lang - The language identifier (e.g., "json", "javascript").
 * @param {string} [fallback="txt"] - The fallback extension to use if the language is not recognized.
 * @returns {string} The file extension corresponding to the language, or the fallback if not found.
 */
export function getLangExtension(lang: string, fallback = "txt"): string {
  const extensions: Record<string, string> = {
    vdl: "vdl",
    json: "json",
    yaml: "yaml",
    yml: "yaml",
    bash: "sh",
    shell: "sh",
    sh: "sh",
    c: "c",
    csharp: "cs",
    cs: "cs",
    clojure: "clj",
    dart: "dart",
    elixir: "ex",
    go: "go",
    http: "http",
    java: "java",
    javascript: "js",
    js: "js",
    typescript: "ts",
    ts: "ts",
    julia: "jl",
    kotlin: "kt",
    lua: "lua",
    matlab: "m",
    objectivec: "m", // Objective-C
    objc: "m",
    ocaml: "ml",
    perl: "pl",
    php: "php",
    powershell: "ps1",
    ps1: "ps1",
    python: "py",
    py: "py",
    r: "r",
    ruby: "rb",
    rb: "rb",
    rust: "rs",
    swift: "swift",
    wget: "sh", // Wget is a bash command, so .sh
    ansible: "yaml", // Ansible uses YAML
  };

  let extension = fallback;

  const normalizedLang = lang.trim().toLowerCase();
  if (normalizedLang in extensions) {
    extension = extensions[normalizedLang];
  }

  return extension;
}
