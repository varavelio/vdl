// Verifies stream reconnection lifecycle: the server fails with 500 on the first
// 2 attempts, then succeeds. The client should reconnect and eventually receive events.
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
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Streams.Events().Handle(func(c *gen.ServiceEventsHandlerContext[AppProps], emit gen.ServiceEventsEmitFunc[AppProps]) error {
		return nil
	})

	var requestCount atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		if count <= 2 {
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

	conf := gen.ReconnectConfig{
		MaxAttempts:     5,
		InitialDelay:    10 * time.Millisecond,
		MaxDelay:        100 * time.Millisecond,
		DelayMultiplier: 1.0,
		Jitter:          0.0,
	}

	stream := client.RPCs.Service().Streams.Events().
		WithReconnectConfig(conf).
		OnConnect(func() { connectCount.Add(1) }).
		OnReconnect(func(attempt int, wait time.Duration) { reconnectCount.Add(1) }).
		Execute(context.Background(), gen.ServiceEventsInput{})

	<-stream

	if reconnectCount.Load() < 2 {
		panic(fmt.Sprintf("expected at least 2 reconnects, got %d", reconnectCount.Load()))
	}
	if connectCount.Load() != 1 {
		panic(fmt.Sprintf("expected 1 connect, got %d", connectCount.Load()))
	}

	fmt.Println("Success")
}
