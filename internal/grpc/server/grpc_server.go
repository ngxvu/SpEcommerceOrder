package server

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"net/http"
	"order/internal/metrics"
	pb "order/pkg/proto"
)

type GRPCServer struct {
	server     *grpc.Server
	grpcAddr   string
	httpAddr   string
	httpServer *http.Server
	lis        net.Listener
}

func NewGRPCServer(handler pb.OrderServiceServer, grpcAddr, httpAddr string) *GRPCServer {
	s := grpc.NewServer(
		// incoming from clients to server will be measured here
		grpc.UnaryInterceptor(metrics.UnaryServerInterceptor("order")),
	)
	pb.RegisterOrderServiceServer(s, handler)
	return &GRPCServer{
		server:   s,
		grpcAddr: grpcAddr,
		httpAddr: httpAddr,
	}
}

func (s *GRPCServer) Run(ctx context.Context) error {
	// start gRPC listener
	lis, err := net.Listen("tcp", s.grpcAddr)
	if err != nil {
		return err
	}
	s.lis = lis

	grpcErrCh := make(chan error, 1)
	go func() {
		log.Printf("gRPC server running on %s", s.grpcAddr)
		if err := s.server.Serve(lis); err != nil {
			grpcErrCh <- err
		}
	}()

	// setup grpc-gateway
	gwMux := runtime.NewServeMux()
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := pb.RegisterOrderServiceHandlerFromEndpoint(ctx, gwMux, s.grpcAddr, dialOpts); err != nil {
		s.server.GracefulStop()
		return err
	}

	// create top-level HTTP mux and mount /metrics and the gateway
	httpMux := http.NewServeMux()
	httpMux.Handle("/metrics", metrics.RegisterMetrics())
	httpMux.Handle("/", gwMux)

	s.httpServer = &http.Server{
		Addr:    s.httpAddr,
		Handler: httpMux,
	}

	httpErrCh := make(chan error, 1)
	go func() {
		log.Printf("HTTP gateway listening on %s", s.httpAddr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			httpErrCh <- err
		}
	}()

	// wait for cancel or server error
	select {
	case <-ctx.Done():
		log.Println("shutting down servers")
		_ = s.httpServer.Shutdown(context.Background())
		s.server.GracefulStop()
		return ctx.Err()
	case err := <-grpcErrCh:
		return fmt.Errorf("gRPC server error: %w", err)
	case err := <-httpErrCh:
		return fmt.Errorf("HTTP gateway error: %w", err)
	}
}

// Stop triggers an immediate graceful shutdown.
func (s *GRPCServer) Stop() {
	if s.httpServer != nil {
		_ = s.httpServer.Shutdown(context.Background())
	}
	if s.server != nil {
		s.server.GracefulStop()
	}
}
