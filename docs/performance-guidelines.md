# Performance Guidelines

Rules for writing and reviewing performance-sensitive code in kessel-sdk-go.

## Token Caching and Authentication

### Rule: Share a single `OAuth2ClientCredentials` instance across the application
The `OAuth2ClientCredentials` struct caches the access token in memory with mutex protection (`tokenMutex sync.Mutex`). Creating multiple instances defeats caching and causes redundant token requests.

### Rule: Pass `OAuth2ClientCredentials` by pointer, never by value
The struct contains a `sync.Mutex` and a cached token. Copying it by value copies the mutex (which is forbidden) and duplicates the cache. All builder methods (`OAuth2ClientAuthenticated`, `OAuth2AuthRequest`) accept `*OAuth2ClientCredentials`.

### Rule: Understand the token expiration window
Tokens are considered invalid when they are within 300 seconds (5 minutes) of expiry (`expirationWindow`). Do not implement your own expiration logic around `GetToken` -- the built-in 5-minute buffer handles it. If the server omits `expires_in`, the SDK defaults to 3600 seconds.

### Rule: Use `ForceRefresh` sparingly
`GetTokenOptions.ForceRefresh` clears the cached token and forces a network round-trip. Only use it after receiving a 401/403 from the server, never preemptively or on every call.

### Rule: Token refresh acquires the mutex; reads of valid tokens do not
`GetToken` checks `isTokenValid()` before acquiring the lock. This means concurrent callers with a valid cached token proceed without contention. Only expired-token paths serialize through the mutex. Do not add locking around `GetToken` calls externally.

## gRPC Connection Management

### Rule: Always close `*grpc.ClientConn` returned by `Build()`
`ClientBuilder.Build()` returns `(client, *grpc.ClientConn, error)`. The caller owns the connection and must close it. Use `defer conn.Close()` immediately after checking the error, following the pattern in all examples:
```go
client, conn, err := v1beta2.NewClientBuilder(endpoint).Insecure().Build()
if err != nil {
    log.Fatal(err)
}
defer func() {
    if closeErr := conn.Close(); closeErr != nil {
        log.Printf("Failed to close gRPC client: %v", closeErr)
    }
}()
```

### Rule: Reuse a single client/connection for the lifetime of the application
`grpc.NewClient` creates a connection that supports multiplexing. Do not create a new `ClientBuilder`/`Build()` per request. Build once, share the client, close on shutdown.

### Rule: The `ClientBuilder` does not accept `DialOption` hooks
The builder constructs dial options internally (`WithTransportCredentials`, `WithDefaultCallOptions`). There is no `WithDialOptions` method. If you need custom call options, pass them per-RPC via the `opts ...grpc.CallOption` parameter on each method.

## Message Size Configuration

### Rule: Default max message size is 4 MB (send and receive)
`CompatibilityConfig` defaults both `MaxReceiveMessageSize` and `MaxSendMessageSize` to `4 * 1024 * 1024`. Increase these via `WithGRPCMaxReceiveMessageSize` / `WithGRPCMaxSendMessageSize` only when bulk responses or large resource representations require it.

### Rule: The `CompatibilityConfig` is not wired into `ClientBuilder`
`CompatibilityConfig` and its `With*` options exist for compatibility-layer consumers. The `ClientBuilder` in `kessel/inventory/internal/builder/` does **not** read `CompatibilityConfig`. If you use `ClientBuilder`, message size limits are governed by gRPC defaults (4 MB) unless you pass `grpc.MaxCallRecvMsgSize` / `grpc.MaxCallSendMsgSize` as per-RPC call options.

## Bulk Operations

### Rule: Prefer `CheckBulk` / `CheckSelfBulk` / `CheckForUpdateBulk` over loops of single checks
Each bulk endpoint is a single unary RPC. Making N individual `Check` calls incurs N round-trips and N token-header injections. Bulk endpoints batch them into one.

### Rule: `CheckForUpdate` and `CheckForUpdateBulk` are strongly consistent
These bypass caches on the server side. Use them only for pre-mutation authorization (write, delete). For read-path filtering, use the standard `Check` / `CheckBulk` which may benefit from server-side caching.

## Streaming APIs and the Iterator Pattern

### Rule: Use `iter.Seq2` for consuming streamed list results
The SDK provides `ListWorkspaces` (and similar functions) that return `iter.Seq2[*Response, error]`. This is the Go 1.23+ iterator pattern. Consume with a range-over-func loop. The iterator handles pagination, stream lifecycle, and early termination internally.

### Rule: Pagination is automatic inside iterators -- do not manually paginate
`ListWorkspaces` manages continuation tokens internally. It sets `Limit: 1000` per page and re-opens streams with the last continuation token until the server returns an empty token. Callers just `range` over the iterator.

### Rule: Breaking out of the iterator range loop is safe and terminates the stream
The iterator checks the return value of `yield`. If the caller breaks or returns from the range loop, the iterator stops fetching. No explicit cleanup is needed by the caller.

### Rule: When consuming streams directly (without an iterator wrapper), always drain or check for `io.EOF`
If you call `StreamedListObjects` or `StreamedListSubjects` directly, loop on `stream.Recv()` until `io.EOF`. Failing to fully consume or cancel the stream leaks the underlying HTTP/2 stream.

## HTTP Client Reuse

### Rule: Pass a shared `*http.Client` or nil (to use `http.DefaultClient`)
Both `FetchOIDCDiscovery` and `FetchWorkspaceOptions` accept an optional `HttpClient`. Passing nil defaults to `http.DefaultClient`, which has built-in connection pooling. Do not create a new `http.Client` per call. When passing a custom client, ensure it has appropriate timeouts and connection pool settings.

### Rule: Always close HTTP response bodies
The SDK closes response bodies internally (`defer func() { _ = response.Body.Close() }()`). If you extend the SDK with new HTTP-based functions, follow the same pattern. Do not use `defer response.Body.Close()` directly -- use the `_ =` discard form to satisfy linters.

## Code Generation

### Rule: Never edit `*.pb.go` files for performance tuning
All files in `kessel/inventory/v1beta2/` (except `client_builder.go`) and `kessel/inventory/v1beta1/` are generated by `protoc-gen-go` / `protoc-gen-go-grpc`. Changes will be overwritten on the next `buf generate`. Performance-related changes to protobuf types must go through the `.proto` schema definitions.

## Testing Patterns

### Rule: Test concurrent token access with multiple goroutines
The existing `TestConcurrentTokenAccess` test verifies that 5 goroutines sharing one `OAuth2ClientCredentials` instance result in at most 1 server call due to caching. Any new authentication code must include a similar concurrency test.

### Rule: Use `httptest.NewServer` for token endpoint tests, not real servers
All auth tests mock the token endpoint with `httptest.NewServer`. This avoids network flakiness and allows controlling response timing (e.g., `time.Sleep(50 * time.Millisecond)` to simulate latency).
