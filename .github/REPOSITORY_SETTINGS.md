# GitHub Repository Settings

This document contains the recommended settings for the jirasdk GitHub repository.

## Repository Description

```
Enterprise-grade Go client for Jira Cloud & Server/Data Center REST APIs. Features resilience patterns, environment config, zero-allocation logging, and comprehensive documentation. Production-ready with full context support.
```

## Repository Topics (Tags)

Add the following topics to make the repository discoverable:

### Primary Topics
- `jira`
- `jira-api`
- `jira-client`
- `go`
- `golang`
- `sdk`

### Technology Topics
- `rest-api`
- `api-client`
- `jira-cloud`
- `jira-server`
- `atlassian`

### Feature Topics
- `resilience-patterns`
- `circuit-breaker`
- `retry-logic`
- `rate-limiting`
- `structured-logging`
- `environment-variables`

### Development Topics
- `enterprise`
- `production-ready`
- `type-safe`
- `context-propagation`
- `oauth2`

### Quality Topics
- `well-documented`
- `tested`
- `ci-cd`

## How to Set Topics on GitHub

1. Go to the repository: https://github.com/felixgeelhaar/jirasdk
2. Click the gear icon ⚙️ next to "About" on the right sidebar
3. In the "Topics" field, add the topics listed above
4. Click "Save changes"

## Repository Settings

### General Settings

**Description:**
```
Enterprise-grade Go client for Jira Cloud & Server/Data Center REST APIs. Features resilience patterns, environment config, zero-allocation logging, and comprehensive documentation. Production-ready with full context support.
```

**Website:**
```
https://pkg.go.dev/github.com/felixgeelhaar/jirasdk
```

**Features to Enable:**
- ✅ Issues
- ✅ Discussions (recommended for community Q&A)
- ✅ Projects
- ✅ Wiki (optional - we have comprehensive docs in repo)
- ✅ Sponsorships (optional)

**Features to Disable:**
- ❌ Merge commits (use squash or rebase instead)

### Branch Protection Rules (main)

**Protect matching branches:**
- ✅ Require a pull request before merging
  - ✅ Require approvals: 1
  - ✅ Dismiss stale pull request approvals when new commits are pushed
- ✅ Require status checks to pass before merging
  - ✅ Require branches to be up to date before merging
  - Required checks: `test`, `lint`, `security`, `build`
- ✅ Require conversation resolution before merging
- ✅ Require linear history
- ✅ Include administrators
- ✅ Allow force pushes: Nobody
- ✅ Allow deletions: Disabled

### Security Settings

**Code security and analysis:**
- ✅ Dependency graph: Enabled
- ✅ Dependabot alerts: Enabled
- ✅ Dependabot security updates: Enabled
- ✅ Code scanning: GitHub CodeQL (configure)
- ✅ Secret scanning: Enabled
- ✅ Secret scanning push protection: Enabled

### Actions Settings

**General:**
- ✅ Allow all actions and reusable workflows

**Workflow permissions:**
- ✅ Read and write permissions
- ✅ Allow GitHub Actions to create and approve pull requests

### Pages (Optional)

If you want to host documentation:
- Source: Deploy from a branch
- Branch: `gh-pages` or create a `docs` branch
- Folder: `/` or `/docs`

## Social Preview

Create a social preview image with:
- Repository name: "jirasdk"
- Tagline: "Enterprise Jira Client for Go"
- Key features: Resilience | Observability | Type-Safe
- Background: Professional gradient or Go gopher theme
- Dimensions: 1280x640px

Upload via:
1. Settings → General → Social preview
2. Upload image
3. Save

## README Badges

The following badges are already in the README:
- Go Version
- Go Reference (pkg.go.dev)
- Go Report Card
- License (MIT)

Consider adding:
- Build Status: `[![CI](https://github.com/felixgeelhaar/jirasdk/actions/workflows/ci.yml/badge.svg)](https://github.com/felixgeelhaar/jirasdk/actions/workflows/ci.yml)`
- Release: `[![Release](https://img.shields.io/github/v/release/felixgeelhaar/jirasdk)](https://github.com/felixgeelhaar/jirasdk/releases)`
- Coverage: `[![codecov](https://codecov.io/gh/felixgeelhaar/jirasdk/branch/main/graph/badge.svg)](https://codecov.io/gh/felixgeelhaar/jirasdk)` (if using Codecov)

## Community Health Files

Already present:
- ✅ README.md
- ✅ LICENSE
- ✅ CONTRIBUTING.md
- ✅ SECURITY.md
- ✅ CHANGELOG.md
- ✅ Issue templates
- ✅ Pull request template

Consider adding:
- [ ] CODE_OF_CONDUCT.md (optional)
- [ ] SUPPORT.md (where to get help)
- [ ] FUNDING.yml (for sponsorships)

## GitHub Discussions Categories

If enabling Discussions, create these categories:
1. **Announcements** - Official updates and releases
2. **General** - General discussion about jirasdk
3. **Q&A** - Questions from the community
4. **Ideas** - Feature requests and ideas
5. **Show and Tell** - Community showcases
6. **Troubleshooting** - Help with issues

## Applying These Settings

### Via GitHub UI:
1. Go to repository Settings
2. Update each section as described above

### Via GitHub CLI:
```bash
# Update repository description and topics
gh repo edit felixgeelhaar/jirasdk \
  --description "Enterprise-grade Go client for Jira Cloud & Server/Data Center REST APIs. Features resilience patterns, environment config, zero-allocation logging, and comprehensive documentation. Production-ready with full context support." \
  --homepage "https://pkg.go.dev/github.com/felixgeelhaar/jirasdk" \
  --add-topic jira \
  --add-topic jira-api \
  --add-topic jira-client \
  --add-topic go \
  --add-topic golang \
  --add-topic sdk \
  --add-topic rest-api \
  --add-topic api-client \
  --add-topic jira-cloud \
  --add-topic jira-server \
  --add-topic atlassian \
  --add-topic resilience-patterns \
  --add-topic circuit-breaker \
  --add-topic retry-logic \
  --add-topic rate-limiting \
  --add-topic structured-logging \
  --add-topic environment-variables \
  --add-topic enterprise \
  --add-topic production-ready \
  --add-topic type-safe \
  --add-topic context-propagation \
  --add-topic oauth2 \
  --add-topic well-documented \
  --add-topic tested \
  --add-topic ci-cd
```

---

**Last Updated**: 2025-01-08
