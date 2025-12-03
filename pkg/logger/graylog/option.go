package graylog

import (
	"fmt"
)

// Option -.
type Option func(*GELFWriter)

// Host - set host
func Host(host string, port uint16) Option {
	return func(s *GELFWriter) {
		s.host = fmt.Sprintf("%s:%d", host, port)
	}
}

// Version - set version
func Version(version string) Option {
	return func(s *GELFWriter) {
		s.version = version
	}
}
