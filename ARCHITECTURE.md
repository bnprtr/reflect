# Reflect Architecture and Implementation Plan

## Goals

- Render rich, navigable API docs for Connect/gRPC (services, methods, messages, enums, options, comments).
- Accept protobuf input from a local directory, file upload, or pasted text.
- Provide secure “Try it” with environment allowlist and server-side proxy for Connect and gRPC.
- Minimal-JS UI: HTMX + Alpine + Tailwind, server-rendered HTML.

## Key Decisions

- UI: Server-rendered Go templates + HTMX partials + Alpine interactions; Tailwind for styling (embedded via go:embed).
- Invocation path: All “Try it” proxied through server for auth, CORS, SSRF protection, auditing.
- Environments: Multiple named environments (dev/stage/prod), each with base URL, TLS settings, default headers.
- Transports: Phase 1 Connect(JSON) + gRPC (unary); Phase 2 streaming (server/client/bidi) + gRPC-web (if needed) via pluggable transport.

## High-Level Architecture

- CLI/server binary `reflect` hosts a web UI and descriptor engine.
- Descriptor engine parses `.proto` sources into a `DescriptorRegistry` used for docs and dynamic invocation.
- UI layer renders pages and HTMX fragments; forms post to server for example generation and invocation.
- Try-It proxy encodes JSON to protobuf using descriptors, calls upstream via Connect or gRPC, returns structured results.
- Config loader manages named environments and secrets; enforcement layer validates requests against allowlists.

## Repo Structure

```
/Users/benporter/code/reflect/
  cmd/reflect/main.go                    # CLI to start server
  internal/config/config.go              # Load reflect.yaml; envs, defaults, secrets
  internal/descriptor/
    loader.go                            # Loaders: dir, upload, paste, descriptor set
    parser.go                            # Parse -> FileDescriptorSet (protoparse)
    registry.go                          # Index services, methods, messages
    examplejson.go                       # Generate example JSON for any message
  internal/docs/
    model.go                             # Doc view models (service/method/field)
    search.go                            # Simple search index
  internal/tryit/
    invoker.go                           # Transport-agnostic Invoke interface
    connectinvoker.go                    # Connect JSON unary (phase 1)
    grpcinvoker.go                       # gRPC unary via grpc-go (dynamic)
    streaming.go                         # Streaming (phase 2)
    sanitize.go                          # Header allowlist, redaction
  internal/server/
    server.go                            # Router, middleware (chi), security
    handlers_docs.go                     # Pages + HTMX fragments
    handlers_tryit.go                    # Proxy endpoints
    handlers_source.go                   # Upload/paste, dir-reload (dev)
    templates/                           # HTML templates (go:embed)
    static/                              # Compiled Tailwind CSS/JS (go:embed)
  internal/security/
    auth.go                              # Optional OIDC/basic auth for UI
    csrf.go                              # CSRF for mutating endpoints
    ratelimit.go                         # Per-IP/user limits
  web/tailwind.config.js                 # Build config (compiled at release)
  reflect.yaml.example                   # Sample env config
```

## Protobuf Ingestion

- Directory loader: recursive read `.proto`, supports `import` paths; configurable include paths; fsnotify watcher in dev.
- Upload: zip/tar or multiple files; parsed in-memory; not persisted by default.
- Paste: multi-file text area with filenames; parsed in-memory.
- Precompiled: accept `FileDescriptorSet` (`.desc`) or Buf image (`bin`/`json`) for speed.
- Parser: use `github.com/jhump/protoreflect/desc/protoparse` to avoid requiring `protoc`. Convert to `protodesc.FileDescriptor` for downstream use.

## Docs Rendering

- Extract comments, options, HTTP paths (if present), deprecations.
- Pages: overview, service detail, method detail, message/enum detail.
- Example JSON generation for any message type with sensible defaults; toggle required/optional/oneof variants.
- Deep links and copyable method paths, sample cURL/connect-curl snippets.

## Try-It (Proxy) Design

- Request model: { environment, fqMethod, transport, headers, jsonBody, timeout }.
- Validation: method must exist in registry; environment must be allowlisted; headers filtered by allowlist.
- Encoding: parse `jsonBody` into dynamic protobuf using descriptors; for Connect(JSON) encode per protojson; for gRPC use grpc-go dynamic invoke.
- Execution: per-transport invoker with context deadlines and size caps.
- Response: structured object with status, headers (redacted), latency, JSON-decoded response, and raw wire (optional).
- Streaming (phase 2): SSE to browser; backpressure; sample viewers.

## Security Model

- SSRF prevention: environment base URLs are the only allowed upstreams; no arbitrary URLs.
- Header policy: allowlist only; redact sensitive headers from logs/UI; secure secret storage (env vars or OS keychain on macOS).
- CSRF protection on upload, config, and try-it; SameSite cookies; strict Origin/Referer checks.
- Rate limits, timeouts, max body size; audit logging (method, env, timings, redacted headers).
- CORS: UI endpoints same-origin; disable cross-origin unless explicitly configured.
- Optional UI auth: OIDC or Basic for shared installs.

## Routing Sketch

- GET `/` overview; GET `/services/{fullName}`; GET `/methods/{fullName}`; GET `/types/{fullName}`
- POST `/load/upload` (CSRF); POST `/load/paste` (CSRF); POST `/load/dir/reload` (dev)
- POST `/tryit/invoke` (CSRF): body defines environment/method/transport/headers/json
- GET `/partial/...` HTMX fragments for panels, examples, and results

## Config (reflect.yaml)

```yaml
environments:
  - name: dev
    baseURL: https://dev.api.example.com
    transport: connect # default; per-invoke override allowed (connect|grpc|grpc-web)
    tls:
      insecureSkipVerify: false
    defaultHeaders:
      x-api-key: ${REFLECT_DEV_API_KEY}
  - name: prod
    baseURL: https://api.example.com
headerAllowlist: [authorization, x-api-key, x-request-id]
maxRequestBodyBytes: 1048576
requestTimeoutSeconds: 15
```

## Dependencies

- `github.com/go-chi/chi/v5` (router)
- `github.com/jhump/protoreflect/desc/protoparse` (parse .proto)
- `google.golang.org/protobuf` (types, protojson)
- `google.golang.org/grpc` (grpc client)
- `connectrpc.com/connect` (Connect transport; JSON unary)
- `github.com/fsnotify/fsnotify` (watcher)

## Build & Assets

- Tailwind built at release-time into `internal/server/static/app.css`; embed templates/static via `go:embed`.
- Single static binary; no node at runtime.

## Milestones

1. Core docs (dir/paste/upload), service/method pages, example JSON, basic search.
2. Try-It: Connect(JSON) unary + gRPC unary, env allowlist, header allowlist, timeouts, audit logs.
3. Streaming support (Connect/gRPC) via SSE; history and shareable permalinks for requests.
4. Optional: gRPC-web client support (if upstream only exposes gRPC-web), OIDC auth for UI, Buf image ingestion.

## Testing

- Golden tests for example JSON and doc rendering.
- Integration tests against a sample Connect/gRPC server.
- Security tests: SSRF, header filtering, body/timeout limits.