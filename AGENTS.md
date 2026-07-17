# AGENTS.md

Onboarding guide for AI agents working in the kessel-sdk-go repository. This document covers cross-cutting conventions and architectural context. Domain-specific details live in the directory-local guideline files linked below.

## Guidelines Index

| File | Scope |
|------|-------|
| [kessel/auth/GUIDELINES.md](kessel/auth/GUIDELINES.md) | OAuth2 client credentials, token caching, OIDC discovery, AuthRequest interface |
| [kessel/rbac/v2/GUIDELINES.md](kessel/rbac/v2/GUIDELINES.md) | REST workspace client, gRPC iterator, utility constructors |
| [kessel/inventory/internal/builder/GUIDELINES.md](kessel/inventory/internal/builder/GUIDELINES.md) | Generic ClientBuilder[C], auth modes, three-value return |
| [examples/GUIDELINES.md](examples/GUIDELINES.md) | Standalone example binaries, env config, error handling |

## Repository Layout

```
kessel/
  auth/             # OAuth2 client credentials, OIDC discovery, AuthRequest interface
  config/           # CompatibilityConfig with functional options (legacy pattern)
  console/          # Console identity helpers (PrincipalFromRHIdentity)
  grpc/             # OAuth2 PerRPCCredentials wrapper for gRPC
  inventory/
    internal/builder/  # Generic ClientBuilder[C] (Go generics)
    v1/                # Generated: health service only (stable)
    v1beta1/           # Generated: legacy per-resource-type services
    v1beta2/           # Generated: current unified API + client_builder.go (hand-written)
  rbac/v2/          # Hand-written: REST workspace client + v1beta2 utility constructors
examples/
  grpc/             # gRPC client examples (standalone binaries)
  rbac/             # RBAC workspace examples
  console/          # Console identity examples
```

## Generated vs. Hand-Written Code

This is the most important distinction in the repo. Violating it wastes effort and breaks CI.

**Generated (never edit):** Every `*.pb.go` and `*_grpc.pb.go` file under `kessel/inventory/`. These are regenerated from `buf.build/project-kessel/inventory-api` via `make generate` (runs `buf generate`). A scheduled GitHub Actions workflow (`buf-generate.yml`) runs this every 6 hours and opens a PR automatically.

**Generation toolchain:** `buf.gen.yaml` configures two remote plugins -- `buf.build/protocolbuffers/go` (message types) and `buf.build/grpc/go` (service stubs). Both use `paths=source_relative` so output mirrors the proto package path. Each proto message gets its own `<snake_case_name>.pb.go` file; each service gets a `<service_name>_grpc.pb.go` plus a companion `.pb.go` for service descriptor registration.

**Hand-written (where all new logic goes):** `kessel/auth/`, `kessel/config/`, `kessel/grpc/`, `kessel/inventory/internal/builder/`, `kessel/inventory/v1beta2/client_builder.go`, `kessel/rbac/v2/`, and `examples/`.

When in doubt, check if the file has a `// Code generated` header comment. If it does, do not edit it. Protobuf field validation (`buf/validate` annotations) is enforced server-side only -- the SDK does not run client-side protobuf validation.

## Build, Test, and Lint Commands

```bash
make test            # go test -v ./kessel/...
make test-coverage   # generates coverage.out + coverage.html
make lint            # golangci-lint via Docker/Podman container
make build           # compiles example binaries into bin/
make generate        # buf generate (regenerate protobuf files)
make fmt             # go fmt ./...
make mod-tidy        # go mod tidy
```

CI runs `make lint` and `make test` (via separate workflows) on every push/PR to `main`. Both must pass before merge.

## Go Version and Module Path

- **Go 1.25** (specified in `go.mod`; CI reads from `go.mod` via `go-version-file`)
- Module path: `github.com/project-kessel/kessel-sdk-go`
- The codebase uses generics (supported since Go 1.18) and Go 1.23+ features: `iter.Seq2`, range-over-func

## API Version to Use

Always use `v1beta2` for new code. It is the current active API version with the unified `KesselInventoryService`. The `v1beta1` package is legacy (per-resource-type services) and `v1` contains only health endpoints. Never mix types from different API versions in the same call.

## Code Style and Naming Conventions

- **Functional options pattern:** Configuration uses `WithXxx` functions returning a closure (see `kessel/config/config.go`). Follow this for any new config.
- **Builder pattern (fluent):** `ClientBuilder` methods return `*ClientBuilder[C]` for chaining. Each method is a single self-contained mutation. See [builder GUIDELINES.md](kessel/inventory/internal/builder/GUIDELINES.md) for details.
- **Package exports:** Struct fields holding secrets are unexported (e.g., `clientId`, `clientSecret` in `OAuth2ClientCredentials`). Use constructor functions, not direct struct literals.
- **Variable naming in tests:** Loop variable is always `tt`, never `tc` or `test`. Subtest names are lowercase with spaces.
- **Import aliasing:** Use the version as the alias when importing versioned packages: `v1beta2 "...kessel/inventory/v1beta2"`, `v2 "...kessel/rbac/v2"`. Use `kesselgrpc` to alias `kessel/grpc` (avoids conflict with `google.golang.org/grpc`).
- **File naming:** Hand-written files use `snake_case.go`. Test files are `<name>_test.go` in the same package (white-box testing).

## Linter Configuration

`.golangci.yaml` enables: `govet`, `errcheck`, `ineffassign`, `staticcheck`, `unused`. CI uses `golangci-lint` v2.x. Examples are linted individually (each file has its own `main()` function), not as a batch.

## Error Handling Conventions

These rules span all hand-written packages. For package-specific error details, see the relevant GUIDELINES.md.

### Error wrapping strategy

The repo uses three wrapping levels depending on context:

1. **Wrap with `%w` and a contextual prefix** -- SDK-internal code that adds context to an upstream error. The prefix describes the SDK operation that failed (e.g., `"failed to create gRPC client: %w"`). Used in `kessel/inventory/internal/builder/` and `kessel/rbac/v2/list_workspaces.go`.
2. **Wrap with `%v`** -- HTTP-layer code in `kessel/rbac/v2/workspace.go` uses `%v` (not `%w`) for body read and JSON unmarshal errors. Follow this established convention for that file.
3. **Return unwrapped** -- When the SDK has nothing meaningful to add, return the error directly. This is the convention in `kessel/auth/` and `kessel/grpc/`. Do not wrap errors from `client.Discover`, `client.CallTokenEndpoint`, or `credentials.GetToken`.

### Validation errors

Builder validation uses plain `fmt.Errorf` with no wrapping (e.g., `fmt.Errorf("target URI is required")`). HTTP status validation embeds the operation context and raw HTTP status string.

### gRPC status codes

All callers must use `status.FromError(err)` to inspect gRPC errors, handling the `!ok` branch for non-gRPC errors. See the standard switch pattern in the [examples GUIDELINES.md](examples/GUIDELINES.md).

### Bulk response per-item errors

Bulk endpoints (`CheckBulk`, `CheckSelfBulk`, `CheckForUpdateBulk`) return both a response-level gRPC error and per-item errors via `google.rpc.Status` fields on each pair. Callers must handle both: check `pair.GetItem()` vs `pair.GetError()`.

### Stream errors

When consuming streaming RPCs directly (without an iterator wrapper), check for `io.EOF` to detect end-of-stream. Wrap any other stream error with `%w`: `fmt.Errorf("error receiving from stream: %w", err)`. Always drain or cancel the stream to avoid leaking the underlying HTTP/2 stream.

### Zero-value return convention

When returning an error, return the zero value of the success type: `nil` for pointer types, `RefreshTokenResponse{}` for struct types, `zero` (via `var zero C`) for generic three-value returns.

## Testing Conventions

These conventions apply across all hand-written packages. For package-specific test patterns, see the relevant GUIDELINES.md.

### Assertion strategy by package

| Packages | Library | Rule |
|----------|---------|------|
| `kessel/auth`, `kessel/config`, `kessel/grpc` | stdlib only | `t.Errorf`, `t.Error`, `t.Fatal`, `t.Fatalf`. Do not introduce testify. |
| `kessel/rbac/v2` | testify | `require` for preconditions, `assert` for assertions. |
| New packages | testify preferred | Unless the package is low-level infrastructure (auth, config, grpc). |

### Table-driven tests

Most tests use table-driven patterns with `t.Run`. The loop variable is always `tt`. Subtest names are lowercase with spaces. Include `expectedError bool` for error case testing. Use `validateReq func(t *testing.T, ...)` fields when asserting captured request properties; guard with `if tt.validateReq != nil`.

### White-box testing

All test files use the same package as the code under test (e.g., `package auth`, not `package auth_test`). This gives access to unexported types and fields. Do not use `_test` package suffixes.

### Mocking

All mocks are hand-written in `_test.go` files. Do not add mockgen, gomock, or any code generation for mocks. Mocking patterns:
- **HTTP services:** `httptest.NewServer`. Never make real HTTP calls in tests.
- **gRPC clients:** Embed the generated client interface, override only methods under test.
- **Auth:** Implement `auth.AuthRequest` interface with a mock struct.
- **Error types:** Define minimal error structs with a `message` field.

### Test error handling

Use `t.Fatal` / `t.Fatalf` only for setup failures that make the test meaningless. Use `t.Errorf` for assertion failures so remaining checks execute.

## Security and Credential Handling

These rules apply across the entire repo. For package-specific auth details, see [auth GUIDELINES.md](kessel/auth/GUIDELINES.md) and [builder GUIDELINES.md](kessel/inventory/internal/builder/GUIDELINES.md).

### TLS configuration

- Default transport is TLS with system CA pool (`&tls.Config{}`). This is the secure default -- do not change it.
- Custom `*tls.Config` should always set `MinVersion: tls.VersionTLS12`.
- Never set `InsecureSkipVerify: true` in non-test code.
- The `CompatibilityConfig.TLSConfig` field is tagged `json:"-"` so it is never serialized.

### HTTP client injection

Every function that makes HTTP calls accepts an optional `*http.Client`. If nil, it falls back to `http.DefaultClient`. Do not create new `http.Client` instances inside SDK functions. The caller controls timeouts, TLS, and transport settings.

### Environment variables

| Variable | Purpose | Required |
|----------|---------|----------|
| `KESSEL_ENDPOINT` | gRPC server address (host:port) | Yes |
| `AUTH_CLIENT_ID` | OAuth2 client ID | For authenticated flows |
| `AUTH_CLIENT_SECRET` | OAuth2 client secret | For authenticated flows |
| `AUTH_DISCOVERY_ISSUER_URL` | OIDC issuer for discovery | For authenticated flows |

Never read `AUTH_CLIENT_ID` or `AUTH_CLIENT_SECRET` at import time. Load them at call time via `os.Getenv` or `godotenv/autoload`. The `.env` file is in `.gitignore` -- never commit credentials.

### Connection lifecycle

Always defer connection close immediately after a successful `Build()` call. Capture and log the close error instead of discarding it:

```go
defer func() {
    if closeErr := conn.Close(); closeErr != nil {
        log.Printf("Failed to close gRPC client: %v", closeErr)
    }
}()
```

Do not use a bare `defer conn.Close()` -- it silently drops close errors. The caller owns the connection. Reuse a single client/connection for the application's lifetime -- `grpc.NewClient` supports multiplexing. Do not create a new `ClientBuilder`/`Build()` per request.

### Dependency boundaries

- `zitadel/oidc/v3` handles OIDC discovery and token endpoint calls. Do not reimplement OIDC discovery.
- `go-jose/go-jose/v4` (transitive via zitadel) handles JWT/JWS. Do not add separate JWT libraries.
- Use `credentials.NewTLS` and `insecure.NewCredentials` from `google.golang.org/grpc` -- do not use raw `net/tls` for gRPC connections.

## Performance Notes

- **Token caching:** Share a single `*OAuth2ClientCredentials` instance. Creating multiple instances defeats caching and causes redundant token requests. See [auth GUIDELINES.md](kessel/auth/GUIDELINES.md) for the generation counter pattern.
- **ForceRefresh:** Only use `GetTokenOptions.ForceRefresh = true` after receiving a 401/403 from the server. Never force-refresh preemptively.
- **Bulk operations:** Prefer `CheckBulk` / `CheckSelfBulk` / `CheckForUpdateBulk` over loops of single checks. Each bulk endpoint is a single unary RPC.
- **Strongly consistent checks:** `CheckForUpdate` and `CheckForUpdateBulk` bypass server-side caches. Use them only for pre-mutation authorization (write, delete). For read-path filtering, use `Check` / `CheckBulk`.
- **Message size limits:** `CompatibilityConfig` defaults to 4 MB for send and receive. The `ClientBuilder` does not read `CompatibilityConfig` -- if using the builder, message size limits follow gRPC defaults unless overridden with per-RPC call options.

## Maintaining Examples

When public API surface is added or changed (new exported functions, new builder methods, changed method signatures, new service operations), update the `examples/` directory so that examples stay accurate and demonstrate current usage. See [examples GUIDELINES.md](examples/GUIDELINES.md) for conventions and registration steps.

## Common Pitfalls

1. **Editing generated protobuf files.** They will be overwritten by `buf generate`. Changes to API types must go through the upstream proto definitions at `buf.build/project-kessel/inventory-api`.

2. **Constructing `OAuth2ClientCredentials` as a struct literal.** The fields are unexported. Always use `auth.NewOAuth2ClientCredentials(...)`. See [auth GUIDELINES.md](kessel/auth/GUIDELINES.md).

3. **Copying `OAuth2ClientCredentials` by value.** It contains a `sync.RWMutex`. Always pass by pointer (`*OAuth2ClientCredentials`).

4. **Using `CompatibilityConfig` with `ClientBuilder`.** They are separate systems. `ClientBuilder` does not read `CompatibilityConfig`. Use one or the other, not both.

5. **Adding `grpc.DialOption` hooks to `ClientBuilder`.** The builder has no `WithDialOptions` method by design. Custom call options should be passed per-RPC. See [builder GUIDELINES.md](kessel/inventory/internal/builder/GUIDELINES.md).

6. **Mixing the two OAuth2 gRPC adapters.** `kesselgrpc.OAuth2CallCredentials` (exported, always requires TLS) is for use with `.Authenticated()`. The internal `oauth2PerRPCCreds` is used automatically by `.OAuth2ClientAuthenticated()`. Don't use both. See [builder GUIDELINES.md](kessel/inventory/internal/builder/GUIDELINES.md).

7. **Forgetting `defer conn.Close()` after `Build()`.** The caller owns the connection. Leaking it leaks the underlying HTTP/2 transport.

8. **Using `_test` package suffix in test files.** All tests in this repo use white-box testing (same package name). This gives access to unexported types and is the established convention.

9. **Introducing mockgen or gomock.** All mocks are hand-written in `_test.go` files. Do not add mock generation tools.

10. **Using `require` from testify in `kessel/auth`, `kessel/config`, or `kessel/grpc`.** Those packages use stdlib testing only. Only `kessel/rbac/v2` uses testify.

## PR and CI Expectations

- All PRs target `main`
- CI runs two workflows in parallel: `golangci-lint` (lint) and `CI Build and Test` (build + test)
- The build-test workflow uses concurrency groups and cancels in-progress runs on PR update
- Dependabot creates daily PRs for both Go module and GitHub Actions version updates
- Protobuf regeneration PRs are created automatically on the `buf-generate-update` branch

## Release Process

Releases follow semantic versioning. SDK versions across languages are independent. The process is: run quality checks, commit any generated code changes, tag `vX.Y.Z`, push the tag, create a GitHub release with `gh release create`. Go modules are consumed directly from GitHub tags -- no separate registry publish step.
