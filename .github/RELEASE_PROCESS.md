# Release Process

This document describes how to release a new version of jirasdk.

## Automated Release Workflow

The repository includes automated GitHub Actions workflows for releasing new versions.

### Prerequisites

- All tests must pass on `main` branch
- CHANGELOG.md should be updated with changes for the new version
- You must have write access to the repository

### Release Steps

#### 1. Update CHANGELOG.md

Add a new section for your release:

```markdown
## [v1.0.0] - 2025-01-08

### Added
- New feature X
- New feature Y

### Changed
- Improved performance of Z

### Fixed
- Bug fix for issue #123

### Breaking Changes
- Changed API signature for method Foo
```

Commit and push:
```bash
git add CHANGELOG.md
git commit -m "docs: Update CHANGELOG for v1.0.0"
git push
```

#### 2. Create Release Tag (Via GitHub UI)

1. Go to: https://github.com/felixgeelhaar/jirasdk/actions/workflows/tag.yml
2. Click "Run workflow"
3. Fill in the form:
   - **Version**: Enter version (e.g., `v1.0.0`, `v1.0.1-rc.1`)
   - **Prerelease**: Check if this is a prerelease (alpha, beta, rc)
4. Click "Run workflow"

The workflow will:
- ✅ Validate version format
- ✅ Check if tag already exists
- ✅ Run full test suite
- ✅ Create and push the tag
- ✅ Trigger the release workflow automatically

#### 3. Automatic Release Creation

Once the tag is pushed, the release workflow automatically:

1. **Verifies the tag is on `main`**: every later job is skipped if it is not,
   so a tag pushed from a branch produces no release
2. **Runs Tests**: full test suite on the Go versions in
   `.github/workflows/release.yml`
3. **Creates Release**: a GitHub Release whose body is this version's section
   extracted from CHANGELOG.md, marked as a prerelease when the tag contains
   `-rc`, `-beta` or `-alpha`
4. **Updates pkg.go.dev**: requests the new version from the Go module proxy,
   which is what makes it appear on pkg.go.dev

**No binaries are built, and the release carries no attached assets.** jirasdk
is a library, so there is nothing to compile for a user to download — they
consume it with `go get`. A release with zero assets is correct and is not a
sign that something failed.

#### 4. Verify Release

The release itself is only half of it. What actually determines whether people
can use the new version is the Go module proxy, so check that too:

1. Check the release page: https://github.com/felixgeelhaar/jirasdk/releases —
   the body should be this version's changelog section. There will be no
   attached assets; see above.
2. Confirm the module is fetchable, which is the real test of a library
   release:

   ```bash
   cd "$(mktemp -d)" && go mod init verify
   go get github.com/felixgeelhaar/jirasdk@vX.Y.Z
   ```

   A version that resolves here is live for every consumer. Note that the proxy
   caches immutably: a published version can never be changed or withdrawn, so
   a mistake is fixed by releasing another version, never by retagging.
3. Check pkg.go.dev: https://pkg.go.dev/github.com/felixgeelhaar/jirasdk
   Documentation can lag the proxy by a few minutes.

## Manual Release (Alternative)

If you prefer manual releases:

### 1. Create Tag Locally

```bash
# Ensure you're on main and up to date
git checkout main
git pull

# Create annotated tag
git tag -a v1.0.0 -m "Release v1.0.0

- Feature A
- Feature B
- Bug fix C
"

# Push tag
git push origin v1.0.0
```

### 2. Wait for Automated Release

The release workflow will automatically trigger when the tag is pushed.

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version (v2.0.0): Incompatible API changes
- **MINOR** version (v1.1.0): New functionality, backwards compatible
- **PATCH** version (v1.0.1): Backwards compatible bug fixes

### Prerelease Versions

- **Alpha** (v1.0.0-alpha.1): Early testing, unstable
- **Beta** (v1.0.0-beta.1): Feature complete, testing phase
- **RC** (v1.0.0-rc.1): Release candidate, final testing

## Release Checklist

Before creating a release:

- [ ] All tests pass on `main`
- [ ] CHANGELOG.md is updated
- [ ] Breaking changes are documented
- [ ] Examples are updated (if needed)
- [ ] README is up to date
- [ ] Migration guide added (for breaking changes)
- [ ] Security issues addressed
- [ ] Dependencies updated

## Post-Release Tasks

After release is published:

1. **Verify pkg.go.dev**: Check documentation appears correctly
2. **Test Installation**: Verify users can install with `go get`
   ```bash
   go get github.com/felixgeelhaar/jirasdk@v1.0.0
   ```
3. **Monitor Issues**: Watch for bug reports from new version
4. **Update Examples**: Ensure all examples work with new version

## Hotfix Process

For urgent bug fixes:

1. Create hotfix branch from tag:
   ```bash
   git checkout -b hotfix/v1.0.1 v1.0.0
   ```

2. Apply fixes and test:
   ```bash
   # Make changes
   git add .
   git commit -m "fix: Critical bug in feature X"
   ```

3. Merge to main:
   ```bash
   git checkout main
   git merge --no-ff hotfix/v1.0.1
   git push
   ```

4. Create patch release:
   - Use "Tag Release" workflow with v1.0.1

## Rollback

If a release has critical issues:

1. **Don't delete the tag** (breaks pkg.go.dev caching)
2. Release a new patch version with fixes:
   - v1.0.1 fixes issues in v1.0.0
3. Document the issue in CHANGELOG
4. Add migration notes if needed

## Troubleshooting

### Release workflow failed

1. Check workflow logs in Actions tab
2. Common issues:
   - Tests failing
   - Build errors
   - Permission issues
3. Fix issues and re-run workflow

### Tag already exists

```bash
# Delete local tag
git tag -d v1.0.0

# Delete remote tag (use carefully!)
git push origin :refs/tags/v1.0.0

# Create new tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### pkg.go.dev not updating

1. Wait 15-30 minutes (can take time)
2. Manually trigger:
   ```bash
   curl "https://proxy.golang.org/github.com/felixgeelhaar/jirasdk/@v/v1.0.0.info"
   ```
3. Check https://pkg.go.dev/github.com/felixgeelhaar/jirasdk

## Getting Help

- **Workflow Issues**: Check [GitHub Actions Documentation](https://docs.github.com/en/actions)
- **Versioning Questions**: See [Semantic Versioning](https://semver.org/)
- **Go Module Issues**: See [Go Modules Reference](https://go.dev/ref/mod)

---

**Last Updated**: 2025-01-08
