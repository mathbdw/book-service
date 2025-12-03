package handlers

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/mathbdw/book/internal/interfaces/observability"
	errs "github.com/mathbdw/book/internal/errors"
	pb "github.com/mathbdw/book/proto"
)

// Delete - deletes the books based on data from a gRPC request.
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
func (bh *BookHandler) Delete(ctx context.Context, req *pb.BookGetRequest) (*emptypb.Empty, error) {
	start := time.Now()
	logger := bh.observ.WithContext(ctx)
	ctx, span := bh.observ.StartSpan(ctx, "v1.BookService.Delete")
	span.SetAttributes([]observability.Attribute{
		{Key: "http.method", Value: "DELETE"},
		{Key: "http.methrouteod", Value: "v1/books"},
	})

	defer span.End()

	var statusCode codes.Code = codes.OK
	defer func() {
		duration := time.Since(start).Seconds()
		bh.observ.RecordHanderRequest(ctx, "DELETE", "v1/books", int(statusCode), duration)
	}()

	if err := req.Validate(); err != nil {
		logger.Info("grpcBook.Delete: validate", map[string]any{"error": err.Error()})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "validation.failed", Value: true}})
		statusCode = codes.InvalidArgument

		return nil, status.Error(statusCode, "invalid arguments")
	}

	span.SetAttributes([]observability.Attribute{{Key: "book.ids", Value: req.GetBookId()}})

	err := bh.uc.Remove.Execute(ctx, req.GetBookId())
	if err != nil {
		logger.Info(
			"grpcBook.Delete: usecase",
			map[string]any{
				"error": err.Error(),
				"ids":   req.GetBookId(),
			},
		)
		span.SetAttributes([]observability.Attribute{{Key: "usecase.failed", Value: true}})

		if errors.Is(err, errs.ErrNotFound) {
			statusCode = codes.InvalidArgument
			return nil, status.Error(statusCode, errs.ErrNotFound.Error())
		}

		statusCode = codes.Internal
		return nil, status.Error(statusCode, err.Error())
	}

	return &emptypb.Empty{}, nil
}
