// @ts-check
import starlight from "@astrojs/starlight";
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
    ],
  },

  integrations: [
    starlight({
      title: "My Docs",
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/withastro/starlight",
        },
      ],
      sidebar: [
        {
          label: "Guides",
          items: [
            // Each item here is one entry in the navigation menu.
            { label: "Example Guide", slug: "guides/example" },
          ],
        },
        {
          label: "Reference",
          autogenerate: { directory: "reference" },
        },
      ],
    }),
  ],

  markdown: {
    syntaxHighlight: "shiki",
    shikiConfig: {
      langs: getShikiLangs(),
    },
  },
});
