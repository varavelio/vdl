// Verifies pattern generation: pattern templates with placeholders should
// generate functions that construct the correct strings.
package main

import (
	"e2e/gen"
	"fmt"
)

func main() {
	// Test pattern with two placeholders
	result := gen.UserEventSubject("user123", "created")
	expected := "events.users.user123.created"
	if result != expected {
		panic(fmt.Sprintf("UserEventSubject: expected %q, got %q", expected, result))
	}

	// Test pattern with one placeholder
	result = gen.SessionCacheKey("sess-abc")
	expected = "cache:session:sess-abc"
	if result != expected {
		panic(fmt.Sprintf("SessionCacheKey: expected %q, got %q", expected, result))
	}

	// Test pattern with two placeholders and different separators
	result = gen.TopicChannel("chan-1", "msg-42")
	expected = "channels.chan-1.messages.msg-42"
	if result != expected {
		panic(fmt.Sprintf("TopicChannel: expected %q, got %q", expected, result))
	}

	// Test static pattern (no placeholders)
	result = gen.SimpleKey()
	expected = "static-key"
	if result != expected {
		panic(fmt.Sprintf("SimpleKey: expected %q, got %q", expected, result))
	}

	// Test with empty strings
	result = gen.UserEventSubject("", "")
	expected = "events.users.."
	if result != expected {
		panic(fmt.Sprintf("UserEventSubject with empty: expected %q, got %q", expected, result))
	}

	// Test with special characters
	result = gen.UserEventSubject("user/123", "event:type")
	expected = "events.users.user/123.event:type"
	if result != expected {
		panic(fmt.Sprintf("UserEventSubject with special chars: expected %q, got %q", expected, result))
	}

	fmt.Println("Success")
}
