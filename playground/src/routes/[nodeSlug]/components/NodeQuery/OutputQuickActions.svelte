<script lang="ts">
  import { Copy, EarthLock, Key, Sparkles } from "@lucide/svelte";
  import Menu from "$lib/components/Menu.svelte";
  import Tooltip from "$lib/components/Tooltip.svelte";
  import { copyTextToClipboard } from "$lib/helpers/copyTextToClipboard";
  import { discoverAuthToken, type TokenInfo } from "$lib/helpers/discoverAuthToken.ts";
  import { setHeader } from "$lib/storeSettings.svelte";

  interface Props {
    output: string | null;
  }

  const { output }: Props = $props();

  // Discover authentication tokens in the response, limit to tokenLimit
  // in order to avoid UI overload
  const tokenLimit = 5;
  let foundTokens = $derived(discoverAuthToken(output));
  let firstTokens = $derived(foundTokens.slice(0, tokenLimit));
  let hasToken = $derived(foundTokens.length > 0);

  function handleSetAuthHeader(token: TokenInfo) {
    setHeader("Authorization", `Bearer ${token.value}`);
  }

  async function handleCopyToClipboard(token: TokenInfo) {
    return copyTextToClipboard(token.value);
  }
</script>

{#if hasToken}
  <div class="flex flex-wrap items-center justify-start gap-2">
    <span class="flex flex-none items-center pr-2 text-xs font-bold">
      <Sparkles class="mr-1 size-3" />
      <span>Quick Actions</span>
    </span>

    {#each firstTokens as token}
      {#snippet menuContent()}
        <div class="flex flex-col space-y-2 py-2">
          <button
            class="btn btn-sm btn-ghost w-full justify-start"
            onclick={() => handleCopyToClipboard(token)}
          >
            <Copy class="mr-1 size-4" />
            <span>Copy token</span>
          </button>
          <Tooltip content={`Authorization: Bearer <${token.path}>`}>
            <button
              class="btn btn-sm btn-ghost w-full justify-start"
              onclick={() => handleSetAuthHeader(token)}
            >
              <EarthLock class="mr-1 size-4" />
              <span>Set as header</span>
            </button>
          </Tooltip>
        </div>
      {/snippet}

      <Menu content={menuContent} placement="top" trigger="mouseenter click">
        <button class="btn btn-xs border-base-content/20 flex-none">
          <Key class="mr-1 size-3" />
          <span>{token.key}</span>
        </button>
      </Menu>
    {/each}
  </div>
{/if}
