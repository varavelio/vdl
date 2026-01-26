// Verifies middleware execution order and coverage across all levels:
// Global middleware applies to all RPCs, RPC middleware applies to all procs in that RPC,
// and Proc middleware applies only to specific procs.
package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

type AppProps struct {
	Trace *[]string
}

func main() {
	server := gen.NewServer[AppProps]()

	server.Use(func(next gen.GlobalHandlerFunc[AppProps]) gen.GlobalHandlerFunc[AppProps] {
		return func(c *gen.HandlerContext[AppProps, any]) (any, error) {
			*c.Props.Trace = append(*c.Props.Trace, "Global")
			return next(c)
		}
	})

	server.RPCs.ServiceA().Use(func(next gen.GlobalHandlerFunc[AppProps]) gen.GlobalHandlerFunc[AppProps] {
		return func(c *gen.HandlerContext[AppProps, any]) (any, error) {
			*c.Props.Trace = append(*c.Props.Trace, "RpcA")
			return next(c)
		}
	})

	server.RPCs.ServiceB().Use(func(next gen.GlobalHandlerFunc[AppProps]) gen.GlobalHandlerFunc[AppProps] {
		return func(c *gen.HandlerContext[AppProps, any]) (any, error) {
			*c.Props.Trace = append(*c.Props.Trace, "RpcB")
			return next(c)
		}
	})

	server.RPCs.ServiceA().Procs.Proc1().Use(func(next gen.ServiceAProc1HandlerFunc[AppProps]) gen.ServiceAProc1HandlerFunc[AppProps] {
		return func(c *gen.ServiceAProc1HandlerContext[AppProps]) (gen.ServiceAProc1Output, error) {
			*c.Props.Trace = append(*c.Props.Trace, "ProcA1")
			return next(c)
		}
	})

	server.RPCs.ServiceA().Procs.Proc1().Handle(func(c *gen.ServiceAProc1HandlerContext[AppProps]) (gen.ServiceAProc1Output, error) {
		*c.Props.Trace = append(*c.Props.Trace, "HandlerA1")
		return gen.ServiceAProc1Output{Trace: *c.Props.Trace}, nil
	})

	server.RPCs.ServiceA().Procs.Proc2().Handle(func(c *gen.ServiceAProc2HandlerContext[AppProps]) (gen.ServiceAProc2Output, error) {
		*c.Props.Trace = append(*c.Props.Trace, "HandlerA2")
		return gen.ServiceAProc2Output{Trace: *c.Props.Trace}, nil
	})

	server.RPCs.ServiceB().Procs.Proc1().Handle(func(c *gen.ServiceBProc1HandlerContext[AppProps]) (gen.ServiceBProc1Output, error) {
		*c.Props.Trace = append(*c.Props.Trace, "HandlerB1")
		return gen.ServiceBProc1Output{Trace: *c.Props.Trace}, nil
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
	ctx := context.Background()

	testCases := []struct {
		name     string
		call     func() ([]string, error)
		expected []string
	}{
		{
			name: "ServiceA.Proc1 (Global + RpcA + ProcA1)",
			call: func() ([]string, error) {
				res, err := client.RPCs.ServiceA().Procs.Proc1().Execute(ctx, gen.ServiceAProc1Input{})
				return res.Trace, err
			},
			expected: []string{"Global", "RpcA", "ProcA1", "HandlerA1"},
		},
		{
			name: "ServiceA.Proc2 (Global + RpcA, no proc middleware)",
			call: func() ([]string, error) {
				res, err := client.RPCs.ServiceA().Procs.Proc2().Execute(ctx, gen.ServiceAProc2Input{})
				return res.Trace, err
			},
			expected: []string{"Global", "RpcA", "HandlerA2"},
		},
		{
			name: "ServiceB.Proc1 (Global + RpcB, no proc middleware)",
			call: func() ([]string, error) {
				res, err := client.RPCs.ServiceB().Procs.Proc1().Execute(ctx, gen.ServiceBProc1Input{})
				return res.Trace, err
			},
			expected: []string{"Global", "RpcB", "HandlerB1"},
		},
	}

	for _, tc := range testCases {
		trace, err := tc.call()
		if err != nil {
			panic(fmt.Sprintf("%s: %v", tc.name, err))
		}
		if !equalSlices(trace, tc.expected) {
			panic(fmt.Sprintf("%s: expected %v, got %v", tc.name, tc.expected, trace))
		}
	}

	fmt.Println("Success")
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	return strings.Join(a, ",") == strings.Join(b, ",")
}
