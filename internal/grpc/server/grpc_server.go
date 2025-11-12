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
		grpc.UnaryInterceptor(metrics.UnaryServerInterceptor("order")),
	)
	pb.RegisterOrderServiceServer(s, handler)
	return &GRPCServer{
		server:   s,
		grpcAddr: grpcAddr,
		httpAddr: httpAddr,
	}
}

// Run starts the gRPC server and the grpc-gateway HTTP server.
// Cancel the provided ctx to trigger graceful shutdown.
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
		// stop gRPC server if gateway registration fails
		s.server.GracefulStop()
		return err
	}

	s.httpServer = &http.Server{
		Addr:    s.httpAddr,
		Handler: gwMux,
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
		// Graceful shutdown of HTTP and gRPC
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
