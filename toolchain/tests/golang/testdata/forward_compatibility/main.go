// Verifies forward compatibility: the client should ignore unknown fields
// returned by the server (simulating a newer server version).
package main

import (
	"context"
	"e2e/gen"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type AppProps struct{}

func main() {
	// Create a custom server that returns extra fields not in the schema
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/Service/GetData", func(w http.ResponseWriter, r *http.Request) {
		// Server returns extra fields that the client doesn't know about
		response := map[string]any{
			"ok": true,
			"output": map[string]any{
				"name":         "test",
				"value":        42,
				"unknownField": "this field doesn't exist in schema",
				"extraNested": map[string]any{
					"foo": "bar",
					"baz": 123,
				},
				"extraArray": []string{"a", "b", "c"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()
	ctx := context.Background()

	result, err := client.RPCs.Service().Procs.GetData().Execute(ctx, gen.ServiceGetDataInput{})
	if err != nil {
		panic(fmt.Sprintf("execute failed: %v", err))
	}

	// Verify the known fields are correctly parsed
	if result.Name != "test" {
		panic(fmt.Sprintf("expected Name='test', got: %s", result.Name))
	}
	if result.Value != 42 {
		panic(fmt.Sprintf("expected Value=42, got: %d", result.Value))
	}

	// The test passes if we got here - unknown fields were ignored without error
	fmt.Println("Success")
}
