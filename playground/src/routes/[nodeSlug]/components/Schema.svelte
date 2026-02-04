<script lang="ts">
  import { Expand, Minimize } from "@lucide/svelte";

  import { storeSettings } from "$lib/storeSettings.svelte";
  import { storeUi } from "$lib/storeUi.svelte";
  import { cmdExtractProc, cmdExtractStream, cmdExtractType } from "$lib/urpc";

  import Code from "$lib/components/Code.svelte";
  import H2 from "$lib/components/H2.svelte";
  import Tabs from "$lib/components/Tabs.svelte";

  interface Props {
    node: (typeof storeSettings.store.jsonSchema.nodes)[number];
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
      let extractFunc = async (_input: string, _name: string) => "";
      if (node.kind === "type") extractFunc = cmdExtractType;
      if (node.kind === "proc") extractFunc = cmdExtractProc;
      if (node.kind === "stream") extractFunc = cmdExtractStream;

      if (isCompact) {
        extractedSchemaCompact = await extractFunc(
          storeSettings.store.urpcSchema,
          node.name,
        );
      }

      if (isExtended) {
        extractedSchemaExpanded = await extractFunc(
          storeSettings.store.urpcSchemaExpanded,
          node.name,
        );
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
