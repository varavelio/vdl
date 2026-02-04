/**
 * version is replaced during the release process by the latest Git tag
 * and should not be manually edited. Should not include the v prefix.
 */
const version = "0.0.0-dev";

/**
 * Takes the playground version, cleans it and returns it
 *
 * @returns cleaned playground version without v prefix
 */
export function getVersion(): string {
  let ver = version.toLowerCase().trim();
  if (!ver.startsWith("v")) return ver;
  return ver.slice(1);
}
