// Verifies server-side input validation for deeply nested structures, arrays, and maps.
// Missing required fields at any nesting level should produce validation errors.
// Also tests that optional nested objects, when present, must have all required inner fields.
package main

import (
	"bytes"
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

	// Original handlers
	server.RPCs.Service().Procs.ValidatePerson().Handle(func(c *gen.ServiceValidatePersonHandlerContext[AppProps]) (gen.ServiceValidatePersonOutput, error) {
		return gen.ServiceValidatePersonOutput{}, nil
	})

	server.RPCs.Service().Procs.ValidateTeam().Handle(func(c *gen.ServiceValidateTeamHandlerContext[AppProps]) (gen.ServiceValidateTeamOutput, error) {
		return gen.ServiceValidateTeamOutput{}, nil
	})

	server.RPCs.Service().Procs.ValidateOrganization().Handle(func(c *gen.ServiceValidateOrganizationHandlerContext[AppProps]) (gen.ServiceValidateOrganizationOutput, error) {
		return gen.ServiceValidateOrganizationOutput{}, nil
	})

	server.RPCs.Service().Procs.ValidateArray().Handle(func(c *gen.ServiceValidateArrayHandlerContext[AppProps]) (gen.ServiceValidateArrayOutput, error) {
		return gen.ServiceValidateArrayOutput{}, nil
	})

	server.RPCs.Service().Procs.ValidateMap().Handle(func(c *gen.ServiceValidateMapHandlerContext[AppProps]) (gen.ServiceValidateMapOutput, error) {
		return gen.ServiceValidateMapOutput{}, nil
	})

	// New handlers for optional nested validation
	server.RPCs.Service().Procs.ValidateOptionalNested().Handle(func(c *gen.ServiceValidateOptionalNestedHandlerContext[AppProps]) (gen.ServiceValidateOptionalNestedOutput, error) {
		return gen.ServiceValidateOptionalNestedOutput{}, nil
	})

	server.RPCs.Service().Procs.ValidateDeepOptionalNested().Handle(func(c *gen.ServiceValidateDeepOptionalNestedHandlerContext[AppProps]) (gen.ServiceValidateDeepOptionalNestedOutput, error) {
		return gen.ServiceValidateDeepOptionalNestedOutput{}, nil
	})

	server.RPCs.Service().Procs.ValidateOptionalArrayNested().Handle(func(c *gen.ServiceValidateOptionalArrayNestedHandlerContext[AppProps]) (gen.ServiceValidateOptionalArrayNestedOutput, error) {
		return gen.ServiceValidateOptionalArrayNestedOutput{}, nil
	})

	server.RPCs.Service().Procs.ValidateOptionalMapNested().Handle(func(c *gen.ServiceValidateOptionalMapNestedHandlerContext[AppProps]) (gen.ServiceValidateOptionalMapNestedOutput, error) {
		return gen.ServiceValidateOptionalMapNestedOutput{}, nil
	})

	server.RPCs.Service().Procs.ValidateComplexOptional().Handle(func(c *gen.ServiceValidateComplexOptionalHandlerContext[AppProps]) (gen.ServiceValidateComplexOptionalOutput, error) {
		return gen.ServiceValidateComplexOptionalOutput{}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Original test cases
	runOriginalTestCases(ts.URL)

	// New test cases for optional nested with required inner fields
	runOptionalNestedTestCases(ts.URL)

	fmt.Println("Success")
}

func runOriginalTestCases(baseURL string) {
	testCases := []struct {
		name        string
		endpoint    string
		payload     string
		shouldFail  bool
		errContains string
	}{
		// Nested object - missing required field at level 2
		{
			name:       "Person valid",
			endpoint:   "/rpc/Service/ValidatePerson",
			payload:    `{"person": {"name": "John", "address": {"street": "123 Main", "city": "NYC"}}}`,
			shouldFail: false,
		},
		{
			name:        "Person missing address.city",
			endpoint:    "/rpc/Service/ValidatePerson",
			payload:     `{"person": {"name": "John", "address": {"street": "123 Main"}}}`,
			shouldFail:  true,
			errContains: "city",
		},
		{
			name:        "Person missing address entirely",
			endpoint:    "/rpc/Service/ValidatePerson",
			payload:     `{"person": {"name": "John"}}`,
			shouldFail:  true,
			errContains: "address",
		},

		// Deeply nested - Team has Person which has Address
		{
			name:       "Team valid",
			endpoint:   "/rpc/Service/ValidateTeam",
			payload:    `{"team": {"leader": {"name": "Boss", "address": {"street": "HQ", "city": "LA"}}, "members": []}}`,
			shouldFail: false,
		},
		{
			name:        "Team leader missing address",
			endpoint:    "/rpc/Service/ValidateTeam",
			payload:     `{"team": {"leader": {"name": "Boss"}, "members": []}}`,
			shouldFail:  true,
			errContains: "address",
		},

		// Array of nested objects
		{
			name:       "Array valid",
			endpoint:   "/rpc/Service/ValidateArray",
			payload:    `{"people": [{"name": "A", "address": {"street": "1", "city": "X"}}, {"name": "B", "address": {"street": "2", "city": "Y"}}]}`,
			shouldFail: false,
		},
		{
			name:        "Array element missing nested field",
			endpoint:    "/rpc/Service/ValidateArray",
			payload:     `{"people": [{"name": "A", "address": {"street": "1", "city": "X"}}, {"name": "B", "address": {"street": "2"}}]}`,
			shouldFail:  true,
			errContains: "city",
		},

		// Map of nested objects
		{
			name:       "Map valid",
			endpoint:   "/rpc/Service/ValidateMap",
			payload:    `{"contacts": {"home": {"street": "123", "city": "NYC"}, "work": {"street": "456", "city": "LA"}}}`,
			shouldFail: false,
		},
		{
			name:        "Map value missing nested field",
			endpoint:    "/rpc/Service/ValidateMap",
			payload:     `{"contacts": {"home": {"street": "123", "city": "NYC"}, "work": {"street": "456"}}}`,
			shouldFail:  true,
			errContains: "city",
		},

		// Triple nesting: Organization -> Team[] -> Person -> Address
		{
			name:       "Organization valid",
			endpoint:   "/rpc/Service/ValidateOrganization",
			payload:    `{"org": {"name": "ACME", "teams": [{"leader": {"name": "L", "address": {"street": "S", "city": "C"}}, "members": []}], "contacts": {}}}`,
			shouldFail: false,
		},
		{
			name:        "Organization nested team leader missing city",
			endpoint:    "/rpc/Service/ValidateOrganization",
			payload:     `{"org": {"name": "ACME", "teams": [{"leader": {"name": "L", "address": {"street": "S"}}, "members": []}], "contacts": {}}}`,
			shouldFail:  true,
			errContains: "city",
		},
	}

	for _, tc := range testCases {
		runTestCase(baseURL, tc.name, tc.endpoint, tc.payload, tc.shouldFail, tc.errContains)
	}
}

func runOptionalNestedTestCases(baseURL string) {
	testCases := []struct {
		name        string
		endpoint    string
		payload     string
		shouldFail  bool
		errContains string
	}{
		// === OptionalNested tests ===
		// When optional field is absent, no validation error
		{
			name:       "OptionalNested - details absent (valid)",
			endpoint:   "/rpc/Service/ValidateOptionalNested",
			payload:    `{"data": {"id": "test-1"}}`,
			shouldFail: false,
		},
		// When optional field is present with all required fields
		{
			name:       "OptionalNested - details present and complete (valid)",
			endpoint:   "/rpc/Service/ValidateOptionalNested",
			payload:    `{"data": {"id": "test-2", "details": {"required1": "value", "required2": 42}}}`,
			shouldFail: false,
		},
		// When optional field is present but missing required inner field
		{
			name:        "OptionalNested - details present but missing required1",
			endpoint:    "/rpc/Service/ValidateOptionalNested",
			payload:     `{"data": {"id": "test-3", "details": {"required2": 42}}}`,
			shouldFail:  true,
			errContains: "required1",
		},
		{
			name:        "OptionalNested - details present but missing required2",
			endpoint:    "/rpc/Service/ValidateOptionalNested",
			payload:     `{"data": {"id": "test-4", "details": {"required1": "value"}}}`,
			shouldFail:  true,
			errContains: "required2",
		},
		// When optional field is present but empty object
		{
			name:        "OptionalNested - details present but empty object",
			endpoint:    "/rpc/Service/ValidateOptionalNested",
			payload:     `{"data": {"id": "test-5", "details": {}}}`,
			shouldFail:  true,
			errContains: "required",
		},

		// === DeepOptionalNested tests ===
		// level1 absent - valid
		{
			name:       "DeepOptionalNested - level1 absent (valid)",
			endpoint:   "/rpc/Service/ValidateDeepOptionalNested",
			payload:    `{"data": {"id": "deep-1"}}`,
			shouldFail: false,
		},
		// level1 present with only required1 (level2 absent) - valid
		{
			name:       "DeepOptionalNested - level1 with required1, level2 absent (valid)",
			endpoint:   "/rpc/Service/ValidateDeepOptionalNested",
			payload:    `{"data": {"id": "deep-2", "level1": {"required1": "val1"}}}`,
			shouldFail: false,
		},
		// level1 present but missing required1 - error
		{
			name:        "DeepOptionalNested - level1 present but missing required1",
			endpoint:    "/rpc/Service/ValidateDeepOptionalNested",
			payload:     `{"data": {"id": "deep-3", "level1": {}}}`,
			shouldFail:  true,
			errContains: "required1",
		},
		// level1 and level2 present, level2 missing required2 - error
		{
			name:        "DeepOptionalNested - level2 present but missing required2",
			endpoint:    "/rpc/Service/ValidateDeepOptionalNested",
			payload:     `{"data": {"id": "deep-4", "level1": {"required1": "val1", "level2": {"level3": {"required3": "val3"}}}}}`,
			shouldFail:  true,
			errContains: "required2",
		},
		// All levels complete - valid
		{
			name:       "DeepOptionalNested - all levels complete (valid)",
			endpoint:   "/rpc/Service/ValidateDeepOptionalNested",
			payload:    `{"data": {"id": "deep-5", "level1": {"required1": "val1", "level2": {"required2": "val2", "level3": {"required3": "val3"}}}}}`,
			shouldFail: false,
		},
		// level3 missing required3 - error
		{
			name:        "DeepOptionalNested - level3 missing required3",
			endpoint:    "/rpc/Service/ValidateDeepOptionalNested",
			payload:     `{"data": {"id": "deep-6", "level1": {"required1": "val1", "level2": {"required2": "val2", "level3": {}}}}}`,
			shouldFail:  true,
			errContains: "required3",
		},

		// === OptionalArrayNested tests ===
		// items absent - valid
		{
			name:       "OptionalArrayNested - items absent (valid)",
			endpoint:   "/rpc/Service/ValidateOptionalArrayNested",
			payload:    `{"data": {"id": "arr-1"}}`,
			shouldFail: false,
		},
		// items present with valid elements - valid
		{
			name:       "OptionalArrayNested - items with valid elements (valid)",
			endpoint:   "/rpc/Service/ValidateOptionalArrayNested",
			payload:    `{"data": {"id": "arr-2", "items": [{"required1": "a", "required2": 1}, {"required1": "b", "required2": 2}]}}`,
			shouldFail: false,
		},
		// items present with empty array - valid
		{
			name:       "OptionalArrayNested - empty items array (valid)",
			endpoint:   "/rpc/Service/ValidateOptionalArrayNested",
			payload:    `{"data": {"id": "arr-3", "items": []}}`,
			shouldFail: false,
		},
		// items present with invalid element - error
		{
			name:        "OptionalArrayNested - array element missing required1",
			endpoint:    "/rpc/Service/ValidateOptionalArrayNested",
			payload:     `{"data": {"id": "arr-4", "items": [{"required1": "a", "required2": 1}, {"required2": 2}]}}`,
			shouldFail:  true,
			errContains: "required1",
		},

		// === OptionalMapNested tests ===
		// entries absent - valid
		{
			name:       "OptionalMapNested - entries absent (valid)",
			endpoint:   "/rpc/Service/ValidateOptionalMapNested",
			payload:    `{"data": {"id": "map-1"}}`,
			shouldFail: false,
		},
		// entries present with valid values - valid
		{
			name:       "OptionalMapNested - entries with valid values (valid)",
			endpoint:   "/rpc/Service/ValidateOptionalMapNested",
			payload:    `{"data": {"id": "map-2", "entries": {"key1": {"required1": "a", "required2": 1}, "key2": {"required1": "b", "required2": 2}}}}`,
			shouldFail: false,
		},
		// entries present with empty map - valid
		{
			name:       "OptionalMapNested - empty entries map (valid)",
			endpoint:   "/rpc/Service/ValidateOptionalMapNested",
			payload:    `{"data": {"id": "map-3", "entries": {}}}`,
			shouldFail: false,
		},
		// entries present with invalid value - error
		{
			name:        "OptionalMapNested - map value missing required2",
			endpoint:    "/rpc/Service/ValidateOptionalMapNested",
			payload:     `{"data": {"id": "map-4", "entries": {"key1": {"required1": "a", "required2": 1}, "key2": {"required1": "b"}}}}`,
			shouldFail:  true,
			errContains: "required2",
		},

		// === ComplexOptionalValidation tests ===
		// All optional fields absent - valid
		{
			name:       "ComplexOptional - all absent (valid)",
			endpoint:   "/rpc/Service/ValidateComplexOptional",
			payload:    `{"data": {"id": "complex-1"}}`,
			shouldFail: false,
		},
		// simpleOptional present - valid (just a string)
		{
			name:       "ComplexOptional - simpleOptional present (valid)",
			endpoint:   "/rpc/Service/ValidateComplexOptional",
			payload:    `{"data": {"id": "complex-2", "simpleOptional": "hello"}}`,
			shouldFail: false,
		},
		// nestedOptional present and complete - valid
		{
			name:       "ComplexOptional - nestedOptional complete (valid)",
			endpoint:   "/rpc/Service/ValidateComplexOptional",
			payload:    `{"data": {"id": "complex-3", "nestedOptional": {"required1": "a", "required2": 1}}}`,
			shouldFail: false,
		},
		// nestedOptional present but incomplete - error
		{
			name:        "ComplexOptional - nestedOptional incomplete",
			endpoint:    "/rpc/Service/ValidateComplexOptional",
			payload:     `{"data": {"id": "complex-4", "nestedOptional": {"required1": "a"}}}`,
			shouldFail:  true,
			errContains: "required2",
		},
		// deepOptional with only outerRequired - valid (innerOptional is optional)
		{
			name:       "ComplexOptional - deepOptional with outerRequired only (valid)",
			endpoint:   "/rpc/Service/ValidateComplexOptional",
			payload:    `{"data": {"id": "complex-5", "deepOptional": {"outerRequired": "outer"}}}`,
			shouldFail: false,
		},
		// deepOptional with innerOptional complete - valid
		{
			name:       "ComplexOptional - deepOptional with innerOptional complete (valid)",
			endpoint:   "/rpc/Service/ValidateComplexOptional",
			payload:    `{"data": {"id": "complex-6", "deepOptional": {"outerRequired": "outer", "innerOptional": {"innerRequired": "inner"}}}}`,
			shouldFail: false,
		},
		// deepOptional missing outerRequired - error
		{
			name:        "ComplexOptional - deepOptional missing outerRequired",
			endpoint:    "/rpc/Service/ValidateComplexOptional",
			payload:     `{"data": {"id": "complex-7", "deepOptional": {}}}`,
			shouldFail:  true,
			errContains: "outerRequired",
		},
		// deepOptional with innerOptional present but missing innerRequired - error
		{
			name:        "ComplexOptional - innerOptional present but missing innerRequired",
			endpoint:    "/rpc/Service/ValidateComplexOptional",
			payload:     `{"data": {"id": "complex-8", "deepOptional": {"outerRequired": "outer", "innerOptional": {}}}}`,
			shouldFail:  true,
			errContains: "innerRequired",
		},
		// Mix of valid optional fields - valid
		{
			name:       "ComplexOptional - mixed valid optionals (valid)",
			endpoint:   "/rpc/Service/ValidateComplexOptional",
			payload:    `{"data": {"id": "complex-9", "simpleOptional": "test", "arrayOptional": [{"required1": "a", "required2": 1}], "mapOptional": {"k": {"required1": "b", "required2": 2}}}}`,
			shouldFail: false,
		},
	}

	for _, tc := range testCases {
		runTestCase(baseURL, tc.name, tc.endpoint, tc.payload, tc.shouldFail, tc.errContains)
	}
}

func runTestCase(baseURL, name, endpoint, payload string, shouldFail bool, errContains string) {
	resp, err := http.Post(baseURL+endpoint, "application/json", bytes.NewBufferString(payload))
	if err != nil {
		panic(fmt.Sprintf("%s: %v", name, err))
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("%s: expected 200, got %d", name, resp.StatusCode))
	}

	var res map[string]any
	if err := json.Unmarshal(body, &res); err != nil {
		panic(fmt.Sprintf("%s: %v", name, err))
	}

	isOk := res["ok"] == true

	if shouldFail && isOk {
		panic(fmt.Sprintf("%s: expected validation error, got success", name))
	}

	if !shouldFail && !isOk {
		panic(fmt.Sprintf("%s: expected success, got error: %v", name, res["error"]))
	}

	if shouldFail && errContains != "" {
		errMsg := fmt.Sprintf("%v", res["error"])
		if !strings.Contains(strings.ToLower(errMsg), strings.ToLower(errContains)) {
			panic(fmt.Sprintf("%s: error should contain '%s', got: %s", name, errContains, errMsg))
		}
	}
}
