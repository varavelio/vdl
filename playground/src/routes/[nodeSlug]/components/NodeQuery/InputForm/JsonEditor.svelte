<script lang="ts">
  import { isEqual } from "lodash-es";
  import { untrack } from "svelte";

  import Editor from "$lib/components/Editor.svelte";

  interface Props {
    input: Record<string, any>;
  }

  let { input = $bindable() }: Props = $props();

  /**
   * The internal state of the editor. It is always a string.
   * It is initialized with the formatted value from the store.
   */
  let editorValue = $state(JSON.stringify(input, null, 2));

  /**
   * Function to compare if the two states are deeply equal.
   * This deeply compares if the two objects have the same content
   * regardless of formatting or key order.
   */
  function statesAreEqual(input: Record<string, any>, editorValue: string): boolean {
    try {
      const comparableInput = JSON.parse(JSON.stringify(input));
      const comparableEditorValue = JSON.parse(editorValue);
      return isEqual(comparableInput, comparableEditorValue);
    } catch {
      // If the editor value is not valid JSON, we consider them not equal.
      return false;
    }
  }

  /**
   * Effect to synchronize changes from the store (`input`) to the editor (`editorValue`).
   * This runs whenever `input` changes externally.
   */
  $effect(() => {
    const trackedInput = input;

    // From here we untrack to avoid tracking `editorValue` changes
    // as we only want to react to external input changes.
    untrack(() => {
      if (statesAreEqual(trackedInput, editorValue)) return;
      editorValue = JSON.stringify(trackedInput, null, 2);
    });
  });

  /**
   * Effect to synchronize changes from the editor (`editorValue`) to the store (`input`).
   * This runs whenever `editorValue` changes (i.e., when the user types).
   */
  $effect(() => {
    const trackedEditorValue = editorValue;

    // From here we untrack to avoid tracking `input` changes
    // as we only want to react to editor changes.
    untrack(() => {
      try {
        if (statesAreEqual(input, trackedEditorValue)) return;
        input = JSON.parse(trackedEditorValue) as Record<string, any>;
      } catch {
        // If the JSON is invalid, we do nothing.
        // The user is likely in the middle of typing.
      }
    });
  });
</script>

<Editor
  class="rounded-box h-[600px] w-full overflow-hidden shadow-md"
  lang="json"
  bind:value={editorValue}
/>
