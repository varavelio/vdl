// Verifies VDLPaths, VDLProcedures, and VDLStreams catalog exports.
import { VDLPaths, VDLProcedures, VDLStreams } from "./gen/index.ts";

function fail(name: string, expected: string, actual: string): never {
  console.error(`${name} mismatch: expected "${expected}", got "${actual}"`);
  process.exit(1);
}

function main() {
  // Verify VDLPaths structure (only procs and streams, no service root path)
  if (VDLPaths.MyService.MyProc !== "/MyService/MyProc") {
    fail("VDLPaths.MyService.MyProc", "/MyService/MyProc", VDLPaths.MyService.MyProc);
  }
  if (VDLPaths.MyService.MyStream !== "/MyService/MyStream") {
    fail("VDLPaths.MyService.MyStream", "/MyService/MyStream", VDLPaths.MyService.MyStream);
  }

  // Verify VDLProcedures contains MyProc
  let foundProc = false;
  for (const op of VDLProcedures) {
    if (op.rpcName === "MyService" && op.name === "MyProc") {
      foundProc = true;
      // Verify path method
      const expectedPath = "/MyService/MyProc";
      const actualPath = `/${op.rpcName}/${op.name}`;
      if (actualPath !== expectedPath) {
        fail("op path for MyProc", expectedPath, actualPath);
      }
    }
  }
  if (!foundProc) {
    console.error("MyProc operation not found in VDLProcedures");
    process.exit(1);
  }

  // Verify VDLStreams contains MyStream
  let foundStream = false;
  for (const op of VDLStreams) {
    if (op.rpcName === "MyService" && op.name === "MyStream") {
      foundStream = true;
      // Verify path method
      const expectedPath = "/MyService/MyStream";
      const actualPath = `/${op.rpcName}/${op.name}`;
      if (actualPath !== expectedPath) {
        fail("op path for MyStream", expectedPath, actualPath);
      }
    }
  }
  if (!foundStream) {
    console.error("MyStream operation not found in VDLStreams");
    process.exit(1);
  }

  console.log("Paths verification successful");
  process.exit(0);
}

main();
