<script lang="ts">
  import { History, Trash, X } from "@lucide/svelte";
  import { toast } from "svelte-sonner";
  import BottomSpace from "$lib/components/BottomSpace.svelte";
  import Offcanvas from "$lib/components/Offcanvas.svelte";
  import { formatISODate } from "$lib/helpers/formatISODate";

  import type { StoreNodeInstance } from "../../../storeNode.svelte";

  import HistoryModal from "./HistoryModal.svelte";

  interface Props {
    isOpen: boolean;
    storeNode: StoreNodeInstance;
  }

  let { isOpen = $bindable(), storeNode }: Props = $props();

  let selectedHistoryItem: any = $state(null);
  let isModalOpen = $state(false);
  let historyLength = $derived(storeNode.store.history.length);
  let hasHistory = $derived(historyLength > 0);

  function viewHistoryItem(item: any) {
    selectedHistoryItem = item;
    isModalOpen = true;
  }

  function isConfirmed() {
    return confirm("Are you sure you want to continue? This action cannot be undone.");
  }

  function deleteHistoryItem(index: number, event: Event) {
    event.stopPropagation();
    if (!isConfirmed()) return;
    storeNode.actions.deleteHistoryItem(index);
    toast.success("History item deleted");
  }

  function clearAllHistory() {
    if (!isConfirmed()) return;
    storeNode.actions.clearHistory();
    toast.success("History cleared");
  }
</script>

<Offcanvas bind:isOpen direction="right" class="desk:w-[400px]">
  <div class="flex h-full flex-col">
    <div class="border-base-content/20 flex items-center justify-between border-b p-4">
      <h2 class="text-lg font-semibold">Request History ({historyLength})</h2>
      <button
        class="btn btn-ghost btn-sm btn-circle"
        onclick={() => (isOpen = false)}
        aria-label="Close"
      >
        <X class="size-4" />
      </button>
    </div>

    <div class="border-base-content/20 border-b p-4">
      <button
        class="btn btn-warning btn-sm w-full"
        onclick={clearAllHistory}
        disabled={!hasHistory}
      >
        <Trash class="size-4" />
        Clear all history
      </button>
    </div>

    <div class="flex-1 overflow-y-auto">
      {#if !hasHistory}
        <div class="flex items-center justify-center p-8 text-center">
          <div class="text-base-content/60">
            <History class="mx-auto mb-2 size-8" />
            <p>No history available</p>
          </div>
        </div>
      {:else}
        <div class="space-y-2 p-4">
          {#each storeNode.store.history as item, index}
            <div
              class={[
                "rounded-box border-base-content/20 bg-base-200 hover:bg-base-300 rounded border shadow-sm transition-colors",
                "flex items-center justify-between",
              ]}
            >
              <button
                class="flex-1 cursor-pointer p-4 text-left"
                onclick={() => viewHistoryItem(item)}
                type="button"
              >
                <p class="text-sm font-medium">{formatISODate(item.date)}</p>
                <p class="text-base-content/60 mt-1 truncate text-xs">Click to view details</p>
              </button>
              <button
                class="btn btn-warning btn-square btn-xs mr-4 ml-2"
                onclick={(e) => deleteHistoryItem(index, e)}
                type="button"
              >
                <Trash class="size-3" />
              </button>
            </div>
          {/each}
        </div>
      {/if}

      <BottomSpace />
    </div>
  </div>
</Offcanvas>

<HistoryModal bind:isOpen={isModalOpen} historyItem={selectedHistoryItem} />
