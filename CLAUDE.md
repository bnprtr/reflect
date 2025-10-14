# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Reflect is a protobuf API documentation server that parses `.proto` files and generates rich, navigable documentation for Connect/gRPC services. It provides a web UI built with Go templates, HTMX, Alpine.js, and Tailwind CSS.

## Build and Test Commands

### Building
```bash
# Build the binary
go build -o reflect ./cmd/reflect

# Build CSS assets (required for UI changes)
npm run build-css

# Watch CSS during development
npm run watch-css
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/descriptor

# Run a single test
go test ./internal/descriptor -run TestLoadDirectory

# Run tests with verbose output
go test -v ./...
```

### Running
```bash
# Run the server with a proto directory
./reflect -addr :8080 -proto-root /path/to/protos

# With additional include paths for imports
./reflect -addr :8080 -proto-root /path/to/protos -proto-include /path/to/shared/protos

# Dev mode (reserved for future use)
./reflect -addr :8080 -proto-root /path/to/protos -dev
```

## Architecture

### Core Components

**Descriptor Package** (`internal/descriptor/`)
- `loader.go`: Discovers and loads `.proto` files from a directory
- `parser.go`: Parses proto files using `github.com/jhump/protoreflect/desc/protoparse`, converts to `FileDescriptorSet`
- `registry.go`: Builds an indexed registry with fast lookups for services, methods, messages, and enums by fully-qualified name
- Comment extraction: Parses source code info to index comments for documentation

**Docs Package** (`internal/docs/`)
- `model.go`: View models for rendering documentation (Index, ServiceView, MethodView, MessageView, EnumView)
- Transforms protobuf descriptors into presentation-friendly structures
- Handles sorting, formatting, and example generation

**Server Package** (`internal/server/`)
- `server.go`: Chi router setup, embedded templates and static assets via `go:embed`
- `handlers_docs.go`: HTTP handlers for documentation pages and HTMX partials
- Routes: `/` (home), `/services/{fullName}`, `/methods/*`, `/types/{fullName}`, `/partial/types/*`

**Main Entry Point** (`cmd/reflect/main.go`)
- CLI flag parsing for address, proto-root, proto-include paths
- Loads proto descriptors and starts HTTP server

### Data Flow

1. **Proto Loading**: `LoadDirectory` → discovers .proto files → `parseFiles` → builds `Registry`
2. **Registry Building**: Indexes all services, methods, messages, enums with FQN → extracts comments from source code info
3. **Doc Generation**: Handlers call `docs.Build*View()` → queries registry → returns view models
4. **Rendering**: Go templates render HTML with embedded HTMX for dynamic loading

### Key Design Patterns

**Registry Pattern**: The `Registry` type provides fast O(1) lookups by fully-qualified name (FQN):
- Services: `pkg.ServiceName`
- Methods: `pkg.ServiceName/MethodName`
- Messages/Enums: `pkg.TypeName`

**Comment Indexing**: Uses protobuf SourceCodeInfo paths to extract leading comments and map them to FQNs. Path format example:
- Service: `[6, serviceIndex]`
- Method: `[6, serviceIndex, 2, methodIndex]`
- Message: `[4, messageIndex]`
- Field: `[4, messageIndex, 2, fieldIndex]`

**Template Embedding**: Templates and CSS are embedded using `go:embed` for single-binary distribution

## Testing Structure

Test files follow Go conventions with `_test.go` suffix. The `internal/descriptor/testdata/` directory contains proto files for testing:
- `basic/`: Simple service definitions
- `import/`: Tests for proto imports
- `wkt/`: Well-known types (e.g., google/protobuf/timestamp.proto)
- `http/`: HTTP annotation tests (future)

## Dependencies

- `github.com/go-chi/chi/v5`: HTTP router
- `github.com/jhump/protoreflect`: Proto parsing and reflection
- `google.golang.org/protobuf`: Protobuf types and descriptor handling
- Tailwind CSS (dev dependency): CSS utility framework

## Important Implementation Details

### Proto Import Resolution
The parser requires relative paths from include directories. `findRelativePath` in `parser.go:58` converts absolute file paths to relative paths by checking each include path.

### Nested Type Indexing
Messages and enums can be nested. The registry recursively indexes nested types using `indexMessages` and `indexEnums` functions in `registry.go:95-116`.

### Method Naming Convention
Methods use the format `ServiceFullName/MethodName` (e.g., `echo.v1.EchoService/Echo`). This is handled in `registry.go:75` during registry building.

### Template Function Map
Custom template functions are registered in `server.go:27` (e.g., `contains` for string matching).

## Future Plans (from ARCHITECTURE.md)

- Try-it functionality with Connect/gRPC invocation
- Environment configuration with allowlists
- Streaming support (server/client/bidirectional)
- Security features: CSRF, rate limiting, header filtering
- File upload and paste interfaces for proto sources
