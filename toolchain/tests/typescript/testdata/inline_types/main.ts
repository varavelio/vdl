import {
  type ComplexInlineTypes,
  type ComplexInlineTypesConfigs,
  type ComplexInlineTypesDeepNest,
  ComplexInlineTypesDeepNestChild,
  ComplexInlineTypesDeepNestChildGrandChild,
  ComplexInlineTypesDeepNestChildGrandChildGreatGrandChild,
  type ComplexInlineTypesFiles,
  type ComplexInlineTypesGrid,
  type ComplexInlineTypesGroupedFiles,
  type ComplexInlineTypesMeta,
  type ComplexInlineTypesSimple,
} from "./gen/index.ts";

function main() {
  // 1. Array of inline objects
  const files: ComplexInlineTypesFiles[] = [{ path: "a", content: "b" }];

  // 2. Map of inline objects
  const meta: Record<string, ComplexInlineTypesMeta> = {
    v1: { createdAt: "2023", author: "me" },
  };

  // 3. Map of arrays of inline objects
  const groupedFiles: Record<string, ComplexInlineTypesGroupedFiles[]> = {
    group1: [{ name: "f1", size: 10 }],
  };

  // 4. Array of maps of inline objects
  const configs: Record<string, ComplexInlineTypesConfigs>[] = [
    { conf1: { key: "k", value: "v" } },
  ];

  // 5. Nested arrays of inline objects
  const grid: ComplexInlineTypesGrid[][] = [[{ x: 1, y: 2 }]];

  // 6. Simple inline object
  const simple: ComplexInlineTypesSimple = {
    name: "test",
    enabled: true,
  };

  // 7. Deeply nested inline objects
  const deepNest: ComplexInlineTypesDeepNest = {
    level1: "l1",
    child: {
      level2: 2,
      grandChild: {
        level3: true,
        greatGrandChild: {
          level4: 4.5,
          data: "end",
        },
      },
    },
  };

  const output: ComplexInlineTypes = {
    files,
    meta,
    groupedFiles,
    configs,
    grid,
    simple,
    deepNest,
  };

  // Verify integrity
  if (output.files[0].path !== "a") throw new Error("files mismatch");
  if (output.meta["v1"].author !== "me") throw new Error("meta mismatch");
  if (output.groupedFiles["group1"][0].size !== 10) throw new Error("groupedFiles mismatch");
  if (output.configs[0]["conf1"].value !== "v") throw new Error("configs mismatch");
  if (output.grid[0][0].y !== 2) throw new Error("grid mismatch");
  if (output.simple.name !== "test") throw new Error("simple mismatch");
  if (output.deepNest.child.grandChild.greatGrandChild.data !== "end")
    throw new Error("deepNest mismatch");

  console.log("Success");
}

main();
