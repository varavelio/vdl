<script lang="ts">
  import { onDestroy, onMount, type Snippet } from "svelte";
  import Portal from "svelte-portal";
  import type { Instance, Props as TippyProps } from "tippy.js";
  import tippy from "tippy.js";
  import "tippy.js/dist/svg-arrow.css";
  import "tippy.js/dist/tippy.css";

  // the following props are set by the component and should
  // not be passed in from the parent
  type customTippyProps = Omit<TippyProps, "arrow" | "appendTo" | "triggerTarget" | "plugins">;

  export interface Props extends Partial<customTippyProps> {
    children: Snippet;
  }

  let { children, ...tippyProps }: Props = $props();

  let hiddenEl: HTMLTemplateElement | undefined = $state(undefined);
  let arrow: SVGElement | undefined = $state(undefined);
  let tippyInstance: Instance<TippyProps> | undefined = $state(undefined);

  /**
   * Waits for the hiddenEl and arrow to be defined, with a maximum number of attempts
   */
  async function waitForElements(): Promise<boolean> {
    const maxAttempts = 20;
    let counter = 0;

    while (!hiddenEl || !arrow) {
      if (counter >= maxAttempts) {
        console.error("Tooltip: hiddenEl or arrow not found");
        return false;
      }
      counter++;
      await new Promise((resolve) => setTimeout(resolve, 50));
    }

    return true;
  }

  onMount(async () => {
    let elsExists = await waitForElements();
    if (!elsExists || !hiddenEl || !arrow) return;

    // Find the next sibling element of the hidden element
    const el = hiddenEl.nextElementSibling;
    if (!el) return;

    // Delete the hidden element from the DOM
    hiddenEl.remove();
    hiddenEl = undefined;

    // Initialize tippy.js on the target element
    tippyInstance = tippy(el, {
      ...tippyProps,
      arrow,
      appendTo: document.body,
      triggerTarget: el,
    });
  });

  onDestroy(() => {
    if (tippyInstance) {
      tippyInstance.destroy();
      tippyInstance = undefined;
    }
  });
</script>

<!-- 
  This element does not render anything, it's just used to reference
  the next sibling element as the tooltip target

  It's only used when the component is mounted, then is deleted from the DOM
-->
<template bind:this={hiddenEl}></template>

{#if children}
  {@render children()}
{/if}

<!-- 
  The arrow is rendered in a portal to the body, so it can be used
  by tippy.js for the tooltip arrow and it never affects the layout
-->
<Portal target="body">
  <template>
    <svg width="16" height="6" bind:this={arrow}>
      <path
        class="svg-arrow-border fill-base-content/30"
        d="M0 6s1.796-.013 4.67-3.615C5.851.9 6.93.006 8 0c1.07-.006 2.148.887 3.343 2.385C14.233 6.005 16 6 16 6H0z"
      />
      <path
        class="svg-arrow fill-base-200"
        d="m0 7s2 0 5-4c1-1 2-2 3-2 1 0 2 1 3 2 3 4 5 4 5 4h-16z"
      />
    </svg>
  </template>
</Portal>

<style lang="postcss">
  @reference "tailwindcss";
  @plugin "daisyui";

  :global(.tippy-box) {
    @apply bg-base-200 text-base-content border-base-content/20 border;
  }
</style>
