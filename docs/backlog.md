
## Bump go.mod min to 1.25 + drop Go 1.24

Strategic Go-version policy change. Current go.mod=go 1.24.6; MEMORY says support last 3 (1.24/1.25/1.26). Conflict: oauth2 0.36, bolt 1.3.0, fortify 1.3.1 all require go≥1.25. Action: bump go.mod to 1.25.0, drop 1.24 from CI matrix, update CHANGELOG + migration note, ship as v1.7.0. Rationale: align with industry norm (2-version support) and unblock dep updates. Risk: consumers on 1.24 must pin v1.6.1. Acceptance: CI matrix=[1.25,1.26], all 3 deferred PRs (#54 #50 #43) re-mergeable, ADR doc filed.

---

## Adapt resilience/fortify to fortify v1.3.x API

fortify v1.3.1 has breaking API: ratelimit.New takes Config by value (not *Config); retry.Retry[T] interface no longer exposes Do (or signature changed). Files: resilience/fortify/fortify.go lines 74, 150. Action: read v1.3.1 retry.go + ratelimit.go, refactor adapter, add contract test that locks expected interface shape. Acceptance: build passes against fortify v1.3.x, existing adapter tests pass, new contract test pinned to current major.

---

## CI guard for dependency Go-version drift

Quality lens finding: 3 dependabot PRs blocked because deps require newer Go than our go.mod min. Need early-warning. Action: nox/CI step that scans go.sum entries' go.mod files for minimum-go directive; fails (or warns) when any dep min > our min. Use `go mod download -json` + parse each dep's go.mod. Acceptance: CI step added to .github/workflows, runs on PR, surfaces drift with actionable message.

---

## Verify CI release pipeline post-major-action-bumps

Three CI actions bumped major: codeql-action 3→4 (#63), upload-artifact 6→7 (#62), action-gh-release 2→3 (#61). Quality risk: release pipeline may break on next tag. Action: dry-run release flow on a no-op tag (e.g., v1.6.2-rc.1) or workflow_dispatch; verify artifact upload, GH release creation, CodeQL findings ingestion. Acceptance: dry-run succeeds OR fixes applied before v1.7.0 cut.

---

## Document OpenTelemetry span attrs for AI/agent observability

otel 1.41 merged. AI lens: SDK frequently used as backend for MCP servers + LLM agent pipelines. Document trace span names + attributes (jira.issue.key, jira.method, jira.endpoint) so consumers wiring traces into LLM observability tooling get consistent labels. Action: add docs/observability.md or expand README; list span schema. Acceptance: doc page committed; example collector pipeline shown.

---

## Cut v1.7.0 release bundling go-bump + deferred deps

GTM/Product lens: bundle Go policy change + 3 deferred dep updates (oauth2 0.36, bolt 1.3.0, fortify 1.3.1) + adapter rewrite into single v1.7.0 minor release. Includes CHANGELOG migration block, semver justification (Go bump = consumer-facing breaking → minor on libs is acceptable but doc clearly), README badge update. Acceptance: tagged v1.7.0, CHANGELOG entry, GitHub release notes, all 3 deferred PRs closed/superseded.

---
