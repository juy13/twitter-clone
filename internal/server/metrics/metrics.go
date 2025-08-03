package metrics

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"twitter-clone/internal/domain/config"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsServer struct {
	server *http.Server
	info   string
}

func NewMetricsServer(config config.MetricsConfig) *MetricsServer {
	var debugServer *http.Server
	debugMux := http.NewServeMux()
	debugMux.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
	debugMux.HandleFunc("/debug/pprof/", pprof.Index)
	debugMux.HandleFunc("/debug/pprof/heap", pprof.Index)
	debugMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	debugMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	debugMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	debugMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	debugAddress := fmt.Sprintf("%s:%d", config.MetricsServerHost(), config.MetricsServerPort())
	debugServer = &http.Server{
		Addr:    debugAddress,
		Handler: debugMux,
	}
	return &MetricsServer{
		server: debugServer,
		info:   fmt.Sprintf("Running metrics server on %v", debugAddress),
	}
}

func (s *MetricsServer) Start() error {
	return s.server.ListenAndServe()
}

func (s *MetricsServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *MetricsServer) Info() string {
	return s.info
}
