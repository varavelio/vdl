// Verifies primitive type serialization: int, float, bool, string, and datetime.
package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

type AppProps struct{}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Procs.Echo().Handle(func(c *gen.ServiceEchoHandlerContext[AppProps]) (gen.ServiceEchoOutput, error) {
		return gen.ServiceEchoOutput{
			I: c.Input.I,
			F: c.Input.F,
			B: c.Input.B,
			S: c.Input.S,
			D: c.Input.D,
		}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()

	now := time.Now().Truncate(time.Second)
	input := gen.ServiceEchoInput{
		I: 42,
		F: 3.14159,
		B: true,
		S: "Hello VDL",
		D: now,
	}

	res, err := client.RPCs.Service().Procs.Echo().Execute(context.Background(), input)
	if err != nil {
		panic(err)
	}

	if res.I != input.I {
		panic("int mismatch")
	}
	if res.F != input.F {
		panic("float mismatch")
	}
	if res.B != input.B {
		panic("bool mismatch")
	}
	if res.S != input.S {
		panic("string mismatch")
	}
	if !res.D.Equal(input.D) {
		panic(fmt.Sprintf("datetime mismatch: sent %v, got %v", input.D, res.D))
	}

	fmt.Println("Success")
}
