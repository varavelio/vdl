<script lang="ts">
  import { Copy, Download } from "@lucide/svelte";
  import { transformerColorizedBrackets } from "@shikijs/colorized-brackets";
  import { toast } from "svelte-sonner";

  import { copyTextToClipboard } from "$lib/helpers/copyTextToClipboard";
  import { getLangExtension } from "$lib/helpers/getLangExtension";
  import { mergeClasses } from "$lib/helpers/mergeClasses";
  import type { ClassValue } from "$lib/helpers/mergeClasses";
  import {
    darkTheme,
    getHighlighter,
    getOrFallbackLanguage,
    lightTheme,
  } from "$lib/shiki";
  import { storeUi } from "$lib/storeUi.svelte";

  interface Props {
    code: string;
    lang: "vdl" | string;
    class?: ClassValue;
    rounded?: boolean;
    withBorder?: boolean;
    scrollY?: boolean;
    scrollX?: boolean;
  }

  let {
    code,
    lang,
    class: className,
    rounded = true,
    withBorder = true,
    scrollY = true,
    scrollX = true,
  }: Props = $props();

  let codeHighlighted = $state("");
  $effect(() => {
    const themeMap = {
      dark: darkTheme,
      light: lightTheme,
    };
    let theme = themeMap[storeUi.store.theme];

    const codeToHighlight = code.trim();

    (async () => {
      const highlighter = await getHighlighter();
      codeHighlighted = highlighter.codeToHtml(codeToHighlight, {
        lang: getOrFallbackLanguage(lang),
        theme: theme,
        transformers: [transformerColorizedBrackets()],
      });
    })();
  });

  async function copyToClipboard() {
    return copyTextToClipboard(code);
  }

  const downloadCode = () => {
    try {
      // Create a blob from the code
      const blob = new Blob([code], { type: "text/plain" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");

      // Find extension from lang
      const extension = getLangExtension(lang);
      const fileName = `code.${extension}`;

      // Download the file
      a.href = url;
      a.download = fileName;
      a.click();

      toast.success("Code downloaded", { duration: 1500 });
    } catch (error) {
      console.error("Failed to download code: ", error);
      toast.error("Failed to download code", {
        description: `Error: ${error}`,
      });
    }
  };
</script>

{#if codeHighlighted !== ""}
  <div
    class={mergeClasses([
      "bg-base-200 flex h-full flex-col overflow-hidden",
      {
        "border-base-content/20 border": withBorder,
        "rounded-box": rounded,
      },
      className,
    ])}
  >
    <div
      class={{
        "w-full border-b px-4 py-2": true,
        "border-base-content/20 border-b": withBorder,
      }}
    >
      <div class="flex items-center justify-between space-x-1">
        <button
          class="btn btn-ghost btn-xs justify-end"
          onclick={() => downloadCode()}
        >
          <Download class="size-3" />
          <span>Download</span>
        </button>
        <button
          class="btn btn-ghost btn-xs justify-start"
          onclick={() => copyToClipboard()}
        >
          <Copy class="size-3" />
          <span>Copy to clipboard</span>
        </button>
      </div>
    </div>
    <div
      class={mergeClasses([
        "code-container flex-1",
        {
          "code-container-scroll-y": scrollY,
          "code-container-scroll-x": scrollX,
        },
      ])}
    >
      {@html codeHighlighted}
    </div>
  </div>
{/if}

<style lang="postcss">
  @reference "tailwindcss";
  @plugin "daisyui";

  .code-container-scroll-y {
    @apply overflow-y-auto;
  }

  .code-container-scroll-x {
    @apply overflow-x-auto;
  }

  .code-container {
    :global(pre) {
      @apply bg-base-200! rounded-box p-4;
    }

    :global(pre:focus-visible) {
      @apply outline-none;
    }

    /* 
      Classes to handle line numbers
      https://github.com/shikijs/shiki/issues/3
    */

    :global(code) {
      counter-reset: step;
      counter-increment: step 0;
    }

    :global(code .line::before) {
      content: counter(step);
      counter-increment: step;
      width: 1rem;
      margin-right: 1.5rem;
      display: inline-block;
      text-align: right;
      @apply text-base-content/40;
    }
  }
</style>
