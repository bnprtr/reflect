module github.com/bnprtr/reflect/examples/grpc-server

go 1.23

toolchain go1.24.9

require (
	github.com/bnprtr/reflect/examples/proto v0.0.0
	google.golang.org/grpc v1.69.4
)

require (
	golang.org/x/net v0.31.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241015192408-796eee8c2d53 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/bnprtr/reflect/examples/proto => ../proto
