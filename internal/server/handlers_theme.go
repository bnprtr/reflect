package server

import (
	"encoding/json"
	"net/http"

	"github.com/bnprtr/reflect/internal/server/theme"
)

// handleThemesList returns all available themes
func (s *Server) handleThemesList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		themes := theme.GetAllThemes()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"themes": themes,
		}); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// handleCurrentTheme returns the currently active theme
func (s *Server) handleCurrentTheme() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"name":   s.theme.Name,
			"colors": s.theme.Colors,
		}); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
