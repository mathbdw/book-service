package grpcserver

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	pbgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// GRPCUnaryServerInterceptor - interceptor для gRPC сервера
func GRPCUnaryServerInterceptor() pbgrpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *pbgrpc.UnaryServerInfo, handler pbgrpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			carrier := GRPCCarrier{MD: md}
			ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
		}

		ctx, span := otel.Tracer("grpc-server").Start(
			ctx,
			info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", extractServiceName(info.FullMethod)),
			attribute.String("rpc.method", extractMethodName(info.FullMethod)),
		)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		span.SetAttributes(
			attribute.Float64("rpc.duration_ms", float64(duration.Milliseconds())),
		)

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		}

		return resp, err
	}
}

type GRPCCarrier struct {
	MD metadata.MD
}

func (c GRPCCarrier) Get(key string) string {
	values := c.MD[strings.ToLower(key)]
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func (c GRPCCarrier) Set(key, value string) {
	key = strings.ToLower(key)
	c.MD[key] = append(c.MD[key], value)
}

func (c GRPCCarrier) Keys() []string {
	keys := make([]string, 0, len(c.MD))
	for k := range c.MD {
		keys = append(keys, k)
	}
	return keys
}

// Вспомогательные функции
func extractServiceName(fullMethod string) string {
	if len(fullMethod) > 0 && fullMethod[0] == '/' {
		fullMethod = fullMethod[1:]
	}
	if pos := strings.LastIndex(fullMethod, "/"); pos >= 0 {
		return fullMethod[:pos]
	}
	return "unknown"
}

func extractMethodName(fullMethod string) string {
	if pos := strings.LastIndex(fullMethod, "/"); pos >= 0 {
		return fullMethod[pos+1:]
	}
	return fullMethod
}
