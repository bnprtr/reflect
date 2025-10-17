package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"connectrpc.com/connect"
	echov1 "github.com/bnprtr/reflect/examples/proto/echo/v1"
	"github.com/bnprtr/reflect/examples/proto/echo/v1/echov1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// echoServer implements the EchoService.
type echoServer struct{}

// Echo implements the unary Echo RPC.
func (s *echoServer) Echo(
	ctx context.Context,
	req *connect.Request[echov1.EchoRequest],
) (*connect.Response[echov1.EchoResponse], error) {
	log.Printf("[Connect] Echo called with message: %q", req.Msg.Message)

	res := connect.NewResponse(&echov1.EchoResponse{
		Message:   req.Msg.Message,
		Timestamp: time.Now().UnixMilli(),
	})

	return res, nil
}

// EchoStream implements the server-streaming EchoStream RPC.
func (s *echoServer) EchoStream(
	ctx context.Context,
	req *connect.Request[echov1.EchoRequest],
	stream *connect.ServerStream[echov1.EchoResponse],
) error {
	count := req.Msg.Count
	if count <= 0 {
		count = 3
	}

	log.Printf("[Connect] EchoStream called with message: %q, count: %d", req.Msg.Message, count)

	for i := int32(0); i < count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := stream.Send(&echov1.EchoResponse{
			Message:   fmt.Sprintf("%s (stream %d/%d)", req.Msg.Message, i+1, count),
			Timestamp: time.Now().UnixMilli(),
		}); err != nil {
			return err
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func main() {
	server := &echoServer{}
	path, handler := echov1connect.NewEchoServiceHandler(server)

	mux := http.NewServeMux()
	mux.Handle(path, handler)

	addr := ":8081"
	log.Printf("Connect server listening on %s", addr)
	log.Printf("Try it: curl -X POST http://localhost:8081/echo.v1.EchoService/Echo -H 'Content-Type: application/json' -d '{\"message\":\"Hello Connect\"}'")

	if err := http.ListenAndServe(addr, h2c.NewHandler(mux, &http2.Server{})); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to serve: %v", err)
	}
}
