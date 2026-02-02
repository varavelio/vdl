package main

import (
	"fmt"
	"os"
	"time"

	"test/gen"
)

func main() {
	verifyConstants()
	verifyEnums()
	verifyPatterns()
	verifyTypes()
	verifyRPCs()
	fmt.Println("Full syntax verification successful")
}

func verifyConstants() {
	if gen.MAX_PAGE_SIZE != 100 {
		fail("MAX_PAGE_SIZE", "100", fmt.Sprintf("%d", gen.MAX_PAGE_SIZE))
	}
	if gen.API_VERSION != "1.0.0" {
		fail("API_VERSION", "1.0.0", gen.API_VERSION)
	}
}

func verifyEnums() {
	// String enum
	s := gen.StatusActive
	if s.String() != "Active" {
		fail("StatusActive.String()", "Active", s.String())
	}
	if !s.IsValid() {
		fail("StatusActive.IsValid()", "true", "false")
	}

	// Int enum
	p := gen.PriorityHigh
	if int(p) != 3 {
		fail("PriorityHigh", "3", fmt.Sprintf("%d", p))
	}
}

func verifyPatterns() {
	topic := gen.UserTopic("123", "login")
	expected := "events.users.123.login"
	if topic != expected {
		fail("UserTopic pattern", expected, topic)
	}
}

func verifyTypes() {
	// Verify struct fields and embedding/spreading
	_ = gen.User{
		Id:        "u1",
		CreatedAt: time.Now(), // From Meta
		Username:  "alice",
		Roles:     []string{"admin"},
		Address: gen.UserAddress{ // Inline type
			Street: "123 Main St",
			City:   "Tech City",
		},
	}

	// Verify optional fields
	u := gen.User{}
	if u.Bio != nil {
		fail("User.Bio optional default", "nil", "not nil")
	}
}

func verifyRPCs() {
	// Verify RPC Catalog
	path := gen.VDLPaths.UserService.GetUser
	expectedPath := "/UserService/GetUser"
	if path != expectedPath {
		fail("VDLPaths.UserService.GetUser", expectedPath, path)
	}

	// Verify Procedure Input/Output types
	_ = gen.UserServiceGetUserInput{Id: "1"}
	_ = gen.UserServiceGetUserOutput{
		User: gen.User{Username: "bob"},
	}

	// Verify Stream Input/Output types
	_ = gen.UserServiceUserActivityInput{UserId: "1"}
	_ = gen.UserServiceUserActivityOutput{Action: "click"}
}

func fail(name, expected, actual string) {
	fmt.Fprintf(os.Stderr, "Verification failed for %s: expected %q, got %q\n", name, expected, actual)
	os.Exit(1)
}
