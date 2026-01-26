// Verifies serialization of complex types: nested structs, arrays, maps, and optional fields.
package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
)

type AppProps struct{}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Procs.Echo().Handle(func(c *gen.ServiceEchoHandlerContext[AppProps]) (gen.ServiceEchoOutput, error) {
		return gen.ServiceEchoOutput{
			User:   c.Input.User,
			Matrix: c.Input.Matrix,
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

	inputUser := gen.User{
		Id:       123,
		Username: "alice",
		IsActive: true,
		Tags:     []string{"admin", "editor"},
		Metadata: map[string]string{"role": "superuser", "level": "9000"},
		Address:  gen.Some(gen.Address{Street: "123 Main St", City: "Wonderland", Zip: 90210}),
	}
	matrix := [][]int64{{1, 2}, {3, 4}}

	input := gen.ServiceEchoInput{User: inputUser, Matrix: matrix}
	res, err := client.RPCs.Service().Procs.Echo().Execute(context.Background(), input)
	if err != nil {
		panic(err)
	}

	if !reflect.DeepEqual(res.User, inputUser) {
		panic(fmt.Sprintf("user mismatch: sent %+v, got %+v", inputUser, res.User))
	}
	if !reflect.DeepEqual(res.Matrix, matrix) {
		panic(fmt.Sprintf("matrix mismatch: sent %+v, got %+v", matrix, res.Matrix))
	}

	fmt.Println("Success")
}
