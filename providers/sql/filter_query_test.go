package sql

import (
	"reflect"
	"testing"

	sq "github.com/Masterminds/squirrel"

	"github.com/ankulikov/rapidmin/config"
	"github.com/ankulikov/rapidmin/providers"
)

func TestMakeFilterCondDefaults(t *testing.T) {
	tests := []struct {
		name          string
		spec          config.FilterSpec
		filter        providers.Filter
		dbName        string
		typeHint      *config.DataType
		expectedSQL   string
		expectedArgs  []any
		expectedError string
	}{
		{
			name:         "eq default",
			spec:         config.FilterSpec{ID: "name", Target: "name", Type: "text"},
			filter:       providers.Filter{Name: "name", Values: []string{"ann"}},
			dbName:       "sqlite3",
			expectedSQL:  "SELECT * FROM src WHERE name = ?",
			expectedArgs: []any{"ann"},
		},
		{
			name:         "contains from operators",
			spec:         config.FilterSpec{ID: "name", Target: "name", Type: "text", Operators: []config.FilterOperator{"contains"}},
			filter:       providers.Filter{Name: "name", Values: []string{"ann"}},
			dbName:       "sqlite3",
			expectedSQL:  "SELECT * FROM src WHERE name LIKE ?",
			expectedArgs: []any{"%ann%"},
		},
		{
			name:         "select multi defaults to in",
			spec:         config.FilterSpec{ID: "tags", Target: "tags", Type: "select_multi"},
			filter:       providers.Filter{Name: "tags", Values: []string{"vip", "active"}},
			dbName:       "sqlite3",
			expectedSQL:  "SELECT * FROM src WHERE tags IN (?,?)",
			expectedArgs: []any{"vip", "active"},
		},
		{
			name:         "operators single override",
			spec:         config.FilterSpec{ID: "age", Target: "age", Type: "number", Operators: []config.FilterOperator{"gt"}},
			filter:       providers.Filter{Name: "age", Values: []string{"18"}},
			dbName:       "sqlite3",
			expectedSQL:  "SELECT * FROM src WHERE age > ?",
			expectedArgs: []any{"18"},
		},
		{
			name:          "between missing value",
			spec:          config.FilterSpec{ID: "created", Target: "created_at", Type: "date"},
			filter:        providers.Filter{Name: "created", Operator: "between", Values: []string{"2024-01-01"}},
			dbName:        "sqlite3",
			expectedError: "filter operator 'between' requires at least two values",
		},
		{
			name:          "unknown operator",
			spec:          config.FilterSpec{ID: "score", Target: "score", Type: "number"},
			filter:        providers.Filter{Name: "score", Operator: "nope", Values: []string{"10"}},
			dbName:        "sqlite3",
			expectedError: "unknown operator in filter 'score':nope",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cond, err := makeFilterCond(tc.spec, tc.filter, tc.dbName, tc.typeHint)
			if tc.expectedError != "" {
				if err == nil || err.Error() != tc.expectedError {
					t.Fatalf("expected error %q, got %v", tc.expectedError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			query, args, err := sq.Select("*").From("src").Where(cond).PlaceholderFormat(sq.Question).ToSql()
			if err != nil {
				t.Fatalf("unexpected sql error: %v", err)
			}

			if query != tc.expectedSQL {
				t.Fatalf("expected query %q, got %q", tc.expectedSQL, query)
			}
			if !reflect.DeepEqual(args, tc.expectedArgs) {
				t.Fatalf("expected args %v, got %v", tc.expectedArgs, args)
			}
		})
	}
}

func TestMakeFilterCondOperators(t *testing.T) {
	tests := []struct {
		name         string
		spec         config.FilterSpec
		filter       providers.Filter
		dbName       string
		expectedSQL  string
		expectedArgs []any
	}{
		{
			name:         "eq",
			spec:         config.FilterSpec{ID: "age", Target: "age", Type: "number"},
			filter:       providers.Filter{Name: "age", Operator: "eq", Values: []string{"10"}},
			dbName:       "sqlite3",
			expectedSQL:  "SELECT * FROM src WHERE age = ?",
			expectedArgs: []any{"10"},
		},
		{
			name:         "gt",
			spec:         config.FilterSpec{ID: "age", Target: "age", Type: "number"},
			filter:       providers.Filter{Name: "age", Operator: "gt", Values: []string{"10"}},
			dbName:       "sqlite3",
			expectedSQL:  "SELECT * FROM src WHERE age > ?",
			expectedArgs: []any{"10"},
		},
		{
			name:         "lt",
			spec:         config.FilterSpec{ID: "age", Target: "age", Type: "number"},
			filter:       providers.Filter{Name: "age", Operator: "lt", Values: []string{"10"}},
			dbName:       "sqlite3",
			expectedSQL:  "SELECT * FROM src WHERE age < ?",
			expectedArgs: []any{"10"},
		},
		{
			name:         "before",
			spec:         config.FilterSpec{ID: "created", Target: "created_at", Type: "date"},
			filter:       providers.Filter{Name: "created", Operator: "before", Values: []string{"2024-01-01"}},
			dbName:       "sqlite3",
			expectedSQL:  "SELECT * FROM src WHERE created_at < ?",
			expectedArgs: []any{"2024-01-01"},
		},
		{
			name:         "after",
			spec:         config.FilterSpec{ID: "created", Target: "created_at", Type: "date"},
			filter:       providers.Filter{Name: "created", Operator: "after", Values: []string{"2024-01-01"}},
			dbName:       "sqlite3",
			expectedSQL:  "SELECT * FROM src WHERE created_at > ?",
			expectedArgs: []any{"2024-01-01"},
		},
		{
			name:         "contains sqlite",
			spec:         config.FilterSpec{ID: "name", Target: "name", Type: "text"},
			filter:       providers.Filter{Name: "name", Operator: "contains", Values: []string{"ann"}},
			dbName:       "sqlite3",
			expectedSQL:  "SELECT * FROM src WHERE name LIKE ?",
			expectedArgs: []any{"%ann%"},
		},
		{
			name:         "contains postgres",
			spec:         config.FilterSpec{ID: "name", Target: "name", Type: "text"},
			filter:       providers.Filter{Name: "name", Operator: "contains", Values: []string{"ann"}},
			dbName:       "postgres",
			expectedSQL:  "SELECT * FROM src WHERE name ILIKE ?",
			expectedArgs: []any{"%ann%"},
		},
		{
			name:         "between",
			spec:         config.FilterSpec{ID: "created", Target: "created_at", Type: "date"},
			filter:       providers.Filter{Name: "created", Operator: "between", Values: []string{"2024-01-01", "2024-01-31"}},
			dbName:       "sqlite3",
			expectedSQL:  "SELECT * FROM src WHERE created_at BETWEEN ? AND ?",
			expectedArgs: []any{"2024-01-01", "2024-01-31"},
		},
		{
			name:         "datetime between",
			spec:         config.FilterSpec{ID: "created", Target: "created_at", Type: "datetime"},
			filter:       providers.Filter{Name: "created", Operator: "between", Values: []string{"1710000000", "1710003600"}},
			dbName:       "sqlite3",
			expectedSQL:  "SELECT * FROM src WHERE created_at BETWEEN ? AND ?",
			expectedArgs: []any{int64(1710000000), int64(1710003600)},
		},
		{
			name:         "in",
			spec:         config.FilterSpec{ID: "tags", Target: "tags", Type: "select_multi"},
			filter:       providers.Filter{Name: "tags", Operator: "in", Values: []string{"vip", "active"}},
			dbName:       "sqlite3",
			expectedSQL:  "SELECT * FROM src WHERE tags IN (?,?)",
			expectedArgs: []any{"vip", "active"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cond, err := makeFilterCond(tc.spec, tc.filter, tc.dbName, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			query, args, err := sq.Select("*").From("src").Where(cond).PlaceholderFormat(sq.Question).ToSql()
			if err != nil {
				t.Fatalf("unexpected sql error: %v", err)
			}

			if query != tc.expectedSQL {
				t.Fatalf("expected query %q, got %q", tc.expectedSQL, query)
			}
			if !reflect.DeepEqual(args, tc.expectedArgs) {
				t.Fatalf("expected args %v, got %v", tc.expectedArgs, args)
			}
		})
	}
}

func TestMakeFilterCondJSONArraySQLite(t *testing.T) {
	typeHint := config.JsonArray
	spec := config.FilterSpec{ID: "tags", Target: "tags", Type: "select_multi"}

	cond, err := makeFilterCond(spec, providers.Filter{Name: "tags", Values: []string{"vip", "active"}}, "sqlite3", &typeHint)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query, args, err := sq.Select("*").From("src").Where(cond).PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		t.Fatalf("unexpected sql error: %v", err)
	}

	expectedSQL := "SELECT * FROM src WHERE EXISTS(select 1 from json_each(tags) where value in (?,?))"
	if query != expectedSQL {
		t.Fatalf("expected query %q, got %q", expectedSQL, query)
	}
	expectedArgs := []any{"vip", "active"}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Fatalf("expected args %v, got %v", expectedArgs, args)
	}
}

func TestMakeFilterCondJSONArrayContainsSQLite(t *testing.T) {
	typeHint := config.JsonArray
	spec := config.FilterSpec{ID: "tags", Target: "tags", Type: "text"}

	cond, err := makeFilterCond(spec, providers.Filter{Name: "tags", Operator: "contains", Values: []string{"vip"}}, "sqlite3", &typeHint)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query, args, err := sq.Select("*").From("src").Where(cond).PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		t.Fatalf("unexpected sql error: %v", err)
	}

	expectedSQL := "SELECT * FROM src WHERE EXISTS(select 1 from json_each(tags) where value like ?)"
	if query != expectedSQL {
		t.Fatalf("expected query %q, got %q", expectedSQL, query)
	}
	expectedArgs := []any{"%vip%"}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Fatalf("expected args %v, got %v", expectedArgs, args)
	}
}
