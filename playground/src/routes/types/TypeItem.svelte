<script lang="ts">
  import { pushState } from "$app/navigation";
  import { Expand, Link, Minimize, TriangleAlert } from "@lucide/svelte";

  import { markdownToHtml } from "$lib/helpers/markdownToHtml";
  import { slugify } from "$lib/helpers/slugify";
  import { storeSettings } from "$lib/storeSettings.svelte";
  import { storeUi } from "$lib/storeUi.svelte";
  import { wasmClient } from "$lib/wasm/index";
  import type { TypeDef } from "$lib/wasm/wasmtypes/types";

  import Code from "$lib/components/Code.svelte";
  import Tabs from "$lib/components/Tabs.svelte";

  interface Props {
    type: TypeDef;
  }

  let { type }: Props = $props();

  let isCompact = $derived(storeUi.store.schemaViewMode === "compact");
  let isExpanded = $derived(storeUi.store.schemaViewMode === "expanded");
  let extractedSchemaCompact = $state("");
  let extractedSchemaExpanded = $state("");

  const slug = slugify(`types#${type.name}`);
  const href = `/#/${slug}`;

  function scrollTo(id: string) {
    document
      .getElementById(id)
      ?.scrollIntoView({ behavior: "smooth", block: "start" });
  }

  function handleClick(e: MouseEvent) {
    e.preventDefault();
    pushState(href, {});
    scrollTo(slug);
  }

  $effect(() => {
    if (isCompact && extractedSchemaCompact !== "") return;
    if (isExpanded && extractedSchemaExpanded !== "") return;

    (async () => {
      const schemaCompact = storeSettings.store.vdlSchema;
      const schemaExpanded = storeSettings.store.vdlSchemaExpanded;

      try {
        if (isCompact) {
          extractedSchemaCompact = await wasmClient.extractType(
            schemaCompact,
            type.name,
          );
        }
        if (isExpanded) {
          extractedSchemaExpanded = await wasmClient.extractType(
            schemaExpanded,
            type.name,
          );
        }
      } catch (err) {
        console.error(`Failed to extract schema for ${type.name}`, err);
      }
    })();
  });

  let schemaToShow = $derived.by(() => {
    if (isCompact) return extractedSchemaCompact;
    if (isExpanded) return extractedSchemaExpanded;
    return "";
  });
</script>

<div class="card bg-base-200 shadow-sm">
  <div class="card-body gap-4">
    <div class="flex items-center gap-2">
      <a
        {href}
        class="btn btn-ghost btn-sm btn-square shrink-0 opacity-50 hover:opacity-100"
        onclick={handleClick}
      >
        <Link class="size-4" />
      </a>

      <a {href} class="group min-w-0 flex-1" onclick={handleClick}>
        <h2
          id={slug}
          class={{
            "scroll-mt-4 truncate text-xl font-bold group-hover:underline": true,
            "line-through opacity-60": typeof type.deprecated === "string",
          }}
        >
          {type.name}
        </h2>
      </a>
    </div>

    {#if typeof type.deprecated === "string"}
      <div class="alert alert-warning">
        <TriangleAlert class="size-5" />
        <span class="font-semibold">
          {type.deprecated || "Deprecated"}
        </span>
      </div>
    {/if}

    {#if type.doc}
      {#await markdownToHtml(type.doc) then html}
        <div class="prose prose-sm max-w-none">{@html html}</div>
      {/await}
    {/if}

    <div class="space-y-2">
      <div class="flex items-center justify-between gap-4">
        <div class="flex-1"></div>
        <Tabs
          containerClass="w-auto flex-none"
          buttonClass="btn-xs"
          iconClass="size-3"
          bind:active={storeUi.store.schemaViewMode}
          items={[
            {
              id: "compact",
              label: "Compact",
              icon: Minimize,
              tooltipText: "Show compact schema with type references",
            },
            {
              id: "expanded",
              label: "Expanded",
              icon: Expand,
              tooltipText: "Show expanded schema with all types inlined",
            },
          ]}
        />
      </div>

      {#if schemaToShow}
        <Code
          code={schemaToShow}
          lang="vdl"
          withBorder={true}
          class="max-h-125"
        />
      {:else}
        <div class="skeleton h-32 w-full"></div>
      {/if}
    </div>
  </div>
</div>
