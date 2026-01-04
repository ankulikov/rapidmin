package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ankulikov/rapidmin/internal/config"
	"github.com/ankulikov/rapidmin/internal/providers"
	sqlprovider "github.com/ankulikov/rapidmin/internal/providers/sql"
)

type dataResponse struct {
	Data       []map[string]any `json:"data"`
	Total      int              `json:"total"`
	NextCursor string           `json:"next_cursor"`
	HasMore    bool             `json:"has_more"`
}

func TestServerConfigAndWidgetData(t *testing.T) {
	db := setupSQLiteDB(t)
	providerRegistry := providers.Registry{
		"db": sqlprovider.NewWithDB(db),
	}

	cfg := sampleConfig()
	app, err := New(cfg, providerRegistry)
	if err != nil {
		t.Fatalf("server init: %v", err)
	}

	srv := httptest.NewServer(app.Handler())
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/api/config")
	if err != nil {
		t.Fatalf("config request: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("config status: %d", resp.StatusCode)
	}

	query := url.Values{}
	query.Set("limit", "2")
	query.Add("name.contains", "ann")
	dataResp := fetchWidgetData(t, srv.URL, query)
	if dataResp.Total != 2 {
		t.Fatalf("expected 2 rows, got %d", dataResp.Total)
	}

	query = url.Values{}
	query.Add("age.gt", "40")
	dataResp = fetchWidgetData(t, srv.URL, query)
	if dataResp.Total != 1 {
		t.Fatalf("expected 1 row, got %d", dataResp.Total)
	}

	query = url.Values{}
	query.Add("tags", "vip")
	query.Add("tags", "active")
	dataResp = fetchWidgetData(t, srv.URL, query)
	if dataResp.Total != 3 {
		t.Fatalf("expected 3 rows, got %d", dataResp.Total)
	}

	query = url.Values{}
	query.Add("created.between", "2024-01-01")
	query.Add("created.between", "2024-01-31")
	dataResp = fetchWidgetData(t, srv.URL, query)
	if dataResp.Total != 1 {
		t.Fatalf("expected 1 row, got %d", dataResp.Total)
	}

	query = url.Values{}
	query.Set("limit", "1")
	dataResp = fetchWidgetData(t, srv.URL, query)
	if dataResp.Total != 1 || extractID(dataResp.Data[0]) != 1 {
		t.Fatalf("expected first row id=1")
	}
	if dataResp.NextCursor != "1" {
		t.Fatalf("expected next cursor 1, got %q", dataResp.NextCursor)
	}
	if !dataResp.HasMore {
		t.Fatalf("expected has_more true")
	}

	query = url.Values{}
	query.Set("limit", "1")
	query.Set("offset", "1")
	dataResp = fetchWidgetData(t, srv.URL, query)
	if dataResp.Total != 1 || extractID(dataResp.Data[0]) != 2 {
		t.Fatalf("expected cursor row id=2")
	}
	if dataResp.NextCursor != "2" {
		t.Fatalf("expected next cursor 2, got %q", dataResp.NextCursor)
	}
	if !dataResp.HasMore {
		t.Fatalf("expected has_more true")
	}
}

func fetchWidgetData(t *testing.T, baseURL string, query url.Values) dataResponse {
	t.Helper()
	endpoint := baseURL + "/api/widgets/users_table"
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	resp, err := http.Get(endpoint)
	if err != nil {
		t.Fatalf("data request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("data status: %d", resp.StatusCode)
	}

	var payload dataResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	return payload
}

func setupSQLiteDB(t *testing.T) *sqlx.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "rapidmin.db")
	db, err := sqlx.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	schema := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT,
			email TEXT,
			age INTEGER,
			created_at TEXT,
			tag TEXT
		);
		INSERT INTO users (id, name, email, age, created_at, tag) VALUES
			(1, 'Ann', 'ann@example.com', 25, '2024-01-10', 'vip'),
			(2, 'Anna', 'anna@example.com', 31, '2024-02-11', 'active'),
			(3, 'Bob', 'bob@example.com', 45, '2023-12-01', 'vip');
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("seed db: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func sampleConfig() config.AppConfig {
	return config.AppConfig{
		Title: "Test",
		Providers: map[string]config.ProviderConfig{
			"db": {
				SQL: &config.SQLProviderConfig{Driver: "sqlite3", DSN: ":memory:"},
			},
		},
		Pages: []config.Page{
			{
				Slug:  "users",
				Title: "Users",
				Widgets: []config.Widget{
					{
						ID:    "users_table",
						Title: "Users",
						Type:  "table",
						Provider: config.ProviderSpec{
							Name: "db",
							SQL: &config.SQLSpec{
								Query: `SELECT id, name, email, age, created_at, tag FROM users ORDER BY id ASC`,
								Pagination: &config.PaginationSpec{
									Column: "id",
									Order:  "asc",
								},
							},
						},
						Table: &config.TableSpec{
							Columns: []config.ColumnSpec{
								{ID: "id", Title: "id"},
								{ID: "name", Title: "name"},
								{ID: "email", Title: "email"},
							},
							Filters: []config.FilterSpec{
								{
									ID:     "name",
									Title:  "Name",
									Type:   "text",
									Column: "name",
									Mode:   "contains",
								},
								{
									ID:        "age",
									Title:     "Age",
									Type:      "number",
									Column:    "age",
									Operators: []string{"gt"},
								},
								{
									ID:        "tags",
									Title:     "Tags",
									Type:      "select_multi",
									Column:    "tag",
									Operators: []string{"in"},
								},
								{
									ID:        "created",
									Title:     "Created",
									Type:      "date",
									Column:    "created_at",
									Operators: []string{"between"},
								},
							},
						},
					},
				},
			},
		},
	}
}

func extractID(row map[string]any) int {
	value, ok := row["id"]
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	default:
		return 0
	}
}
