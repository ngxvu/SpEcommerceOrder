package bootstrap

import (
	"log"
	"net/http"
	"order/internal/metrics"
	"order/pkg/core/configloader"
)

func StartMetricsServer(config *configloader.Config) {
	addr := config.MetricsAddress
	if addr == "" {
		addr = ":9090"
	}

	handler := metrics.RegisterMetrics()
	http.Handle("/metrics", handler)
	log.Printf("prometheus metrics listening on %s/metrics", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("metrics server error: %v", err)
	}
}
