// Verifies SSE ping handling: the server sends pings during a long-running stream.
// Uses a raw HTTP client to verify pings are actually on the wire (the generated client filters them out).
package main

import (
	"bufio"
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

type AppProps struct{}

func main() {
	// Test 1: Raw HTTP client to verify pings are actually sent
	testRawPings()

	// Test 2: Generated client correctly filters pings and receives events
	testGeneratedClient()

	fmt.Println("Success")
}

func createServer() (*gen.Server[AppProps], *httptest.Server) {
	server := gen.NewServer[AppProps]()

	server.SetStreamConfig(gen.StreamConfig{
		PingInterval: 50 * time.Millisecond,
	})

	server.RPCs.Clock().Streams.Ticks().Handle(func(c *gen.ClockTicksHandlerContext[AppProps], emit gen.ClockTicksEmitFunc[AppProps]) error {
		time.Sleep(200 * time.Millisecond)
		if err := emit(c, gen.ClockTicksOutput{Iso: "event1"}); err != nil {
			return err
		}
		time.Sleep(200 * time.Millisecond)
		return emit(c, gen.ClockTicksOutput{Iso: "event2"})
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	return server, ts
}

func testRawPings() {
	_, ts := createServer()
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", ts.URL+"/rpc/Clock/Ticks", strings.NewReader("{}"))
	if err != nil {
		panic(fmt.Sprintf("failed to create request: %v", err))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(fmt.Sprintf("request failed: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
	}

	scanner := bufio.NewScanner(resp.Body)
	var pingCount, eventCount int

	for scanner.Scan() {
		line := scanner.Text()

		if line == ": ping" {
			pingCount++
		}
		if strings.HasPrefix(line, "data:") {
			eventCount++
		}
	}

	// Ignore scanner errors from context cancellation or connection close
	// The stream naturally ends when handler completes

	// With 50ms ping interval and ~400ms total stream time, expect at least 3 pings
	if pingCount < 3 {
		panic(fmt.Sprintf("expected at least 3 pings, got %d", pingCount))
	}

	if eventCount != 2 {
		panic(fmt.Sprintf("expected 2 events, got %d", eventCount))
	}
}

func testGeneratedClient() {
	_, ts := createServer()
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stream := client.RPCs.Clock().Streams.Ticks().Execute(ctx, gen.ClockTicksInput{})

	var received []string
	for evt := range stream {
		if !evt.Ok {
			panic(fmt.Sprintf("stream error: %v", evt.Error))
		}
		received = append(received, evt.Output.Iso)
	}

	if len(received) != 2 || received[0] != "event1" || received[1] != "event2" {
		panic(fmt.Sprintf("expected [event1, event2], got %v", received))
	}
}
