package handlers

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mathbdw/book/internal/domain/entities"
	errs "github.com/mathbdw/book/internal/errors"
	"github.com/mathbdw/book/mocks"
	pb "github.com/mathbdw/book/proto"
	"github.com/stretchr/testify/assert"
)

func TestBook_List_ErrorValidate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookRepo := mocks.NewMockBookRepository(ctrl)
	//createMockObservability - add_book_test.go
	observHandler := createMockHandlerObservability(ctrl)
	uc := getMockUC(ctrl, bookRepo)
	bookHandler := &BookHandler{uc: uc, observ: observHandler}
	ctx := context.Background()

	tests := []struct {
		name    string
		dataReq pb.BookListRequest_CursorPagination
		errStr  string
	}{
		{
			"pageSize",
			pb.BookListRequest_CursorPagination{
				PageSize: 1,
			},
			"value must be in list [2 10 50]",
		},
		{
			"sortBy",
			pb.BookListRequest_CursorPagination{
				PageSize: 2,
				SortBy:   "re",
			},
			"value must be in list [id title year]",
		},
		{
			"sortOrder",
			pb.BookListRequest_CursorPagination{
				PageSize:  2,
				SortBy:    "id",
				SortOrder: "re",
			},
			"value must be in list [asc desc]",
		},
	}

	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			res, err := bookHandler.List(ctx, &pb.BookListRequest{
				Pagination: &tt.dataReq,
			})

			assert.Nil(t, res)
			assert.Error(t, err)
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
			assert.Contains(t, err.Error(), tt.errStr)
		})
	}
}

func TestBook_List_ErrorCursorDecode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookRepo := mocks.NewMockBookRepository(ctrl)
	//createMockObservability - add_book_test.go
	observHandler := createMockHandlerObservability(ctrl)
	uc := getMockUC(ctrl, bookRepo)
	bookHandler := &BookHandler{uc: uc, observ: observHandler}
	ctx := context.Background()

	res, err := bookHandler.List(ctx, &pb.BookListRequest{Pagination: &pb.BookListRequest_CursorPagination{
		Cursor:    "wer",
		PageSize:  2,
		SortBy:    "id",
		SortOrder: "asc",
	}})

	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Contains(t, err.Error(), "invalid cursor")
}

func TestBook_List_ErrorUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookRepo := mocks.NewMockBookRepository(ctrl)
	//createMockObservability - add_book_test.go
	observHandler := createMockHandlerObservability(ctrl)
	uc := getMockUC(ctrl, bookRepo)
	bookHandler := &BookHandler{uc: uc, observ: observHandler}
	ctx := context.Background()

	bookRepo.EXPECT().
		List(gomock.Any(), gomock.Any()).
		Return(nil, errs.New("not found"))

	res, err := bookHandler.List(ctx, &pb.BookListRequest{Pagination: &pb.BookListRequest_CursorPagination{
		PageSize:  2,
		SortBy:    "id",
		SortOrder: "asc",
	}})

	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestBook_List_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bookRepo := mocks.NewMockBookRepository(ctrl)
	//createMockObservability - add_book_test.go
	observHandler := createMockHandlerObservability(ctrl)
	uc := getMockUC(ctrl, bookRepo)
	bookHandler := &BookHandler{uc: uc, observ: observHandler}
	ctx := context.Background()
	
	bookRepo.EXPECT().
		List(gomock.Any(), gomock.Any()).
		Return(&entities.ResponseBooks{Data: []entities.Book{{ID: 1, Title: "Test"}}, PageInfo: entities.PageInfo{}}, nil)

	protoBook, err := bookHandler.List(ctx, &pb.BookListRequest{
		Pagination: &pb.BookListRequest_CursorPagination{
			PageSize:  2,
			SortBy:    "id",
			SortOrder: "asc",
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, "", protoBook.GetPagination().GetCursorNext())
	assert.Equal(t, 1, len(protoBook.GetBooks()))
}
