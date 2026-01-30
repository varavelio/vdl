// Verifies stream lifecycle hooks: onConnect, onDisconnect, and onReconnect.
// The server fails with 500 on the first 2 attempts, then succeeds and emits events.
package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"
)

type AppProps struct{}

func main() {
	// Test 1: onConnect and onDisconnect with normal completion
	testConnectDisconnect()

	// Test 2: onDisconnect with context cancellation
	testDisconnectOnCancel()

	// Test 3: onReconnect with flaky server
	testReconnect()

	fmt.Println("All tests passed!")
}

func testConnectDisconnect() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Streams.Events().Handle(func(c *gen.ServiceEventsHandlerContext[AppProps], emit gen.ServiceEventsEmitFunc[AppProps]) error {
		// Emit 3 events
		for range 3 {
			if err := emit(c, gen.ServiceEventsOutput{}); err != nil {
				return err
			}
		}
		return nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()

	var connectCalled, disconnectCalled atomic.Bool
	var disconnectHadError atomic.Bool

	stream := client.RPCs.Service().Streams.Events().
		OnConnect(func() {
			connectCalled.Store(true)
		}).
		OnDisconnect(func(err error) {
			disconnectCalled.Store(true)
			disconnectHadError.Store(err != nil)
		}).
		Execute(context.Background(), gen.ServiceEventsInput{})

	// Consume all events
	eventCount := 0
	for range stream {
		eventCount++
	}

	// Give a moment for disconnect callback
	time.Sleep(10 * time.Millisecond)

	if !connectCalled.Load() {
		panic("Test 1 FAILED: onConnect was not called")
	}

	if !disconnectCalled.Load() {
		panic("Test 1 FAILED: onDisconnect was not called")
	}

	// Normal completion should have nil error
	if disconnectHadError.Load() {
		panic("Test 1 FAILED: onDisconnect should have nil error on normal completion")
	}

	if eventCount != 3 {
		panic(fmt.Sprintf("Test 1 FAILED: expected 3 events, got %d", eventCount))
	}

	fmt.Println("Test 1 passed: onConnect and onDisconnect work correctly")
}

func testDisconnectOnCancel() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Streams.Events().Handle(func(c *gen.ServiceEventsHandlerContext[AppProps], emit gen.ServiceEventsEmitFunc[AppProps]) error {
		// Emit many events slowly
		for range 100 {
			if err := emit(c, gen.ServiceEventsOutput{}); err != nil {
				return err
			}
			time.Sleep(10 * time.Millisecond)
		}
		return nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()

	var disconnectCalled atomic.Bool

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream := client.RPCs.Service().Streams.Events().
		OnDisconnect(func(err error) {
			disconnectCalled.Store(true)
		}).
		Execute(ctx, gen.ServiceEventsInput{})

	// Receive 2 events then cancel
	eventCount := 0
	for range stream {
		eventCount++
		if eventCount >= 2 {
			cancel()
			// Drain remaining events after cancel to allow cleanup
			for range stream {
			}
			break
		}
	}

	// Give more time for disconnect callback
	time.Sleep(100 * time.Millisecond)

	if !disconnectCalled.Load() {
		panic("Test 2 FAILED: onDisconnect was not called after context cancellation")
	}

	fmt.Println("Test 2 passed: onDisconnect called after context cancellation")
}

func testReconnect() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Streams.Events().Handle(func(c *gen.ServiceEventsHandlerContext[AppProps], emit gen.ServiceEventsEmitFunc[AppProps]) error {
		// Emit 2 events
		for range 2 {
			if err := emit(c, gen.ServiceEventsOutput{}); err != nil {
				return err
			}
		}
		return nil
	})

	var requestCount atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		// Fail first request to trigger reconnect
		if count == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()

	var connectCount, reconnectCount atomic.Int32
	var reconnectAttempt atomic.Int32
	var reconnectDelay atomic.Int64

	conf := gen.ReconnectConfig{
		MaxAttempts:     5,
		InitialDelay:    50 * time.Millisecond,
		MaxDelay:        200 * time.Millisecond,
		DelayMultiplier: 1.5,
		Jitter:          0.0,
	}

	stream := client.RPCs.Service().Streams.Events().
		WithReconnectConfig(conf).
		OnConnect(func() {
			connectCount.Add(1)
		}).
		OnReconnect(func(attempt int, wait time.Duration) {
			reconnectCount.Add(1)
			reconnectAttempt.Store(int32(attempt))
			reconnectDelay.Store(int64(wait))
		}).
		Execute(context.Background(), gen.ServiceEventsInput{})

	// Consume all events
	eventCount := 0
	for range stream {
		eventCount++
	}

	if reconnectCount.Load() < 1 {
		panic(fmt.Sprintf("Test 3 FAILED: expected at least 1 reconnect, got %d", reconnectCount.Load()))
	}

	if reconnectAttempt.Load() != 1 {
		panic(fmt.Sprintf("Test 3 FAILED: expected reconnect attempt 1, got %d", reconnectAttempt.Load()))
	}

	if reconnectDelay.Load() <= 0 {
		panic(fmt.Sprintf("Test 3 FAILED: expected positive delay, got %d", reconnectDelay.Load()))
	}

	if connectCount.Load() != 1 {
		panic(fmt.Sprintf("Test 3 FAILED: expected 1 connect, got %d", connectCount.Load()))
	}

	if eventCount != 2 {
		panic(fmt.Sprintf("Test 3 FAILED: expected 2 events after reconnect, got %d", eventCount))
	}

	fmt.Println("Test 3 passed: onReconnect called on server failure")
}
