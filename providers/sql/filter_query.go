package sql

import (
	"fmt"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/ankulikov/rapidmin/config"
	"github.com/ankulikov/rapidmin/providers"
)

func makeFilterCond(spec config.FilterSpec, value providers.Filter, dbName string,
	typeHint *config.DataType) (sq.Sqlizer, error) {
	operator := value.Operator

	if operator == "" {
		if len(spec.Operators) == 1 {
			operator = spec.Operators[0]
		} else if spec.Type == "select_multi" {
			operator = "in"
		} else {
			operator = "eq"
		}
	}

	vals := make([]any, 0, len(value.Values))

	if spec.Type == "datetime" {
		tsValues, err := parseUnixValues(value.Values)
		if err != nil {
			return nil, err
		}

		for _, tsValue := range tsValues {
			vals = append(vals, tsValue)
		}
	} else {
		for _, s := range value.Values {
			vals = append(vals, s)
		}
	}

	column := spec.Target
	switch operator {
	case config.EqOperator:
		return sq.Eq{column: vals[0]}, nil
	case config.GtOperator:
		return sq.Gt{column: vals[0]}, nil
	case config.LtOperator:
		return sq.Lt{column: vals[0]}, nil
	case config.AfterOperator:
		return sq.Gt{column: vals[0]}, nil
	case config.BeforeOperator:
		return sq.Lt{column: vals[0]}, nil
	case config.ContainsOperator:
		return makeContainsFilterCond(spec, vals, dbName, typeHint)
	case config.BetweenOperator:
		if len(vals) < 2 {
			return nil, fmt.Errorf("filter operator '%s' requires at least two values", operator)
		}

		return sq.Expr(fmt.Sprintf("%s BETWEEN ? AND ?", column), vals[0], vals[1]), nil
	case config.InOperator:
		return makeInFilterCond(spec, vals, dbName, typeHint)
	}

	return nil, fmt.Errorf("unknown operator in filter '%s':%s", value.Name, operator)
}

func parseUnixValues(values []string) ([]int64, error) {
	parsed := make([]int64, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return nil, fmt.Errorf("invalid unix timestamp %q", value)
		}
		number, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid unix timestamp %q", value)
		}
		parsed = append(parsed, number)
	}

	return parsed, nil
}

func makeContainsFilterCond(spec config.FilterSpec, values []any, dbName string,
	typeHint *config.DataType) (sq.Sqlizer, error) {
	if typeHint != nil && *typeHint == config.JsonArray {
		if dbName == "sqlite3" {
			return sq.Expr(
				fmt.Sprintf("EXISTS(select 1 from json_each(%s) where value like ?)", spec.Target),
				fmt.Sprintf("%%%v%%", values[0]),
			), nil
		}
	}

	return sq.Expr(
		fmt.Sprintf("%s %s ?", spec.Target, likeOperator(dbName)),
		fmt.Sprintf("%%%v%%", values[0]),
	), nil

}

func makeInFilterCond(spec config.FilterSpec, vals []any, dbName string,
	typeHint *config.DataType) (sq.Sqlizer, error) {
	if typeHint != nil && *typeHint == config.JsonArray {
		parts := make([]string, 0, len(vals))

		for range vals {
			parts = append(parts, "?")
		}

		if dbName == "sqlite3" {
			return sq.Expr(fmt.Sprintf("EXISTS(select 1 from json_each(%s) where value in (%s))",
				spec.Target, strings.Join(parts, ",")), vals...), nil
		}
	}

	return sq.Eq{spec.Target: vals}, nil
}

func likeOperator(driver string) string {
	if driver == "postgres" {
		return "ILIKE"
	}
	return "LIKE"
}
