# Error Handling Guidelines

Guidelines for AI agents implementing or reviewing error handling in kessel-sdk-go.

## 1. Error Wrapping Rules

### SDK-internal code: wrap with `%w` and a contextual prefix

When the SDK adds context to an error before returning it, use `fmt.Errorf` with `%w` so callers can unwrap.

```go
return zero, nil, fmt.Errorf("failed to create gRPC client: %w", err)
```

This is the convention in `kessel/inventory/internal/builder/builder.go` and `kessel/rbac/v2/list_workspaces.go`. Prefix messages describe the SDK operation that failed (e.g., `"failed to start stream"`, `"error receiving from stream"`).

### HTTP layer code in `kessel/rbac/v2`: use `%v` for body/unmarshal errors

The `workspace.go` file uses `%v` (not `%w`) when wrapping `io.ReadAll` and `json.Unmarshal` errors. Follow this existing convention for that package:

```go
return nil, fmt.Errorf("error reading response body: %v", err)
return nil, fmt.Errorf("error unmarshalling response: %v", err)
```

### Passthrough: return errors unwrapped from upstream libraries

When the SDK has nothing meaningful to add, return the error directly. This is the convention in `kessel/auth/auth.go`, `kessel/auth/auth_request.go`, and `kessel/grpc/grpc.go`:

```go
return RefreshTokenResponse{}, err   // auth.go
return nil, err                       // grpc.go
```

Do not wrap errors from `client.Discover`, `client.CallTokenEndpoint`, or `credentials.GetToken` -- these are returned as-is.

## 2. Validation Errors

### Builder validation uses plain `fmt.Errorf` (no wrapping)

The `ClientBuilder.Build()` method validates preconditions and returns non-wrapped errors for configuration problems:

```go
return zero, nil, fmt.Errorf("target URI is required")
```

Use this pattern for any new builder-level validation. These are configuration errors, not wrapped upstream failures.

### HTTP status validation produces descriptive errors

Non-200 HTTP responses produce errors that embed the workspace type and HTTP status string:

```go
return nil, fmt.Errorf("error fetching %s workspace - http status %s", workspaceType, response.Status)
```

Follow this format for any new HTTP client code: include the operation context and the raw HTTP status.

### Cardinality validation

When exactly one result is expected, validate the count and include the actual data in the error:

```go
return nil, fmt.Errorf("unexpected number of %s workspaces: %d. %v", workspaceType, len(data), data)
```

### Protobuf validation is server-side only

The generated `.pb.go` files import `buf/validate` annotations, but the SDK does not call a protovalidate validator client-side. Validation is enforced by the server. Do not add client-side protobuf validation unless explicitly requested.

## 3. gRPC Error Handling

### Callers must use `status.FromError` to inspect gRPC errors

All gRPC examples follow this exact pattern for handling errors from inventory client methods:

```go
if st, ok := status.FromError(err); ok {
    switch st.Code() {
    case codes.Unavailable:
        log.Fatal("Service unavailable: ", err)
    case codes.PermissionDenied:
        log.Fatal("Permission denied: ", err)
    default:
        log.Fatal("gRPC connection error: ", err)
    }
} else {
    log.Fatal("Unknown error: ", err)
}
```

When adding new examples or consumer-facing documentation, replicate this structure. Always handle the `!ok` branch (non-gRPC error).

### Bulk response per-item errors

Bulk endpoints (`CheckBulk`, `CheckSelfBulk`, `CheckForUpdateBulk`) return a response-level error via gRPC status AND per-item errors via `google.rpc.Status` fields on each pair. Callers must handle both:

```go
// Response-level: standard gRPC error check on err
// Per-item: check pair.GetError() vs pair.GetItem()
if item := pair.GetItem(); item != nil { ... }
else if err := pair.GetError(); err != nil { ... }
```

### Stream errors

When consuming streaming RPCs, check for `io.EOF` to detect end-of-stream, then wrap any other error with `%w`:

```go
if err == io.EOF {
    break
}
if err != nil {
    yield(nil, fmt.Errorf("error receiving from stream: %w", err))
    return
}
```

## 4. Connection Lifecycle Errors

### `Build()` errors are fatal

Every example treats `ClientBuilder.Build()` failure as fatal (`log.Fatal`). The SDK returns a three-value tuple `(client, conn, error)` -- callers must check the error before using either the client or connection.

### `conn.Close()` errors are logged, not fatal

All examples use this exact deferred cleanup pattern:

```go
defer func() {
    if closeErr := conn.Close(); closeErr != nil {
        log.Printf("Failed to close gRPC client: %v", closeErr)
    }
}()
```

Use `closeErr` as the variable name. Log with `log.Printf`, not `log.Fatal`.

### `response.Body.Close()` errors are discarded

HTTP response bodies are closed with the error explicitly discarded:

```go
defer func() { _ = response.Body.Close() }()
```

## 5. Auth Errors

### Auth errors propagate without wrapping

The auth layer (`auth.go`, `auth_request.go`, `grpc.go`) returns errors from token endpoints and OIDC discovery directly to callers without adding context. This is intentional -- the upstream library errors are already descriptive.

### OIDC discovery failures in examples use `panic`, not `log.Fatal`

When writing example binaries, use `panic(err)` for OIDC discovery failures and `log.Fatal` for inventory client build failures. This asymmetry is the established convention across the existing examples in `examples/grpc/` and `examples/rbac/`.

## 6. Testing Conventions

### Table-driven tests with `expectedError bool`

All test files use table-driven tests with an `expectedError` (or `expectError`) boolean field. The assertion pattern is:

```go
if tt.expectedError {
    if err == nil {
        t.Errorf("Expected error but got none")
    }
} else {
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
}
```

### Error message assertions use `strings.Contains`

When asserting on error message content (rare in this repo), use `strings.Contains`:

```go
if !strings.Contains(err.Error(), "unmarshalling") {
    t.Errorf("Expected unmarshalling error, got: %v", err)
}
```

### Mock error types implement the `error` interface directly

Test mocks define minimal error structs with a `message` field:

```go
type mockStreamError struct {
    message string
}
func (e *mockStreamError) Error() string {
    return e.message
}
```

### Use `t.Fatal` / `t.Fatalf` only for test setup failures

`t.Fatal` is reserved for cases where continuing the test is impossible (e.g., failed to create an HTTP request). Assertion failures use `t.Errorf`.

### Use `assert`/`require` from testify in newer code

The `kessel/rbac/v2` package uses `github.com/stretchr/testify/assert` and `require`. Use `require` for preconditions that would make subsequent assertions meaningless. Use `assert` for the actual assertions under test. Older packages (`auth`, `grpc`, `config`) use stdlib testing only -- follow the existing style when modifying those packages.

## 7. Iterator Error Propagation

The `ListWorkspaces` function returns `iter.Seq2[*Response, error]`. Errors are yielded as the second value, then the iterator returns:

```go
yield(nil, fmt.Errorf("failed to start stream: %w", err))
return
```

Callers consume this by breaking on the first error:

```go
for resp, err := range v2.ListWorkspaces(...) {
    if err != nil {
        log.Fatalf("Error listing workspaces: %v", err)
    }
}
```

Always `return` after yielding an error -- never continue iteration.

## 8. Zero-Value Return Convention

When returning an error from a function, return the zero value of the success type:

- Pointer types: `return nil, err`
- Struct types: `return RefreshTokenResponse{}, err`
- Three-value (generics): `return zero, nil, err` where `var zero C` is declared at function start
