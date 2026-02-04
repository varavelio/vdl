<script lang="ts">
  import { untrack } from "svelte";

  import { getMarkdownTitle } from "$lib/helpers/getMarkdownTitle";
  import { slugify } from "$lib/helpers/slugify";
  import { storeSettings } from "$lib/storeSettings.svelte";
  import { storeUi } from "$lib/storeUi.svelte";

  import type { PageProps } from "./$types";
  import Node from "./components/Node.svelte";
  import NotFound from "./components/NotFound.svelte";
  import { createStoreNode } from "./storeNode.svelte";

  let { data }: PageProps = $props();

  // Create a per-slug store and recreate it whenever the slug changes
  let storeNode = $state(createStoreNode(data.nodeSlug));
  $effect(() => {
    storeNode = createStoreNode(data.nodeSlug);
  });

  let nodeIndex = $derived.by(() => {
    for (const [
      index,
      node,
    ] of storeSettings.store.jsonSchema.nodes.entries()) {
      if (node.kind !== data.nodeKind) continue;

      const isDoc = node.kind === "doc";
      let nodeName = isDoc ? getMarkdownTitle(node.content) : node.name;
      nodeName = slugify(nodeName);

      if (data.nodeName === nodeName) return index;
    }

    return -1; // Node not found in store
  });

  let nodeExists = $derived(nodeIndex !== -1);

  let node = $derived(storeSettings.store.jsonSchema.nodes[nodeIndex]);

  let name = $derived.by(() => {
    if (node.kind === "type") return node.name;
    if (node.kind === "proc") return node.name;
    if (node.kind === "stream") return node.name;
    if (node.kind === "doc") {
      return getMarkdownTitle(node.content);
    }

    return "unknown";
  });

  let humanKind = $derived.by(() => {
    if (node.kind === "type") return "type";
    if (node.kind === "proc") return "procedure";
    if (node.kind === "stream") return "stream";
    if (node.kind === "doc") return "documentation";
    return "unknown";
  });

  let title = $derived.by(() => {
    if (!nodeExists) return "VDL Playground";

    return `${name} ${humanKind} | VDL Playground`;
  });

  // Scroll to top of page when node changes
  $effect(() => {
    nodeIndex; // Just to add a dependency to trigger the effect
    untrack(() => {
      // Untrack the storeUi.contentWrapper.element to avoid infinite loop
      storeUi.store.contentWrapper.element?.scrollTo({
        top: 0,
        behavior: "smooth",
      });
    });
  });
</script>

<svelte:head>
  <title>{title}</title>
</svelte:head>

{#if !nodeExists}
  <NotFound />
{/if}

{#if nodeExists && storeNode.status.ready}
  {#key nodeIndex}
    <Node {node} bind:storeNode />
  {/key}
{/if}
