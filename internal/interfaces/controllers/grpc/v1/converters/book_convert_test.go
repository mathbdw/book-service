package converters

import (
	"testing"
	"time"

	"github.com/mathbdw/book/internal/domain/entities"
	errs "github.com/mathbdw/book/internal/errors"
	pb "github.com/mathbdw/book/proto"
	"github.com/stretchr/testify/assert"
)

func TestBookAddRequestToBook(t *testing.T) {
	book := entities.Book{
		Title:       "Test",
		Description: "Desc",
		Year:        1900,
		Genre:       "Genre",
	}
	req := pb.BookAddRequest{
		Title:       book.Title,
		Description: book.Description,
		Year:        int32(book.Year),
		Genre:       book.Genre,
	}

	res := BookAddRequestToBook(&req)

	assert.Equal(t, book.Title, res.Title)
	assert.Equal(t, book.Description, res.Description)
	assert.Equal(t, book.Year, res.Year)
	assert.Equal(t, book.Genre, res.Genre)
}

func TestBookToProtoBook(t *testing.T) {
	book := entities.Book{
		ID:          1,
		Title:       "Test",
		Description: "Desc",
		Year:        1900,
		Genre:       "Genre",
	}
	pbBook := pb.Book{
		Id:          book.ID,
		Title:       book.Title,
		Description: book.Description,
		Year:        int32(book.Year),
		Genre:       book.Genre,
	}

	res := BookToProtoBook(&book)

	assert.Equal(t, pbBook.Id, res.Id)
	assert.Equal(t, pbBook.Title, res.Title)
	assert.Equal(t, pbBook.Description, res.Description)
	assert.Equal(t, pbBook.Year, res.Year)
	assert.Equal(t, pbBook.Genre, res.Genre)
}

func TestCursorPaginationToPaginationParams_Error(t *testing.T) {
	paginationParams := entities.PaginationParams{
		Limit:     1,
		SortBy:    entities.CursorTypeBookID,
		SortOrder: entities.SortOrderTypeAsc,
	}
	pbCP := pb.BookListRequest_CursorPagination{
		Cursor:    ":",
		PageSize:  paginationParams.Limit,
		SortBy:    string(paginationParams.SortBy),
		SortOrder: string(paginationParams.SortOrder),
	}

	res, err := CursorPaginationToPaginationParams(&pbCP)

	assert.Error(t, err)
	assert.Equal(t, entities.PaginationParams{}, res)
}

func TestCursorPaginationToPaginationParams_Success(t *testing.T) {
	timeCreated := time.Now()
	cursor := &entities.Cursor{
		Value:     1,
		CreatedAt: &timeCreated,
	}
	paginationParams := entities.PaginationParams{
		Limit:     1,
		Cursor:    cursor,
		SortBy:    entities.CursorTypeBookID,
		SortOrder: entities.SortOrderTypeAsc,
	}
	pbCP := pb.BookListRequest_CursorPagination{
		Cursor:    "qwer",
		PageSize:  paginationParams.Limit,
		SortBy:    string(paginationParams.SortBy),
		SortOrder: string(paginationParams.SortOrder),
	}

	res, err := CursorPaginationToPaginationParams(&pbCP)

	assert.NoError(t, err)
	assert.Equal(t, paginationParams.Limit, res.Limit)
	assert.Equal(t, paginationParams.SortBy, res.SortBy)
	assert.Equal(t, paginationParams.SortOrder, res.SortOrder)
}

func TestLevelToZerolog(t *testing.T) {
	tests := []struct {
		name     string
		expected int8
	}{
		{
			name:     "debug",
			expected: 0,
		},
		{
			name:     "info",
			expected: 1,
		},
		{
			name:     "warn",
			expected: 2,
		},
		{
			name:     "error",
			expected: 3,
		},
		{
			name:     "no_level",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, err := LevelToZerolog(tt.name)

			if err != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), errs.ErrNotFound.Error())
			} else {
				assert.Equal(t, tt.expected, level)
			}
		})
	}
}
