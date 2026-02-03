package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
)

type AppProps struct{}

func main() {
	server := gen.NewServer[AppProps]()

	server.RPCs.TestService().Procs.TestCircular().Handle(func(c *gen.TestServiceTestCircularHandlerContext[AppProps]) (gen.TestServiceTestCircularOutput, error) {
		return gen.TestServiceTestCircularOutput{
			Self:     c.Input.Self,
			Chain:    c.Input.Chain,
			Optional: c.Input.Optional,
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

	testSelfReferencing(client)
	testChainWithOptional(client)
	testFullyOptionalChain(client)

	fmt.Println("Success")
}

func testSelfReferencing(client *gen.Client) {
	selfRef := gen.SelfReferencing{
		Id:   "root",
		Name: "Root Node",
		Parent: gen.Ptr(gen.SelfReferencing{
			Id:   "parent",
			Name: "Parent Node",
			Parent: gen.Ptr(gen.SelfReferencing{
				Id:     "grandparent",
				Name:   "Grandparent Node",
				Parent: nil,
			}),
		}),
	}

	res, err := client.RPCs.TestService().Procs.TestCircular().Execute(context.Background(), gen.TestServiceTestCircularInput{
		Self:     selfRef,
		Chain:    buildNodeChain(),
		Optional: buildOptionalChain(),
	})
	if err != nil {
		panic(fmt.Sprintf("self-referencing test failed: %v", err))
	}

	if !reflect.DeepEqual(res.Self, selfRef) {
		panic("self-referencing data mismatch")
	}
}

func testChainWithOptional(client *gen.Client) {
	chain := buildNodeChain()

	res, err := client.RPCs.TestService().Procs.TestCircular().Execute(context.Background(), gen.TestServiceTestCircularInput{
		Self:     gen.SelfReferencing{Id: "test", Name: "Test"},
		Chain:    chain,
		Optional: buildOptionalChain(),
	})
	if err != nil {
		panic(fmt.Sprintf("chain test failed: %v", err))
	}

	if !reflect.DeepEqual(res.Chain, chain) {
		panic("chain data mismatch")
	}
}

func testFullyOptionalChain(client *gen.Client) {
	optional := buildOptionalChain()

	res, err := client.RPCs.TestService().Procs.TestCircular().Execute(context.Background(), gen.TestServiceTestCircularInput{
		Self:     gen.SelfReferencing{Id: "test", Name: "Test"},
		Chain:    buildNodeChain(),
		Optional: optional,
	})
	if err != nil {
		panic(fmt.Sprintf("fully optional test failed: %v", err))
	}

	if !reflect.DeepEqual(res.Optional, optional) {
		panic("fully optional data mismatch")
	}
}

func buildNodeChain() gen.NodeA {
	return gen.NodeA{
		Value: "A",
		NodeB: gen.NodeB{
			Value: "B",
			NodeC: gen.NodeC{
				Value: "C",
				NodeD: gen.NodeD{
					Value: "D",
					NodeE: gen.NodeE{
						Value:   "E",
						BackToA: nil,
					},
				},
			},
		},
	}
}

func buildOptionalChain() gen.FullyOptionalA {
	return gen.FullyOptionalA{
		Id: "A",
		B: gen.Ptr(gen.FullyOptionalB{
			Id: "B",
			C: gen.Ptr(gen.FullyOptionalC{
				Id: "C",
				D: gen.Ptr(gen.FullyOptionalD{
					Id: "D",
					A:  nil,
				}),
			}),
		}),
	}
}
