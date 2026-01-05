package server

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ankulikov/rapidmin/internal/config"
	"github.com/ankulikov/rapidmin/internal/providers"
)

const defaultLimit = 50

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.cfg)
}

func (s *Server) handleWidgetData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	widgetID, ok := s.parseWidgetID(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	widget, ok := s.findWidget(widgetID)
	if !ok {
		http.NotFound(w, r)
		return
	}

	provider, ok := s.providers.Get(widget.Provider.Name)
	if !ok {
		http.Error(w, "unknown provider", http.StatusBadRequest)
		return
	}

	req := providers.DataRequest{
		Limit:   parseInt(r.URL.Query().Get("limit"), defaultLimit),
		Cursor:  r.URL.Query().Get("offset"),
		Filters: parseFilters(r.URL.Query()),
	}

	data, err := provider.Fetch(r.Context(), widget, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotImplemented)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) parseWidgetID(path string) (string, bool) {
	prefix := s.apiWidgetsPrefix()
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}

	trimmed := strings.TrimPrefix(path, prefix)
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" {
		return "", false
	}
	if strings.Contains(trimmed, "/") {
		return "", false
	}

	return trimmed, true
}

func parseFilters(values url.Values) []providers.Filter {
	filtersByKey := map[string]*providers.Filter{}
	order := []string{}
	for key, values := range values {
		if key == "limit" || key == "offset" {
			continue
		}
		if len(values) == 0 {
			continue
		}

		name := key
		operator := ""
		if dot := strings.LastIndex(key, "."); dot != -1 {
			name = key[:dot]
			operator = key[dot+1:]
		}
		if name == "" {
			continue
		}

		mapKey := name + "|" + operator
		filter, ok := filtersByKey[mapKey]
		if !ok {
			filter = &providers.Filter{Name: name, Operator: config.FilterOperator(operator)}
			filtersByKey[mapKey] = filter
			order = append(order, mapKey)
		}
		for _, value := range values {
			if value == "" {
				continue
			}
			filter.Values = append(filter.Values, value)
		}
	}

	filters := make([]providers.Filter, 0, len(order))
	for _, key := range order {
		filter, ok := filtersByKey[key]
		if !ok || filter == nil {
			continue
		}
		filters = append(filters, *filter)
	}
	return filters
}

func parseInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func (s *Server) findWidget(id string) (config.Widget, bool) {
	for _, page := range s.cfg.Pages {
		for _, widget := range page.Widgets {
			if widget.ID == id {
				return widget, true
			}
		}
	}

	return config.Widget{}, false
}
