// Verifies dynamic header providers: headers are computed per-request using a counter
// that increments on each call, and the server echoes the header value back.
package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
)

type AppProps struct {
	Count string
}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Procs.Get().Handle(func(c *gen.ServiceGetHandlerContext[AppProps]) (gen.ServiceGetOutput, error) {
		return gen.ServiceGetOutput{Val: c.Props.Count}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		props := AppProps{Count: r.Header.Get("X-Count")}
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), props, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	var count atomic.Int32
	client := gen.NewClient(ts.URL + "/rpc").
		WithHeaderProvider(func(ctx context.Context, h http.Header) error {
			h.Set("X-Count", fmt.Sprintf("%d", count.Add(1)))
			return nil
		}).
		Build()

	for expected := 1; expected <= 3; expected++ {
		res, _ := client.RPCs.Service().Procs.Get().Execute(context.Background(), gen.ServiceGetInput{})
		if res.Val != fmt.Sprintf("%d", expected) {
			panic(fmt.Sprintf("expected %d, got %s", expected, res.Val))
		}
	}

	fmt.Println("Success")
}
