package entities

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBookPagination_GetFieldAsString_NotFound(t *testing.T) {
	book := &Book{ID: 1, Title: "test", Year: 1900, CreatedAt: time.Now()}
	_, err := book.GetFieldAsString("testFiled")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown field")
}

func TestBookPagination_GetFieldAsString_Empty(t *testing.T) {
	book := &Book{ID: 0, Title: "test", Year: 1900, CreatedAt: time.Now()}
	res, err := book.GetFieldAsString("id")

	assert.NoError(t, err)
	assert.Equal(t, "", res)
}

func TestBookPagination_GetFieldAsString_Success(t *testing.T) {
	book := Book{ID: 1, Title: "test", Year: 1900, CreatedAt: time.Now()}

	tests := []struct {
		name  string
		field string
		value any
	}{
		{"ById", "id", fmt.Sprint(book.ID)},
		{"ByTitle", "title", book.Title},
		{"ByYear", "year", fmt.Sprint(book.Year)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := book.GetFieldAsString(tt.field)

			assert.NoError(t, err)
			assert.Equal(t, tt.value, value)
		})
	}
}

func TestBookPagination_GetCreatedAt(t *testing.T) {
	book := &Book{ID: 1, Title: "test", Year: 1900, CreatedAt: time.Now()}
	res := book.GetCreatedAt()

	assert.Equal(t, book.CreatedAt, res)
}
