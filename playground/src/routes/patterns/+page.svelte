<script lang="ts">
  import { pushState } from "$app/navigation";
  import { Copy, Link, Regex, TriangleAlert } from "@lucide/svelte";
  import { onMount } from "svelte";

  import { copyTextToClipboard } from "$lib/helpers/copyTextToClipboard";
  import { markdownToHtml } from "$lib/helpers/markdownToHtml";
  import { slugify } from "$lib/helpers/slugify";
  import { storeSettings } from "$lib/storeSettings.svelte";
  import { storeUi } from "$lib/storeUi.svelte";

  import BottomSpace from "$lib/components/BottomSpace.svelte";
  import H1 from "$lib/components/H1.svelte";

  let patterns = $derived(storeSettings.store.irSchema.patterns);
  let isMobile = $derived(storeUi.store.isMobile);

  const getSlug = (name: string) => slugify(`patterns#${name}`);
  const getHref = (name: string) => `/#/${getSlug(name)}`;

  function scrollTo(slug: string) {
    document
      .getElementById(slug)
      ?.scrollIntoView({ behavior: "smooth", block: "start" });
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
      setTimeout(() => scrollTo(`patterns#${parts.at(-1)}`), 100);
    }
  });
</script>

<svelte:head>
  <title>Patterns | VDL Playground</title>
</svelte:head>

<div class="h-full overflow-y-auto scroll-smooth">
  {#if patterns.length === 0}
    <div
      class="flex h-full flex-col items-center justify-center p-4 text-center"
    >
      <div class="card bg-base-200 w-full max-w-md shadow-lg">
        <div class="card-body items-center text-center">
          <Regex class="text-base-content/40 mb-4 size-16" />
          <H1 class="text-2xl">No Patterns Defined</H1>
          <p class="text-base-content/60 mt-2">
            Your schema doesn't have any patterns yet.
          </p>
        </div>
      </div>
    </div>
  {:else}
    <div class={{ "h-full": true, flex: !isMobile }}>
      <div class={{ "flex-1 p-4": true, "overflow-y-auto": !isMobile }}>
        <div class="mb-8">
          <H1>Schema Patterns</H1>
          <p class="text-base-content/60 mt-2">
            Patterns are templates for generating dynamic strings with
            placeholders.
          </p>
        </div>

        <div class="space-y-4">
          {#each patterns as p (p.name)}
            {@const slug = getSlug(p.name)}
            {@const href = getHref(p.name)}

            <div class="card bg-base-200 shadow-sm">
              <div class="card-body gap-4">
                <div class="flex items-center gap-2">
                  <a
                    {href}
                    class="btn btn-ghost btn-sm btn-square shrink-0 opacity-50 hover:opacity-100"
                    onclick={(e) => handleClick(e, p.name)}
                  >
                    <Link class="size-4" />
                  </a>

                  <a
                    {href}
                    class="group min-w-0 flex-1"
                    onclick={(e) => handleClick(e, p.name)}
                  >
                    <h2
                      id={slug}
                      class={{
                        "scroll-mt-4 truncate text-xl font-bold group-hover:underline": true,
                        "line-through opacity-60":
                          typeof p.deprecated === "string",
                      }}
                    >
                      {p.name}
                    </h2>
                  </a>
                </div>

                {#if typeof p.deprecated === "string"}
                  <div class="alert alert-warning">
                    <TriangleAlert class="size-5" />
                    <span class="font-semibold">
                      {p.deprecated || "Deprecated"}
                    </span>
                  </div>
                {/if}

                {#if p.doc}
                  {#await markdownToHtml(p.doc) then html}
                    <div class="prose prose-sm max-w-none">{@html html}</div>
                  {/await}
                {/if}

                <div>
                  <span class="text-base-content/60 mb-2 block text-sm">
                    Template
                  </span>
                  <div class="flex items-center gap-2">
                    <input
                      type="text"
                      readonly
                      value={p.template}
                      class="input input-bordered flex-1 font-mono"
                    />
                    <button
                      class="btn btn-square btn-ghost"
                      onclick={() => copyTextToClipboard(p.template)}
                      title="Copy template"
                    >
                      <Copy class="size-4" />
                    </button>
                  </div>
                </div>

                {#if p.placeholders.length > 0}
                  <div>
                    <span class="text-base-content/60 mb-2 block text-sm">
                      Placeholders
                    </span>
                    <div class="flex flex-wrap gap-2">
                      {#each p.placeholders as placeholder}
                        <span class="badge badge-soft badge-neutral font-mono">
                          {placeholder}
                        </span>
                      {/each}
                    </div>
                  </div>
                {/if}
              </div>
            </div>
          {/each}
        </div>

        <BottomSpace />
      </div>

      {#if !isMobile}
        <aside
          class="border-base-300 flex h-full w-64 shrink-0 flex-col border-l p-4"
        >
          <h3
            class="text-base-content/60 mb-4 shrink-0 px-2 text-sm font-semibold tracking-wide uppercase"
          >
            On this page
          </h3>
          <nav class="min-h-0 flex-1 overflow-y-auto">
            <ul class="menu menu-sm w-full">
              {#each patterns as p (p.name)}
                <li>
                  <a
                    href={getHref(p.name)}
                    class="gap-2"
                    onclick={(e) => handleClick(e, p.name)}
                  >
                    <Link class="size-3 shrink-0 opacity-50" />
                    <span class="truncate">{p.name}</span>
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
