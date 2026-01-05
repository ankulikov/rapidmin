package server

import (
	"embed"
	"html"
	"net/http"
	"strings"
	"sync"

	"github.com/ankulikov/rapidmin/internal/config"
	"github.com/ankulikov/rapidmin/internal/providers"
)

//go:embed web/index.html
var indexFS embed.FS

const indexPath = "web/index.html"

type Server struct {
	cfg          config.AppConfig
	providers    providers.Registry
	indexHTML    []byte
	pathPrefix   string
	renderedOnce sync.Once
	renderedHTML []byte
}

type Option func(*Server)

func WithPathPrefix(prefix string) Option {
	return func(s *Server) {
		s.pathPrefix = normalizePrefix(prefix)
	}
}

func New(cfg config.AppConfig, providers providers.Registry, opts ...Option) (*Server, error) {
	indexHTML, err := indexFS.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		cfg:        cfg,
		providers:  providers,
		indexHTML:  indexHTML,
		pathPrefix: normalizePrefix(cfg.PathPrefix),
	}

	for _, opt := range opts {
		opt(srv)
	}

	return srv, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(s.apiConfigPath(), s.handleConfig)
	mux.HandleFunc(s.apiWidgetsPrefix(), s.handleWidgetData)
	if s.pathPrefix == "" {
		mux.HandleFunc("/", s.handleIndex)
	} else {
		mux.HandleFunc(s.pathPrefix, s.handleIndex)
		mux.HandleFunc(s.pathPrefix+"/", s.handleIndex)
	}
	return mux
}

func (s *Server) apiPrefix() string {
	return s.pathPrefix + "/api/"
}

func (s *Server) apiConfigPath() string {
	return s.pathPrefix + "/api/config"
}

func (s *Server) apiWidgetsPrefix() string {
	return s.pathPrefix + "/api/widgets/"
}

func normalizePrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || prefix == "/" {
		return ""
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	return strings.TrimSuffix(prefix, "/")
}

func (s *Server) renderIndexHTML() []byte {
	s.renderedOnce.Do(func() {
		htmlStr := string(s.indexHTML)
		s.renderedHTML = []byte(strings.Replace(htmlStr, "{{pathPrefix}}", html.EscapeString(s.pathPrefix), 1))
	})
	return s.renderedHTML
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(s.renderIndexHTML())
}
