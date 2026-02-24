<script lang="ts">
  import { ChevronDown, ChevronRight } from "@lucide/svelte";
  import type { Snippet } from "svelte";
  import { slide } from "svelte/transition";

  import { storeUi } from "$lib/storeUi.svelte";

  import SnippetsSdkDownload from "./SnippetsSdkDownload.svelte";
  import SnippetsSdkSetup from "./SnippetsSdkSetup.svelte";
  import SnippetsSdkUsage from "./SnippetsSdkUsage.svelte";

  interface Props {
    type: "proc" | "stream";
    name: string;
  }

  const { type, name }: Props = $props();

  function toggleStep(step: "download" | "setup" | "usage") {
    if (storeUi.store.codeSnippetsSdkStep === step) {
      storeUi.store.codeSnippetsSdkStep = "";
      return;
    }
    storeUi.store.codeSnippetsSdkStep = step;
  }
</script>

<label class="fieldset mb-4">
  <legend class="fieldset-legend">Language</legend>
  <select
    id="sdk-generator-select"
    class="select w-full"
    bind:value={storeUi.store.codeSnippetsSdkLang}
  >
    <option value="typescript">TypeScript</option>
    <option value="go">Go</option>
    <option value="dart">Dart</option>
    <option value="python">Python</option>
  </select>
  <div class="prose prose-sm text-base-content/50 max-w-none">
    <b>Can't find your language?</b> No problem. You can still use the HTTP
    request snippets (Curl and others) to get started, or generate a client SDK
    using the
    <a href="./openapi.yaml" target="_blank" class="text-base-content/50">
      OpenAPI specification.
    </a>
  </div>
</label>

{#snippet step(
  isOpen: boolean,
  stepName: string,
  onToggle: () => void,
  children: Snippet,
)}
  <div class="rounded-box bg-base-200 border-base-content/20 border">
    <button
      class=" flex w-full cursor-pointer items-center justify-start space-x-2 px-4 py-2"
      onclick={onToggle}
    >
      {#if isOpen}
        <ChevronDown class="size-4" />
      {:else}
        <ChevronRight class="size-4" />
      {/if}
      <span>{stepName}</span>
    </button>

    {#if isOpen}
      <div
        class="border-base-content/20 border-t p-4"
        transition:slide={{ duration: 200 }}
      >
        {@render children()}
      </div>
    {/if}
  </div>
{/snippet}

{#snippet download()}
  <SnippetsSdkDownload />
{/snippet}

{#snippet setup()}
  <SnippetsSdkSetup />
{/snippet}

{#snippet usage()}
  <SnippetsSdkUsage {type} {name} />
{/snippet}

<div class="space-y-4">
  {@render step(
    storeUi.store.codeSnippetsSdkStep === "download",
    "1. Download SDK",
    () => toggleStep("download"),
    download,
  )}

  {@render step(
    storeUi.store.codeSnippetsSdkStep === "setup",
    "2. Setup SDK",
    () => toggleStep("setup"),
    setup,
  )}

  {@render step(
    storeUi.store.codeSnippetsSdkStep === "usage",
    "3. Usage example",
    () => toggleStep("usage"),
    usage,
  )}
</div>
