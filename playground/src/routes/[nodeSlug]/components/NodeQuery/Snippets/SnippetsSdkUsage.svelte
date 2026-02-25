<script lang="ts">
  import Code from "$lib/components/Code.svelte";
  import { storeUi } from "$lib/storeUi.svelte";

  interface Props {
    type: "proc" | "stream";
    name: string;
  }

  const { type, name }: Props = $props();

  function toWords(value: string): string[] {
    return value
      .replace(/[^a-zA-Z0-9]+/g, " ")
      .trim()
      .split(/\s+/)
      .filter(Boolean);
  }

  function toCamelCase(value: string): string {
    const words = toWords(value);
    if (words.length === 0) return value;
    const [first, ...rest] = words;
    return (
      first.charAt(0).toLowerCase() +
      first.slice(1) +
      rest.map((w) => w.charAt(0).toUpperCase() + w.slice(1)).join("")
    );
  }

  function toPascalCase(value: string): string {
    const words = toWords(value);
    return words.map((w) => w.charAt(0).toUpperCase() + w.slice(1)).join("");
  }

  const isProc = $derived.by(() => type === "proc");
  const nameCamel = $derived.by(() => toCamelCase(name));
  const namePascal = $derived.by(() => toPascalCase(name));

  const tsProc = $derived.by(
    () => `// Assuming \`client\` is already created (see Setup)
const result = await client.procs.${nameCamel}().execute(...);
console.log(result);`,
  );

  const tsStream = $derived.by(
    () => `// Assuming \`client\` is already created (see Setup)
const { stream, cancel } = client.streams.${nameCamel}().execute(...);

for await (const event of stream) {
  console.log(event);
}

cancel();`,
  );

  const goProc = $derived.by(
    () => `// Assuming \`client\` is already created (see Setup)
output, err := client.Procs.${namePascal}().Execute(context.Background(), ...)
if err != nil {
  // handle error
}
fmt.Println(output)`,
  );

  const goStream = $derived.by(
    () => `// Assuming \`client\` is already created (see Setup)
stream := client.Streams.${namePascal}().Execute(context.Background(), ...)

for event := range stream {
  fmt.Println(event)
}`,
  );

  const dartProc = $derived.by(
    () => `// Assuming \`client\` is already created (see Setup)
final result = await client.procs.${nameCamel}().execute(...);
print(result);`,
  );

  const dartStream = $derived.by(
    () => `// Assuming \`client\` is already created (see Setup)
final handle = client.streams.${nameCamel}().execute(...);

await for (final event in handle.stream) {
  print(event.output?.toJson());
}

handle.cancel();`,
  );

  const pythonProc = $derived.by(
    () => `# Assuming \`client\` is already created (see Setup)
result = client.procs.${nameCamel}().execute(...)
print(result)`,
  );

  const pythonStream = $derived.by(
    () => `# Assuming \`client\` is already created (see Setup)
stream = client.streams.${nameCamel}().execute(...)

for event in stream:
    print(event)`,
  );
</script>

<div class="prose prose-sm text-base-content max-w-none space-y-4">
  <p>
    Here is a simple example of how to call this
    {isProc
      ? "procedure"
      : "stream"}
    with the SDK.
  </p>

  <p>Use your editor/IDE to help you with the types, the SDK is fully typed.</p>

  {#if storeUi.store.codeSnippetsSdkLang === "typescript"}
    <div class="not-prose"><Code code={isProc ? tsProc : tsStream} lang="ts" /></div>
  {:else if storeUi.store.codeSnippetsSdkLang === "go"}
    <div class="not-prose"><Code code={isProc ? goProc : goStream} lang="go" /></div>
  {:else if storeUi.store.codeSnippetsSdkLang === "dart"}
    <div class="not-prose"><Code code={isProc ? dartProc : dartStream} lang="dart" /></div>
  {:else if storeUi.store.codeSnippetsSdkLang === "python"}
    <div class="not-prose"><Code code={isProc ? pythonProc : pythonStream} lang="python" /></div>
  {/if}
</div>
