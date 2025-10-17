package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	echov1 "github.com/bnprtr/reflect/examples/proto/echo/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// echoServer implements the EchoService.
type echoServer struct {
	echov1.UnimplementedEchoServiceServer
}

// Echo implements the unary Echo RPC.
func (s *echoServer) Echo(ctx context.Context, req *echov1.EchoRequest) (*echov1.EchoResponse, error) {
	log.Printf("[gRPC] Echo called with message: %q", req.Message)

	return &echov1.EchoResponse{
		Message:   req.Message,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

// EchoStream implements the server-streaming EchoStream RPC.
func (s *echoServer) EchoStream(req *echov1.EchoRequest, stream echov1.EchoService_EchoStreamServer) error {
	count := req.Count
	if count <= 0 {
		count = 3
	}

	log.Printf("[gRPC] EchoStream called with message: %q, count: %d", req.Message, count)

	for i := int32(0); i < count; i++ {
		if err := stream.Context().Err(); err != nil {
			return err
		}

		if err := stream.Send(&echov1.EchoResponse{
			Message:   fmt.Sprintf("%s (stream %d/%d)", req.Message, i+1, count),
			Timestamp: time.Now().UnixMilli(),
		}); err != nil {
			return err
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func main() {
	addr := ":8082"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	echov1.RegisterEchoServiceServer(s, &echoServer{})

	// Enable reflection for tools like grpcurl
	reflection.Register(s)

	log.Printf("gRPC server listening on %s", addr)
	log.Printf("Try it: grpcurl -plaintext -d '{\"message\":\"Hello gRPC\"}' localhost:8082 echo.v1.EchoService/Echo")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
