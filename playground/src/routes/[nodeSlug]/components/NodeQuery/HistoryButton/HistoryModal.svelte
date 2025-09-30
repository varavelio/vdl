<script lang="ts">
  import { MoveDownLeft, MoveUpRight, X } from "@lucide/svelte";

  import { formatISODate } from "$lib/helpers/formatISODate";
  import { storeUi } from "$lib/storeUi.svelte";

  import Code from "$lib/components/Code.svelte";
  import Modal from "$lib/components/Modal.svelte";
  import Tabs from "$lib/components/Tabs.svelte";

  import type { HistoryItem } from "../../../storeNode.svelte";

  interface Props {
    isOpen: boolean;
    historyItem: HistoryItem | null;
  }

  let { isOpen = $bindable(), historyItem }: Props = $props();
</script>

<Modal bind:isOpen class="h-full w-full max-w-4xl">
  {#if historyItem}
    <div class="flex h-full flex-col space-y-4 overflow-hidden">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold">
          {formatISODate(historyItem.date)}
        </h2>
        <button
          class="btn btn-ghost btn-sm btn-circle"
          onclick={() => (isOpen = false)}
          aria-label="Close"
        >
          <X class="size-4" />
        </button>
      </div>

      <div>
        <Tabs
          items={[
            { id: "input", label: "Input", icon: MoveUpRight },
            { id: "output", label: "Output", icon: MoveDownLeft },
          ]}
          bind:active={storeUi.store.historyTab}
        />
      </div>

      <div class="flex-1 flex-grow overflow-hidden">
        {#if storeUi.store.historyTab === "input"}
          <Code lang="json" code={historyItem.input} />
        {/if}

        {#if storeUi.store.historyTab === "output"}
          <Code lang="json" code={historyItem.output} />
        {/if}
      </div>
    </div>
  {/if}
</Modal>
