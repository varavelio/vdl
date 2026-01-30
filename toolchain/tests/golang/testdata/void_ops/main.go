// Verifies procs with empty input/output (void operations) work correctly.
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

	server.RPCs.Service().Procs.Ping().Handle(func(c *gen.ServicePingHandlerContext[AppProps]) (gen.ServicePingOutput, error) {
		return gen.ServicePingOutput{}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()

	if _, err := client.RPCs.Service().Procs.Ping().Execute(context.Background(), gen.ServicePingInput{}); err != nil {
		panic(err)
	}

	fmt.Println("Success")
}
