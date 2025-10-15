package descriptor

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateExampleJSON_Golden(t *testing.T) {
	tests := []struct {
		name     string
		msgName  string
		options  ExampleOptions
		filename string
	}{
		{
			name:     "basic echo request",
			msgName:  "echo.v1.EchoRequest",
			options:  DefaultExampleOptions(),
			filename: "basic_echo_request.json",
		},
		{
			name:     "basic echo response",
			msgName:  "echo.v1.EchoResponse",
			options:  DefaultExampleOptions(),
			filename: "basic_echo_response.json",
		},
		{
			name:     "user message minimal",
			msgName:  "users.v1.User",
			options:  ExampleOptions{MinimalMode: true, MaxDepth: 2},
			filename: "user_minimal.json",
		},
		{
			name:     "user message full",
			msgName:  "users.v1.User",
			options:  DefaultExampleOptions(),
			filename: "user_full.json",
		},
		{
			name:     "create user request",
			msgName:  "users.v1.CreateUserRequest",
			options:  DefaultExampleOptions(),
			filename: "create_user_request.json",
		},
		{
			name:     "list users request",
			msgName:  "users.v1.ListUsersRequest",
			options:  DefaultExampleOptions(),
			filename: "list_users_request.json",
		},
		{
			name:     "timestamp message",
			msgName:  "testdata.wkt.UsesTimestamp",
			options:  DefaultExampleOptions(),
			filename: "uses_timestamp.json",
		},
	}

	// Load test registries
	basicRegistry, err := LoadDirectory(context.Background(), "testdata/basic", nil)
	if err != nil {
		t.Fatalf("Failed to load basic test registry: %v", err)
	}

	comprehensiveRegistry, err := LoadDirectory(context.Background(), "testdata/comprehensive", nil)
	if err != nil {
		t.Fatalf("Failed to load comprehensive test registry: %v", err)
	}

	wktRegistry, err := LoadDirectory(context.Background(), "testdata/wkt", nil)
	if err != nil {
		t.Fatalf("Failed to load WKT test registry: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Select appropriate registry
			var registry *Registry
			switch {
			case tt.msgName == "echo.v1.EchoRequest" || tt.msgName == "echo.v1.EchoResponse":
				registry = basicRegistry
			case tt.msgName == "testdata.wkt.UsesTimestamp":
				registry = wktRegistry
			default:
				registry = comprehensiveRegistry
			}

			msg, exists := registry.FindMessage(tt.msgName)
			if !exists {
				t.Skipf("Message %s not found in test registry", tt.msgName)
			}

			// Generate example JSON
			result, err := GenerateExampleJSON(msg, tt.options)
			if err != nil {
				t.Fatalf("GenerateExampleJSON() error = %v", err)
			}

			// Validate JSON
			var jsonData interface{}
			if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
				t.Errorf("Generated JSON is invalid: %v\nJSON: %s", err, result)
			}

			// Compare with golden file
			goldenPath := filepath.Join("testdata/golden", tt.filename)
			compareWithGolden(t, result, goldenPath)
		})
	}
}

func compareWithGolden(t *testing.T, actual, goldenPath string) {
	// Read golden file if it exists
	goldenData, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create golden file if it doesn't exist
			t.Logf("Creating golden file: %s", goldenPath)
			if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
				t.Fatalf("Failed to create golden directory: %v", err)
			}
			if err := os.WriteFile(goldenPath, []byte(actual), 0644); err != nil {
				t.Fatalf("Failed to write golden file: %v", err)
			}
			return
		}
		t.Fatalf("Failed to read golden file: %v", err)
	}

	// Compare actual with golden
	if string(goldenData) != actual {
		t.Errorf("Generated JSON does not match golden file %s", goldenPath)
		t.Logf("Expected:\n%s", string(goldenData))
		t.Logf("Actual:\n%s", actual)

		// Optionally update golden file (uncomment for development)
		// t.Logf("Updating golden file: %s", goldenPath)
		// if err := os.WriteFile(goldenPath, []byte(actual), 0644); err != nil {
		// 	t.Fatalf("Failed to update golden file: %v", err)
		// }
	}
}

func TestGenerateExampleJSON_WellKnownTypes_Golden(t *testing.T) {
	// Test well-known types by using the WKT test registry
	wktRegistry, err := LoadDirectory(context.Background(), "testdata/wkt", nil)
	if err != nil {
		t.Skipf("Failed to load WKT test registry: %v", err)
	}

	// Test timestamp message which uses well-known types
	msgName := "testdata.wkt.UsesTimestamp"
	msg, exists := wktRegistry.FindMessage(msgName)
	if !exists {
		t.Skipf("Message %s not found in WKT test registry", msgName)
	}

	options := DefaultExampleOptions()
	result, err := GenerateExampleJSON(msg, options)
	if err != nil {
		t.Fatalf("GenerateExampleJSON() error = %v", err)
	}

	// Compare with golden file
	goldenPath := filepath.Join("testdata/golden", "wkt_uses_timestamp.json")
	compareWithGolden(t, result, goldenPath)
}
