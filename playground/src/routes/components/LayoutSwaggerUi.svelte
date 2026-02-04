<script lang="ts">
  import { Download, X } from "@lucide/svelte";
  import SwaggerUI from "swagger-ui";
  import "swagger-ui/dist/swagger-ui.css";

  import Logo from "$lib/components/Logo.svelte";

  interface Props {
    onClose: () => void;
  }

  const { onClose }: Props = $props();

  const elId = $props.id();

  $effect(() => {
    SwaggerUI({
      domNode: document.getElementById(elId),
      url: "./openapi.yaml",
    });
  });
</script>

<header
  class={[
    "bg-base-100 border-base-content/20 flex justify-between border-b px-4 py-4",
    "sticky top-0 left-0 z-50",
  ]}
>
  <div class="flex-grow"></div>
  <Logo class="h-[32px] flex-grow" />
  <div class="flex flex-grow justify-end">
    <button class="btn btn-ghost btn-circle" onclick={onClose}>
      <X class="size-6" />
    </button>
  </div>
</header>

<p class="text-base-content/50 mt-8 block px-4 text-center text-sm">
  VDL is not affiliated with OpenAPI or Swagger. It just generates OpenAPI
  schemas from VDL code.
</p>

<div class="mt-2 flex justify-center">
  <a href="./openapi.yaml" class="btn btn-sm btn-ghost" download>
    <Download class="size-4" />
    <span>Download OpenAPI schema</span>
  </a>
</div>

<div id={elId}></div>

<!-- Space for the switch button to avoid hiding content behind it -->
<div class="h-[75px]"></div>

<style>
  :global([data-theme="dark"] .swagger-ui) {
    filter: invert(88%) hue-rotate(180deg);
  }
  :global([data-theme="dark"] .swagger-ui .microlight) {
    filter: invert(100%) hue-rotate(180deg);
  }
</style>
