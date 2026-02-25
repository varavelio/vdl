<script lang="ts">
  import { Copy, Link, Lock, TriangleAlert } from "@lucide/svelte";
  import { onMount } from "svelte";
  import { pushState } from "$app/navigation";
  import BottomSpace from "$lib/components/BottomSpace.svelte";
  import H1 from "$lib/components/H1.svelte";
  import { copyTextToClipboard } from "$lib/helpers/copyTextToClipboard";
  import { markdownToHtml } from "$lib/helpers/markdownToHtml";
  import { slugify } from "$lib/helpers/slugify";
  import { storeSettings } from "$lib/storeSettings.svelte";
  import { storeUi } from "$lib/storeUi.svelte";

  let constants = $derived(storeSettings.store.irSchema.constants);
  let isMobile = $derived(storeUi.store.isMobile);

  const getSlug = (name: string) => slugify(`constants#${name}`);
  const getHref = (name: string) => `/#/${getSlug(name)}`;

  const badgeClasses: Record<string, string> = {
    string: "badge badge-soft badge-success font-mono",
    int: "badge badge-soft badge-info font-mono",
    float: "badge badge-soft badge-warning font-mono",
    bool: "badge badge-soft badge-secondary font-mono",
  };
  const getBadgeClass = (type: string) =>
    badgeClasses[type] ?? "badge badge-soft badge-neutral font-mono";

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
      setTimeout(() => scrollTo(`constants#${parts.at(-1)}`), 100);
    }
  });
</script>

<svelte:head> <title>Constants | VDL Playground</title> </svelte:head>

<div class="h-full overflow-y-auto scroll-smooth">
  {#if constants.length === 0}
    <div class="flex h-full flex-col items-center justify-center p-4 text-center">
      <div class="card bg-base-200 w-full max-w-md shadow-lg">
        <div class="card-body items-center text-center">
          <Lock class="text-base-content/40 mb-4 size-16" />
          <H1 class="text-2xl">No Constants Defined</H1>
          <p class="text-base-content/60 mt-2">Your schema doesn't have any constants yet.</p>
        </div>
      </div>
    </div>
  {:else}
    <div class={{ "h-full": true, flex: !isMobile }}>
      <div class={{ "flex-1 p-4": true, "overflow-y-auto": !isMobile }}>
        <div class="mb-8">
          <H1>Schema Constants</H1>
          <p class="text-base-content/60 mt-2">
            Constants define fixed values that remain unchanged throughout your application.
          </p>
        </div>

        <div class="space-y-4">
          {#each constants as c (c.name)}
            {@const slug = getSlug(c.name)}
            {@const href = getHref(c.name)}

            <div class="card bg-base-200 shadow-sm">
              <div class="card-body gap-4">
                <div class="flex items-center gap-2">
                  <a
                    {href}
                    class="btn btn-ghost btn-sm btn-square shrink-0 opacity-50 hover:opacity-100"
                    onclick={(e) => handleClick(e, c.name)}
                  >
                    <Link class="size-4" />
                  </a>

                  <a {href} class="group min-w-0 flex-1" onclick={(e) => handleClick(e, c.name)}>
                    <h2
                      id={slug}
                      class={{
                        "scroll-mt-4 truncate text-xl font-bold group-hover:underline": true,
                        "line-through opacity-60":
                          typeof c.deprecated === "string",
                      }}
                    >
                      {c.name}
                    </h2>
                  </a>

                  <span class={getBadgeClass(c.constType)}>{c.constType}</span>
                </div>

                {#if typeof c.deprecated === "string"}
                  <div class="alert alert-warning">
                    <TriangleAlert class="size-5" />
                    <span class="font-semibold"> {c.deprecated || "Deprecated"} </span>
                  </div>
                {/if}

                {#if c.doc}
                  {#await markdownToHtml(c.doc) then html}
                    <div class="prose prose-sm max-w-none">{@html html}</div>
                  {/await}
                {/if}

                <div>
                  <span class="text-base-content/60 mb-2 block text-sm"> Value </span>
                  <div class="flex items-center gap-2">
                    <input
                      type="text"
                      readonly
                      value={c.value}
                      class="input input-bordered flex-1 font-mono"
                    >
                    <button
                      class="btn btn-square btn-ghost"
                      onclick={() => copyTextToClipboard(c.value)}
                      title="Copy value"
                    >
                      <Copy class="size-4" />
                    </button>
                  </div>
                </div>
              </div>
            </div>
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
              {#each constants as c (c.name)}
                <li>
                  <a href={getHref(c.name)} class="gap-2" onclick={(e) => handleClick(e, c.name)}>
                    <Link class="size-3 shrink-0 opacity-50" />
                    <span class="truncate">{c.name}</span>
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
