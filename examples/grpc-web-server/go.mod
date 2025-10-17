module github.com/bnprtr/reflect/examples/grpc-web-server

go 1.23

toolchain go1.24.9

require (
	github.com/bnprtr/reflect/examples/proto v0.0.0
	github.com/improbable-eng/grpc-web v0.15.0
	google.golang.org/grpc v1.69.4
)

require (
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/klauspost/compress v1.11.7 // indirect
	github.com/rs/cors v1.8.0 // indirect
	golang.org/x/net v0.31.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241015192408-796eee8c2d53 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	nhooyr.io/websocket v1.8.6 // indirect
)

replace github.com/bnprtr/reflect/examples/proto => ../proto
