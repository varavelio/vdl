import { createStore } from "$lib/createStore.svelte";

type Input = Record<string, unknown>;
type Output = string;
type Date = string;
type HistoryInput = string;
type HistoryOutput = string;
type HistoryDate = string;

export interface HistoryItem {
  input: HistoryInput;
  output: HistoryOutput;
  date: HistoryDate;
}

export interface StoreNode {
  input: Input;
  output: Output;
  outputDate: Date;
  history: HistoryItem[];
}

export type StoreNodeInstance = ReturnType<typeof createStoreNode>;

type StoreNodeKey = keyof StoreNode;

const storeNodeDefault: StoreNode = {
  input: {},
  output: "",
  outputDate: "",
  history: [],
};

const storeNodeKeysToPersist: StoreNodeKey[] = ["input", "output", "outputDate", "history"];

export const createStoreNode = (nodeSlug: string) => {
  return createStore({
    initialValue: () => storeNodeDefault,
    keysToPersist: storeNodeKeysToPersist,
    dbName: "storeNode",
    tableName: nodeSlug,
    actions: (store, status) => {
      /**
       * Save the current input and output to the history.
       *
       * Limits the history to the most recent 150 entries.
       *
       * @param operationID the operation ID to save the current input and output for
       */
      function saveCurrentToHistory() {
        const historyLimit = 150;
        const input = store.input;
        const output = store.output;

        if (status.loading) return;
        if (!input && !output) return;

        store.history.unshift({
          input: JSON.stringify(input, null, 2),
          output: output,
          date: new Date().toISOString(),
        });

        while (store.history.length > historyLimit) {
          store.history.pop();
        }
      }

      /**
       * Delete a specific history item by its index.
       *
       * @param index The index of the history item to delete.
       */
      function deleteHistoryItem(index: number) {
        if (status.loading) return;
        if (index < 0 || index >= store.history.length) return;
        store.history.splice(index, 1);
      }

      /**
       * Clear the entire history.
       */
      function clearHistory() {
        if (status.loading) return;
        store.history = [];
      }

      /**
       * Clear the input and date fields.
       */
      function clearInput() {
        if (status.loading) return;
        store.input = {};
      }

      /**
       * Clear the output and date fields.
       */
      function clearOutput() {
        if (status.loading) return;
        store.output = "";
        store.outputDate = "";
      }

      /**
       * Clear the input and output fields.
       */
      function clearInputOutput() {
        clearInput();
        clearOutput();
      }

      return {
        saveCurrentToHistory,
        deleteHistoryItem,
        clearHistory,
        clearInput,
        clearOutput,
        clearInputOutput,
      };
    },
  });
};
