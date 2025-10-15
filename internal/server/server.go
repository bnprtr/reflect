package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
	"sync"
	"text/template"

	"github.com/bnprtr/reflect/internal/descriptor"
	"github.com/bnprtr/reflect/internal/docs"
	"github.com/bnprtr/reflect/internal/server/theme"
	"github.com/go-chi/chi/v5"
)

//go:embed templates/*.html templates/partials/*.html static/*.css static/*.js
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Server struct {
	router      *chi.Mux
	templates   *template.Template
	registry    *descriptor.Registry
	searchIndex *docs.SearchIndex
	theme       *theme.Theme
	mu          sync.RWMutex // Protects registry and searchIndex during hot reload
}

func New(registry *descriptor.Registry) (*Server, error) {
	return NewWithTheme(registry, theme.GetDefaultTheme())
}

func NewWithTheme(registry *descriptor.Registry, themeConfig *theme.Theme) (*Server, error) {
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

	// Build search index
	searchIndex := docs.BuildSearchIndex(registry)

	s := &Server{router: r, templates: t, registry: registry, searchIndex: searchIndex, theme: themeConfig}
	s.routes()
	return s, nil
}

// SetRegistry atomically updates the registry and rebuilds the search index
func (s *Server) SetRegistry(registry *descriptor.Registry) {
	searchIndex := docs.BuildSearchIndex(registry)

	s.mu.Lock()
	s.registry = registry
	s.searchIndex = searchIndex
	s.mu.Unlock()
}

// getRegistry safely retrieves the current registry
func (s *Server) getRegistry() (*descriptor.Registry, *docs.SearchIndex) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.registry, s.searchIndex
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
