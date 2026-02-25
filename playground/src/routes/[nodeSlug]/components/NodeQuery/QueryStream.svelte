<script lang="ts">
  import { Info, Loader, MoveDownLeft, MoveUpRight, Trash, Zap } from "@lucide/svelte";
  import { toast } from "svelte-sonner";
  import H2 from "$lib/components/H2.svelte";
  import Menu from "$lib/components/Menu.svelte";
  import Tabs from "$lib/components/Tabs.svelte";
  import { ctrlSymbol } from "$lib/helpers/ctrlSymbol";
  import { joinPath } from "$lib/helpers/joinPath";
  import { getHeadersObject, type StreamDef, storeSettings } from "$lib/storeSettings.svelte";
  import { storeUi } from "$lib/storeUi.svelte";

  import type { StoreNodeInstance } from "../../storeNode.svelte";

  import HistoryButton from "./HistoryButton/HistoryButton.svelte";
  import InputForm from "./InputForm/InputForm.svelte";
  import Output from "./Output.svelte";
  import Snippets from "./Snippets/Snippets.svelte";

  interface Props {
    stream: StreamDef & { kind: "stream" };
    storeNode: StoreNodeInstance;
  }

  let { stream, storeNode = $bindable() }: Props = $props();

  let outputArray: unknown[] = $state([]);
  let isExecuting = $state(false);
  let cancelRequest = $state<() => void>(() => {});

  $effect(() => {
    if (outputArray.length === 0) {
      storeNode.actions.clearOutput();
    } else {
      storeNode.store.output = JSON.stringify(outputArray, null, 2);
      storeNode.store.outputDate = new Date().toISOString();
    }
  });

  async function executeStream() {
    if (isExecuting) return;
    isExecuting = true;
    outputArray = [];

    try {
      openOutputTab(true);
      const controller = new AbortController();
      const signal = controller.signal;

      cancelRequest = () => {
        controller.abort();
        toast.info("Stream stopped");
      };

      const endpoint = joinPath([storeSettings.store.baseUrl, stream.rpcName, stream.name]);
      const headers = getHeadersObject();
      headers.set("Accept", "text/event-stream");
      headers.set("Cache-Control", "no-cache");

      const response = await fetch(endpoint, {
        method: "POST",
        body: JSON.stringify(storeNode.store.input ?? {}),
        headers,
        signal: signal,
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const reader = response.body?.getReader();
      if (!reader) {
        throw new Error("Response body is null");
      }
      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) {
          toast.info("Stream ended by server");
          break;
        }

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");

        buffer = lines.pop() || "";

        for (const line of lines) {
          if (line.trim() === "") continue;

          if (line.startsWith("data: ")) {
            const eventData = line.slice(6);

            if (eventData.trim() === "" || eventData.trim() === "heartbeat") {
              continue;
            }

            try {
              const parsedData = JSON.parse(eventData);
              outputArray.unshift(parsedData);
            } catch (parseError) {
              outputArray.unshift(eventData);
            }
          }
        }
      }
    } catch (error: unknown) {
      if (!(error instanceof Error && error.name === "AbortError")) {
        console.error(error);
        toast.error("Failed to send HTTP request", {
          description: `Error: ${error}`,
          duration: 5000,
        });
      }
    } finally {
      isExecuting = false;
      cancelRequest = () => {};
      if (outputArray.length > 0) {
        storeNode.actions.saveCurrentToHistory();
      }
    }
  }

  async function executeStreamFromKbd(event: KeyboardEvent) {
    if (event.key === "Enter" && (event.ctrlKey || event.metaKey)) {
      event.preventDefault();
      await executeStream();
    }
  }

  let tab: "input" | "output" = $state("input");
  let wrapper: HTMLDivElement | null = $state(null);
  function openOutputTab(scroll = false) {
    if (tab === "output") return;
    tab = "output";
    if (scroll) wrapper?.scrollIntoView({ behavior: "smooth", block: "start" });
  }
</script>

<div bind:this={wrapper}>
  <div
    class={{
      "bg-base-100 sticky top-0 z-20 pt-4": !storeUi.store.isMobile,
    }}
  >
    <H2 class="mb-4 flex items-center justify-between">
      <span>Try it out</span>
      <HistoryButton {storeNode} nodeName={stream.name} />
    </H2>

    <Tabs
      items={[
        { id: "input", label: "Input", icon: MoveUpRight },
        { id: "output", label: "Output", icon: MoveDownLeft },
      ]}
      bind:active={tab}
    />
  </div>

  <div
    class={{
      "space-y-4": true,
      hidden: tab === "output",
      block: tab === "input",
    }}
  >
    {#if stream.input && stream.input.length > 0}
      <div class="space-y-4" onkeydown={executeStreamFromKbd} role="button" tabindex="0">
        <InputForm fields={stream.input} bind:input={storeNode.store.input} />
      </div>
    {:else}
      <div role="alert" class="alert alert-soft alert-warning mt-6 w-fit">
        <Info class="size-4" />
        <span>This stream does not require any input</span>
      </div>
    {/if}

    <div class="flex w-full justify-end gap-2">
      {#snippet kbd()}
        <span>
          <kbd class="kbd kbd-sm">{ctrlSymbol()}</kbd>
          <kbd class="kbd kbd-sm">⤶</kbd>
        </span>
      {/snippet}

      <button
        class="btn btn-primary btn-ghost"
        disabled={isExecuting}
        onclick={storeNode.actions.clearInput}
      >
        <Trash class="size-4" />
        <span>Clear input</span>
      </button>

      <Menu content={kbd} placement="bottom" trigger="mouseenter">
        <button class="btn btn-primary" disabled={isExecuting} onclick={executeStream}>
          {#if isExecuting}
            <Loader class="animate size-4 animate-spin" />
          {:else}
            <Zap class="size-4" />
          {/if}
          <span>Start stream</span>
        </button>
      </Menu>
    </div>
  </div>

  <div
    class={{
      "mt-4": true,
      hidden: tab === "input",
      block: tab === "output",
    }}
  >
    <Output type="stream" {cancelRequest} {isExecuting} {storeNode} />
  </div>
</div>

{#if storeUi.store.isMobile}
  <div class="mt-12">
    <Snippets type="stream" name={stream.name} input={storeNode.store.input} />
  </div>
{/if}
