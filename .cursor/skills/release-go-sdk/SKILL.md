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

### Step 0: Preflight -- Clean Working Tree

Run `git status --porcelain` to check for uncommitted changes. If the working tree is dirty, present the list of changed files and ask the user whether to:
1. Abort the release (recommended if unsure)
2. Stash changes for later: `git stash --include-untracked`

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

### Step 4: Review Changes

Before committing, summarize the release for the user and ask for confirmation.

1. Run `git diff --stat` and `git status` to gather all pending changes.
2. Compare `$VERSION` against the latest git tag (`git describe --tags --abbrev=0`) to determine the bump type (major/minor/patch).
3. Present a summary to the user including:
   - The version being released and the bump type
   - List of files that will be committed
   - Quality check results
4. **Wait for user confirmation before proceeding.**

### Step 5: Commit, Push Branch, and Create PR

```bash
git checkout -b release/${VERSION}
git add .
git commit -m "chore: regenerate protobuf code"
git push -u origin release/${VERSION}
gh pr create --title "Release v${VERSION}" --body "Release version ${VERSION}"
```

Include any other changed files (generated code, lock files) in the commit.

**The remaining steps (tag, GitHub release) should be performed after the PR is merged to main.**

### Step 6: Tag the Release

After the PR is merged, switch back to main and pull:

```bash
git checkout main && git pull origin main
git tag -a v${VERSION} -m "Release version ${VERSION}"
git push origin v${VERSION}
```

### Step 7: Create GitHub Release

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
- [ ] Preflight: clean working tree
- [ ] Check existing tags and determine new version
- [ ] Set VERSION env var
- [ ] Regenerate gRPC code if needed (make generate)
- [ ] Run make lint, make test, make build
- [ ] Review changes and get user confirmation
- [ ] Commit, push release/${VERSION} branch, create PR
- [ ] Merge PR to main
- [ ] Create and push git tag (v${VERSION})
- [ ] Create GitHub release
```
