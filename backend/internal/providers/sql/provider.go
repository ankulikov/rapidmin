package sql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/ankulikov/rapidmin/internal/config"
	"github.com/ankulikov/rapidmin/internal/providers"
)

type Provider struct {
	db *sqlx.DB
}

func New() *Provider {
	return &Provider{}
}

func NewWithDB(db *sqlx.DB) *Provider {
	return &Provider{db: db}
}

func (p *Provider) Init(ctx context.Context, name string, providerConfig config.ProviderConfig) (err error) {
	if p.db != nil {
		return nil
	}

	if providerConfig.SQL == nil {
		return fmt.Errorf("sql provider %s missing config", name)
	}

	driver := strings.TrimSpace(providerConfig.SQL.Driver)
	dsn := strings.TrimSpace(providerConfig.SQL.DSN)
	if driver == "" {
		return fmt.Errorf("sql provider %s missing driver", name)
	}
	if dsn == "" {
		return fmt.Errorf("sql provider %s missing dsn", name)
	}

	p.db, err = sqlx.Open(driver, dsn)

	return err
}

func (p *Provider) Fetch(ctx context.Context, widget config.Widget, req providers.DataRequest) (providers.DataResponse, error) {
	if p.db == nil {
		return providers.DataResponse{}, errors.New("sql provider not configured")
	}
	if widget.Provider.SQL == nil {
		return providers.DataResponse{}, errors.New("sql provider missing query")
	}

	query, args, err := buildQuery(widget, req, likeOperator(p.db.DriverName()))
	if err != nil {
		return providers.DataResponse{}, err
	}
	query = p.db.Rebind(query)
	rows, err := p.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return providers.DataResponse{}, err
	}
	defer rows.Close()

	data := make([]map[string]any, 0)
	nextCursor := ""
	hasMore := false
	for rows.Next() {
		row := map[string]any{}
		if err := rows.MapScan(row); err != nil {
			return providers.DataResponse{}, err
		}
		normalizeRow(row)
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return providers.DataResponse{}, err
	}

	if req.Limit > 0 && len(data) > req.Limit {
		hasMore = true
		data = data[:req.Limit]
	}

	if len(data) > 0 {
		cursorColumn := paginationColumn(widget.Provider.SQL.Pagination)
		if cursorColumn != "" {
			nextCursor = cursorValue(data[len(data)-1], cursorColumn)
		}
	}

	return providers.DataResponse{
		Data:       data,
		Total:      len(data),
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func buildQuery(widget config.Widget, req providers.DataRequest, likeOp string) (string, []any, error) {
	trimmed := strings.TrimSuffix(strings.TrimSpace(widget.Provider.SQL.Query), ";")
	base, baseOrderBy := splitByOrderBy(trimmed)
	builder := sq.Select("*").From("(" + base + ") AS src")

	conds := buildFilterConditions(widget, req.Filters, likeOp)
	for _, cond := range conds {
		builder = builder.Where(cond)
	}

	paginationCond, paginationOrder := buildPagination(widget.Provider.SQL.Pagination, req.Cursor)
	if paginationCond != nil {
		builder = builder.Where(paginationCond)
	}

	orderBy := trimOrderByPrefix(baseOrderBy)
	if paginationOrder != "" {
		orderBy = paginationOrder
	}
	if orderBy != "" {
		builder = builder.OrderBy(orderBy)
	}

	if req.Limit > 0 {
		builder = builder.Limit(uint64(req.Limit + 1))
	}

	builder = builder.PlaceholderFormat(sq.Question)
	query, args, err := builder.ToSql()
	if err != nil {
		return "", nil, err
	}
	return query, args, nil
}

func buildFilterConditions(widget config.Widget, filters []providers.Filter, likeOp string) []sq.Sqlizer {
	if widget.Table == nil {
		return nil
	}

	filterIndex := map[string]config.FilterSpec{}
	for _, filter := range widget.Table.Filters {
		filterIndex[filter.ID] = filter
	}

	conds := make([]sq.Sqlizer, 0, len(filters))
	for _, filter := range filters {
		spec, ok := filterIndex[filter.Name]
		if !ok || spec.Target == "" {
			continue
		}

		operator := filter.Operator
		if operator == "" {
			if spec.Mode != "" {
				operator = spec.Mode
			} else if spec.Type == "select_multi" {
				operator = "in"
			} else if len(spec.Operators) == 1 {
				operator = spec.Operators[0]
			} else {
				operator = "eq"
			}
		}

		if len(filter.Values) == 0 {
			continue
		}

		col := spec.Target
		switch operator {
		case "eq":
			conds = append(conds, sq.Eq{col: filter.Values[0]})
		case "gt":
			conds = append(conds, sq.Gt{col: filter.Values[0]})
		case "lt":
			conds = append(conds, sq.Lt{col: filter.Values[0]})
		case "before":
			conds = append(conds, sq.Lt{col: filter.Values[0]})
		case "after":
			conds = append(conds, sq.Gt{col: filter.Values[0]})
		case "contains":
			conds = append(conds, sq.Expr(fmt.Sprintf("%s %s ?", col, likeOp), "%"+filter.Values[0]+"%"))
		case "between":
			if len(filter.Values) < 2 {
				continue
			}
			conds = append(conds, sq.Expr(fmt.Sprintf("%s BETWEEN ? AND ?", col), filter.Values[0], filter.Values[1]))
		case "in":
			if len(filter.Values) == 0 {
				continue
			}
			conds = append(conds, sq.Eq{col: filter.Values})
		}
	}

	return conds
}

func likeOperator(driver string) string {
	if driver == "postgres" {
		return "ILIKE"
	}
	return "LIKE"
}

func buildPagination(pagination *config.PaginationSpec, cursor string) (sq.Sqlizer, string) {
	if pagination == nil || pagination.Column == "" || cursor == "" {
		return nil, ""
	}

	order := strings.ToLower(strings.TrimSpace(pagination.Order))
	if order != "desc" {
		order = "asc"
	}

	operator := ">"
	if order == "desc" {
		operator = "<"
	}

	cond := sq.Expr(fmt.Sprintf("%s %s ?", pagination.Column, operator), cursor)
	orderBy := fmt.Sprintf("%s %s", pagination.Column, strings.ToUpper(order))
	return cond, orderBy
}

func paginationColumn(pagination *config.PaginationSpec) string {
	if pagination == nil {
		return ""
	}
	return pagination.Column
}

func trimOrderByPrefix(orderBy string) string {
	trimmed := strings.TrimSpace(orderBy)
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "order by ") {
		return strings.TrimSpace(trimmed[len("order by "):])
	}
	return trimmed
}

func splitByOrderBy(query string) (string, string) {
	lower := strings.ToLower(query)
	depth := 0
	orderIdx := -1
	for i := 0; i < len(lower); i++ {
		switch lower[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		default:
			if depth != 0 {
				continue
			}
			if strings.HasPrefix(lower[i:], " order by ") {
				orderIdx = i + 1
			}
		}
	}

	if orderIdx == -1 {
		return query, ""
	}
	return strings.TrimSpace(query[:orderIdx-1]), strings.TrimSpace(query[orderIdx:])
}

func normalizeRow(row map[string]any) {
	for key, value := range row {
		if bytes, ok := value.([]byte); ok {
			row[key] = string(bytes)
		}
	}
}

func cursorValue(row map[string]any, column string) string {
	value, ok := row[column]
	if !ok || value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
