// Verifies SSE ping handling: the server sends pings during a long-running stream,
// and the client should correctly filter them out and still receive the actual event.
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
	server := gen.NewServer[AppProps]()

	server.SetStreamConfig(gen.StreamConfig{
		PingInterval: 100 * time.Millisecond,
	})

	server.RPCs.Clock().Streams.Ticks().Handle(func(c *gen.ClockTicksHandlerContext[AppProps], emit gen.ClockTicksEmitFunc[AppProps]) error {
		time.Sleep(500 * time.Millisecond)
		return emit(c, gen.ClockTicksOutput{Iso: time.Now().Format(time.RFC3339)})
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stream := client.RPCs.Clock().Streams.Ticks().Execute(ctx, gen.ClockTicksInput{})

	select {
	case evt := <-stream:
		if !evt.Ok {
			panic(fmt.Sprintf("stream error: %v", evt.Error))
		}
		fmt.Printf("Tick received: %s\n", evt.Output.Iso)
	case <-ctx.Done():
		panic("timeout waiting for event")
	}

	fmt.Println("Success")
}
