// Verifies multiple RPC blocks with different procs work correctly.
package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type AppProps struct{}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.A().Procs.X().Handle(func(c *gen.AXHandlerContext[AppProps]) (gen.AXOutput, error) {
		return gen.AXOutput{}, nil
	})
	server.RPCs.B().Procs.Y().Handle(func(c *gen.BYHandlerContext[AppProps]) (gen.BYOutput, error) {
		return gen.BYOutput{}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()

	if _, err := client.RPCs.A().Procs.X().Execute(context.Background(), gen.AXInput{}); err != nil {
		panic(err)
	}
	if _, err := client.RPCs.B().Procs.Y().Execute(context.Background(), gen.BYInput{}); err != nil {
		panic(err)
	}

	fmt.Println("Success")
}
