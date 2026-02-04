import MiniSearch from "minisearch";
import { toast } from "svelte-sonner";

import { createStore } from "./createStore.svelte.ts";
import { getCurrentHost } from "./helpers/getCurrentHost.ts";
import { getMarkdownTitle } from "./helpers/getMarkdownTitle.ts";
import { markdownToText } from "./helpers/markdownToText.ts";
import { slugify } from "./helpers/slugify.ts";
import { wasmClient } from "./wasm/index.ts";
import type {
  ConstantDef,
  DocDef,
  EnumDef,
  IrSchema,
  PatternDef,
  ProcedureDef,
  RpcDef,
  StreamDef,
  TypeDef,
} from "./wasm/wasmtypes/types.ts";

export const primitiveTypes = ["string", "int", "float", "bool", "datetime"];

type SearchItemKind =
  | "constant"
  | "pattern"
  | "enum"
  | "type"
  | "rpc"
  | "proc"
  | "stream"
  | "doc";

type SearchItem = {
  id: number;
  kind: SearchItemKind;
  name: string;
  slug: string;
  doc: string;
};

export const miniSearch = new MiniSearch({
  fields: ["kind", "name", "doc"],
  storeFields: ["kind", "name", "doc", "slug"],
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
  vdlSchema: string;
  vdlSchemaExpanded: string;
  irSchema: IrSchema;
}

const fetchConfig = async () => {
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
      baseUrl: `${getCurrentHost()}/rpc`,
      headers: [],
    };
  }

  let baseUrl = "";
  let headers: Header[] = [];

  if (typeof config.baseUrl === "string" && config.baseUrl.trim() !== "") {
    baseUrl = config.baseUrl;
  } else {
    baseUrl = `${getCurrentHost()}/rpc`;
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
    vdlSchema: "",
    vdlSchemaExpanded: "",
    irSchema: {
      constants: [],
      enums: [],
      types: [],
      patterns: [],
      rpcs: [],
      procedures: [],
      streams: [],
      docs: [],
    },
  };
}

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
 * Fetches and loads an VDL schema from a specified URL.
 *
 * This function attempts to retrieve a schema from the given URL and, if successful,
 * updates the `vdlSchema` store with the fetched content. If the fetch fails,
 * an error is logged to the console.
 *
 * It also expands the types in the fetched schema and updates the `vdlSchemaExpanded` store.
 *
 * @param url The URL from which to fetch the VDL schema.
 * @throws Logs an error to the console if the fetch operation fails.
 */
export const loadVdlSchema = async (url: string) => {
  const response = await fetch(url);
  if (!response.ok) {
    console.error(`Failed to fetch schema from ${url}`);
    return;
  }
  const sch = await response.text();
  await loadVdlSchemaFromString(sch);
};

/**
 * Updates the `vdlSchema` store with a provided URPC schema string.
 *
 * This function directly sets the `vdlSchema` store to the provided schema string,
 * allowing for immediate updates to the schema without fetching from a URL.
 *
 * It also expands the types in the provided schema and updates the `vdlSchemaExpanded` store.
 *
 * @param vdlSchema The URPC schema string to be loaded into the store.
 */
const loadVdlSchemaFromString = async (vdlSchema: string) => {
  storeSettings.store.vdlSchema = vdlSchema;
  storeSettings.store.vdlSchemaExpanded =
    await wasmClient.expandTypes(vdlSchema);
  storeSettings.store.irSchema = await wasmClient.irgen({
    vdlSchema: vdlSchema,
  });
  await indexSearchItems();
};

/**
 * Indexes the search items for the current VDL IR schema.
 */
const indexSearchItems = async () => {
  type Node =
    | (ConstantDef & { kind: "constant" })
    | (PatternDef & { kind: "pattern" })
    | (EnumDef & { kind: "enum" })
    | (TypeDef & { kind: "type" })
    | (RpcDef & { kind: "rpc" })
    | (ProcedureDef & { kind: "proc" })
    | (StreamDef & { kind: "stream" })
    | (DocDef & { kind: "doc" });

  const addKind = <T, K extends SearchItemKind>(items: T[], kind: K) => {
    return items.map((item) => ({ ...item, kind }));
  };

  const nodes: Node[] = [
    ...addKind(storeSettings.store.irSchema.constants, "constant"),
    ...addKind(storeSettings.store.irSchema.patterns, "pattern"),
    ...addKind(storeSettings.store.irSchema.enums, "enum"),
    ...addKind(storeSettings.store.irSchema.types, "type"),
    ...addKind(storeSettings.store.irSchema.rpcs, "rpc"),
    ...addKind(storeSettings.store.irSchema.procedures, "proc"),
    ...addKind(storeSettings.store.irSchema.streams, "stream"),
  ];

  const searchItems = await Promise.all(
    nodes.map(async (node, index) => {
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
