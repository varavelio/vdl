<script lang="ts">
  import { TriangleAlert } from "@lucide/svelte";

  import { deleteMarkdownHeadings } from "$lib/helpers/deleteMarkdownHeadings";
  import { getMarkdownTitle } from "$lib/helpers/getMarkdownTitle";
  import { markdownToHtml } from "$lib/helpers/markdownToHtml";
  import { type IrNode } from "$lib/storeSettings.svelte";
  import { storeUi } from "$lib/storeUi.svelte";

  import BottomSpace from "$lib/components/BottomSpace.svelte";

  import type { StoreNodeInstance } from "../storeNode.svelte";

  import NodeQueryProc from "./NodeQuery/QueryProc.svelte";
  import NodeQueryStream from "./NodeQuery/QueryStream.svelte";
  import Snippets from "./NodeQuery/Snippets/Snippets.svelte";
  import Schema from "./Schema.svelte";

  interface Props {
    node: IrNode;
    storeNode: StoreNodeInstance;
  }

  let { node, storeNode = $bindable() }: Props = $props();

  let name = $derived.by(() => {
    if (node.kind === "type") return node.name;
    if (node.kind === "proc") return node.name;
    if (node.kind === "stream") return node.name;
    if (node.kind === "doc") {
      return getMarkdownTitle(node.content);
    }

    return "unknown";
  });

  let deprecatedMessage = $derived.by(() => {
    if (node.kind === "doc") return "";
    if (typeof node.deprecated !== "string") return "";
    if (node.deprecated !== "") return node.deprecated;
    return `This ${node.kind} is deprecated and it's use is not recommended`;
  });

  let documentation = $state("");
  $effect(() => {
    (async () => {
      if (node.kind === "doc") {
        documentation = await markdownToHtml(
          deleteMarkdownHeadings(node.content),
        );
      }
      if (
        (node.kind === "stream" ||
          node.kind === "type" ||
          node.kind === "proc") &&
        typeof node.doc === "string" &&
        node.doc !== ""
      ) {
        documentation = await markdownToHtml(deleteMarkdownHeadings(node.doc));
      }
    })();
  });

  let isProcOrStream = $derived(node.kind === "proc" || node.kind === "stream");
</script>

<div
  class={{
    "h-full overflow-hidden": true,
    "grid grid-cols-12": !storeUi.store.isMobile,
  }}
>
  <section
    class={{
      "h-full space-y-12 overflow-y-auto p-4 pt-0": true,
      "col-span-8": !storeUi.store.isMobile && isProcOrStream,
      "col-span-12": !storeUi.store.isMobile && !isProcOrStream,
    }}
  >
    <div
      class={{
        "prose pt-4": true,
        "max-w-none": isProcOrStream,
        "max-w-5xl": !isProcOrStream,
      }}
    >
      <h1 class="break-all">{name}</h1>

      {#if deprecatedMessage !== ""}
        <div
          role="alert"
          class="alert alert-soft alert-error w-fit gap-2 font-bold italic"
        >
          <TriangleAlert class="size-4" />
          <span>Deprecated: {deprecatedMessage}</span>
        </div>
      {/if}

      {#if documentation !== ""}
        {@html documentation}
      {/if}
    </div>

    {#if node.kind === "proc"}
      <div>
        <NodeQueryProc proc={node} bind:storeNode />
      </div>
    {/if}

    {#if node.kind === "stream"}
      <div>
        <NodeQueryStream stream={node} bind:storeNode />
      </div>
    {/if}

    <Schema {node} />

    {#if storeUi.store.isMobile || !isProcOrStream}
      <BottomSpace />
    {/if}
  </section>

  {#if !storeUi.store.isMobile && (node.kind == "proc" || node.kind == "stream")}
    <div class="col-span-4 overflow-y-auto p-4 pt-0">
      <Snippets
        type={node.kind}
        name={node.name}
        input={storeNode.store.input}
      />
      <BottomSpace />
    </div>
  {/if}
</div>
