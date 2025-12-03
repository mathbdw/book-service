package postgres

import (
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/mathbdw/book/internal/domain/entities"
	"github.com/stretchr/testify/assert"
)

func TestBuilderQuery_ConditionBuilder_CursorNil(t *testing.T) {
	builder := sq.Select("*").From("test")

	builder = conditionBuilder(builder, entities.PaginationParams{})

	sql, _, _ := builder.ToSql()

	assert.Equal(t, "SELECT * FROM test", sql)
}

func TestBuilderQuery_ConditionBuilder_CursorCreatedNil(t *testing.T) {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("test")

	builder = conditionBuilder(builder, entities.PaginationParams{
		Cursor:    &entities.Cursor{Value: "titleTest"},
		SortBy:    entities.CursorTypeBookTitle,
		SortOrder: entities.SortOrderTypeDesc,
	})

	sql, _, _ := builder.ToSql()

	assert.Equal(t, "SELECT * FROM test WHERE title < $1", sql)
}

func TestBuilderQuery_ConditionBuilder_CursorCreatedSortOrderTypeDesc(t *testing.T) {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("test")

	date := time.Now()
	builder = conditionBuilder(builder, entities.PaginationParams{
		Cursor:    &entities.Cursor{Value: "titleTest", CreatedAt: &date},
		SortBy:    entities.CursorTypeBookTitle,
		SortOrder: entities.SortOrderTypeDesc,
	})

	sql, _, _ := builder.ToSql()

	assert.Equal(t, "SELECT * FROM test WHERE (title < $1 OR (title = $2 AND created_at < $3))", sql)
}

func TestBuilderQuery_ConditionBuilder_CursorCreated(t *testing.T) {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).Select("*").From("test")

	date := time.Now()
	builder = conditionBuilder(builder, entities.PaginationParams{
		Cursor:    &entities.Cursor{Value: "titleTest", CreatedAt: &date},
		SortBy:    entities.CursorTypeBookTitle,
		SortOrder: entities.SortOrderTypeAsc,
	})

	sql, _, _ := builder.ToSql()

	assert.Equal(t, "SELECT * FROM test WHERE (title > $1 OR (title = $2 AND created_at > $3))", sql)
}

func TestBuilderQuery_OrderByBuilder_CursorNil(t *testing.T) {
	builder := sq.Select("*").From("test")

	builder = orderByBuilder(builder, entities.PaginationParams{
		SortBy:    entities.CursorTypeBookTitle,
		SortOrder: entities.SortOrderTypeDesc,
	})

	sql, _, _ := builder.ToSql()

	assert.Equal(t, "SELECT * FROM test ORDER BY title desc", sql)
}

func TestBuilderQuery_OrderByBuilder_CursorCreatedAt(t *testing.T) {
	builder := sq.Select("*").From("test")
	date := time.Now()

	builder = orderByBuilder(builder, entities.PaginationParams{
		Cursor:    &entities.Cursor{CreatedAt: &date},
		SortBy:    entities.CursorTypeBookTitle,
		SortOrder: entities.SortOrderTypeDesc,
	})

	sql, _, _ := builder.ToSql()

	assert.Equal(t, "SELECT * FROM test ORDER BY title desc, created_at desc", sql)
}
