# Testing Guidelines

## Running Tests

```bash
make test              # go test -v ./kessel/...
make test-coverage     # generates coverage.out and coverage.html
```

CI runs `go test -v ./kessel/...` on every push/PR to `main`. Tests must pass before merge.

## Assertion Strategy

The repo uses **two styles** depending on the package. Follow whichever style the package already uses:

1. **stdlib only** (`kessel/auth`, `kessel/config`, `kessel/grpc`): Use `t.Errorf`, `t.Error`, `t.Fatal`, `t.Fatalf` directly. Do not introduce testify into these packages.
2. **testify** (`kessel/rbac/v2`): Use `assert` and `require` from `github.com/stretchr/testify`. Use `require` for preconditions that should abort the test (e.g., length checks before indexing), and `assert` for all other checks.

When creating a new package, prefer testify for new test files unless the package under test is low-level infrastructure (auth, config, grpc).

## Test Organization

### Table-driven tests

Most tests use table-driven patterns with `t.Run` subtests. Follow this structure:

```go
tests := []struct {
    name          string
    // inputs
    expectedError bool
    // validation function (optional)
    validateReq   func(t *testing.T, ...)
}{
    { name: "descriptive case name", ... },
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test body
    })
}
```

Key conventions:
- The loop variable is always `tt` (not `tc`, `test`, etc.)
- The `name` field is lowercase, descriptive, and uses spaces (e.g., `"successful default workspace fetch"`, `"handles stream errors"`)
- Include a `validateReq` or `validateAuth` function field when the test needs to assert properties of captured requests. Set it to `nil` for cases that don't need it, and guard the call: `if tt.validateReq != nil { tt.validateReq(t, ...) }`

### Standalone tests

Simple unit tests for single behaviors can be standalone functions (not table-driven). This is common in `kessel/grpc` and `kessel/config`. Use standalone tests when there is only one interesting case or the setup differs significantly between cases.

## Naming Conventions

- Test functions: `TestTypeName_MethodName_Scenario` or `TestFunctionName_Scenario` (e.g., `TestOAuth2ClientCredentials_GetToken`, `TestFetchDefaultWorkspace`, `TestCallCredentials_GetRequestMetadata_ErrorHandling`)
- For integration-style tests: suffix with `_Integration` (e.g., `TestOAuth2CallCredentials_Integration`)
- Subtest names (in `t.Run`): lowercase with spaces, no `Test` prefix

## Mocking Patterns

### HTTP services: use `net/http/httptest`

All tests that interact with HTTP endpoints use `httptest.NewServer`. Never make real HTTP calls in tests.

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    response := map[string]interface{}{...}
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        t.Errorf("Failed to encode test response: %v", err)
    }
}))
defer server.Close()
```

When the server handler can vary per test case, use a `setupServer func() *httptest.Server` or `serverHandler func(w http.ResponseWriter, r *http.Request)` field in the test table. Always `defer server.Close()` inside the subtest.

### gRPC clients: embed the generated client interface

Mock gRPC clients by embedding the generated client interface and overriding only the methods under test:

```go
type mockInventoryClient struct {
    v1beta2.KesselInventoryServiceClient  // embed the interface
    responses        []*v1beta2.StreamedListObjectsResponse
    err              error
    capturedRequests []*v1beta2.StreamedListObjectsRequest
}
```

For streaming RPCs, implement a mock stream struct that returns canned responses from a slice and `io.EOF` at the end.

### Auth: implement the `auth.AuthRequest` interface

```go
type mockAuthRequest struct {
    token      string
    shouldFail bool
}

func (m *mockAuthRequest) ConfigureRequest(ctx context.Context, request *http.Request) error {
    if m.shouldFail {
        return &mockAuthError{message: "auth failed"}
    }
    request.Header.Set("authorization", "Bearer "+m.token)
    return nil
}
```

### Custom error types for mocks

Define simple error structs for mock failures rather than using `errors.New` or `fmt.Errorf`:

```go
type mockStreamError struct{ message string }
func (e *mockStreamError) Error() string { return e.message }
```

## No External Mock Frameworks

The repo does not use mockgen, gomock, or any code generation for mocks. All mocks are hand-written in `_test.go` files within the same package. Keep mocks local to the test file that uses them.

## Package-Level Testing (White-Box)

All test files use the **same package** as the code under test (e.g., `package auth`, not `package auth_test`). This gives tests access to unexported types and fields (e.g., `oauth2Auth`, `credentials.cachedToken`). Follow this convention; do not use `_test` package suffixes.

## Error Handling in Tests

- Use `t.Fatal` / `t.Fatalf` only for setup failures that make the rest of the test meaningless (e.g., `http.NewRequest` failure, type assertion failure)
- Use `t.Error` / `t.Errorf` for assertion failures so remaining checks still execute
- For expected errors, check `err == nil` and call `t.Error("Expected error but got none")` rather than using `require.Error`
- In stdlib-style tests, format messages as: `t.Errorf("Expected X to be %v, got %v", expected, actual)`

## Test Scope

- Tests cover: config option application, auth token lifecycle (fetch, cache, refresh, expiration, concurrency), gRPC credential wrappers, HTTP API clients, utility/builder functions
- Generated protobuf files (`*.pb.go`) are not tested directly
- Example binaries in `examples/` are not tested; they are build-checked only (`go build`)

## Coverage

- Coverage artifacts (`coverage.out`, `coverage.html`) are gitignored
- Run `make test-coverage` locally to generate an HTML coverage report
- CI does not enforce a coverage threshold

## Linting

CI runs `golangci-lint` with these linters enabled: `govet`, `errcheck`, `ineffassign`, `staticcheck`, `unused`. All test code must pass the linter. Lint is scoped to `./kessel/...` (examples are linted separately).

## Concurrency Tests

When testing concurrent access (e.g., token caching under contention), use goroutines with channels for collecting results, `sync.Mutex` for test-server state, and `time.After` for timeout guards. See `TestConcurrentTokenAccess` in `kessel/auth/auth_test.go` for the pattern.

## Context Usage

- Always pass `context.Background()` or `context.TODO()` in tests
- Test context cancellation explicitly where relevant (create a context, cancel it, assert error)
- Test context timeout with `context.WithTimeout` and a short duration
