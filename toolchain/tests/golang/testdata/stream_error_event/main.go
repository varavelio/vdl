// Verifies stream error events: when a handler returns an error, the client
// receives it as an error event with the expected message.
package main

import (
	"context"
	"e2e/gen"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

type AppProps struct{}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Streams.Data().Handle(func(c *gen.ServiceDataHandlerContext[AppProps], emit gen.ServiceDataEmitFunc[AppProps]) error {
		time.Sleep(50 * time.Millisecond)
		return errors.New("something went wrong")
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()
	stream := client.RPCs.Service().Streams.Data().
		WithReconnectConfig(gen.ReconnectConfig{MaxAttempts: 0}).
		Execute(context.Background(), gen.ServiceDataInput{})

	select {
	case evt := <-stream:
		if evt.Ok {
			panic("expected error event, got success")
		}
		if evt.Error.Message != "something went wrong" {
			panic(fmt.Sprintf("expected 'something went wrong', got '%s'", evt.Error.Message))
		}
	case <-time.After(1 * time.Second):
		panic("timeout waiting for error event")
	}

	fmt.Println("Success")
}
