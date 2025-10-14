package server

import (
	"net/http"
)

func (s *Server) routes() {
	s.router.Get("/", s.handleHome())
}

func (s *Server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = s.templates.ExecuteTemplate(w, "home.html", map[string]any{"Title": "Reflect"})
	}
}
