package handlers

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/mathbdw/book/internal/interfaces/controllers/grpc/v1/converters"
	"github.com/mathbdw/book/internal/interfaces/controllers/grpc/v1/response"
	errs "github.com/mathbdw/book/internal/errors"
	pb "github.com/mathbdw/book/proto"
	"github.com/mathbdw/book/internal/interfaces/observability"
)

// GetByIDs - returns the books based on data from a gRPC request.
// Returns:
// - *pb.BooksResponse: empty response on successful creation
// - error: validation or business logic error
//
// Errors:
// - codes.InvalidArgument: input data validation error
// - codes.Internal: database or usecase level error
//
// Logging:
// - Info level: validation and business logic errors
func (bh *BookHandler) GetByIDs(ctx context.Context, req *pb.BookGetRequest) (*pb.BooksResponse, error) {
	start := time.Now()
	logger := bh.observ.WithContext(ctx)
	ctx, span := bh.observ.StartSpan(ctx, "v1.BookService.GetByIDs")
	span.SetAttributes([]observability.Attribute{
		{Key: "http.method", Value: "GET"},
		{Key: "http.route", Value: "v1/books"},
	})

	defer span.End()

	var statusCode codes.Code = codes.OK
	defer func() {
		duration := time.Since(start).Seconds()
		bh.observ.RecordHanderRequest(ctx, "GET", "v1/books", int(statusCode), duration)
	}()

	if err := req.Validate(); err != nil {
		logger.Info("grpcBook.GetByIDs: validate", map[string]any{
			"error": err.Error(),
			"ids":   req.GetBookId(),
		})
		span.RecordError(err)
		span.SetAttributes([]observability.Attribute{{Key: "validation.failed", Value: true}})
		statusCode = codes.InvalidArgument

		return nil, status.Error(statusCode, err.Error())
	}

	span.SetAttributes([]observability.Attribute{{Key: "book.ids", Value: req.GetBookId()}})

	books, err := bh.uc.Get.GetByIDs(ctx, req.GetBookId())
	if err != nil {
		span.SetAttributes([]observability.Attribute{{Key: "usecase.failed", Value: true}})

		if errors.Is(err, errs.ErrNotFound) {
			logger.Info("grpcBook.GetByIDs: usecase not found IDs", map[string]any{
				"error": err.Error(),
				"ids":   req.GetBookId(),
			})
			statusCode = codes.NotFound

			return nil, status.Error(statusCode, errs.ErrNotFound.Error())
		} else {
			logger.Info("grpcBook.GetByIDs: usecase IDs", map[string]any{
				"error": err.Error(),
				"ids":   req.GetBookId(),
			})
			statusCode = codes.Internal

			return nil, status.Error(statusCode, err.Error())
		}
	}

	if data, ok := metadata.FromIncomingContext(ctx); ok {
		bh.observ.Info("meta", map[string]any{"meta": data})

		if val, ok := data["loglevel"]; ok {
			if len(val) > 0 {
				newLevel, err := converters.LevelToZerolog(val[0])
				if err != nil {
					bh.observ.Info("grpcBook.GetByIDs: error newLevel", map[string]any{
						"error":    err.Error(),
						"newLevel": val,
					})

					return response.GetBooksResponse(books), nil
				}

				bh.observ.Error("grpcBook.GetByIDs: set newLevel", map[string]any{
					"newLevel": val,
				})
				bh.observ.SetLevel(newLevel)
			}
		}
	}

	return response.GetBooksResponse(books), nil
}
