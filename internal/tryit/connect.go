package tryit

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/dynamicpb"
)

// ConnectInvoker implements the Invoker interface for the Connect protocol.
type ConnectInvoker struct {
	client *http.Client
}

// NewConnectInvoker creates a new Connect invoker.
func NewConnectInvoker() *ConnectInvoker {
	return &ConnectInvoker{
		client: &http.Client{
			// Client timeout is controlled per-request via context
			Timeout: 0,
		},
	}
}

// Invoke executes a Connect RPC.
// Connect uses HTTP POST with JSON encoding to /{package}.{Service}/{Method}.
func (c *ConnectInvoker) Invoke(ctx context.Context, req *Request) (*Response, error) {
	start := time.Now()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create HTTP client with TLS configuration
	client := c.getHTTPClient(req.InsecureSkipVerify)

	// Parse JSON into dynamic protobuf message
	inputMsg := dynamicpb.NewMessage(req.InputMessageDescriptor())
	if req.JSONBody != "" {
		if err := protojson.Unmarshal([]byte(req.JSONBody), inputMsg); err != nil {
			return &Response{
				Status:     http.StatusBadRequest,
				StatusText: "Bad Request",
				Latency:    time.Since(start),
				Error: &InvocationError{
					Code:    http.StatusBadRequest,
					Message: fmt.Sprintf("failed to parse JSON request: %v", err),
				},
			}, nil
		}
	}

	// Marshal to Connect JSON format (protojson)
	requestBytes, err := protojson.Marshal(inputMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build Connect URL: {baseURL}/{package.Service/Method}
	url := c.buildConnectURL(req.BaseURL, req.MethodFullName())

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set Connect protocol headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Add user-provided headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return &Response{
			Status:     0,
			StatusText: "Request Failed",
			Latency:    time.Since(start),
			Error: &InvocationError{
				Code:    0,
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

	// Handle non-200 responses
	if httpResp.StatusCode != http.StatusOK {
		return &Response{
			Status:     httpResp.StatusCode,
			StatusText: httpResp.Status,
			Headers:    httpResp.Header,
			JSONBody:   string(respBody),
			Latency:    time.Since(start),
			Error: &InvocationError{
				Code:    httpResp.StatusCode,
				Message: fmt.Sprintf("RPC failed with status %d", httpResp.StatusCode),
				Details: []string{string(respBody)},
			},
		}, nil
	}

	// Parse response JSON into dynamic message
	outputMsg := dynamicpb.NewMessage(req.OutputMessageDescriptor())
	if len(respBody) > 0 {
		if err := protojson.Unmarshal(respBody, outputMsg); err != nil {
			return &Response{
				Status:     httpResp.StatusCode,
				StatusText: httpResp.Status,
				Headers:    httpResp.Header,
				JSONBody:   string(respBody),
				Latency:    time.Since(start),
				Error: &InvocationError{
					Code:    httpResp.StatusCode,
					Message: fmt.Sprintf("failed to parse response JSON: %v", err),
					Details: []string{string(respBody)},
				},
			}, nil
		}
	}

	// Marshal back to formatted JSON for display
	formattedJSON, err := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		EmitUnpopulated: false,
	}.Marshal(outputMsg)
	if err != nil {
		// Fall back to raw response if we can't format it
		formattedJSON = respBody
	}

	return &Response{
		Status:     httpResp.StatusCode,
		StatusText: httpResp.Status,
		Headers:    httpResp.Header,
		JSONBody:   string(formattedJSON),
		Latency:    time.Since(start),
	}, nil
}

// buildConnectURL constructs the Connect protocol URL.
// Format: {baseURL}/{package.Service/Method}
// The method full name is already in the format "package.Service/Method".
func (c *ConnectInvoker) buildConnectURL(baseURL, methodFullName string) string {
	// Ensure base URL doesn't end with slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Ensure method name starts with slash
	if !strings.HasPrefix(methodFullName, "/") {
		methodFullName = "/" + methodFullName
	}

	return baseURL + methodFullName
}

// getHTTPClient returns an HTTP client with the appropriate TLS configuration.
func (c *ConnectInvoker) getHTTPClient(insecureSkipVerify bool) *http.Client {
	if !insecureSkipVerify {
		return c.client
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
