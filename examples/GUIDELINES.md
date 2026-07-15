# Examples Directory Guidelines

Rules for writing and maintaining standalone example binaries in `examples/`.

## Directory Structure

| Subdirectory | Purpose |
|---|---|
| `grpc/` | gRPC client usage -- ClientBuilder auth modes, inventory operations |
| `rbac/` | REST workspace API helpers (FetchDefaultWorkspace, ListWorkspaces) |
| `console/` | Console identity helpers (PrincipalFromRHIdentity) |

## File and Package Rules

- Every example file is `package main` with its own `main()` function.
- One file per example -- no multi-file packages within a subdirectory.
- File naming: `snake_case.go` matching the operation (e.g., `report_resource.go`, `check_bulk.go`).
- The `main()` function delegates to a named function: `func main() { reportResource() }`.

## Environment Variables

Examples that need configuration read from environment variables via `os.Getenv`. Available variables (defined in `.env.sample`):

| Variable | Purpose |
|---|---|
| `KESSEL_ENDPOINT` | gRPC server address (e.g., `localhost:9000`) |
| `AUTH_CLIENT_ID` | OAuth2 client ID |
| `AUTH_CLIENT_SECRET` | OAuth2 client secret |
| `AUTH_DISCOVERY_ISSUER_URL` | OIDC issuer URL for discovery |

### godotenv/autoload Import

Examples that read env vars typically import `_ "github.com/joho/godotenv/autoload"` for automatic `.env` loading. Place this import in a separate block (blank-line-separated) from other imports. Some examples (e.g., `authenticated.go`, `insecure.go`, `check_bulk.go`) omit this import because they are designed to work without a `.env` file or rely on environment variables being set externally -- follow the pattern of whichever auth mode you are demonstrating.

The `.env` file is gitignored. Running `make build` auto-copies `.env.sample` to `.env` if it does not exist.

## Import Aliasing

When importing versioned SDK packages, use consistent aliasing:

```go
v2 "github.com/project-kessel/kessel-sdk-go/kessel/rbac/v2"
kesselgrpc "github.com/project-kessel/kessel-sdk-go/kessel/grpc"
```

The `kesselgrpc` alias avoids collision with `google.golang.org/grpc`. For `kessel/inventory/v1beta2`, examples currently import directly without an alias, though SDK code uses `v1beta2` as the alias.

## Error Handling Conventions

Four cases of error handling are used in examples, depending on what failed:

1. **OIDC discovery errors**: Use `panic(err)`. Discovery failure means the environment is misconfigured and there is nothing to clean up.
2. **ClientBuilder errors**: Use `log.Fatal("Failed to create gRPC client:", err)`.
3. **gRPC call errors**: Check `status.FromError(err)` and switch on gRPC status codes (`codes.Unavailable`, `codes.PermissionDenied`, default). Use `log.Fatal` for each branch.
4. **Console/utility errors**: Use `panic(err)` for identity parsing failures.

Do not use `fmt.Println` for errors. Do not silently ignore errors.

## Connection Cleanup

Always defer `conn.Close()` immediately after a successful `Build()` call. Log close errors instead of discarding them:

```go
defer func() {
    if closeErr := conn.Close(); closeErr != nil {
        log.Printf("Failed to close gRPC client: %v", closeErr)
    }
}()
```

Every gRPC example in the repo uses this exact pattern. Do not use a bare `defer conn.Close()`.

## Example Structure Pattern

All gRPC examples follow this three-phase structure:

1. **Setup**: Construct credentials (if authenticated), build the client via `ClientBuilder`, defer connection close.
2. **Execute**: Build the request struct, call the API method.
3. **Output**: Print the response with `fmt.Printf("... response: %+v\n", response)`.

## Helper Patterns

- `addr[T any](t T) *T` -- a generic helper for creating pointers to literals (see `report_resource.go`). Define it locally in the file that needs it; do not create a shared utils file.
- RBAC utility constructors (`v2.PrincipalSubject`, `v2.WorkspaceResource`) -- use these instead of building `ResourceReference` and `SubjectReference` structs manually when the resource type is an RBAC concept.

## Adding a New Example

1. Create a single `.go` file in the appropriate subdirectory (`grpc/`, `rbac/`, or `console/`).
2. Add a `go build -o bin/<output-name> ./examples/<subdir>/<file>.go` line to the `build` target in the root `Makefile`.
3. Run `make build` to verify the example compiles.
4. If creating a new subdirectory, the CI lint workflow (`examples/*/*.go` glob) picks it up automatically, but the Makefile `build` target needs an explicit entry.

## What Examples Are Not

- Examples are not tests. They require a live Kessel server and are never run by `make test`.
- Examples are build-checked only (`make build` and CI lint).
- Do not add `_test.go` files under `examples/`.

## API Version

Always use `v1beta2` for new gRPC examples. The `v1beta1` package is legacy. Never mix types from different API versions in the same example.
