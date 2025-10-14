package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bnprtr/reflect/internal/descriptor"
)

func TestDocHandlers(t *testing.T) {
	// Load test registry
	ctx := context.Background()
	testDataPath := filepath.Join("..", "descriptor", "testdata", "basic")
	reg, err := descriptor.LoadDirectory(ctx, testDataPath, []string{})
	if err != nil {
		t.Fatalf("Failed to load test registry: %v", err)
	}

	// Debug: check if registry has services
	if len(reg.ServicesByName) == 0 {
		t.Fatalf("Registry has no services loaded")
	}
	t.Logf("Registry loaded with %d services: %v", len(reg.ServicesByName), reg.ServicesByName)

	// Create server with test registry
	srv, err := New(reg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// srv is a chi.Router, so we can use it directly

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedText   []string
	}{
		{
			name:           "home page",
			method:         "GET",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"API Documentation", "EchoService", "echo.v1.EchoService"},
		},
		{
			name:           "service detail",
			method:         "GET",
			path:           "/services/echo.v1.EchoService",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"EchoService", "Echo", "EchoStream", "echo.v1.EchoRequest", "echo.v1.EchoResponse"},
		},
		{
			name:           "method detail",
			method:         "GET",
			path:           "/methods/echo.v1.EchoService/Echo",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"Echo", "echo.v1.EchoRequest", "echo.v1.EchoResponse"},
		},
		{
			name:           "message type detail",
			method:         "GET",
			path:           "/types/echo.v1.EchoRequest",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"EchoRequest", "message", "message", "count"},
		},
		{
			name:           "enum type detail",
			method:         "GET",
			path:           "/types/echo.v1.Status",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"Status", "STATUS_UNSPECIFIED", "STATUS_SUCCESS", "STATUS_ERROR"},
		},
		{
			name:           "type partial",
			method:         "GET",
			path:           "/partial/types/echo.v1.EchoRequest",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"EchoRequest", "message", "count"},
		},
		{
			name:           "non-existent service",
			method:         "GET",
			path:           "/services/non.existent.Service",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "non-existent method",
			method:         "GET",
			path:           "/methods/echo.v1.EchoService/NonExistent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "non-existent type",
			method:         "GET",
			path:           "/types/non.existent.Type",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			srv.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedText != nil {
				body := w.Body.String()
				for _, text := range tt.expectedText {
					if !strings.Contains(body, text) {
						t.Errorf("Expected body to contain %q, but it didn't. Body: %s", text, body)
					}
				}
			}
		})
	}
}

func TestDocHandlersWithNilRegistry(t *testing.T) {
	// Create server with nil registry
	srv, err := New(nil)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedText   []string
	}{
		{
			name:           "home page with nil registry",
			method:         "GET",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"No services found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			srv.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedText != nil {
				body := w.Body.String()
				for _, text := range tt.expectedText {
					if !strings.Contains(body, text) {
						t.Errorf("Expected body to contain %q, but it didn't. Body: %s", text, body)
					}
				}
			}
		})
	}
}
