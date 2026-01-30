// Verifies basic RPC functionality: a simple Add proc that sums two numbers.
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

	server.RPCs.Calculator().Procs.Add().Handle(func(c *gen.CalculatorAddHandlerContext[AppProps]) (gen.CalculatorAddOutput, error) {
		return gen.CalculatorAddOutput{Sum: c.Input.A + c.Input.B}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()
	output, err := client.RPCs.Calculator().Procs.Add().Execute(context.Background(), gen.CalculatorAddInput{A: 10, B: 32})
	if err != nil {
		panic(err)
	}
	if output.Sum != 42 {
		panic(fmt.Sprintf("expected 42, got %d", output.Sum))
	}

	fmt.Println("Success")
}
