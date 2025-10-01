import MiniSearch from "minisearch";
import { toast } from "svelte-sonner";

import { createStore } from "./createStore.svelte.ts";
import { getCurrentHost } from "./helpers/getCurrentHost.ts";
import { getMarkdownTitle } from "./helpers/getMarkdownTitle.ts";
import { markdownToText } from "./helpers/markdownToText.ts";
import { slugify } from "./helpers/slugify.ts";
import { cmdExpandTypes, transpileUrpcToJson } from "./urpc.ts";
import type { Schema } from "./urpcTypes.ts";

export const primitiveTypes = ["string", "int", "float", "bool", "datetime"];

type SearchItem = {
  id: number;
  kind: "doc" | "type" | "proc" | "stream";
  name: string;
  slug: string;
  doc: string;
};

export const miniSearch = new MiniSearch({
  fields: ["kind", "name", "doc"],
  storeFields: ["kind", "name", "slug", "doc"],
  searchOptions: {
    boost: { title: 2 },
    fuzzy: 0.2,
    prefix: true,
  },
  tokenize: (text: string, _?: string): string[] => {
    const tokens: string[] = [];

    // First split by spaces
    const spaceTokens = text.split(" ");
    tokens.push(...spaceTokens);

    // Then split each space token by uppercase letters
    for (const token of spaceTokens) {
      const upperCaseTokens = token.split(/(?=[A-Z])/);
      tokens.push(...upperCaseTokens);
    }

    return tokens;
  },
});

export interface Header {
  key: string;
  value: string;
  enabled: boolean;
  description: string;
}

export interface StoreSettings {
  baseUrl: string;
  headers: Header[];
  urpcSchema: string;
  urpcSchemaExpanded: string;
  jsonSchema: Schema;
}

export const fetchConfig = async () => {
  // biome-ignore lint/suspicious/noExplicitAny: the fetch can return anything
  let config: any = {};

  try {
    const response = await fetch("./config.json");
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    config = await response.json();
  } catch (error) {
    console.error("Failed to fetch default config", error);
    return {
      baseUrl: `${getCurrentHost()}/api/v1/urpc`,
      headers: [],
    };
  }

  let baseUrl = "";
  let headers: Header[] = [];

  if (typeof config.baseUrl === "string" && config.baseUrl.trim() !== "") {
    baseUrl = config.baseUrl;
  } else {
    baseUrl = `${getCurrentHost()}/api/v1/urpc`;
  }

  if (Array.isArray(config.headers)) {
    headers = normalizeHeaders(config.headers);
  }

  return { baseUrl, headers };
};

async function storeSettingsGetInitialValue(): Promise<StoreSettings> {
  let baseUrl = "";
  let headers: Header[] = [];

  try {
    const config = await fetchConfig();
    baseUrl = config.baseUrl;
    headers = config.headers;
  } catch (error) {
    toast.error("Failed to load default config", {
      description: `Error: ${error}`,
    });
  }

  return {
    baseUrl,
    headers,
    urpcSchema: "version 1",
    urpcSchemaExpanded: "version 1",
    jsonSchema: { version: 1, nodes: [] },
  };
}

// Cannot use createStore because of the http request needed
// maybe it can be refactored later
export const storeSettings = createStore<StoreSettings>({
  initialValue: storeSettingsGetInitialValue,
  keysToPersist: ["baseUrl", "headers"],
  dbName: "storeSettings",
});

/**
 * Add or update a header in the store.headers array.
 *
 * @param key The key of the header to add or update.
 * @param value The value of the header to add or update.
 */
export const setHeader = (key: string, value: string) => {
  const trimmedKey = key.trim();
  const targetKeyLower = trimmedKey.toLowerCase();
  const existingIndex = storeSettings.store.headers.findIndex(
    (h) => h.key.trim().toLowerCase() === targetKeyLower,
  );

  if (existingIndex !== -1) {
    // Update existing header value and ensure it is enabled
    storeSettings.store.headers[existingIndex] = {
      ...storeSettings.store.headers[existingIndex],
      key: trimmedKey,
      value,
      enabled: true,
    };
    // Reassign to trigger reactivity in Svelte
    storeSettings.store.headers = [...storeSettings.store.headers];
  } else {
    // Add a new enabled header
    storeSettings.store.headers = [
      ...storeSettings.store.headers,
      { key: trimmedKey, value, enabled: true, description: "" },
    ];
  }
};

/**
 * Converts the headers array to a Headers object for use in fetch requests
 * it will set the "Content-Type" header to "application/json" by default
 * if that header is not present in the store.headers array.
 *
 * @returns A Headers object
 */
export const getHeadersObject = (): Headers => {
  const headers = new Headers();
  headers.set("Content-Type", "application/json");

  for (const header of storeSettings.store.headers) {
    if (!header.enabled) continue;
    if (header.key.trim()) headers.set(header.key, header.value);
  }

  return headers;
};

/**
 * Loads the default configuration from the static/config.json file.
 */
export const loadDefaultConfig = async () => {
  const config = await fetchConfig();
  storeSettings.store.baseUrl = config.baseUrl;
  storeSettings.store.headers = config.headers;
};

/**
 * Loads the default configuration from the static/config.json file.
 */
export const loadDefaultBaseURL = async () => {
  const config = await fetchConfig();
  storeSettings.store.baseUrl = config.baseUrl;
};

/**
 * Loads the default configuration from the static/config.json file.
 */
export const loadDefaultHeaders = async () => {
  const config = await fetchConfig();
  storeSettings.store.headers = config.headers;
};

type RawHeader = {
  key?: unknown;
  value?: unknown;
  enabled?: unknown;
  description?: unknown;
};

/**
 * Normalize an unknown headers payload (from localStorage or config.json)
 * into a strongly-typed array of Header with defaults of enabled=true and description="".
 */
const normalizeHeaders = (raw: unknown): Header[] => {
  if (!Array.isArray(raw)) return [];

  return (raw as RawHeader[]).map((item) => {
    const key = typeof item?.key === "string" ? item.key : "";
    const value = typeof item?.value === "string" ? item.value : "";
    const enabled = typeof item?.enabled === "boolean" ? item.enabled : true;
    const description =
      typeof item?.description === "string" ? item.description : "";
    return { key, value, enabled, description } satisfies Header;
  });
};

/**
 * Transpiles the current URPC schema to JSON format and updates the `jsonSchema` store.
 *
 * This asynchronous function takes the current value of `urpcSchema`, transpiles it to JSON
 * using the `transpileUrpcToJson` utility, and then updates the `jsonSchema` store with the result.
 */
export const loadJsonSchemaFromCurrentUrpcSchema = async () => {
  storeSettings.store.jsonSchema = await transpileUrpcToJson(
    storeSettings.store.urpcSchema,
  );
  await indexSearchItems();
};

/**
 * Indexes the search items for the current URPC JSON schema.
 */
const indexSearchItems = async () => {
  const searchItems = await Promise.all(
    storeSettings.store.jsonSchema.nodes.map(async (node, index) => {
      let name = "";
      let doc = "";

      if (node.kind === "doc") {
        name = getMarkdownTitle(node.content);
        doc = node.content;
      } else {
        name = node.name;
        doc = node.doc ?? "";
      }

      const item: SearchItem = {
        id: index,
        kind: node.kind,
        name,
        doc,
        slug: slugify(`${node.kind}-${name}`),
      };

      item.doc = await markdownToText(item.doc);

      return item;
    }),
  );

  miniSearch.removeAll();
  miniSearch.addAll(searchItems);
};

/**
 * Fetches and loads an URPC schema from a specified URL.
 *
 * This function attempts to retrieve a schema from the given URL and, if successful,
 * updates the `urpcSchema` store with the fetched content. If the fetch fails,
 * an error is logged to the console.
 *
 * It also expands the types in the fetched schema and updates the `urpcSchemaExpanded` store.
 *
 * @param url The URL from which to fetch the URPC schema.
 * @throws Logs an error to the console if the fetch operation fails.
 */
export const loadUrpcSchemaFromUrl = async (url: string) => {
  const response = await fetch(url);
  if (!response.ok) {
    console.error(`Failed to fetch schema from ${url}`);
    return;
  }

  const sch = await response.text();
  await loadUrpcSchemaFromString(sch);
};

/**
 * Updates the `urpcSchema` store with a provided URPC schema string.
 *
 * This function directly sets the `urpcSchema` store to the provided schema string,
 * allowing for immediate updates to the schema without fetching from a URL.
 *
 * It also expands the types in the provided schema and updates the `urpcSchemaExpanded` store.
 *
 * @param sch The URPC schema string to be loaded into the store.
 */
export const loadUrpcSchemaFromString = async (sch: string) => {
  storeSettings.store.urpcSchema = sch;
  storeSettings.store.urpcSchemaExpanded = await cmdExpandTypes(sch);
};

/**
 * Fetches an URPC schema from a URL, loads it, and then transpiles it to JSON.
 *
 * This function combines the operations of `loadUrpcSchemaFromUrl` and
 * `loadJsonSchemaFromCurrentUrpcSchema`. It first fetches and loads the URPC schema
 * from the specified URL, then transpiles that schema to JSON, updating both
 * the `urpcSchema` and `jsonSchema` stores in the process.
 *
 * @param url The URL from which to fetch the URPC schema.
 */
export const loadJsonSchemaFromUrpcSchemaUrl = async (url: string) => {
  await loadUrpcSchemaFromUrl(url);
  await loadJsonSchemaFromCurrentUrpcSchema();
};

/**
 * Loads an URPC schema from a string and transpiles it to JSON.
 *
 * This function takes an URPC schema as a string, loads it into the `urpcSchema` store,
 * and then transpiles it to JSON, updating both the `urpcSchema` and `jsonSchema` stores.
 * It's useful for processing schemas that are already available as strings without needing
 * to fetch from a URL.
 *
 * It also expands the types in the provided schema and updates the `urpcSchemaExpanded` store.
 *
 * @param sch The URPC schema string to load and transpile.
 */
export const loadJsonSchemaFromUrpcSchemaString = async (sch: string) => {
  await loadUrpcSchemaFromString(sch);
  await loadJsonSchemaFromCurrentUrpcSchema();
};
