// @ts-check
import starlight from "@astrojs/starlight";
import svelte from "@astrojs/svelte";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "astro/config";
import fse from "fs-extra";
import { bundledLanguages } from "shiki";
import { viteStaticCopy } from "vite-plugin-static-copy";

// https://docs.astro.build/en/reference/configuration-reference/#markdown-options
// https://shiki.style/guide/load-lang
function getShikiLangs() {
  const syntaxPath = "../editors/vscode/language/vdl.tmLanguage.json";
  const vdlLang = fse.readJSONSync(syntaxPath);
  vdlLang.name = "vdl";
  return [...Object.keys(bundledLanguages), vdlLang];
}

// https://astro.build/config
export default defineConfig({
  vite: {
    plugins: [
      viteStaticCopy({
        targets: [
          {
            src: "../assets/favicon.ico",
            dest: "./",
          },
          {
            src: "../assets/*",
            dest: "./_vdl/assets",
          },
          {
            src: "../toolchain/dist/vdl.wasm",
            dest: "./_vdl/wasm",
          },
          {
            src: "../toolchain/dist/wasm_exec.js",
            dest: "./_vdl/wasm",
          },
        ],
      }),
      tailwindcss(),
    ],
  },

  integrations: [
    svelte(),
    starlight({
      favicon: "/_vdl/assets/png/icon.png",
      title: "VDL",
      description:
        "Open-source cross-language definition engine for modern stacks. Define your data structures, APIs, contracts, and generate type-safe code for your backend and frontend instantly.",
      logo: {
        light: "../assets/svg/vdl.svg",
        dark: "../assets/svg/vdl-white.svg",
        alt: "VDL Logo",
        replacesTitle: true,
      },
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/varavelio/vdl",
        },
        {
          icon: "discord",
          label: "Discord",
          href: "https://vdl.varavel.com/discord",
        },
        {
          icon: "reddit",
          label: "Reddit",
          href: "https://vdl.varavel.com/reddit",
        },
        {
          icon: "x.com",
          label: "X (Twitter)",
          href: "https://vdl.varavel.com/x",
        },
      ],
      sidebar: [
        {
          label: "Guides",
          autogenerate: { directory: "guides" },
        },
        {
          label: "Reference",
          autogenerate: { directory: "reference" },
        },
      ],
      editLink: {
        baseUrl: "https://github.com/varavelio/vdl/tree/main/docs/",
      },
      customCss: ["./src/styles/global.css"],
    }),
  ],

  markdown: {
    syntaxHighlight: "shiki",
    shikiConfig: {
      langs: getShikiLangs(),
    },
  },
});
