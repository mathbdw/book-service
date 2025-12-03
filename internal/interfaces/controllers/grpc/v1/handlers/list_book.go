package handlers

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mathbdw/book/internal/interfaces/observability"
	"github.com/mathbdw/book/internal/interfaces/controllers/grpc/v1/converters"
	"github.com/mathbdw/book/internal/interfaces/controllers/grpc/v1/response"
	pb "github.com/mathbdw/book/proto"
)

// List - returns a list of books based on data from a gRPC request.
// Returns:
// - *pb.BookListResponse: empty response on successful creation
// - error: validation or business logic error
//
// Errors:
// - codes.InvalidArgument: input data validation error
// - codes.Internal: database or usecase level error
//
// Logging:
// - Info level: validation and business logic errors
func (bh *BookHandler) List(ctx context.Context, req *pb.BookListRequest) (*pb.BookListResponse, error) {
	start := time.Now()
	logger := bh.observ.WithContext(ctx)
	ctx, span := bh.observ.StartSpan(ctx, "v1.BookService.List")
	span.SetAttributes([]observability.Attribute{
		{Key: "http.method", Value: "GET"},
		{Key: "http.methrouteod", Value: "v1/book-list"},
	})

	defer span.End()

	var statusCode codes.Code = codes.OK
	defer func() {
		duration := time.Since(start).Seconds()
		bh.observ.RecordHanderRequest(ctx, "GET", "v1/book-list", int(statusCode), duration)
	}()

	if err := req.Validate(); err != nil {
		logger.Info("grpcBook.List: validate", map[string]any{"error": err.Error()})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "validation.failed", Value: true}})
		statusCode = codes.InvalidArgument

		return nil, status.Error(statusCode, err.Error())
	}

	params, err := converters.CursorPaginationToPaginationParams(req.GetPagination())
	if err != nil {
		logger.Info("grpcBook.List: validate cursor", map[string]any{"error": err.Error()})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "validation_cursor.failed", Value: true}})
		statusCode = codes.InvalidArgument

		return nil, status.Error(statusCode, "invalid cursor")
	}

	span.SetAttributes([]observability.Attribute{
		{Key: "limit", Value: int64(params.Limit)},
		{Key: "sort_by", Value: string(params.SortBy)},
		{Key: "sort_order", Value: string(params.SortOrder)},
	})

	if params.Cursor != nil {
		span.SetAttributes([]observability.Attribute{{Key: "cursor.value", Value: fmt.Sprintf("%v", params.Cursor.Value)}})
		if params.Cursor.CreatedAt != nil {
			span.SetAttributes([]observability.Attribute{{Key: "cursor.created_at_unix", Value: params.Cursor.CreatedAt.Unix()}})
		}
	}

	respBook, err := bh.uc.List.Execute(ctx, params)
	if err != nil {
		logger.Info("grpcBook.List: usecase", map[string]any{"error": err.Error()})
		span.SetAttributes([]observability.Attribute{{Key: "usecase.failed", Value: true}})
		statusCode = codes.Internal

		return nil, status.Error(statusCode, err.Error())
	}

	return response.GetListResponse(respBook), nil
}
