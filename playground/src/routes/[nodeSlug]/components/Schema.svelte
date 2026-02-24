<script lang="ts">
  import { Expand, Minimize } from "@lucide/svelte";

  import { type IrNode, storeSettings } from "$lib/storeSettings.svelte";
  import { storeUi } from "$lib/storeUi.svelte";
  import { wasmClient } from "$lib/wasm/index";

  import Code from "$lib/components/Code.svelte";
  import H2 from "$lib/components/H2.svelte";
  import Tabs from "$lib/components/Tabs.svelte";

  interface Props {
    node: IrNode;
  }

  let { node }: Props = $props();

  let isProcOrStream = $derived(node.kind === "proc" || node.kind === "stream");
  let isCompact = $derived(storeUi.store.schemaViewMode === "compact");
  let isExtended = $derived(storeUi.store.schemaViewMode === "expanded");
  let extractedSchemaCompact = $state("");
  let extractedSchemaExpanded = $state("");

  $effect(() => {
    if (node.kind === "doc") return;
    if (isCompact && extractedSchemaCompact !== "") return;
    if (isExtended && extractedSchemaExpanded !== "") return;

    (async () => {
      const schemaCompact = storeSettings.store.vdlSchema;
      const schemaExpanded = storeSettings.store.vdlSchemaExpanded;

      if (node.kind === "type") {
        if (isCompact) {
          extractedSchemaCompact = await wasmClient.extractType(
            schemaCompact,
            node.name,
          );
        }
        if (isExtended) {
          extractedSchemaExpanded = await wasmClient.extractType(
            schemaExpanded,
            node.name,
          );
        }
      }
      if (node.kind === "proc") {
        if (isCompact) {
          extractedSchemaCompact = await wasmClient.extractProc(
            schemaCompact,
            node.rpcName,
            node.name,
          );
        }
        if (isExtended) {
          extractedSchemaExpanded = await wasmClient.extractProc(
            schemaExpanded,
            node.rpcName,
            node.name,
          );
        }
      }
      if (node.kind === "stream") {
        if (isCompact) {
          extractedSchemaCompact = await wasmClient.extractStream(
            schemaCompact,
            node.rpcName,
            node.name,
          );
        }
        if (isExtended) {
          extractedSchemaExpanded = await wasmClient.extractStream(
            schemaExpanded,
            node.rpcName,
            node.name,
          );
        }
      }
    })();
  });

  let schemaToShow = $derived.by(() => {
    if (isCompact) return extractedSchemaCompact;
    if (isExtended) return extractedSchemaExpanded;
    return "";
  });
</script>

{#if schemaToShow !== ""}
  <div
    class={{
      "space-y-4": true,
      "max-w-5xl": !isProcOrStream,
    }}
  >
    <div class="flex items-center justify-between gap-4">
      <H2>Schema</H2>
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

    <Code lang="vdl" code={schemaToShow} />
  </div>
{/if}
