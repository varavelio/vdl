<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import { CircleCheck, CircleDashed, CircleX, Loader } from "@lucide/svelte";
  import { onMount } from "svelte";
  import { toast, Toaster } from "svelte-sonner";
  import { fade } from "svelte/transition";

  import { initializeShiki } from "$lib/shiki";
  import { loadVdlSchema, storeSettings } from "$lib/storeSettings.svelte";
  import {
    dimensionschangeAction,
    initTheme,
    storeUi,
  } from "$lib/storeUi.svelte";
  import { wasmClient } from "$lib/wasm";

  import Logo from "$lib/components/Logo.svelte";

  import "../app.css";

  import LayoutAside from "./components/LayoutAside.svelte";
  import LayoutHeader from "./components/LayoutHeader.svelte";
  import LayoutSwaggerSwitch from "./components/LayoutSwaggerSwitch.svelte";

  let { children } = $props();

  type LoadingStep = {
    id: string;
    label: string;
    status: "pending" | "loading" | "completed" | "error";
  };

  // Initialize playground
  let initialized = $state(false);
  let loadingSteps = $state<LoadingStep[]>([
    { id: "wasm", label: "Load WebAssembly binary", status: "loading" },
    { id: "highlighter", label: "Load Code highlighter", status: "pending" },
    { id: "config", label: "Load Configuration", status: "pending" },
    { id: "schema", label: "Load VDL Schema", status: "pending" },
  ]);

  function updateStepStatus(id: string, status: LoadingStep["status"]) {
    const step = loadingSteps.find((s) => s.id === id);
    if (step) step.status = status;
  }

  onMount(async () => {
    const handleError = (error: unknown) => {
      console.error(error);
      toast.error("Failed to initialize the Playground", {
        description: `Please try again or contact the developers if the problem persists. Error: ${error}`,
        duration: 15000,
      });
    };

    updateStepStatus("wasm", "loading");
    try {
      await wasmClient.init();
      updateStepStatus("wasm", "completed");
    } catch (error) {
      updateStepStatus("wasm", "error");
      handleError(error);
      return;
    }

    updateStepStatus("highlighter", "loading");
    try {
      await initializeShiki();
      updateStepStatus("highlighter", "completed");
    } catch (error) {
      updateStepStatus("highlighter", "error");
      handleError(error);
      return;
    }

    updateStepStatus("config", "loading");
    try {
      await Promise.all([
        storeSettings.status.waitUntilReady(),
        storeUi.status.waitUntilReady(),
      ]);
      initTheme();
      updateStepStatus("config", "completed");
    } catch (error) {
      updateStepStatus("config", "error");
      handleError(error);
      return;
    }

    updateStepStatus("schema", "loading");
    try {
      await loadVdlSchema("./schema.vdl");
      updateStepStatus("schema", "completed");
    } catch (error) {
      updateStepStatus("schema", "error");
      handleError(error);
      return;
    }

    initialized = true;
  });

  // Handle view transitions
  onNavigate((navigation) => {
    if (!document.startViewTransition) return;

    return new Promise((resolve) => {
      document.startViewTransition(async () => {
        resolve();
        await navigation.complete;
      });
    });
  });

  let mainWidth = $derived.by(() => {
    if (storeUi.store.isMobile) return storeUi.store.app.size.offsetWidth;
    return (
      storeUi.store.app.size.offsetWidth - storeUi.store.aside.size.offsetWidth
    );
  });

  let mainHeight = $derived.by(() => {
    return (
      storeUi.store.app.size.offsetHeight -
      storeUi.store.header.size.offsetHeight
    );
  });

  let mainStyle = $derived.by(() => {
    return `width: ${mainWidth}px; height: ${mainHeight}px;`;
  });
</script>

{#if !initialized}
  <main
    out:fade={{ duration: 200 }}
    class="fixed top-0 left-0 flex h-[100dvh] w-[100dvw] flex-col items-center justify-center"
  >
    <Logo class="mb-12 h-auto w-[90dvw] max-w-[600px]" />

    <div class="space-y-2">
      {#each loadingSteps as step (step.id)}
        <div class="flex items-center space-x-2">
          <div class="flex-shrink-0">
            {#if step.status === "pending"}
              <CircleDashed class="size-4" />
            {:else if step.status === "loading"}
              <Loader class="size-4 animate-spin" />
            {:else if step.status === "completed"}
              <CircleCheck class="text-success size-4" />
            {:else if step.status === "error"}
              <CircleX class="text-error size-4" />
            {/if}
          </div>

          <span class="text-sm font-medium">{step.label}</span>
        </div>
      {/each}
    </div>
  </main>
{/if}

{#if initialized}
  <div
    transition:fade={{ duration: 200 }}
    use:dimensionschangeAction
    ondimensionschange={(e) => (storeUi.store.app = e.detail)}
    class="flex h-[100dvh] w-[100dvw] justify-start"
  >
    <LayoutAside />
    <div
      use:dimensionschangeAction
      ondimensionschange={(e) => (storeUi.store.contentWrapper = e.detail)}
      class="h-[100dvh] flex-grow scroll-p-[90px]"
    >
      <LayoutHeader />
      <main
        class="overflow-hidden"
        style={mainStyle}
        use:dimensionschangeAction
        ondimensionschange={(e) => (storeUi.store.main = e.detail)}
      >
        {@render children()}
      </main>
    </div>
  </div>

  <!-- Requires a space at bottom of the page to fit the switch button without covering other content -->
  <LayoutSwaggerSwitch />
{/if}

<Toaster richColors closeButton duration={5000} />
