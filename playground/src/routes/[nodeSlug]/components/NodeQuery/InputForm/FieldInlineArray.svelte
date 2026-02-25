<!--
  This component handles the case where a field is an inline array of objects, it acts only
  as a container for preparing and rendering its sub-object-forms.

  It handles adding/removing items from the array, and iterates over the items
  rendering a Field component for each item.
-->

<script lang="ts">
  import { BrushCleaning, Minus, PackageOpen, Plus, Trash } from "@lucide/svelte";
  import { get, set, unset } from "lodash-es";
  import Tooltip from "$lib/components/Tooltip.svelte";
  import type { Field as FieldType } from "$lib/storeSettings.svelte";

  import CommonFieldDoc from "./CommonFieldDoc.svelte";
  import CommonFieldset from "./CommonFieldset.svelte";
  import CommonLabel from "./CommonLabel.svelte";
  import Field from "./Field.svelte";

  interface Props {
    path: string;
    field: FieldType;
    input: Record<string, unknown>;
  }

  let { field, input = $bindable(), path }: Props = $props();

  let noArrayField = $derived({
    ...field,
    doc: undefined,
    typeRef: field.typeRef.arrayType!,
  } as FieldType);
  let arrayLen = $derived((get(input, path) as unknown[])?.length || 0);
  let arrayIndexes = $derived(Array.from({ length: arrayLen }, (_, i) => i));
  let lastIndex = $derived(arrayIndexes[arrayIndexes.length - 1]);

  function clearArray() {
    input = set(input, path, []);
  }

  function deleteArray() {
    unset(input, path);
  }

  function removeItem() {
    if (arrayLen <= 0) return;
    unset(input, `${path}[${lastIndex}]`);
  }

  function addItem() {
    input = set(input, `${path}[${arrayLen}]`, null);
  }
</script>

<CommonFieldset>
  <legend class="fieldset-legend"><CommonLabel optional={field.optional} label={path} /></legend>

  <CommonFieldDoc doc={field.doc} class="-mt-2" />

  {#if arrayLen == 0}
    <PackageOpen class="mx-auto size-6" />
    <p class="text-center text-sm italic">No items, add one using the button below</p>
  {/if}

  {#each arrayIndexes as index}
    <Field field={noArrayField} path={`${path}[${index}]`} disableDelete={true} bind:input />
  {/each}

  <div class="flex justify-end">
    <Tooltip content={`Clear and reset ${path} to an empty array`} placement="left">
      <button class="btn btn-sm btn-ghost btn-square" onclick={clearArray}>
        <BrushCleaning class="size-4" />
      </button>
    </Tooltip>

    <Tooltip content={`Delete ${path} array from the JSON object`} placement="left">
      <button class="btn btn-sm btn-ghost btn-square" onclick={deleteArray}>
        <Trash class="size-4" />
      </button>
    </Tooltip>

    {#if arrayLen > 0}
      <Tooltip content={`Remove last item from ${path} array`} placement="left">
        <button class="btn btn-sm btn-ghost btn-square" onclick={removeItem}>
          <Minus class="size-4" />
        </button>
      </Tooltip>
    {/if}

    <Tooltip content={`Add item to ${path} array`} placement="left">
      <button class="btn btn-sm btn-ghost btn-square" onclick={addItem}>
        <Plus class="size-4" />
      </button>
    </Tooltip>
  </div>
</CommonFieldset>
