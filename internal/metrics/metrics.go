package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
)

var (
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rpc_request_duration_seconds",
			Help:    "Histogram of RPC handler duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rpc_requests_total",
			Help: "Total number of RPC requests",
		},
		[]string{"service", "method"},
	)

	ErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rpc_errors_total",
			Help: "Total number of RPC errors",
		},
		[]string{"service", "method", "code"},
	)
)

// RegisterMetrics registers metrics with Prometheus and returns an HTTP handler for /metrics.
func RegisterMetrics() http.Handler {
	prometheus.MustRegister(RequestDuration, RequestsTotal, ErrorsTotal)
	return promhttp.Handler()
}

// UnaryServerInterceptor returns a grpc.UnaryServerInterceptor that records metrics.
// serviceName is a short identifier like "order" or "payment".
func UnaryServerInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()
		resp, err = handler(ctx, req)

		method := info.FullMethod // e.g. /package.Service/Method
		RequestsTotal.WithLabelValues(serviceName, method).Inc()
		RequestDuration.WithLabelValues(serviceName, method).Observe(time.Since(start).Seconds())

		if err != nil {
			code := status.Code(err).String()
			ErrorsTotal.WithLabelValues(serviceName, method, code).Inc()
		}
		return resp, err
	}
}

// UnaryClientInterceptor returns a grpc.UnaryClientInterceptor that records metrics for outgoing calls.
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

		RequestsTotal.WithLabelValues(serviceName, method).Inc()
		RequestDuration.WithLabelValues(serviceName, method).Observe(time.Since(start).Seconds())
		if err != nil {
			code := status.Code(err).String()
			ErrorsTotal.WithLabelValues(serviceName, method, code).Inc()
		}
		return err
	}
}
