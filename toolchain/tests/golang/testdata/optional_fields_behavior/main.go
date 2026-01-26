// Verifies optional field behavior: server can detect whether optional fields
// were explicitly provided vs absent, and empty values are distinct from absent.
package main

import (
	"bytes"
	"context"
	"e2e/gen"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

type AppProps struct{}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Procs.Echo().Handle(func(c *gen.ServiceEchoHandlerContext[AppProps]) (gen.ServiceEchoOutput, error) {
		return gen.ServiceEchoOutput{
			Required:               c.Input.Required,
			Optional:               c.Input.Optional,
			OptionalInt:            c.Input.OptionalInt,
			OptionalBool:           c.Input.OptionalBool,
			WasOptionalPresent:     c.Input.Optional.Present,
			WasOptionalIntPresent:  c.Input.OptionalInt.Present,
			WasOptionalBoolPresent: c.Input.OptionalBool.Present,
		}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Test 1: All fields absent
	testAllAbsent(ts.URL)

	// Test 2: All fields present (including empty values)
	testAllPresent(ts.URL)

	// Test 3: Mixed presence
	testMixedPresence(ts.URL)

	// Test 4: Using generated client
	testGeneratedClient(ts.URL)

	fmt.Println("Success")
}

func testAllAbsent(baseURL string) {
	payload := `{"required": "hello"}`
	resp, err := http.Post(baseURL+"/rpc/Service/Echo", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		panic(fmt.Sprintf("request failed: %v", err))
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %v", err))
	}

	if result["ok"] != true {
		panic(fmt.Sprintf("expected ok=true, got: %s", string(body)))
	}

	output := result["output"].(map[string]any)

	if output["wasOptionalPresent"] != false {
		panic("expected wasOptionalPresent=false when optional is absent")
	}
	if output["wasOptionalIntPresent"] != false {
		panic("expected wasOptionalIntPresent=false when optionalInt is absent")
	}
	if output["wasOptionalBoolPresent"] != false {
		panic("expected wasOptionalBoolPresent=false when optionalBool is absent")
	}
}

func testAllPresent(baseURL string) {
	// Provide all optional fields, including empty/zero values
	payload := `{"required": "hello", "optional": "", "optionalInt": 0, "optionalBool": false}`
	resp, err := http.Post(baseURL+"/rpc/Service/Echo", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		panic(fmt.Sprintf("request failed: %v", err))
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %v", err))
	}

	if result["ok"] != true {
		panic(fmt.Sprintf("expected ok=true, got: %s", string(body)))
	}

	output := result["output"].(map[string]any)

	if output["wasOptionalPresent"] != true {
		panic("expected wasOptionalPresent=true when optional is explicitly empty string")
	}
	if output["wasOptionalIntPresent"] != true {
		panic("expected wasOptionalIntPresent=true when optionalInt is explicitly 0")
	}
	if output["wasOptionalBoolPresent"] != true {
		panic("expected wasOptionalBoolPresent=true when optionalBool is explicitly false")
	}
}

func testMixedPresence(baseURL string) {
	// Only provide optional string
	payload := `{"required": "hello", "optional": "world"}`
	resp, err := http.Post(baseURL+"/rpc/Service/Echo", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		panic(fmt.Sprintf("request failed: %v", err))
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %v", err))
	}

	if result["ok"] != true {
		panic(fmt.Sprintf("expected ok=true, got: %s", string(body)))
	}

	output := result["output"].(map[string]any)

	if output["wasOptionalPresent"] != true {
		panic("expected wasOptionalPresent=true")
	}
	if output["wasOptionalIntPresent"] != false {
		panic("expected wasOptionalIntPresent=false")
	}
	if output["wasOptionalBoolPresent"] != false {
		panic("expected wasOptionalBoolPresent=false")
	}

	if output["optional"] != "world" {
		panic(fmt.Sprintf("expected optional='world', got: %v", output["optional"]))
	}
}

func testGeneratedClient(baseURL string) {
	client := gen.NewClient(baseURL + "/rpc").Build()
	ctx := context.Background()

	// Test with all optionals absent
	result, err := client.RPCs.Service().Procs.Echo().Execute(ctx, gen.ServiceEchoInput{
		Required: "test",
	})
	if err != nil {
		panic(fmt.Sprintf("execute failed: %v", err))
	}

	if result.WasOptionalPresent {
		panic("expected WasOptionalPresent=false")
	}

	// Test with optional present
	result2, err := client.RPCs.Service().Procs.Echo().Execute(ctx, gen.ServiceEchoInput{
		Required: "test",
		Optional: gen.Some("value"),
	})
	if err != nil {
		panic(fmt.Sprintf("execute failed: %v", err))
	}

	if !result2.WasOptionalPresent {
		panic("expected WasOptionalPresent=true")
	}
	if result2.Optional.Value != "value" {
		panic(fmt.Sprintf("expected Optional.Value='value', got: %s", result2.Optional.Value))
	}
}
