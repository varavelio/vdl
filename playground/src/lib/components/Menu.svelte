<script lang="ts">
  import type { Snippet } from "svelte";
  import type { Props as TooltipProps } from "./Tooltip.svelte";
  import Tooltip from "./Tooltip.svelte";

  interface Props extends Omit<TooltipProps, "content"> {
    content: Snippet;
  }

  let {
    content,
    children,
    interactive = true,
    trigger = "click",
    ...tooltipProps
  }: Props = $props();

  let contentWrapper: HTMLDivElement | undefined = $state(undefined);
</script>

<Tooltip content={contentWrapper ?? ""} {interactive} {trigger} {...tooltipProps}>
  {@render children()}
</Tooltip>

<div bind:this={contentWrapper}>{@render content()}</div>
