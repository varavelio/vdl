<script lang="ts">
  import * as curlconverter from "curlconverter";
  import { onMount } from "svelte";
  import Code from "$lib/components/Code.svelte";
  import { storeUi } from "$lib/storeUi.svelte";

  interface Props {
    curl: string;
  }

  const { curl }: Props = $props();

  interface Lang {
    group: string;
    langCode: string;
    label: string;
    func: (code: string) => string;
  }

  interface LangGroup {
    group: string;
    langs: Lang[];
  }

  const langs: Lang[] = [
    {
      group: "Ansible",
      langCode: "yaml",
      label: "Ansible",
      func: curlconverter.toAnsible,
    },
    {
      group: "Bash",
      langCode: "bash",
      label: "Curl",
      func: (code: string) => code,
    },
    {
      group: "Bash",
      langCode: "bash",
      label: "Wget",
      func: curlconverter.toWget,
    },
    {
      group: "C",
      langCode: "c",
      label: "C",
      func: curlconverter.toC,
    },
    {
      group: "C#",
      langCode: "csharp",
      label: "C#",
      func: curlconverter.toCSharp,
    },
    {
      group: "Clojure",
      langCode: "clojure",
      label: "Clojure",
      func: curlconverter.toClojure,
    },
    {
      group: "Dart",
      langCode: "dart",
      label: "Dart",
      func: curlconverter.toDart,
    },
    {
      group: "Elixir",
      langCode: "elixir",
      label: "Elixir",
      func: curlconverter.toElixir,
    },
    {
      group: "Go",
      langCode: "go",
      label: "Go",
      func: curlconverter.toGo,
    },
    {
      group: "HTTP",
      langCode: "http",
      label: "HTTP",
      func: curlconverter.toHTTP,
    },
    {
      group: "HTTPie",
      langCode: "bash",
      label: "HTTPie",
      func: curlconverter.toHttpie,
    },
    {
      group: "Java",
      langCode: "java",
      label: "Java HttpClient",
      func: curlconverter.toJava,
    },
    {
      group: "Java",
      langCode: "java",
      label: "Java HttpUrlConnection",
      func: curlconverter.toJavaHttpUrlConnection,
    },
    {
      group: "Java",
      langCode: "java",
      label: "Java Jsoup",
      func: curlconverter.toJavaJsoup,
    },
    {
      group: "Java",
      langCode: "java",
      label: "Java OkHttp",
      func: curlconverter.toJavaOkHttp,
    },
    {
      group: "JavaScript",
      langCode: "javascript",
      label: "JavaScript Fetch",
      func: curlconverter.toJavaScript,
    },
    {
      group: "JavaScript",
      langCode: "javascript",
      label: "JavaScript JQuery",
      func: curlconverter.toJavaScriptJquery,
    },
    {
      group: "JavaScript",
      langCode: "javascript",
      label: "JavaScript XHR",
      func: curlconverter.toJavaScriptXHR,
    },
    {
      group: "Julia",
      langCode: "julia",
      label: "Julia",
      func: curlconverter.toJulia,
    },
    {
      group: "JSON",
      langCode: "json",
      label: "JSON",
      func: curlconverter.toJsonString,
    },
    {
      group: "Kotlin",
      langCode: "kotlin",
      label: "Kotlin",
      func: curlconverter.toKotlin,
    },
    {
      group: "Lua",
      langCode: "lua",
      label: "Lua",
      func: curlconverter.toLua,
    },
    {
      group: "MATLAB",
      langCode: "matlab",
      label: "MATLAB",
      func: curlconverter.toMATLAB,
    },
    {
      group: "Node.js",
      langCode: "javascript",
      label: "Node.js Axios",
      func: curlconverter.toNodeAxios,
    },
    {
      group: "Node.js",
      langCode: "javascript",
      label: "Node.js Got",
      func: curlconverter.toNodeGot,
    },
    {
      group: "Node.js",
      langCode: "javascript",
      label: "Node.js Ky",
      func: curlconverter.toNodeKy,
    },
    {
      group: "Node.js",
      langCode: "javascript",
      label: "Node.js node-fetch",
      func: curlconverter.toNodeFetch,
    },
    {
      group: "Node.js",
      langCode: "javascript",
      label: "Node.js request",
      func: curlconverter.toNodeRequest,
    },
    {
      group: "Node.js",
      langCode: "javascript",
      label: "Node.js SuperAgent",
      func: curlconverter.toNodeSuperAgent,
    },
    {
      group: "Node.js",
      langCode: "javascript",
      label: "Node.js http",
      func: curlconverter.toNodeHttp,
    },
    {
      group: "Objective-C",
      langCode: "objective-c",
      label: "Objective-C",
      func: curlconverter.toObjectiveC,
    },
    {
      group: "OCaml",
      langCode: "ocaml",
      label: "OCaml",
      func: curlconverter.toOCaml,
    },
    {
      group: "Perl",
      langCode: "perl",
      label: "Perl",
      func: curlconverter.toPerl,
    },
    {
      group: "PHP",
      langCode: "php",
      label: "PHP Curl",
      func: curlconverter.toPhp,
    },
    {
      group: "PHP",
      langCode: "php",
      label: "PHP Guzzle",
      func: curlconverter.toPhpGuzzle,
    },
    {
      group: "PowerShell",
      langCode: "powershell",
      label: "PowerShell",
      func: curlconverter.toPowershellWebRequest,
    },
    {
      group: "Python",
      langCode: "python",
      label: "Python Requests",
      func: curlconverter.toPython,
    },
    {
      group: "Python",
      langCode: "python",
      label: "Python http.client",
      func: curlconverter.toPythonHttp,
    },
    {
      group: "R",
      langCode: "r",
      label: "R httr",
      func: curlconverter.toR,
    },
    {
      group: "R",
      langCode: "r",
      label: "R httr2",
      func: curlconverter.toRHttr2,
    },
    {
      group: "Ruby",
      langCode: "ruby",
      label: "Ruby Net::HTTP",
      func: curlconverter.toRuby,
    },
    {
      group: "Ruby",
      langCode: "ruby",
      label: "Ruby HTTParty",
      func: curlconverter.toRubyHttparty,
    },
    {
      group: "Rust",
      langCode: "rust",
      label: "Rust",
      func: curlconverter.toRust,
    },
    {
      group: "Swift",
      langCode: "swift",
      label: "Swift",
      func: curlconverter.toSwift,
    },
  ];

  // Make sure to always have Curl as the default lang
  const defaultLang = langs[1];

  // This takes every lang and puts it into it's group
  const langGroups = $derived.by(() => {
    const groups: LangGroup[] = [];

    for (const lang of langs) {
      const group = groups.find((group) => group.group === lang.group);
      if (group) {
        group.langs.push(lang);
      } else {
        groups.push({ group: lang.group, langs: [lang] });
      }
    }

    return groups;
  });

  let pickedLang = $derived.by(() => {
    const lang = langs.find((lang) => lang.label === storeUi.store.codeSnippetsCurlLang);
    if (!lang) return defaultLang.langCode;
    return lang.langCode;
  });

  let pickedCode = $derived.by(() => {
    const lang = langs.find((lang) => lang.label === storeUi.store.codeSnippetsCurlLang);
    if (!lang) return defaultLang.func(curl);
    return lang.func(curl);
  });

  onMount(() => {
    const lang = langs.find((lang) => lang.label === storeUi.store.codeSnippetsCurlLang);
    if (!lang) {
      storeUi.store.codeSnippetsCurlLang = defaultLang.label;
    }
  });
</script>

<label class="fieldset mb-4">
  <legend class="fieldset-legend">Language</legend>
  <select class="select w-full appearance-none" bind:value={storeUi.store.codeSnippetsCurlLang}>
    {#each langGroups as langGroup}
      {#if langGroup.langs.length > 1}
        <optgroup label={langGroup.group}>
          {#each langGroup.langs as lang}
            <option value={lang.label}>{lang.label}</option>
          {/each}
        </optgroup>
      {:else}
        <option value={langGroup.langs[0].label}>{langGroup.langs[0].label}</option>
      {/if}
    {/each}
  </select>
  <div class="prose prose-sm text-base-content/50 max-w-none">
    The easiest way to test this API is with the code snippets below. However, if your language is
    supported, <b>using the SDK is recommended</b> for a smoother integration experience.
  </div>
</label>

<Code code={pickedCode} lang={pickedLang} scrollY={false} />
