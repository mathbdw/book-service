package status

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	"github.com/mathbdw/book/config"
	status_server "github.com/mathbdw/book/pkg/status"
	"github.com/mathbdw/book/internal/interfaces/observability"
)

func NewRouter(app *status_server.Status, cfg *config.Config, isReady *atomic.Value, l observability.Logger) {
	mux := http.DefaultServeMux

	mux.HandleFunc(cfg.Status.LivenessPath, livenessHandler)
	mux.HandleFunc(cfg.Status.ReadinessPath, readinessHandler(isReady))
	mux.HandleFunc(cfg.Status.VersionPath, versionHandler(cfg, l))

	app.Server.Handler = mux
}

func livenessHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func readinessHandler(isReady *atomic.Value) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if isReady == nil || !isReady.Load().(bool) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)

			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func versionHandler(cfg *config.Config, l observability.Logger) func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		data := map[string]interface{}{
			"name":    cfg.Project.Name,
			"debug":   cfg.Project.Debug,
			"version": cfg.Project.Version,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(data); err != nil {
			l.Error("status.versionHandler: service information encoding error", map[string]any{"error": err.Error()})
		}
	}
}
