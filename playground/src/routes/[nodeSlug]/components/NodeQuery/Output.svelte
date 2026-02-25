<script lang="ts">
  import { CircleX, Clock, CloudAlert, Copy, Loader, Trash } from "@lucide/svelte";
  import Editor from "$lib/components/Editor.svelte";
  import H3 from "$lib/components/H3.svelte";
  import Tabs from "$lib/components/Tabs.svelte";
  import { copyTextToClipboard } from "$lib/helpers/copyTextToClipboard";
  import { formatISODate } from "$lib/helpers/formatISODate";

  import type { StoreNodeInstance } from "../../storeNode.svelte";

  import OutputQuickActions from "./OutputQuickActions.svelte";

  interface Props {
    type: "stream" | "proc";
    cancelRequest: () => void;
    isExecuting: boolean;
    storeNode: StoreNodeInstance;
  }

  let { cancelRequest, isExecuting, type, storeNode }: Props = $props();
  let hasOutput = $derived.by(() => {
    if (!storeNode.store.output) return false;
    if (storeNode.store.output === "{}") return false;
    if (storeNode.store.output === "[]") return false;
    return true;
  });

  /**
   * Formats the output date to be more human-readable.
   *
   * The date is in ISO 8601 format (e.g., "2000-10-05T14:48:00.000Z").
   *
   * And the result will be like "2000-10-05 14:48:00"
   */
  let prettyOutputDate = $derived.by(() => {
    if (!storeNode.store.outputDate) return "unknown output date";
    return formatISODate(storeNode.store.outputDate);
  });

  async function copyToClipboard() {
    return copyTextToClipboard(storeNode.store.output);
  }
</script>

{#snippet CancelButton()}
  {#if isExecuting}
    <button class="btn btn-error" onclick={cancelRequest}>
      <CircleX class="size-4" />
      {type === "proc" ? "Cancel procedure call" : "Stop stream"}
    </button>
  {/if}
{/snippet}

<div class="w-full space-y-2">
  {#if !hasOutput && !isExecuting}
    <div class="mt-[100px] flex w-full flex-col items-center justify-center space-y-2">
      <CloudAlert class="size-10" />
      <H3 class="flex items-center justify-start space-x-2">No Output</H3>
    </div>

    <p class="mb-[100px] pt-4 text-center">
      Please execute the {type === "proc" ? "procedure" : "stream"} from the input tab to see the
      output.
    </p>
  {/if}

  {#if !hasOutput && isExecuting}
    <div class="mt-12 mb-4 flex w-full flex-col items-center justify-center space-y-2">
      <Loader class="animate size-10 animate-spin" />
      <H3 class="flex items-center justify-start space-x-2">
        {type === "proc" ? "Executing procedure" : "Starting data stream"}
      </H3>
    </div>

    <div class="flex justify-center">{@render CancelButton()}</div>
  {/if}

  {#if hasOutput}
    {#if type == "stream" && isExecuting}
      <p class="pb-2 text-sm">
        The data stream is currently active using Server Sent Events (SSE). You can stop it by
        clicking the button below. New messages will be added to the top of the output.
      </p>

      <div class="pb-2">{@render CancelButton()}</div>
    {/if}

    <div class="flex w-full flex-wrap items-start justify-between gap-2">
      <div>
        {#if type == "proc"}
          <OutputQuickActions output={storeNode.store.output} />
        {/if}
      </div>

      <div class="flex justify-end">
        <Tabs
          buttonClass="btn-xs"
          iconClass="size-3"
          items={[
            {
              id: "outputDate",
              icon: Clock,
              action: () => {},
              label: prettyOutputDate,
              tooltipText: "Latest output date",
            },
            {
              id: "copy",
              icon: Copy,
              action: copyToClipboard,
              tooltipText: "Copy output to clipboard",
            },
            {
              id: "clear",
              icon: Trash,
              action: storeNode.actions.clearOutput,
              tooltipText: "Clear output",
            },
          ]}
        />
      </div>
    </div>

    <Editor
      class="rounded-box h-[600px] w-full overflow-hidden shadow-md"
      lang="json"
      value={storeNode.store.output ?? ""}
    />
  {/if}
</div>
