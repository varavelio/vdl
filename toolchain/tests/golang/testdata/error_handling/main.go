// Verifies error handler precedence: RPC-level error handlers override global handlers.
// Users.Get uses global handler ("Global: fail"), Auth.Login uses RPC-specific handler ("Auth: fail").
package main

import (
	"context"
	"e2e/gen"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type AppProps struct{}

func main() {
	server := gen.NewServer[AppProps]()

	server.SetErrorHandler(func(c *gen.HandlerContext[AppProps, any], err error) gen.Error {
		return gen.Error{Message: "Global: " + err.Error()}
	})

	server.RPCs.Auth().SetErrorHandler(func(c *gen.HandlerContext[AppProps, any], err error) gen.Error {
		return gen.Error{Message: "Auth: " + err.Error()}
	})

	server.RPCs.Users().Procs.Get().Handle(func(c *gen.UsersGetHandlerContext[AppProps]) (gen.UsersGetOutput, error) {
		return gen.UsersGetOutput{}, errors.New("fail")
	})
	server.RPCs.Auth().Procs.Login().Handle(func(c *gen.AuthLoginHandlerContext[AppProps]) (gen.AuthLoginOutput, error) {
		return gen.AuthLoginOutput{}, errors.New("fail")
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, r.PathValue("rpc"), r.PathValue("proc"), adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()
	ctx := context.Background()

	_, err := client.RPCs.Users().Procs.Get().Execute(ctx, gen.UsersGetInput{})
	assertError(err, "Global: fail")

	_, err = client.RPCs.Auth().Procs.Login().Execute(ctx, gen.AuthLoginInput{})
	assertError(err, "Auth: fail")

	fmt.Println("Success")
}

func assertError(err error, expectedMsg string) {
	if err == nil {
		panic("expected error, got nil")
	}
	vdlErr, ok := err.(gen.Error)
	if !ok {
		panic(fmt.Sprintf("expected gen.Error, got %T: %v", err, err))
	}
	if vdlErr.Message != expectedMsg {
		panic(fmt.Sprintf("expected message '%s', got '%s'", expectedMsg, vdlErr.Message))
	}
}
