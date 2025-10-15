package descriptor

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestGenerateExampleJSON(t *testing.T) {
	tests := []struct {
		name     string
		msgName  string
		options  ExampleOptions
		expected string
	}{
		{
			name:    "basic message with defaults",
			msgName: "echo.v1.EchoRequest",
			options: DefaultExampleOptions(),
		},
		{
			name:    "minimal mode",
			msgName: "echo.v1.EchoRequest",
			options: ExampleOptions{
				MinimalMode:     true,
				IncludeOptional: false,
				MaxDepth:        3,
			},
		},
		{
			name:    "well-known timestamp",
			msgName: "google.protobuf.Timestamp",
			options: DefaultExampleOptions(),
		},
		{
			name:    "well-known duration",
			msgName: "google.protobuf.Duration",
			options: DefaultExampleOptions(),
		},
		{
			name:    "well-known any",
			msgName: "google.protobuf.Any",
			options: DefaultExampleOptions(),
		},
		{
			name:    "well-known struct",
			msgName: "google.protobuf.Struct",
			options: DefaultExampleOptions(),
		},
		{
			name:    "well-known field mask",
			msgName: "google.protobuf.FieldMask",
			options: DefaultExampleOptions(),
		},
	}

	// Load test registry
	registry, err := LoadDirectory(context.Background(), "testdata/basic", nil)
	if err != nil {
		t.Fatalf("Failed to load test registry: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, exists := registry.FindMessage(tt.msgName)
			if !exists {
				t.Skipf("Message %s not found in test registry", tt.msgName)
			}

			result, err := GenerateExampleJSON(msg, tt.options)
			if err != nil {
				t.Fatalf("GenerateExampleJSON() error = %v", err)
			}

			if result == "" {
				t.Error("GenerateExampleJSON() returned empty string")
			}

			// Basic validation - should be valid JSON
			var jsonData interface{}
			if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
				t.Errorf("Generated JSON is invalid: %v\nJSON: %s", err, result)
			}

			t.Logf("Generated JSON for %s:\n%s", tt.msgName, result)
		})
	}
}

func TestGenerateExampleJSON_ComplexMessage(t *testing.T) {
	// Load comprehensive test registry
	registry, err := LoadDirectory(context.Background(), "testdata/comprehensive", nil)
	if err != nil {
		t.Fatalf("Failed to load comprehensive test registry: %v", err)
	}

	// Test a complex message with nested types
	msgName := "users.v1.User"
	msg, exists := registry.FindMessage(msgName)
	if !exists {
		t.Skipf("Message %s not found in comprehensive test registry", msgName)
	}

	options := DefaultExampleOptions()
	result, err := GenerateExampleJSON(msg, options)
	if err != nil {
		t.Fatalf("GenerateExampleJSON() error = %v", err)
	}

	if result == "" {
		t.Error("GenerateExampleJSON() returned empty string")
	}

	// Validate JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
		t.Errorf("Generated JSON is invalid: %v\nJSON: %s", err, result)
	}

	t.Logf("Generated JSON for complex message %s:\n%s", msgName, result)
}

func TestGenerateExampleJSON_RepeatedFields(t *testing.T) {
	registry, err := LoadDirectory(context.Background(), "testdata/comprehensive", nil)
	if err != nil {
		t.Fatalf("Failed to load test registry: %v", err)
	}

	// Find a message with repeated fields
	msgName := "users.v1.ListUsersResponse"
	msg, exists := registry.FindMessage(msgName)
	if !exists {
		t.Skipf("Message %s not found in test registry", msgName)
	}

	options := DefaultExampleOptions()
	result, err := GenerateExampleJSON(msg, options)
	if err != nil {
		t.Fatalf("GenerateExampleJSON() error = %v", err)
	}

	// Should contain arrays for repeated fields
	if !strings.Contains(result, "[") {
		t.Error("Expected repeated fields to generate arrays")
	}

	t.Logf("Generated JSON for message with repeated fields:\n%s", result)
}

func TestGenerateExampleJSON_EnumFields(t *testing.T) {
	registry, err := LoadDirectory(context.Background(), "testdata/comprehensive", nil)
	if err != nil {
		t.Fatalf("Failed to load test registry: %v", err)
	}

	// Find a message with enum fields
	msgName := "users.v1.User"
	msg, exists := registry.FindMessage(msgName)
	if !exists {
		t.Skipf("Message %s not found in test registry", msgName)
	}

	options := DefaultExampleOptions()
	result, err := GenerateExampleJSON(msg, options)
	if err != nil {
		t.Fatalf("GenerateExampleJSON() error = %v", err)
	}

	// Should contain enum values as strings
	if !strings.Contains(result, "USER_ROLE") && !strings.Contains(result, "ADMIN") {
		t.Error("Expected enum fields to generate string values")
	}

	t.Logf("Generated JSON for message with enum fields:\n%s", result)
}

func TestGenerateExampleJSON_MaxDepth(t *testing.T) {
	registry, err := LoadDirectory(context.Background(), "testdata/comprehensive", nil)
	if err != nil {
		t.Fatalf("Failed to load test registry: %v", err)
	}

	msgName := "users.v1.User"
	msg, exists := registry.FindMessage(msgName)
	if !exists {
		t.Skipf("Message %s not found in test registry", msgName)
	}

	// Test with very low max depth
	options := ExampleOptions{
		MaxDepth: 1,
	}
	result, err := GenerateExampleJSON(msg, options)
	if err != nil {
		t.Fatalf("GenerateExampleJSON() error = %v", err)
	}

	// Should contain max depth indicator
	if !strings.Contains(result, "max_depth_reached") {
		t.Error("Expected max depth to be enforced")
	}

	t.Logf("Generated JSON with max depth limit:\n%s", result)
}

func TestGenerateExampleJSON_MinimalMode(t *testing.T) {
	registry, err := LoadDirectory(context.Background(), "testdata/basic", nil)
	if err != nil {
		t.Fatalf("Failed to load test registry: %v", err)
	}

	msgName := "echo.v1.EchoRequest"
	msg, exists := registry.FindMessage(msgName)
	if !exists {
		t.Skipf("Message %s not found in test registry", msgName)
	}

	// Test minimal mode
	options := ExampleOptions{
		MinimalMode:     true,
		IncludeOptional: false,
	}
	result, err := GenerateExampleJSON(msg, options)
	if err != nil {
		t.Fatalf("GenerateExampleJSON() error = %v", err)
	}

	// In minimal mode, should only include required fields
	// For proto3, this typically means no fields (since proto3 has no required fields)
	t.Logf("Generated JSON in minimal mode:\n%s", result)
}

func TestGenerateExampleJSON_NilMessage(t *testing.T) {
	_, err := GenerateExampleJSON(nil, DefaultExampleOptions())
	if err == nil {
		t.Error("Expected error for nil message descriptor")
	}
}

func TestDefaultExampleOptions(t *testing.T) {
	options := DefaultExampleOptions()

	if !options.IncludeOptional {
		t.Error("Expected IncludeOptional to be true by default")
	}

	if options.IncludeComments {
		t.Error("Expected IncludeComments to be false by default")
	}

	if options.MaxDepth != 5 {
		t.Error("Expected MaxDepth to be 5 by default")
	}

	if options.MinimalMode {
		t.Error("Expected MinimalMode to be false by default")
	}
}

func TestGenerateWellKnownType(t *testing.T) {
	// Test well-known types by loading them from the registry
	registry, err := LoadDirectory(context.Background(), "testdata/wkt", nil)
	if err != nil {
		t.Skipf("Failed to load WKT test registry: %v", err)
	}

	// Test timestamp message
	msgName := "testdata.wkt.UsesTimestamp"
	msg, exists := registry.FindMessage(msgName)
	if !exists {
		t.Skipf("Message %s not found in WKT test registry", msgName)
	}

	options := DefaultExampleOptions()
	result, err := GenerateExampleJSON(msg, options)
	if err != nil {
		t.Fatalf("GenerateExampleJSON() error = %v", err)
	}

	// Should contain timestamp fields
	if !strings.Contains(result, "seconds") || !strings.Contains(result, "nanos") {
		t.Error("Expected timestamp fields in generated JSON")
	}

	t.Logf("Generated JSON for WKT message:\n%s", result)
}
