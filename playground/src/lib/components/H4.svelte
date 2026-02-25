<script lang="ts">
  import type { Snippet } from "svelte";
  import type { ClassValue } from "$lib/helpers/mergeClasses";
  import { mergeClasses } from "$lib/helpers/mergeClasses";

  interface Props {
    children: Snippet;
    class?: ClassValue;
    // biome-ignore lint/suspicious/noExplicitAny: can be any other attribute
    rest?: any;
  }

  let { children, class: className, ...rest }: Props = $props();
</script>

<h4 class={mergeClasses("text-lg font-bold", className)} {...rest}>{@render children()}</h4>
