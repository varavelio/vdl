// Verifies Safe Accessors (Getters): GetField() and GetFieldOr() methods work correctly
// for both nil receivers, nil optional fields, and populated values.
package main

import (
	"e2e/gen"
	"encoding/json"
	"fmt"
)

func main() {
	// Test 1: Getters on nil receiver
	testNilReceiver()

	// Test 2: Getters on struct with nil optional fields
	testNilOptionalFields()

	// Test 3: Getters on fully populated struct
	testPopulatedStruct()

	// Test 4: GetFieldOr with default values
	testGetterWithDefaults()

	// Test 5: JSON marshaling with pointers (omitempty behavior)
	testJSONMarshal()

	// Test 6: JSON unmarshaling with null vs undefined
	testJSONUnmarshal()

	// Test 7: Ptr, Val, Or utility functions
	testPointerUtilities()

	fmt.Println("Success")
}

func testNilReceiver() {
	var config *gen.UserConfig

	// GetField() on nil receiver should return zero value
	if config.GetHost() != "" {
		panic("GetHost on nil receiver should return empty string")
	}
	if config.GetTimeout() != 0 {
		panic("GetTimeout on nil receiver should return 0")
	}
	if config.GetAdvanced() != (gen.AdvancedConfig{}) {
		panic("GetAdvanced on nil receiver should return zero AdvancedConfig")
	}

	// GetFieldOr() on nil receiver should return default value
	if config.GetHostOr("default") != "default" {
		panic("GetHostOr on nil receiver should return default")
	}
	if config.GetTimeoutOr(30) != 30 {
		panic("GetTimeoutOr on nil receiver should return default")
	}
}

func testNilOptionalFields() {
	config := &gen.UserConfig{
		Host: "localhost",
		// Timeout and Advanced are nil (not set)
	}

	// Required field returns value
	if config.GetHost() != "localhost" {
		panic("GetHost should return 'localhost'")
	}

	// Optional nil fields return zero values
	if config.GetTimeout() != 0 {
		panic("GetTimeout on nil field should return 0")
	}
	if config.GetAdvanced() != (gen.AdvancedConfig{}) {
		panic("GetAdvanced on nil field should return zero AdvancedConfig")
	}
}

func testPopulatedStruct() {
	config := &gen.UserConfig{
		Host:    "example.com",
		Timeout: gen.Ptr(int64(60)),
		Advanced: gen.Ptr(gen.AdvancedConfig{
			MaxRetries: 3,
			Debug:      gen.Ptr(true),
		}),
	}

	if config.GetHost() != "example.com" {
		panic("GetHost should return 'example.com'")
	}
	if config.GetTimeout() != 60 {
		panic("GetTimeout should return 60")
	}
	if config.GetAdvanced().MaxRetries != 3 {
		panic("GetAdvanced().MaxRetries should be 3")
	}
	if !*config.Advanced.Debug {
		panic("Advanced.Debug should be true")
	}
}

func testGetterWithDefaults() {
	config := &gen.UserConfig{
		Host: "localhost",
	}

	// GetFieldOr returns default when optional field is nil
	if config.GetTimeoutOr(42) != 42 {
		panic("GetTimeoutOr should return default 42")
	}

	// GetFieldOr returns value when field is set
	config.Timeout = gen.Ptr(int64(100))
	if config.GetTimeoutOr(42) != 100 {
		panic("GetTimeoutOr should return actual value 100")
	}
}

func testJSONMarshal() {
	// Struct with nil optional field should omit the field
	config := gen.UserConfig{
		Host: "localhost",
		// Timeout is nil, should be omitted
	}

	data, err := json.Marshal(config)
	if err != nil {
		panic(fmt.Sprintf("JSON marshal failed: %v", err))
	}

	jsonStr := string(data)
	// Should NOT contain "timeout" field
	if contains(jsonStr, "timeout") {
		panic(fmt.Sprintf("Optional nil field should be omitted: %s", jsonStr))
	}

	// Struct with set optional field should include it
	config.Timeout = gen.Ptr(int64(30))
	data, err = json.Marshal(config)
	if err != nil {
		panic(fmt.Sprintf("JSON marshal failed: %v", err))
	}

	jsonStr = string(data)
	if !contains(jsonStr, `"timeout":30`) {
		panic(fmt.Sprintf("Optional set field should be present: %s", jsonStr))
	}
}

func testJSONUnmarshal() {
	// Unmarshal with missing optional field
	jsonStr := `{"host":"localhost"}`
	var config gen.UserConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		panic(fmt.Sprintf("JSON unmarshal failed: %v", err))
	}
	if config.Host != "localhost" {
		panic("Host should be 'localhost'")
	}
	if config.Timeout != nil {
		panic("Timeout should be nil when omitted")
	}

	// Unmarshal with null optional field
	jsonStr = `{"host":"localhost","timeout":null}`
	config = gen.UserConfig{}
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		panic(fmt.Sprintf("JSON unmarshal failed: %v", err))
	}
	if config.Timeout != nil {
		panic("Timeout should be nil when null")
	}

	// Unmarshal with value optional field
	jsonStr = `{"host":"localhost","timeout":45}`
	config = gen.UserConfig{}
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		panic(fmt.Sprintf("JSON unmarshal failed: %v", err))
	}
	if config.Timeout == nil || *config.Timeout != 45 {
		panic("Timeout should be 45")
	}
}

func testPointerUtilities() {
	// Test Ptr
	ptr := gen.Ptr(42)
	if *ptr != 42 {
		panic("Ptr should create pointer to value")
	}

	// Test Val
	if gen.Val(ptr) != 42 {
		panic("Val should dereference pointer")
	}
	var nilPtr *int
	if gen.Val(nilPtr) != 0 {
		panic("Val should return zero for nil pointer")
	}

	// Test Or
	if gen.Or(ptr, 100) != 42 {
		panic("Or should return value when pointer is not nil")
	}
	if gen.Or(nilPtr, 100) != 100 {
		panic("Or should return default when pointer is nil")
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
