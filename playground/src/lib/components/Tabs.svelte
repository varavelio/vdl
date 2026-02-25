<script lang="ts" generics="T extends string">
  import type { IconProps } from "@lucide/svelte";
  import type { Component } from "svelte";

  import { type ClassValue, mergeClasses } from "$lib/helpers/mergeClasses";

  import Tooltip from "./Tooltip.svelte";

  export interface TabItem<T extends string = string> {
    id: T;
    label?: string;
    icon?: Component<IconProps, {}, "">;
    action?: () => void;
    tooltipText?: string;
  }

  interface Props<T extends string> {
    items: TabItem<T>[];
    active?: T;
    onSelect?: (id: T) => void;
    containerClass?: ClassValue;
    buttonClass?: ClassValue;
    activeButtonClass?: ClassValue;
    inactiveButtonClass?: ClassValue;
    iconClass?: ClassValue;
    spanTextClass?: ClassValue;
  }

  let {
    items,
    active = $bindable(),
    onSelect,
    containerClass,
    buttonClass,
    activeButtonClass,
    inactiveButtonClass,
    iconClass,
    spanTextClass,
  }: Props<T> = $props();

  const handleSelect = (id: T) => {
    active = id;
    if (onSelect) onSelect(id);
  };

  const handleClick = (id: T) => {
    const tab = items.find((t) => t.id === id);

    if (tab?.action) {
      tab.action();
      return;
    }

    handleSelect(id);
  };
</script>

<div
  class={mergeClasses(
    "join bg-base-100 flex w-full overflow-x-auto overflow-y-hidden",
    containerClass,
  )}
>
  {#each items as tab}
    {#snippet buttonContent()}
      {#if tab.icon}
        <tab .icon class={mergeClasses("size-4", iconClass)} />
      {/if}
      {#if tab.label}
        <span class={mergeClasses(spanTextClass)}> {tab.label} </span>
      {/if}
    {/snippet}

    {#snippet button()}
      <button
        class={[
          "btn join-item border-base-content/20 flex-grow",
          buttonClass,
          active === tab.id && "btn-primary",
          active === tab.id && activeButtonClass,
          active !== tab.id && inactiveButtonClass,
        ]}
        onclick={() => handleClick(tab.id)}
        type="button"
      >
        {@render buttonContent()}
      </button>
    {/snippet}

    {#if tab.tooltipText}
      <Tooltip content={tab.tooltipText} placement="top"> {@render button()} </Tooltip>
    {:else}
      {@render button()}
    {/if}
  {/each}
</div>
