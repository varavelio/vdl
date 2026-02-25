<script lang="ts">
  import { Info, Loader, MoveDownLeft, MoveUpRight, Trash, Zap } from "@lucide/svelte";
  import { toast } from "svelte-sonner";
  import H2 from "$lib/components/H2.svelte";
  import Menu from "$lib/components/Menu.svelte";
  import Tabs from "$lib/components/Tabs.svelte";
  import { ctrlSymbol } from "$lib/helpers/ctrlSymbol";
  import { joinPath } from "$lib/helpers/joinPath";
  import { getHeadersObject, type ProcedureDef, storeSettings } from "$lib/storeSettings.svelte";
  import { storeUi } from "$lib/storeUi.svelte";

  import type { StoreNodeInstance } from "../../storeNode.svelte";

  import HistoryButton from "./HistoryButton/HistoryButton.svelte";
  import InputForm from "./InputForm/InputForm.svelte";
  import Output from "./Output.svelte";
  import Snippets from "./Snippets/Snippets.svelte";

  interface Props {
    proc: ProcedureDef & { kind: "proc" };
    storeNode: StoreNodeInstance;
  }

  let { proc, storeNode = $bindable() }: Props = $props();

  let isExecuting = $state(false);
  let cancelRequest = $state<() => void>(() => {});

  async function executeProcedure() {
    if (isExecuting) return;
    isExecuting = true;
    storeNode.store.output = "";
    storeNode.store.outputDate = "";

    try {
      openOutputTab(true);
      const controller = new AbortController();
      const signal = controller.signal;

      cancelRequest = () => {
        controller.abort();
        toast.info("Procedure call cancelled");
      };

      const endpoint = joinPath([storeSettings.store.baseUrl, proc.rpcName, proc.name]);
      const response = await fetch(endpoint, {
        method: "POST",
        body: JSON.stringify(storeNode.store.input ?? {}),
        headers: getHeadersObject(),
        signal: signal,
      });

      const data = await response.json();
      storeNode.store.output = JSON.stringify(data, null, 2);
      storeNode.store.outputDate = new Date().toISOString();
      storeNode.actions.saveCurrentToHistory();
    } catch (error) {
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
    }
  }

  async function executeProcedureFromKbd(event: KeyboardEvent) {
    if (event.key === "Enter" && (event.ctrlKey || event.metaKey)) {
      event.preventDefault();
      await executeProcedure();
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
      <HistoryButton {storeNode} nodeName={proc.name} />
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
    {#if proc.input && proc.input.length > 0}
      <div class="space-y-4" onkeydown={executeProcedureFromKbd} role="button" tabindex="0">
        <InputForm fields={proc.input} bind:input={storeNode.store.input} />
      </div>
    {:else}
      <div role="alert" class="alert alert-soft alert-warning mt-6 w-fit">
        <Info class="size-4" />
        <span>This procedure does not require any input</span>
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
        <button class="btn btn-primary" disabled={isExecuting} onclick={executeProcedure}>
          {#if isExecuting}
            <Loader class="animate size-4 animate-spin" />
          {:else}
            <Zap class="size-4" />
          {/if}
          <span>Execute procedure</span>
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
    <Output type="proc" {cancelRequest} {isExecuting} {storeNode} />
  </div>
</div>

{#if storeUi.store.isMobile}
  <div class="mt-12"><Snippets type="proc" name={proc.name} input={storeNode.store.input} /></div>
{/if}
