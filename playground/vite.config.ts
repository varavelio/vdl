import { sveltekit } from "@sveltejs/kit/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import { viteStaticCopy } from "vite-plugin-static-copy";

export default defineConfig({
  plugins: [
    tailwindcss(),
    sveltekit(),
    viteStaticCopy({
      targets: [
        {
          src: "../assets/favicon.ico",
          dest: "./",
        },
        {
          src: "../assets/*",
          dest: "./_app/vdl/assets",
        },
        {
          src: "../toolchain/dist/vdl.wasm",
          dest: "./_app/vdl/wasm",
        },
        {
          src: "../toolchain/dist/wasm_exec.js",
          dest: "./_app/vdl/wasm",
        },
        {
          src: "../editors/vscode/language/vdl.tmLanguage.json",
          dest: "./_app/vdl/vscode",
        },
        {
          src: "node_modules/web-tree-sitter/tree-sitter.wasm",
          dest: "_app/curlconverter",
        },
        {
          src: "node_modules/curlconverter/dist/tree-sitter-bash.wasm",
          dest: "_app/curlconverter",
        },
      ],
    }),
  ],
  server: {
    host: "0.0.0.0",
  },
  optimizeDeps: {
    esbuildOptions: {
      target: "esnext",
    },
  },
  build: {
    target: "ES2022",
  },
});
