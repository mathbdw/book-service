package status

import (
	"context"
	"fmt"
	"net/http"
)

type Status struct {
	Server  *http.Server
	notify  chan error
	address string
}

// New -.
func New(opts ...Option) *Status {
	s := &Status{
		notify: make(chan error, 1),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.Server = &http.Server{
		Addr: s.address,
	}
	return s
}

func (s *Status) Start() {
	go func() {
		err := s.Server.ListenAndServe()
		if err != nil {
			s.notify <- fmt.Errorf("failed to listen: %w", err)
			close(s.notify)

			return
		}

	}()
}

// Notify -.
func (s *Status) Notify() <-chan error {
	return s.notify
}

// Shutdown -.
func (s *Status) Shutdown(ctx context.Context) error {
	err := s.Server.Shutdown(ctx)

	if err != nil {
		return err
	}

	return nil
}
