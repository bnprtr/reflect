package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	echov1 "github.com/bnprtr/reflect/examples/proto/echo/v1"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// echoServer implements the EchoService.
type echoServer struct {
	echov1.UnimplementedEchoServiceServer
}

// Echo implements the unary Echo RPC.
func (s *echoServer) Echo(ctx context.Context, req *echov1.EchoRequest) (*echov1.EchoResponse, error) {
	log.Printf("[gRPC-Web] Echo called with message: %q", req.Message)

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

	log.Printf("[gRPC-Web] EchoStream called with message: %q, count: %d", req.Message, count)

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
	grpcServer := grpc.NewServer()
	echov1.RegisterEchoServiceServer(grpcServer, &echoServer{})

	// Enable reflection
	reflection.Register(grpcServer)

	// Wrap the gRPC server with grpc-web
	wrappedGrpc := grpcweb.WrapServer(grpcServer,
		grpcweb.WithOriginFunc(func(origin string) bool {
			// Allow all origins for testing
			return true
		}),
	)

	httpServer := &http.Server{
		Addr: ":8083",
		Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			if wrappedGrpc.IsGrpcWebRequest(req) {
				wrappedGrpc.ServeHTTP(resp, req)
				return
			}
			// Fall back to regular gRPC if not a grpc-web request
			// This requires HTTP/2
			http.NotFound(resp, req)
		}),
	}

	log.Printf("gRPC-Web server listening on %s", httpServer.Addr)
	log.Println("This server accepts gRPC-Web requests from browsers")
	log.Println("You can test with a gRPC-Web client or browser")

	// Start a goroutine to listen for regular gRPC on a different port
	go func() {
		grpcAddr := ":8084"
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Printf("failed to listen for gRPC: %v", err)
			return
		}
		log.Printf("Also serving regular gRPC on %s for testing", grpcAddr)
		log.Printf("Try it: grpcurl -plaintext -d '{\"message\":\"Hello gRPC-Web\"}' localhost:8084 echo.v1.EchoService/Echo")
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to serve gRPC: %v", err)
		}
	}()

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to serve: %v", err)
	}
}
