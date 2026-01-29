// Verifies enum serialization: both string enums and int enums are echoed correctly.
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
	// Test ColorList contains all Color values
	if len(gen.ColorList) != 3 {
		panic(fmt.Sprintf("expected ColorList to have 3 elements, got %d", len(gen.ColorList)))
	}
	expectedColors := []gen.Color{gen.ColorRed, gen.ColorGreen, gen.ColorBlue}
	for i, c := range gen.ColorList {
		if c != expectedColors[i] {
			panic(fmt.Sprintf("ColorList[%d]: expected %s, got %s", i, expectedColors[i], c))
		}
	}

	// Test PriorityList contains all Priority values
	if len(gen.PriorityList) != 2 {
		panic(fmt.Sprintf("expected PriorityList to have 2 elements, got %d", len(gen.PriorityList)))
	}
	expectedPriorities := []gen.Priority{gen.PriorityLow, gen.PriorityHigh}
	for i, p := range gen.PriorityList {
		if p != expectedPriorities[i] {
			panic(fmt.Sprintf("PriorityList[%d]: expected %d, got %d", i, expectedPriorities[i], p))
		}
	}

	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Procs.Test().Handle(func(c *gen.ServiceTestHandlerContext[AppProps]) (gen.ServiceTestOutput, error) {
		return gen.ServiceTestOutput{C: c.Input.C, P: c.Input.P}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()

	testCases := []struct {
		color    gen.Color
		priority gen.Priority
	}{
		{gen.ColorRed, gen.PriorityHigh},
		{gen.ColorBlue, gen.PriorityLow},
	}

	for _, tc := range testCases {
		res, err := client.RPCs.Service().Procs.Test().Execute(context.Background(), gen.ServiceTestInput{C: tc.color, P: tc.priority})
		if err != nil {
			panic(err)
		}
		if res.C != tc.color {
			panic(fmt.Sprintf("expected color %s, got %s", tc.color, res.C))
		}
		if res.P != tc.priority {
			panic(fmt.Sprintf("expected priority %d, got %d", tc.priority, res.P))
		}
	}

	fmt.Println("Success")
}
