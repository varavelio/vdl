<script lang="ts">
  import { page } from "$app/state";
  import { House, X } from "@lucide/svelte";
  import { onMount } from "svelte";

  import { storeSettings } from "$lib/storeSettings.svelte";
  import { dimensionschangeAction, storeUi } from "$lib/storeUi.svelte";
  import { versionWithPrefix } from "$lib/version";

  import Logo from "$lib/components/Logo.svelte";
  import Offcanvas from "$lib/components/Offcanvas.svelte";
  import Tooltip from "$lib/components/Tooltip.svelte";

  import LayoutAsideFilters from "./LayoutAsideFilters.svelte";
  import LayoutAsideItemWrapper from "./LayoutAsideItemWrapper.svelte";

  // if has hash anchor navigate to it
  onMount(async () => {
    // wait 500ms to ensure the content is rendered
    await new Promise((resolve) => setTimeout(resolve, 500));

    if (window.location.hash) {
      const id = `navlink-${window.location.hash.replaceAll("#/", "")}`;
      const element = document.getElementById(id);
      if (element) {
        element.scrollIntoView({ behavior: "smooth" });
      }
    }
  });

  let isHome = $derived(page.url.hash === "" || page.url.hash === "#/");

  const defaultSize = 280;
  const minSize = 210;
  const maxSize = 600;
  let isResizing = $state(false);

  function startResize(e: MouseEvent) {
    isResizing = true;
    e.preventDefault();
  }

  function stopResize() {
    isResizing = false;
  }

  function resize(e: MouseEvent) {
    if (!isResizing) return;
    const newWidth = e.clientX;
    if (newWidth >= minSize && newWidth <= maxSize) {
      storeUi.store.asideWidth = newWidth;
    }
  }

  function handleDoubleClick() {
    storeUi.store.asideWidth = defaultSize;
  }

  let asideStyle = $derived.by(() => {
    if (storeUi.store.isMobile) {
      return `width: 100%; max-width: ${defaultSize}px;`;
    }

    return `width: ${storeUi.store.asideWidth}px;`;
  });
</script>

{#snippet aside()}
  <aside
    use:dimensionschangeAction
    ondimensionschange={(e) => (storeUi.store.aside = e.detail)}
    style={asideStyle}
    class={[
      "bg-base-100 relative h-[100dvh] flex-none scroll-p-[130px]",
      "overflow-x-hidden overflow-y-auto",
    ]}
  >
    <header class="bg-base-100 sticky top-0 z-10 w-full shadow-xs">
      {#if !storeUi.store.isMobile}
        <a
          class="sticky top-0 z-10 flex h-[72px] w-full items-end p-4"
          href="https://uforpc.uforg.dev"
          target="_blank"
        >
          <Tooltip content={versionWithPrefix} placement="right">
            <Logo class="mx-auto w-[180px]" />
          </Tooltip>
        </a>
      {/if}

      {#if storeUi.store.isMobile}
        <div class="flex items-center justify-between p-4">
          <Logo class="mx-3 w-[180px]" />

          <button
            class="btn btn-ghost btn-square btn-sm"
            onclick={() => (storeUi.store.asideOpen = !storeUi.store.asideOpen)}
          >
            <X class="size-6" />
          </button>
        </div>
      {/if}

      <LayoutAsideFilters />
    </header>

    <nav class="p-4 pb-8">
      <Tooltip content="RPC Home">
        <a
          href="#/"
          onclick={() => (storeUi.store.asideOpen = false)}
          class={[
            "btn btn-ghost btn-block justify-start space-x-2 border-transparent",
            "hover:bg-blue-500/20",
            { "bg-blue-500/20": isHome },
          ]}
        >
          <House class="size-4" />
          <span>Home</span>
        </a>
      </Tooltip>

      {#each storeSettings.store.jsonSchema.nodes as node}
        <LayoutAsideItemWrapper {node} />
      {/each}
    </nav>
  </aside>

  {#if !storeUi.store.isMobile}
    <button
      aria-label="Resize"
      class={[
        "group border-base-content/20 fixed top-0 z-40 h-[100dvh] w-[6px] cursor-col-resize border-l",
        "hover:bg-primary/20 hover:border-l-0",
      ]}
      style="left: {storeUi.store.asideWidth - 2}px;"
      onmousedown={startResize}
      ondblclick={handleDoubleClick}
    ></button>
  {/if}
{/snippet}

<svelte:window onmousemove={resize} onmouseup={stopResize} />

{#if !storeUi.store.isMobile}
  {@render aside()}
{/if}

{#if storeUi.store.isMobile}
  <Offcanvas bind:isOpen={storeUi.store.asideOpen}>
    {@render aside()}
  </Offcanvas>
{/if}
