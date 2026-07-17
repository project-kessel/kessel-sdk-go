# GUIDELINES.md -- kessel/rbac/v2

## Package Purpose

This package provides two distinct integration surfaces for RBAC:
1. **REST workspace client** (`workspace.go`) -- plain HTTP calls to `/api/rbac/v2/workspaces/`, no gRPC.
2. **gRPC iterator + utility constructors** (`list_workspaces.go`, `utils.go`) -- wraps `v1beta2.StreamedListObjects` with pagination and provides convenience builders for RBAC-specific protobuf references.

These two surfaces share no transport code. REST functions use `net/http`; the iterator uses the `v1beta2.KesselInventoryServiceClient` gRPC client.

## REST Workspace Functions

### Standalone Functions, Not a Client

`FetchRootWorkspace` and `FetchDefaultWorkspace` are package-level functions, not methods on a struct. Pass the RBAC base endpoint, org ID, and options each time.

### Required Header

Every REST workspace request must carry `x-rh-rbac-org-id`. The SDK sets this from the `orgId` parameter -- never set it manually on the request.

### FetchWorkspaceOptions

- `HttpClient` -- optional; defaults to `http.DefaultClient` when nil.
- `Auth` -- an `auth.AuthRequest` (interface with `ConfigureRequest(ctx, *http.Request) error`). When nil, no auth header is set.

### Endpoint Normalization

The base endpoint is trimmed of trailing slashes via `strings.TrimRight` before appending the path constant `/api/rbac/v2/workspaces/`. Tests cover single, multiple, and zero trailing slashes.

### Response Contract

The REST endpoint returns `{"data": [...]}`. The SDK expects exactly one workspace in `data` for `FetchRootWorkspace`/`FetchDefaultWorkspace`. Zero or more than one results in an error.

## ListWorkspaces Iterator (gRPC)

### Return Type: iter.Seq2

`ListWorkspaces` returns `iter.Seq2[*v1beta2.StreamedListObjectsResponse, error]`. Consume it with Go range-over-func:

```go
for resp, err := range v2.ListWorkspaces(ctx, client, subject, "viewer", "") {
```

### Pagination Is Internal

The iterator loops over continuation tokens automatically. Subsequent pages (those with a continuation token) request limit 1000. The initial page has no explicit limit. When the last response has an empty continuation token, iteration stops. Pass a non-empty `continuationToken` to resume from a prior position.

### Functional Options

Use `WithConsistency(c)` to attach a `*v1beta2.Consistency` to every request in the pagination loop. Extend options by adding new `ListWorkspacesOption` functions following the same closure pattern.

### Early Termination

If the caller breaks out of the range loop, the iterator returns immediately (`yield` returns false). No cleanup is needed.

## Utility Constructors (utils.go)

### All RBAC Resources Use ReporterType "rbac"

Every constructor (`PrincipalResource`, `WorkspaceResource`, `RoleResource`, `WorkspaceType`, `RoleType`) hardcodes `ReporterType: "rbac"`. Do not construct RBAC resource references manually.

### PrincipalResource ID Format

`PrincipalResource(id, domain)` produces `ResourceId: "domain/id"` (e.g., `"redhat/alice"`). The domain comes first.

### Subject Relation Semantics

`Subject(ref, relation)` sets the `Relation` field only when `relation != ""`. For direct subjects (like principals), use `PrincipalSubject` which omits the relation. This distinction matters for authorization checks.

## Testing Conventions

Repo-wide testing rules (white-box packaging, `tt` loop variable, table-driven structure) are in [AGENTS.md -- Testing Conventions](../../../AGENTS.md#testing-conventions). This section covers v2-specific patterns only.

### Assertion Library

This package uses `testify/assert` and `testify/require`. Do not cross-pollinate testify into `auth`, `config`, or `grpc` packages.

### REST Tests: httptest

HTTP tests use `httptest.NewServer` with per-case handler functions. The test table includes a `serverHandler` field and optionally a `validateReq` function to assert request properties (method, path, query params, headers). Always guard: `if tt.validateReq != nil`.

### gRPC Tests: Embedded Interface Mocks

Mock the `v1beta2.KesselInventoryServiceClient` interface by embedding it in a struct and overriding `StreamedListObjects`. Mock streams implement `Recv()` returning canned responses then `io.EOF`. Capture requests in a `capturedRequests` slice for assertion.

## Import Aliases

```go
v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
```

Always alias versioned inventory imports by their version string.

## Adding New Functionality

- New REST endpoints: follow the `fetchWorkspace` pattern -- unexported implementation function, exported wrappers with specific defaults, standalone (not methods).
- New gRPC iterators: return `iter.Seq2[*ResponseType, error]`, handle pagination internally, support functional options.
- New resource types: add a `XxxType() *RepresentationType` and `XxxResource(id) *ResourceReference` to `utils.go`, always with `ReporterType: "rbac"`.
- New subject constructors: follow `PrincipalSubject` for direct subjects (no relation) or `Subject` for relation-bearing subjects.
