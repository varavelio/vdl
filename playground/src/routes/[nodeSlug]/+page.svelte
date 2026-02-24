<script lang="ts">
  import { untrack } from "svelte";

  import { getMarkdownTitle } from "$lib/helpers/getMarkdownTitle";
  import { slugify } from "$lib/helpers/slugify";
  import { getIrNodes, type IrNode } from "$lib/storeSettings.svelte";
  import { storeUi } from "$lib/storeUi.svelte";

  import type { PageProps } from "./$types";
  import Node from "./components/Node.svelte";
  import NotFound from "./components/NotFound.svelte";
  import { createStoreNode } from "./storeNode.svelte";

  let { data }: PageProps = $props();

  let storeNode = $state(createStoreNode(data.nodeSlug));
  $effect(() => {
    storeNode = createStoreNode(data.nodeSlug);
  });

  let nodes = $derived(getIrNodes());

  let foundNode = $derived.by((): IrNode | undefined => {
    for (const node of nodes) {
      if (data.nodeKind === "rpc") {
        if (node.kind !== "proc" && node.kind !== "stream") continue;
        if (node.rpcName !== data.rpcName) continue;
        if (slugify(node.name) !== slugify(data.nodeName)) continue;
        return node;
      }

      if (node.kind !== data.nodeKind) continue;

      const isDoc = node.kind === "doc";
      let nodeName = isDoc ? getMarkdownTitle(node.content) : node.name;
      nodeName = slugify(nodeName);

      if (data.nodeName === nodeName) return node;
    }

    return undefined;
  });

  let nodeExists = $derived(foundNode !== undefined);

  let name = $derived.by(() => {
    if (!foundNode) return "unknown";
    if (foundNode.kind === "type") return foundNode.name;
    if (foundNode.kind === "proc") return foundNode.name;
    if (foundNode.kind === "stream") return foundNode.name;
    if (foundNode.kind === "doc") {
      return getMarkdownTitle(foundNode.content);
    }

    return "unknown";
  });

  let humanKind = $derived.by(() => {
    if (!foundNode) return "unknown";
    if (foundNode.kind === "type") return "type";
    if (foundNode.kind === "proc") return "procedure";
    if (foundNode.kind === "stream") return "stream";
    if (foundNode.kind === "doc") return "documentation";
    return "unknown";
  });

  let title = $derived.by(() => {
    if (!nodeExists) return "VDL Playground";

    return `${name} ${humanKind} | VDL Playground`;
  });

  $effect(() => {
    data.nodeSlug;
    untrack(() => {
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

{#if nodeExists && foundNode && storeNode.status.ready}
  {#key data.nodeSlug}
    <Node node={foundNode} bind:storeNode />
  {/key}
{/if}
