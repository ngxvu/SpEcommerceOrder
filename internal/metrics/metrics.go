package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"service", "method", "code", "direction"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "direction"},
	)
)

func init() {
	prometheus.MustRegister(requestCount, requestDuration)
}

// RegisterMetrics returns an http.Handler for Prometheus scraping.
func RegisterMetrics() http.Handler {
	return promhttp.Handler()
}

// UnaryServerInterceptor gets the gRPC server interceptor for metrics collection.
func UnaryServerInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err).String()
		requestCount.WithLabelValues(serviceName, info.FullMethod, code, "incoming").Inc()
		requestDuration.WithLabelValues(serviceName, info.FullMethod, "incoming").Observe(time.Since(start).Seconds())
		return resp, err
	}
}

// UnaryClientInterceptor gets the gRPC client interceptor for metrics collection.
func UnaryClientInterceptor(serviceName string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		code := status.Code(err).String()
		requestCount.WithLabelValues(serviceName, method, code, "outgoing").Inc()
		requestDuration.WithLabelValues(serviceName, method, "outgoing").Observe(time.Since(start).Seconds())
		return err
	}
}
