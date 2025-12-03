package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/mathbdw/book/internal/interfaces/observability"
	pb "github.com/mathbdw/book/proto"
)

type Server struct {
	Server      *http.Server
	notify      chan error
	addressGrpc string
	address     string
}

var (
	httpRequestCounter  metric.Int64Counter
	httpRequestDuration metric.Float64Histogram
)

// New -.
func New(opts ...Option) *Server {
	s := &Server{
		notify: make(chan error, 1),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Start -.
func (s *Server) Start(log observability.Logger) {
	conn, err := grpc.NewClient(
		s.addressGrpc,
		grpc.WithUnaryInterceptor(CustomUnaryClientInterceptor()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)

	if err != nil {
		log.Error("gateway.Start: failed to dial server", map[string]any{"error": err.Error()})
	}

	mux := runtime.NewServeMux()
	if err := pb.RegisterBookServiceHandler(context.Background(), mux, conn); err != nil {
		log.Error("gateway.Start: failed registration handler book", map[string]any{"error": err.Error()})

		return
	}

	var handler http.Handler = mux
	handler = GatewayMiddleware(log, handler)

	s.Server = &http.Server{
		Addr:    s.address,
		Handler: handler,
	}

	go func() {
		log.Info("gateway: starting HTTP gateway", map[string]any{
			"address": s.address,
			"grpc":    s.addressGrpc,
		})

		err := s.Server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.notify <- fmt.Errorf("failed to listen: %w", err)
			close(s.notify)
			return
		}
	}()
}

// Notify -.
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown -.
func (s *Server) Shutdown(ctx context.Context) error {
	err := s.Server.Shutdown(ctx)

	if err != nil {
		return err
	}

	return nil
}
