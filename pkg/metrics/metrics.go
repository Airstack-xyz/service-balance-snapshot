package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func GetMetricsHandler(reg *prometheus.Registry) *http.Handler {
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	return &promHandler
}
