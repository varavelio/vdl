<script lang="ts">
  import {
    BookOpenText,
    Braces,
    FileText,
    House,
    LibraryIcon,
    List,
    Lock,
    NetworkIcon,
    Radio,
    Regex,
    ServerCog,
    X,
    Zap,
  } from "@lucide/svelte";
  import { onMount } from "svelte";

  import {
    getNodeName,
    getNodeSlug,
    storeSettings,
  } from "$lib/storeSettings.svelte";
  import { dimensionschangeAction, storeUi } from "$lib/storeUi.svelte";
  import { getVersion } from "$lib/version";

  import Logo from "$lib/components/Logo.svelte";
  import Offcanvas from "$lib/components/Offcanvas.svelte";
  import Tooltip from "$lib/components/Tooltip.svelte";

  import LayoutAsideCollapse from "./LayoutAsideCollapse.svelte";
  import LayoutAsideLink from "./LayoutAsideLink.svelte";

  // if has hash anchor navigate to it, the id's are set in LayoutAsideLink component
  // and follows the `navlink-{hash}` pattern without #/
  onMount(async () => {
    // wait 500ms to ensure the content is rendered
    await new Promise((resolve) => setTimeout(resolve, 500));

    if (window.location.hash) {
      const id = `navlink-${window.location.hash.replaceAll("#/", "")}`;
      const element = document.getElementById(id);
      if (element) {
        element.scrollIntoView({ behavior: "smooth" });
      }
    }
  });

  const defaultSize = 280;
  const minSize = 225;
  const maxSize = 600;
  let isResizing = $state(false);

  function startResize(e: MouseEvent) {
    isResizing = true;
    e.preventDefault();
  }

  function stopResize() {
    isResizing = false;
  }

  function resize(e: MouseEvent) {
    if (!isResizing) return;
    const newWidth = e.clientX;
    if (newWidth >= minSize && newWidth <= maxSize) {
      storeUi.store.asideWidth = newWidth;
    }
  }

  function handleDoubleClick() {
    storeUi.store.asideWidth = defaultSize;
  }

  let asideStyle = $derived.by(() => {
    if (storeUi.store.isMobile) {
      return `width: 100%; max-width: ${defaultSize}px;`;
    }

    return `width: ${storeUi.store.asideWidth}px;`;
  });

  type AsideNavItem = {
    icon: typeof House;
    label: string;
    deprecated?: string;
  } & (
    | {
        kind: "link";
        href: string;
      }
    | {
        kind: "collapse";
        storageKey: string;
        children?: AsideNavItem[];
      }
  );

  let asideNavItems = $derived.by(() => {
    const ir = storeSettings.store.irSchema;
    let globalDocs = ir.docs.filter((doc) => !doc.rpcName);

    const items: AsideNavItem[] = [
      {
        kind: "link",
        label: "Home",
        icon: House,
        href: "#/",
      },
    ];

    if (globalDocs.length > 0) {
      const globalDocsItems: AsideNavItem[] = globalDocs.map((doc) => {
        const label = getNodeName({ kind: "doc", ...doc });
        const slug = getNodeSlug({ kind: "doc", ...doc });
        return {
          kind: "link",
          label: label,
          icon: FileText,
          href: `#/${slug}`,
        };
      });

      items.push({
        kind: "collapse",
        label: "Docs",
        icon: BookOpenText,
        storageKey: "asideDocsOpen",
        children: globalDocsItems,
      });
    }

    if (ir.rpcs.length > 0) {
      let rpcsItems: AsideNavItem[] = [];

      for (const rpc of ir.rpcs) {
        let rpcDocs = ir.docs.filter((d) => d.rpcName == rpc.name);
        let rpcProcs = ir.procedures.filter((p) => p.rpcName == rpc.name);
        let rpcStreams = ir.streams.filter((s) => s.rpcName == rpc.name);

        let rpcDocsItems: AsideNavItem[] = rpcDocs.map((doc) => {
          const label = getNodeName({ kind: "doc", ...doc });
          const slug = getNodeSlug({ kind: "doc", ...doc });
          return {
            kind: "link",
            label: label,
            icon: FileText,
            href: `#/${slug}`,
          };
        });

        if (rpc.doc) {
          const slug = getNodeSlug({ kind: "rpc", ...rpc });
          rpcDocsItems.unshift({
            kind: "link",
            label: rpc.name,
            icon: FileText,
            href: `#/${slug}`,
          });
        }

        let rpcProcsItems: AsideNavItem[] = rpcProcs.map((proc) => {
          const label = getNodeName({ kind: "proc", ...proc });
          const slug = getNodeSlug({ kind: "proc", ...proc });
          return {
            kind: "link",
            label: label,
            icon: Zap,
            href: `#/${slug}`,
            deprecated: proc.deprecated,
          };
        });

        let rpcStreamsItems: AsideNavItem[] = rpcStreams.map((stream) => {
          const label = getNodeName({ kind: "stream", ...stream });
          const slug = getNodeSlug({ kind: "stream", ...stream });
          return {
            kind: "link",
            label: label,
            icon: Radio,
            href: `#/${slug}`,
            deprecated: stream.deprecated,
          };
        });

        rpcsItems.push({
          kind: "collapse",
          label: rpc.name,
          icon: ServerCog,
          storageKey: `asideRpcOpen:${rpc.name}`,
          deprecated: rpc.deprecated,
          children: [...rpcDocsItems, ...rpcProcsItems, ...rpcStreamsItems],
        });
      }

      items.push({
        kind: "collapse",
        label: "RPC's",
        icon: NetworkIcon,
        storageKey: "asideRpcsOpen",
        children: rpcsItems,
      });
    }

    if (
      ir.constants.length > 0 ||
      ir.patterns.length > 0 ||
      ir.enums.length > 0 ||
      ir.types.length > 0
    ) {
      const referenceItems: AsideNavItem[] = [];

      if (ir.constants.length > 0) {
        referenceItems.push({
          kind: "link",
          label: "Constants",
          icon: Lock,
          href: `#/constants`,
        });
      }
      if (ir.patterns.length > 0) {
        referenceItems.push({
          kind: "link",
          label: "Patterns",
          icon: Regex,
          href: `#/patterns`,
        });
      }
      if (ir.enums.length > 0) {
        referenceItems.push({
          kind: "link",
          label: "Enums",
          icon: List,
          href: `#/enums`,
        });
      }
      if (ir.types.length > 0) {
        referenceItems.push({
          kind: "link",
          label: "Types",
          icon: Braces,
          href: `#/types`,
        });
      }

      items.push({
        kind: "collapse",
        label: "Reference",
        icon: LibraryIcon,
        storageKey: "asideReferenceOpen",
        children: referenceItems,
      });
    }

    return items;
  });
</script>

{#snippet asideNavItem(item: AsideNavItem)}
  {#if item.kind == "link"}
    <LayoutAsideLink
      label={item.label}
      icon={item.icon}
      href={item.href ?? ""}
      deprecated={item.deprecated}
    />
  {/if}
  {#if item.kind == "collapse"}
    <LayoutAsideCollapse
      label={item.label}
      icon={item.icon}
      storageKey={item.storageKey}
      deprecated={item.deprecated}
    >
      {#each item.children as child}
        {@render asideNavItem(child)}
      {/each}
    </LayoutAsideCollapse>
  {/if}
{/snippet}

{#snippet aside()}
  <aside
    use:dimensionschangeAction
    ondimensionschange={(e) => (storeUi.store.aside = e.detail)}
    style={asideStyle}
    class={[
      "bg-base-100 relative h-dvh flex-none scroll-p-32.5",
      "overflow-x-hidden overflow-y-auto",
    ]}
  >
    <header class="bg-base-100 sticky top-0 z-10 w-full shadow-xs">
      {#if !storeUi.store.isMobile}
        <a
          class="sticky top-0 z-10 flex h-18 w-full items-end p-4"
          href="https://vdl.varavel.com"
          target="_blank"
        >
          <Tooltip content={getVersion()} placement="right">
            <Logo class="mx-auto w-30" />
          </Tooltip>
        </a>
      {/if}

      {#if storeUi.store.isMobile}
        <div class="flex items-center justify-between p-4">
          <Logo class="mx-3 w-27.5" />

          <button
            class="btn btn-ghost btn-square btn-sm"
            onclick={() => (storeUi.store.asideOpen = !storeUi.store.asideOpen)}
          >
            <X class="size-6" />
          </button>
        </div>
      {/if}
    </header>

    <nav class="space-y-1.5 p-4 pb-8">
      {#each asideNavItems as item}
        {@render asideNavItem(item)}
      {/each}
    </nav>
  </aside>

  {#if !storeUi.store.isMobile}
    <button
      aria-label="Resize"
      class={[
        "group border-base-content/20 fixed top-0 z-40 h-dvh w-1.5 cursor-col-resize border-l",
        "hover:bg-primary/50 hover:border-l-0",
      ]}
      style="left: {storeUi.store.asideWidth - 2}px;"
      onmousedown={startResize}
      ondblclick={handleDoubleClick}
    ></button>
  {/if}
{/snippet}

<svelte:window onmousemove={resize} onmouseup={stopResize} />

{#if !storeUi.store.isMobile}
  {@render aside()}
{/if}

{#if storeUi.store.isMobile}
  <Offcanvas bind:isOpen={storeUi.store.asideOpen}>
    {@render aside()}
  </Offcanvas>
{/if}
