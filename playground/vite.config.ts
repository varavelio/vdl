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
          src: "../toolchain/dist/vdl.wasm",
          dest: "_app/vdl",
        },
        {
          src: "../toolchain/dist/wasm_exec.js",
          dest: "_app/vdl",
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
