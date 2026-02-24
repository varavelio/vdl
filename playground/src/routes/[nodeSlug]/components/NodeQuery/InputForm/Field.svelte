<!--
  This component is responsible for initiating the rendering of a field based on its type.
  It handles different scenarios based on the typeRef.kind:

    - primitive: A primitive type field (string, int, float, bool, datetime)
    - type: A named type reference (expands to its fields)
    - enum: An enum reference
    - array: An array of items
    - map: A map type
    - object: An inline object with nested fields

  If the named type is a custom type (not a primitive), it expands the underlying fields
  and treats it as an inline type for rendering purposes.
-->

<script lang="ts">
  import {
    type Field as FieldType,
    storeSettings,
  } from "$lib/storeSettings.svelte";

  import FieldInlineArray from "./FieldInlineArray.svelte";
  import FieldInlineSingle from "./FieldInlineSingle.svelte";
  import FieldNamedArray from "./FieldNamedArray.svelte";
  import FieldNamedSingle from "./FieldNamedSingle.svelte";

  interface Props {
    path: string;
    field: FieldType;
    input: Record<string, unknown>;
    disableDelete?: boolean;
  }

  let {
    field: originalField,
    input = $bindable(),
    path,
    disableDelete,
  }: Props = $props();

  function getTypeDefFields(typeName: string): FieldType[] {
    for (const typeDef of storeSettings.store.irSchema.types) {
      if (typeDef.name === typeName) {
        return typeDef.fields;
      }
    }
    return [];
  }

  let field = $derived.by((): FieldType => {
    const typeRef = originalField.typeRef;

    if (typeRef.kind === "type" && typeRef.typeName) {
      const typeFields = getTypeDefFields(typeRef.typeName);
      if (typeFields.length > 0) {
        return {
          ...originalField,
          typeRef: {
            kind: "object",
            objectFields: typeFields,
          },
        };
      }
    }

    if (
      typeRef.kind === "array" &&
      typeRef.arrayType?.kind === "type" &&
      typeRef.arrayType.typeName
    ) {
      const typeFields = getTypeDefFields(typeRef.arrayType.typeName);
      if (typeFields.length > 0) {
        return {
          ...originalField,
          typeRef: {
            kind: "array",
            arrayType: {
              kind: "object",
              objectFields: typeFields,
            },
          },
        };
      }
    }

    return originalField;
  });

  let isArray = $derived(field.typeRef.kind === "array");
  let isObject = $derived(field.typeRef.kind === "object");
  let isPrimitive = $derived(field.typeRef.kind === "primitive");

  let isInlineArray = $derived(
    isArray && field.typeRef.arrayType?.kind === "object",
  );
  let isInlineSingle = $derived(isObject && !isArray);
  let isNamedArray = $derived(
    isArray && field.typeRef.arrayType?.kind === "primitive",
  );
  let isNamedSingle = $derived(isPrimitive && !isArray);
</script>

{#if isInlineArray}
  <FieldInlineArray {field} {path} bind:input />
{/if}

{#if isInlineSingle}
  <FieldInlineSingle {field} {path} {disableDelete} bind:input />
{/if}

{#if isNamedArray}
  <FieldNamedArray {field} {path} bind:input />
{/if}

{#if isNamedSingle}
  <FieldNamedSingle {field} {path} {disableDelete} bind:input />
{/if}
