package descriptor

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Registry holds parsed protobuf descriptors with fast lookup capabilities.
type Registry struct {
	Files *protoregistry.Files
	Types *protoregistry.Types
	// FileDescriptorSet for comment extraction
	FileDescriptorSet *descriptorpb.FileDescriptorSet
	// Comment index for documentation
	CommentIndex map[string]string
	// Fast lookups by fully-qualified name
	ServicesByName map[string]protoreflect.ServiceDescriptor
	MethodsByName  map[string]protoreflect.MethodDescriptor
	MessagesByName map[string]protoreflect.MessageDescriptor
	EnumsByName    map[string]protoreflect.EnumDescriptor
}

// FindService returns a service descriptor by its fully-qualified name.
func (r *Registry) FindService(fullName string) (protoreflect.ServiceDescriptor, bool) {
	service, exists := r.ServicesByName[fullName]
	return service, exists
}

// FindMethod returns a method descriptor by its fully-qualified name.
// Method names use the format "pkg.Service/Method".
func (r *Registry) FindMethod(fullName string) (protoreflect.MethodDescriptor, bool) {
	method, exists := r.MethodsByName[fullName]
	return method, exists
}

// FindMessage returns a message descriptor by its fully-qualified name.
func (r *Registry) FindMessage(fullName string) (protoreflect.MessageDescriptor, bool) {
	message, exists := r.MessagesByName[fullName]
	return message, exists
}

// FindEnum returns an enum descriptor by its fully-qualified name.
func (r *Registry) FindEnum(fullName string) (protoreflect.EnumDescriptor, bool) {
	enum, exists := r.EnumsByName[fullName]
	return enum, exists
}

// buildRegistry creates a Registry from parsed files.
func buildRegistry(files *protoregistry.Files, fdSet *descriptorpb.FileDescriptorSet) (*Registry, error) {
	registry := &Registry{
		Files:             files,
		Types:             &protoregistry.Types{},
		FileDescriptorSet: fdSet,
		CommentIndex:      make(map[string]string),
		ServicesByName:    make(map[string]protoreflect.ServiceDescriptor),
		MethodsByName:     make(map[string]protoreflect.MethodDescriptor),
		MessagesByName:    make(map[string]protoreflect.MessageDescriptor),
		EnumsByName:       make(map[string]protoreflect.EnumDescriptor),
	}

	// Iterate through all files to build indexes
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		// Index services and methods
		for i := 0; i < fd.Services().Len(); i++ {
			service := fd.Services().Get(i)
			serviceName := string(service.FullName())
			registry.ServicesByName[serviceName] = service

			// Index methods
			for j := 0; j < service.Methods().Len(); j++ {
				method := service.Methods().Get(j)
				methodName := fmt.Sprintf("%s/%s", serviceName, method.Name())
				registry.MethodsByName[methodName] = method
			}
		}

		// Index messages
		indexMessages(fd.Messages(), registry)

		// Index enums
		indexEnums(fd.Enums(), registry)

		return true
	})

	// Build comment index
	buildCommentIndex(fdSet, registry)

	return registry, nil
}

// indexMessages recursively indexes all message types.
func indexMessages(messages protoreflect.MessageDescriptors, registry *Registry) {
	for i := 0; i < messages.Len(); i++ {
		msg := messages.Get(i)
		msgName := string(msg.FullName())
		registry.MessagesByName[msgName] = msg

		// Recursively index nested messages
		indexMessages(msg.Messages(), registry)
		// Index nested enums
		indexEnums(msg.Enums(), registry)
	}
}

// indexEnums recursively indexes all enum types.
func indexEnums(enums protoreflect.EnumDescriptors, registry *Registry) {
	for i := 0; i < enums.Len(); i++ {
		enum := enums.Get(i)
		enumName := string(enum.FullName())
		registry.EnumsByName[enumName] = enum
	}
}

// buildCommentIndex extracts comments from FileDescriptorSet and indexes them by FQN.
func buildCommentIndex(fdSet *descriptorpb.FileDescriptorSet, registry *Registry) {
	for _, file := range fdSet.File {
		if file.SourceCodeInfo == nil {
			continue
		}

		// Extract comments for services
		for i, service := range file.Service {
			servicePath := []int32{6, int32(i)} // 6 = service
			comment := extractComment(file.SourceCodeInfo, servicePath)
			if comment != "" {
				// Use full name instead of just name
				serviceFullName := fmt.Sprintf("%s.%s", file.GetPackage(), *service.Name)
				registry.CommentIndex[serviceFullName] = comment
			}

			// Extract comments for methods
			for j, method := range service.Method {
				methodPath := []int32{6, int32(i), 2, int32(j)} // 6 = service, 2 = method
				comment := extractComment(file.SourceCodeInfo, methodPath)
				if comment != "" {
					// Use full name format
					methodName := fmt.Sprintf("%s.%s/%s", file.GetPackage(), *service.Name, *method.Name)
					registry.CommentIndex[methodName] = comment
				}
			}
		}

		// Extract comments for messages
		for i, message := range file.MessageType {
			extractMessageComments(file.SourceCodeInfo, message, registry, []int32{4, int32(i)}, file.GetPackage())
		}

		// Extract comments for enums
		for i, enum := range file.EnumType {
			extractEnumComments(file.SourceCodeInfo, enum, registry, []int32{5, int32(i)}, file.GetPackage())
		}
	}
}

// extractMessageComments recursively extracts comments from message types and their fields.
func extractMessageComments(sourceInfo *descriptorpb.SourceCodeInfo, message *descriptorpb.DescriptorProto, registry *Registry, path []int32, packageName string) {
	// Extract comment for the message itself
	comment := extractComment(sourceInfo, path)
	if comment != "" {
		// Use full name
		messageFullName := fmt.Sprintf("%s.%s", packageName, *message.Name)
		registry.CommentIndex[messageFullName] = comment
	}

	// Extract comments for fields
	for i, field := range message.Field {
		fieldPath := append(path, 2, int32(i)) // 2 = field
		comment := extractComment(sourceInfo, fieldPath)
		if comment != "" {
			// Use full name
			fieldName := fmt.Sprintf("%s.%s.%s", packageName, *message.Name, *field.Name)
			registry.CommentIndex[fieldName] = comment
		}
	}

	// Extract comments for nested messages
	for i, nested := range message.NestedType {
		nestedPath := append(path, 3, int32(i)) // 3 = nested_type
		extractMessageComments(sourceInfo, nested, registry, nestedPath, packageName)
	}

	// Extract comments for nested enums
	for i, nested := range message.EnumType {
		nestedPath := append(path, 4, int32(i)) // 4 = enum_type
		extractEnumComments(sourceInfo, nested, registry, nestedPath, packageName)
	}
}

// extractEnumComments recursively extracts comments from enum types and their values.
func extractEnumComments(sourceInfo *descriptorpb.SourceCodeInfo, enum *descriptorpb.EnumDescriptorProto, registry *Registry, path []int32, packageName string) {
	// Extract comment for the enum itself
	comment := extractComment(sourceInfo, path)
	if comment != "" {
		// Use full name
		enumFullName := fmt.Sprintf("%s.%s", packageName, *enum.Name)
		registry.CommentIndex[enumFullName] = comment
	}

	// Extract comments for enum values
	for i, value := range enum.Value {
		valuePath := append(path, 2, int32(i)) // 2 = value
		comment := extractComment(sourceInfo, valuePath)
		if comment != "" {
			// Use full name
			valueName := fmt.Sprintf("%s.%s.%s", packageName, *enum.Name, *value.Name)
			registry.CommentIndex[valueName] = comment
		}
	}
}

// extractComment extracts the leading comment from SourceCodeInfo for a given path.
func extractComment(sourceInfo *descriptorpb.SourceCodeInfo, path []int32) string {
	if sourceInfo == nil {
		return ""
	}

	for _, location := range sourceInfo.Location {
		if pathEqual(location.Path, path) {
			if location.LeadingComments != nil && *location.LeadingComments != "" {
				return *location.LeadingComments
			}
		}
	}
	return ""
}

// pathEqual compares two path slices for equality.
func pathEqual(a, b []int32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
