@AGENTS.md

## Claude Code Build and Test Commands

Before committing changes, run:
```bash
make lint        # golangci-lint via Docker/Podman (examples linted individually)
make test        # go test -v ./kessel/...
```

Optional quality commands:
```bash
make test-coverage   # generates coverage.out + coverage.html
make fmt             # go fmt ./...
make mod-tidy        # go mod tidy
make build           # compiles example binaries into bin/
```

## CI Expectations

Both `golangci-lint` and `CI Build and Test` workflows must pass before merge. CI runs on every push/PR to `main`.

## Pre-Commit Hook

The repository has a `rh-multi-pre-commit` hook installed (Red Hat security scanning). If the hook fails, investigate and fix the underlying issue - never skip hooks with `--no-verify` unless explicitly requested.

## Generated Code

Never edit `*.pb.go` or `*_grpc.pb.go` files - they are regenerated every 6 hours by the `buf-generate.yml` workflow. Changes to protobuf definitions must go through the upstream `buf.build/project-kessel/inventory-api` repository.
