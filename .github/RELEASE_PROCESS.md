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
## [v1.9.1] - YYYY-MM-DD

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

**The `v` prefix is load-bearing.** The release workflow extracts the release
notes by matching `^## [<tag>]`, and tags carry a `v`, so a heading written
`## [1.9.1]` does not match its own tag. The release is still created, but its
body silently falls back to "See CHANGELOG.md for details" — which is how some
older versions in this repo ended up with no notes.

Open a pull request with the change; `main` requires one:

```bash
git checkout -b docs/changelog-v1.9.1
git add CHANGELOG.md
git commit -m "docs: update CHANGELOG for v1.9.1"
git push -u origin docs/changelog-v1.9.1
gh pr create --fill
```

Direct pushes to `main` are rejected. Branch protection requires a pull request
and the `Lint`, `Build`, `Test (1.26)` and `Security (nox)` checks, and it
applies to administrators too.

#### 2. Create Release Tag (Via GitHub UI)

1. Go to: https://github.com/felixgeelhaar/jirasdk/actions/workflows/tag.yml
2. Click "Run workflow"
3. Fill in the form:
   - **Version**: Enter version (e.g., `v1.9.1`, `v1.10.0-rc.1`)
   - **Prerelease**: annotates the tag message only — see below
4. Click "Run workflow"

The workflow will:
- ✅ Validate version format
- ✅ Check if tag already exists
- ✅ Run full test suite
- ✅ Create and push the tag
- ✅ Trigger the release workflow automatically

**The "Prerelease" checkbox does not mark the GitHub Release as a prerelease.**
It only adds a note to the tag message. The release flag is derived from the tag
*name*: `release.yml` marks a release as a prerelease when the version contains
`-rc`, `-beta` or `-alpha`. So `v2.0.0` with the box ticked publishes a full
release, and `v2.0.0-rc.1` with it unticked publishes a prerelease. Name the tag
correctly and the checkbox does not matter.

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
3. Check pkg.go.dev at the version you released:
   https://pkg.go.dev/github.com/felixgeelhaar/jirasdk@vX.Y.Z
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

1. **Verify pkg.go.dev**: check the new version renders, substituting the
   version you released:
   https://pkg.go.dev/github.com/felixgeelhaar/jirasdk@vX.Y.Z
   The bare URL https://pkg.go.dev/github.com/felixgeelhaar/jirasdk shows
   whatever pkg.go.dev considers latest, which is not proof this release
   indexed.
2. **Test Installation**: verify users can install it
   ```bash
   go get github.com/felixgeelhaar/jirasdk@vX.Y.Z
   ```
3. **Monitor Issues**: Watch for bug reports from new version
4. **Update Examples**: Ensure all examples work with new version

## Hotfix Process

For urgent bug fixes:

1. Create hotfix branch from the tag being fixed:
   ```bash
   git checkout -b hotfix/v1.9.1 v1.9.0
   ```

2. Apply fixes and test:
   ```bash
   # Make changes
   git add .
   git commit -m "fix: Critical bug in feature X"
   ```

3. Merge to main through a pull request. A direct push is rejected, including
   for administrators, so the hotfix branch goes through review and checks like
   anything else:
   ```bash
   git push -u origin hotfix/v1.9.1
   gh pr create --fill --base main
   # once the required checks pass:
   gh pr merge --rebase --delete-branch
   ```

   If the urgency is such that waiting on checks is itself the problem, the
   checks are the thing to make faster — not the gate to route around. There is
   no supported way to bypass it.

4. Create patch release:
   - Use the "Tag Release" workflow with v1.9.1

## Rollback

If a release has critical issues:

1. **Don't delete or move the tag.** The module proxy and the checksum database
   have already recorded the original permanently; see "Tag already exists"
   under Troubleshooting for what actually happens if you try.
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

Which case you are in decides everything.

**The tag exists only locally and was never pushed.** Delete it and start again;
nothing outside your clone has seen it.

```bash
git tag -d vX.Y.Z
```

**The tag was pushed.** Treat that version number as spent and release the next
one. Do not delete and recreate it.

```bash
# Fix the problem on main, then tag the next patch version.
# If v1.9.0 is the spent one, that is v1.9.1:
git tag -a v1.9.1 -m "Release v1.9.1"
git push origin v1.9.1
```

Recreating a pushed tag does not do what it looks like it does. Two independent
services have already recorded the original:

- `proxy.golang.org` caches module content immutably. Consumers keep receiving
  the original code no matter what the tag now points at, so the "fix" is
  invisible to exactly the people it was meant for.
- `sum.golang.org` is an append-only transparency log. Once a version's hash is
  recorded it cannot be removed — check any published version with
  `curl https://sum.golang.org/lookup/github.com/felixgeelhaar/jirasdk@vX.Y.Z`.

So after a retag the tag and the published module disagree, and anyone fetching
without a warm cache gets a checksum mismatch reported as a SECURITY ERROR —
which reads like a supply-chain attack on your own library. A new patch version
costs nothing by comparison.

### pkg.go.dev not updating

1. Wait 15-30 minutes (can take time)
2. Manually trigger, substituting the version you released:
   ```bash
   curl "https://proxy.golang.org/github.com/felixgeelhaar/jirasdk/@v/vX.Y.Z.info"
   ```
3. Check https://pkg.go.dev/github.com/felixgeelhaar/jirasdk@vX.Y.Z

## Getting Help

- **Workflow Issues**: Check [GitHub Actions Documentation](https://docs.github.com/en/actions)
- **Versioning Questions**: See [Semantic Versioning](https://semver.org/)
- **Go Module Issues**: See [Go Modules Reference](https://go.dev/ref/mod)

---

**Last Updated**: 2026-09-11
