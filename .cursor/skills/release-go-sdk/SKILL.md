---
name: release-go-sdk
description: Release a new version of the Kessel Go SDK (kessel-sdk-go). Guides through version selection, code generation, quality checks, git tagging, and GitHub release creation. Use when the user wants to release, publish, bump version, or cut a new release of the Go SDK.
---

# Release Kessel Go SDK

## Prerequisites

- Write access to the GitHub repository
- Go 1.24+
- [buf](https://github.com/bufbuild/buf) for protobuf/gRPC code generation
- Docker or Podman (for linting)

## Release Process

### Step 1: Determine the Version

Check existing tags to find the current version:

```bash
git fetch --tags
git tag --sort=-v:refname | head -5
```

Or via GitHub:

```bash
gh release list --limit 5
```

Choose the new version following [Semantic Versioning](https://semver.org/):
- **MAJOR**: incompatible API changes
- **MINOR**: backward-compatible new functionality
- **PATCH**: backward-compatible bug fixes

Then set the `VERSION` env var for use in subsequent steps:

```bash
export VERSION=X.Y.Z
echo "Releasing version: v${VERSION}"
```

### Step 2: Update Dependencies (if needed)

Regenerate gRPC code if there are updates to the Kessel Inventory API:

```bash
make generate
```

### Step 3: Run Quality Checks

```bash
make lint
make test
make build
```

### Step 4: Commit and Push

Before committing, analyze the changes and present a summary to the user:

1. Run `git diff --stat` and `git status` to inspect what changed.
2. Summarize the changes for the user: how many files changed, the nature of the changes (e.g. regenerated protobuf code, updated dependencies, new API surface).
3. **Ask the user for confirmation before committing.** Do not proceed until confirmed.

Once confirmed, commit and push:

```bash
git add .
git commit -m "chore: regenerate protobuf code"
git push origin main
```

### Step 5: Tag the Release

```bash
git tag -a v${VERSION} -m "Release version ${VERSION}"
git push origin v${VERSION}
```

### Step 6: Create GitHub Release

```bash
gh release create v${VERSION} --title "v${VERSION}" --generate-notes
```

Or manually:

- Go to the [GitHub Releases page](https://github.com/project-kessel/kessel-sdk-go/releases)
- Click "Create a new release"
- Select the tag you just created
- Add release notes describing the changes
- Publish the release

Go modules are consumed directly from GitHub via `go get` -- no separate package registry publish step is needed. Once the tag is pushed, users can install the new version with:

```bash
go get github.com/project-kessel/kessel-sdk-go@v${VERSION}
```

## Quick Reference Checklist

```
Release v${VERSION}:
- [ ] Check existing tags and determine new version
- [ ] Set VERSION env var
- [ ] Regenerate gRPC code if needed (make generate)
- [ ] Run make lint, make test, make build
- [ ] Commit and push any changes
- [ ] Create and push git tag (v${VERSION})
- [ ] Create GitHub release
```
