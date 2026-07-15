# kessel/auth -- Directory Guidelines

## Package Purpose

This package implements OAuth2 client-credentials authentication for the Kessel SDK. It provides token acquisition, thread-safe caching, OIDC discovery, and the `AuthRequest` interface used by both the HTTP (RBAC workspace) and gRPC (inventory) layers.

## Key Files

| File | Contains |
|------|----------|
| `auth.go` | `OAuth2ClientCredentials` struct, `GetToken`, `FetchOIDCDiscovery`, token caching logic |
| `auth_request.go` | `AuthRequest` interface, `OAuth2AuthRequest` constructor, `oauth2Auth` implementation |
| `auth_test.go` | Tests for credentials, token lifecycle, OIDC discovery, concurrent access |
| `auth_request_test.go` | Tests for `AuthRequest` construction, `ConfigureRequest`, caching through the interface |

## Construction Rules

- `OAuth2ClientCredentials` fields (`clientId`, `clientSecret`, `tokenEndpoint`) are **unexported**. Always construct via `NewOAuth2ClientCredentials(clientId, clientSecret, tokenEndpoint)`. Struct literals will not compile outside this package.
- `NewOAuth2ClientCredentials` returns a **value**, not a pointer. The caller must take its address (`&creds`) before passing it to any consumer. All downstream consumers (`OAuth2AuthRequest`, `OAuth2CallCredentials`, `OAuth2ClientAuthenticated`) accept `*OAuth2ClientCredentials`.
- `oauth2Auth` is unexported. Callers obtain an `AuthRequest` via `OAuth2AuthRequest(creds, options)` -- they never see the concrete type.

## Thread-Safe Token Caching -- The Generation Counter Pattern

`GetToken` uses a three-phase concurrency protocol. Understanding this is critical before modifying any caching logic:

1. **Pre-lock snapshot:** `atomic.LoadUint64(&o.generation)` is read *before* any lock acquisition. All goroutines that observe the token as stale record the same generation value.
2. **Fast path (RLock):** If `!ForceRefresh && isTokenValid()`, return the cached token immediately. Multiple readers proceed concurrently.
3. **Slow path (Lock + double-check):** After acquiring the write lock, compare the current generation to the pre-lock snapshot. If they differ and the token is valid, another goroutine already refreshed -- return the cached token. Otherwise, call `refreshToken` and increment the generation via `atomic.AddUint64`.

**Why this matters:** The generation counter prevents thundering-herd token refreshes. Concurrent `ForceRefresh: true` callers also coalesce -- the post-lock recheck intentionally does not re-check `ForceRefresh`. Three dedicated tests (`TestConcurrentTokenAccess`, `TestConcurrentTokenAccess_stale_token`, `TestConcurrentForceRefresh`) enforce exactly-1-SSO-call semantics with 20 goroutines. Run with `-race` after any change to this code.

## Token Validity

- `isTokenValid()` returns false when the token is empty OR within `expirationWindow` (300 seconds / 5 minutes) of expiry.
- If the token response omits `expires_in`, `refreshToken` defaults to `defaultExpiresIn` (3600 seconds). Do not assume IdPs always return this field.
- `ExpiresAt` is computed as `time.Now().Add(duration)` at refresh time -- `time.Now()` includes a monotonic reading, which `Add` preserves for in-process comparisons. Serialization or reconstruction from wall-clock components strips the monotonic reading. The 5-minute buffer makes this acceptable in practice.

## AuthRequest Interface Contract

```go
type AuthRequest interface {
    ConfigureRequest(ctx context.Context, request *http.Request) error
}
```

- Used by `kessel/rbac/v2/workspace.go` to attach the `authorization: Bearer <token>` header to HTTP requests.
- `ConfigureRequest` calls `GetToken` internally. Callers do not manage tokens directly.
- The header key is lowercase `"authorization"` (Go's `http.Header.Set` canonicalizes it, but the string literal is lowercase in the source).
- If the `Auth` field is nil in consumer options, no auth header is sent -- the consumer skips calling `ConfigureRequest` entirely.

## Error Handling

- **No wrapping.** Errors from `client.Discover`, `client.CallTokenEndpoint`, and `GetToken` are returned as-is. Do not add `fmt.Errorf("...: %w", err)` wrapping in this package.
- On error, return the zero-value struct: `RefreshTokenResponse{}` (not a nil pointer -- it is a value type).

## HTTP Client Convention

Every function accepting HTTP options follows this exact nil-fallback pattern:
```go
httpClient := options.HttpClient
if httpClient == nil {
    httpClient = http.DefaultClient
}
```
Do not create new `http.Client` instances inside this package. The caller controls timeouts and TLS.

## OIDC Discovery

- `FetchOIDCDiscovery` delegates to `zitadel/oidc/v3`'s `client.Discover`. Do not reimplement OIDC discovery.
- Returns only `TokenEndpoint` from the discovery document (via `OIDCDiscoveryMetadata`). Other fields are not exposed.
- The issuer URL should come from the `AUTH_DISCOVERY_ISSUER_URL` environment variable (loaded at call time, not import time).

## Downstream Consumers

The auth package has exactly three consumers -- changes here affect all of them:

| Consumer | How it uses auth |
|----------|-----------------|
| `kessel/grpc/grpc.go` | Wraps `*OAuth2ClientCredentials` in `OAuth2CallCredentials` (gRPC `PerRPCCredentials`). Always requires TLS. |
| `kessel/inventory/internal/builder/builder.go` | Internal `oauth2PerRPCCreds` adapter. `RequireTransportSecurity` returns `!insecure`. |
| `kessel/rbac/v2/workspace.go` | Calls `AuthRequest.ConfigureRequest` to set HTTP auth headers. |

Do not add new exported types without considering the impact on all three consumers.

## Testing Conventions

- **stdlib only.** This package uses `testing` from the standard library. Do not introduce testify (`assert`/`require`). That is reserved for `kessel/rbac/v2`.
- **White-box tests.** Test files use `package auth` (no `_test` suffix). Tests access unexported fields directly (e.g., `credentials.cachedToken = ...`).
- **Table-driven with `tt`.** Loop variable is always `tt`. Subtest names are lowercase with spaces.
- **Token endpoint mocking.** Use `httptest.NewServer` returning JSON `{"access_token": "...", "token_type": "Bearer", "expires_in": N}`. Never hit a real IdP.
- **Concurrency tests.** Use barrier synchronization (WaitGroup + gate channel) with 20 goroutines. Assert exactly 1 SSO call. Always run with `-race`.
- **Context cancellation.** Test `ConfigureRequest` with a short-timeout context to verify error propagation.

## Non-Obvious Gotchas

1. **Mutex in value type.** `OAuth2ClientCredentials` contains `sync.RWMutex`. Copying by value silently copies the mutex. Always pass by pointer after construction.
2. **Generation counter is not under mutex.** It uses `atomic` operations deliberately. Do not move it under the mutex or replace with a non-atomic read.
3. **`expirationWindow` is in seconds, not `time.Duration`.** It is multiplied by `time.Second` at usage sites. Do not change it to a `time.Duration` constant.
4. **`NewOAuth2ClientCredentials` returns a value.** This is unlike most Go constructors. The caller must `&` it before passing downstream.
