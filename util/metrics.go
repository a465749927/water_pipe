package util

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry           *prometheus.Registry
	ConnectionsTotal   *prometheus.CounterVec
	ConnectionsActive  *prometheus.GaugeVec
	TrafficBytes       *prometheus.CounterVec
	ConnectionDuration *prometheus.HistogramVec
	HealthCheckStatus  *prometheus.GaugeVec
	server             *http.Server
}

type MetricsConfig struct {
	Enabled bool
	Address string
}

func NewMetrics(config MetricsConfig) (*Metrics, error) {
	if !config.Enabled {
		return nil, nil
	}

	registry := prometheus.NewRegistry()

	connectionsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "water_pipe_connections_total",
			Help: "Total number of connections",
		},
		[]string{"node_id", "direction", "status"},
	)

	connectionsActive := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "water_pipe_connections_active",
			Help: "Number of active connections",
		},
		[]string{"node_id", "direction"},
	)

	trafficBytes := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "water_pipe_traffic_bytes",
			Help: "Total traffic in bytes",
		},
		[]string{"node_id", "direction"},
	)

	connectionDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "water_pipe_connection_duration_seconds",
			Help:    "Connection duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"node_id", "direction"},
	)

	healthCheckStatus := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "water_pipe_health_check_status",
			Help: "Health check status (1 = healthy, 0 = unhealthy)",
		},
		[]string{"node_id"},
	)

	registry.MustRegister(connectionsTotal)
	registry.MustRegister(connectionsActive)
	registry.MustRegister(trafficBytes)
	registry.MustRegister(connectionDuration)
	registry.MustRegister(healthCheckStatus)

	metrics := &Metrics{
		registry:           registry,
		ConnectionsTotal:   connectionsTotal,
		ConnectionsActive:  connectionsActive,
		TrafficBytes:       trafficBytes,
		ConnectionDuration: connectionDuration,
		HealthCheckStatus:  healthCheckStatus,
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	server := &http.Server{
		Addr:    config.Address,
		Handler: mux,
	}

	metrics.server = server

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Metrics server error: %v\n", err)
		}
	}()

	return metrics, nil
}

func (m *Metrics) Stop(ctx context.Context) error {
	if m == nil || m.server == nil {
		return nil
	}
	return m.server.Shutdown(ctx)
}
