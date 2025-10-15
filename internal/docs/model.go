package docs

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bnprtr/reflect/internal/descriptor"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Index represents the main overview page with all services.
type Index struct {
	Services []ServiceSummary
}

// ServiceSummary represents a service in the index.
type ServiceSummary struct {
	Name, FullName, Package, Comment string
}

// ServiceView represents a detailed service view.
type ServiceView struct {
	Name, FullName, Package, Comment string
	Methods                          []MethodSummary
}

// HTTPRule represents a single HTTP mapping rule.
type HTTPRule struct {
	Method string
	Path   string
	Body   string
	Query  string
}

// MethodSummary represents a method in a service.
type MethodSummary struct {
	Name, FullName, Comment          string
	InputType, OutputType            string
	ClientStreaming, ServerStreaming bool
	Deprecated                       bool
	HTTPRules                        []HTTPRule
	Examples                         struct {
		Curl    string
		Grpcurl string
	}
	ExampleRequest  string
	ExampleResponse string
}

// MessageView represents a detailed message view.
type MessageView struct {
	Name, FullName, Package, Comment string
	Fields                           []FieldView
	ExampleJSON                      string
}

// FieldView represents a field in a message.
type FieldView struct {
	Name    string
	Number  int
	Type    string // resolved display (e.g., pkg.Msg, string, int32, repeated pkg.Msg)
	Label   string // repeated / optional / required (proto2)
	Oneof   string // if part of a oneof
	Comment string
}

// EnumView represents a detailed enum view.
type EnumView struct {
	Name, FullName, Package, Comment string
	Values                           []EnumValueView
}

// EnumValueView represents a value in an enum.
type EnumValueView struct {
	Name    string
	Number  int32
	Comment string
}

// BuildIndex creates an index view from the registry.
func BuildIndex(reg *descriptor.Registry) (*Index, error) {
	if reg == nil {
		return &Index{Services: []ServiceSummary{}}, nil
	}

	var services []ServiceSummary
	for _, service := range reg.ServicesByName {
		summary := ServiceSummary{
			Name:     string(service.Name()),
			FullName: string(service.FullName()),
			Package:  string(service.ParentFile().Package()),
			Comment:  reg.CommentIndex[string(service.FullName())],
		}
		services = append(services, summary)
	}

	// Sort services by full name
	sort.Slice(services, func(i, j int) bool {
		return services[i].FullName < services[j].FullName
	})

	return &Index{Services: services}, nil
}

// BuildServiceView creates a service view from the registry.
func BuildServiceView(reg *descriptor.Registry, fullName string) (*ServiceView, error) {
	if reg == nil {
		return nil, fmt.Errorf("registry is nil")
	}

	service, exists := reg.FindService(fullName)
	if !exists {
		return nil, fmt.Errorf("service %q not found", fullName)
	}

	var methods []MethodSummary
	for i := 0; i < service.Methods().Len(); i++ {
		method := service.Methods().Get(i)
		methodName := fmt.Sprintf("%s/%s", fullName, method.Name())

		summary := MethodSummary{
			Name:            string(method.Name()),
			FullName:        methodName,
			Comment:         reg.CommentIndex[methodName],
			InputType:       string(method.Input().FullName()),
			OutputType:      string(method.Output().FullName()),
			ClientStreaming: method.IsStreamingClient(),
			ServerStreaming: method.IsStreamingServer(),
			Deprecated:      false, // TODO: implement deprecated detection
		}

		// Generate example request and response JSON
		if inputMsg, exists := reg.FindMessage(string(method.Input().FullName())); exists {
			if example, err := descriptor.GenerateExampleJSON(inputMsg, descriptor.DefaultExampleOptions()); err == nil {
				summary.ExampleRequest = example
			}
		}
		if outputMsg, exists := reg.FindMessage(string(method.Output().FullName())); exists {
			if example, err := descriptor.GenerateExampleJSON(outputMsg, descriptor.DefaultExampleOptions()); err == nil {
				summary.ExampleResponse = example
			}
		}

		methods = append(methods, summary)
	}

	// Sort methods by name
	sort.Slice(methods, func(i, j int) bool {
		return methods[i].Name < methods[j].Name
	})

	return &ServiceView{
		Name:     string(service.Name()),
		FullName: fullName,
		Package:  string(service.ParentFile().Package()),
		Comment:  reg.CommentIndex[fullName],
		Methods:  methods,
	}, nil
}

// BuildMethodView creates a method view from the registry.
func BuildMethodView(reg *descriptor.Registry, fullName string) (*MethodSummary, error) {
	if reg == nil {
		return nil, fmt.Errorf("registry is nil")
	}

	method, exists := reg.FindMethod(fullName)
	if !exists {
		return nil, fmt.Errorf("method %q not found", fullName)
	}

	summary := &MethodSummary{
		Name:            string(method.Name()),
		FullName:        fullName,
		Comment:         reg.CommentIndex[fullName],
		InputType:       string(method.Input().FullName()),
		OutputType:      string(method.Output().FullName()),
		ClientStreaming: method.IsStreamingClient(),
		ServerStreaming: method.IsStreamingServer(),
		Deprecated:      false, // TODO: implement deprecated detection
	}

	// Extract HTTP rules
	httpRules, err := extractHTTPRules(method)
	if err != nil {
		// Log error but don't fail - HTTP rules are optional
		fmt.Printf("Warning: failed to extract HTTP rules for %s: %v\n", fullName, err)
	} else {
		summary.HTTPRules = httpRules
	}

	// Generate examples
	summary.Examples.Curl = generateCurlExample(summary)
	summary.Examples.Grpcurl = generateGrpcurlExample(summary)

	// Generate example request and response JSON
	if reg != nil {
		if inputMsg, exists := reg.FindMessage(string(method.Input().FullName())); exists {
			if example, err := descriptor.GenerateExampleJSON(inputMsg, descriptor.DefaultExampleOptions()); err == nil {
				summary.ExampleRequest = example
			}
		}
		if outputMsg, exists := reg.FindMessage(string(method.Output().FullName())); exists {
			if example, err := descriptor.GenerateExampleJSON(outputMsg, descriptor.DefaultExampleOptions()); err == nil {
				summary.ExampleResponse = example
			}
		}
	}

	return summary, nil
}

// BuildMessageView creates a message view from the registry.
func BuildMessageView(reg *descriptor.Registry, fullName string) (*MessageView, error) {
	if reg == nil {
		return nil, fmt.Errorf("registry is nil")
	}

	message, exists := reg.FindMessage(fullName)
	if !exists {
		return nil, fmt.Errorf("message %q not found", fullName)
	}

	var fields []FieldView
	for i := 0; i < message.Fields().Len(); i++ {
		field := message.Fields().Get(i)
		fieldName := fmt.Sprintf("%s.%s", fullName, field.Name())

		fieldView := FieldView{
			Name:    string(field.Name()),
			Number:  int(field.Number()),
			Type:    formatFieldType(field),
			Label:   formatFieldLabel(field),
			Oneof:   formatOneofName(field),
			Comment: reg.CommentIndex[fieldName],
		}
		fields = append(fields, fieldView)
	}

	// Sort fields by number
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Number < fields[j].Number
	})

	// Generate example JSON
	exampleJSON := ""
	if reg != nil {
		if example, err := descriptor.GenerateExampleJSON(message, descriptor.DefaultExampleOptions()); err == nil {
			exampleJSON = example
		}
	}

	return &MessageView{
		Name:        string(message.Name()),
		FullName:    fullName,
		Package:     string(message.ParentFile().Package()),
		Comment:     reg.CommentIndex[fullName],
		Fields:      fields,
		ExampleJSON: exampleJSON,
	}, nil
}

// BuildEnumView creates an enum view from the registry.
func BuildEnumView(reg *descriptor.Registry, fullName string) (*EnumView, error) {
	if reg == nil {
		return nil, fmt.Errorf("registry is nil")
	}

	enum, exists := reg.FindEnum(fullName)
	if !exists {
		return nil, fmt.Errorf("enum %q not found", fullName)
	}

	var values []EnumValueView
	for i := 0; i < enum.Values().Len(); i++ {
		value := enum.Values().Get(i)
		valueName := fmt.Sprintf("%s.%s", fullName, value.Name())

		valueView := EnumValueView{
			Name:    string(value.Name()),
			Number:  int32(value.Number()),
			Comment: reg.CommentIndex[valueName],
		}
		values = append(values, valueView)
	}

	// Sort values by number
	sort.Slice(values, func(i, j int) bool {
		return values[i].Number < values[j].Number
	})

	return &EnumView{
		Name:     string(enum.Name()),
		FullName: fullName,
		Package:  string(enum.ParentFile().Package()),
		Comment:  reg.CommentIndex[fullName],
		Values:   values,
	}, nil
}

// formatFieldType formats a field type for display.
func formatFieldType(field protoreflect.FieldDescriptor) string {
	switch field.Kind() {
	case protoreflect.MessageKind:
		return string(field.Message().FullName())
	case protoreflect.EnumKind:
		return string(field.Enum().FullName())
	default:
		return strings.ToLower(field.Kind().String())
	}
}

// formatFieldLabel formats a field label for display.
func formatFieldLabel(field protoreflect.FieldDescriptor) string {
	if field.Cardinality() == protoreflect.Repeated {
		return "repeated"
	}
	if field.HasOptionalKeyword() {
		return "optional"
	}
	if field.Cardinality() == protoreflect.Required {
		return "required"
	}
	return ""
}

// formatOneofName formats a oneof name for display.
func formatOneofName(field protoreflect.FieldDescriptor) string {
	if field.ContainingOneof() != nil {
		return string(field.ContainingOneof().Name())
	}
	return ""
}

// extractHTTPRules extracts HTTP rules from a method descriptor.
func extractHTTPRules(method protoreflect.MethodDescriptor) ([]HTTPRule, error) {
	// TODO: Implement proper HTTP rule extraction using proto.GetExtension
	// This requires importing the google.api.annotations package and
	// properly resolving the extension descriptor.
	// For now, we'll return empty rules as the extension handling
	// requires more complex protobuf extension resolution.

	return []HTTPRule{}, nil
}

// generateCurlExample generates a curl example for the method.
func generateCurlExample(method *MethodSummary) string {
	if len(method.HTTPRules) == 0 {
		return ""
	}

	rule := method.HTTPRules[0]       // Use first rule
	host := "https://api.example.com" // Placeholder host

	var curlCmd string
	switch strings.ToUpper(rule.Method) {
	case "GET":
		curlCmd = fmt.Sprintf("curl -X GET %s%s", host, rule.Path)
	case "POST":
		curlCmd = fmt.Sprintf("curl -X POST %s%s \\\n  -H \"Content-Type: application/json\" \\\n  -d '{}'", host, rule.Path)
	case "PUT":
		curlCmd = fmt.Sprintf("curl -X PUT %s%s \\\n  -H \"Content-Type: application/json\" \\\n  -d '{}'", host, rule.Path)
	case "PATCH":
		curlCmd = fmt.Sprintf("curl -X PATCH %s%s \\\n  -H \"Content-Type: application/json\" \\\n  -d '{}'", host, rule.Path)
	case "DELETE":
		curlCmd = fmt.Sprintf("curl -X DELETE %s%s", host, rule.Path)
	default:
		curlCmd = fmt.Sprintf("curl -X %s %s%s", strings.ToUpper(rule.Method), host, rule.Path)
	}

	return curlCmd
}

// generateGrpcurlExample generates a grpcurl example for the method.
func generateGrpcurlExample(method *MethodSummary) string {
	host := "localhost:8080" // Placeholder host
	return fmt.Sprintf("grpcurl -plaintext -d '{}' %s %s", host, method.FullName)
}
