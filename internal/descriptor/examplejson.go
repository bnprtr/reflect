package descriptor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// ExampleOptions configures how examples are generated.
type ExampleOptions struct {
	IncludeOptional bool // Whether to include optional fields (default: true)
	IncludeComments bool // Whether to include field comments as JSON comments (default: false)
	MaxDepth        int  // Maximum recursion depth to prevent cycles (default: 5)
	MinimalMode     bool // Only include required fields (default: false)
}

// DefaultExampleOptions returns sensible defaults for example generation.
func DefaultExampleOptions() ExampleOptions {
	return ExampleOptions{
		IncludeOptional: true,
		IncludeComments: false,
		MaxDepth:        5,
		MinimalMode:     false,
	}
}

// GenerateExampleJSON generates a formatted JSON example for a message type.
func GenerateExampleJSON(msg protoreflect.MessageDescriptor, options ExampleOptions) (string, error) {
	if msg == nil {
		return "", fmt.Errorf("message descriptor is nil")
	}

	// Set defaults for unset options
	if options.MaxDepth == 0 {
		options.MaxDepth = 5
	}

	visited := make(map[string]bool)
	value, err := generateMessageValue(msg, options, visited, 0)
	if err != nil {
		return "", fmt.Errorf("failed to generate message value: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// generateMessageValue generates example values for a message type.
func generateMessageValue(msg protoreflect.MessageDescriptor, options ExampleOptions, visited map[string]bool, depth int) (map[string]any, error) {
	if depth >= options.MaxDepth {
		return map[string]any{"<max_depth_reached>": true}, nil
	}

	msgName := string(msg.FullName())
	if visited[msgName] {
		return map[string]any{"<recursive>": true}, nil
	}

	visited[msgName] = true
	defer delete(visited, msgName)

	result := make(map[string]any)

	// Handle well-known types specially
	if wktValue := generateWellKnownType(msg); wktValue != nil {
		return wktValue, nil
	}

	// Generate fields
	for i := 0; i < msg.Fields().Len(); i++ {
		field := msg.Fields().Get(i)

		// Skip fields based on options
		if !shouldIncludeField(field, options) {
			continue
		}

		fieldValue, err := generateFieldValue(field, options, visited, depth)
		if err != nil {
			return nil, fmt.Errorf("failed to generate value for field %s: %w", field.Name(), err)
		}

		// Skip nil values unless they're explicitly set
		if fieldValue != nil {
			result[string(field.JSONName())] = fieldValue
		}
	}

	return result, nil
}

// generateFieldValue generates an appropriate value for a field based on its type.
func generateFieldValue(field protoreflect.FieldDescriptor, options ExampleOptions, visited map[string]bool, depth int) (any, error) {
	switch {
	case field.IsMap():
		return generateMapValue(field, options, visited, depth)
	case field.Cardinality() == protoreflect.Repeated:
		return generateRepeatedValue(field, options, visited, depth)
	case field.ContainingOneof() != nil:
		return generateOneofValue(field, options, visited, depth)
	default:
		return generateScalarValue(field, options, visited, depth)
	}
}

// generateScalarValue generates a value for a scalar field.
func generateScalarValue(field protoreflect.FieldDescriptor, options ExampleOptions, visited map[string]bool, depth int) (any, error) {
	switch field.Kind() {
	case protoreflect.BoolKind:
		return true, nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return 42, nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return int64(42), nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return uint32(42), nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return uint64(42), nil
	case protoreflect.FloatKind:
		return float32(3.14), nil
	case protoreflect.DoubleKind:
		return 3.14, nil
	case protoreflect.StringKind:
		return fmt.Sprintf("example_%s", field.Name()), nil
	case protoreflect.BytesKind:
		return base64.StdEncoding.EncodeToString([]byte("example data")), nil
	case protoreflect.EnumKind:
		return generateEnumValue(field.Enum())
	case protoreflect.MessageKind:
		return generateMessageValue(field.Message(), options, visited, depth+1)
	default:
		return nil, fmt.Errorf("unsupported field kind: %v", field.Kind())
	}
}

// generateRepeatedValue generates an array value for a repeated field.
func generateRepeatedValue(field protoreflect.FieldDescriptor, options ExampleOptions, visited map[string]bool, depth int) (any, error) {
	// Generate 1-2 example items
	itemCount := 2
	if field.Kind() == protoreflect.MessageKind {
		// For complex message types, just generate 1 item
		itemCount = 1
	}

	result := make([]any, 0, itemCount)
	for i := 0; i < itemCount; i++ {
		itemValue, err := generateScalarValue(field, options, visited, depth)
		if err != nil {
			return nil, err
		}
		result = append(result, itemValue)
	}

	return result, nil
}

// generateMapValue generates a map value for a map field.
func generateMapValue(field protoreflect.FieldDescriptor, options ExampleOptions, visited map[string]bool, depth int) (any, error) {
	// Generate 1-2 example key-value pairs
	result := make(map[string]any)

	keyField := field.MapKey()
	valueField := field.MapValue()

	// Generate example keys and values
	for i := 0; i < 2; i++ {
		keyValue, err := generateScalarValue(keyField, options, visited, depth)
		if err != nil {
			return nil, err
		}

		valueValue, err := generateScalarValue(valueField, options, visited, depth)
		if err != nil {
			return nil, err
		}

		// Convert key to string for JSON
		keyStr := fmt.Sprintf("%v", keyValue)
		if i > 0 {
			keyStr = fmt.Sprintf("%s_%d", keyStr, i)
		}

		result[keyStr] = valueValue
	}

	return result, nil
}

// generateOneofValue generates a value for a oneof field.
func generateOneofValue(field protoreflect.FieldDescriptor, options ExampleOptions, visited map[string]bool, depth int) (any, error) {
	// Generate value for the first field in the oneof
	oneof := field.ContainingOneof()
	if oneof.Fields().Len() == 0 {
		return nil, nil
	}

	firstField := oneof.Fields().Get(0)
	return generateScalarValue(firstField, options, visited, depth)
}

// generateEnumValue generates an example value for an enum.
func generateEnumValue(enum protoreflect.EnumDescriptor) (any, error) {
	// Try to find the first non-zero value, otherwise use zero
	for i := 0; i < enum.Values().Len(); i++ {
		value := enum.Values().Get(i)
		if value.Number() != 0 {
			return string(value.Name()), nil
		}
	}

	// If no non-zero value found, use the first value
	if enum.Values().Len() > 0 {
		return string(enum.Values().Get(0).Name()), nil
	}

	return "UNKNOWN", nil
}

// generateWellKnownType generates examples for well-known protobuf types.
func generateWellKnownType(msg protoreflect.MessageDescriptor) map[string]any {
	msgName := string(msg.FullName())

	switch msgName {
	case "google.protobuf.Timestamp":
		// Generate a deterministic timestamp for testing
		return map[string]any{
			"seconds": int64(1640995200), // 2022-01-01 00:00:00 UTC
			"nanos":   0,
		}
	case "google.protobuf.Duration":
		return map[string]any{
			"seconds": 300,
			"nanos":   0,
		}
	case "google.protobuf.Any":
		return map[string]any{
			"@type": "type.googleapis.com/google.protobuf.StringValue",
			"value": "example value",
		}
	case "google.protobuf.Struct":
		return map[string]any{
			"fields": map[string]any{
				"example_key": map[string]any{
					"string_value": "example value",
				},
			},
		}
	case "google.protobuf.FieldMask":
		return map[string]any{
			"paths": []string{"field1", "field2.nested"},
		}
	case "google.protobuf.StringValue":
		return map[string]any{
			"value": "example string",
		}
	case "google.protobuf.Int32Value":
		return map[string]any{
			"value": 42,
		}
	case "google.protobuf.Int64Value":
		return map[string]any{
			"value": int64(42),
		}
	case "google.protobuf.UInt32Value":
		return map[string]any{
			"value": uint32(42),
		}
	case "google.protobuf.UInt64Value":
		return map[string]any{
			"value": uint64(42),
		}
	case "google.protobuf.FloatValue":
		return map[string]any{
			"value": float32(3.14),
		}
	case "google.protobuf.DoubleValue":
		return map[string]any{
			"value": 3.14,
		}
	case "google.protobuf.BoolValue":
		return map[string]any{
			"value": true,
		}
	case "google.protobuf.BytesValue":
		return map[string]any{
			"value": base64.StdEncoding.EncodeToString([]byte("example data")),
		}
	case "google.protobuf.Empty":
		return map[string]any{}
	case "google.protobuf.NullValue":
		return nil
	}

	return nil
}

// shouldIncludeField determines whether a field should be included in the example.
func shouldIncludeField(field protoreflect.FieldDescriptor, options ExampleOptions) bool {
	// In minimal mode, only include required fields
	if options.MinimalMode {
		return field.Cardinality() == protoreflect.Required
	}

	// Skip optional fields if not including them
	if !options.IncludeOptional && field.HasOptionalKeyword() {
		return false
	}

	return true
}
