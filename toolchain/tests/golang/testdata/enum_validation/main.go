// Verifies enum validation: IsValid, MarshalJSON, and UnmarshalJSON methods
// reject invalid values while accepting valid ones.
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
	"strings"
)

type AppProps struct{}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Procs.Echo().Handle(func(c *gen.ServiceEchoHandlerContext[AppProps]) (gen.ServiceEchoOutput, error) {
		return gen.ServiceEchoOutput{
			Color:    c.Input.Color,
			Priority: c.Input.Priority,
		}, nil
	})

	server.RPCs.Service().Procs.EchoOptional().Handle(func(c *gen.ServiceEchoOptionalHandlerContext[AppProps]) (gen.ServiceEchoOptionalOutput, error) {
		return gen.ServiceEchoOptionalOutput{
			Container: c.Input.Container,
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

	// Test IsValid method
	testIsValid()

	// Test valid enum round-trip
	testValidEnumRoundTrip(client)

	// Test MarshalJSON rejects invalid values
	testMarshalInvalid()

	// Test UnmarshalJSON rejects invalid values
	testUnmarshalInvalid()

	// Test raw HTTP with invalid enum values (server rejection)
	testServerRejectsInvalidEnum(ts.URL)

	// Test optional enum fields
	testOptionalEnums(client)

	fmt.Println("Success")
}

func testIsValid() {
	// String enum - valid values
	if !gen.ColorRed.IsValid() {
		panic("ColorRed should be valid")
	}
	if !gen.ColorGreen.IsValid() {
		panic("ColorGreen should be valid")
	}
	if !gen.ColorBlue.IsValid() {
		panic("ColorBlue should be valid")
	}

	// String enum - invalid values
	invalid := gen.Color("Purple")
	if invalid.IsValid() {
		panic("Color('Purple') should be invalid")
	}
	empty := gen.Color("")
	if empty.IsValid() {
		panic("Color('') should be invalid")
	}

	// Int enum - valid values
	if !gen.PriorityLow.IsValid() {
		panic("PriorityLow should be valid")
	}
	if !gen.PriorityMedium.IsValid() {
		panic("PriorityMedium should be valid")
	}
	if !gen.PriorityHigh.IsValid() {
		panic("PriorityHigh should be valid")
	}

	// Int enum - invalid values
	invalidPriority := gen.Priority(999)
	if invalidPriority.IsValid() {
		panic("Priority(999) should be invalid")
	}
	zeroPriority := gen.Priority(0)
	if zeroPriority.IsValid() {
		panic("Priority(0) should be invalid")
	}
}

func testValidEnumRoundTrip(client *gen.Client) {
	ctx := context.Background()

	testCases := []struct {
		color    gen.Color
		priority gen.Priority
	}{
		{gen.ColorRed, gen.PriorityLow},
		{gen.ColorGreen, gen.PriorityMedium},
		{gen.ColorBlue, gen.PriorityHigh},
	}

	for _, tc := range testCases {
		res, err := client.RPCs.Service().Procs.Echo().Execute(ctx, gen.ServiceEchoInput{
			Color:    tc.color,
			Priority: tc.priority,
		})
		if err != nil {
			panic(fmt.Sprintf("Echo failed: %v", err))
		}
		if res.Color != tc.color {
			panic(fmt.Sprintf("expected color %s, got %s", tc.color, res.Color))
		}
		if res.Priority != tc.priority {
			panic(fmt.Sprintf("expected priority %d, got %d", tc.priority, res.Priority))
		}
	}
}

func testMarshalInvalid() {
	// String enum - invalid value should fail to marshal
	invalidColor := gen.Color("InvalidColor")
	_, err := json.Marshal(invalidColor)
	if err == nil {
		panic("marshaling invalid Color should fail")
	}
	if !strings.Contains(err.Error(), "InvalidColor") {
		panic(fmt.Sprintf("error should mention the invalid value, got: %v", err))
	}

	// Int enum - invalid value should fail to marshal
	invalidPriority := gen.Priority(999)
	_, err = json.Marshal(invalidPriority)
	if err == nil {
		panic("marshaling invalid Priority should fail")
	}
	if !strings.Contains(err.Error(), "999") {
		panic(fmt.Sprintf("error should mention the invalid value, got: %v", err))
	}

	// Struct with invalid enum should fail to marshal
	type TestStruct struct {
		Color gen.Color `json:"color"`
	}
	_, err = json.Marshal(TestStruct{Color: gen.Color("BadValue")})
	if err == nil {
		panic("marshaling struct with invalid enum should fail")
	}
}

func testUnmarshalInvalid() {
	// String enum - invalid value should fail to unmarshal
	var color gen.Color
	err := json.Unmarshal([]byte(`"InvalidColor"`), &color)
	if err == nil {
		panic("unmarshaling invalid Color should fail")
	}
	if !strings.Contains(err.Error(), "InvalidColor") {
		panic(fmt.Sprintf("error should mention the invalid value, got: %v", err))
	}

	// String enum - empty string should fail
	err = json.Unmarshal([]byte(`""`), &color)
	if err == nil {
		panic("unmarshaling empty string to Color should fail")
	}

	// Int enum - invalid value should fail to unmarshal
	var priority gen.Priority
	err = json.Unmarshal([]byte(`999`), &priority)
	if err == nil {
		panic("unmarshaling invalid Priority should fail")
	}
	if !strings.Contains(err.Error(), "999") {
		panic(fmt.Sprintf("error should mention the invalid value, got: %v", err))
	}

	// Int enum - zero should fail (not a valid member)
	err = json.Unmarshal([]byte(`0`), &priority)
	if err == nil {
		panic("unmarshaling 0 to Priority should fail")
	}

	// Wrong type should fail
	err = json.Unmarshal([]byte(`123`), &color)
	if err == nil {
		panic("unmarshaling int to Color should fail")
	}

	err = json.Unmarshal([]byte(`"string"`), &priority)
	if err == nil {
		panic("unmarshaling string to Priority should fail")
	}
}

func testServerRejectsInvalidEnum(baseURL string) {
	// Invalid string enum value
	payload := `{"color": "Purple", "priority": 1}`
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

	if result["ok"] == true {
		panic("server should reject invalid enum value 'Purple'")
	}

	// Invalid int enum value
	payload = `{"color": "Red", "priority": 999}`
	resp2, err := http.Post(baseURL+"/rpc/Service/Echo", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		panic(fmt.Sprintf("request failed: %v", err))
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)
	var result2 map[string]any
	if err := json.Unmarshal(body2, &result2); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %v", err))
	}

	if result2["ok"] == true {
		panic("server should reject invalid enum value 999")
	}
}

func testOptionalEnums(client *gen.Client) {
	ctx := context.Background()

	// Test with absent optional enums
	res, err := client.RPCs.Service().Procs.EchoOptional().Execute(ctx, gen.ServiceEchoOptionalInput{
		Container: gen.Container{},
	})
	if err != nil {
		panic(fmt.Sprintf("EchoOptional failed: %v", err))
	}
	if res.Container.Color.Present {
		panic("color should be absent")
	}
	if res.Container.Priority.Present {
		panic("priority should be absent")
	}

	// Test with present valid optional enums
	res2, err := client.RPCs.Service().Procs.EchoOptional().Execute(ctx, gen.ServiceEchoOptionalInput{
		Container: gen.Container{
			Color:    gen.Some(gen.ColorBlue),
			Priority: gen.Some(gen.PriorityHigh),
		},
	})
	if err != nil {
		panic(fmt.Sprintf("EchoOptional failed: %v", err))
	}
	if !res2.Container.Color.Present {
		panic("color should be present")
	}
	if res2.Container.Color.Value != gen.ColorBlue {
		panic(fmt.Sprintf("expected ColorBlue, got %s", res2.Container.Color.Value))
	}
	if !res2.Container.Priority.Present {
		panic("priority should be present")
	}
	if res2.Container.Priority.Value != gen.PriorityHigh {
		panic(fmt.Sprintf("expected PriorityHigh, got %d", res2.Container.Priority.Value))
	}
}
