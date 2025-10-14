package server

import (
	"embed"
	"io/fs"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Server struct {
	router    *chi.Mux
	templates *template.Template
}

func New() (http.Handler, error) {
	t, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()
	// Static assets
	staticSub, _ := fs.Sub(staticFS, "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	s := &Server{router: r, templates: t}
	s.routes()
	return s.router, nil
}
