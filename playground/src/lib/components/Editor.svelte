<script lang="ts">
  import { Loader } from "@lucide/svelte";
  import loader from "@monaco-editor/loader";
  import { shikiToMonaco } from "@shikijs/monaco";
  import type * as Monaco from "monaco-editor/esm/vs/editor/editor.api";
  import { onMount } from "svelte";

  import { type ClassValue, mergeClasses } from "$lib/helpers/mergeClasses";
  import { darkTheme, getHighlighter, lightTheme } from "$lib/shiki";
  import { storeUi } from "$lib/storeUi.svelte";

  interface Props {
    lang: string;
    value: string;
    onChange?: (value: string) => void;
    class?: ClassValue;
    rest?: any;
  }

  let {
    lang = "vdl",
    value = $bindable(),
    onChange = undefined,
    class: className,
    ...rest
  }: Props = $props();

  let editorContainer: HTMLElement;
  let monaco: typeof Monaco | null = $state(null);
  let editor: Monaco.editor.IStandaloneCodeEditor | null = $state(null);
  let isLoading = $state(true);

  onMount(async () => {
    const [highlighter, monacoEditor] = await Promise.all([
      getHighlighter(),
      loader.init(),
    ]);

    monaco = monacoEditor;
    monaco.languages.register({ id: "vdl" });
    shikiToMonaco(highlighter, monaco);

    editor = monaco.editor.create(editorContainer, {
      value: value,
      language: lang,
      tabSize: 2,
      insertSpaces: true,
      padding: { top: 30, bottom: 30 },
      scrollbar: {
        alwaysConsumeMouseWheel: false,
      },
      scrollBeyondLastLine: false,
      automaticLayout: true,
      minimap: { enabled: false },
    });

    editor.onDidChangeModelContent(() => {
      value = editor?.getValue() ?? "";
      if (onChange) onChange(value);
    });

    isLoading = false;
  });

  // Effect that manages the editor's value
  $effect(() => {
    if (!editor) return;
    if (value === editor.getValue()) return;
    editor.setValue(value);
  });

  // Effect that manages the editor's theme
  $effect(() => {
    if (!monaco) return;

    const themeMap = {
      light: lightTheme,
      dark: darkTheme,
    };

    monaco.editor.setTheme(themeMap[storeUi.store.theme]);
  });
</script>

{#if isLoading}
  <div
    class={mergeClasses(
      className,
      "flex h-[400px] w-full items-center justify-center",
    )}
  >
    <Loader class="animate size-10 animate-spin" />
  </div>
{/if}

<div
  bind:this={editorContainer}
  class={mergeClasses(className, { hidden: isLoading })}
  {...rest}
></div>
