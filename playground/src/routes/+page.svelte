<script lang="ts">
  import {
    ArrowRight,
    BookOpenText,
    FileJson,
    Github,
    Languages,
    Laptop,
    MonitorSmartphone,
  } from "@lucide/svelte";

  import { storeUi } from "$lib/storeUi.svelte";

  import BottomSpace from "$lib/components/BottomSpace.svelte";
  import Logo from "$lib/components/Logo.svelte";
  import VersionBadge from "$lib/components/VersionBadge.svelte";

  import UfoAbduction from "../lib/components/UfoAbduction.svelte";
</script>

<svelte:head>
  <title>UFO RPC Playground</title>
</svelte:head>

<div
  class={{
    "h-full overflow-y-auto p-4": true,
    "pt-16": !storeUi.store.isMobile,
  }}
>
  <div class="container mx-auto">
    <div class="mb-16 text-center">
      {#if !storeUi.store.isMobile}
        <div class="mb-8">
          <Logo class="mx-auto h-full max-w-[600px]" />
        </div>
      {/if}

      <p
        class={{
          "mx-auto mb-8 max-w-2xl text-xl": true,
          "text-base-content/70": !storeUi.store.isMobile,
        }}
      >
        Test and experiment with UFO RPC in this interactive playground using
        the left sidebar to navigate the schema.
      </p>

      <div class="mb-8 flex items-center justify-center">
        <UfoAbduction width={200} />
      </div>

      <div class="mb-8 flex flex-wrap items-center justify-center gap-2">
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
          href="https://uforpc.uforg.dev"
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
    </div>

    <div class="desk:grid-cols-2 mb-16 grid gap-8">
      <div class="card bg-base-300 shadow-md">
        <div class="card-body">
          <h2 class="card-title mb-4 text-2xl">What is UFO RPC?</h2>
          <p class="text-base-content/80 leading-relaxed">
            UFO RPC is a DX-focused RPC system that provides type safety and
            cross-language compatibility. It includes its own DSL (Domain
            Specific Language) that makes it easy to define and validate RPC
            services with a clean, intuitive syntax
          </p>
        </div>
      </div>

      <div class="card bg-base-300 shadow-md">
        <div class="card-body">
          <h2 class="card-title mb-4 text-2xl">About this Playground</h2>
          <p class="text-base-content/80 leading-relaxed">
            This playground (generated from a UFO RPC schema) provides an
            interactive environment where you can explore the API, read the
            documentation, and experiment by executing procedures and
            subscribing to streams
          </p>
        </div>
      </div>
    </div>

    <div class="mb-16">
      <h2 class="mb-4 text-3xl font-bold">Who is UFO RPC for?</h2>
      <p class="text-base-content/80 leading-relaxed">
        UFO RPC is the ideal tool for teams and developers who:
      </p>
      <ul class="text-base-content/80 mt-4 list-disc space-y-2 pl-6">
        <li>
          Want the <span class="font-semibold">type safety of gRPC</span>
          without the complexity of Protobuf and HTTP/2
        </li>
        <li>
          Need clear, well-defined APIs but don't require the
          <span class="font-semibold">querying flexibility of GraphQL</span>
        </li>
        <li>
          Are looking for the <span class="font-semibold"
            >simplicity of tRPC</span
          >
          but need to support a
          <span class="font-semibold">polyglot ecosystem</span>
          (beyond just TypeScript)
        </li>
        <li>
          Value a fast workflow, exceptional tooling, and the ability to move
          confidently between the backend and frontend
        </li>
      </ul>
      <p class="text-base-content/70 mt-4">
        In short, if you want to build modern APIs quickly and safely, UFO RPC
        is built for you.
      </p>
    </div>

    <div>
      <h2 class="mb-8 text-3xl font-bold">Features</h2>
      <div class="desk:grid-cols-4 grid gap-6">
        <div class="card bg-base-200 shadow-sm">
          <div class="card-body">
            <div class="flex flex-col gap-4">
              <div
                class="bg-primary/10 inline-flex items-center justify-start rounded-full p-3"
              >
                <Laptop class="mr-2 size-6 flex-none" />
                <h3 class="text-xl font-semibold">DX first</h3>
              </div>

              <p class="text-base-content/80">
                Every single detail of UFO RPC is designed to make your life as
                a developer easier and more productive
              </p>

              <ul class="text-base-content/70 list-disc space-y-1 pl-6 text-sm">
                <li>
                  <span class="font-semibold">Intuitive DSL</span>: A simple,
                  human-readable language for defining APIs
                </li>
                <li>
                  <span class="font-semibold">Powerful Tooling</span>:
                  Playground, LSP support, formatter, VS Code extension, etc
                </li>
                <li>
                  <span class="font-semibold">Type Safety</span>: End-to-end
                  type safety from schema to client and server
                </li>
              </ul>
            </div>
          </div>
        </div>

        <div class="card bg-base-200 shadow-sm">
          <div class="card-body">
            <div class="flex flex-col gap-4">
              <div
                class="bg-secondary/10 inline-flex items-center justify-start rounded-full p-3"
              >
                <FileJson class="mr-2 size-6 flex-none" />
                <h3 class="text-xl font-semibold">Open Source</h3>
              </div>

              <p class="text-base-content/80">
                UFO RPC is <span class="font-semibold">100% open source</span> and
                uses a dual-open-license approach
              </p>

              <ul class="text-base-content/70 list-disc space-y-1 pl-6 text-sm">
                <li>
                  <span class="font-semibold">UFO RPC Generator</span>: AGPL-3.0
                </li>
                <li><span class="font-semibold">Generated Code</span>: MIT</li>
                <li>
                  Ensures UFO RPC remains free and open source while your
                  generated code can be used freely in any project
                </li>
              </ul>
            </div>
          </div>
        </div>

        <div class="card bg-base-200 shadow-sm">
          <div class="card-body">
            <div class="flex flex-col gap-4">
              <div
                class="bg-accent/10 inline-flex items-center justify-start rounded-full p-3"
              >
                <Languages class="mr-2 size-6 flex-none" />
                <h3 class="text-xl font-semibold">Cross language</h3>
              </div>

              <p class="text-base-content/80">
                Designed to work in a polyglot ecosystem from a single source of
                truth
              </p>

              <ul class="text-base-content/70 list-disc space-y-1 pl-6 text-sm">
                <li>
                  <span class="font-semibold">Server & Client</span>:
                  Zero-dependency, high-performance servers and clients for
                  supported languages
                </li>
                <li>
                  <span class="font-semibold">OpenAPI</span>: Generate OpenAPI
                  specs from your schema
                </li>
                <li>
                  <span class="font-semibold">Code Snippets</span>: Ready-to-use
                  curl and many languages
                </li>
              </ul>
            </div>
          </div>
        </div>

        <div class="card bg-base-200 shadow-sm">
          <div class="card-body">
            <div class="flex flex-col gap-4">
              <div
                class="bg-info/10 inline-flex items-center justify-start rounded-full p-3"
              >
                <MonitorSmartphone class="mr-2 size-6 flex-none" />
                <h3 class="text-xl font-semibold">Playground</h3>
              </div>

              <p class="text-base-content/80">
                Generate a static, zero-dependency, responsive web portal for
                your schema
              </p>

              <ul class="text-base-content/70 list-disc space-y-1 pl-6 text-sm">
                <li>
                  <span class="font-semibold">Explore</span>: Procedures,
                  streams, and data types
                </li>
                <li>
                  <span class="font-semibold">Read the Docs</span>: Markdown
                  docs alongside your schema
                </li>
                <li>
                  <span class="font-semibold">Test the API</span>: Execute
                  procedures and subscribe to streams
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <BottomSpace />
</div>
