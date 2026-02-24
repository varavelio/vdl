<script lang="ts">
  import { Download, Loader } from "@lucide/svelte";
  import { downloadZip } from "client-zip";
  import { toast } from "svelte-sonner";

  import { storeSettings } from "$lib/storeSettings.svelte";
  import { storeUi } from "$lib/storeUi.svelte";
  import { wasmClient } from "$lib/wasm/index";
  import type {
    CodegenOutput,
    CodegenOutputFile,
  } from "$lib/wasm/wasmtypes/types";

  let isGenerating: boolean = $state(false);

  let downloadFileName = $derived.by(() => {
    const lang = storeUi.store.codeSnippetsSdkLang;
    if (lang === "typescript") return "vdl-client-sdk.ts";
    if (lang === "go") return "vdl-client-sdk.go";
    if (lang === "dart") return "vdl-dart-client-sdk.zip";
    if (lang === "python") return "vdl-client-sdk.py";
    return "unknown";
  });

  function downloadSingleFile(file: CodegenOutputFile) {
    const blob = new Blob([file.content], { type: "text/plain" });
    const link = document.createElement("a");
    link.href = URL.createObjectURL(blob);
    link.download = downloadFileName;
    link.click();
    link.remove();
  }

  async function downloadMultipleFiles(files: CodegenOutputFile[]) {
    const zipInputFiles = files.map((file) => ({
      name: file.path,
      input: file.content,
      lastModified: new Date(),
    }));

    const blob = await downloadZip(zipInputFiles).blob();
    const link = document.createElement("a");
    link.href = URL.createObjectURL(blob);
    link.download = downloadFileName;
    link.click();
    link.remove();
  }

  async function generateAndDownload() {
    if (isGenerating) return;
    isGenerating = true;

    try {
      const lang = storeUi.store.codeSnippetsSdkLang;
      const vdlSchema = storeSettings.store.vdlSchema;

      let result: CodegenOutput;

      if (lang === "typescript") {
        result = await wasmClient.codegen({
          vdlSchema,
          target: {
            typescript: {
              output: "",
              genClient: true,
            },
          },
        });
      } else if (lang === "go") {
        const packageName =
          storeUi.store.codeSnippetsSdkGolangPackageName.trim() || "vdlclient";
        result = await wasmClient.codegen({
          vdlSchema,
          target: {
            go: {
              output: "",
              package: packageName,
              genClient: true,
            },
          },
        });
      } else if (lang === "dart") {
        result = await wasmClient.codegen({
          vdlSchema,
          target: {
            dart: {
              output: "",
            },
          },
        });
      } else if (lang === "python") {
        result = await wasmClient.codegen({
          vdlSchema,
          target: {
            python: {
              output: "",
            },
          },
        });
      } else {
        throw new Error(`Unsupported language: ${lang}`);
      }

      if (result.files.length === 1) {
        downloadSingleFile(result.files[0]);
      }
      if (result.files.length > 1) {
        await downloadMultipleFiles(result.files);
      }
    } catch (error) {
      console.error(error);
      toast.error("Failed to generate SDK", {
        description: String(error),
        duration: 5000,
      });
    } finally {
      isGenerating = false;
    }
  }
</script>

<div>
  {#if storeUi.store.codeSnippetsSdkLang === "go"}
    <label class="fieldset">
      <legend class="fieldset-legend">Go package name</legend>
      <input
        id="go-pkg"
        class="input w-full"
        placeholder="Package name..."
        bind:value={storeUi.store.codeSnippetsSdkGolangPackageName}
      />
      <div class="prose prose-sm text-base-content/50 max-w-none">
        The generated SDK and code examples will use this package name.
      </div>
    </label>
  {/if}

  {#if storeUi.store.codeSnippetsSdkLang === "dart"}
    <label class="fieldset">
      <legend class="fieldset-legend">Dart package name</legend>
      <input
        id="dart-pkg"
        class="input w-full"
        placeholder="Package name..."
        bind:value={storeUi.store.codeSnippetsSdkDartPackageName}
      />
      <div class="prose prose-sm text-base-content/50 max-w-none">
        The generated SDK and code examples will use this package name.
      </div>
    </label>
  {/if}

  <div class="fieldset">
    <legend class="fieldset-legend">Download SDK</legend>
    <button
      class="btn btn-primary btn-block"
      disabled={isGenerating}
      onclick={generateAndDownload}
      type="button"
    >
      {#if isGenerating}
        <Loader class="animate size-4 animate-spin" />
      {/if}
      {#if !isGenerating}
        <Download class="size-4" />
      {/if}
      <span>Download SDK</span>
    </button>
  </div>
</div>
