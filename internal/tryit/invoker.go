package tryit

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Invoker represents a transport-agnostic RPC invoker.
type Invoker interface {
	// Invoke executes an RPC and returns the response.
	Invoke(ctx context.Context, req *Request) (*Response, error)
}

// Request represents a "Try It" invocation request.
type Request struct {
	// Environment is the name of the configured environment to invoke against.
	Environment string

	// MethodDescriptor is the protobuf method descriptor for the RPC.
	MethodDescriptor protoreflect.MethodDescriptor

	// JSONBody is the user-provided JSON request body (to be converted to protobuf).
	JSONBody string

	// Headers are additional HTTP/metadata headers to include with the request.
	// These should already be filtered through the header allowlist.
	Headers map[string]string

	// BaseURL is the base URL of the upstream service (from environment config).
	BaseURL string

	// Timeout is the maximum duration for the request.
	Timeout time.Duration

	// InsecureSkipVerify indicates whether to skip TLS certificate verification.
	InsecureSkipVerify bool
}

// Response represents the result of an RPC invocation.
type Response struct {
	// Status is the HTTP status code (for Connect) or gRPC status code.
	Status int

	// StatusText is a human-readable status description.
	StatusText string

	// Headers are the response headers/metadata returned by the server.
	// Sensitive headers should be redacted before returning to the user.
	Headers map[string][]string

	// JSONBody is the response body converted to JSON for display.
	JSONBody string

	// Latency is the total time taken for the request (including network and processing).
	Latency time.Duration

	// Error contains error details if the invocation failed.
	Error *InvocationError
}

// InvocationError represents detailed error information from an invocation.
type InvocationError struct {
	// Code is the error code (gRPC code or HTTP status code).
	Code int

	// Message is the error message.
	Message string

	// Details contains additional error details (e.g., gRPC error details).
	Details []string
}

// Transport represents the RPC transport protocol.
type Transport string

const (
	// TransportConnect represents the Connect protocol (JSON over HTTP).
	TransportConnect Transport = "connect"

	// TransportGRPC represents the gRPC protocol.
	TransportGRPC Transport = "grpc"

	// TransportGRPCWeb represents the gRPC-Web protocol.
	TransportGRPCWeb Transport = "grpc-web"
)

// ParseTransport converts a string to a Transport type.
func ParseTransport(s string) (Transport, error) {
	switch s {
	case string(TransportConnect), "":
		return TransportConnect, nil
	case string(TransportGRPC):
		return TransportGRPC, nil
	case string(TransportGRPCWeb):
		return TransportGRPCWeb, nil
	default:
		return "", fmt.Errorf("invalid transport: %q (must be connect, grpc, or grpc-web)", s)
	}
}

// String returns the string representation of the transport.
func (t Transport) String() string {
	return string(t)
}

// Validate validates that a Request has all required fields.
func (r *Request) Validate() error {
	if r.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	if r.MethodDescriptor == nil {
		return fmt.Errorf("method descriptor is required")
	}
	if r.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	if r.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	return nil
}

// MethodFullName returns the fully-qualified method name in the format "package.Service/Method".
func (r *Request) MethodFullName() string {
	if r.MethodDescriptor == nil {
		return ""
	}
	service := r.MethodDescriptor.Parent().(protoreflect.ServiceDescriptor)
	return fmt.Sprintf("%s/%s", service.FullName(), r.MethodDescriptor.Name())
}

// InputMessageDescriptor returns the descriptor for the input message type.
func (r *Request) InputMessageDescriptor() protoreflect.MessageDescriptor {
	if r.MethodDescriptor == nil {
		return nil
	}
	return r.MethodDescriptor.Input()
}

// OutputMessageDescriptor returns the descriptor for the output message type.
func (r *Request) OutputMessageDescriptor() protoreflect.MessageDescriptor {
	if r.MethodDescriptor == nil {
		return nil
	}
	return r.MethodDescriptor.Output()
}
