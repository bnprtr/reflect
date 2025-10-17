package tryit

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"
)

// GRPCWebInvoker implements the Invoker interface for the gRPC-Web protocol.
type GRPCWebInvoker struct {
	client *http.Client
}

// NewGRPCWebInvoker creates a new gRPC-Web invoker.
func NewGRPCWebInvoker() *GRPCWebInvoker {
	return &GRPCWebInvoker{
		client: &http.Client{
			// Client timeout is controlled per-request via context
			Timeout: 0,
		},
	}
}

// Invoke executes a gRPC-Web RPC.
// gRPC-Web uses HTTP POST with binary protobuf encoding to {baseURL}/{package.Service/Method}.
func (g *GRPCWebInvoker) Invoke(ctx context.Context, req *Request) (*Response, error) {
	start := time.Now()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create HTTP client with TLS configuration
	client := g.getHTTPClient(req.InsecureSkipVerify)

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

	// Marshal to binary protobuf
	requestBytes, err := proto.Marshal(inputMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build gRPC-Web message frame
	// Frame format: 1 byte flags + 4 bytes length + message
	frameBuffer := new(bytes.Buffer)

	// Compression flag (0 = no compression)
	frameBuffer.WriteByte(0)

	// Message length (4 bytes, big-endian)
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, uint32(len(requestBytes)))
	frameBuffer.Write(lengthBytes)

	// Message data
	frameBuffer.Write(requestBytes)

	// Build gRPC-Web URL: {baseURL}/{package.Service/Method}
	url := g.buildGRPCWebURL(req.BaseURL, req.MethodFullName())

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, frameBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set gRPC-Web protocol headers
	httpReq.Header.Set("Content-Type", "application/grpc-web+proto")
	// Accept both binary and text formats
	httpReq.Header.Set("Accept", "application/grpc-web+proto, application/grpc-web-text+proto")
	httpReq.Header.Set("X-Grpc-Web", "1")
	httpReq.Header.Set("X-User-Agent", "grpc-web-reflect/1.0")

	// Add user-provided headers (as gRPC metadata)
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Log the outgoing request
	slog.Info("Sending gRPC-Web request",
		"url", url,
		"method", httpReq.Method,
		"contentType", httpReq.Header.Get("Content-Type"),
		"bodyLength", frameBuffer.Len())

	// Execute request
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return &Response{
			Status:     int(codes.Unavailable),
			StatusText: "Request Failed",
			Latency:    time.Since(start),
			Error: &InvocationError{
				Code:    int(codes.Unavailable),
				Message: fmt.Sprintf("HTTP request failed: %v", err),
			},
		}, nil
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return &Response{
			Status:     httpResp.StatusCode,
			StatusText: httpResp.Status,
			Headers:    httpResp.Header,
			Latency:    time.Since(start),
			Error: &InvocationError{
				Code:    httpResp.StatusCode,
				Message: fmt.Sprintf("failed to read response body: %v", err),
			},
		}, nil
	}

	// Debug logging
	debugLen := len(respBody)
	if debugLen > 64 {
		debugLen = 64
	}
	contentType := httpResp.Header.Get("Content-Type")
	slog.Debug("gRPC-Web response received",
		"bodyLength", len(respBody),
		"contentType", contentType,
		"hexDump", hex.EncodeToString(respBody[:debugLen]))

	// Check if response is base64-encoded (grpc-web-text format)
	// First check content-type
	isTextFormat := strings.Contains(contentType, "grpc-web-text") || strings.Contains(contentType, "text")

	// Also check if the response looks like base64 (all printable ASCII)
	// gRPC-Web binary format starts with 0x00 or 0x80, not ASCII characters
	looksLikeBase64 := len(respBody) > 0 && respBody[0] >= 0x20 && respBody[0] <= 0x7E

	if isTextFormat || looksLikeBase64 {
		slog.Info("Detected text/base64 format response", "contentType", contentType, "firstByte", respBody[0])
		// Decode base64 response
		decoded, err := base64.StdEncoding.DecodeString(string(respBody))
		if err != nil {
			// Log the actual response body to see what we're dealing with
			slog.Warn("Failed to decode base64 response, trying as binary",
				"error", err,
				"responseBody", string(respBody))
		} else {
			slog.Info("Successfully decoded base64 response", "originalLen", len(respBody), "decodedLen", len(decoded))
			respBody = decoded
		}
	}

	// Check for gRPC status in headers/trailers
	grpcStatus := g.extractGRPCStatus(httpResp.Header)
	grpcMessage := httpResp.Header.Get("grpc-message")

	// Parse the response frame
	outputMsg := dynamicpb.NewMessage(req.OutputMessageDescriptor())
	var jsonBody string

	if len(respBody) > 0 {
		// Try to parse the gRPC-Web frame
		messageData, err := g.parseGRPCWebFrame(respBody)
		if err != nil {
			return &Response{
				Status:     int(codes.Internal),
				StatusText: "Internal Error",
				Headers:    httpResp.Header,
				Latency:    time.Since(start),
				Error: &InvocationError{
					Code:    int(codes.Internal),
					Message: fmt.Sprintf("failed to parse gRPC-Web frame: %v", err),
				},
			}, nil
		}

		// Unmarshal the protobuf message
		if len(messageData) > 0 {
			if err := proto.Unmarshal(messageData, outputMsg); err != nil {
				return &Response{
					Status:     int(codes.Internal),
					StatusText: "Internal Error",
					Headers:    httpResp.Header,
					Latency:    time.Since(start),
					Error: &InvocationError{
						Code:    int(codes.Internal),
						Message: fmt.Sprintf("failed to unmarshal response: %v", err),
					},
				}, nil
			}

			// Marshal to JSON for display
			formattedJSON, err := protojson.MarshalOptions{
				Multiline:       true,
				Indent:          "  ",
				EmitUnpopulated: false,
			}.Marshal(outputMsg)
			if err == nil {
				jsonBody = string(formattedJSON)
			}
		}
	}

	// Check if there was a gRPC error
	if grpcStatus != 0 {
		return &Response{
			Status:     grpcStatus,
			StatusText: codes.Code(grpcStatus).String(),
			Headers:    httpResp.Header,
			JSONBody:   jsonBody,
			Latency:    time.Since(start),
			Error: &InvocationError{
				Code:    grpcStatus,
				Message: grpcMessage,
			},
		}, nil
	}

	return &Response{
		Status:     int(codes.OK),
		StatusText: codes.OK.String(),
		Headers:    httpResp.Header,
		JSONBody:   jsonBody,
		Latency:    time.Since(start),
	}, nil
}

// buildGRPCWebURL constructs the gRPC-Web protocol URL.
// Format: {baseURL}/{package.Service/Method}
func (g *GRPCWebInvoker) buildGRPCWebURL(baseURL, methodFullName string) string {
	// Ensure base URL doesn't end with slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Ensure method name starts with slash
	if !strings.HasPrefix(methodFullName, "/") {
		methodFullName = "/" + methodFullName
	}

	return baseURL + methodFullName
}

// getHTTPClient returns an HTTP client with the appropriate TLS configuration.
func (g *GRPCWebInvoker) getHTTPClient(insecureSkipVerify bool) *http.Client {
	if !insecureSkipVerify {
		return g.client
	}

	// Create a client with TLS verification disabled
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

// parseGRPCWebFrame parses a gRPC-Web response frame.
// gRPC-Web responses can contain multiple frames:
// - Data frames: flag 0x00, 4 bytes length, message data
// - Trailer frames: flag 0x80, 4 bytes length, trailer data (text)
func (g *GRPCWebInvoker) parseGRPCWebFrame(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	offset := 0
	var messageData []byte

	// Parse all frames in the response
	for offset < len(data) {
		// Need at least 5 bytes for frame header (1 flag + 4 length)
		if offset+5 > len(data) {
			// Incomplete frame header - this might be a partial response
			break
		}

		// Read frame flag
		frameFlag := data[offset]
		offset++

		// Read frame length (big-endian uint32)
		frameLength := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4

		// Check if we have enough data for this frame
		if offset+int(frameLength) > len(data) {
			// Incomplete frame body
			return nil, fmt.Errorf("incomplete frame: expected %d bytes, got %d", offset+int(frameLength), len(data))
		}

		frameData := data[offset : offset+int(frameLength)]
		offset += int(frameLength)

		// Process based on frame type
		if frameFlag == 0x00 {
			// Data frame - this is the message
			messageData = frameData
		} else if frameFlag == 0x80 {
			// Trailer frame - contains HTTP headers as text (we can ignore for now)
			// Trailers might contain grpc-status and grpc-message
			continue
		} else {
			// Unknown frame type - skip it
			continue
		}
	}

	return messageData, nil
}

// extractGRPCStatus extracts the gRPC status code from response headers.
func (g *GRPCWebInvoker) extractGRPCStatus(headers http.Header) int {
	statusStr := headers.Get("grpc-status")
	if statusStr == "" {
		// No explicit status means OK (0)
		return 0
	}

	status, err := strconv.Atoi(statusStr)
	if err != nil {
		// Invalid status, treat as unknown error
		return int(codes.Unknown)
	}

	return status
}
