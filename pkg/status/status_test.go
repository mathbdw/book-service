package status

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		s := New()

		assert.NotNil(t, s)
		assert.NotNil(t, s.Server)
		assert.NotNil(t, s.notify)
		assert.Equal(t, "", s.address)
		assert.Equal(t, 1, cap(s.notify))
	})

	t.Run("with options", func(t *testing.T) {
		s := New(Address("", 8080))

		assert.NotNil(t, s)
		assert.Equal(t, ":8080", s.address)
		assert.Equal(t, ":8080", s.Server.Addr)
	})
}

func TestStart(t *testing.T) {
	t.Run("successful start", func(t *testing.T) {
		s := New(Address("", 0))
		s.Server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		s.Start()

		select {
		case err := <-s.Notify():
			t.Fatalf("Server should not return error: %v", err)
		default:
		}

		ctx := context.Background()
		err := s.Shutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("start with invalid address", func(t *testing.T) {
		s := New(Address("invalid", 0))
		s.Start()

		select {
		case err := <-s.Notify():
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to listen")
		}
	})
}

func TestShutdown(t *testing.T) {
	t.Run("successful shutdown", func(t *testing.T) {
		s := New(Address("", 0))
		s.Server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		s.Start()

		time.Sleep(10 * time.Millisecond)
		ctx := context.Background()
		err := s.Shutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("shutdown not started server", func(t *testing.T) {
		s := New()
		
		ctx := context.Background()
		err := s.Shutdown(ctx)
		assert.NoError(t, err)
	})
}
