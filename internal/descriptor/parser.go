package descriptor

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

// parseFiles parses the given proto files using protoparse with the specified include paths.
func parseFiles(ctx context.Context, protoFiles []string, includePaths []string) (*protoregistry.Files, *descriptorpb.FileDescriptorSet, error) {
	// Create the parser with include paths
	parser := protoparse.Parser{
		ImportPaths: includePaths,
		// Enable stdlib resolver for WKTs like google/protobuf/timestamp.proto
		IncludeSourceCodeInfo: true,
	}

	// Convert absolute paths to relative paths for protoparse
	var fileNames []string
	for _, file := range protoFiles {
		// Find the best include path for this file
		relPath, err := findRelativePath(file, includePaths)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to find relative path for %q: %w", file, err)
		}
		fileNames = append(fileNames, relPath)
	}

	// Parse the files
	fileDescriptors, err := parser.ParseFiles(fileNames...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse proto files: %w", err)
	}

	// Convert to FileDescriptorSet
	fdSet, err := convertToFileDescriptorSet(fileDescriptors)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert to FileDescriptorSet: %w", err)
	}

	// Create protoregistry.Files
	files, err := protodesc.NewFiles(fdSet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create protoregistry.Files: %w", err)
	}

	return files, fdSet, nil
}

// findRelativePath finds the relative path of a file given a list of include paths.
func findRelativePath(absPath string, includePaths []string) (string, error) {
	for _, includePath := range includePaths {
		relPath, err := filepath.Rel(includePath, absPath)
		if err != nil {
			continue
		}
		// Check if the relative path doesn't go outside the include path
		if !strings.HasPrefix(relPath, "..") {
			return relPath, nil
		}
	}
	return "", fmt.Errorf("file %q is not under any include path", absPath)
}

// convertToFileDescriptorSet converts a slice of FileDescriptor to a FileDescriptorSet.
func convertToFileDescriptorSet(fileDescriptors []*desc.FileDescriptor) (*descriptorpb.FileDescriptorSet, error) {
	fdSet := &descriptorpb.FileDescriptorSet{}

	// Collect all dependencies first
	allFiles := make(map[string]*desc.FileDescriptor)
	for _, fd := range fileDescriptors {
		collectDependencies(fd, allFiles)
	}

	// Add files in dependency order
	added := make(map[string]bool)
	for _, fd := range fileDescriptors {
		addFileWithDependencies(fd, fdSet, added, allFiles)
	}

	return fdSet, nil
}

// collectDependencies recursively collects all dependencies of a file descriptor.
func collectDependencies(fd *desc.FileDescriptor, allFiles map[string]*desc.FileDescriptor) {
	if _, exists := allFiles[fd.GetName()]; exists {
		return
	}
	allFiles[fd.GetName()] = fd

	for _, dep := range fd.GetDependencies() {
		collectDependencies(dep, allFiles)
	}
}

// addFileWithDependencies adds a file and its dependencies to the FileDescriptorSet in the correct order.
func addFileWithDependencies(fd *desc.FileDescriptor, fdSet *descriptorpb.FileDescriptorSet, added map[string]bool, allFiles map[string]*desc.FileDescriptor) {
	if added[fd.GetName()] {
		return
	}

	// Add dependencies first
	for _, dep := range fd.GetDependencies() {
		addFileWithDependencies(dep, fdSet, added, allFiles)
	}

	// Add this file
	fdSet.File = append(fdSet.File, fd.AsFileDescriptorProto())
	added[fd.GetName()] = true
}
