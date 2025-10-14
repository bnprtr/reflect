package descriptor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadDirectory discovers and parses all .proto files in the given root directory.
// It uses the provided includePaths for import resolution, plus the root directory itself.
func LoadDirectory(ctx context.Context, root string, includePaths []string) (*Registry, error) {
	if root == "" {
		return nil, fmt.Errorf("root directory cannot be empty")
	}

	// Check if root exists and is a directory
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("failed to stat root directory %q: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root path %q is not a directory", root)
	}

	// Discover all .proto files recursively
	protoFiles, err := discoverProtoFiles(root)
	if err != nil {
		return nil, fmt.Errorf("failed to discover proto files: %w", err)
	}

	if len(protoFiles) == 0 {
		return nil, fmt.Errorf("no .proto files found in %q", root)
	}

	// Build include paths: dedupe(append(includePaths, root))
	allIncludePaths := dedupeStrings(append(includePaths, root))

	// Parse the files
	files, fdSet, err := parseFiles(ctx, protoFiles, allIncludePaths)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proto files: %w", err)
	}

	// Build the registry
	registry, err := buildRegistry(files, fdSet)
	if err != nil {
		return nil, fmt.Errorf("failed to build registry: %w", err)
	}

	return registry, nil
}

// discoverProtoFiles recursively finds all .proto files in the given directory.
func discoverProtoFiles(root string) ([]string, error) {
	var protoFiles []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a .proto file
		if strings.HasSuffix(strings.ToLower(path), ".proto") {
			protoFiles = append(protoFiles, path)
		}

		return nil
	})

	return protoFiles, err
}

// dedupeStrings removes duplicate strings from a slice while preserving order.
func dedupeStrings(strs []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, str := range strs {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}

	return result
}
