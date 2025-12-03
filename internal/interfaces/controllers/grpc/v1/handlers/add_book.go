package handlers

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/mathbdw/book/internal/interfaces/controllers/grpc/v1/converters"
	"github.com/mathbdw/book/internal/interfaces/observability"
	pb "github.com/mathbdw/book/proto"
)

// Add - creates a new book based on data from a gRPC request.
// Returns:
// - *emptypb.Empty: empty response on successful creation
// - error: validation or business logic error
//
// Errors:
// - codes.InvalidArgument: input data validation error
// - codes.Internal: database or usecase level error
//
// Logging:
// - Info level: validation and business logic errors
func (bh *BookHandler) Add(ctx context.Context, req *pb.BookAddRequest) (*emptypb.Empty, error) {
	start := time.Now()
	logger := bh.observ.WithContext(ctx)
	ctx, span := bh.observ.StartSpan(ctx, "v1.BookService.Add")
	span.SetAttributes([]observability.Attribute{
		{Key: "http.method", Value: "POST"},
		{Key: "http.route", Value: "v1/books"},
	})
	defer span.End()

	var statusCode codes.Code = codes.OK
	defer func() {
		duration := time.Since(start).Seconds()
		bh.observ.RecordHanderRequest(ctx, "POST", "v1/books", int(statusCode), duration)
	}()

	if err := req.Validate(); err != nil {
		logger.Info("grpcBook.Add: validate", map[string]any{"error": err.Error()})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "validation.failed", Value: true}})
		statusCode = codes.InvalidArgument

		return nil, status.Error(statusCode, err.Error())
	}

	book := converters.BookAddRequestToBook(req)
	span.SetAttributes([]observability.Attribute{
		{Key: "book.title", Value: book.Title},
		{Key: "book.description", Value: book.Description},
		{Key: "book.year", Value: book.Year},
		{Key: "book.genre", Value: book.Genre},
	})

	err := bh.uc.Add.Execute(ctx, book)
	if err != nil {
		logger.Info("grpcBook.Add: usecase", map[string]any{"error": err.Error()})

		span.SetAttributes([]observability.Attribute{{Key: "usecase.failed", Value: true}})
		statusCode = codes.Internal

		return nil, status.Error(statusCode, err.Error())
	}

	return &emptypb.Empty{}, nil
}
