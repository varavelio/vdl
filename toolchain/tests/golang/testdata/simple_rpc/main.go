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
	// 1. Setup Server
	server := gen.NewServer[AppProps]()

	server.RPCs.Calculator().Procs.Add().Handle(func(c *gen.CalculatorAddHandlerContext[AppProps]) (gen.CalculatorAddOutput, error) {
		return gen.CalculatorAddOutput{
			Sum: c.Input.A + c.Input.B,
		}, nil
	})

	mux := http.NewServeMux()
	// Use wildcard pattern for VDL routing (Go 1.22+)
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		rpcName := r.PathValue("rpc")
		procName := r.PathValue("proc")
		adapter := gen.NewNetHTTPAdapter(w, r)
		if err := server.HandleRequest(r.Context(), AppProps{}, rpcName, procName, adapter); err != nil {
			fmt.Printf("Server Error: %v\n", err)
		}
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// 2. Setup Client (pointing to /rpc base URL)
	client := gen.NewClient(ts.URL + "/rpc").Build()

	// 3. Execute
	ctx := context.Background()
	output, err := client.RPCs.Calculator().Procs.Add().Execute(ctx, gen.CalculatorAddInput{
		A: 10,
		B: 32,
	})

	if err != nil {
		panic(fmt.Sprintf("RPC failed: %v", err))
	}

	if output.Sum != 42 {
		panic(fmt.Sprintf("Expected 42, got %d", output.Sum))
	}

	fmt.Println("Success: 10 + 32 = 42")
}
