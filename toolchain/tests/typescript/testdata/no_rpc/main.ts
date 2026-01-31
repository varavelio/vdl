import * as fs from "node:fs";
import * as path from "node:path";
import { Something } from "./gen/index.ts";

function main() {
  const s: Something = { field: "value" };
  if (s.field !== "value") throw new Error("field mismatch");

  const catalogPath = path.join(process.cwd(), "gen", "catalog.ts");
  if (fs.existsSync(catalogPath)) {
    throw new Error("catalog.ts should not exist");
  }

  console.log("Success");
}

main();
