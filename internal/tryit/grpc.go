package tryit

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
)

// GRPCInvoker implements the Invoker interface for the gRPC protocol.
type GRPCInvoker struct {
}

// NewGRPCInvoker creates a new gRPC invoker.
func NewGRPCInvoker() *GRPCInvoker {
	return &GRPCInvoker{}
}

// Invoke executes a gRPC RPC using dynamic invocation.
func (g *GRPCInvoker) Invoke(ctx context.Context, req *Request) (*Response, error) {
	start := time.Now()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Determine credentials
	var creds credentials.TransportCredentials
	if req.InsecureSkipVerify {
		creds = credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})
	} else {
		// Use TLS with system cert pool
		creds = credentials.NewTLS(&tls.Config{})
	}

	// Determine if we should use TLS based on the URL scheme
	target := req.BaseURL
	if target[:4] == "http" {
		// Strip http:// or https:// prefix for gRPC dial
		if target[:8] == "https://" {
			target = target[8:]
		} else if target[:7] == "http://" {
			target = target[7:]
			// For http:// URLs, use insecure credentials
			creds = insecure.NewCredentials()
		}
	}

	// Create gRPC connection
	conn, err := grpc.Dial(
		target,
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(false)),
	)
	if err != nil {
		return &Response{
			Status:     int(codes.Unavailable),
			StatusText: "Connection Failed",
			Latency:    time.Since(start),
			Error: &InvocationError{
				Code:    int(codes.Unavailable),
				Message: fmt.Sprintf("failed to connect to gRPC server: %v", err),
			},
		}, nil
	}
	defer conn.Close()

	// Parse JSON into dynamic protobuf message
	inputMsg := dynamicpb.NewMessage(req.InputMessageDescriptor())
	if req.JSONBody != "" {
		if err := protojson.Unmarshal([]byte(req.JSONBody), inputMsg); err != nil {
			return &Response{
				Status:     int(codes.InvalidArgument),
				StatusText: "Invalid Argument",
				Latency:    time.Since(start),
				Error: &InvocationError{
					Code:    int(codes.InvalidArgument),
					Message: fmt.Sprintf("failed to parse JSON request: %v", err),
				},
			}, nil
		}
	}

	// Create output message
	outputMsg := dynamicpb.NewMessage(req.OutputMessageDescriptor())

	// Add metadata from headers
	md := metadata.New(req.Headers)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Build full method name for gRPC: /package.Service/Method
	fullMethod := "/" + req.MethodFullName()

	// Invoke the RPC
	var responseHeader metadata.MD
	err = conn.Invoke(
		ctx,
		fullMethod,
		inputMsg,
		outputMsg,
		grpc.Header(&responseHeader),
	)

	latency := time.Since(start)

	// Convert metadata to header map
	headers := make(map[string][]string)
	for k, v := range responseHeader {
		headers[k] = v
	}

	// Handle error
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return &Response{
				Status:     int(codes.Unknown),
				StatusText: "Unknown Error",
				Headers:    headers,
				Latency:    latency,
				Error: &InvocationError{
					Code:    int(codes.Unknown),
					Message: fmt.Sprintf("RPC failed: %v", err),
				},
			}, nil
		}

		// Extract error details
		details := make([]string, 0, len(st.Details()))
		for _, detail := range st.Details() {
			details = append(details, fmt.Sprintf("%v", detail))
		}

		return &Response{
			Status:     int(st.Code()),
			StatusText: st.Code().String(),
			Headers:    headers,
			Latency:    latency,
			Error: &InvocationError{
				Code:    int(st.Code()),
				Message: st.Message(),
				Details: details,
			},
		}, nil
	}

	// Marshal response to JSON for display
	formattedJSON, err := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		EmitUnpopulated: false,
	}.Marshal(outputMsg)
	if err != nil {
		// Fall back to binary format description
		formattedJSON = []byte(fmt.Sprintf("{\"error\": \"failed to format response: %v\"}", err))
	}

	return &Response{
		Status:     int(codes.OK),
		StatusText: codes.OK.String(),
		Headers:    headers,
		JSONBody:   string(formattedJSON),
		Latency:    latency,
	}, nil
}

// marshalProto is a helper to marshal a proto message (unused but kept for reference).
func marshalProto(msg proto.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

// unmarshalProto is a helper to unmarshal a proto message (unused but kept for reference).
func unmarshalProto(data []byte, msg proto.Message) error {
	return proto.Unmarshal(data, msg)
}
