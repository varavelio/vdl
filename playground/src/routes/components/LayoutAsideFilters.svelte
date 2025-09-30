<script lang="ts">
  import {
    ArrowLeftRight,
    BookOpenText,
    CornerRightDown,
    Funnel,
    FunnelX,
    Search,
    Type,
    X,
  } from "@lucide/svelte";

  import { storeUi } from "$lib/storeUi.svelte";

  import Tooltip from "$lib/components/Tooltip.svelte";

  const searchTooltip = $derived(
    storeUi.store.asideSearchOpen ? "Close search" : "Open search",
  );
  const docsTooltip = $derived(
    storeUi.store.asideHideDocs ? "Show documentation" : "Hide documentation",
  );
  const typesTooltip = $derived(
    storeUi.store.asideHideTypes ? "Show data types" : "Hide data types",
  );
  const procsTooltip = $derived(
    storeUi.store.asideHideProcs ? "Show procedures" : "Hide procedures",
  );
  const streamsTooltip = $derived(
    storeUi.store.asideHideStreams ? "Show streams" : "Hide streams",
  );

  let searchInput: HTMLInputElement | null = $state(null);

  function openSearch() {
    storeUi.store.asideSearchOpen = true;
    storeUi.store.asideSearchQuery = "";

    setTimeout(() => {
      searchInput?.focus();
    }, 100);
  }

  function closeSearch() {
    storeUi.store.asideSearchOpen = false;
    storeUi.store.asideSearchQuery = "";
  }

  function toggleDocs() {
    storeUi.store.asideHideDocs = !storeUi.store.asideHideDocs;
  }

  function toggleTypes() {
    storeUi.store.asideHideTypes = !storeUi.store.asideHideTypes;
  }

  function toggleProcs() {
    storeUi.store.asideHideProcs = !storeUi.store.asideHideProcs;
  }

  function toggleStreams() {
    storeUi.store.asideHideStreams = !storeUi.store.asideHideStreams;
  }

  function resetFilters() {
    storeUi.store.asideSearchOpen = false;
    storeUi.store.asideSearchQuery = "";
    storeUi.store.asideHideDocs = false;
    storeUi.store.asideHideTypes = true;
    storeUi.store.asideHideProcs = false;
    storeUi.store.asideHideStreams = false;
  }
</script>

<div class="flex w-full justify-center space-x-1 px-4 py-2">
  <Tooltip content="Reset filters to default" placement="bottom">
    <button class="btn btn-sm btn-square group" onclick={resetFilters}>
      <Funnel class="size-4 group-hover:hidden" />
      <FunnelX class="hidden size-4 group-hover:inline" />
    </button>
  </Tooltip>

  {#if storeUi.store.asideSearchOpen}
    <input
      type="text"
      class="input input-sm flex-grow"
      placeholder="Search..."
      bind:this={searchInput}
      bind:value={storeUi.store.asideSearchQuery}
    />

    <Tooltip content={searchTooltip} placement="bottom">
      <button class={["btn btn-sm btn-square relative"]} onclick={closeSearch}>
        <X class="size-4" />
      </button>
    </Tooltip>
  {/if}

  {#if !storeUi.store.asideSearchOpen}
    <Tooltip content={searchTooltip} placement="bottom">
      <button class={["btn btn-sm btn-square relative"]} onclick={openSearch}>
        <Search class="size-4" />
      </button>
    </Tooltip>
    <Tooltip content={docsTooltip} placement="bottom">
      <button
        class={[
          "btn btn-sm btn-square relative",
          storeUi.store.asideHideDocs && "toggle-disabled",
        ]}
        onclick={toggleDocs}
      >
        <BookOpenText class="size-4" />
      </button>
    </Tooltip>
    <Tooltip content={typesTooltip} placement="bottom">
      <button
        class={[
          "btn btn-sm btn-square relative",
          storeUi.store.asideHideTypes && "toggle-disabled",
        ]}
        onclick={toggleTypes}
      >
        <Type class="size-4" />
      </button>
    </Tooltip>
    <Tooltip content={procsTooltip} placement="bottom">
      <button
        class={[
          "btn btn-sm btn-square relative",
          storeUi.store.asideHideProcs && "toggle-disabled",
        ]}
        onclick={toggleProcs}
      >
        <ArrowLeftRight class="size-4" />
      </button>
    </Tooltip>
    <Tooltip content={streamsTooltip} placement="bottom">
      <button
        class={[
          "btn btn-sm btn-square relative",
          storeUi.store.asideHideStreams && "toggle-disabled",
        ]}
        onclick={toggleStreams}
      >
        <CornerRightDown class="size-4" />
      </button>
    </Tooltip>
  {/if}
</div>

<style lang="postcss">
  .toggle-disabled::before {
    content: "";
    position: absolute;
    top: 50%;
    left: 10%;
    width: 80%;
    height: 2px;
    background-color: currentColor;
    transform: translateY(-50%) rotate(-45deg);
    opacity: 0.7;
    z-index: 1;
    pointer-events: none;
  }

  .toggle-disabled {
    opacity: 0.6;
  }
</style>
