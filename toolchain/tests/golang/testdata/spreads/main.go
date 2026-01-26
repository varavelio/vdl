// Verifies spread (type composition with ...) works correctly.
// Tests simple spreads, chained spreads, multiple spreads, and spreads with complex nested types.
package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"time"
)

type AppProps struct{}

func main() {
	server := gen.NewServer[AppProps]()

	// Register all handlers as echo handlers
	server.RPCs.Service().Procs.EchoUser().Handle(func(c *gen.ServiceEchoUserHandlerContext[AppProps]) (gen.ServiceEchoUserOutput, error) {
		return gen.ServiceEchoUserOutput{User: c.Input.User}, nil
	})

	server.RPCs.Service().Procs.EchoUserWithTimestamps().Handle(func(c *gen.ServiceEchoUserWithTimestampsHandlerContext[AppProps]) (gen.ServiceEchoUserWithTimestampsOutput, error) {
		return gen.ServiceEchoUserWithTimestampsOutput{User: c.Input.User}, nil
	})

	server.RPCs.Service().Procs.EchoAuditedEntity().Handle(func(c *gen.ServiceEchoAuditedEntityHandlerContext[AppProps]) (gen.ServiceEchoAuditedEntityOutput, error) {
		return gen.ServiceEchoAuditedEntityOutput{Entity: c.Input.Entity}, nil
	})

	server.RPCs.Service().Procs.EchoPerson().Handle(func(c *gen.ServiceEchoPersonHandlerContext[AppProps]) (gen.ServiceEchoPersonOutput, error) {
		return gen.ServiceEchoPersonOutput{Person: c.Input.Person}, nil
	})

	server.RPCs.Service().Procs.EchoDocument().Handle(func(c *gen.ServiceEchoDocumentHandlerContext[AppProps]) (gen.ServiceEchoDocumentOutput, error) {
		return gen.ServiceEchoDocumentOutput{Doc: c.Input.Doc}, nil
	})

	server.RPCs.Service().Procs.EchoServiceConfig().Handle(func(c *gen.ServiceEchoServiceConfigHandlerContext[AppProps]) (gen.ServiceEchoServiceConfigOutput, error) {
		return gen.ServiceEchoServiceConfigOutput{Config: c.Input.Config}, nil
	})

	server.RPCs.Service().Procs.EchoOrder().Handle(func(c *gen.ServiceEchoOrderHandlerContext[AppProps]) (gen.ServiceEchoOrderOutput, error) {
		return gen.ServiceEchoOrderOutput{Order: c.Input.Order}, nil
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
	now := time.Now().UTC().Truncate(time.Second)

	// Test 1: Simple spread (User has ...Identifiable)
	testSimpleSpread(ctx, client)

	// Test 2: Multiple spreads (UserWithTimestamps has ...Identifiable and ...Timestamps)
	testMultipleSpreads(ctx, client, now)

	// Test 3: Chained spreads (AuditedEntity has ...Auditable which has ...Timestamps)
	testChainedSpreads(ctx, client, now)

	// Test 4: Spread with nested objects (Person has ...Identifiable, ...Timestamps, and contact: ContactInfo)
	testSpreadWithNestedObjects(ctx, client, now)

	// Test 5: Spread with arrays and maps (Document has ...Identifiable, ...Timestamps, ...Taggable)
	testSpreadWithArraysAndMaps(ctx, client, now)

	// Test 6: Deep chain of spreads (ServiceConfig has ...AdvancedConfig which has ...BaseConfig)
	testDeepChainOfSpreads(ctx, client)

	// Test 7: Spread with array of objects (Order has ...Identifiable, ...Timestamps, items: OrderItem[])
	testSpreadWithArrayOfObjects(ctx, client, now)

	fmt.Println("Success")
}

func testSimpleSpread(ctx context.Context, client *gen.Client) {
	user := gen.User{
		Id:    "user-123",
		Name:  "Alice",
		Email: "alice@example.com",
	}

	res, err := client.RPCs.Service().Procs.EchoUser().Execute(ctx, gen.ServiceEchoUserInput{User: user})
	if err != nil {
		panic(fmt.Sprintf("EchoUser failed: %v", err))
	}

	if !reflect.DeepEqual(res.User, user) {
		panic(fmt.Sprintf("User mismatch: got %+v, want %+v", res.User, user))
	}
}

func testMultipleSpreads(ctx context.Context, client *gen.Client, now time.Time) {
	user := gen.UserWithTimestamps{
		Id:        "user-456",
		CreatedAt: now,
		UpdatedAt: now.Add(time.Hour),
		Name:      "Bob",
		Email:     "bob@example.com",
		DeletedAt: gen.Some(now.Add(24 * time.Hour)),
	}

	res, err := client.RPCs.Service().Procs.EchoUserWithTimestamps().Execute(ctx, gen.ServiceEchoUserWithTimestampsInput{User: user})
	if err != nil {
		panic(fmt.Sprintf("EchoUserWithTimestamps failed: %v", err))
	}

	if !reflect.DeepEqual(res.User, user) {
		panic(fmt.Sprintf("UserWithTimestamps mismatch: got %+v, want %+v", res.User, user))
	}
}

func testChainedSpreads(ctx context.Context, client *gen.Client, now time.Time) {
	// AuditedEntity has ...Auditable which has ...Timestamps
	// So it should have: createdAt, updatedAt (from Timestamps via Auditable), createdBy, updatedBy (from Auditable), entityType, entityId
	entity := gen.AuditedEntity{
		CreatedAt:  now,
		UpdatedAt:  now.Add(time.Hour),
		CreatedBy:  "system",
		UpdatedBy:  "admin",
		EntityType: "document",
		EntityId:   "doc-789",
	}

	res, err := client.RPCs.Service().Procs.EchoAuditedEntity().Execute(ctx, gen.ServiceEchoAuditedEntityInput{Entity: entity})
	if err != nil {
		panic(fmt.Sprintf("EchoAuditedEntity failed: %v", err))
	}

	if !reflect.DeepEqual(res.Entity, entity) {
		panic(fmt.Sprintf("AuditedEntity mismatch: got %+v, want %+v", res.Entity, entity))
	}
}

func testSpreadWithNestedObjects(ctx context.Context, client *gen.Client, now time.Time) {
	person := gen.Person{
		Id:        "person-111",
		CreatedAt: now,
		UpdatedAt: now.Add(time.Minute),
		Name:      "Charlie",
		Contact: gen.ContactInfo{
			Email: "charlie@example.com",
			Phone: gen.Some("+1-555-1234"),
			Address: gen.Address{
				Street:  "123 Main St",
				City:    "Springfield",
				Country: "USA",
			},
		},
	}

	res, err := client.RPCs.Service().Procs.EchoPerson().Execute(ctx, gen.ServiceEchoPersonInput{Person: person})
	if err != nil {
		panic(fmt.Sprintf("EchoPerson failed: %v", err))
	}

	if !reflect.DeepEqual(res.Person, person) {
		panic(fmt.Sprintf("Person mismatch: got %+v, want %+v", res.Person, person))
	}
}

func testSpreadWithArraysAndMaps(ctx context.Context, client *gen.Client, now time.Time) {
	doc := gen.Document{
		Id:        "doc-222",
		CreatedAt: now,
		UpdatedAt: now.Add(time.Second),
		Tags:      []string{"important", "urgent", "draft"},
		Metadata: map[string]string{
			"author":   "Diana",
			"version":  "1.0",
			"category": "technical",
		},
		Title:   "Technical Specification",
		Content: "This document describes the system architecture...",
	}

	res, err := client.RPCs.Service().Procs.EchoDocument().Execute(ctx, gen.ServiceEchoDocumentInput{Doc: doc})
	if err != nil {
		panic(fmt.Sprintf("EchoDocument failed: %v", err))
	}

	if !reflect.DeepEqual(res.Doc, doc) {
		panic(fmt.Sprintf("Document mismatch: got %+v, want %+v", res.Doc, doc))
	}
}

func testDeepChainOfSpreads(ctx context.Context, client *gen.Client) {
	// ServiceConfig has ...AdvancedConfig which has ...BaseConfig
	// So it should have: enabled, priority (from BaseConfig), retryCount, timeout (from AdvancedConfig), serviceName, endpoints
	config := gen.ServiceConfig{
		Enabled:     true,
		Priority:    10,
		RetryCount:  3,
		Timeout:     5000,
		ServiceName: "api-gateway",
		Endpoints:   []string{"https://api.example.com", "https://api-backup.example.com"},
	}

	res, err := client.RPCs.Service().Procs.EchoServiceConfig().Execute(ctx, gen.ServiceEchoServiceConfigInput{Config: config})
	if err != nil {
		panic(fmt.Sprintf("EchoServiceConfig failed: %v", err))
	}

	if !reflect.DeepEqual(res.Config, config) {
		panic(fmt.Sprintf("ServiceConfig mismatch: got %+v, want %+v", res.Config, config))
	}
}

func testSpreadWithArrayOfObjects(ctx context.Context, client *gen.Client, now time.Time) {
	order := gen.Order{
		Id:        "order-333",
		CreatedAt: now,
		UpdatedAt: now.Add(time.Millisecond * 500),
		Items: []gen.OrderItem{
			{ProductId: "prod-1", Quantity: 2, Price: 29.99},
			{ProductId: "prod-2", Quantity: 1, Price: 149.50},
			{ProductId: "prod-3", Quantity: 5, Price: 9.99},
		},
		Total:  259.43,
		Status: "pending",
	}

	res, err := client.RPCs.Service().Procs.EchoOrder().Execute(ctx, gen.ServiceEchoOrderInput{Order: order})
	if err != nil {
		panic(fmt.Sprintf("EchoOrder failed: %v", err))
	}

	if !reflect.DeepEqual(res.Order, order) {
		panic(fmt.Sprintf("Order mismatch: got %+v, want %+v", res.Order, order))
	}
}
