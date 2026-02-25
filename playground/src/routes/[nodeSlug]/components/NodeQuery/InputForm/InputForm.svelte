<script lang="ts">
  import { BookText, Braces, Copy, Info } from "@lucide/svelte";
  import Tabs from "$lib/components/Tabs.svelte";
  import Tooltip from "$lib/components/Tooltip.svelte";
  import { copyTextToClipboard } from "$lib/helpers/copyTextToClipboard";
  import type { Field as FieldType } from "$lib/storeSettings.svelte";
  import { storeUi } from "$lib/storeUi.svelte";

  import Field from "./Field.svelte";
  import JsonEditor from "./JsonEditor.svelte";

  interface Props {
    fields: FieldType[];
    input: Record<string, unknown>;
  }

  let { fields, input = $bindable() }: Props = $props();

  let isFormTab = $derived(storeUi.store.inputFormTab === "form");
  let isJsonTab = $derived(storeUi.store.inputFormTab === "json");

  async function copyToClipboard() {
    return copyTextToClipboard(JSON.stringify(input, null, 2));
  }
</script>

<div class="desk:flex-row desk:items-start mt-4 flex flex-col items-end justify-between gap-4">
  <div role="alert" class="alert alert-soft alert-info w-fit">
    <Info class="size-4" />
    <span>
      Requests are made from your browser and validations are performed on the server side
    </span>
  </div>

  <div class="flex items-center space-x-2">
    <Tabs
      containerClass="w-auto flex-none"
      buttonClass="btn-xs"
      iconClass="size-3"
      items={[
        {
          id: "form",
          label: "Form",
          icon: BookText,
          tooltipText: "Edit input using form editor",
        },
        {
          id: "json",
          label: "JSON",
          icon: Braces,
          tooltipText: "Edit input using JSON editor",
        },
        {
          id: "copy",
          icon: Copy,
          action: copyToClipboard,
          tooltipText: "Copy input to clipboard",
        },
      ]}
      bind:active={storeUi.store.inputFormTab}
    />
  </div>
</div>

{#if isFormTab}
  {#each fields as field}
    <Field {field} path={field.name} bind:input />
  {/each}
{/if}

{#if isJsonTab}
  <JsonEditor bind:input />
{/if}
