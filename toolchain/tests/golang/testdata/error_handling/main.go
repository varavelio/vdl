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

	// Global Handler
	server.SetErrorHandler(func(c *gen.HandlerContext[AppProps, any], err error) gen.Error {
		return gen.Error{Message: "Global: " + err.Error()}
	})

	// RPC Specific Handler
	server.RPCs.Auth().SetErrorHandler(func(c *gen.HandlerContext[AppProps, any], err error) gen.Error {
		return gen.Error{Message: "Auth: " + err.Error()}
	})

	// Implement handlers that always fail
	server.RPCs.Users().Procs.Get().Handle(func(c *gen.UsersGetHandlerContext[AppProps]) (gen.UsersGetOutput, error) {
		return gen.UsersGetOutput{}, errors.New("fail")
	})
	server.RPCs.Auth().Procs.Login().Handle(func(c *gen.AuthLoginHandlerContext[AppProps]) (gen.AuthLoginOutput, error) {
		return gen.AuthLoginOutput{}, errors.New("fail")
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /rpc/{rpc}/{proc}", func(w http.ResponseWriter, r *http.Request) {
		rpcName := r.PathValue("rpc")
		procName := r.PathValue("proc")
		adapter := gen.NewNetHTTPAdapter(w, r)
		_ = server.HandleRequest(r.Context(), AppProps{}, rpcName, procName, adapter)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := gen.NewClient(ts.URL + "/rpc").Build()
	ctx := context.Background()

	// Test 1: Users.Get -> Should use Global Handler
	_, err := client.RPCs.Users().Procs.Get().Execute(ctx, gen.UsersGetInput{})
	if err == nil {
		panic("Expected error from Users.Get")
	}

	// Check error message
	if vdlErr, ok := err.(gen.Error); ok {
		if vdlErr.Message != "Global: fail" {
			panic(fmt.Sprintf("Expected 'Global: fail', got '%s'", vdlErr.Message))
		}
	} else {
		// If generated Error doesn't implement Error interface or is wrapped differently
		panic(fmt.Sprintf("Expected gen.Error, got %T: %v", err, err))
	}

	// Test 2: Auth.Login -> Should use Auth Handler
	_, err = client.RPCs.Auth().Procs.Login().Execute(ctx, gen.AuthLoginInput{})
	if err == nil {
		panic("Expected error from Auth.Login")
	}
	if vdlErr, ok := err.(gen.Error); ok {
		if vdlErr.Message != "Auth: fail" {
			panic(fmt.Sprintf("Expected 'Auth: fail', got '%s'", vdlErr.Message))
		}
	} else {
		panic(fmt.Sprintf("Expected gen.Error, got %T: %v", err, err))
	}

	fmt.Println("Success: Error handling precedence verified")
}
