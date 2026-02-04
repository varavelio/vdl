<script lang="ts">
  import { Link, Plus, RefreshCcw, Settings, Trash, X } from "@lucide/svelte";

  import { ctrlSymbol } from "$lib/helpers/ctrlSymbol";
  import {
    loadDefaultBaseURL,
    loadDefaultHeaders,
    storeSettings,
  } from "$lib/storeSettings.svelte";
  import { storeUi } from "$lib/storeUi.svelte";

  import Modal from "$lib/components/Modal.svelte";
  import Tooltip from "$lib/components/Tooltip.svelte";

  let isOpen = $state(false);
  const openModal = () => (isOpen = true);
  const closeModal = () => (isOpen = false);

  const onKeydown = (e: KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === ",") {
      e.preventDefault();
      openModal();
    }
  };

  $effect(() => {
    window.addEventListener("keydown", onKeydown);
    return () => {
      window.removeEventListener("keydown", onKeydown);
    };
  });

  const addHeader = () => {
    storeSettings.store.headers = [
      ...storeSettings.store.headers,
      { key: "", value: "", enabled: true, description: "" },
    ];
  };

  const removeHeader = (index: number) => {
    storeSettings.store.headers = storeSettings.store.headers.filter(
      (_, i) => i !== index,
    );
  };

  const loadDefaultBaseUrlWithConfirm = () => {
    if (confirm("Are you sure you want to load the default base URL?")) {
      loadDefaultBaseURL();
    }
  };

  const loadDefaultHeadersWithConfirm = () => {
    if (confirm("Are you sure you want to load the default headers?")) {
      loadDefaultHeaders();
    }
  };
</script>

<button
  class="btn btn-ghost flex items-center justify-start space-x-1 text-sm"
  onclick={openModal}
>
  <Settings class="size-4" />
  <span>Settings</span>
  {#if !storeUi.store.isMobile}
    <span class="ml-4">
      <kbd class="kbd kbd-sm">{ctrlSymbol()}</kbd>
      <kbd class="kbd kbd-sm">,</kbd>
    </span>
  {/if}
</button>

<Modal bind:isOpen class="max-w-3xl">
  <div class="flex w-full items-center justify-between">
    <h3 class="text-xl font-bold">Settings</h3>
    <button class="btn btn-circle btn-ghost" onclick={closeModal}>
      <X class="size-4" />
    </button>
  </div>

  <p>Settings are saved in your browser's local storage.</p>

  <div class="mt-4 space-y-4">
    <fieldset class="fieldset">
      <legend class="fieldset-legend">Base URL</legend>
      <div class="flex space-x-1">
        <label class="input w-full">
          <Link class="size-4" />
          <input
            type="url"
            class="grow"
            spellcheck="false"
            placeholder="https://api.example.com/rpc"
            bind:value={storeSettings.store.baseUrl}
          />
        </label>
        <Tooltip content="Reset base URL to default">
          <button
            class="btn btn-ghost btn-square"
            onclick={loadDefaultBaseUrlWithConfirm}
          >
            <RefreshCcw class="size-4" />
          </button>
        </Tooltip>
      </div>
      <p class="label text-wrap">
        This is the base URL where the VDL RPC server is running, all requests
        will be sent to {`<base-url>/{rpcName}/{operationName}`}.
      </p>
    </fieldset>

    <fieldset class="fieldset">
      <legend class="fieldset-legend">Headers</legend>
      <p class="label mb-1">Headers to send with requests to the endpoint.</p>

      {#if storeSettings.store.headers.length > 0}
        <div class="overflow-x-auto">
          <table class="table-xs table w-full min-w-[720px]">
            <thead>
              <tr>
                <th class="w-0"></th>
                <th>Key</th>
                <th>Value</th>
                <th>Description (optional)</th>
                <th class="w-0"></th>
              </tr>
            </thead>
            <tbody>
              {#each storeSettings.store.headers as header, index}
                <tr>
                  <td>
                    <Tooltip
                      content={header.enabled
                        ? "Disable header"
                        : "Enable header"}
                    >
                      <input
                        type="checkbox"
                        class="checkbox"
                        bind:checked={header.enabled}
                      />
                    </Tooltip>
                  </td>
                  <td>
                    <input
                      type="text"
                      class="input w-full"
                      spellcheck="false"
                      placeholder="Key"
                      bind:value={header.key}
                    />
                  </td>
                  <td>
                    <input
                      type="text"
                      class="input w-full"
                      spellcheck="false"
                      placeholder="Value"
                      bind:value={header.value}
                    />
                  </td>
                  <td>
                    <input
                      type="text"
                      class="input w-full"
                      spellcheck="false"
                      placeholder="Description (optional)"
                      bind:value={header.description}
                    />
                  </td>
                  <td>
                    <Tooltip content="Remove header">
                      <button
                        class="btn btn-square hover:btn-error"
                        onclick={() => removeHeader(index)}
                      >
                        <Trash class="size-4" />
                      </button>
                    </Tooltip>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}

      <div class="mt-2 flex justify-end space-x-2">
        <Tooltip content="Reset headers to default">
          <button class="btn btn-ghost" onclick={loadDefaultHeadersWithConfirm}>
            <RefreshCcw class="mr-1 size-4" />
            Reset
          </button>
        </Tooltip>
        <Tooltip content="Add a new header">
          <button class="btn btn-ghost" onclick={addHeader}>
            <Plus class="mr-1 size-4" />
            Add
          </button>
        </Tooltip>
      </div>
    </fieldset>
  </div>
</Modal>
