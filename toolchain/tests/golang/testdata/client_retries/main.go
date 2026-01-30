// Verifies client retry logic: the server fails with 500 on first 2 attempts,
// then succeeds on the 3rd. The client should automatically retry and eventually succeed.
package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"
)

type AppProps struct{}

func main() {
	server := gen.NewServer[AppProps]()
	var attempts atomic.Int32

	server.RPCs.Service().Procs.Flaky().Handle(func(c *gen.ServiceFlakyHandlerContext[AppProps]) (gen.ServiceFlakyOutput, error) {
		if attempts.Add(1) < 3 {
			panic("simulated server error")
		}
		return gen.ServiceFlakyOutput{}, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()
	retryConf := gen.RetryConfig{
		MaxAttempts:     3,
		InitialDelay:    10 * time.Millisecond,
		MaxDelay:        100 * time.Millisecond,
		DelayMultiplier: 1.0,
		Jitter:          0.0,
	}

	_, err := client.RPCs.Service().Procs.Flaky().
		WithRetryConfig(retryConf).
		Execute(context.Background(), gen.ServiceFlakyInput{})

	if err != nil {
		panic(fmt.Sprintf("expected success after retries, got: %v", err))
	}
	if attempts.Load() != 3 {
		panic(fmt.Sprintf("expected 3 attempts, got %d", attempts.Load()))
	}

	fmt.Println("Success")
}
