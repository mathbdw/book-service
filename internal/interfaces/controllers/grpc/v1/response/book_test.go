package response

import (
	"testing"

	"github.com/mathbdw/book/internal/interfaces/controllers/grpc/v1/converters"
	"github.com/mathbdw/book/internal/domain/entities"
	"github.com/stretchr/testify/assert"
)

func TestBook_ConvertBookToProtoBook(t *testing.T) {
	book := entities.Book{ID: 1, Title: "Title"}
	resp := converters.BookToProtoBook(&book)

	assert.Equal(t, book.ID, resp.Id)
	assert.Equal(t, book.Title, resp.Title)
	assert.Equal(t, book.Description, resp.Description)
	assert.Equal(t, int32(book.Year), resp.Year)
	assert.Equal(t, book.Genre, resp.Genre)
}

func TestBook_GetBooksResponse(t *testing.T) {
	books := []entities.Book{
		{ID: 1, Title: "Test Title", Description: "Test Desc", Year: 1900, Genre: "Test Genre"},
	}

	res := GetBooksResponse(books)

	assert.Equal(t, 1, len(res.GetBook()))
}

func TestBook_GetListResponse(t *testing.T) {
	respBook := entities.ResponseBooks{
		PageInfo: entities.PageInfo{
			NextCursor: "adf",
		},
		Data: []entities.Book{{ID: 1, Title: "Test Title", Description: "Test Desc", Year: 1900, Genre: "Test Genre"}},
	}

	resProto := GetListResponse(&respBook)
	assert.Equal(t, respBook.PageInfo.NextCursor, resProto.GetPagination().GetCursorNext())
	assert.Equal(t, 1, len(resProto.GetBooks()))
}
