package status

import (
	"fmt"
)

// Option -.
type Option func(*Status)

// Address -.
func Address(host string, port uint16) Option {
	return func(s *Status) {
		s.address = fmt.Sprintf("%s:%d", host, port)
	}
}
