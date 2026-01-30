// Verifies SSE payload escaping: messages with newlines, unicode,
// and special characters should be correctly transmitted through SSE.
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

	server.SetStreamConfig(gen.StreamConfig{
		PingInterval: 10 * time.Second, // Disable pings for this test
	})

	testMessages := []string{
		"simple message",
		"message with\nnewline",
		"message with\r\nCRLF",
		"message with\ttab",
		"unicode: 你好世界 🎉 émojis",
		`message with "quotes" and 'apostrophes'`,
		"message with backslash: \\path\\to\\file",
		"message with special chars: <>&",
		"multi\nline\nmessage\nwith\nmany\nbreaks",
		"",
	}

	server.RPCs.Service().Streams.Events().Handle(func(c *gen.ServiceEventsHandlerContext[AppProps], emit gen.ServiceEventsEmitFunc[AppProps]) error {
		for _, msg := range testMessages {
			if err := emit(c, gen.ServiceEventsOutput{Message: msg}); err != nil {
				return err
			}
		}
		return nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := client.RPCs.Service().Streams.Events().Execute(ctx, gen.ServiceEventsInput{})

	var received []string
	for evt := range stream {
		if !evt.Ok {
			panic(fmt.Sprintf("stream error: %v", evt.Error))
		}
		received = append(received, evt.Output.Message)
	}

	if len(received) != len(testMessages) {
		panic(fmt.Sprintf("expected %d messages, got %d", len(testMessages), len(received)))
	}

	for i, expected := range testMessages {
		if received[i] != expected {
			panic(fmt.Sprintf("message %d mismatch:\nexpected: %q\ngot: %q", i, expected, received[i]))
		}
	}

	fmt.Println("Success")
}
