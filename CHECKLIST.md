# Reflect Implementation Checklist

This checklist breaks down the implementation of the Reflect project into phases based on the architecture plan.

## Milestone 1: Core Docs & Protobuf Ingestion

- [x] **Project Skeleton**: Create `cmd/reflect`, all `internal` packages, set up the Chi router, and configure `go:embed` for templates and static assets.
- [ ] **Protobuf Parsing**: Implement the descriptor loader and parser. Support loading from a directory, file uploads, and pasted text. This will populate the `DescriptorRegistry`.
- [ ] **Doc Rendering**: Build the view models and HTML templates for rendering services, methods, messages, and enums. Use HTMX for any partial page updates.
- [ ] **Example Generation**: Implement the logic to generate example JSON request bodies for any given message type.
- [ ] **Search**: Add a lightweight search feature to find services, methods, and messages.
- [ ] **Dev Watcher**: Use `fsnotify` to watch for changes in the proto directory and automatically reload the descriptors in development mode.

## Milestone 2: "Try It" Functionality (Unary)

- [ ] **Configuration**: Implement the `reflect.yaml` loader to configure named environments, base URLs, and security policies.
- [ ] **Connect Invoker**: Implement the "Try It" proxy endpoint for unary Connect RPCs (JSON over HTTP).
- [ ] **gRPC Invoker**: Implement the "Try It" proxy endpoint for unary gRPC RPCs using `grpc-go`'s dynamic invocation.
- [ ] **Security Hardening**: Enforce all security policies: SSRF allowlist for environments, header allowlists, CSRF protection on forms, request size/timeout limits, and audit logging.

## Milestone 3: Streaming & Usability

- [ ] **Streaming Proxy**: Add support for streaming RPCs (server, client, and bidirectional) using Server-Sent Events (SSE) to stream data to the browser.
- [ ] **Request History**: Implement a feature to view and reuse past requests.
- [ ] **Shareable Permalinks**: Create a way to generate a shareable link that pre-fills a request in the UI.

## Milestone 4: Advanced Features & Polish

- [ ] **UI Authentication**: Add optional OIDC or Basic Authentication for securing access to the web UI in shared environments.
- [ ] **gRPC-Web Support**: (Optional) Add gRPC-Web as a supported transport for the "Try It" feature.
- [ ] **Buf Image Ingestion**: (Optional) Add support for loading pre-compiled `FileDescriptorSet`s from Buf images.
- [ ] **Testing**: Write golden tests for doc rendering, integration tests for the proxy, and security tests for the enforcement layer.
