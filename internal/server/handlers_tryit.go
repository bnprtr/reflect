package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bnprtr/reflect/internal/tryit"
)

// TryItRequest represents the JSON request body for the /api/tryit/invoke endpoint.
type TryItRequest struct {
	// Environment is the name of the configured environment to invoke against.
	Environment string `json:"environment"`

	// Method is the fully-qualified method name (e.g., "echo.v1.EchoService/Echo").
	Method string `json:"method"`

	// Transport is the RPC transport to use (connect, grpc, or grpc-web).
	// If empty, uses the environment's default transport.
	Transport string `json:"transport,omitempty"`

	// Headers are additional headers to include with the request.
	Headers map[string]string `json:"headers,omitempty"`

	// Body is the JSON request body.
	Body string `json:"body"`
}

// TryItResponse represents the JSON response for the /api/tryit/invoke endpoint.
type TryItResponse struct {
	// Success indicates whether the invocation succeeded.
	Success bool `json:"success"`

	// Status is the HTTP status code or gRPC status code.
	Status int `json:"status"`

	// StatusText is a human-readable status description.
	StatusText string `json:"statusText"`

	// Headers are the response headers (with sensitive values redacted).
	Headers map[string][]string `json:"headers,omitempty"`

	// Body is the response body as JSON.
	Body string `json:"body,omitempty"`

	// Latency is the request duration in milliseconds.
	LatencyMs int64 `json:"latencyMs"`

	// Error contains error details if the invocation failed.
	Error *TryItError `json:"error,omitempty"`
}

// TryItError represents error details in the Try It response.
type TryItError struct {
	// Code is the error code.
	Code int `json:"code"`

	// Message is the error message.
	Message string `json:"message"`

	// Details contains additional error information.
	Details []string `json:"details,omitempty"`
}

// handleTryItInvoke handles POST /api/tryit/invoke requests.
func (s *Server) handleTryItInvoke(w http.ResponseWriter, r *http.Request) {
	// Ensure we have a config
	if s.config == nil {
		s.writeJSONError(w, http.StatusServiceUnavailable, "Try It functionality is not configured (missing reflect.yaml)")
		return
	}

	// Parse form data from request
	if err := r.ParseForm(); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse form data: %v", err))
		return
	}

	// Extract form values into TryItRequest
	tryItReq := TryItRequest{
		Environment: r.FormValue("environment"),
		Method:      r.FormValue("method"),
		Transport:   r.FormValue("transport"),
		Body:        r.FormValue("body"),
	}

	// Parse headers JSON if provided
	headersJSON := r.FormValue("headers")
	if headersJSON != "" && headersJSON != "{}" {
		if err := json.Unmarshal([]byte(headersJSON), &tryItReq.Headers); err != nil {
			s.writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse headers JSON: %v", err))
			return
		}
	}

	// Validate request size
	if err := tryit.ValidateJSONSize(tryItReq.Body, s.config.MaxRequestBodyBytes); err != nil {
		s.writeJSONError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}

	// Get registry
	registry, _ := s.getRegistry()
	if registry == nil {
		s.writeJSONError(w, http.StatusServiceUnavailable, "No protobuf descriptors loaded")
		return
	}

	// Look up method descriptor
	methodDesc, exists := registry.FindMethod(tryItReq.Method)
	if !exists {
		s.writeJSONError(w, http.StatusNotFound, fmt.Sprintf("method %q not found", tryItReq.Method))
		return
	}

	// Look up environment configuration
	env, err := s.config.GetEnvironment(tryItReq.Environment)
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("environment %q not found", tryItReq.Environment))
		return
	}

	// Determine transport
	transport := tryItReq.Transport
	if transport == "" {
		transport = env.Transport
	}

	parsedTransport, err := tryit.ParseTransport(transport)
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Filter headers through allowlist
	filteredHeaders := tryit.FilterHeaders(tryItReq.Headers, s.config.HeaderAllowlist)

	// Merge with environment default headers
	mergedHeaders := tryit.MergeHeaders(env.DefaultHeaders, filteredHeaders)

	// Create invoker request
	invokerReq := &tryit.Request{
		Environment:      tryItReq.Environment,
		MethodDescriptor: methodDesc,
		JSONBody:         tryItReq.Body,
		Headers:          mergedHeaders,
		BaseURL:          env.BaseURL,
		Timeout:          s.config.GetTimeout(),
		InsecureSkipVerify: env.TLS.InsecureSkipVerify,
	}

	// Select appropriate invoker
	var invoker tryit.Invoker
	switch parsedTransport {
	case tryit.TransportConnect:
		invoker = tryit.NewConnectInvoker()
	case tryit.TransportGRPC:
		invoker = tryit.NewGRPCInvoker()
	case tryit.TransportGRPCWeb:
		s.writeJSONError(w, http.StatusNotImplemented, "gRPC-Web transport is not yet implemented")
		return
	default:
		s.writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("unsupported transport: %s", parsedTransport))
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), s.config.GetTimeout())
	defer cancel()

	// Execute invocation
	resp, err := invoker.Invoke(ctx, invokerReq)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("invocation failed: %v", err))
		return
	}

	// Redact sensitive headers
	redactedHeaders := tryit.RedactSensitiveHeaders(resp.Headers)

	// Build response
	tryItResp := TryItResponse{
		Success:    resp.Error == nil,
		Status:     resp.Status,
		StatusText: resp.StatusText,
		Headers:    redactedHeaders,
		Body:       resp.JSONBody,
		LatencyMs:  resp.Latency.Milliseconds(),
	}

	if resp.Error != nil {
		tryItResp.Error = &TryItError{
			Code:    resp.Error.Code,
			Message: resp.Error.Message,
			Details: resp.Error.Details,
		}
	}

	// Render response template
	w.Header().Set("Content-Type", "text/html")
	if err := s.templates.ExecuteTemplate(w, "tryit_response.html", tryItResp); err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
		return
	}
}

// writeJSONError writes a JSON error response.
func (s *Server) writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    statusCode,
			"message": message,
		},
	}

	json.NewEncoder(w).Encode(resp)
}
