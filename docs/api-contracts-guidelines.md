# API Contracts Guidelines

## Code Generation

### Source of Truth
- All protobuf types are generated from the external BSR module `buf.build/project-kessel/inventory-api`, not from local `.proto` files. This repo contains only the generated Go output.
- Never hand-edit any `*.pb.go` or `*_grpc.pb.go` file. They are regenerated in full by `buf generate`.

### Generation Toolchain
- **buf.gen.yaml** configures two remote plugins: `buf.build/protocolbuffers/go` (message types) and `buf.build/grpc/go` (service stubs). Both use `paths=source_relative` so output mirrors the proto package path.
- Run `make generate` (which calls `buf generate`) to regenerate. A GitHub Actions workflow (`buf-generate.yml`) runs this on a schedule (every 6 hours) and opens a PR automatically on branch `buf-generate-update`.

### Generated File Layout
- Output lands at `kessel/inventory/<version>/` mirroring the proto package path.
- Each proto message gets its own `<snake_case_name>.pb.go` file. Each service gets a `<service_name>_grpc.pb.go` file plus a companion `<service_name>.pb.go` for the service descriptor registration.

## API Versioning

### Version Hierarchy
- **`kessel/inventory/v1`** -- Stable. Contains only the health service (`KesselInventoryHealthService`: `GetLivez`, `GetReadyz`).
- **`kessel/inventory/v1beta1`** -- Legacy. Per-resource-type services with domain-specific messages (e.g., `KesselK8sClusterService` with `CreateK8sCluster`/`UpdateK8sCluster`/`DeleteK8sCluster`). Subdivided into `resources/`, `relationships/`, `authz/` sub-packages.
- **`kessel/inventory/v1beta2`** -- Current/active. Single unified `KesselInventoryService` with generic resource operations (`ReportResource`, `DeleteResource`) and authz checks (`Check`, `CheckSelf`, `CheckForUpdate`, plus `*Bulk` variants and `StreamedList*`).

### Import Convention
- Always import the active API version (`v1beta2`) for new code. Import the specific package, not parent paths:
  ```go
  import "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
  ```
- The `kessel/rbac/v2` package already depends on `v1beta2` types -- never mix `v1beta1` and `v1beta2` types in the same call.

## Hand-Written SDK Code (Non-Generated)

### Package Boundaries
The following packages contain hand-written code and are the only places where new non-generated logic should go:

| Package | Purpose |
|---|---|
| `kessel/inventory/internal/builder` | Generic gRPC client builder (generics-based) |
| `kessel/inventory/v1beta2/client_builder.go` | Type alias wiring `ClientBuilder` to `KesselInventoryServiceClient` |
| `kessel/rbac/v2` | REST-based RBAC workspace helpers + v1beta2 utility constructors |
| `kessel/grpc` | OAuth2 `PerRPCCredentials` wrapper |
| `kessel/auth` | OAuth2 client credentials, OIDC discovery, `AuthRequest` interface |
| `kessel/config` | `CompatibilityConfig` with functional options pattern |

### ClientBuilder Pattern
- Each service version exposes a `NewClientBuilder(target)` that returns a `*ClientBuilder[T]` (type alias from the internal generic builder).
- Builder methods chain: `.Insecure()`, `.Authenticated(creds, tlsCreds)`, `.OAuth2ClientAuthenticated(oauthCreds, tlsCreds)`, `.Unauthenticated(tlsCreds)`.
- `.Build()` returns `(client, *grpc.ClientConn, error)`. Callers must `defer conn.Close()`.
- When adding a new service version, create a `client_builder.go` in the version package with a one-liner type alias and `NewClientBuilder` function.

### RBAC REST API (kessel/rbac/v2)
- The RBAC workspace API uses plain HTTP (`net/http`), not gRPC. Endpoint: `GET /api/rbac/v2/workspaces/?type={root|default}`.
- Auth is pluggable via the `auth.AuthRequest` interface (method `ConfigureRequest(ctx, *http.Request) error`).
- The `x-rh-rbac-org-id` header is required on workspace requests.
- Utility functions (`WorkspaceResource`, `PrincipalSubject`, `Subject`, etc.) construct `v1beta2` protobuf reference types for RBAC-specific resource/subject patterns. Always use these helpers instead of manually constructing `ResourceReference`/`SubjectReference` for RBAC resources.

### Streaming and Pagination
- `StreamedListObjects` and `StreamedListSubjects` use gRPC server-streaming.
- Pagination uses `RequestPagination` (with `limit` and optional `continuation_token`) and `ResponsePagination` (with `continuation_token`).
- The `ListWorkspaces` function in `kessel/rbac/v2` demonstrates the canonical pagination loop pattern: it returns `iter.Seq2` and handles continuation tokens internally.

