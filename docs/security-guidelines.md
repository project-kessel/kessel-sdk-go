# Security Guidelines for kessel-sdk-go

These guidelines document the security conventions, patterns, and rules that apply to this repository. They are intended for AI agents performing implementation and review tasks.

## Authentication Architecture

### OAuth2 Client Credentials Flow

All OAuth2 authentication uses the `kessel/auth` package backed by `github.com/zitadel/oidc/v3`. The SDK supports exactly one OAuth2 grant type: `client_credentials`.

**Rules:**

- Always use `auth.NewOAuth2ClientCredentials(clientId, clientSecret, tokenEndpoint)` to create credential objects. Never construct `OAuth2ClientCredentials` structs directly -- fields are unexported by design.
- Obtain the token endpoint via OIDC discovery (`auth.FetchOIDCDiscovery`) rather than hardcoding it. The issuer URL comes from the environment variable `AUTH_DISCOVERY_ISSUER_URL`.
- Never read `AUTH_CLIENT_ID` or `AUTH_CLIENT_SECRET` at import time. Load them at call time via `os.Getenv` or the `godotenv/autoload` pattern used in examples.
- The `.env` file is in `.gitignore`. Never commit credentials. Use `.env.sample` (which contains placeholder values) as the template.

### Token Caching and Refresh

- `OAuth2ClientCredentials.GetToken` caches tokens in memory with a `sync.Mutex` for thread safety.
- Tokens are considered invalid 5 minutes (`expirationWindow = 300` seconds) before their actual expiry. Do not change this buffer without understanding the downstream impact on concurrent callers.
- If the token response omits `expires_in`, the SDK defaults to 3600 seconds (`defaultExpiresIn`). Do not assume the IdP always returns this field.
- Use `GetTokenOptions.ForceRefresh = true` only when the caller has evidence the token was rejected (e.g., a 401 response). Do not force-refresh on every call.
- Tests for concurrent token access exist (`TestConcurrentTokenAccess`). Any change to the caching or mutex logic must pass that test.

## Transport Security (TLS and gRPC)

### ClientBuilder Security Modes

The `kessel/inventory/internal/builder` package defines four mutually exclusive security modes via the `ClientBuilder` fluent API. Use exactly one per client:

| Method | Transport | Per-RPC Auth | Use Case |
|--------|-----------|--------------|----------|
| `OAuth2ClientAuthenticated(creds, tlsCreds)` | TLS (default or custom) | OAuth2 token via internal adapter | Production with SDK-managed auth |
| `Authenticated(perRPC, tlsCreds)` | TLS (default or custom) | Caller-provided `PerRPCCredentials` | Production with custom auth |
| `Unauthenticated(tlsCreds)` | TLS (default or custom) | None | Health checks, public endpoints |
| `Insecure()` | Plaintext | None | Local development only |

**Rules:**

- `NewClientBuilder` defaults to TLS with an empty `tls.Config{}` (system CA pool). This is the secure default -- do not change it.
- `WithGRPCTLSConfig` forces `Insecure = false`. This override ordering is intentional and tested (`TestOptionsOrder`).
- `Insecure()` clears both channel credentials and per-RPC credentials. It must never be used in production code paths.
- The `oauth2PerRPCCreds` adapter in the builder sets `RequireTransportSecurity()` to `!insecure`. The standalone `grpc.OAuth2CallCredentials` always returns `true`. Keep these consistent with their intended use contexts.
- Per-RPC credentials are applied via `grpc.WithDefaultCallOptions(grpc.PerRPCCredentials(...))`, not `grpc.WithPerRPCCredentials(...)`. This is deliberate -- it attaches creds to every call on the connection.

### TLS Configuration

- When providing a custom `*tls.Config`, always set `MinVersion: tls.VersionTLS12` (see test pattern in `config_test.go`).
- The `CompatibilityConfig.TLSConfig` field is tagged `json:"-"` so it is never serialized. Do not change this tag.
- Never set `InsecureSkipVerify: true` in any non-test code.

## HTTP Client Handling

- Every function that makes HTTP calls accepts an optional `*http.Client`. If nil, it falls back to `http.DefaultClient`. Follow this exact pattern:
```go
httpClient := options.HttpClient
if httpClient == nil {
    httpClient = http.DefaultClient
}
```
- Do not create new `http.Client` instances inside SDK functions. The caller controls timeouts, TLS, and transport settings via the injected client.

## RBAC and Authorization APIs

### HTTP-Based Workspace API

- The RBAC workspace client (`kessel/rbac/v2/workspace.go`) sends the organization ID via the `x-rh-rbac-org-id` header. This header is required for all RBAC API calls.
- Authentication is applied via the `auth.AuthRequest` interface's `ConfigureRequest` method, which sets the `authorization` header. If `Auth` is nil in options, no auth header is sent.
- The workspace endpoint is `/api/rbac/v2/workspaces/`. The trailing slash is significant. The base endpoint is trimmed of trailing slashes before concatenation.

### gRPC-Based Authorization Checks

- Use the helper functions in `kessel/rbac/v2/utils.go` (`WorkspaceResource`, `PrincipalSubject`, `PrincipalResource`, etc.) to build authorization request objects. These ensure the correct `reporter_type: "rbac"` is always set.
- Principal IDs follow the format `domain/id` (e.g., `redhat/alice`). This format is enforced by `PrincipalResource`.
- Handle gRPC errors by extracting status codes via `status.FromError(err)`. The examples demonstrate the standard switch pattern for `codes.Unavailable`, `codes.PermissionDenied`, and default.

## Environment Variables

| Variable | Purpose | Required |
|----------|---------|----------|
| `KESSEL_ENDPOINT` | gRPC server address (host:port) | Yes |
| `AUTH_CLIENT_ID` | OAuth2 client ID | For authenticated flows |
| `AUTH_CLIENT_SECRET` | OAuth2 client secret | For authenticated flows |
| `AUTH_DISCOVERY_ISSUER_URL` | OIDC issuer for discovery | For authenticated flows |

These are loaded via `os.Getenv` or `godotenv/autoload`. The `CompatibilityConfig` struct also supports `env` tags (`KESSEL_INSECURE`, `KESSEL_TIMEOUT`, etc.) but these are not used by the builder -- they exist for compatibility with external config loaders.

## Generated Code

- All `*.pb.go` and `*_grpc.pb.go` files are generated from `buf.build/project-kessel/inventory-api` via `buf generate`. Never hand-edit these files.
- The generated protobuf types import `buf.build/gen/go/bufbuild/protovalidate` for field validation annotations. Validation is defined in the proto source and enforced server-side; the SDK does not run client-side validation.

## Testing Security Code

- Use `httptest.NewServer` for token endpoint mocking. Tests must never make real HTTP requests to identity providers.
- Auth tests must cover: valid cached tokens, expired tokens, force refresh, server errors, context cancellation, and concurrent access.
- Mock `auth.AuthRequest` via a struct implementing `ConfigureRequest(ctx, *http.Request) error` (see `mockAuthRequest` in `workspace_test.go`).
- The `grpc_test.go` tests verify that `RequireTransportSecurity()` returns `true` for `OAuth2CallCredentials`. Any new credential type must have an equivalent test.

## Dependency Security Notes

- `github.com/zitadel/oidc/v3` handles OIDC discovery and token endpoint calls. Do not reimplement OIDC discovery manually.
- `github.com/go-jose/go-jose/v4` (transitive via zitadel) handles JWT/JWS operations. Do not add separate JWT libraries.
- `google.golang.org/grpc` provides transport security primitives. Use `credentials.NewTLS` and `insecure.NewCredentials` from this package -- do not use raw `net/tls` for gRPC connections.
- Dependabot is configured (`.github/dependabot.yml`) for automated dependency updates.

## Connection Lifecycle

- Always `defer conn.Close()` after a successful `Build()` call. The examples show the idiomatic pattern with error logging on close failure:
```go
defer func() {
    if closeErr := conn.Close(); closeErr != nil {
        log.Printf("Failed to close gRPC client: %v", closeErr)
    }
}()
```
- `Build()` returns an error if `target` is empty. Always check this error before using the client or connection.
