package converters

import (
	"github.com/mathbdw/book/internal/domain/entities"
	errs "github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/internal/infrastructure/persistence/postgres"
	pb "github.com/mathbdw/book/proto"
)

// BookAddRequestToBook - converts pb.BookAddRequest to entities.Book.
func BookAddRequestToBook(req *pb.BookAddRequest) entities.Book {
	return entities.Book{
		Title:       req.GetTitle(),
		Description: req.Description,
		Genre:       req.GetGenre(),
		Year:        int(req.GetYear()),
	}
}

// BookToProtoBook - converts entities.Book to pb.Book
func BookToProtoBook(book *entities.Book) *pb.Book {
	return &pb.Book{
		Id:          book.ID,
		Title:       book.Title,
		Description: book.Description,
		Genre:       book.Genre,
		Year:        int32(book.Year),
	}
}

// ToPagination - converts proto BookListRequest_CursorPagination to PaginationParams entities.
func CursorPaginationToPaginationParams(req *pb.BookListRequest_CursorPagination) (entities.PaginationParams, error) {
	var cursor *entities.Cursor
	var err error
	if req.GetCursor() != "" {
		cursor, err = postgres.DecodeCursor(req.GetCursor())
		if err != nil {
			return entities.PaginationParams{}, errs.New("invalid cursor")
		}
	}

	return entities.PaginationParams{
		Limit:     req.GetPageSize(),
		Cursor:    cursor,
		SortBy:    entities.CursorType(req.GetSortBy()),
		SortOrder: entities.SortOrderType(req.GetSortOrder()),
	}, nil
}

// LevelToZerolog - converts string to int8 Level zerolog
func LevelToZerolog(level string) (int8, error) {
	switch level {
	case "debug":
		return 0, nil
	case "info":
		return 1, nil
	case "warn":
		return 2, nil
	case "error":
		return 3, nil
	default:
		return -1, errs.ErrNotFound
	}
}
