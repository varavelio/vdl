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
			WasOptionalPresent:     c.Input.Optional != nil,
			WasOptionalIntPresent:  c.Input.OptionalInt != nil,
			WasOptionalBoolPresent: c.Input.OptionalBool != nil,
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
			WasPresent: c.Input.Wrapper != nil,
		}, nil
	})

	server.RPCs.Service().Procs.EchoInline().Handle(func(c *gen.ServiceEchoInlineHandlerContext[AppProps]) (gen.ServiceEchoInlineOutput, error) {
		return gen.ServiceEchoInlineOutput{
			Data:                     c.Input.Data,
			InlineObjectPresent:      c.Input.Data.InlineObject != nil,
			InlineArrayPresent:       c.Input.Data.InlineArray != nil,
			InlineMapPresent:         c.Input.Data.InlineMap != nil,
			NestedInlinePresent:      c.Input.Data.NestedInline != nil,
			InlineMatrixPresent:      c.Input.Data.InlineMatrix != nil,
			MapOfInlineArraysPresent: c.Input.Data.MapOfInlineArrays != nil,
			ArrayOfInlineMapsPresent: c.Input.Data.ArrayOfInlineMaps != nil,
			UltraComplexPresent:      c.Input.Data.UltraComplex != nil,
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

	// Inline object tests
	testInlineAllAbsent(client)
	testInlineAllPresent(client)
	testInlineNestedAndMatrix(client)
	testInlineUltraComplex(client)

	fmt.Println("Success")
}

// detectPresentFields returns a list of field names that are present in ComplexOptional
func detectPresentFields(data gen.ComplexOptional) []string {
	var present []string
	if data.Name != nil {
		present = append(present, "name")
	}
	if data.Tags != nil {
		present = append(present, "tags")
	}
	if data.Metadata != nil {
		present = append(present, "metadata")
	}
	if data.Address != nil {
		present = append(present, "address")
	}
	if data.Coordinates != nil {
		present = append(present, "coordinates")
	}
	if data.Locations != nil {
		present = append(present, "locations")
	}
	if data.Wrapper != nil {
		present = append(present, "wrapper")
	}
	if data.Matrix != nil {
		present = append(present, "matrix")
	}
	if data.MapOfArrays != nil {
		present = append(present, "mapOfArrays")
	}
	if data.ArrayOfMaps != nil {
		present = append(present, "arrayOfMaps")
	}
	if data.NestedObjectMatrix != nil {
		present = append(present, "nestedObjectMatrix")
	}
	if data.MapOfMaps != nil {
		present = append(present, "mapOfMaps")
	}
	if data.MapOfObjectArrays != nil {
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
		Optional: gen.Ptr("value"),
	})
	if err != nil {
		panic(fmt.Sprintf("execute failed: %v", err))
	}

	if !result2.WasOptionalPresent {
		panic("expected WasOptionalPresent=true")
	}
	if *result2.Optional != "value" {
		panic(fmt.Sprintf("expected Optional='value', got: %s", *result2.Optional))
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
		Name:     gen.Ptr("Alice"),
		Tags:     gen.Ptr([]string{"tag1", "tag2"}),
		Metadata: gen.Ptr(map[string]string{"key": "value"}),
		Address: gen.Ptr(gen.Address{
			Street: "123 Main St",
			City:   "NYC",
		}),
		Coordinates: gen.Ptr([]gen.Coordinates{
			{Lat: 40.7128, Lng: -74.0060},
			{Lat: 34.0522, Lng: -118.2437},
		}),
		Locations: gen.Ptr(map[string]gen.Address{
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
		Matrix: gen.Ptr([][]int64{
			{1, 2, 3},
			{4, 5, 6},
		}),
		NestedObjectMatrix: gen.Ptr([][]gen.Level3Data{
			{{Value: "a", Score: 1}, {Value: "b", Score: 2}},
			{{Value: "c", Score: 3}},
		}),
		ArrayOfMaps: gen.Ptr([]map[string]string{
			{"k1": "v1", "k2": "v2"},
			{"k3": "v3"},
		}),
	}

	res, err := client.RPCs.Service().Procs.EchoComplex().Execute(ctx, gen.ServiceEchoComplexInput{Data: input})
	if err != nil {
		panic(fmt.Sprintf("EchoComplex failed: %v", err))
	}

	if res.Data.Matrix == nil {
		panic("expected matrix to be present")
	}
	if res.Data.NestedObjectMatrix == nil {
		panic("expected nestedObjectMatrix to be present")
	}
	if res.Data.ArrayOfMaps == nil {
		panic("expected arrayOfMaps to be present")
	}

	if !reflect.DeepEqual(*res.Data.Matrix, *input.Matrix) {
		panic("matrix mismatch")
	}
	if !reflect.DeepEqual(*res.Data.NestedObjectMatrix, *input.NestedObjectMatrix) {
		panic("nestedObjectMatrix mismatch")
	}
}

func testComplexNestedMaps(client *gen.Client) {
	ctx := context.Background()

	input := gen.ComplexOptional{
		Id: "test-maps",
		MapOfArrays: gen.Ptr(map[string][]int64{
			"evens": {2, 4, 6, 8},
			"odds":  {1, 3, 5, 7},
		}),
		MapOfMaps: gen.Ptr(map[string]map[string]string{
			"outer1": {"inner1": "val1", "inner2": "val2"},
			"outer2": {"inner3": "val3"},
		}),
		MapOfObjectArrays: gen.Ptr(map[string][]gen.Level3Data{
			"group1": {{Value: "x", Score: 10}, {Value: "y", Score: 20}},
			"group2": {{Value: "z", Score: 30}},
		}),
	}

	res, err := client.RPCs.Service().Procs.EchoComplex().Execute(ctx, gen.ServiceEchoComplexInput{Data: input})
	if err != nil {
		panic(fmt.Sprintf("EchoComplex failed: %v", err))
	}

	if res.Data.MapOfArrays == nil {
		panic("expected mapOfArrays to be present")
	}
	if res.Data.MapOfMaps == nil {
		panic("expected mapOfMaps to be present")
	}
	if res.Data.MapOfObjectArrays == nil {
		panic("expected mapOfObjectArrays to be present")
	}

	if !reflect.DeepEqual(*res.Data.MapOfArrays, *input.MapOfArrays) {
		panic("mapOfArrays mismatch")
	}
	if !reflect.DeepEqual(*res.Data.MapOfMaps, *input.MapOfMaps) {
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
			Items: gen.Ptr([]gen.Level3Data{
				{Value: "item1", Score: 1},
				{Value: "item2", Score: 2},
			}),
		},
		Metadata: gen.Ptr(map[string]string{
			"version": "1.0",
			"author":  "test",
		}),
	}

	res, err := client.RPCs.Service().Procs.EchoDeep().Execute(ctx, gen.ServiceEchoDeepInput{
		Wrapper: gen.Ptr(wrapper),
	})
	if err != nil {
		panic(fmt.Sprintf("EchoDeep failed: %v", err))
	}

	if !res.WasPresent {
		panic("expected wrapper to be present")
	}
	if !reflect.DeepEqual(*res.Wrapper, wrapper) {
		panic(fmt.Sprintf("wrapper mismatch:\nexpected: %+v\ngot:      %+v", wrapper, *res.Wrapper))
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
		Wrapper: gen.Ptr(gen.Level1Wrapper{
			Id: "nested-wrapper",
			Container: gen.Level2Container{
				Name: "deep-container",
				Data: gen.Level3Data{Value: "innermost", Score: 999},
			},
		}),
		Matrix: gen.Ptr([][]int64{
			{1, 2},
			{3, 4},
			{5, 6},
		}),
		NestedObjectMatrix: gen.Ptr([][]gen.Level3Data{
			{{Value: "r0c0", Score: 0}, {Value: "r0c1", Score: 1}},
			{{Value: "r1c0", Score: 2}, {Value: "r1c1", Score: 3}},
		}),
	}

	res, err := client.RPCs.Service().Procs.EchoComplex().Execute(ctx, gen.ServiceEchoComplexInput{Data: input})
	if err != nil {
		panic(fmt.Sprintf("EchoComplex failed: %v", err))
	}

	if res.Data.Wrapper == nil {
		panic("expected wrapper to be present")
	}
	if res.Data.Wrapper.Container.Data.Score != 999 {
		panic("deeply nested score mismatch")
	}

	if !reflect.DeepEqual(res.Data, input) {
		panic("matrix and nested structures mismatch")
	}
}

// ===== Inline object tests =====

func testInlineAllAbsent(client *gen.Client) {
	ctx := context.Background()

	input := gen.ComplexOptional{Id: "inline-absent"}
	res, err := client.RPCs.Service().Procs.EchoInline().Execute(ctx, gen.ServiceEchoInlineInput{Data: input})
	if err != nil {
		panic(fmt.Sprintf("EchoInline failed: %v", err))
	}

	// All inline fields should be absent
	if res.InlineObjectPresent {
		panic("expected InlineObject to be absent")
	}
	if res.InlineArrayPresent {
		panic("expected InlineArray to be absent")
	}
	if res.InlineMapPresent {
		panic("expected InlineMap to be absent")
	}
	if res.NestedInlinePresent {
		panic("expected NestedInline to be absent")
	}
	if res.InlineMatrixPresent {
		panic("expected InlineMatrix to be absent")
	}
	if res.MapOfInlineArraysPresent {
		panic("expected MapOfInlineArrays to be absent")
	}
	if res.ArrayOfInlineMapsPresent {
		panic("expected ArrayOfInlineMaps to be absent")
	}
	if res.UltraComplexPresent {
		panic("expected UltraComplex to be absent")
	}
}

func testInlineAllPresent(client *gen.Client) {
	ctx := context.Background()

	input := gen.ComplexOptional{
		Id: "inline-present",
		InlineObject: gen.Ptr(gen.ComplexOptionalInlineObject{
			Label: "test-label",
			Count: 42,
		}),
		InlineArray: gen.Ptr([]gen.ComplexOptionalInlineArray{
			{Name: "item1", Active: true},
			{Name: "item2", Active: false},
		}),
		InlineMap: gen.Ptr(map[string]gen.ComplexOptionalInlineMap{
			"key1": {Key: "k1", Value: 100},
			"key2": {Key: "k2", Value: 200},
		}),
		NestedInline: gen.Ptr(gen.ComplexOptionalNestedInline{
			Outer: "outer-value",
			Inner: gen.ComplexOptionalNestedInlineInner{
				Deep:  "deep-value",
				Score: 3.14,
			},
		}),
	}

	res, err := client.RPCs.Service().Procs.EchoInline().Execute(ctx, gen.ServiceEchoInlineInput{Data: input})
	if err != nil {
		panic(fmt.Sprintf("EchoInline failed: %v", err))
	}

	// All specified inline fields should be present
	if !res.InlineObjectPresent {
		panic("expected InlineObject to be present")
	}
	if !res.InlineArrayPresent {
		panic("expected InlineArray to be present")
	}
	if !res.InlineMapPresent {
		panic("expected InlineMap to be present")
	}
	if !res.NestedInlinePresent {
		panic("expected NestedInline to be present")
	}

	// Verify data integrity
	if res.Data.InlineObject.Label != "test-label" {
		panic("InlineObject.Label mismatch")
	}
	if res.Data.InlineObject.Count != 42 {
		panic("InlineObject.Count mismatch")
	}
	if len(*res.Data.InlineArray) != 2 {
		panic("InlineArray length mismatch")
	}
	if res.Data.NestedInline.Inner.Deep != "deep-value" {
		panic("NestedInline.Inner.Deep mismatch")
	}
}

func testInlineNestedAndMatrix(client *gen.Client) {
	ctx := context.Background()

	input := gen.ComplexOptional{
		Id: "inline-matrix",
		InlineMatrix: gen.Ptr([][]gen.ComplexOptionalInlineMatrix{
			{{X: 0, Y: 0}, {X: 0, Y: 1}},
			{{X: 1, Y: 0}, {X: 1, Y: 1}},
			{{X: 2, Y: 0}},
		}),
		MapOfInlineArrays: gen.Ptr(map[string][]gen.ComplexOptionalMapOfInlineArrays{
			"high": {{Item: "urgent", Priority: 1}, {Item: "critical", Priority: 0}},
			"low":  {{Item: "later", Priority: 10}},
		}),
		ArrayOfInlineMaps: gen.Ptr([]map[string]gen.ComplexOptionalArrayOfInlineMaps{
			{"first": {Data: "data1"}, "second": {Data: "data2"}},
			{"third": {Data: "data3"}},
		}),
	}

	res, err := client.RPCs.Service().Procs.EchoInline().Execute(ctx, gen.ServiceEchoInlineInput{Data: input})
	if err != nil {
		panic(fmt.Sprintf("EchoInline failed: %v", err))
	}

	if !res.InlineMatrixPresent {
		panic("expected InlineMatrix to be present")
	}
	if !res.MapOfInlineArraysPresent {
		panic("expected MapOfInlineArrays to be present")
	}
	if !res.ArrayOfInlineMapsPresent {
		panic("expected ArrayOfInlineMaps to be present")
	}

	// Verify matrix dimensions
	if len(*res.Data.InlineMatrix) != 3 {
		panic("InlineMatrix row count mismatch")
	}
	if len((*res.Data.InlineMatrix)[0]) != 2 {
		panic("InlineMatrix first row column count mismatch")
	}
	if (*res.Data.InlineMatrix)[1][1].X != 1 || (*res.Data.InlineMatrix)[1][1].Y != 1 {
		panic("InlineMatrix[1][1] values mismatch")
	}

	// Verify map of inline arrays
	highPriority := (*res.Data.MapOfInlineArrays)["high"]
	if len(highPriority) != 2 {
		panic("MapOfInlineArrays['high'] length mismatch")
	}
	if highPriority[0].Priority != 1 {
		panic("MapOfInlineArrays['high'][0].Priority mismatch")
	}

	// Verify array of inline maps
	if len(*res.Data.ArrayOfInlineMaps) != 2 {
		panic("ArrayOfInlineMaps length mismatch")
	}
	if (*res.Data.ArrayOfInlineMaps)[0]["first"].Data != "data1" {
		panic("ArrayOfInlineMaps[0]['first'].Data mismatch")
	}
}

func testInlineUltraComplex(client *gen.Client) {
	ctx := context.Background()

	input := gen.ComplexOptional{
		Id: "inline-ultra",
		UltraComplex: gen.Ptr(map[string][]gen.ComplexOptionalUltraComplex{
			"group1": {
				{
					Level1: "g1-item1",
					Nested: gen.ComplexOptionalUltraComplexNested{
						Level2: 100,
						Items:  []string{"a", "b", "c"},
					},
				},
				{
					Level1: "g1-item2",
					Nested: gen.ComplexOptionalUltraComplexNested{
						Level2: 200,
						Items:  []string{"d", "e"},
					},
				},
			},
			"group2": {
				{
					Level1: "g2-item1",
					Nested: gen.ComplexOptionalUltraComplexNested{
						Level2: 300,
						Items:  []string{"x", "y", "z"},
					},
				},
			},
		}),
	}

	res, err := client.RPCs.Service().Procs.EchoInline().Execute(ctx, gen.ServiceEchoInlineInput{Data: input})
	if err != nil {
		panic(fmt.Sprintf("EchoInline failed: %v", err))
	}

	if !res.UltraComplexPresent {
		panic("expected UltraComplex to be present")
	}

	// Verify the ultra complex structure
	group1 := (*res.Data.UltraComplex)["group1"]
	if len(group1) != 2 {
		panic("UltraComplex['group1'] length mismatch")
	}
	if group1[0].Level1 != "g1-item1" {
		panic("UltraComplex['group1'][0].Level1 mismatch")
	}
	if group1[0].Nested.Level2 != 100 {
		panic("UltraComplex['group1'][0].Nested.Level2 mismatch")
	}
	if len(group1[0].Nested.Items) != 3 {
		panic("UltraComplex['group1'][0].Nested.Items length mismatch")
	}

	group2 := (*res.Data.UltraComplex)["group2"]
	if len(group2) != 1 {
		panic("UltraComplex['group2'] length mismatch")
	}
	if !reflect.DeepEqual(group2[0].Nested.Items, []string{"x", "y", "z"}) {
		panic("UltraComplex['group2'][0].Nested.Items mismatch")
	}

	// Verify round-trip
	if !reflect.DeepEqual(*res.Data.UltraComplex, *input.UltraComplex) {
		panic("UltraComplex round-trip mismatch")
	}
}
