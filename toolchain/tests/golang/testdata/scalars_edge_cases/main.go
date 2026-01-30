// Verifies scalar edge cases: zero values (0, 0.0, false, "", zero datetime)
// should be correctly transmitted and distinguishable from absent values.
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
	"time"
)

type AppProps struct{}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Procs.Echo().Handle(func(c *gen.ServiceEchoHandlerContext[AppProps]) (gen.ServiceEchoOutput, error) {
		return gen.ServiceEchoOutput{
			IntVal:      c.Input.IntVal,
			FloatVal:    c.Input.FloatVal,
			BoolVal:     c.Input.BoolVal,
			StringVal:   c.Input.StringVal,
			DatetimeVal: c.Input.DatetimeVal,
		}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Test 1: Zero values via raw HTTP
	testZeroValuesRaw(ts.URL)

	// Test 2: Zero values via generated client
	testZeroValuesClient(ts.URL)

	// Test 3: Non-zero values
	testNonZeroValues(ts.URL)

	fmt.Println("Success")
}

func testZeroValuesRaw(baseURL string) {
	// Zero datetime in ISO8601 format
	zeroTime := time.Time{}.UTC().Format(time.RFC3339)

	payload := fmt.Sprintf(`{
		"intVal": 0,
		"floatVal": 0.0,
		"boolVal": false,
		"stringVal": "",
		"datetimeVal": "%s"
	}`, zeroTime)

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

	// Verify int zero
	if output["intVal"] != float64(0) {
		panic(fmt.Sprintf("expected intVal=0, got: %v (%T)", output["intVal"], output["intVal"]))
	}

	// Verify float zero
	if output["floatVal"] != float64(0) {
		panic(fmt.Sprintf("expected floatVal=0, got: %v", output["floatVal"]))
	}

	// Verify bool false
	if output["boolVal"] != false {
		panic(fmt.Sprintf("expected boolVal=false, got: %v", output["boolVal"]))
	}

	// Verify empty string
	if output["stringVal"] != "" {
		panic(fmt.Sprintf("expected stringVal='', got: %v", output["stringVal"]))
	}

	// Verify datetime is present (format may vary)
	if output["datetimeVal"] == nil {
		panic("expected datetimeVal to be present")
	}
}

func testZeroValuesClient(baseURL string) {
	client := gen.NewClient(baseURL + "/rpc").Build()
	ctx := context.Background()

	zeroTime := time.Time{}.UTC()

	result, err := client.RPCs.Service().Procs.Echo().Execute(ctx, gen.ServiceEchoInput{
		IntVal:      0,
		FloatVal:    0.0,
		BoolVal:     false,
		StringVal:   "",
		DatetimeVal: zeroTime,
	})
	if err != nil {
		panic(fmt.Sprintf("execute failed: %v", err))
	}

	if result.IntVal != 0 {
		panic(fmt.Sprintf("expected IntVal=0, got: %d", result.IntVal))
	}
	if result.FloatVal != 0.0 {
		panic(fmt.Sprintf("expected FloatVal=0.0, got: %f", result.FloatVal))
	}
	if result.BoolVal != false {
		panic(fmt.Sprintf("expected BoolVal=false, got: %v", result.BoolVal))
	}
	if result.StringVal != "" {
		panic(fmt.Sprintf("expected StringVal='', got: %s", result.StringVal))
	}
	if !result.DatetimeVal.Equal(zeroTime) {
		panic(fmt.Sprintf("expected DatetimeVal=%v, got: %v", zeroTime, result.DatetimeVal))
	}
}

func testNonZeroValues(baseURL string) {
	client := gen.NewClient(baseURL + "/rpc").Build()
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)

	result, err := client.RPCs.Service().Procs.Echo().Execute(ctx, gen.ServiceEchoInput{
		IntVal:      42,
		FloatVal:    3.14159,
		BoolVal:     true,
		StringVal:   "hello world",
		DatetimeVal: now,
	})
	if err != nil {
		panic(fmt.Sprintf("execute failed: %v", err))
	}

	if result.IntVal != 42 {
		panic(fmt.Sprintf("expected IntVal=42, got: %d", result.IntVal))
	}
	if result.FloatVal != 3.14159 {
		panic(fmt.Sprintf("expected FloatVal=3.14159, got: %f", result.FloatVal))
	}
	if result.BoolVal != true {
		panic(fmt.Sprintf("expected BoolVal=true, got: %v", result.BoolVal))
	}
	if result.StringVal != "hello world" {
		panic(fmt.Sprintf("expected StringVal='hello world', got: %s", result.StringVal))
	}
	if !result.DatetimeVal.Equal(now) {
		panic(fmt.Sprintf("expected DatetimeVal=%v, got: %v", now, result.DatetimeVal))
	}
}
