# GUIDELINES.md -- kessel/inventory/internal/builder/

Rules for working in the generic gRPC ClientBuilder package.

## What This Package Does

`builder.go` provides `ClientBuilder[C any]` -- a generic, fluent builder that constructs a typed gRPC client stub and returns it alongside the raw `*grpc.ClientConn`. `keepalive.go` provides the keepalive configuration and typed options. Every service version (currently only `v1beta2`) exposes its own thin wrapper via a type alias and `NewClientBuilder` function.

## Generic Pattern

`ClientBuilder[C any]` is parameterized on the client interface type. The `newStub` field holds the generated constructor (e.g., `v1beta2.NewKesselInventoryServiceClient`). This is the only use of generics in the repo.

When adding a new service version, create a `client_builder.go` in that version's package with exactly two things:
1. A type alias: `type ClientBuilder = genericBuilder.ClientBuilder[YourServiceClient]`
2. A constructor: `func NewClientBuilder(target string) *ClientBuilder { ... }`

Import alias the internal builder as `genericBuilder`. See `kessel/inventory/v1beta2/client_builder.go` for the canonical example.

## Four Mutually Exclusive Auth Modes

Each mode is a single method that fully configures both transport and per-RPC credentials. Call exactly one before `Build()`. Calling a second one silently overwrites the first.

| Method | Transport | Per-RPC | Notes |
|---|---|---|---|
| `Insecure()` | Plaintext | None | Clears any previously set credentials. Dev only. |
| `Unauthenticated(tlsCreds)` | TLS | None | Explicitly sets `perRPCCredentials = nil`. |
| `Authenticated(perRPC, tlsCreds)` | TLS | Caller-provided | Use with `kesselgrpc.OAuth2CallCredentials`. |
| `OAuth2ClientAuthenticated(creds, tlsCreds)` | TLS | Internal adapter | Wraps `*auth.OAuth2ClientCredentials` automatically. |

For the three TLS modes, passing `nil` as `channelCredentials` falls back to `credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})` (system CA pool, TLS 1.2 minimum). Do not pass `insecure.NewCredentials()` as the channel creds argument -- use `Insecure()` instead.

## Internal oauth2PerRPCCreds Adapter

This unexported type bridges `*auth.OAuth2ClientCredentials` to `credentials.PerRPCCredentials`. It differs from the exported `kesselgrpc.OAuth2CallCredentials` in one way: `RequireTransportSecurity()` returns `!o.insecure`, allowing it to work with `Insecure()` mode. The exported adapter always returns `true`.

Do not use both adapters for the same client. `.OAuth2ClientAuthenticated()` uses the internal one; `.Authenticated()` expects the caller to use `kesselgrpc.OAuth2CallCredentials` or a custom implementation.

## Three-Value Return from Build()

`Build()` returns `(C, *grpc.ClientConn, error)` -- not just a client. The zero value of `C` is returned on error via `var zero C`. Callers must always check the error and immediately defer a closure that closes the connection and logs any close error. The connection is caller-owned; the builder does not track or close it.

## Client Keepalive

Every client built by `ClientBuilder.Build()` has client keepalive enabled by default with a 45-second interval, a 10-second acknowledgment timeout, and `permitWithoutCalls = true`. Callers do not need to set an option to get these defaults.

The `v1beta2` package exposes the builder's typed, opaque `KeepaliveOption` API through `Keepalive` and the `WithKeepaliveInterval`, `WithKeepaliveTimeout`, and `WithKeepalivePermitWithoutCalls` functions. Applying options changes only the fields they specify; later partial `Keepalive` calls preserve other configured values. An explicit `false` for `permitWithoutCalls` is a valid override. Durations must be positive; `Build()` returns a validation error when an interval or timeout is zero or negative. grpc-go clamps keepalive intervals below 10 seconds to 10 seconds.

These pings are transport-level behavior, not an application health check, retry mechanism, or guarantee about server or load-balancer idle timeouts. Servers and gateways must allow the configured cadence, including pings when there are no active calls, or may close the connection with a `GOAWAY` carrying `too_many_pings`.

## No WithDialOptions Hook (By Design)

The builder deliberately omits a general `WithDialOptions` method. All dial options are assembled internally in `Build()`: transport credentials, optional per-RPC credentials, and the builder's supported keepalive configuration. Custom per-call options should be passed at the call site, not injected into the connection. Do not add a general `WithDialOptions` method without an explicit design decision to change this constraint.

## Per-RPC Credential Attachment

Credentials are attached via `grpc.WithDefaultCallOptions(grpc.PerRPCCredentials(...))`. While gRPC also provides `grpc.WithPerRPCCredentials(...)` as a dial option, the builder uses `WithDefaultCallOptions` to maintain consistency with how other default call options would be configured if added in the future. Both approaches attach credentials to every RPC on the connection.

## setChannelCredentialsOrDefault

This private helper is called by `OAuth2ClientAuthenticated`, `Authenticated`, and `Unauthenticated`. It:
1. Sets `b.insecure = false`
2. Uses the provided `channelCredentials` if non-nil, otherwise falls back to default TLS

`Insecure()` bypasses this helper entirely and sets its own state. The ordering matters: calling `Insecure()` then `OAuth2ClientAuthenticated()` results in TLS mode (last writer wins).

## Testing

Repo-wide testing rules (white-box packaging, `tt` loop variable, stdlib-only for infrastructure packages) are in [AGENTS.md -- Testing Conventions](../../../../AGENTS.md#testing-conventions).

`builder_test.go` contains standard-library unit tests for TLS defaults, auth-mode state, target validation, and successful connection creation/cleanup. Keep tests in the same package and use the standard library. For keepalive changes, cover the defaults, partial overrides across repeated calls, explicit `false`, and `Build()` validation errors for non-positive durations; verify observable builder behavior rather than merely repeating implementation constants.

## Dependencies

Only three external packages are imported:
- `crypto/tls` -- default TLS config construction
- `google.golang.org/grpc` + subpackages -- gRPC dial, credentials
- `kessel/auth` -- `OAuth2ClientCredentials` type (for the internal adapter)

Do not add dependencies on `kessel/config` (CompatibilityConfig) or `kessel/grpc` (exported adapter). Those are separate systems.

## Common Mistakes

- Passing `insecure.NewCredentials()` to a TLS auth mode instead of calling `Insecure()`.
- Adding `WithDialOptions` -- the closed dial option set is a deliberate design constraint.
- Exporting `oauth2PerRPCCreds` -- it must stay internal; the exported equivalent lives in `kessel/grpc`.
- Mixing `CompatibilityConfig` with `ClientBuilder` -- they do not interact.
- Treating keepalive pings as health checks, retries, or a load-balancer timeout guarantee.
- Assuming a server accepts client pings while idle or at every configured cadence; unsupported pings can trigger `GOAWAY` with `too_many_pings`.
- Adding testify assertions -- this package uses stdlib testing only.
