<script lang="ts">
  import { page } from "$app/state";
  import { ChevronDown, TriangleAlert } from "@lucide/svelte";

  import { storeUi } from "$lib/storeUi.svelte";

  import Tooltip from "$lib/components/Tooltip.svelte";

  interface Props {
    icon: typeof ChevronDown;
    label: string;
    href: string;
    deprecated?: string;
  }
  const { icon: Icon, label: rawLabel, href, deprecated }: Props = $props();

  let id = $derived(`navlink-${href.replaceAll("#/", "")}`);
  let isActive = $derived(page.url.hash == href);
  let isDeprecated = $derived(deprecated || deprecated === "");
  let label = $derived(isDeprecated ? `${rawLabel} (Deprecated)` : rawLabel);
</script>

<Tooltip content={label}>
  <a
    {id}
    {href}
    onclick={() => (storeUi.store.asideOpen = false)}
    class={{
      "btn btn-ghost btn-block items-center justify-between space-x-2 border-transparent hover:bg-blue-500/20": true,
      "bg-blue-500/20": isActive,
    }}
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
  </a>
</Tooltip>
