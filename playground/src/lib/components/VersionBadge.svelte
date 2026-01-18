<script lang="ts">
  import { ExternalLink } from "@lucide/svelte";

  import { mergeClasses } from "$lib/helpers/mergeClasses";
  import { versionWithPrefix } from "$lib/version";

  interface Props {
    class?: string;
  }

  let { class: className = "" }: Props = $props();
</script>

<a
  href="https://github.com/varavelio/vdl/releases/tag/{versionWithPrefix}"
  target="_blank"
  class={mergeClasses("btn btn-ghost flex items-center", className)}
>
  <ExternalLink class="size-4" />
  <span>
    UFO RPC {versionWithPrefix}
  </span>
</a>
