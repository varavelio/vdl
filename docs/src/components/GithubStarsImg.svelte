<script lang="ts">
  interface Props {
    class?: string;
  }

  const { class: className = "" }: Props = $props();
</script>

<img
  alt="GitHub Repo stars"
  src="https://img.shields.io/github/stars/varavelio/vdl?style=plastic&label=%20"
  class={className}
>
