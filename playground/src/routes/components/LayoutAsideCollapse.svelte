<script lang="ts">
  import { ChevronDown, ChevronRight, TriangleAlert } from "@lucide/svelte";
  import { onMount, type Snippet } from "svelte";
  import { browser } from "$app/environment";
  import Tooltip from "$lib/components/Tooltip.svelte";
  import { createGlobalDbNamePrefix } from "$lib/createStore.svelte";

  interface Props {
    icon: typeof ChevronDown;
    label: string;
    storageKey?: string;
    deprecated?: string;
    children?: Snippet;
  }
  const { icon: Icon, label: rawLabel, storageKey, deprecated, children }: Props = $props();

  // safeStorageKey key avoids collisions across deployments on same domain
  let safeStorageKey = $derived.by(() => {
    if (!browser || !storageKey) return null;
    return `${createGlobalDbNamePrefix()}:${storageKey}`;
  });

  let isOpen = $state(false);
  let isDeprecated = $derived(deprecated || deprecated === "");
  let label = $derived(isDeprecated ? `${rawLabel} (Deprecated)` : rawLabel);

  onMount(() => {
    if (!safeStorageKey) return;
    if (localStorage.getItem(safeStorageKey) === "true") isOpen = true;
  });
  $effect(() => {
    if (safeStorageKey) localStorage.setItem(safeStorageKey, String(isOpen));
  });
</script>

<div>
  <Tooltip content={label}>
    <button
      onclick={() => (isOpen = !isOpen)}
      class="btn btn-ghost btn-block items-center justify-between space-x-2 border-transparent hover:bg-blue-500/20"
    >
      <span class="flex min-w-0 items-center justify-start space-x-2">
        {#if isDeprecated}
          <TriangleAlert class="text-warning size-4 flex-none" />
        {:else}
          <Icon class="size-4 flex-none" />
        {/if}
        <span
          class={{
            "overflow-hidden overflow-ellipsis whitespace-nowrap": true,
            "line-through": isDeprecated,
          }}
        >
          {label}
        </span>
      </span>
      {#if isOpen}
        <ChevronDown class="size-4 flex-none" />
      {:else}
        <ChevronRight class="size-4 flex-none" />
      {/if}
    </button>

    {#if isOpen && children}
      <div class="border-base-content/20 mt-1 ml-6 space-y-1 border-l-2 border-dashed pl-1">
        {@render children()}
      </div>
    {/if}
  </Tooltip>
</div>
