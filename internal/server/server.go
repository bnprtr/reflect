package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
	"text/template"

	"github.com/bnprtr/reflect/internal/descriptor"
	"github.com/bnprtr/reflect/internal/server/theme"
	"github.com/go-chi/chi/v5"
)

//go:embed templates/*.html templates/partials/*.html static/*.css static/*.js
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Server struct {
	router    *chi.Mux
	templates *template.Template
	registry  *descriptor.Registry
	theme     *theme.Theme
}

func New(registry *descriptor.Registry) (http.Handler, error) {
	return NewWithTheme(registry, theme.GetDefaultTheme())
}

func NewWithTheme(registry *descriptor.Registry, themeConfig *theme.Theme) (http.Handler, error) {
	t, err := template.New("").Funcs(template.FuncMap{
		"contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},
	}).ParseFS(templatesFS, "templates/*.html", "templates/partials/*.html")
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()
	// Static assets
	staticSub, _ := fs.Sub(staticFS, "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	s := &Server{router: r, templates: t, registry: registry, theme: themeConfig}
	s.routes()
	return s.router, nil
}
