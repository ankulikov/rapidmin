package sql

import (
	"reflect"
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"

	"github.com/ankulikov/rapidmin/internal/config"
	"github.com/ankulikov/rapidmin/internal/providers"
)

func TestBuildFilterConditions(t *testing.T) {
	widget := config.Widget{
		Table: &config.TableSpec{
			Filters: []config.FilterSpec{
				{ID: "age", Target: "age", Type: "number", Mode: "gt"},
				{ID: "tags", Target: "tag", Type: "select_multi"},
				{ID: "score", Target: "score", Type: "number", Operators: []string{"lt"}},
				{ID: "name", Target: "name", Type: "text"},
				{ID: "desc", Target: "description", Type: "text"},
				{ID: "range", Target: "created_at", Type: "date"},
				{ID: "skip", Target: ""},
			},
		},
	}
	filters := []providers.Filter{
		{Name: "age", Values: []string{"18"}},
		{Name: "tags", Values: []string{"vip", "active"}},
		{Name: "score", Values: []string{"90"}},
		{Name: "name", Operator: "contains", Values: []string{"bob"}},
		{Name: "desc", Values: []string{"foo"}},
		{Name: "range", Operator: "between", Values: []string{"2024-01-01", "2024-01-31"}},
		{Name: "skip", Values: []string{"x"}},
		{Name: "missing", Values: []string{"x"}},
	}

	conds := buildFilterConditions(widget, filters, "ILIKE")
	query, args, err := sq.Select("*").From("src").Where(sq.And(conds)).PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedQuery := "SELECT * FROM src WHERE (age > ? AND tag IN (?,?) AND score < ? AND name ILIKE ? AND description = ? AND created_at BETWEEN ? AND ?)"
	if query != expectedQuery {
		t.Fatalf("expected query %q, got %q", expectedQuery, query)
	}

	expectedArgs := []any{
		"18",
		"vip",
		"active",
		"90",
		"%bob%",
		"foo",
		"2024-01-01",
		"2024-01-31",
	}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Fatalf("expected args %v, got %v", expectedArgs, args)
	}
}

func TestBuildFilterConditionsBetweenMissingValue(t *testing.T) {
	widget := config.Widget{
		Table: &config.TableSpec{
			Filters: []config.FilterSpec{
				{ID: "range", Target: "created_at", Type: "date"},
			},
		},
	}
	filters := []providers.Filter{
		{Name: "range", Operator: "between", Values: []string{"2024-01-01"}},
	}

	conds := buildFilterConditions(widget, filters, "LIKE")

	if len(conds) != 0 {
		t.Fatalf("expected no conds, got %v", conds)
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
				{ID: "name", Target: "name", Type: "text", Mode: "contains"},
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

	query, args, err := buildQuery(widget, req, "ILIKE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
