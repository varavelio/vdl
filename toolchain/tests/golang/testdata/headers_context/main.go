// Verifies header passing through context: client-level and request-level headers
// are extracted by the HTTP handler into Props and returned by the server.
package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type AppProps struct {
	Headers map[string]string
}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Procs.GetHeaders().Handle(func(c *gen.ServiceGetHeadersHandlerContext[AppProps]) (gen.ServiceGetHeadersOutput, error) {
		return gen.ServiceGetHeadersOutput{Values: c.Props.Headers}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		headers := map[string]string{
			"X-Custom":      r.Header.Get("X-Custom"),
			"Authorization": r.Header.Get("Authorization"),
		}
		props := AppProps{Headers: headers}
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), props, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL+"/rpc").
		WithHeader("Authorization", "Bearer secret").
		Build()

	res, err := client.RPCs.Service().Procs.GetHeaders().
		WithHeader("X-Custom", "123").
		Execute(context.Background(), gen.ServiceGetHeadersInput{})

	if err != nil {
		panic(err)
	}
	if res.Values["Authorization"] != "Bearer secret" {
		panic("missing Authorization header")
	}
	if res.Values["X-Custom"] != "123" {
		panic("missing X-Custom header")
	}

	fmt.Println("Success")
}
