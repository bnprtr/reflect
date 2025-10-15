package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bnprtr/reflect/internal/descriptor"
	"github.com/bnprtr/reflect/internal/docs"
	"github.com/bnprtr/reflect/internal/server/theme"
	"github.com/go-chi/chi/v5"
)

// baseData returns common template data with theme configuration
func (s *Server) baseData(r *http.Request) map[string]any {
	// Check for theme parameter in URL
	themeName := r.URL.Query().Get("theme")
	if themeName == "" {
		themeName = s.theme.Name
	}

	// Get theme by name (fallback to current theme if not found)
	themeConfig := theme.GetThemeByName(themeName)

	return map[string]any{
		"ThemeVars": themeConfig.ToCSSVariables(),
		"ThemeName": themeConfig.Name,
	}
}

// mergeData merges additional data with base theme data
func (s *Server) mergeData(r *http.Request, data map[string]any) map[string]any {
	base := s.baseData(r)
	for k, v := range data {
		base[k] = v
	}
	return base
}

func (s *Server) routes() {
	// Documentation routes
	s.router.Get("/", s.handleHome())
	s.router.Get("/services/{fullName}", s.handleServiceDetail())
	s.router.Get("/methods/*", s.handleMethodDetail())
	s.router.Get("/types/{fullName}", s.handleTypeDetail())
	s.router.Get("/partial/types/*", s.handleTypePartial())

	// Theme API routes
	s.router.Get("/api/themes", s.handleThemesList())
	s.router.Get("/api/themes/current", s.handleCurrentTheme())

	// Example generation API
	s.router.Post("/api/examples/generate", s.handleGenerateExample())

	// Search API
	s.router.Get("/api/search", s.handleSearch())
}

func (s *Server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		index, err := docs.BuildIndex(s.registry)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to build index: %v", err), http.StatusInternalServerError)
			return
		}

		data := s.mergeData(r, map[string]any{
			"Title":    "Reflect",
			"Services": index.Services,
		})

		err = s.templates.ExecuteTemplate(w, "home.html", data)
		if err != nil {
			http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleServiceDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fullName := chi.URLParam(r, "fullName")
		if fullName == "" {
			http.Error(w, "Service name required", http.StatusBadRequest)
			return
		}

		serviceView, err := docs.BuildServiceView(s.registry, fullName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Service not found: %v", err), http.StatusNotFound)
			return
		}

		// Get all services for sidebar navigation
		index, err := docs.BuildIndex(s.registry)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to build index: %v", err), http.StatusInternalServerError)
			return
		}

		data := s.mergeData(r, map[string]any{
			"Title":          fmt.Sprintf("Service: %s", serviceView.Name),
			"Service":        serviceView,
			"Services":       index.Services,
			"CurrentService": serviceView.FullName,
		})
		err = s.templates.ExecuteTemplate(w, "service_detail.html", data)
		if err != nil {
			http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleMethodDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fullName := chi.URLParam(r, "*")
		if fullName == "" {
			http.Error(w, "Method name required", http.StatusBadRequest)
			return
		}

		methodView, err := docs.BuildMethodView(s.registry, fullName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Method not found: %v", err), http.StatusNotFound)
			return
		}

		// Extract service name from method full name
		parts := strings.Split(fullName, "/")
		serviceName := ""
		if len(parts) >= 2 {
			serviceName = parts[0]
		}

		// Get all services for sidebar navigation
		index, err := docs.BuildIndex(s.registry)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to build index: %v", err), http.StatusInternalServerError)
			return
		}

		data := s.mergeData(r, map[string]any{
			"Title":          fmt.Sprintf("Method: %s", methodView.Name),
			"Method":         methodView,
			"ServiceName":    serviceName,
			"Services":       index.Services,
			"CurrentService": serviceName,
		})
		err = s.templates.ExecuteTemplate(w, "method_detail.html", data)
		if err != nil {
			http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleTypeDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fullName := chi.URLParam(r, "fullName")
		if fullName == "" {
			http.Error(w, "Type name required", http.StatusBadRequest)
			return
		}

		// Get all services for sidebar navigation
		index, err := docs.BuildIndex(s.registry)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to build index: %v", err), http.StatusInternalServerError)
			return
		}

		// Try to find as message first, then as enum
		messageView, err := docs.BuildMessageView(s.registry, fullName)
		if err == nil {
			data := s.mergeData(r, map[string]any{
				"Title":    fmt.Sprintf("Message: %s", messageView.Name),
				"Message":  messageView,
				"Services": index.Services,
			})
			_ = s.templates.ExecuteTemplate(w, "type_detail.html", data)
			return
		}

		enumView, err := docs.BuildEnumView(s.registry, fullName)
		if err == nil {
			data := s.mergeData(r, map[string]any{
				"Title":    fmt.Sprintf("Enum: %s", enumView.Name),
				"Enum":     enumView,
				"Services": index.Services,
			})
			_ = s.templates.ExecuteTemplate(w, "type_detail.html", data)
			return
		}

		http.Error(w, fmt.Sprintf("Type not found: %s", fullName), http.StatusNotFound)
	}
}

func (s *Server) handleTypePartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fullName := chi.URLParam(r, "*")
		if fullName == "" {
			http.Error(w, "Type name required", http.StatusBadRequest)
			return
		}

		// Try to find as message first, then as enum
		messageView, err := docs.BuildMessageView(s.registry, fullName)
		if err == nil {
			data := map[string]any{
				"Message": messageView,
			}
			err = s.templates.ExecuteTemplate(w, "type_detail_partial.html", data)
			if err != nil {
				http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
				return
			}
			return
		}

		enumView, err := docs.BuildEnumView(s.registry, fullName)
		if err == nil {
			data := map[string]any{
				"Enum": enumView,
			}
			_ = s.templates.ExecuteTemplate(w, "partials/type_detail_partial.html", data)
			return
		}

		http.Error(w, fmt.Sprintf("Type not found: %s", fullName), http.StatusNotFound)
	}
}

// GenerateExampleRequest represents the request body for example generation.
type GenerateExampleRequest struct {
	MessageType string                    `json:"messageType"`
	Options     descriptor.ExampleOptions `json:"options"`
}

// GenerateExampleResponse represents the response for example generation.
type GenerateExampleResponse struct {
	ExampleJSON string `json:"exampleJson"`
	Error       string `json:"error,omitempty"`
}

func (s *Server) handleGenerateExample() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req GenerateExampleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if req.MessageType == "" {
			http.Error(w, "messageType is required", http.StatusBadRequest)
			return
		}

		// Find the message in the registry
		msg, exists := s.registry.FindMessage(req.MessageType)
		if !exists {
			http.Error(w, fmt.Sprintf("Message type %s not found", req.MessageType), http.StatusNotFound)
			return
		}

		// Generate example JSON
		exampleJSON, err := descriptor.GenerateExampleJSON(msg, req.Options)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate example: %v", err), http.StatusInternalServerError)
			return
		}

		// Return response
		response := GenerateExampleResponse{
			ExampleJSON: exampleJSON,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) handleSearch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if len(query) < 2 {
			// Return empty results for short queries
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]docs.SearchResult{})
			return
		}

		results := s.searchIndex.Search(query)

		// Set content type for HTMX
		w.Header().Set("Content-Type", "text/html")

		// Render the search results template
		data := map[string]any{
			"Results": results,
			"Query":   query,
		}

		err := s.templates.ExecuteTemplate(w, "search_results.html", data)
		if err != nil {
			http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
