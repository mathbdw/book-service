package response

import (
	"github.com/mathbdw/book/internal/interfaces/controllers/grpc/v1/converters"
	"github.com/mathbdw/book/internal/domain/entities"
	pb "github.com/mathbdw/book/proto"
)

// GetBooksResponse - Sets *pb.BooksResponse from slice entities.Book
func GetBooksResponse(books []entities.Book) *pb.BooksResponse {
	res := make([]*pb.Book, 0, len(books))
	for _, book := range books {
		res = append(res, converters.BookToProtoBook(&book))
	}

	return &pb.BooksResponse{Book: res}
}

// GetBooks - Sets *pb.Book from slice entities.Book
func GetBooks(books []entities.Book) []*pb.Book {
	res := make([]*pb.Book, 0, len(books))
	for _, book := range books {
		res = append(res, converters.BookToProtoBook(&book))
	}

	return res
}

// GetListResponse - Sets *pb.BookListResponse from *entities.ResponseBooks
func GetListResponse(respBook *entities.ResponseBooks) *pb.BookListResponse {
	return &pb.BookListResponse{
		Pagination: &pb.BookListResponse_CursorPagination{
			CursorNext: respBook.PageInfo.NextCursor,
		},
		Books: GetBooks(respBook.Data),
	}
}
