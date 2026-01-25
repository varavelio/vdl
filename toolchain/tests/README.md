# VDL e2e Tests

This directory contains End-to-End (E2E) tests for the VDL toolchain. These tests verify that the **generated code** compiles and runs correctly in a real environment.

## Structure

- `e2e_test.go`: The test runner. It builds the VDL binary and executes scenarios found in `testdata/`.
- `testdata/`: Contains individual test scenarios (folders).

## How it works

For each test case (e.g., `testdata/simple_rpc`), the runner:

1.  **Builds** the VDL binary from the current source.
2.  **Generates** code inside the test case folder (`vdl generate`).
3.  **Runs** the `main.go` file using `go run`.

If `main.go` runs successfully (exit code 0), the test passes. If it panics or fails, the test fails.

## Adding a test case

1.  Create a folder in `tests/golang/testdata/<case_name>`.
2.  Add these 4 files:

    - `go.mod`:
      ```go
      module e2e
      go 1.23
      ```
    - `vdl.yaml`: Config pointing output to `gen/vdl.go`.
    - `schema.vdl`: Your VDL definitions.
    - `main.go`: The test logic. Must implement server/client, run the flow, and **panic** on failure.

    **Example `main.go`:**

    ```go
    package main

    import (
        "fmt"
        "e2e/gen" // Import the generated package
    )

    func main() {
        // 1. Setup Server
        // 2. Setup Client
        // 3. Execute RPC

        // Panic if something is wrong
        // if err != nil { panic(err) }

        fmt.Println("Success")
    }
    ```
