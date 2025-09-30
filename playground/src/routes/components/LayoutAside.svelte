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
</script>

{#snippet aside()}
  <aside
    use:dimensionschangeAction
    ondimensionschange={(e) => (storeUi.store.aside = e.detail)}
    class={[
      "bg-base-100 h-[100dvh] w-full max-w-[280px] flex-none scroll-p-[130px]",
      "overflow-x-hidden overflow-y-auto",
    ]}
  >
    <header class="bg-base-100 sticky top-0 z-10 w-full shadow-xs">
      {#if !storeUi.store.isMobile}
        <a
          class="sticky top-0 z-10 flex h-[72px] w-full items-end p-4 shadow-xs"
          href="https://uforpc.uforg.dev"
          target="_blank"
        >
          <Tooltip content={versionWithPrefix} placement="right">
            <Logo class="mx-auto h-full" />
          </Tooltip>
        </a>
      {/if}

      {#if storeUi.store.isMobile}
        <div class="flex items-center justify-between p-4">
          <Logo class="w-[180px]" />

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
{/snippet}

{#if !storeUi.store.isMobile}
  {@render aside()}
{/if}

{#if storeUi.store.isMobile}
  <Offcanvas bind:isOpen={storeUi.store.asideOpen}>
    {@render aside()}
  </Offcanvas>
{/if}
