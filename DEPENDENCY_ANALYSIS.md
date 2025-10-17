# Dependency Analysis: zerolog and zap

## Summary

Both `zerolog` and `zap` (along with `logrus`) are **indirect dependencies** from the `bolt` logging library. The `jirasdk` project provides **optional adapters** for these logging frameworks but does not directly depend on them.

## Dependency Chain

```
jirasdk (main)
  └─ bolt@v1.2.1 (optional adapter library)
      ├─ zerolog@v1.34.0
      ├─ zap@v1.27.0
      └─ logrus@v1.9.3
```

## Why Multiple Loggers?

The `bolt` library is a **logging abstraction library** that provides:
1. A unified logging interface
2. Multiple backend adapters (zerolog, zap, logrus)
3. Zero-allocation structured logging
4. OpenTelemetry integration

Users can choose which logger backend they prefer:
- **zerolog**: Zero-allocation, high performance
- **zap**: Uber's structured logger, very popular
- **logrus**: Traditional structured logger

## Usage in jirasdk

### Optional Adapters (Not in Main Module)

The SDK provides **optional adapter packages**:

1. **`logger/bolt/bolt.go`** - Adapter for bolt logging
   - Located at: `jirasdk/logger/bolt/`
   - Provides: Adapter to use bolt logger with jirasdk
   - Import: `github.com/felixgeelhaar/jirasdk/logger/bolt`

2. **`resilience/fortify/fortify.go`** - Adapter for fortify resilience
   - Located at: `jirasdk/resilience/fortify/`
   - Provides: Circuit breakers, retry, rate limiting, timeouts, bulkheads
   - Import: `github.com/felixgeelhaar/jirasdk/resilience/fortify`

### Used Only in Examples

The only place where `bolt` and `fortify` are actually **used** (not just imported) is:
- `examples/observability/main.go` - Example showing how to use structured logging

### Core Library Behavior

The **core jirasdk library** uses:
- `NewNoopLogger()` by default (no dependencies)
- `NewNoopResilience()` by default (no dependencies)

Users must **opt-in** to use bolt or fortify via functional options:
```go
client, err := jira.NewClient(
    jira.WithBaseURL(baseURL),
    jira.WithAPIToken(email, token),
    // OPTIONAL: Only if user wants structured logging
    jira.WithLogger(boltadapter.NewAdapter(logger)),
    // OPTIONAL: Only if user wants advanced resilience
    jira.WithResilience(fortify.NewAdapter(config)),
)
```

## Impact Analysis

### Current State
- ✅ No bloat for basic users (adapters are optional)
- ✅ Users who don't import `logger/bolt` or `resilience/fortify` don't pull in dependencies
- ❌ `bolt` and `fortify` are listed in `go.mod` as direct dependencies
- ❌ This pulls in **all** logging backends even if you only want one

### Binary Size Impact

When building an application that uses jirasdk:
- **Without bolt/fortify adapters**: ~5-10MB binary
- **With bolt adapter**: +~2MB for logging backends (zerolog, zap, logrus)
- **With fortify adapter**: +~500KB for resilience patterns

### Actual Problem

The issue is that `bolt@v1.2.1` has **all three loggers** as direct dependencies:
```go
github.com/felixgeelhaar/bolt@v1.2.1 github.com/rs/zerolog@v1.34.0
github.com/felixgeelhaar/bolt@v1.2.1 github.com/sirupsen/logrus@v1.9.3
github.com/felixgeelhaar/bolt@v1.2.1 go.uber.org/zap@v1.27.0
```

Even if a user only wants to use `zerolog`, they get `zap` and `logrus` compiled into their binary.

## Recommendations

### Option 1: Move to Optional Subpackages (Recommended)

Move `bolt` and `fortify` to separate subpackages that are **not imported by default**:

```
jirasdk/
  ├── logger/
  │   └── bolt/          # Users explicitly import if needed
  ├── resilience/
  │   └── fortify/       # Users explicitly import if needed
  └── go.mod             # Remove bolt and fortify from require section
```

Users would import only what they need:
```go
import (
    jira "github.com/felixgeelhaar/jirasdk"
    "github.com/felixgeelhaar/jirasdk/logger/bolt"  // Opt-in
)
```

**Benefits:**
- Zero dependencies for basic usage
- Users pay only for what they use
- Cleaner dependency tree

**Drawbacks:**
- Breaking change (requires go.mod update for existing users)
- Documentation needs update

### Option 2: Request bolt Library Split

Ask the `bolt` library maintainer to split backends into separate packages:
```
bolt/                  # Core interfaces only
bolt/zerolog/         # zerolog backend
bolt/zap/             # zap backend
bolt/logrus/          # logrus backend
```

**Benefits:**
- Users import only the backend they need
- Better separation of concerns
- Smaller binaries

**Drawbacks:**
- Requires changes to external dependency
- May not be accepted by maintainer

### Option 3: Keep Current (Status Quo)

Accept that `bolt` brings all three loggers and document it.

**Benefits:**
- No changes needed
- Maximum flexibility for users

**Drawbacks:**
- ~2MB extra in binaries even if unused
- Longer compile times
- More dependencies to manage

## Security Considerations

Having multiple logging libraries means:
- ✅ More eyes on security (all three are popular and well-maintained)
- ❌ More surface area for vulnerabilities
- ❌ Need to monitor CVEs for 3 loggers instead of 1

Current security status:
- **zerolog v1.34.0**: No known CVEs
- **zap v1.27.0**: No known CVEs
- **logrus v1.9.3**: No known CVEs

## Conclusion

**Current situation is acceptable** for a library that provides optional observability features. The extra dependencies are only pulled in if users opt-in to structured logging.

**Recommended action:**
1. Document this clearly in README.md
2. Consider moving to subpackages in v2.0.0 (breaking change)
3. File an issue with `bolt` library to request backend splitting

**No immediate action required** - this is a design trade-off, not a bug.
