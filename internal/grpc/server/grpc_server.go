package server

import (
	"fmt"
	//pb "basesource/pkg/proto/orderpb"
	"google.golang.org/grpc"
	"log"
	"net"
	pb "order/pkg/proto/orderpb"
)

type GRPCServer struct {
	server *grpc.Server
}

func NewGRPCServer(handler pb.OrderServiceServer) *GRPCServer {
	s := grpc.NewServer()
	pb.RegisterOrderServiceServer(s, handler)
	return &GRPCServer{server: s}
}

func (s *GRPCServer) Run(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	log.Printf("gRPC server running on port %d", port)
	return s.server.Serve(lis)
}

func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}
