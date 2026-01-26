// Verifies custom header propagation: X-Trace-ID and other headers
// should be correctly passed from client to server and accessible in handlers.
package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type AppProps struct {
	TraceID string
}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Procs.Echo().Handle(func(c *gen.ServiceEchoHandlerContext[AppProps]) (gen.ServiceEchoOutput, error) {
		return gen.ServiceEchoOutput{
			Data:            c.Input.Data,
			ReceivedTraceId: c.Props.TraceID,
		}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-ID")
		props := AppProps{TraceID: traceID}
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), props, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Test 1: With trace ID header at client level
	testClientLevelHeader(ts.URL)

	// Test 2: Without trace ID header
	testWithoutTraceID(ts.URL)

	// Test 3: With request-level header
	testRequestLevelHeader(ts.URL)

	fmt.Println("Success")
}

func testClientLevelHeader(baseURL string) {
	client := gen.NewClient(baseURL+"/rpc").
		WithHeader("X-Trace-ID", "trace-123-abc").
		Build()

	ctx := context.Background()
	result, err := client.RPCs.Service().Procs.Echo().Execute(ctx, gen.ServiceEchoInput{
		Data: "hello",
	})
	if err != nil {
		panic(fmt.Sprintf("execute failed: %v", err))
	}

	if result.ReceivedTraceId != "trace-123-abc" {
		panic(fmt.Sprintf("expected ReceivedTraceId='trace-123-abc', got: %s", result.ReceivedTraceId))
	}
}

func testWithoutTraceID(baseURL string) {
	client := gen.NewClient(baseURL + "/rpc").Build()

	ctx := context.Background()
	result, err := client.RPCs.Service().Procs.Echo().Execute(ctx, gen.ServiceEchoInput{
		Data: "hello",
	})
	if err != nil {
		panic(fmt.Sprintf("execute failed: %v", err))
	}

	if result.ReceivedTraceId != "" {
		panic(fmt.Sprintf("expected ReceivedTraceId='', got: %s", result.ReceivedTraceId))
	}
}

func testRequestLevelHeader(baseURL string) {
	client := gen.NewClient(baseURL + "/rpc").Build()

	ctx := context.Background()
	result, err := client.RPCs.Service().Procs.Echo().
		WithHeader("X-Trace-ID", "trace-456-def").
		Execute(ctx, gen.ServiceEchoInput{
			Data: "test",
		})
	if err != nil {
		panic(fmt.Sprintf("execute failed: %v", err))
	}

	if result.ReceivedTraceId != "trace-456-def" {
		panic(fmt.Sprintf("expected ReceivedTraceId='trace-456-def', got: %s", result.ReceivedTraceId))
	}
}
