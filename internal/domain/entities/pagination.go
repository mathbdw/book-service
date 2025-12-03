package entities

import (
	"time"
)

type (
	CursorType    string
	SortOrderType string
)

const (
	CursorTypeBookID    CursorType = "id"
	CursorTypeBookTitle CursorType = "title"
	CursorTypeBookYear  CursorType = "year"

	SortOrderTypeAsc  SortOrderType = "asc"
	SortOrderTypeDesc SortOrderType = "desc"
)

// SortTypesUnique - unique fields that do not require created_at for cursorPagination
var SortTypesUnique = map[string]bool{
	string(CursorTypeBookID): true,
}

type Cursor struct {
	Value     any
	CreatedAt *time.Time
}

// PageInfo информация о странице
type PageInfo struct {
	NextCursor string `json:"next_cursor,omitempty"`
}

// PaginationParams параметры пагинации
type PaginationParams struct {
	Limit     uint64
	Cursor    *Cursor
	SortBy    CursorType
	SortOrder SortOrderType `json:"sort_order"` // "asc" или "desc"
}

type Paginatable interface {
	GetFieldAsString(field string) (string, error)
	GetCreatedAt() time.Time
}
