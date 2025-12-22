package server

import (
	"embed"
	"net/http"
	"strings"

	"github.com/ankulikov/rapidmin/internal/config"
	"github.com/ankulikov/rapidmin/internal/providers"
)

//go:embed web/index.html
var indexFS embed.FS

const indexPath = "web/index.html"

type Server struct {
	cfg       config.AppConfig
	providers providers.Registry
	indexHTML []byte
}

func New(cfg config.AppConfig, providers providers.Registry) (*Server, error) {
	indexHTML, err := indexFS.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	return &Server{
		cfg:       cfg,
		providers: providers,
		indexHTML: indexHTML,
	}, nil
}

func (s *Server) Handler() http.Handler {
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/api/config", s.handleConfig)
	apiMux.HandleFunc("/api/widgets/", s.handleWidgetData)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			apiMux.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(s.indexHTML)
	})
}
