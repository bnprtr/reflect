# Reflect Example Test Servers

This directory contains simple test servers implementing the same Echo service across three different RPC protocols:

- **Connect** (HTTP/JSON-based RPC)
- **gRPC** (Standard Protocol Buffers over HTTP/2)
- **gRPC-Web** (Browser-compatible gRPC)

These servers are designed to validate that Reflect's "Try it" functionality works correctly with all three transport modes.

## Directory Structure

```
examples/
├── .mise.toml                    # Tool version management
├── proto/                        # Shared protocol definitions
│   ├── buf.yaml                  # Buf configuration
│   ├── buf.gen.yaml              # Code generation config
│   ├── go.mod
│   └── echo/v1/
│       ├── echo.proto            # Service definition
│       ├── echo.pb.go            # Generated protobuf code
│       ├── echo_grpc.pb.go       # Generated gRPC code
│       └── echov1connect/        # Generated Connect code
├── connect-server/               # Connect server (port 8081)
│   ├── go.mod
│   └── main.go
├── grpc-server/                  # gRPC server (port 8082)
│   ├── go.mod
│   └── main.go
└── grpc-web-server/              # gRPC-Web server (ports 8083, 8084)
    ├── go.mod
    └── main.go
```

## Prerequisites

### Using mise (Recommended)

We use [mise](https://mise.jdx.dev/) to manage tool dependencies. If you have mise installed:

```bash
cd examples
mise install
mise trust
```

This will install:
- protoc (Protocol Buffer compiler)
- buf (Modern protobuf tooling)
- go 1.23

### Manual Installation

If you prefer not to use mise, install these tools manually:

- [Go 1.23+](https://go.dev/dl/)
- [protoc](https://grpc.io/docs/protoc-installation/)
- [buf](https://buf.build/docs/installation/) (optional, for regenerating code)

## Setup

### 1. Generate Proto Code (if needed)

The generated code is already committed, but to regenerate:

```bash
cd proto
buf generate
```

Or with mise:

```bash
cd proto
mise exec -- buf generate
```

### 2. Install Dependencies

Each server has its own `go.mod`. Install dependencies for all servers:

```bash
# Connect server
cd connect-server
go mod download

# gRPC server
cd ../grpc-server
go mod download

# gRPC-Web server
cd ../grpc-web-server
go mod download
```

## Running the Servers

Each server runs on a different port and can be started independently.

### Connect Server (Port 8081)

```bash
cd connect-server
go run main.go
```

The server will start on `http://localhost:8081`.

**Test with curl:**

```bash
# Unary Echo
curl -X POST http://localhost:8081/echo.v1.EchoService/Echo \
  -H 'Content-Type: application/json' \
  -d '{"message":"Hello Connect"}'

# Streaming Echo
curl -X POST http://localhost:8081/echo.v1.EchoService/EchoStream \
  -H 'Content-Type: application/connect+json' \
  -d '{"message":"Hello Stream","count":5}'
```

### gRPC Server (Port 8082)

```bash
cd grpc-server
go run main.go
```

The server will start on `localhost:8082`.

**Test with grpcurl:**

First, install grpcurl if you haven't:

```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

Then test:

```bash
# List services (reflection enabled)
grpcurl -plaintext localhost:8082 list

# Unary Echo
grpcurl -plaintext -d '{"message":"Hello gRPC"}' \
  localhost:8082 echo.v1.EchoService/Echo

# Streaming Echo
grpcurl -plaintext -d '{"message":"Hello Stream","count":5}' \
  localhost:8082 echo.v1.EchoService/EchoStream
```

### gRPC-Web Server (Ports 8083 & 8084)

```bash
cd grpc-web-server
go run main.go
```

This server runs on two ports:
- **Port 8083**: gRPC-Web endpoint (for browser clients)
- **Port 8084**: Standard gRPC endpoint (for testing)

**Test the gRPC endpoint on port 8084:**

```bash
# Unary Echo
grpcurl -plaintext -d '{"message":"Hello gRPC-Web"}' \
  localhost:8084 echo.v1.EchoService/Echo

# Streaming Echo
grpcurl -plaintext -d '{"message":"Hello Stream","count":5}' \
  localhost:8084 echo.v1.EchoService/EchoStream
```

**Note:** Testing the gRPC-Web endpoint (port 8083) requires a gRPC-Web client or browser-based testing tool.

## Running All Servers at Once

You can run all servers in parallel using separate terminal windows, or use a tool like `tmux` or `foreman`.

**Using background processes (for quick testing):**

```bash
# From the examples directory
cd connect-server && go run main.go &
cd ../grpc-server && go run main.go &
cd ../grpc-web-server && go run main.go &
```

**Stop all:**

```bash
pkill -f "go run main.go"
```

## Testing with Reflect

Once the servers are running, you can use Reflect to document and test them:

```bash
# From the root of the reflect repository
./reflect -proto-root examples/proto -addr :8080
```

Then open http://localhost:8080 to view the documentation for the Echo service.

### Configuring Environments

Create a `reflect.yaml` configuration to define environments for each server:

```yaml
environments:
  - name: connect
    baseURL: http://localhost:8081
    transport: connect
  - name: grpc
    baseURL: http://localhost:8082
    transport: grpc
  - name: grpc-web
    baseURL: http://localhost:8083
    transport: grpc-web
```

## Service Definition

The Echo service provides two RPCs:

### Echo (Unary)

Echoes back the received message with a timestamp.

**Request:**
```json
{
  "message": "Your message here"
}
```

**Response:**
```json
{
  "message": "Your message here",
  "timestamp": 1697500000000
}
```

### EchoStream (Server Streaming)

Streams back the message multiple times.

**Request:**
```json
{
  "message": "Your message here",
  "count": 3
}
```

**Response:** A stream of:
```json
{
  "message": "Your message here (stream 1/3)",
  "timestamp": 1697500000000
}
```

## Troubleshooting

### Port Already in Use

If you see an error like "address already in use", check if a server is already running:

```bash
lsof -i :8081  # Or :8082, :8083, :8084
```

Kill the process:

```bash
kill -9 <PID>
```

### Proto Code Generation Issues

If you encounter issues with generated code:

1. Ensure buf is installed: `buf --version`
2. Regenerate code: `cd proto && buf generate`
3. Check that all imports resolve: `go mod tidy` in each server directory

### Module Resolution Errors

The servers use local module replacement for the proto package. If you move the examples directory, update the `replace` directive in each `go.mod`:

```go
replace github.com/bnprtr/reflect/examples/proto => ../proto
```

## Development

### Adding New RPCs

1. Edit `proto/echo/v1/echo.proto`
2. Regenerate code: `cd proto && buf generate`
3. Update server implementations in each server's `main.go`
4. Test with clients

### Building Binaries

To build standalone binaries:

```bash
# Connect server
cd connect-server && go build -o connect-server

# gRPC server
cd grpc-server && go build -o grpc-server

# gRPC-Web server
cd grpc-web-server && go build -o grpc-web-server
```

## License

Same as the parent Reflect project. See [../LICENSE](../LICENSE).
