<script lang="ts">
  import { type ClassValue, mergeClasses } from "$lib/helpers/mergeClasses";

  interface Props {
    class?: ClassValue;
    // biome-ignore lint/suspicious/noExplicitAny: can be any other attribute
    rest?: any;
  }

  let { class: className, ...rest }: Props = $props();
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  class={mergeClasses("vdl-logo", className)}
  viewBox="0 0 940 275"
  {...rest}
>
  <path fill="currentColor" d="M680 275v-70h25V70h-25V0h145v70h-25v135h25l20-25h95v95z" />
  <path
    fill="currentColor"
    d="M360 0v70h25v135h-25v70l229.188.063L645 205V70L588.674 0Zm120 70h50l20 25v85l-20 25h-50z"
  />
  <path
    fill="currentColor"
    d="M25 70H0V0h145v70h-25l42.368 134.146L205 70h-25V0h145v70h-25l-65 205H90Z"
  />
</svg>

<style>
  :global([data-theme="light"]) {
    .vdl-logo {
      color: #2962ff;
    }
  }
  :global([data-theme="dark"]) {
    .vdl-logo {
      color: #ffffff;
    }
  }
</style>
