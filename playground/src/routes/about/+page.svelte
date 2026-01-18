<script lang="ts">
  import { BookOpenText, Github } from "@lucide/svelte";

  import { storeUi } from "$lib/storeUi.svelte";

  import BottomSpace from "$lib/components/BottomSpace.svelte";
  import Logo from "$lib/components/Logo.svelte";
  import VersionBadge from "$lib/components/VersionBadge.svelte";
</script>

<svelte:head>
  <title>About | UFO RPC Playground</title>
</svelte:head>

<div
  class={{
    "h-full overflow-y-auto p-4": true,
    "pt-16": !storeUi.store.isMobile,
    "pt-8": storeUi.store.isMobile,
  }}
>
  <div class="container mx-auto space-y-8 pb-8">
    {#if !storeUi.store.isMobile}
      <Logo class="mx-auto h-full max-w-[600px]" />
    {/if}

    <div class="flex items-center justify-center space-x-2">
      <a
        href="https://github.com/varavelio/vdl"
        target="_blank"
        class={{
          "btn btn-primary": true,
          "btn-lg": !storeUi.store.isMobile,
        }}
      >
        <Github class="size-5" />
        View on GitHub
      </a>
      <a
        href="https://vdl.varavel.com"
        target="_blank"
        class={{
          "btn btn-outline": true,
          "btn-lg": !storeUi.store.isMobile,
        }}
      >
        <BookOpenText class="size-5" />
        Documentation
      </a>
    </div>

    <div class="flex justify-center">
      <VersionBadge />
    </div>

    <div class="card bg-base-200 shadow-sm">
      <div class="card-body">
        <h2 class="card-title mb-4 text-2xl">License</h2>
        <div class="space-y-4">
          <p class="text-base-content/80 leading-relaxed">
            UFO RPC is 100% open source and uses a dual-license approach:
          </p>

          <ul class="list-inside list-disc">
            <li>
              <a
                href="https://vdl.varavel.com/license"
                target="_blank"
                class="link"
              >
                VDL Generator: MIT
              </a>
              - Ensures this project remains free and open source forever
            </li>
            <li>
              <a
                href="https://vdl.varavel.com/license"
                target="_blank"
                class="link"
              >
                Generated Code: MIT
              </a>
              - Your generated code is yours to use freely in any way you want
            </li>
          </ul>

          <p class="text-base-content/80 leading-relaxed">
            VDL is provided "AS IS" without warranty of any kind. The
            authors assume no responsibility for damages or losses from using
            this software. You are responsible for testing generated code before
            production use.
          </p>
        </div>
      </div>
    </div>

    <div class="card bg-base-200 shadow-sm">
      <div class="card-body">
        <h2 class="card-title mb-4 text-2xl">Dependencies</h2>
        <div class="space-y-4">
          <p class="text-base-content/80 leading-relaxed">
            VDL is possible thanks to the following awesome open-source
            projects
          </p>

          <ul class="list-inside list-disc">
            <li>
              <a
                href="https://github.com/varavelio/vdl/blob/main/urpc/go.mod"
                target="_blank"
                class="link"
              >
                URPC dependencies
              </a>
            </li>
            <li>
              <a
                href="https://github.com/varavelio/vdl/blob/main/playground/package.json"
                target="_blank"
                class="link"
              >
                Playground dependencies
              </a>
            </li>
          </ul>
        </div>
      </div>
    </div>

    <div class="card bg-base-200 shadow-sm">
      <div class="card-body">
        <h2 class="card-title mb-4 text-2xl">Author</h2>
        <div class="space-y-4">
          <p class="text-base-content/80 leading-relaxed">
            This project was created and is maintained with ❤️ by
            <a
              href="https://vdl.varavel.com/author"
              target="_blank"
              class="link"
            >
              Luis Eduardo
            </a>
          </p>

          <ul class="list-inside list-disc">
            <li>
              <a
                href="https://vdl.varavel.com/author"
                target="_blank"
                class="link"
              >
                Website
              </a>
            </li>
            <li>
              <a
                href="https://vdl.varavel.com/author/github"
                target="_blank"
                class="link"
              >
                GitHub
              </a>
            </li>
            <li>
              <a
                href="https://vdl.varavel.com/author/linkedin"
                target="_blank"
                class="link"
              >
                LinkedIn
              </a>
            </li>
            <li>
              <a
                href="https://vdl.varavel.com/author/x"
                target="_blank"
                class="link"
              >
                X (Twitter)
              </a>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>

  <BottomSpace />
</div>
