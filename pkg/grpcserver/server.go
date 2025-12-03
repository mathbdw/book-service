// Package grpcserver implements HTTP server.
package grpcserver

import (
	"fmt"
	"net"

	"google.golang.org/grpc/reflection"

	pbgrpc "google.golang.org/grpc"
)

const (
	_defaultAddr = ":80"
)

// Server -.
type Server struct {
	App     *pbgrpc.Server
	notify  chan error
	address string
	mode    bool
}

// New -.
func New(opts ...Option) *Server {
	// Создаем gRPC сервер С interceptors
	grpcServer := pbgrpc.NewServer(
		pbgrpc.UnaryInterceptor(GRPCUnaryServerInterceptor()),
		// Можно добавить StreamInterceptor если нужен
	)

	s := &Server{
		App:     grpcServer,
		notify:  make(chan error, 1),
		address: _defaultAddr,
	}

	// Custom options
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Start -.

func (s *Server) Start() {
	go func() {
		ln, err := net.Listen("tcp", s.address)
		if err != nil {
			s.notify <- fmt.Errorf("failed to listen: %w", err)
			close(s.notify)

			return
		}

		if s.mode {
			reflection.Register(s.App)
		}

		s.notify <- s.App.Serve(ln)
		close(s.notify)
	}()
}

// Notify -.
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown -.
func (s *Server) Shutdown() {
	s.App.GracefulStop()
}
