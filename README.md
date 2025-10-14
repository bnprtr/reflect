# Reflect

A modern, web-based protobuf reflection server that provides beautiful documentation for your Protocol Buffer services.

## Features

- 🚀 **Instant Setup**: Point to your `.proto` files and get documentation immediately
- 🎨 **Modern UI**: Beautiful, responsive interface with Tailwind CSS
- 🌙 **Dark Mode**: Built-in light/dark mode toggle with system preference detection
- 📚 **Rich Documentation**: Displays service, method, and field descriptions from proto comments
- 🔗 **HTTP Mappings**: Shows `google.api.http` annotations and generates example requests
- 📋 **Copy-Paste Ready**: One-click copy for `curl` and `grpcurl` commands
- 🔍 **Type Navigation**: Deep linking between services, methods, and types
- 📱 **Mobile Friendly**: Responsive design that works on all devices

## Quick Start

### Installation

```bash
git clone <repository-url>
cd reflect
go build ./cmd/reflect
```

### Usage

```bash
# Basic usage - point to your proto files
./reflect --proto-root=./protos

# With include paths for dependencies
./reflect --proto-root=./protos --proto-include=./third_party

# Custom port
./reflect --proto-root=./protos --addr=:8080
```

Then open http://localhost:8080 in your browser.

## Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `--proto-root` | Root directory containing `.proto` files | Required |
| `--proto-include` | Additional include directories (can be used multiple times) | None |
| `--addr` | Address to listen on | `:8080` |

## Example Proto Files

Here's what your proto files should look like to get the best documentation:

```protobuf
syntax = "proto3";

package echo.v1;

import "google/api/annotations.proto";

// EchoService provides simple echo functionality with HTTP annotations.
service EchoService {
  // Echo echoes back the received message.
  rpc Echo(EchoRequest) returns (EchoResponse) {
    option (google.api.http) = {
      post: "/v1/echo"
      body: "*"
    };
  }
}

// EchoRequest contains the message to echo.
message EchoRequest {
  // The message to echo back.
  string message = 1;
}

// EchoResponse contains the echoed message.
message EchoResponse {
  // The echoed message.
  string message = 1;
}
```

## Architecture

Reflect consists of several key components:

### Core Components

- **Descriptor Package** (`internal/descriptor/`): Loads and parses `.proto` files using `protoparse`
- **Registry** (`internal/descriptor/registry.go`): Indexes services, methods, and types with full names
- **Documentation Models** (`internal/docs/`): Builds view models for templates
- **Web Server** (`internal/server/`): Serves HTML documentation with Tailwind CSS styling

### Key Features

- **Comment Extraction**: Automatically extracts and displays proto comments as descriptions
- **HTTP Annotation Support**: Shows REST API mappings from `google.api.http` options
- **Example Generation**: Creates ready-to-use `curl` and `grpcurl` commands
- **Type Linking**: Deep navigation between related types and services

## Development

### Prerequisites

- Go 1.21+
- Node.js (for Tailwind CSS compilation)

### Building CSS

The project uses Tailwind CSS for styling. To rebuild the CSS:

```bash
npm install
npm run build-css

# Or watch for changes during development
npm run watch-css
```

### Running Tests

```bash
go test ./...
```

### Project Structure

```
reflect/
├── cmd/reflect/           # Main application entry point
├── internal/
│   ├── descriptor/        # Proto file loading and registry
│   ├── docs/             # Documentation model builders
│   └── server/           # Web server and templates
├── ARCHITECTURE.md       # Detailed architecture documentation
├── CHECKLIST.md          # Development progress tracking
└── README.md            # This file
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

See [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Go](https://golang.org/) and [Tailwind CSS](https://tailwindcss.com/)
- Uses [protoparse](https://github.com/jhump/protoreflect/tree/master/desc/protoparse) for proto file parsing
- Inspired by the need for better protobuf documentation tools
