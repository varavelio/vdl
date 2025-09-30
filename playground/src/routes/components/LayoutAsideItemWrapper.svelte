<script lang="ts">
  import { getMarkdownTitle } from "$lib/helpers/getMarkdownTitle";
  import type { storeSettings } from "$lib/storeSettings.svelte";
  import { storeUi } from "$lib/storeUi.svelte";
  import type { Schema } from "$lib/urpcTypes";

  import LayoutAsideItem from "./LayoutAsideItem.svelte";

  interface Props {
    node: (typeof storeSettings.store.jsonSchema.nodes)[number];
  }

  const { node }: Props = $props();

  function getNodeName(node: Schema["nodes"][number]) {
    if (node.kind === "type") return node.name;
    if (node.kind === "proc") return node.name;
    if (node.kind === "stream") return node.name;
    if (node.kind === "doc") return getMarkdownTitle(node.content);
    return "unknown";
  }

  let showNode = $derived.by(() => {
    if (storeUi.store.asideSearchOpen) {
      if (storeUi.store.asideSearchQuery === "") return true;

      // Do the search
      const name = getNodeName(node).toLowerCase();
      const query = storeUi.store.asideSearchQuery.toLowerCase();
      return name.includes(query);
    }

    if (node.kind === "doc") return !storeUi.store.asideHideDocs;
    if (node.kind === "type") return !storeUi.store.asideHideTypes;
    if (node.kind === "proc") return !storeUi.store.asideHideProcs;
    if (node.kind === "stream") return !storeUi.store.asideHideStreams;

    return false;
  });
</script>

{#if showNode}
  {#key showNode}
    <LayoutAsideItem {node} />
  {/key}
{/if}
