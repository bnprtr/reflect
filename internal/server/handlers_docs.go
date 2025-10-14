package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bnprtr/reflect/internal/docs"
	"github.com/go-chi/chi/v5"
)

// baseData returns common template data with theme configuration
func (s *Server) baseData() map[string]any {
	return map[string]any{
		"ThemeVars": s.theme.ToCSSVariables(),
		"ThemeName": s.theme.Name,
	}
}

// mergeData merges additional data with base theme data
func (s *Server) mergeData(data map[string]any) map[string]any {
	base := s.baseData()
	for k, v := range data {
		base[k] = v
	}
	return base
}

func (s *Server) routes() {
	s.router.Get("/", s.handleHome())
	s.router.Get("/services/{fullName}", s.handleServiceDetail())
	s.router.Get("/methods/*", s.handleMethodDetail())
	s.router.Get("/types/{fullName}", s.handleTypeDetail())
	s.router.Get("/partial/types/*", s.handleTypePartial())
}

func (s *Server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		index, err := docs.BuildIndex(s.registry)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to build index: %v", err), http.StatusInternalServerError)
			return
		}

		data := s.mergeData(map[string]any{
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

		data := s.mergeData(map[string]any{
			"Title":          fmt.Sprintf("Service: %s", serviceView.Name),
			"Service":        serviceView,
			"Services":       []docs.ServiceSummary{{Name: serviceView.Name, FullName: serviceView.FullName, Package: serviceView.Package, Comment: serviceView.Comment}},
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

		data := s.mergeData(map[string]any{
			"Title":       fmt.Sprintf("Method: %s", methodView.Name),
			"Method":      methodView,
			"ServiceName": serviceName,
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

		// Try to find as message first, then as enum
		messageView, err := docs.BuildMessageView(s.registry, fullName)
		if err == nil {
			data := s.mergeData(map[string]any{
				"Title":   fmt.Sprintf("Message: %s", messageView.Name),
				"Message": messageView,
			})
			_ = s.templates.ExecuteTemplate(w, "type_detail.html", data)
			return
		}

		enumView, err := docs.BuildEnumView(s.registry, fullName)
		if err == nil {
			data := s.mergeData(map[string]any{
				"Title": fmt.Sprintf("Enum: %s", enumView.Name),
				"Enum":  enumView,
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
