package descriptor

import (
	"context"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestRegistryLookups(t *testing.T) {
	ctx := context.Background()
	testDataDir := "testdata"

	// Load the basic test data
	reg, err := LoadDirectory(ctx, filepath.Join(testDataDir, "basic"), []string{})
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	tests := []struct {
		name     string
		fullName string
		lookup   func(string) (protoreflect.Descriptor, bool)
		wantName string
	}{
		{
			name:     "find service",
			fullName: "echo.v1.EchoService",
			lookup: func(name string) (protoreflect.Descriptor, bool) {
				desc, ok := reg.FindService(name)
				return desc, ok
			},
			wantName: "EchoService",
		},
		{
			name:     "find method",
			fullName: "echo.v1.EchoService/Echo",
			lookup: func(name string) (protoreflect.Descriptor, bool) {
				desc, ok := reg.FindMethod(name)
				return desc, ok
			},
			wantName: "Echo",
		},
		{
			name:     "find message",
			fullName: "echo.v1.EchoRequest",
			lookup: func(name string) (protoreflect.Descriptor, bool) {
				desc, ok := reg.FindMessage(name)
				return desc, ok
			},
			wantName: "EchoRequest",
		},
		{
			name:     "find enum",
			fullName: "echo.v1.Status",
			lookup: func(name string) (protoreflect.Descriptor, bool) {
				desc, ok := reg.FindEnum(name)
				return desc, ok
			},
			wantName: "Status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, exists := tt.lookup(tt.fullName)
			if !exists {
				t.Fatalf("Descriptor %q not found", tt.fullName)
			}
			if desc.Name() != protoreflect.Name(tt.wantName) {
				t.Errorf("Expected name %q, got %q", tt.wantName, desc.Name())
			}
		})
	}
}

func TestRegistryNotFound(t *testing.T) {
	ctx := context.Background()
	testDataDir := "testdata"

	reg, err := LoadDirectory(ctx, filepath.Join(testDataDir, "basic"), []string{})
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	// Test lookups that should fail
	notFoundTests := []struct {
		name   string
		lookup func(string) bool
	}{
		{
			name: "non-existent service",
			lookup: func(name string) bool {
				_, exists := reg.FindService(name)
				return exists
			},
		},
		{
			name: "non-existent method",
			lookup: func(name string) bool {
				_, exists := reg.FindMethod(name)
				return exists
			},
		},
		{
			name: "non-existent message",
			lookup: func(name string) bool {
				_, exists := reg.FindMessage(name)
				return exists
			},
		},
		{
			name: "non-existent enum",
			lookup: func(name string) bool {
				_, exists := reg.FindEnum(name)
				return exists
			},
		},
	}

	for _, tt := range notFoundTests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.lookup("non.existent.Name") {
				t.Error("Expected lookup to return false for non-existent descriptor")
			}
		})
	}
}

func TestRegistryWithImports(t *testing.T) {
	ctx := context.Background()
	testDataDir := "testdata"

	// Load the import test data
	reg, err := LoadDirectory(ctx, filepath.Join(testDataDir, "import"), []string{filepath.Join(testDataDir, "import")})
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	// Check that we can find descriptors from both files
	tests := []struct {
		name     string
		fullName string
		lookup   func(string) (protoreflect.Descriptor, bool)
	}{
		{
			name:     "import service",
			fullName: "import.v1.ImportEchoService",
			lookup: func(name string) (protoreflect.Descriptor, bool) {
				desc, ok := reg.FindService(name)
				return desc, ok
			},
		},
		{
			name:     "shared message",
			fullName: "shared.v1.CommonMessage",
			lookup: func(name string) (protoreflect.Descriptor, bool) {
				desc, ok := reg.FindMessage(name)
				return desc, ok
			},
		},
		{
			name:     "shared error details",
			fullName: "shared.v1.ErrorDetails",
			lookup: func(name string) (protoreflect.Descriptor, bool) {
				desc, ok := reg.FindMessage(name)
				return desc, ok
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, exists := tt.lookup(tt.fullName)
			if !exists {
				t.Fatalf("Descriptor %q not found", tt.fullName)
			}
			if desc == nil {
				t.Fatal("Descriptor is nil")
			}
		})
	}
}

func TestRegistryWithWKTs(t *testing.T) {
	ctx := context.Background()
	testDataDir := "testdata"

	// Load the WKT test data
	reg, err := LoadDirectory(ctx, filepath.Join(testDataDir, "wkt"), []string{})
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	// Check that we can find the service and messages
	_, exists := reg.FindService("wkt.v1.TimestampService")
	if !exists {
		t.Fatal("TimestampService not found")
	}

	// Check that the service has the expected methods
	methods := []string{"GetCurrentTime", "CalculateDuration"}
	for _, methodName := range methods {
		fullMethodName := "wkt.v1.TimestampService/" + methodName
		method, exists := reg.FindMethod(fullMethodName)
		if !exists {
			t.Fatalf("Method %q not found", fullMethodName)
		}
		if method.Name() != protoreflect.Name(methodName) {
			t.Errorf("Expected method name %q, got %q", methodName, method.Name())
		}
	}

	// Check that we can find the messages
	messages := []string{"GetCurrentTimeRequest", "GetCurrentTimeResponse", "CalculateDurationRequest", "CalculateDurationResponse"}
	for _, msgName := range messages {
		fullMsgName := "wkt.v1." + msgName
		msg, exists := reg.FindMessage(fullMsgName)
		if !exists {
			t.Fatalf("Message %q not found", fullMsgName)
		}
		if msg.Name() != protoreflect.Name(msgName) {
			t.Errorf("Expected message name %q, got %q", msgName, msg.Name())
		}
	}
}
