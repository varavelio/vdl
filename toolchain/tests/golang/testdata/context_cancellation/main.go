// Verifies context cancellation behavior: when the client cancels the context
// mid-stream, the server should stop sending events and clean up resources.
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
	var emitCount atomic.Int32
	var cleanedUp atomic.Bool

	server := gen.NewServer[AppProps]()

	server.SetStreamConfig(gen.StreamConfig{
		PingInterval: 10 * time.Second, // Disable pings
	})

	server.RPCs.Service().Streams.Counter().Handle(func(c *gen.ServiceCounterHandlerContext[AppProps], emit gen.ServiceCounterEmitFunc[AppProps]) error {
		defer func() {
			cleanedUp.Store(true)
		}()

		for i := 0; ; i++ {
			select {
			case <-c.Context.Done():
				return c.Context.Err()
			default:
				if err := emit(c, gen.ServiceCounterOutput{Count: int64(i)}); err != nil {
					return err
				}
				emitCount.Add(1)
				time.Sleep(50 * time.Millisecond)
			}
		}
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	stream := client.RPCs.Service().Streams.Counter().Execute(ctx, gen.ServiceCounterInput{})

	// Receive a few events
	receivedCount := 0
	for evt := range stream {
		if !evt.Ok {
			// If cancelled, stream should end
			break
		}
		receivedCount++
		if receivedCount >= 3 {
			// Cancel after receiving 3 events
			cancel()
		}
	}

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	if receivedCount < 3 {
		panic(fmt.Sprintf("expected at least 3 events before cancel, got %d", receivedCount))
	}

	if !cleanedUp.Load() {
		panic("expected server handler to clean up after cancellation")
	}

	// Verify server stopped emitting (count should stabilize)
	countBefore := emitCount.Load()
	time.Sleep(200 * time.Millisecond)
	countAfter := emitCount.Load()

	if countAfter > countBefore+2 {
		panic(fmt.Sprintf("server continued emitting after cancellation: before=%d, after=%d", countBefore, countAfter))
	}

	fmt.Println("Success")
}
