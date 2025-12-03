package gateway

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func CustomUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()

		ctx, span := otel.Tracer("grpc-gateway").Start(
			ctx,
			method,
			trace.WithSpanKind(trace.SpanKindClient),
		)
		defer span.End()

		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", method),
			attribute.String("net.peer.name", cc.Target()),
		)

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		carrier := GRPCCarrier{MD: &md}
		otel.GetTextMapPropagator().Inject(ctx, carrier)

		ctx = metadata.NewOutgoingContext(ctx, md)

		err := invoker(ctx, method, req, reply, cc, opts...)

		duration := time.Since(start)
		span.SetAttributes(
			attribute.Float64("rpc.duration_ms", float64(duration.Milliseconds())),
		)

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		}

		return err
	}
}

type GRPCCarrier struct {
	MD *metadata.MD
}

func (c GRPCCarrier) Set(key, value string) {
	key = strings.ToLower(key)
	c.MD.Set(key, value)
}

func (c GRPCCarrier) Get(key string) string {
	key = strings.ToLower(key)
	values := c.MD.Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func (c GRPCCarrier) Keys() []string {
	keys := make([]string, 0, len(*c.MD))
	for k := range *c.MD {
		keys = append(keys, k)
	}
	return keys
}
