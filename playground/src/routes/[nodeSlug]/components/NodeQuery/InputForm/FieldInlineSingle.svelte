<!--
  This component handles the case where a field is an inline single object, it acts only
  as a container for preparing and rendering its sub-fields.

  It iterates over the fields of the inline object and renders a Field component for each sub-field.
-->

<script lang="ts">
  import { BrushCleaning, Trash } from "@lucide/svelte";
  import { set, unset } from "lodash-es";
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
    disableDelete?: boolean;
  }

  let { field, input = $bindable(), path, disableDelete }: Props = $props();

  function clearObject() {
    input = set(input, path, {});
  }

  function deleteObject() {
    unset(input, path);
  }
</script>

<CommonFieldset>
  <legend class="fieldset-legend"><CommonLabel label={path} optional={field.optional} /></legend>

  <CommonFieldDoc doc={field.doc} class="-mt-2" />

  {#each field.typeRef.objectFields! as childField}
    <Field field={childField} path={`${path}.${childField.name}`} bind:input />
  {/each}

  <div class="flex justify-end">
    <Tooltip content={`Clear and reset ${path} to an empty object`} placement="left">
      <button class="btn btn-sm btn-ghost btn-square" onclick={clearObject}>
        <BrushCleaning class="size-4" />
      </button>
    </Tooltip>

    {#if !disableDelete}
      <Tooltip content={`Delete ${path} from the JSON object`} placement="left">
        <button class="btn btn-sm btn-ghost btn-square" onclick={deleteObject}>
          <Trash class="size-4" />
        </button>
      </Tooltip>
    {/if}
  </div>
</CommonFieldset>
