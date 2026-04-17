# AGENTS.md

Onboarding guide for AI agents working in the kessel-sdk-go repository. This document covers cross-cutting conventions and architectural context. Domain-specific details live in the guideline files linked below.

## Docs Index

| File | Scope |
|------|-------|
| [docs/security-guidelines.md](docs/security-guidelines.md) | Auth architecture, TLS, token caching, credential handling, environment variables |
| [docs/api-contracts-guidelines.md](docs/api-contracts-guidelines.md) | Protobuf code generation, API versioning, ClientBuilder pattern, domain concepts |
| [docs/error-handling-guidelines.md](docs/error-handling-guidelines.md) | Error wrapping rules, gRPC status codes, stream/bulk error patterns, zero-value returns |
| [docs/testing-guidelines.md](docs/testing-guidelines.md) | Test styles, table-driven patterns, mocking, naming, coverage, linting |
| [docs/performance-guidelines.md](docs/performance-guidelines.md) | Connection reuse, token caching, bulk operations, streaming, message size limits |
| [docs/integration-guidelines.md](docs/integration-guidelines.md) | Client construction, OIDC flow, REST workspace API, RBAC utilities, credential adapters |

## Repository Layout

```
kessel/
  auth/             # OAuth2 client credentials, OIDC discovery, AuthRequest interface
  config/           # CompatibilityConfig with functional options (legacy pattern)
  grpc/             # OAuth2 PerRPCCredentials wrapper for gRPC
  inventory/
    internal/builder/  # Generic ClientBuilder[C] (Go generics)
    v1/                # Generated: health service only (stable)
    v1beta1/           # Generated: legacy per-resource-type services
    v1beta2/           # Generated: current unified API + client_builder.go (hand-written)
  rbac/v2/          # Hand-written: REST workspace client + v1beta2 utility constructors
examples/           # Standalone main() files; each is a separate binary
docs/               # Guideline files for AI agents
```

## Generated vs. Hand-Written Code

This is the most important distinction in the repo. Violating it wastes effort and breaks CI.

**Generated (never edit):** Every `*.pb.go` and `*_grpc.pb.go` file under `kessel/inventory/`. These are regenerated from `buf.build/project-kessel/inventory-api` via `make generate` (runs `buf generate`). A scheduled GitHub Actions workflow (`buf-generate.yml`) runs this every 6 hours and opens a PR automatically.

**Hand-written (where all new logic goes):** `kessel/auth/`, `kessel/config/`, `kessel/grpc/`, `kessel/inventory/internal/builder/`, `kessel/inventory/v1beta2/client_builder.go`, `kessel/rbac/v2/`, and `examples/`.

When in doubt, check if the file has a `// Code generated` header comment. If it does, do not edit it.

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

- **Go 1.24** (specified in `go.mod`; CI reads from `go.mod` via `go-version-file`)
- Module path: `github.com/project-kessel/kessel-sdk-go`
- The codebase uses Go 1.23+ features: `iter.Seq2` (range-over-func), generics

## Code Style and Naming Conventions

- **Functional options pattern:** Configuration uses `WithXxx` functions returning a closure (see `kessel/config/config.go`). Follow this for any new config.
- **Builder pattern (fluent):** `ClientBuilder` methods return `*ClientBuilder[C]` for chaining. Each method is a single self-contained mutation.
- **Package exports:** Struct fields holding secrets are unexported (e.g., `clientId`, `clientSecret` in `OAuth2ClientCredentials`). Use constructor functions, not direct struct literals.
- **Variable naming in tests:** Loop variable is always `tt`, never `tc` or `test`. Subtest names are lowercase with spaces.
- **Import aliasing:** Use the version as the alias when importing versioned packages: `v1beta2 "...kessel/inventory/v1beta2"`, `v2 "...kessel/rbac/v2"`. Use `kesselgrpc` to alias `kessel/grpc` (avoids conflict with `google.golang.org/grpc`).
- **File naming:** Hand-written files use `snake_case.go`. Test files are `<name>_test.go` in the same package (white-box testing).

## Linter Configuration

`.golangci.yaml` enables: `govet`, `errcheck`, `ineffassign`, `staticcheck`, `unused`. CI uses `golangci-lint` v2.x. Examples are linted individually (each file has its own `main()` function), not as a batch.

## Examples Are Standalone Binaries

Each file in `examples/grpc/` and `examples/rbac/` is a standalone `package main` with its own `main()` function. They are built separately by the Makefile and linted individually by CI. When adding a new example:
1. Create a single `.go` file with `package main`
2. Add a `go build -o bin/<name> ./examples/<path>/<file>.go` line to the `build` target in `Makefile`
3. The CI lint workflow loops over `examples/grpc/*.go` and `examples/rbac/*.go` -- new subdirectories need to be added to both the Makefile lint target and `.github/workflows/lint.yml`

## API Version to Use

Always use `v1beta2` for new code. It is the current active API version with the unified `KesselInventoryService`. The `v1beta1` package is legacy (per-resource-type services) and `v1` contains only health endpoints. Never mix types from different API versions in the same call.

## Common Pitfalls

1. **Editing generated protobuf files.** They will be overwritten by `buf generate`. Changes to API types must go through the upstream proto definitions at `buf.build/project-kessel/inventory-api`.

2. **Constructing `OAuth2ClientCredentials` as a struct literal.** The fields are unexported. Always use `auth.NewOAuth2ClientCredentials(...)`.

3. **Copying `OAuth2ClientCredentials` by value.** It contains a `sync.Mutex`. Always pass by pointer (`*OAuth2ClientCredentials`).

4. **Using `CompatibilityConfig` with `ClientBuilder`.** They are separate systems. `ClientBuilder` does not read `CompatibilityConfig`. Use one or the other, not both.

5. **Adding `grpc.DialOption` hooks to `ClientBuilder`.** The builder has no `WithDialOptions` method by design. Custom call options should be passed per-RPC.

6. **Mixing the two OAuth2 gRPC adapters.** `kesselgrpc.OAuth2CallCredentials` (exported, always requires TLS) is for use with `.Authenticated()`. The internal `oauth2PerRPCCreds` is used automatically by `.OAuth2ClientAuthenticated()`. Don't use both.

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
