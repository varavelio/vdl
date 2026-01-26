// Verifies server-side input validation: sending invalid JSON (missing required field)
// should result in a validation error with ok=false.
package main

import (
	"bytes"
	"e2e/gen"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

type AppProps struct{}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Procs.Send().Handle(func(c *gen.ServiceSendHandlerContext[AppProps]) (gen.ServiceSendOutput, error) {
		return gen.ServiceSendOutput{}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	invalidJSON := `{"data": {}}`
	resp, err := http.Post(ts.URL+"/rpc/Service/Send", "application/json", bytes.NewBufferString(invalidJSON))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("expected 200 OK, got %d", resp.StatusCode))
	}

	var res map[string]any
	if err := json.Unmarshal(body, &res); err != nil {
		panic(err)
	}

	if res["ok"] == true {
		panic(fmt.Sprintf("expected validation error, got success: %s", string(body)))
	}

	errMap := res["error"].(map[string]any)
	fmt.Printf("Validation Error: %v\n", errMap["message"])

	fmt.Println("Success")
}
