import { browser } from "$app/environment";
import localforage from "localforage";
import { debounce } from "lodash-es";
import { toast } from "svelte-sonner";

/**
 * The type of function that can be used as an action in the store.
 */
// biome-ignore lint/suspicious/noExplicitAny: the args and return types are dynamic and can't be typed here.
type ActionFunc = (...args: any[]) => any | Promise<any>;

/**
 * Options for configuring the store creation.
 */
interface CreateStoreOptions<
  T extends Record<string, unknown>,
  M extends Record<string, ActionFunc> = Record<string, ActionFunc>,
> {
  /**
   * A function that returns the initial value of the store.
   * @returns A promise that resolves to the initial value of the store.
   */
  initialValue: () => T | Promise<T>;
  /**
   * An array of keys from the store that should be persisted to IndexedDB.
   * Only top-level keys are supported.
   */
  keysToPersist: (keyof T)[];
  /**
   * An optional name for the database, used to create a unique isolated database instead of the global one.
   */
  dbName?: string;
  /**
   * An optional table name within the database to further isolate the store data.
   */
  tableName?: string;
  /**
   * An optional object containing additional actions to be added to the returned store.
   * The actions will receive the store and its status as arguments.
   */
  actions?: (store: T, status: StoreStatus) => M;
}

/**
 * Interface representing the status of the store.
 */
interface StoreStatus {
  /**
   * A function that returns a promise that resolves when the store is ready.
   */
  waitUntilReady: () => Promise<void>;
  /**
   * Indicates whether the store is ready for use (the opposite of loading).
   */
  ready: boolean;
  /**
   * Indicates whether the store is currently loading (the opposite of ready).
   */
  loading: boolean;
  /**
   * Indicates whether the store is currently saving data to the persistent storage.
   */
  saving: boolean;
}

/**
 * The result returned by the createStore function.
 */
interface StoreResult<
  T extends Record<string, unknown>,
  M extends Record<string, ActionFunc> = Record<string, ActionFunc>,
> {
  /**
   * The store object, read-write reactive to changes.
   */
  store: T;
  /**
   * The lifecycle status of the store, read-only reactive to changes.
   */
  status: StoreStatus;
  /**
   * Custom additional actions added to the store when it was created.
   */
  actions: M;
}

/**
 * Creates an asynchronous Svelte store that initializes its value from a
 * provided async function and persists specified keys to IndexedDB using
 * localforage with fallback to localStorage if IndexedDB is unavailable.
 *
 * Important: this function works only with JSON-serializable values.
 *
 * @template T - The type of the store object, which should be a record with string keys and unknown values.
 * @param opts - Configuration options for creating the async store.
 * @param opts.initialValue - A sync or async function that returns the initial value of the store.
 * @param opts.keysToPersist - An array of keys from the store that should be persisted to IndexedDB.
 * @param opts.dbName - An optional name for the store, used to create a unique isolated database instead of the global one.
 * @param opts.tableName - An optional table name within the database to further isolate the store data.
 * @param opts.actions - An optional object containing additional actions to be added to the returned store. The actions will receive the store and its status as arguments.
 * @returns An object containing the Svelte store and its lifecycle (status) state.
 *
 * @example
 * ```ts
 * const { store, status, actions } = createStore({
 *   initialValue: async () => ({ theme: 'light', fontSize: 14 }),
 *   keysToPersist: ['theme'],
 *   storeName: 'userPreferences',
 * });
 * ```
 */
export function createStore<
  // biome-ignore lint/suspicious/noExplicitAny: the values are dynamic and varied between different stores
  T extends Record<string, any>,
  M extends Record<string, ActionFunc> = Record<string, ActionFunc>,
>(opts: CreateStoreOptions<T, M>): StoreResult<T, M> {
  // Promise for waiting for initialization
  let readyPromiseResolve: () => void;
  const readyPromise = new Promise<void>((resolve) => {
    readyPromiseResolve = resolve;
  });

  // Initialize Svelte stores
  let store = $state<T>({} as T);
  const status = $state<StoreStatus>({
    waitUntilReady: () => readyPromise,
    ready: false,
    loading: true,
    saving: false,
  });

  // Asynchronously manage the store lifecycle
  (async () => {
    // Browser-only check
    if (!browser) {
      // biome-ignore lint/style/noNonNullAssertion: the function is always defined above
      readyPromiseResolve!();
      status.ready = true;
      status.loading = false;
      return;
    }

    // Create the localforage database name, it' will be used to isolate
    // different stores between themselves
    let dbName = createGlobalDbNamePrefix();
    if (opts.dbName && opts.dbName.trim() !== "") {
      dbName += `-${opts.dbName.trim()}`;
    }

    // Create localforage database instance
    // https://localforage.github.io/localForage/#multiple-instances-createinstance
    // https://localforage.github.io/localForage/#settings-api-config
    const db = localforage.createInstance({
      driver: localforage.INDEXEDDB,
      name: dbName ?? "defaultDb",
      storeName: opts.tableName ?? "defaultTable",
    });

    // Load the initial store value
    try {
      const initialValue = await opts.initialValue();
      Object.assign(store, initialValue);
    } catch (error) {
      toast.error("Failed to load initial store value", {
        description: `Error: ${error}`,
      });
    }

    // Load persisted values from the database
    try {
      const promises = opts.keysToPersist.map(async (keyToPersist) => {
        const value = await db.getItem(keyToPersist as string);
        if (value === null) return;
        (store[keyToPersist] as unknown) = value;
      });
      await Promise.all(promises);
    } catch (error) {
      toast.error(`Failed to load persisted store values from ${dbName}`, {
        description: `Error: ${error}`,
      });
    }

    // Create map with the debounced persist functions for each key
    const persistDebouncedMap = new Map<string, (value: unknown) => void>();
    for (const keyToPersist of opts.keysToPersist) {
      const persistFn = async (value: unknown) => {
        status.saving = true;

        try {
          // Delete null or undefined values from the database
          if (value === null || value === undefined) {
            await db.removeItem(keyToPersist as string);
          } else {
            await db.setItem(keyToPersist as string, value);
          }
        } catch (error) {
          // On error, remove the item to avoid stale or corrupted data
          db.removeItem(keyToPersist as string);

          // Show error toast and log to console for debugging
          toast.error(
            `Failed to persist ${keyToPersist as string} value to the database ${dbName}`,
            {
              description: `Error: ${error}`,
            },
          );
          console.error(
            `Failed to persist ${keyToPersist as string} value to the database ${dbName}`,
            error,
          );
          console.log({ value });
        } finally {
          status.saving = false;
        }
      };

      const delayMs = 300;
      const persistFnDebounced = debounce(persistFn, delayMs);
      persistDebouncedMap.set(keyToPersist as string, persistFnDebounced);
    }

    // Create an $effect to persist changes to the database
    $effect.root(() => {
      for (const keyToPersist of opts.keysToPersist) {
        $effect(() => {
          const persistFn = persistDebouncedMap.get(keyToPersist as string);
          if (!persistFn) return;

          // We must unwrap the Svelte Proxy before saving because it can't be easily saved as-is.
          // The JSON cycle is a simple way to do this, but limits us to JSON-serializable data.
          //
          // This also force svelte to track changes for complex or nested object values
          // by using JSON.stringify that involves reading all properties
          const value = JSON.parse(JSON.stringify(store[keyToPersist]));
          persistFn(value);
        });
      }
    });

    // biome-ignore lint/style/noNonNullAssertion: the function is always defined above
    readyPromiseResolve!();
    status.ready = true;
    status.loading = false;
  })();

  // Create actions if provided
  const actions = opts.actions ? opts.actions(store, status) : ({} as M);

  return {
    store,
    status,
    actions: actions,
  };
}

/**
 * Creates a prefix string based on the current URL path. This prefix allows
 * to use the same database names across different deployments under the
 * same domain, avoiding collisions.
 *
 * @returns A prefix string based on the current URL path, suitable for use in database names.
 */
export function createGlobalDbNamePrefix(): string {
  if (!browser) return "";

  const prefix = globalThis.location.pathname
    .replace(/[^a-z0-9]/gi, "-")
    .toLowerCase();

  return prefix;
}
