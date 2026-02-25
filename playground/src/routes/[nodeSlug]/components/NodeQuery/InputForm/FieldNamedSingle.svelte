<!--
  This component handles the final single field rendering for a named type that is not an array.

  It handles primitive types only: string, int, float, bool, datetime

  It should handle reactivity and binding of default values correctly.
-->

<script lang="ts">
  import { BrushCleaning, EllipsisVertical, Trash } from "@lucide/svelte";
  import flatpickr from "flatpickr";
  import { get, set, unset } from "lodash-es";
  import { onMount, untrack } from "svelte";
  import Menu from "$lib/components/Menu.svelte";
  import Tooltip from "$lib/components/Tooltip.svelte";
  import type { Field as FieldType } from "$lib/storeSettings.svelte";

  import CommonFieldDoc from "./CommonFieldDoc.svelte";
  import CommonLabel from "./CommonLabel.svelte";

  interface Props {
    path: string;
    field: FieldType;
    input: Record<string, unknown>;
    disableDelete?: boolean;
  }

  let { field, input = $bindable(), path, disableDelete }: Props = $props();
  const fieldId = $props.id();

  // biome-ignore lint/suspicious/noExplicitAny: localValue can be any primitive type
  let localValue = $state<any>(get(input, path) ?? undefined);

  $effect(() => {
    const reactiveLocalValue = localValue;
    untrack(() => {
      if (get(input, path) !== reactiveLocalValue) {
        set(input, path, reactiveLocalValue);
      }
    });
  });

  $effect(() => {
    const reactiveInputValue = get(input, path);
    untrack(() => {
      if (reactiveInputValue !== localValue) {
        localValue = reactiveInputValue;
      }
    });
  });

  export const deleteValue = () => {
    if (flatpickrInstance) flatpickrInstance.clear();
    localValue = undefined;
    unset(input, path);
  };

  export const clearValue = () => {
    const typeName = field.typeRef.primitiveName;
    if (typeName === "string") localValue = "";
    if (typeName === "int") localValue = 0;
    if (typeName === "float") localValue = 0;
    if (typeName === "bool") localValue = false;
    if (typeName === "datetime") {
      let now = new Date();
      if (flatpickrInstance) flatpickrInstance.setDate(now);
      localValue = now;
    }
  };

  let inputType = $derived.by(() => {
    const typeName = field.typeRef.primitiveName;
    if (!typeName) return "text";

    if (typeName === "string") return "text";
    if (typeName === "int" || typeName === "float") return "number";
    if (typeName === "bool") return "checkbox";
    if (typeName === "datetime") return "datetime";

    return "text";
  });

  let inputStep = $derived.by(() => {
    const typeName = field.typeRef.primitiveName;
    if (typeName === "float") return 0.01;
    if (typeName === "int") return 1;
    return undefined;
  });

  let flatpickrInstance: flatpickr.Instance | null = $state(null);
  onMount(() => {
    if (field.typeRef.primitiveName !== "datetime") return;
    let inst = flatpickr(`#${fieldId}`, {
      enableTime: true,
      enableSeconds: true,
      dateFormat: "Z",
      altInput: true,
      altFormat: "F j, Y H:i:S",
    });
    if (Array.isArray(inst)) inst = inst[0];
    flatpickrInstance = inst;
  });
</script>

{#snippet menuContent()}
  <div class="py-1">
    <Tooltip content={`Clear and reset ${path} to its default value`} placement="left">
      <button
        class="btn btn-ghost btn-block flex items-center justify-start space-x-2"
        onclick={clearValue}
      >
        <BrushCleaning class="size-4" />
        <span>Clear</span>
      </button>
    </Tooltip>

    {#if !disableDelete}
      <Tooltip content={`Delete ${path} from the JSON object`} placement="left">
        <button
          class="btn btn-ghost btn-block flex items-center justify-start space-x-2"
          onclick={deleteValue}
        >
          <Trash class="size-4" />
          <span>Delete</span>
        </button>
      </Tooltip>
    {/if}
  </div>
{/snippet}

{#snippet menu()}
  <Menu content={menuContent} placement="bottom" trigger="mouseenter click">
    <button class="btn btn-ghost btn-square"><EllipsisVertical class="size-4" /></button>
  </Menu>
{/snippet}

<div>
  <label class="group/field block w-full">
    <span class="mb-1 block font-semibold">
      <CommonLabel optional={field.optional} label={path} />
    </span>

    {#if inputType !== "checkbox" && inputType !== "datetime"}
      <div class="mb-1 flex items-center justify-start">
        <input
          type={inputType}
          step={inputStep}
          bind:value={localValue}
          class="input group-hover/field:border-base-content/50 mr-1 flex-grow"
          placeholder={`Enter ${path} here...`}
        >

        {@render menu()}
      </div>
    {/if}

    {#if inputType === "datetime"}
      <div class="mb-1 flex items-center justify-start">
        <input
          id={fieldId}
          type={inputType}
          step={inputStep}
          bind:value={localValue}
          class="input group-hover/field:border-base-content/50 mr-1 flex-grow"
          placeholder={`Enter ${path} here...`}
        >

        {@render menu()}
      </div>
      <div class="prose prose-sm text-base-content/50 max-w-none font-bold">
        Time is shown in your local timezone and will be sent as UTC
      </div>
    {/if}

    {#if inputType === "checkbox"}
      <div class="flex items-center justify-start space-x-2">
        <input type="checkbox" bind:checked={localValue} class="toggle toggle-lg">

        {@render menu()}
      </div>
    {/if}
  </label>

  <CommonFieldDoc doc={field.doc} />
</div>
