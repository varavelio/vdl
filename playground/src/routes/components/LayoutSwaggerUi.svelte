<script lang="ts">
  import { Download, X } from "@lucide/svelte";
  import { untrack } from "svelte";
  import { toast } from "svelte-sonner";
  import SwaggerUI from "swagger-ui";
  import "swagger-ui/dist/swagger-ui.css";

  import Logo from "$lib/components/Logo.svelte";
  import { storeSettings } from "$lib/storeSettings.svelte";

  interface Props {
    onClose: () => void;
  }

  const { onClose }: Props = $props();

  // Debounced states to prevent lag during typing
  let debouncedJson = $state(storeSettings.store.openApiJsonSchema);
  let debouncedYaml = $state(storeSettings.store.openApiYamlSchema);
  let debouncedBaseUrl = $state(storeSettings.store.baseUrl);

  let yamlSchemaURL = $state("");

  // Debounce logic: Updates internal state after 500ms of inactivity
  $effect(() => {
    const json = storeSettings.store.openApiJsonSchema;
    const yaml = storeSettings.store.openApiYamlSchema;
    const base = storeSettings.store.baseUrl;

    const timeout = setTimeout(() => {
      debouncedJson = json;
      debouncedYaml = yaml;
      debouncedBaseUrl = base;
    }, 1000);

    return () => clearTimeout(timeout);
  });

  // SwaggerUI Initialization
  $effect(() => {
    const jsonStr = debouncedJson;
    const yamlStr = debouncedYaml;
    const baseUrl = debouncedBaseUrl;
    const container = document.getElementById("swagger-container");

    if (container && jsonStr) {
      untrack(() => {
        try {
          container.innerHTML = "";
          const jsonSpec = JSON.parse(jsonStr);
          jsonSpec.servers = [{ url: baseUrl }];

          SwaggerUI({
            domNode: container,
            spec: jsonSpec,
          });

          // Update Download Blob
          if (yamlSchemaURL) URL.revokeObjectURL(yamlSchemaURL);
          const blob = new Blob([yamlStr], { type: "text/yaml" });
          yamlSchemaURL = URL.createObjectURL(blob);
        } catch (error) {
          console.error("Swagger render error:", error);
          toast.error("Failed to render OpenAPI spec");
        }
      });
    }

    return () => {
      if (yamlSchemaURL) URL.revokeObjectURL(yamlSchemaURL);
    };
  });
</script>

<header
  class={[
    "bg-base-100 border-base-content/20 flex justify-between border-b px-4 py-4",
    "sticky top-0 left-0 z-50",
  ]}
>
  <div class="grow"></div>
  <Logo class="h-8 grow" />
  <div class="flex grow justify-end">
    <button class="btn btn-ghost btn-circle" onclick={onClose}><X class="size-6" /></button>
  </div>
</header>

<p class="text-base-content/50 mt-8 block px-4 text-center text-sm">
  VDL is not affiliated with OpenAPI or Swagger. It just generates OpenAPI schemas from VDL code.
</p>

<div class="mt-2 flex justify-center">
  <a href={yamlSchemaURL} class="btn btn-sm btn-ghost" download="openapi.yaml">
    <Download class="size-4" />
    <span>Download OpenAPI schema</span>
  </a>
</div>

{#key debouncedBaseUrl}
  <div id="swagger-container"></div>
{/key}

<div class="h-18.75"></div>

<style>
  :global([data-theme="dark"] .swagger-ui) {
    filter: invert(88%) hue-rotate(180deg);
  }
  :global([data-theme="dark"] .swagger-ui .microlight) {
    filter: invert(100%) hue-rotate(180deg);
  }
</style>
