<script lang="ts">
  import { Braces, Link } from "@lucide/svelte";
  import { onMount } from "svelte";
  import { pushState } from "$app/navigation";
  import BottomSpace from "$lib/components/BottomSpace.svelte";
  import H1 from "$lib/components/H1.svelte";
  import { slugify } from "$lib/helpers/slugify";
  import { storeSettings } from "$lib/storeSettings.svelte";
  import { storeUi } from "$lib/storeUi.svelte";

  import TypeItem from "./TypeItem.svelte";

  let types = $derived(storeSettings.store.irSchema.types);
  let isMobile = $derived(storeUi.store.isMobile);

  const getSlug = (name: string) => slugify(`types#${name}`);
  const getHref = (name: string) => `/#/${getSlug(name)}`;

  function scrollTo(slug: string) {
    document.getElementById(slug)?.scrollIntoView({ behavior: "smooth", block: "start" });
  }

  function handleClick(e: MouseEvent, name: string) {
    e.preventDefault();
    const slug = getSlug(name);
    pushState(getHref(name), {});
    scrollTo(slug);
  }

  onMount(() => {
    const parts = window.location.hash.split("#");
    if (parts.length >= 2) {
      setTimeout(() => scrollTo(`types#${parts.at(-1)}`), 100);
    }
  });
</script>

<svelte:head> <title>Types | VDL Playground</title> </svelte:head>

<div class="h-full overflow-y-auto scroll-smooth">
  {#if types.length === 0}
    <div class="flex h-full flex-col items-center justify-center p-4 text-center">
      <div class="card bg-base-200 w-full max-w-md shadow-lg">
        <div class="card-body items-center text-center">
          <Braces class="text-base-content/40 mb-4 size-16" />
          <H1 class="text-2xl">No Types Defined</H1>
          <p class="text-base-content/60 mt-2">Your schema doesn't have any types yet.</p>
        </div>
      </div>
    </div>
  {:else}
    <div class={{ "h-full": true, flex: !isMobile }}>
      <div class={{ "flex-1 p-4": true, "overflow-y-auto": !isMobile }}>
        <div class="mb-8">
          <H1>Schema Types</H1>
          <p class="text-base-content/60 mt-2">
            Types define the structure of the data models used in your schema.
          </p>
        </div>

        <div class="space-y-4">
          {#each types as t (t.name)}
            <TypeItem type={t} />
          {/each}
        </div>

        <BottomSpace />
      </div>

      {#if !isMobile}
        <aside class="border-base-300 flex h-full w-64 shrink-0 flex-col border-l p-4">
          <h3
            class="text-base-content/60 mb-4 shrink-0 px-2 text-sm font-semibold tracking-wide uppercase"
          >
            On this page
          </h3>
          <nav class="min-h-0 flex-1 overflow-y-auto">
            <ul class="menu menu-sm w-full">
              {#each types as t (t.name)}
                <li>
                  <a href={getHref(t.name)} class="gap-2" onclick={(e) => handleClick(e, t.name)}>
                    <Link class="size-3 shrink-0 opacity-50" />
                    <span class="truncate">{t.name}</span>
                  </a>
                </li>
              {/each}
            </ul>
            <BottomSpace />
          </nav>
        </aside>
      {/if}
    </div>
  {/if}
</div>
