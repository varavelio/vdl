// Verifies middleware execution order: Global -> RPC -> Proc -> Handler.
// The trace captured at handler execution time shows the "Pre" middlewares in order.
package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type AppProps struct {
	Trace *[]string
}

func main() {
	server := gen.NewServer[AppProps]()

	server.Use(func(next gen.GlobalHandlerFunc[AppProps]) gen.GlobalHandlerFunc[AppProps] {
		return func(c *gen.HandlerContext[AppProps, any]) (any, error) {
			*c.Props.Trace = append(*c.Props.Trace, "GlobalPre")
			res, err := next(c)
			*c.Props.Trace = append(*c.Props.Trace, "GlobalPost")
			return res, err
		}
	})

	server.RPCs.Service().Use(func(next gen.GlobalHandlerFunc[AppProps]) gen.GlobalHandlerFunc[AppProps] {
		return func(c *gen.HandlerContext[AppProps, any]) (any, error) {
			*c.Props.Trace = append(*c.Props.Trace, "RPCPre")
			res, err := next(c)
			*c.Props.Trace = append(*c.Props.Trace, "RPCPost")
			return res, err
		}
	})

	server.RPCs.Service().Procs.Test().Use(func(next gen.ServiceTestHandlerFunc[AppProps]) gen.ServiceTestHandlerFunc[AppProps] {
		return func(c *gen.ServiceTestHandlerContext[AppProps]) (gen.ServiceTestOutput, error) {
			*c.Props.Trace = append(*c.Props.Trace, "ProcPre")
			res, err := next(c)
			*c.Props.Trace = append(*c.Props.Trace, "ProcPost")
			return res, err
		}
	})

	server.RPCs.Service().Procs.Test().Handle(func(c *gen.ServiceTestHandlerContext[AppProps]) (gen.ServiceTestOutput, error) {
		*c.Props.Trace = append(*c.Props.Trace, "Handler")
		return gen.ServiceTestOutput{Trace: *c.Props.Trace}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		trace := make([]string, 0)
		props := AppProps{Trace: &trace}
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), props, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()
	res, err := client.RPCs.Service().Procs.Test().Execute(context.Background(), gen.ServiceTestInput{})
	if err != nil {
		panic(err)
	}

	expected := []string{"GlobalPre", "RPCPre", "ProcPre", "Handler"}
	if len(res.Trace) != len(expected) {
		panic(fmt.Sprintf("trace mismatch: got %v, want %v", res.Trace, expected))
	}
	for i, v := range expected {
		if res.Trace[i] != v {
			panic(fmt.Sprintf("mismatch at index %d: got %s, want %s", i, res.Trace[i], v))
		}
	}

	fmt.Println("Success")
}
