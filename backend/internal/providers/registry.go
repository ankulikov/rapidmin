package providers

import (
	"context"

	"github.com/ankulikov/rapidmin/internal/config"
)

type DataRequest struct {
	Limit   int
	Cursor  string
	Filters []Filter
}

type Filter struct {
	Name     string
	Operator string
	Values   []string
}

type DataResponse struct {
	Data       []map[string]any `json:"data"`
	Total      int              `json:"total"`
	NextCursor string           `json:"next_cursor,omitempty"`
	HasMore    bool             `json:"has_more,omitempty"`
}

type Provider interface {
	Fetch(ctx context.Context, widget config.Widget, req DataRequest) (DataResponse, error)
}

type Registry map[string]Provider

func (r Registry) Get(name string) (Provider, bool) {
	provider, ok := r[name]
	return provider, ok
}
