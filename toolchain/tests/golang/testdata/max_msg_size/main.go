// Verifies max message size enforcement: server sends a 2MB payload but client
// is configured with 1MB limit, expecting a MESSAGE_TOO_LARGE error.
package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

type AppProps struct{}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Streams.Data().Handle(func(c *gen.ServiceDataHandlerContext[AppProps], emit gen.ServiceDataEmitFunc[AppProps]) error {
		payload := strings.Repeat("a", 2*1024*1024)
		return emit(c, gen.ServiceDataOutput{Payload: payload})
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
		WithMaxMessageSize(1*1024*1024).
		Execute(context.Background(), gen.ServiceDataInput{})

	evt := <-stream
	if evt.Ok {
		panic("expected error due to message size, got success")
	}
	if evt.Error.Code != "MESSAGE_TOO_LARGE" {
		panic(fmt.Sprintf("expected MESSAGE_TOO_LARGE, got %s", evt.Error.Code))
	}

	fmt.Println("Success")
}
