package descriptor

import (
	"context"
	"path/filepath"
	"testing"
)

func TestLoadDirectory(t *testing.T) {
	ctx := context.Background()
	testDataDir := "testdata"

	tests := []struct {
		name          string
		root          string
		includePaths  []string
		wantError     bool
		checkRegistry func(t *testing.T, reg *Registry)
	}{
		{
			name:         "basic proto files",
			root:         filepath.Join(testDataDir, "basic"),
			includePaths: []string{},
			wantError:    false,
			checkRegistry: func(t *testing.T, reg *Registry) {
				// Check that we can find the EchoService
				_, exists := reg.FindService("echo.v1.EchoService")
				if !exists {
					t.Fatal("EchoService not found")
				}

				// Check methods
				method, exists := reg.FindMethod("echo.v1.EchoService/Echo")
				if !exists {
					t.Fatal("Echo method not found")
				}
				if method.Name() != "Echo" {
					t.Errorf("Expected method name 'Echo', got %q", method.Name())
				}

				// Check messages
				_, exists = reg.FindMessage("echo.v1.EchoRequest")
				if !exists {
					t.Fatal("EchoRequest message not found")
				}

				// Check enums
				enum, exists := reg.FindEnum("echo.v1.Status")
				if !exists {
					t.Fatal("Status enum not found")
				}
				if enum.Name() != "Status" {
					t.Errorf("Expected enum name 'Status', got %q", enum.Name())
				}
			},
		},
		{
			name:         "imported proto files",
			root:         filepath.Join(testDataDir, "import"),
			includePaths: []string{filepath.Join(testDataDir, "import")},
			wantError:    false,
			checkRegistry: func(t *testing.T, reg *Registry) {
				// Check that we can find the ImportEchoService
				_, exists := reg.FindService("import.v1.ImportEchoService")
				if !exists {
					t.Fatal("ImportEchoService not found")
				}

				// Check that shared messages are available
				_, exists = reg.FindMessage("shared.v1.CommonMessage")
				if !exists {
					t.Fatal("CommonMessage not found")
				}
			},
		},
		{
			name:         "well-known types",
			root:         filepath.Join(testDataDir, "wkt"),
			includePaths: []string{},
			wantError:    false,
			checkRegistry: func(t *testing.T, reg *Registry) {
				// Check that we can find the TimestampService
				_, exists := reg.FindService("wkt.v1.TimestampService")
				if !exists {
					t.Fatal("TimestampService not found")
				}

				// Check that WKT messages are available
				_, exists = reg.FindMessage("wkt.v1.GetCurrentTimeRequest")
				if !exists {
					t.Fatal("GetCurrentTimeRequest not found")
				}
			},
		},
		{
			name:         "non-existent directory",
			root:         "non-existent",
			includePaths: []string{},
			wantError:    true,
		},
		{
			name:         "empty root",
			root:         "",
			includePaths: []string{},
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg, err := LoadDirectory(ctx, tt.root, tt.includePaths)
			if tt.wantError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if reg == nil {
				t.Fatal("Registry is nil")
			}

			if tt.checkRegistry != nil {
				tt.checkRegistry(t, reg)
			}
		})
	}
}

func TestDiscoverProtoFiles(t *testing.T) {
	testDataDir := "testdata"

	tests := []struct {
		name      string
		root      string
		wantCount int
		wantError bool
	}{
		{
			name:      "basic directory",
			root:      filepath.Join(testDataDir, "basic"),
			wantCount: 1,
			wantError: false,
		},
		{
			name:      "import directory",
			root:      filepath.Join(testDataDir, "import"),
			wantCount: 2, // echo.proto and shared/common.proto
			wantError: false,
		},
		{
			name:      "wkt directory",
			root:      filepath.Join(testDataDir, "wkt"),
			wantCount: 1,
			wantError: false,
		},
		{
			name:      "entire testdata directory",
			root:      testDataDir,
			wantCount: 10, // All proto files including http, comprehensive/*
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := discoverProtoFiles(tt.root)
			if tt.wantError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(files) != tt.wantCount {
				t.Errorf("Expected %d files, got %d: %v", tt.wantCount, len(files), files)
			}
		})
	}
}

func TestDedupeStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single element",
			input:    []string{"a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dedupeStrings(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Expected %q at index %d, got %q", tt.expected[i], i, v)
				}
			}
		})
	}
}
