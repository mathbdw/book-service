package grpcserver

import (
	"fmt"
)

// Option -.
type Option func(*Server)

// Address -.
func Address(host string, port uint16) Option {
	return func(s *Server) {
		s.address = fmt.Sprintf("%s:%d", host, port)
	}
}

// Mode run -.
func Mode(mode bool) Option {
	return func(s *Server) {
		s.mode = mode
	}
}
