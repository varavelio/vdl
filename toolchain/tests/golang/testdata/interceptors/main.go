// Verifies client interceptors: the interceptor adds a context value that a HeaderProvider
// reads and converts into a header, which the server then echoes back in the response.
package main

import (
	"context"
	"e2e/gen"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type AppProps struct {
	Intercepted bool
}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Procs.Test().Handle(func(c *gen.ServiceTestHandlerContext[AppProps]) (gen.ServiceTestOutput, error) {
		return gen.ServiceTestOutput{Intercepted: c.Props.Intercepted}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		intercepted := r.Header.Get("X-Intercepted") == "true"
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{Intercepted: intercepted}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").
		WithInterceptor(func(ctx context.Context, req gen.RequestInfo, next gen.Invoker) (gen.Response[json.RawMessage], error) {
			ctx = context.WithValue(ctx, "intercepted", "true")
			return next(ctx, req)
		}).
		WithHeaderProvider(func(ctx context.Context, h http.Header) error {
			if v, ok := ctx.Value("intercepted").(string); ok && v == "true" {
				h.Set("X-Intercepted", "true")
			}
			return nil
		}).
		Build()

	res, err := client.RPCs.Service().Procs.Test().Execute(context.Background(), gen.ServiceTestInput{})
	if err != nil {
		panic(err)
	}
	if !res.Intercepted {
		panic("interceptor did not work")
	}

	fmt.Println("Success")
}
