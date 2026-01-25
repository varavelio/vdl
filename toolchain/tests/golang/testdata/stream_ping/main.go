package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

type AppProps struct{}

func main() {
	// 1. Setup Server
	server := gen.NewServer[AppProps]()

	// Configure aggressive ping to force pings during the test
	server.SetStreamConfig(gen.StreamConfig{
		PingInterval: 100 * time.Millisecond,
	})

	server.RPCs.Clock().Streams.Ticks().Handle(func(c *gen.ClockTicksHandlerContext[AppProps], emit gen.ClockTicksEmitFunc[AppProps]) error {
		// Wait for 500ms (5 pings should be sent)
		time.Sleep(500 * time.Millisecond)

		return emit(c, gen.ClockTicksOutput{Iso: time.Now().Format(time.RFC3339)})
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		rpcName := r.PathValue("rpc")
		procName := r.PathValue("proc")
		adapter := gen.NewNetHTTPAdapter(w, r)
		if err := server.HandleRequest(r.Context(), AppProps{}, rpcName, procName, adapter); err != nil {
			fmt.Printf("Server Error: %v\n", err)
		}
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// 2. Setup Client
	client := gen.NewClient(ts.URL + "/rpc").Build()

	// 3. Execute
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stream := client.RPCs.Clock().Streams.Ticks().Execute(ctx, gen.ClockTicksInput{})

	fmt.Println("Waiting for tick...")

	// We expect exactly one event
	select {
	case evt := <-stream:
		if !evt.Ok {
			panic(fmt.Sprintf("Stream error: %v", evt.Error))
		}
		fmt.Printf("Tick received: %s\n", evt.Output.Iso)
	case <-ctx.Done():
		panic("Timeout waiting for event (did pings break the JSON parser?)")
	}

	fmt.Println("Success")
}
