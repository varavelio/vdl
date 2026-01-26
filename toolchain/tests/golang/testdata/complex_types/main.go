// Verifies complex type serialization: deeply nested structures, maps of arrays,
// arrays of maps, nested objects, and multi-dimensional arrays.
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
		return gen.ServiceEchoOutput{Data: c.Input.Data}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()

	input := buildComplexInput()
	res, err := client.RPCs.Service().Procs.Echo().Execute(context.Background(), gen.ServiceEchoInput{Data: input})
	if err != nil {
		panic(err)
	}

	if !reflect.DeepEqual(res.Data, input) {
		panic(fmt.Sprintf("data mismatch:\nsent: %+v\ngot:  %+v", input, res.Data))
	}

	fmt.Println("Success")
}

func buildComplexInput() gen.Container {
	return gen.Container{
		User: gen.User{
			Id:       123,
			Username: "alice",
			IsActive: true,
			Tags:     []string{"admin", "editor", "viewer"},
			Metadata: map[string]string{"role": "superuser", "level": "9000"},
			Address:  gen.Some(gen.Address{Street: "123 Main St", City: "Wonderland", Zip: 90210}),
		},
		Matrix: [][]int64{
			{1, 2, 3},
			{4, 5, 6},
			{7, 8, 9},
		},
		NestedArrays: [][][]string{
			{{"a", "b"}, {"c", "d"}},
			{{"e", "f"}, {"g", "h"}},
		},
		MapOfArrays: map[string][]int64{
			"primes":    {2, 3, 5, 7, 11},
			"fibonacci": {1, 1, 2, 3, 5, 8},
		},
		ArrayOfMaps: []map[string]string{
			{"key1": "value1", "key2": "value2"},
			{"key3": "value3"},
		},
		MapOfObjects: map[string]gen.User{
			"alice": {Id: 1, Username: "alice", IsActive: true, Tags: []string{"a"}, Metadata: map[string]string{}},
			"bob":   {Id: 2, Username: "bob", IsActive: false, Tags: []string{"b"}, Metadata: map[string]string{"x": "y"}},
		},
		DeepNest: gen.Level1{
			Id: 1,
			Level2: gen.Level2{
				Name: "level2",
				Level3: gen.Level3{
					Value: "deepest",
				},
			},
		},
		Points: []gen.Point{
			{X: 10, Y: 20},
			{X: 30, Y: 40},
		},
		Scores: map[string]gen.Score{
			"player1": {Name: "Alice", Value: 95.5},
			"player2": {Name: "Bob", Value: 87.3},
		},
		ArrayOfMapOfArrays: []map[string][]int64{
			{"set1": {1, 2, 3}, "set2": {4, 5}},
			{"set3": {6, 7, 8, 9}},
		},
	}
}
