package gateway

import (
	"fmt"
)

// Option -.
type Option func(*Server)

// Address gateway - .
func Address(host string, port uint16) Option {
	return func(s *Server) {
		s.address = fmt.Sprintf("%s:%d", host, port)
	}
}

// AddressGrpc - .
func AddressGrpc(host string, port uint16) Option {
	return func(s *Server) {
		s.addressGrpc = fmt.Sprintf("%s:%d", host, port)
	}
}
