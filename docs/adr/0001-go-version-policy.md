# ADR 0001: Go Version Support Policy

- Status: Accepted
- Date: 2026-05-09
- Deciders: maintainers

## Context

Three open dependency upgrades blocked merging because the upstream libraries
require a newer Go toolchain than this module declares:

- `golang.org/x/oauth2` v0.36 → `go ≥ 1.25.0`
- `github.com/felixgeelhaar/bolt` v1.3.0 → `go ≥ 1.25.0`
- `github.com/felixgeelhaar/fortify` v1.3.1 → `go ≥ 1.25.0`

Prior policy (recorded in maintainer memory) was "support the last three Go
versions" — currently 1.24, 1.25, 1.26. Holding 1.24 forces holding back the
above dependencies, which carries security and feature debt and creates a
recurring backlog of blocked Dependabot PRs.

## Decision

Adopt a **last-two-Go-versions** support policy.

- Minimum: `go 1.25.0` (declared in `go.mod`).
- CI matrix: `[1.25, 1.26]`.
- Branch protection required checks updated to remove `Test (1.24)`.

This brings jirasdk in line with industry norm (`x/...`, `kubernetes`,
`prometheus`) and unblocks dependency hygiene.

## Consequences

### Positive

- Dependency upgrades unblock immediately (oauth2, bolt, fortify, and any future
  library that follows the Go release cadence).
- Reduced security debt — no more "security patch is in newer dep but newer dep
  needs newer Go" stalemates.
- Aligns with toolchain-aware Go (`GOTOOLCHAIN`) defaults that ship Go ≥ 1.25.

### Negative / Mitigations

- Consumers still on Go 1.24 cannot upgrade past `v1.6.1`. Mitigation:
  the `v1.6.x` line stays buildable on 1.24; security fixes can backport on
  request.
- Policy departure from prior commitment. Mitigation: documented in CHANGELOG
  under "BREAKING" and called out in the v1.7.0 release notes.

## References

- Deferred PRs: #54 (oauth2), #50 (bolt), #43 (fortify)
- CHANGELOG entry: v1.7.0 Unreleased section
