// Verifies enum serialization: enums should be transmitted as strings on the wire,
// and round-trip correctly through client->server->client.
// Tests both implicit-value enums (name=value) and explicit-value enums (name!=value).
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
			Color:  c.Input.Color,
			Status: c.Input.Status,
		}, nil
	})

	server.RPCs.Service().Procs.GetDefaults().Handle(func(c *gen.ServiceGetDefaultsHandlerContext[AppProps]) (gen.ServiceGetDefaultsOutput, error) {
		return gen.ServiceGetDefaultsOutput{
			Color:  gen.ColorRed,
			Status: gen.StatusPending,
		}, nil
	})

	server.RPCs.Service().Procs.EchoHttpStatus().Handle(func(c *gen.ServiceEchoHttpStatusHandlerContext[AppProps]) (gen.ServiceEchoHttpStatusOutput, error) {
		return gen.ServiceEchoHttpStatusOutput{
			Status: c.Input.Status,
		}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Implicit-value enum tests
	testWireFormat(ts.URL)
	testGeneratedClient(ts.URL)
	testAllEnumValues(ts.URL)

	// Explicit-value enum tests
	testExplicitValueWireFormat(ts.URL)
	testExplicitValueClient(ts.URL)
	testExplicitValueAllMembers(ts.URL)
	testExplicitValueStringMethod()

	fmt.Println("Success")
}

func testWireFormat(baseURL string) {
	// Send raw JSON to verify wire format
	payload := `{"color": "Blue", "status": "Active"}`
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
	if output["color"] != "Blue" {
		panic(fmt.Sprintf("expected color='Blue', got: %v", output["color"]))
	}
	if output["status"] != "Active" {
		panic(fmt.Sprintf("expected status='Active', got: %v", output["status"]))
	}
}

func testGeneratedClient(baseURL string) {
	client := gen.NewClient(baseURL + "/rpc").Build()
	ctx := context.Background()

	result, err := client.RPCs.Service().Procs.Echo().Execute(ctx, gen.ServiceEchoInput{
		Color:  gen.ColorGreen,
		Status: gen.StatusCompleted,
	})
	if err != nil {
		panic(fmt.Sprintf("execute failed: %v", err))
	}

	if result.Color != gen.ColorGreen {
		panic(fmt.Sprintf("expected ColorGreen, got: %v", result.Color))
	}
	if result.Status != gen.StatusCompleted {
		panic(fmt.Sprintf("expected StatusCompleted, got: %v", result.Status))
	}
}

func testAllEnumValues(baseURL string) {
	client := gen.NewClient(baseURL + "/rpc").Build()
	ctx := context.Background()

	colors := []gen.Color{gen.ColorRed, gen.ColorGreen, gen.ColorBlue}
	statuses := []gen.Status{gen.StatusPending, gen.StatusActive, gen.StatusCompleted, gen.StatusCancelled}

	for _, color := range colors {
		for _, status := range statuses {
			result, err := client.RPCs.Service().Procs.Echo().Execute(ctx, gen.ServiceEchoInput{
				Color:  color,
				Status: status,
			})
			if err != nil {
				panic(fmt.Sprintf("execute failed for %s/%s: %v", color, status, err))
			}
			if result.Color != color {
				panic(fmt.Sprintf("color mismatch: expected %v, got %v", color, result.Color))
			}
			if result.Status != status {
				panic(fmt.Sprintf("status mismatch: expected %v, got %v", status, result.Status))
			}
		}
	}
}

// testExplicitValueWireFormat verifies that explicit-value enums use the VALUE (not the name) on the wire.
func testExplicitValueWireFormat(baseURL string) {
	// Send explicit value string - should work
	payload := `{"status": "BAD_REQUEST"}`
	resp, err := http.Post(baseURL+"/rpc/Service/EchoHttpStatus", "application/json", bytes.NewBufferString(payload))
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
		panic(fmt.Sprintf("expected ok=true for BAD_REQUEST, got: %s", string(body)))
	}

	output := result["output"].(map[string]any)
	// Wire format must use the VALUE, not the name
	if output["status"] != "BAD_REQUEST" {
		panic(fmt.Sprintf("expected status='BAD_REQUEST' on wire, got: %v", output["status"]))
	}

	// Verify that using the NAME (not value) is rejected
	payload = `{"status": "BadRequest"}`
	resp2, err := http.Post(baseURL+"/rpc/Service/EchoHttpStatus", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		panic(fmt.Sprintf("request failed: %v", err))
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)
	var result2 map[string]any
	if err := json.Unmarshal(body2, &result2); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %v", err))
	}

	// Using the name instead of value should fail validation
	if result2["ok"] == true {
		panic("using enum NAME 'BadRequest' instead of VALUE 'BAD_REQUEST' should be rejected")
	}
}

// testExplicitValueClient verifies the generated client uses correct values.
func testExplicitValueClient(baseURL string) {
	client := gen.NewClient(baseURL + "/rpc").Build()
	ctx := context.Background()

	// Test each explicit-value enum member
	testCases := []struct {
		input    gen.HttpStatus
		expected string // Expected wire value
	}{
		{gen.HttpStatusOk, "OK"},
		{gen.HttpStatusCreated, "CREATED"},
		{gen.HttpStatusBadRequest, "BAD_REQUEST"},
		{gen.HttpStatusNotFound, "NOT_FOUND"},
		{gen.HttpStatusInternalError, "INTERNAL_SERVER_ERROR"},
	}

	for _, tc := range testCases {
		result, err := client.RPCs.Service().Procs.EchoHttpStatus().Execute(ctx, gen.ServiceEchoHttpStatusInput{
			Status: tc.input,
		})
		if err != nil {
			panic(fmt.Sprintf("execute failed for %s: %v", tc.input, err))
		}
		if result.Status != tc.input {
			panic(fmt.Sprintf("round-trip failed: expected %v, got %v", tc.input, result.Status))
		}
		// Verify the constant value matches expected wire format
		if string(tc.input) != tc.expected {
			panic(fmt.Sprintf("constant value mismatch: expected %q, got %q", tc.expected, string(tc.input)))
		}
	}
}

// testExplicitValueAllMembers verifies all explicit-value enum members round-trip correctly.
func testExplicitValueAllMembers(baseURL string) {
	client := gen.NewClient(baseURL + "/rpc").Build()
	ctx := context.Background()

	allStatuses := []gen.HttpStatus{
		gen.HttpStatusOk,
		gen.HttpStatusCreated,
		gen.HttpStatusBadRequest,
		gen.HttpStatusNotFound,
		gen.HttpStatusInternalError,
	}

	for _, status := range allStatuses {
		result, err := client.RPCs.Service().Procs.EchoHttpStatus().Execute(ctx, gen.ServiceEchoHttpStatusInput{
			Status: status,
		})
		if err != nil {
			panic(fmt.Sprintf("execute failed for %s: %v", status, err))
		}
		if result.Status != status {
			panic(fmt.Sprintf("status mismatch: expected %v, got %v", status, result.Status))
		}
	}
}

// testExplicitValueStringMethod verifies String() returns the value, not the name.
func testExplicitValueStringMethod() {
	// For explicit-value enums, String() should return the value
	cases := []struct {
		status   gen.HttpStatus
		expected string
	}{
		{gen.HttpStatusOk, "OK"},
		{gen.HttpStatusCreated, "CREATED"},
		{gen.HttpStatusBadRequest, "BAD_REQUEST"},
		{gen.HttpStatusNotFound, "NOT_FOUND"},
		{gen.HttpStatusInternalError, "INTERNAL_SERVER_ERROR"},
	}

	for _, tc := range cases {
		got := tc.status.String()
		if got != tc.expected {
			panic(fmt.Sprintf("String() mismatch: expected %q, got %q", tc.expected, got))
		}
	}

	// Compare with implicit-value enum where name=value
	if gen.StatusPending.String() != "Pending" {
		panic(fmt.Sprintf("implicit enum String() should return 'Pending', got %q", gen.StatusPending.String()))
	}
}
