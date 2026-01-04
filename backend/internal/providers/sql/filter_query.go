package sql

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/ankulikov/rapidmin/internal/config"
	"github.com/ankulikov/rapidmin/internal/providers"
)

func makeFilterCond(spec config.FilterSpec, value providers.Filter, dbName string,
	typeHint *config.DataType) (sq.Sqlizer, error) {
	operator := value.Operator
	if operator == "" {
		if spec.Mode != "" {
			operator = config.FilterOperator(spec.Mode)
		} else if spec.Type == "select_multi" {
			operator = "in"
		} else if len(spec.Operators) == 1 {
			operator = spec.Operators[0]
		} else {
			operator = "eq"
		}
	}

	column := spec.Target
	switch operator {
	case config.EqOperator:
		return sq.Eq{column: value.Values[0]}, nil
	case config.GtOperator:
		return sq.Gt{column: value.Values[0]}, nil
	case config.LtOperator:
		return sq.Lt{column: value.Values[0]}, nil
	case config.AfterOperator:
		return sq.Gt{column: value.Values[0]}, nil
	case config.BeforeOperator:
		return sq.Lt{column: value.Values[0]}, nil
	case config.ContainsOperator:
		return makeContainsFilterCond(spec, value, dbName, typeHint)
	case config.BetweenOperator:
		if len(value.Values) < 2 {
			return nil, fmt.Errorf("filter operator '%s' requires at least two values", operator)
		}

		return sq.Expr(fmt.Sprintf("%s BETWEEN ? AND ?", column), value.Values[0], value.Values[1]), nil
	case config.InOperator:
		return makeInFilterCond(spec, value, dbName, typeHint)
	}

	return nil, fmt.Errorf("unknown operator in filter '%s':%s", value.Name, operator)
}

func makeContainsFilterCond(spec config.FilterSpec, value providers.Filter, dbName string,
	typeHint *config.DataType) (sq.Sqlizer, error) {
	if typeHint != nil && *typeHint == config.JsonArray {
		if dbName == "sqlite3" {
			return sq.Expr(fmt.Sprintf("EXISTS(select 1 from json_each(%s) where value like ?)",
				spec.Target), "%"+value.Values[0]+"%"), nil
		}
	}

	return sq.Expr(fmt.Sprintf("%s %s ?", spec.Target, likeOperator(dbName)), "%"+value.Values[0]+"%"), nil

}

func makeInFilterCond(spec config.FilterSpec, value providers.Filter, dbName string,
	typeHint *config.DataType) (sq.Sqlizer, error) {
	if typeHint != nil && *typeHint == config.JsonArray {
		parts := make([]string, 0, len(value.Values))
		args := make([]any, 0, len(value.Values))

		for _, val := range value.Values {
			parts = append(parts, "?")
			args = append(args, val)
		}

		if dbName == "sqlite3" {
			return sq.Expr(fmt.Sprintf("EXISTS(select 1 from json_each(%s) where value in (%s))",
				spec.Target, strings.Join(parts, ",")), args...), nil
		}
	}

	return sq.Eq{spec.Target: value.Values}, nil
}

func likeOperator(driver string) string {
	if driver == "postgres" {
		return "ILIKE"
	}
	return "LIKE"
}
