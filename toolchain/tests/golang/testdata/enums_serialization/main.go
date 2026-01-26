// Verifies enum serialization: enums should be transmitted as strings on the wire,
// and round-trip correctly through client->server->client.
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

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Test 1: Verify enums are serialized as strings on the wire
	testWireFormat(ts.URL)

	// Test 2: Verify generated client handles enums correctly
	testGeneratedClient(ts.URL)

	// Test 3: Verify all enum values
	testAllEnumValues(ts.URL)

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
