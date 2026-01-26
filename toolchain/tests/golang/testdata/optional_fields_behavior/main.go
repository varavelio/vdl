// Verifies optional field behavior: server can detect whether optional fields
// were explicitly provided vs absent, and empty values are distinct from absent.
// Also tests deeply nested optional structures with arrays, maps, and complex types.
package main

import (
	"bytes"
	"context"
	"e2e/gen"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
)

type AppProps struct{}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Procs.Echo().Handle(func(c *gen.ServiceEchoHandlerContext[AppProps]) (gen.ServiceEchoOutput, error) {
		return gen.ServiceEchoOutput{
			Required:               c.Input.Required,
			Optional:               c.Input.Optional,
			OptionalInt:            c.Input.OptionalInt,
			OptionalBool:           c.Input.OptionalBool,
			WasOptionalPresent:     c.Input.Optional.Present,
			WasOptionalIntPresent:  c.Input.OptionalInt.Present,
			WasOptionalBoolPresent: c.Input.OptionalBool.Present,
		}, nil
	})

	server.RPCs.Service().Procs.EchoComplex().Handle(func(c *gen.ServiceEchoComplexHandlerContext[AppProps]) (gen.ServiceEchoComplexOutput, error) {
		presentFields := detectPresentFields(c.Input.Data)
		return gen.ServiceEchoComplexOutput{
			Data:          c.Input.Data,
			PresentFields: presentFields,
		}, nil
	})

	server.RPCs.Service().Procs.EchoDeep().Handle(func(c *gen.ServiceEchoDeepHandlerContext[AppProps]) (gen.ServiceEchoDeepOutput, error) {
		return gen.ServiceEchoDeepOutput{
			Wrapper:    c.Input.Wrapper,
			WasPresent: c.Input.Wrapper.Present,
		}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Original tests
	testAllAbsent(ts.URL)
	testAllPresent(ts.URL)
	testMixedPresence(ts.URL)
	testGeneratedClient(ts.URL)

	// New complex tests
	client := gen.NewClient(ts.URL + "/rpc").Build()
	testComplexAllAbsent(client)
	testComplexAllPresent(client)
	testComplexNestedArrays(client)
	testComplexNestedMaps(client)
	testDeepNesting(client)
	testMatrixAndNestedStructures(client)

	fmt.Println("Success")
}

// detectPresentFields returns a list of field names that are present in ComplexOptional
func detectPresentFields(data gen.ComplexOptional) []string {
	var present []string
	if data.Name.Present {
		present = append(present, "name")
	}
	if data.Tags.Present {
		present = append(present, "tags")
	}
	if data.Metadata.Present {
		present = append(present, "metadata")
	}
	if data.Address.Present {
		present = append(present, "address")
	}
	if data.Coordinates.Present {
		present = append(present, "coordinates")
	}
	if data.Locations.Present {
		present = append(present, "locations")
	}
	if data.Wrapper.Present {
		present = append(present, "wrapper")
	}
	if data.Matrix.Present {
		present = append(present, "matrix")
	}
	if data.MapOfArrays.Present {
		present = append(present, "mapOfArrays")
	}
	if data.ArrayOfMaps.Present {
		present = append(present, "arrayOfMaps")
	}
	if data.NestedObjectMatrix.Present {
		present = append(present, "nestedObjectMatrix")
	}
	if data.MapOfMaps.Present {
		present = append(present, "mapOfMaps")
	}
	if data.MapOfObjectArrays.Present {
		present = append(present, "mapOfObjectArrays")
	}
	return present
}

func testAllAbsent(baseURL string) {
	payload := `{"required": "hello"}`
	resp, err := http.Post(baseURL+"/rpc/Service/Echo", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		panic(fmt.Sprintf("request failed: %v", err))
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %v", err))
	}

	if result["ok"] != true {
		panic(fmt.Sprintf("expected ok=true, got: %s", string(body)))
	}

	output := result["output"].(map[string]any)

	if output["wasOptionalPresent"] != false {
		panic("expected wasOptionalPresent=false when optional is absent")
	}
	if output["wasOptionalIntPresent"] != false {
		panic("expected wasOptionalIntPresent=false when optionalInt is absent")
	}
	if output["wasOptionalBoolPresent"] != false {
		panic("expected wasOptionalBoolPresent=false when optionalBool is absent")
	}
}

func testAllPresent(baseURL string) {
	payload := `{"required": "hello", "optional": "", "optionalInt": 0, "optionalBool": false}`
	resp, err := http.Post(baseURL+"/rpc/Service/Echo", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		panic(fmt.Sprintf("request failed: %v", err))
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %v", err))
	}

	if result["ok"] != true {
		panic(fmt.Sprintf("expected ok=true, got: %s", string(body)))
	}

	output := result["output"].(map[string]any)

	if output["wasOptionalPresent"] != true {
		panic("expected wasOptionalPresent=true when optional is explicitly empty string")
	}
	if output["wasOptionalIntPresent"] != true {
		panic("expected wasOptionalIntPresent=true when optionalInt is explicitly 0")
	}
	if output["wasOptionalBoolPresent"] != true {
		panic("expected wasOptionalBoolPresent=true when optionalBool is explicitly false")
	}
}

func testMixedPresence(baseURL string) {
	payload := `{"required": "hello", "optional": "world"}`
	resp, err := http.Post(baseURL+"/rpc/Service/Echo", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		panic(fmt.Sprintf("request failed: %v", err))
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %v", err))
	}

	if result["ok"] != true {
		panic(fmt.Sprintf("expected ok=true, got: %s", string(body)))
	}

	output := result["output"].(map[string]any)

	if output["wasOptionalPresent"] != true {
		panic("expected wasOptionalPresent=true")
	}
	if output["wasOptionalIntPresent"] != false {
		panic("expected wasOptionalIntPresent=false")
	}
	if output["wasOptionalBoolPresent"] != false {
		panic("expected wasOptionalBoolPresent=false")
	}

	if output["optional"] != "world" {
		panic(fmt.Sprintf("expected optional='world', got: %v", output["optional"]))
	}
}

func testGeneratedClient(baseURL string) {
	client := gen.NewClient(baseURL + "/rpc").Build()
	ctx := context.Background()

	result, err := client.RPCs.Service().Procs.Echo().Execute(ctx, gen.ServiceEchoInput{
		Required: "test",
	})
	if err != nil {
		panic(fmt.Sprintf("execute failed: %v", err))
	}

	if result.WasOptionalPresent {
		panic("expected WasOptionalPresent=false")
	}

	result2, err := client.RPCs.Service().Procs.Echo().Execute(ctx, gen.ServiceEchoInput{
		Required: "test",
		Optional: gen.Some("value"),
	})
	if err != nil {
		panic(fmt.Sprintf("execute failed: %v", err))
	}

	if !result2.WasOptionalPresent {
		panic("expected WasOptionalPresent=true")
	}
	if result2.Optional.Value != "value" {
		panic(fmt.Sprintf("expected Optional.Value='value', got: %s", result2.Optional.Value))
	}
}

func testComplexAllAbsent(client *gen.Client) {
	ctx := context.Background()

	input := gen.ComplexOptional{Id: "test-1"}
	res, err := client.RPCs.Service().Procs.EchoComplex().Execute(ctx, gen.ServiceEchoComplexInput{Data: input})
	if err != nil {
		panic(fmt.Sprintf("EchoComplex failed: %v", err))
	}

	if len(res.PresentFields) != 0 {
		panic(fmt.Sprintf("expected no present fields, got: %v", res.PresentFields))
	}

	if res.Data.Id != "test-1" {
		panic(fmt.Sprintf("expected id='test-1', got: %s", res.Data.Id))
	}
}

func testComplexAllPresent(client *gen.Client) {
	ctx := context.Background()

	input := gen.ComplexOptional{
		Id:       "test-2",
		Name:     gen.Some("Alice"),
		Tags:     gen.Some([]string{"tag1", "tag2"}),
		Metadata: gen.Some(map[string]string{"key": "value"}),
		Address: gen.Some(gen.Address{
			Street: "123 Main St",
			City:   "NYC",
		}),
		Coordinates: gen.Some([]gen.Coordinates{
			{Lat: 40.7128, Lng: -74.0060},
			{Lat: 34.0522, Lng: -118.2437},
		}),
		Locations: gen.Some(map[string]gen.Address{
			"home": {Street: "Home St", City: "HomeTown"},
			"work": {Street: "Work Ave", City: "WorkCity"},
		}),
	}

	res, err := client.RPCs.Service().Procs.EchoComplex().Execute(ctx, gen.ServiceEchoComplexInput{Data: input})
	if err != nil {
		panic(fmt.Sprintf("EchoComplex failed: %v", err))
	}

	expectedFields := []string{"name", "tags", "metadata", "address", "coordinates", "locations"}
	if len(res.PresentFields) != len(expectedFields) {
		panic(fmt.Sprintf("expected %d present fields, got: %v", len(expectedFields), res.PresentFields))
	}

	if !reflect.DeepEqual(res.Data, input) {
		panic(fmt.Sprintf("data mismatch:\nsent: %+v\ngot:  %+v", input, res.Data))
	}
}

func testComplexNestedArrays(client *gen.Client) {
	ctx := context.Background()

	input := gen.ComplexOptional{
		Id: "test-arrays",
		Matrix: gen.Some([][]int64{
			{1, 2, 3},
			{4, 5, 6},
		}),
		NestedObjectMatrix: gen.Some([][]gen.Level3Data{
			{{Value: "a", Score: 1}, {Value: "b", Score: 2}},
			{{Value: "c", Score: 3}},
		}),
		ArrayOfMaps: gen.Some([]map[string]string{
			{"k1": "v1", "k2": "v2"},
			{"k3": "v3"},
		}),
	}

	res, err := client.RPCs.Service().Procs.EchoComplex().Execute(ctx, gen.ServiceEchoComplexInput{Data: input})
	if err != nil {
		panic(fmt.Sprintf("EchoComplex failed: %v", err))
	}

	if !res.Data.Matrix.Present {
		panic("expected matrix to be present")
	}
	if !res.Data.NestedObjectMatrix.Present {
		panic("expected nestedObjectMatrix to be present")
	}
	if !res.Data.ArrayOfMaps.Present {
		panic("expected arrayOfMaps to be present")
	}

	if !reflect.DeepEqual(res.Data.Matrix.Value, input.Matrix.Value) {
		panic("matrix mismatch")
	}
	if !reflect.DeepEqual(res.Data.NestedObjectMatrix.Value, input.NestedObjectMatrix.Value) {
		panic("nestedObjectMatrix mismatch")
	}
}

func testComplexNestedMaps(client *gen.Client) {
	ctx := context.Background()

	input := gen.ComplexOptional{
		Id: "test-maps",
		MapOfArrays: gen.Some(map[string][]int64{
			"evens": {2, 4, 6, 8},
			"odds":  {1, 3, 5, 7},
		}),
		MapOfMaps: gen.Some(map[string]map[string]string{
			"outer1": {"inner1": "val1", "inner2": "val2"},
			"outer2": {"inner3": "val3"},
		}),
		MapOfObjectArrays: gen.Some(map[string][]gen.Level3Data{
			"group1": {{Value: "x", Score: 10}, {Value: "y", Score: 20}},
			"group2": {{Value: "z", Score: 30}},
		}),
	}

	res, err := client.RPCs.Service().Procs.EchoComplex().Execute(ctx, gen.ServiceEchoComplexInput{Data: input})
	if err != nil {
		panic(fmt.Sprintf("EchoComplex failed: %v", err))
	}

	if !res.Data.MapOfArrays.Present {
		panic("expected mapOfArrays to be present")
	}
	if !res.Data.MapOfMaps.Present {
		panic("expected mapOfMaps to be present")
	}
	if !res.Data.MapOfObjectArrays.Present {
		panic("expected mapOfObjectArrays to be present")
	}

	if !reflect.DeepEqual(res.Data.MapOfArrays.Value, input.MapOfArrays.Value) {
		panic("mapOfArrays mismatch")
	}
	if !reflect.DeepEqual(res.Data.MapOfMaps.Value, input.MapOfMaps.Value) {
		panic("mapOfMaps mismatch")
	}
}

func testDeepNesting(client *gen.Client) {
	ctx := context.Background()

	// Test with deeply nested structure present
	wrapper := gen.Level1Wrapper{
		Id: "wrapper-1",
		Container: gen.Level2Container{
			Name: "container-1",
			Data: gen.Level3Data{
				Value: "deep-value",
				Score: 100,
			},
			Items: gen.Some([]gen.Level3Data{
				{Value: "item1", Score: 1},
				{Value: "item2", Score: 2},
			}),
		},
		Metadata: gen.Some(map[string]string{
			"version": "1.0",
			"author":  "test",
		}),
	}

	res, err := client.RPCs.Service().Procs.EchoDeep().Execute(ctx, gen.ServiceEchoDeepInput{
		Wrapper: gen.Some(wrapper),
	})
	if err != nil {
		panic(fmt.Sprintf("EchoDeep failed: %v", err))
	}

	if !res.WasPresent {
		panic("expected wrapper to be present")
	}
	if !reflect.DeepEqual(res.Wrapper.Value, wrapper) {
		panic(fmt.Sprintf("wrapper mismatch:\nexpected: %+v\ngot:      %+v", wrapper, res.Wrapper.Value))
	}

	// Test with wrapper absent
	res2, err := client.RPCs.Service().Procs.EchoDeep().Execute(ctx, gen.ServiceEchoDeepInput{})
	if err != nil {
		panic(fmt.Sprintf("EchoDeep (absent) failed: %v", err))
	}

	if res2.WasPresent {
		panic("expected wrapper to be absent")
	}
}

func testMatrixAndNestedStructures(client *gen.Client) {
	ctx := context.Background()

	input := gen.ComplexOptional{
		Id: "test-matrix",
		Wrapper: gen.Some(gen.Level1Wrapper{
			Id: "nested-wrapper",
			Container: gen.Level2Container{
				Name: "deep-container",
				Data: gen.Level3Data{Value: "innermost", Score: 999},
			},
		}),
		Matrix: gen.Some([][]int64{
			{1, 2},
			{3, 4},
			{5, 6},
		}),
		NestedObjectMatrix: gen.Some([][]gen.Level3Data{
			{{Value: "r0c0", Score: 0}, {Value: "r0c1", Score: 1}},
			{{Value: "r1c0", Score: 2}, {Value: "r1c1", Score: 3}},
		}),
	}

	res, err := client.RPCs.Service().Procs.EchoComplex().Execute(ctx, gen.ServiceEchoComplexInput{Data: input})
	if err != nil {
		panic(fmt.Sprintf("EchoComplex failed: %v", err))
	}

	if !res.Data.Wrapper.Present {
		panic("expected wrapper to be present")
	}
	if res.Data.Wrapper.Value.Container.Data.Score != 999 {
		panic("deeply nested score mismatch")
	}

	if !reflect.DeepEqual(res.Data, input) {
		panic("matrix and nested structures mismatch")
	}
}
