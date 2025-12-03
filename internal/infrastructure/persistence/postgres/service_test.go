package postgres

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mathbdw/book/internal/domain/entities"
)

func TestBookPagination_CreateCursor_Error(t *testing.T) {
	book := entities.Book{ID: 1, Title: "test", Year: 1900, CreatedAt: time.Now()}
	service := NewService()

	_, err := service.CreateCursor(book, "noField")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "servicePostgres.CreateCursor: field not found")
}

func TestBookPagination_CreateCursor_ErrorFailedEncodingTwoField(t *testing.T) {
	book := entities.Book{ID: 1, Title: "", Year: 1900, CreatedAt: time.Now()}
	service := NewService()

	_, err := service.CreateCursor(book, "title")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "servicePostgres.CreateCursor: failed to encode cursor")
}

func TestBookPagination_CreateCursor_ErrorFailedEncodingOneField(t *testing.T) {
	book := entities.Book{ID: 0, Title: "Test", Year: 1900, CreatedAt: time.Now()}
	service := NewService()

	_, err := service.CreateCursor(book, "id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "servicePostgres.CreateCursor: failed to encode cursor")
}

func TestBookPagination_CreateCursor_Success(t *testing.T) {
	now := time.Now()
	book := entities.Book{ID: 1, Title: "test", Year: 1900, CreatedAt: now}
	service := NewService()

	tests := []struct {
		name  string
		field string
		value string
	}{
		{"ById", "id",
			base64.StdEncoding.EncodeToString([]byte(fmt.Sprint(book.ID))),
		},
		{"ByTitle", "title",
			base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%d", book.Title, now.UnixNano()))),
		},
		{"ByYear", "year",
			base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d:%d", book.Year, now.UnixNano()))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			strCursor, err := service.CreateCursor(book, tt.field)

			assert.NoError(t, err)
			assert.Equal(t, tt.value, strCursor)
		})
	}
}

var books = []entities.Book{
	{ID: 3, Title: "Title3", Year: 1900, CreatedAt: time.Date(2021, 1, 5, 19, 0, 0, 0, time.UTC)},
	{ID: 4, Title: "Title4", Year: 1900, CreatedAt: time.Date(2021, 1, 5, 20, 0, 0, 0, time.UTC)},
	{ID: 5, Title: "Title5", Year: 1900, CreatedAt: time.Date(2021, 1, 5, 20, 0, 0, 0, time.UTC)},
	{ID: 6, Title: "Title6", Year: 1900, CreatedAt: time.Date(2021, 1, 5, 21, 0, 0, 0, time.UTC)},
}

func ConvertPaginatable(books []entities.Book) []entities.Paginatable {
	// Конвертируем в интерфейс для пагинации
	paginatable := make([]entities.Paginatable, len(books))
	for i := range books {
		paginatable[i] = books[i]
	}

	return paginatable
}

func TestPageInfo_Create_ErrorNextCursor(t *testing.T) {
	service := NewService()
	paginatables := ConvertPaginatable(books)
	pageInfo, err := service.CreatePageInfo(paginatables, entities.PaginationParams{
		Limit:     3,
		SortBy:    "title4",
		SortOrder: entities.SortOrderTypeAsc,
	})

	assert.Error(t, err, "expected error")
	assert.Contains(t, err.Error(), "servicePostgres.CreatePageInfo: error nextCursor")
	assert.Empty(t, pageInfo)
}

func TestPageInfo_Create_SuccessEmptyPrevCursor(t *testing.T) {
	service := NewService()
	paginatables := ConvertPaginatable(books)
	tests := []struct {
		name           string
		cursor         *entities.Cursor
		limit          uint64
		wantPrevCursor bool
		wantNextCursor bool
	}{
		{"emptyPrevAndNext", nil, 5, false, false},
		{"emptyPrevAndYesNext", nil, 3, false, true},
		{"yesPrevAndEmptyNext", &entities.Cursor{Value: 5}, 5, true, false},
		{"yesPrevAndNext", &entities.Cursor{Value: 5}, 3, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageInfo, err := service.CreatePageInfo(paginatables, entities.PaginationParams{
				Cursor:    tt.cursor,
				Limit:     tt.limit,
				SortBy:    "id",
				SortOrder: entities.SortOrderTypeAsc,
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.wantNextCursor, len(pageInfo.NextCursor) > 0)
		})
	}
}
