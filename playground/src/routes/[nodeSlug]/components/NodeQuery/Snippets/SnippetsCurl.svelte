<script lang="ts">
  import { joinPath } from "$lib/helpers/joinPath";
  import { getHeadersObject, storeSettings } from "$lib/storeSettings.svelte";

  import Code from "$lib/components/Code.svelte";

  import SnippetsCode from "./SnippetsCurlCode.svelte";

  interface Props {
    input: any;
    type: "proc" | "stream";
    name: string;
  }

  let { input, type, name }: Props = $props();

  let curl = $derived.by(() => {
    const endpoint = joinPath([storeSettings.store.baseUrl, name]);
    const payload = input ?? {};
    let payloadStr = JSON.stringify(payload, null, 2);
    payloadStr = payloadStr.replace(/'/g, "'\\''");

    let c = `curl -X POST ${endpoint} \\\n`;

    if (type === "stream") {
      c += "-N \\\n";
    }

    let headers = getHeadersObject();
    if (type === "stream") {
      headers.set("Accept", "text/event-stream");
      headers.set("Cache-Control", "no-cache");
    }

    headers.forEach((value, key) => {
      let rawHeader = `${key}: ${value}`;
      c += `-H ${JSON.stringify(rawHeader)} \\\n`;
    });

    c += `-d '${payloadStr}'`;

    return c;
  });
</script>

<div>
  {#if type === "stream"}
    <p class="pb-4 text-sm">
      Streams use Server-Sent Events. Only curl examples are provided. Build a
      client manually, or generate one with the urpc CLI if your language is
      supported.
      <br />
      <a href="https://vdl.varavel.com/sse" target="_blank" class="link">
        Learn more here
      </a>
    </p>

    <Code code={curl} lang="bash" scrollY={false} />
  {/if}

  {#if type === "proc"}
    <SnippetsCode {curl} />
  {/if}
</div>
