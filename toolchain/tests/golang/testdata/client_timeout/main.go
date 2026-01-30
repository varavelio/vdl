// Verifies client timeout configuration: the server handler sleeps for 500ms,
// but the client is configured with a 100ms timeout, so it should fail with REQUEST_TIMEOUT.
package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

type AppProps struct{}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.Service().Procs.Slow().Handle(func(c *gen.ServiceSlowHandlerContext[AppProps]) (gen.ServiceSlowOutput, error) {
		time.Sleep(500 * time.Millisecond)
		return gen.ServiceSlowOutput{}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()
	timeoutConf := gen.TimeoutConfig{Timeout: 100 * time.Millisecond}

	start := time.Now()
	_, err := client.RPCs.Service().Procs.Slow().
		WithTimeoutConfig(timeoutConf).
		Execute(context.Background(), gen.ServiceSlowInput{})
	duration := time.Since(start)

	if err == nil {
		panic("expected timeout error, got nil")
	}

	vdlErr, ok := err.(gen.Error)
	if !ok {
		panic(fmt.Sprintf("expected gen.Error, got %T: %v", err, err))
	}
	if vdlErr.Code != "REQUEST_TIMEOUT" {
		panic(fmt.Sprintf("expected error code REQUEST_TIMEOUT, got %s", vdlErr.Code))
	}
	if duration > 300*time.Millisecond {
		panic(fmt.Sprintf("client waited too long: %v", duration))
	}

	fmt.Println("Success")
}
