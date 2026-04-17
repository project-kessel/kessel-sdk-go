# Integration Guidelines

Rules and conventions for integrating with Kessel services using this Go SDK. Intended for AI agents implementing or reviewing integration code.

## Architecture Overview

The SDK provides two transport paths to Kessel services:
- **gRPC** (Inventory API via `v1beta2`): Check, CheckSelf, CheckForUpdate, CheckBulk, CheckSelfBulk, CheckForUpdateBulk, ReportResource, DeleteResource, StreamedListObjects, StreamedListSubjects
- **REST/HTTP** (RBAC API via `kessel/rbac/v2`): Workspace lookup endpoints (`/api/rbac/v2/workspaces/`)

Authentication is handled by a shared `kessel/auth` package used by both transports.

## gRPC Client Construction

### Use the ClientBuilder, Not Raw grpc.NewClient

Always construct inventory clients via the builder pattern in `kessel/inventory/v1beta2`. Do not call `grpc.NewClient` directly.

```go
client, conn, err := v1beta2.NewClientBuilder(endpoint).
    Insecure().     // or .OAuth2ClientAuthenticated(...) or .Authenticated(...)
    Build()
```

`Build()` returns three values: the typed client, the `*grpc.ClientConn`, and an error. The caller owns the connection and must defer `conn.Close()`.

### Builder Authentication Modes

The builder exposes exactly four mutually exclusive modes. Call exactly one before `Build()`:

| Method | Use Case |
|---|---|
| `.Insecure()` | Local dev; no TLS, no auth. Clears any previously set credentials. |
| `.OAuth2ClientAuthenticated(creds, channelCreds)` | Pass `*auth.OAuth2ClientCredentials` directly. SDK handles per-RPC token injection. Pass `nil` for `channelCreds` to use default TLS. |
| `.Authenticated(perRPCCreds, channelCreds)` | Bring your own `credentials.PerRPCCredentials` (e.g., `kesselgrpc.OAuth2CallCredentials`). Pass `nil` for `channelCreds` to use default TLS. |
| `.Unauthenticated(channelCreds)` | TLS but no auth token. Pass `nil` for default TLS. |

When `channelCredentials` is `nil`, the builder defaults to `credentials.NewTLS(&tls.Config{})` (system root CA pool). Never pass `insecure.NewCredentials()` as `channelCreds` -- use `.Insecure()` instead.

### Connection Lifecycle

Always defer connection close immediately after `Build()`:

```go
defer func() {
    if closeErr := conn.Close(); closeErr != nil {
        log.Printf("Failed to close gRPC client: %v", closeErr)
    }
}()
```

## OIDC Discovery and Token Management

### Two-Step Flow: Discover Then Construct Credentials

OIDC integration always follows this order:

1. Discover the token endpoint from the issuer URL
2. Construct `OAuth2ClientCredentials` with the discovered endpoint
3. Pass credentials to the builder or REST auth adapter

```go
discovered, err := auth.FetchOIDCDiscovery(ctx, issuerURL, auth.FetchOIDCDiscoveryOptions{})
creds := auth.NewOAuth2ClientCredentials(clientID, clientSecret, discovered.TokenEndpoint)
```

### Token Caching Is Built In

`OAuth2ClientCredentials` caches tokens internally with mutex protection. Tokens are refreshed automatically when they expire (with a 5-minute buffer window). Do not implement your own caching layer around `GetToken`.

### Custom HTTP Client Support

Both `FetchOIDCDiscovery` and `GetToken` accept an optional `HttpClient`. When `nil`, they use `http.DefaultClient`. Pass a custom client only when you need custom TLS, proxies, or timeouts for the token endpoint calls.

## REST API Integration (RBAC Workspaces)

### Workspace Functions Are Standalone, Not Methods on a Client

The RBAC workspace API uses plain functions, not a client object:

```go
ws, err := v2.FetchDefaultWorkspace(ctx, rbacBaseEndpoint, orgId, v2.FetchWorkspaceOptions{...})
ws, err := v2.FetchRootWorkspace(ctx, rbacBaseEndpoint, orgId, v2.FetchWorkspaceOptions{...})
```

### Required Header: x-rh-rbac-org-id

The SDK automatically sets the `x-rh-rbac-org-id` header from the `orgId` parameter. Do not set this header manually.

### Auth for REST Calls Uses the AuthRequest Interface

To authenticate REST calls, create an `auth.AuthRequest` via `auth.OAuth2AuthRequest` and pass it in options:

```go
authReq := auth.OAuth2AuthRequest(&creds, auth.OAuth2AuthRequestOptions{HttpClient: httpClient})
v2.FetchDefaultWorkspace(ctx, endpoint, orgId, v2.FetchWorkspaceOptions{
    HttpClient: httpClient,
    Auth:       authReq,
})
```

The `AuthRequest` interface has a single method `ConfigureRequest(ctx, *http.Request) error` that sets the Authorization header. When `Auth` is `nil` in options, no authentication is applied.

### Endpoint URL Handling

The SDK trims trailing slashes from `rbacBaseEndpoint` before appending `/api/rbac/v2/workspaces/`. Both `http://host:port` and `http://host:port/` work correctly.

## RBAC Utility Functions (kessel/rbac/v2)

### Resource and Subject Constructors

Use the helper functions in `kessel/rbac/v2/utils.go` to build protobuf references. All RBAC-related resources use `ReporterType: "rbac"`:

| Function | Returns | ResourceId Format |
|---|---|---|
| `PrincipalResource(id, domain)` | `*ResourceReference` | `domain/id` (e.g., `redhat/alice`) |
| `PrincipalSubject(id, domain)` | `*SubjectReference` | `domain/id`, no relation |
| `WorkspaceResource(resourceId)` | `*ResourceReference` | as-is |
| `RoleResource(resourceId)` | `*ResourceReference` | as-is |
| `WorkspaceType()` | `*RepresentationType` | N/A |
| `RoleType()` | `*RepresentationType` | N/A |
| `Subject(ref, relation)` | `*SubjectReference` | from ref; relation set only if non-empty |

### ListWorkspaces Returns an Iterator

`ListWorkspaces` returns `iter.Seq2[*StreamedListObjectsResponse, error]` and handles pagination internally (continuation pages use limit 1000, follows continuation tokens). Use Go range-over-func:

```go
for resp, err := range v2.ListWorkspaces(ctx, inventoryClient, subject, relation, "") {
    if err != nil { break }
    // process resp
}
```

Pass an initial `continuationToken` (last arg) to resume from a previous position; pass `""` to start from the beginning.

## Two Separate gRPC Credential Adapters

The SDK has two distinct ways to adapt OAuth2 credentials for gRPC. Do not confuse them:

1. **`kesselgrpc.OAuth2CallCredentials`** (in `kessel/grpc/grpc.go`): Wraps `*OAuth2ClientCredentials` as `credentials.PerRPCCredentials`. Always requires transport security (`RequireTransportSecurity() == true`). Use with `.Authenticated()`.

2. **Internal `oauth2PerRPCCreds`** (in `kessel/inventory/internal/builder/builder.go`): Used internally by `.OAuth2ClientAuthenticated()`. Respects the builder's insecure flag. Not exported.

If using `.Authenticated()`, you must use adapter #1 explicitly. If using `.OAuth2ClientAuthenticated()`, adapter #2 is applied automatically.

## CompatibilityConfig (Legacy Pattern)

`kessel/config/config.go` provides `CompatibilityConfig` with functional options (`WithGRPCEndpoint`, `WithGRPCInsecure`, `WithGRPCTLSConfig`, etc.). This config struct is **not consumed by the ClientBuilder**. It exists for manual `grpc.NewClient` usage as shown in the README. Prefer the ClientBuilder for new code.

## Environment Variables in Examples

When writing example binaries under `examples/`, read configuration from these env vars (see `.env.sample`):

| Variable | Purpose |
|---|---|
| `KESSEL_ENDPOINT` | gRPC server address (e.g., `localhost:9000`) |
| `AUTH_CLIENT_ID` | OAuth2 client ID |
| `AUTH_CLIENT_SECRET` | OAuth2 client secret |
| `AUTH_DISCOVERY_ISSUER_URL` | OIDC issuer URL for discovery |

Use `github.com/joho/godotenv/autoload` as the import for automatic `.env` file loading, matching the pattern in all existing examples.

## Generated Code

All `*.pb.go` files under `kessel/inventory/` are generated by buf (`make generate`). Never edit these files manually. Only `client_builder.go` and files under `kessel/rbac/`, `kessel/auth/`, `kessel/grpc/`, and `kessel/config/` are hand-written.
