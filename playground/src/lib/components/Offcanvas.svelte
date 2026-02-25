<script lang="ts">
  import type { Snippet } from "svelte";
  import { fade } from "svelte/transition";
  import Portal from "svelte-portal";
  import type { ClassValue } from "$lib/helpers/mergeClasses";
  import { mergeClasses } from "$lib/helpers/mergeClasses";

  interface Props {
    children?: Snippet;
    direction?: "left" | "right";
    isOpen?: boolean;
    class?: ClassValue;
    backdropClass?: ClassValue;
    backdropClose?: boolean;
    escapeClose?: boolean;
  }

  let {
    children,
    direction = "left",
    isOpen = $bindable(false),
    class: className,
    backdropClass: backdropClassName,
    backdropClose = true,
    escapeClose = true,
  }: Props = $props();

  let offcanvasWrapper: HTMLDivElement | undefined = $state(undefined);
  const closeOffcanvas = () => (isOpen = false);

  const handleEscapeKey = (event: KeyboardEvent) => {
    if (!escapeClose) return;
    if (event.key === "Escape") closeOffcanvas();
  };

  // Focus the offcanvas when it opens
  $effect(() => {
    if (!isOpen) return;
    offcanvasWrapper?.focus();
  });
</script>

{#if isOpen}
  <Portal target="body">
    <div
      class="fixed top-0 left-0 z-40 h-screen w-screen"
      transition:fade={{ duration: 100 }}
      onkeydown={handleEscapeKey}
      role="button"
      tabindex="0"
    >
      <button
        class={mergeClasses(
          "z-10 h-full w-full bg-black/30",
          backdropClassName,
        )}
        onclick={backdropClose ? closeOffcanvas : undefined}
        aria-label="Close modal"
      ></button>

      <div
        class={mergeClasses(
          "absolute top-0 z-20",
          "bg-base-100 h-[100dvh] w-[280px]",
          "overflow-x-hidden overflow-y-auto",
          {
            "left-0": direction === "left",
            "right-0": direction === "right",
          },
          className,
        )}
        bind:this={offcanvasWrapper}
        tabindex="-1"
      >
        {#if children}
          {@render children()}
        {/if}
      </div>
    </div>
  </Portal>
{/if}
