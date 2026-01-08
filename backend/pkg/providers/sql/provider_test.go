package sql

import (
	"reflect"
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/require"

	"github.com/ankulikov/rapidmin/backend/pkg/config"
	"github.com/ankulikov/rapidmin/backend/pkg/providers"
)

func TestBuildFilterConditions(t *testing.T) {
	widget := config.Widget{
		Table: &config.TableSpec{
			Filters: []config.FilterSpec{
				{ID: "name", Target: "name", Type: "text", Operators: []config.FilterOperator{"contains"}},
				{ID: "skip", Target: ""},
			},
		},
		Provider: config.ProviderSpec{
			SQL: &config.SQLSpec{
				Types: map[string]config.DataType{},
			}},
	}
	filters := []providers.Filter{
		{Name: "name", Values: []string{"bob"}},
		{Name: "skip", Values: []string{"x"}},
		{Name: "missing", Values: []string{"x"}},
	}

	conds, err := buildFilterConditions(widget, filters, "postgres")
	require.NoError(t, err)

	query, args, err := sq.Select("*").From("src").Where(sq.And(conds)).PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedQuery := "SELECT * FROM src WHERE (name ILIKE ?)"
	if query != expectedQuery {
		t.Fatalf("expected query %q, got %q", expectedQuery, query)
	}

	expectedArgs := []any{
		"%bob%",
	}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Fatalf("expected args %v, got %v", expectedArgs, args)
	}
}

func TestBuildQuery(t *testing.T) {
	widget := config.Widget{
		Provider: config.ProviderSpec{
			SQL: &config.SQLSpec{
				Query: "SELECT id, name FROM users ORDER BY id ASC",
				Pagination: &config.PaginationSpec{
					Column: "created_at",
					Order:  "desc",
				},
			},
		},
		Table: &config.TableSpec{
			Filters: []config.FilterSpec{
				{ID: "name", Target: "name", Type: "text", Operators: []config.FilterOperator{"contains"}},
			},
		},
	}
	req := providers.DataRequest{
		Limit:  10,
		Cursor: "2024-03-01",
		Filters: []providers.Filter{
			{Name: "name", Values: []string{"bob"}},
		},
	}

	query, args, err := buildQuery(widget, req, "postgres")
	require.NoError(t, err)

	if !strings.HasPrefix(query, "SELECT * FROM (SELECT id, name FROM users) AS src") {
		t.Fatalf("unexpected query: %q", query)
	}
	if !strings.Contains(query, "WHERE name ILIKE ? AND created_at < ?") {
		t.Fatalf("missing filters or pagination in query: %q", query)
	}
	if !strings.Contains(query, "ORDER BY created_at DESC") {
		t.Fatalf("missing order by in query: %q", query)
	}
	if !strings.Contains(query, "LIMIT 11") {
		t.Fatalf("missing limit in query: %q", query)
	}

	expectedArgs := []any{"%bob%", "2024-03-01"}
	if !reflect.DeepEqual(args[:2], expectedArgs) {
		t.Fatalf("expected args %v, got %v", expectedArgs, args[:2])
	}
}

func TestBuildPagination(t *testing.T) {
	pagination := &config.PaginationSpec{Column: "created_at", Order: "desc"}
	cond, orderBy := buildPagination(pagination, "2024-02-01")
	query, args, err := sq.Select("*").From("src").Where(cond).PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedQuery := "SELECT * FROM src WHERE created_at < ?"
	if query != expectedQuery {
		t.Fatalf("expected query %q, got %q", expectedQuery, query)
	}
	if len(args) != 1 || args[0] != "2024-02-01" {
		t.Fatalf("expected cursor arg, got %v", args)
	}
	if orderBy != "created_at DESC" {
		t.Fatalf("expected orderBy, got %q", orderBy)
	}

	pagination = &config.PaginationSpec{Column: "created_at", Order: "invalid"}
	cond, orderBy = buildPagination(pagination, "1")
	query, args, err = sq.Select("*").From("src").Where(cond).PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedQuery = "SELECT * FROM src WHERE created_at > ?"
	if query != expectedQuery {
		t.Fatalf("expected query %q, got %q", expectedQuery, query)
	}
	if len(args) != 1 || args[0] != "1" {
		t.Fatalf("expected cursor arg, got %v", args)
	}
	if orderBy != "created_at ASC" {
		t.Fatalf("expected asc orderBy, got %q", orderBy)
	}
}

func TestNormalizeRowJSONType(t *testing.T) {
	row := map[string]any{
		"tags":  []byte(`["vip","active"]`),
		"attrs": `{"tier":2}`,
		"name":  []byte(`Ann`),
	}
	types := map[string]config.DataType{
		"tags":  config.JsonArray,
		"attrs": config.JsonArray,
	}

	normalizeRow(row, types)

	tags, ok := row["tags"].([]any)
	if !ok || len(tags) != 2 || tags[0] != "vip" {
		t.Fatalf("expected parsed tags array, got %T: %v", row["tags"], row["tags"])
	}

	attrs, ok := row["attrs"].(map[string]any)
	if !ok || attrs["tier"] != float64(2) {
		t.Fatalf("expected parsed attrs object, got %T: %v", row["attrs"], row["attrs"])
	}

	if row["name"] != "Ann" {
		t.Fatalf("expected name to be string, got %T: %v", row["name"], row["name"])
	}
}
