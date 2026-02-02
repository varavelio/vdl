// Verifies complex type serialization: deeply nested structures, maps of arrays,
// arrays of maps, nested objects, multi-dimensional arrays, and inline object definitions.
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

	server.RPCs.Service().Procs.EchoInline().Handle(func(c *gen.ServiceEchoInlineHandlerContext[AppProps]) (gen.ServiceEchoInlineOutput, error) {
		return gen.ServiceEchoInlineOutput{Data: c.Input.Data}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()

	// Test 1: Original complex types
	testOriginalComplexTypes(client)

	// Test 2: Inline object types
	testInlineTypes(client)

	fmt.Println("Success")
}

func testOriginalComplexTypes(client *gen.Client) {
	input := buildComplexInput()
	res, err := client.RPCs.Service().Procs.Echo().Execute(context.Background(), gen.ServiceEchoInput{Data: input})
	if err != nil {
		panic(err)
	}

	if !reflect.DeepEqual(res.Data, input) {
		panic(fmt.Sprintf("data mismatch:\nsent: %+v\ngot:  %+v", input, res.Data))
	}
}

func testInlineTypes(client *gen.Client) {
	input := buildInlineInput()
	res, err := client.RPCs.Service().Procs.EchoInline().Execute(context.Background(), gen.ServiceEchoInlineInput{Data: input})
	if err != nil {
		panic(fmt.Sprintf("inline types failed: %v", err))
	}

	if !reflect.DeepEqual(res.Data, input) {
		panic(fmt.Sprintf("inline data mismatch:\nsent: %+v\ngot:  %+v", input, res.Data))
	}
}

func buildComplexInput() gen.Container {
	return gen.Container{
		User: gen.User{
			Id:       123,
			Username: "alice",
			IsActive: true,
			Tags:     []string{"admin", "editor", "viewer"},
			Metadata: map[string]string{"role": "superuser", "level": "9000"},
			Address:  gen.Ptr(gen.Address{Street: "123 Main St", City: "Wonderland", Zip: 90210}),
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

func buildInlineInput() gen.InlineContainer {
	return gen.InlineContainer{
		// Simple inline object
		SimpleInline: gen.InlineContainerSimpleInline{
			Name:  "test",
			Value: 42,
		},

		// Array of inline objects
		ArrayOfInline: []gen.InlineContainerArrayOfInline{
			{Id: 1, Label: "first"},
			{Id: 2, Label: "second"},
			{Id: 3, Label: "third"},
		},

		// Map of inline objects
		MapOfInline: map[string]gen.InlineContainerMapOfInline{
			"alpha": {Code: "A", Active: true},
			"beta":  {Code: "B", Active: false},
		},

		// Nested inline objects (inline within inline)
		NestedInline: gen.InlineContainerNestedInline{
			Outer: "outer-value",
			Inner: gen.InlineContainerNestedInlineInner{
				Deep: "deep-value",
				Deeper: gen.InlineContainerNestedInlineInnerDeeper{
					Deepest: "deepest-value",
				},
			},
		},

		// Array of arrays of inline objects (matrix)
		MatrixOfInline: [][]gen.InlineContainerMatrixOfInline{
			{{X: 0, Y: 0}, {X: 1, Y: 0}},
			{{X: 0, Y: 1}, {X: 1, Y: 1}},
		},

		// Map of arrays of inline objects
		MapOfArrayOfInline: map[string][]gen.InlineContainerMapOfArrayOfInline{
			"group1": {{Item: "a", Count: 1}, {Item: "b", Count: 2}},
			"group2": {{Item: "c", Count: 3}},
		},

		// Array of maps of inline objects
		ArrayOfMapOfInline: []map[string]gen.InlineContainerArrayOfMapOfInline{
			{
				"entry1": {Key: "k1", Data: gen.InlineContainerArrayOfMapOfInlineData{Nested: "n1"}},
				"entry2": {Key: "k2", Data: gen.InlineContainerArrayOfMapOfInlineData{Nested: "n2"}},
			},
			{
				"entry3": {Key: "k3", Data: gen.InlineContainerArrayOfMapOfInlineData{Nested: "n3"}},
			},
		},

		// Complex: map of arrays of arrays of inline objects with nested inline
		UltraComplex: []map[string][][]gen.InlineContainerUltraComplex{
			{
				"dimension1": {
					{
						{Level1: gen.InlineContainerUltraComplexLevel1{Level2: "a1"}},
						{Level1: gen.InlineContainerUltraComplexLevel1{Level2: "a2"}},
					},
					{
						{Level1: gen.InlineContainerUltraComplexLevel1{Level2: "b1"}},
					},
				},
			},
			{
				"dimension2": {
					{
						{Level1: gen.InlineContainerUltraComplexLevel1{Level2: "c1"}},
					},
				},
			},
		},
	}
}
