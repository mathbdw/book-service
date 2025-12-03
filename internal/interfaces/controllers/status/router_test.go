package status

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathbdw/book/config"
	pzlog "github.com/mathbdw/book/pkg/logger/zerolog"
	pserver_status "github.com/mathbdw/book/pkg/status"
)

func TestNewRouter(t *testing.T) {
	cfg := &config.Config{
		Status: config.Status{
			LivenessPath:  "/healthz",
			ReadinessPath: "/readyz",
			VersionPath:   "/version",
		},
		Project: config.Project{
			Name:    "test-service",
			Debug:   true,
			Version: "1.0.0",
		},
	}

	isReady := &atomic.Value{}
	isReady.Store(false)
	l := pzlog.New(&config.Config{})
	app := &pserver_status.Status{
		Server: &http.Server{},
	}

	NewRouter(app, cfg, isReady, l)

	assert.NotNil(t, app.Server.Handler)

	mux := app.Server.Handler.(*http.ServeMux)
	req, _ := http.NewRequest("GET", cfg.Status.LivenessPath, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestLivenessHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/healthz", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	livenessHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Body.String())
}

func TestReadinessHandler(t *testing.T) {
	tests := []struct {
		name     string
		isReady  *atomic.Value
		expected int
	}{
		{
			name:     "service is ready",
			isReady:  func() *atomic.Value { v := &atomic.Value{}; v.Store(true); return v }(),
			expected: http.StatusOK,
		},
		{
			name:     "service is not ready",
			isReady:  func() *atomic.Value { v := &atomic.Value{}; v.Store(false); return v }(),
			expected: http.StatusServiceUnavailable,
		},
		{
			name:     "isReady is nil",
			isReady:  nil,
			expected: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := readinessHandler(tt.isReady)

			req, err := http.NewRequest("GET", "/readyz", nil)
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			handler(rr, req)

			assert.Equal(t, tt.expected, rr.Code)
			if tt.expected == http.StatusServiceUnavailable {
				assert.Contains(t, rr.Body.String(), "Service Unavailable")
			}
		})
	}
}

func TestVersionHandler(t *testing.T) {
	cfg := &config.Config{
		Project: config.Project{
			Name:    "test-service",
			Debug:   true,
			Version: "1.0.0",
		},
	}
	l := pzlog.New(&config.Config{})

	handler := versionHandler(cfg, l)
	req, err := http.NewRequest("GET", "/version", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()
	handler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	expectedBody := `{"debug":true,"name":"test-service","version":"1.0.0"}`
	assert.JSONEq(t, expectedBody, rr.Body.String())
}
