package gateway

import (
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/mathbdw/book/internal/interfaces/observability"
)

func GatewayMiddleware(log observability.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := otel.Tracer("grpc-gateway-middleware").Start(
			ctx,
			"HTTP "+r.Method+" "+r.URL.Path,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
				attribute.String("component", "grpc-gateway"),
			),
		)
		defer span.End()

		rw := &responseWriter{ResponseWriter: w, statusCode: 200}
		r = r.WithContext(ctx)
		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		span.SetAttributes(
			attribute.Int("http.status_code", rw.statusCode),
			attribute.Float64("http.duration_ms", float64(duration.Milliseconds())),
			attribute.Int64("http.response_size", rw.responseSize),
		)
		if rw.statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", rw.statusCode))
		}

		if httpRequestCounter != nil && httpRequestDuration != nil {
			attrs := []attribute.KeyValue{
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
				attribute.Int("http.status_code", rw.statusCode),
			}
			httpRequestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
			httpRequestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
		}

		log.Info("gateway request", map[string]any{
			"method":   r.Method,
			"path":     r.URL.Path,
			"status":   rw.statusCode,
			"duration": duration.String(),
			"size":     rw.responseSize,
		})
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
	wroteHeader  bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(200)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.responseSize += int64(n)
	return n, err
}
