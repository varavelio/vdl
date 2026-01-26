// Verifies server-side input validation for deeply nested structures, arrays, and maps.
// Missing required fields at any nesting level should produce validation errors.
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

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

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
		resp, err := http.Post(ts.URL+tc.endpoint, "application/json", bytes.NewBufferString(tc.payload))
		if err != nil {
			panic(fmt.Sprintf("%s: %v", tc.name, err))
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != 200 {
			panic(fmt.Sprintf("%s: expected 200, got %d", tc.name, resp.StatusCode))
		}

		var res map[string]any
		if err := json.Unmarshal(body, &res); err != nil {
			panic(fmt.Sprintf("%s: %v", tc.name, err))
		}

		isOk := res["ok"] == true

		if tc.shouldFail && isOk {
			panic(fmt.Sprintf("%s: expected validation error, got success", tc.name))
		}

		if !tc.shouldFail && !isOk {
			panic(fmt.Sprintf("%s: expected success, got error: %v", tc.name, res["error"]))
		}

		if tc.shouldFail && tc.errContains != "" {
			errMsg := fmt.Sprintf("%v", res["error"])
			if !strings.Contains(strings.ToLower(errMsg), strings.ToLower(tc.errContains)) {
				panic(fmt.Sprintf("%s: error should contain '%s', got: %s", tc.name, tc.errContains, errMsg))
			}
		}
	}

	fmt.Println("Success")
}
