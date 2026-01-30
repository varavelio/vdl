// Verifies spread (type composition with ...) works correctly.
// Tests simple spreads, chained spreads, multiple spreads, spreads with complex nested types,
// and spreads directly in input/output blocks and nested anonymous objects.
package main

import (
	"context"
	"e2e/gen"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
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

	// New handlers for spread in input/output/anonymous
	server.RPCs.Service().Procs.SpreadInInput().Handle(func(c *gen.ServiceSpreadInInputHandlerContext[AppProps]) (gen.ServiceSpreadInInputOutput, error) {
		return gen.ServiceSpreadInInputOutput{
			Id:          c.Input.Id,
			CreatedAt:   c.Input.CreatedAt,
			UpdatedAt:   c.Input.UpdatedAt,
			CustomField: c.Input.CustomField,
		}, nil
	})

	server.RPCs.Service().Procs.SpreadInOutput().Handle(func(c *gen.ServiceSpreadInOutputHandlerContext[AppProps]) (gen.ServiceSpreadInOutputOutput, error) {
		now := time.Now().UTC().Truncate(time.Second)
		return gen.ServiceSpreadInOutputOutput{
			Id:        c.Input.Id,
			CreatedAt: now,
			UpdatedAt: now.Add(time.Hour),
			Tags:      []string{"generated", "processed"},
			Metadata:  map[string]string{"source": "server"},
			Name:      c.Input.Name,
			Processed: true,
		}, nil
	})

	server.RPCs.Service().Procs.SpreadInNestedAnonymous().Handle(func(c *gen.ServiceSpreadInNestedAnonymousHandlerContext[AppProps]) (gen.ServiceSpreadInNestedAnonymousOutput, error) {
		return gen.ServiceSpreadInNestedAnonymousOutput{
			Wrapper: gen.ServiceSpreadInNestedAnonymousOutputWrapper{
				Id: c.Input.Wrapper.Id,
				Data: gen.ServiceSpreadInNestedAnonymousOutputWrapperData{
					CreatedAt: c.Input.Wrapper.Data.CreatedAt,
					UpdatedAt: c.Input.Wrapper.Data.UpdatedAt,
					Tags:      c.Input.Wrapper.Data.Tags,
					Metadata:  c.Input.Wrapper.Data.Metadata,
					Value:     c.Input.Wrapper.Data.Value,
				},
			},
		}, nil
	})

	server.RPCs.Service().Procs.DeepNestedSpreads().Handle(func(c *gen.ServiceDeepNestedSpreadsHandlerContext[AppProps]) (gen.ServiceDeepNestedSpreadsOutput, error) {
		return gen.ServiceDeepNestedSpreadsOutput{
			Level1: gen.ServiceDeepNestedSpreadsOutputLevel1{
				Id: c.Input.Level1.Id,
				Level2: gen.ServiceDeepNestedSpreadsOutputLevel1Level2{
					CreatedAt: c.Input.Level1.Level2.CreatedAt,
					UpdatedAt: c.Input.Level1.Level2.UpdatedAt,
					Level3: gen.ServiceDeepNestedSpreadsOutputLevel1Level2Level3{
						Tags:      c.Input.Level1.Level2.Level3.Tags,
						Metadata:  c.Input.Level1.Level2.Level3.Metadata,
						DeepValue: c.Input.Level1.Level2.Level3.DeepValue,
					},
				},
			},
		}, nil
	})

	server.RPCs.Service().Procs.SpreadInArrayAnonymous().Handle(func(c *gen.ServiceSpreadInArrayAnonymousHandlerContext[AppProps]) (gen.ServiceSpreadInArrayAnonymousOutput, error) {
		outputItems := make([]gen.ServiceSpreadInArrayAnonymousOutputItems, len(c.Input.Items))
		for i, item := range c.Input.Items {
			outputItems[i] = gen.ServiceSpreadInArrayAnonymousOutputItems{
				Id:        item.Id,
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
				Name:      item.Name,
			}
		}
		return gen.ServiceSpreadInArrayAnonymousOutput{
			Items: outputItems,
			Count: int64(len(c.Input.Items)),
		}, nil
	})

	server.RPCs.Service().Procs.SpreadInMapAnonymous().Handle(func(c *gen.ServiceSpreadInMapAnonymousHandlerContext[AppProps]) (gen.ServiceSpreadInMapAnonymousOutput, error) {
		outputEntries := make(map[string]gen.ServiceSpreadInMapAnonymousOutputEntries)
		keys := make([]string, 0, len(c.Input.Entries))
		for k, v := range c.Input.Entries {
			keys = append(keys, k)
			outputEntries[k] = gen.ServiceSpreadInMapAnonymousOutputEntries{
				CreatedAt: v.CreatedAt,
				UpdatedAt: v.UpdatedAt,
				CreatedBy: v.CreatedBy,
				UpdatedBy: v.UpdatedBy,
				Value:     v.Value,
			}
		}
		sort.Strings(keys)
		return gen.ServiceSpreadInMapAnonymousOutput{
			Entries: outputEntries,
			Keys:    keys,
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

	// Test 8: Spread directly in input block
	testSpreadInInput(ctx, client, now)

	// Test 9: Spread directly in output block
	testSpreadInOutput(ctx, client)

	// Test 10: Spread in nested anonymous objects
	testSpreadInNestedAnonymous(ctx, client, now)

	// Test 11: Deep nested spreads in anonymous objects
	testDeepNestedSpreads(ctx, client, now)

	// Test 12: Spread in anonymous object inside array
	testSpreadInArrayAnonymous(ctx, client, now)

	// Test 13: Spread in anonymous object inside map
	testSpreadInMapAnonymous(ctx, client, now)

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

func testSpreadInInput(ctx context.Context, client *gen.Client, now time.Time) {
	// Input block has ...Identifiable and ...Timestamps spread directly
	input := gen.ServiceSpreadInInputInput{
		Id:          "input-spread-1",
		CreatedAt:   now,
		UpdatedAt:   now.Add(time.Minute),
		CustomField: "my-custom-value",
	}

	res, err := client.RPCs.Service().Procs.SpreadInInput().Execute(ctx, input)
	if err != nil {
		panic(fmt.Sprintf("SpreadInInput failed: %v", err))
	}

	if res.Id != input.Id {
		panic(fmt.Sprintf("Id mismatch: got %s, want %s", res.Id, input.Id))
	}
	if !res.CreatedAt.Equal(input.CreatedAt) {
		panic(fmt.Sprintf("CreatedAt mismatch: got %v, want %v", res.CreatedAt, input.CreatedAt))
	}
	if !res.UpdatedAt.Equal(input.UpdatedAt) {
		panic(fmt.Sprintf("UpdatedAt mismatch: got %v, want %v", res.UpdatedAt, input.UpdatedAt))
	}
	if res.CustomField != input.CustomField {
		panic(fmt.Sprintf("CustomField mismatch: got %s, want %s", res.CustomField, input.CustomField))
	}
}

func testSpreadInOutput(ctx context.Context, client *gen.Client) {
	// Output block has ...Identifiable, ...Timestamps, ...Taggable spread directly
	input := gen.ServiceSpreadInOutputInput{
		Id:   "output-spread-1",
		Name: "test-name",
	}

	res, err := client.RPCs.Service().Procs.SpreadInOutput().Execute(ctx, input)
	if err != nil {
		panic(fmt.Sprintf("SpreadInOutput failed: %v", err))
	}

	// Verify spread fields from Identifiable
	if res.Id != input.Id {
		panic(fmt.Sprintf("Id mismatch: got %s, want %s", res.Id, input.Id))
	}

	// Verify spread fields from Timestamps (server generates these)
	if res.CreatedAt.IsZero() {
		panic("CreatedAt should not be zero")
	}
	if res.UpdatedAt.IsZero() {
		panic("UpdatedAt should not be zero")
	}

	// Verify spread fields from Taggable
	if len(res.Tags) == 0 {
		panic("Tags should not be empty")
	}
	if len(res.Metadata) == 0 {
		panic("Metadata should not be empty")
	}

	// Verify own fields
	if res.Name != input.Name {
		panic(fmt.Sprintf("Name mismatch: got %s, want %s", res.Name, input.Name))
	}
	if !res.Processed {
		panic("Processed should be true")
	}
}

func testSpreadInNestedAnonymous(ctx context.Context, client *gen.Client, now time.Time) {
	// Nested anonymous object has spreads: wrapper: { ...Identifiable, data: { ...Timestamps, ...Taggable, value } }
	input := gen.ServiceSpreadInNestedAnonymousInput{
		Wrapper: gen.ServiceSpreadInNestedAnonymousInputWrapper{
			Id: "nested-anon-1",
			Data: gen.ServiceSpreadInNestedAnonymousInputWrapperData{
				CreatedAt: now,
				UpdatedAt: now.Add(time.Hour),
				Tags:      []string{"nested", "anonymous"},
				Metadata:  map[string]string{"key": "value"},
				Value:     "inner-value",
			},
		},
	}

	res, err := client.RPCs.Service().Procs.SpreadInNestedAnonymous().Execute(ctx, input)
	if err != nil {
		panic(fmt.Sprintf("SpreadInNestedAnonymous failed: %v", err))
	}

	// Verify wrapper.id from Identifiable spread
	if res.Wrapper.Id != input.Wrapper.Id {
		panic(fmt.Sprintf("Wrapper.Id mismatch: got %s, want %s", res.Wrapper.Id, input.Wrapper.Id))
	}

	// Verify wrapper.data fields from Timestamps spread
	if !res.Wrapper.Data.CreatedAt.Equal(input.Wrapper.Data.CreatedAt) {
		panic("Wrapper.Data.CreatedAt mismatch")
	}
	if !res.Wrapper.Data.UpdatedAt.Equal(input.Wrapper.Data.UpdatedAt) {
		panic("Wrapper.Data.UpdatedAt mismatch")
	}

	// Verify wrapper.data fields from Taggable spread
	if !reflect.DeepEqual(res.Wrapper.Data.Tags, input.Wrapper.Data.Tags) {
		panic("Wrapper.Data.Tags mismatch")
	}
	if !reflect.DeepEqual(res.Wrapper.Data.Metadata, input.Wrapper.Data.Metadata) {
		panic("Wrapper.Data.Metadata mismatch")
	}

	// Verify own field
	if res.Wrapper.Data.Value != input.Wrapper.Data.Value {
		panic("Wrapper.Data.Value mismatch")
	}
}

func testDeepNestedSpreads(ctx context.Context, client *gen.Client, now time.Time) {
	// Deep nesting: level1: { ...Identifiable, level2: { ...Timestamps, level3: { ...Taggable, deepValue } } }
	input := gen.ServiceDeepNestedSpreadsInput{
		Level1: gen.ServiceDeepNestedSpreadsInputLevel1{
			Id: "deep-1",
			Level2: gen.ServiceDeepNestedSpreadsInputLevel1Level2{
				CreatedAt: now,
				UpdatedAt: now.Add(time.Minute * 30),
				Level3: gen.ServiceDeepNestedSpreadsInputLevel1Level2Level3{
					Tags:      []string{"level3", "deep"},
					Metadata:  map[string]string{"depth": "3"},
					DeepValue: "innermost-value",
				},
			},
		},
	}

	res, err := client.RPCs.Service().Procs.DeepNestedSpreads().Execute(ctx, input)
	if err != nil {
		panic(fmt.Sprintf("DeepNestedSpreads failed: %v", err))
	}

	// Verify level1 (Identifiable spread)
	if res.Level1.Id != input.Level1.Id {
		panic("Level1.Id mismatch")
	}

	// Verify level2 (Timestamps spread)
	if !res.Level1.Level2.CreatedAt.Equal(input.Level1.Level2.CreatedAt) {
		panic("Level1.Level2.CreatedAt mismatch")
	}
	if !res.Level1.Level2.UpdatedAt.Equal(input.Level1.Level2.UpdatedAt) {
		panic("Level1.Level2.UpdatedAt mismatch")
	}

	// Verify level3 (Taggable spread)
	if !reflect.DeepEqual(res.Level1.Level2.Level3.Tags, input.Level1.Level2.Level3.Tags) {
		panic("Level1.Level2.Level3.Tags mismatch")
	}
	if !reflect.DeepEqual(res.Level1.Level2.Level3.Metadata, input.Level1.Level2.Level3.Metadata) {
		panic("Level1.Level2.Level3.Metadata mismatch")
	}
	if res.Level1.Level2.Level3.DeepValue != input.Level1.Level2.Level3.DeepValue {
		panic("Level1.Level2.Level3.DeepValue mismatch")
	}
}

func testSpreadInArrayAnonymous(ctx context.Context, client *gen.Client, now time.Time) {
	// Array of anonymous objects with spreads: items: { ...Identifiable, ...Timestamps, name }[]
	input := gen.ServiceSpreadInArrayAnonymousInput{
		Items: []gen.ServiceSpreadInArrayAnonymousInputItems{
			{Id: "item-1", CreatedAt: now, UpdatedAt: now.Add(time.Second), Name: "First"},
			{Id: "item-2", CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute * 2), Name: "Second"},
			{Id: "item-3", CreatedAt: now.Add(time.Hour), UpdatedAt: now.Add(time.Hour * 2), Name: "Third"},
		},
	}

	res, err := client.RPCs.Service().Procs.SpreadInArrayAnonymous().Execute(ctx, input)
	if err != nil {
		panic(fmt.Sprintf("SpreadInArrayAnonymous failed: %v", err))
	}

	if res.Count != 3 {
		panic(fmt.Sprintf("Count mismatch: got %d, want 3", res.Count))
	}

	if len(res.Items) != 3 {
		panic(fmt.Sprintf("Items length mismatch: got %d, want 3", len(res.Items)))
	}

	for i, item := range res.Items {
		inputItem := input.Items[i]
		if item.Id != inputItem.Id {
			panic(fmt.Sprintf("Item[%d].Id mismatch", i))
		}
		if !item.CreatedAt.Equal(inputItem.CreatedAt) {
			panic(fmt.Sprintf("Item[%d].CreatedAt mismatch", i))
		}
		if !item.UpdatedAt.Equal(inputItem.UpdatedAt) {
			panic(fmt.Sprintf("Item[%d].UpdatedAt mismatch", i))
		}
		if item.Name != inputItem.Name {
			panic(fmt.Sprintf("Item[%d].Name mismatch", i))
		}
	}
}

func testSpreadInMapAnonymous(ctx context.Context, client *gen.Client, now time.Time) {
	// Map of anonymous objects with spreads: entries: map<{ ...Auditable, value }>
	// Auditable has ...Timestamps, so entries have: createdAt, updatedAt, createdBy, updatedBy, value
	input := gen.ServiceSpreadInMapAnonymousInput{
		Entries: map[string]gen.ServiceSpreadInMapAnonymousInputEntries{
			"alpha": {
				CreatedAt: now,
				UpdatedAt: now.Add(time.Second),
				CreatedBy: "user-a",
				UpdatedBy: "user-b",
				Value:     "alpha-value",
			},
			"beta": {
				CreatedAt: now.Add(time.Hour),
				UpdatedAt: now.Add(time.Hour * 2),
				CreatedBy: "user-c",
				UpdatedBy: "user-d",
				Value:     "beta-value",
			},
		},
	}

	res, err := client.RPCs.Service().Procs.SpreadInMapAnonymous().Execute(ctx, input)
	if err != nil {
		panic(fmt.Sprintf("SpreadInMapAnonymous failed: %v", err))
	}

	// Verify keys are returned sorted
	expectedKeys := []string{"alpha", "beta"}
	if !reflect.DeepEqual(res.Keys, expectedKeys) {
		panic(fmt.Sprintf("Keys mismatch: got %v, want %v", res.Keys, expectedKeys))
	}

	// Verify each entry
	for key, inputEntry := range input.Entries {
		outputEntry, ok := res.Entries[key]
		if !ok {
			panic(fmt.Sprintf("Missing entry for key: %s", key))
		}

		if !outputEntry.CreatedAt.Equal(inputEntry.CreatedAt) {
			panic(fmt.Sprintf("Entry[%s].CreatedAt mismatch", key))
		}
		if !outputEntry.UpdatedAt.Equal(inputEntry.UpdatedAt) {
			panic(fmt.Sprintf("Entry[%s].UpdatedAt mismatch", key))
		}
		if outputEntry.CreatedBy != inputEntry.CreatedBy {
			panic(fmt.Sprintf("Entry[%s].CreatedBy mismatch", key))
		}
		if outputEntry.UpdatedBy != inputEntry.UpdatedBy {
			panic(fmt.Sprintf("Entry[%s].UpdatedBy mismatch", key))
		}
		if outputEntry.Value != inputEntry.Value {
			panic(fmt.Sprintf("Entry[%s].Value mismatch", key))
		}
	}
}
