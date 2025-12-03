package postgres

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/mathbdw/book/internal/domain/entities"
)

// conditionBuilder - SelectBuilder query condition builder
func conditionBuilder(query sq.SelectBuilder, params entities.PaginationParams) sq.SelectBuilder {
	if params.Cursor != nil {
		if params.Cursor.CreatedAt == nil {
			if params.SortOrder == entities.SortOrderTypeAsc {
				return query.Where(sq.Gt{string(params.SortBy): params.Cursor.Value})
			} else {
				return query.Where(sq.Lt{string(params.SortBy): params.Cursor.Value})
			}
		} else {
			if params.SortOrder == entities.SortOrderTypeAsc {
				return query.Where(
					sq.Or{
						sq.Gt{string(params.SortBy): params.Cursor.Value},
						sq.And{
							sq.Eq{string(params.SortBy): params.Cursor.Value},
							sq.Gt{"created_at": params.Cursor.CreatedAt.UTC()},
						},
					},
				)
			} else {
				return query.Where(
					sq.Or{
						sq.Lt{string(params.SortBy): params.Cursor.Value},
						sq.And{
							sq.Eq{string(params.SortBy): params.Cursor.Value},
							sq.Lt{"created_at": params.Cursor.CreatedAt.UTC()},
						},
					},
				)
			}
		}
	}

	return query
}

// orderByBuilder - SelectBuilder Query Sort Builder
func orderByBuilder(query sq.SelectBuilder, params entities.PaginationParams) sq.SelectBuilder {
	var strOrder string

	if params.Cursor != nil && params.Cursor.CreatedAt != nil {
		strOrder = fmt.Sprintf(", created_at %s", params.SortOrder)
	}

	return query.OrderBy(fmt.Sprintf("%s %s%s", params.SortBy, params.SortOrder, strOrder))
}
