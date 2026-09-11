#!/usr/bin/env bash
# Runs golangci-lint at exactly the version and Go toolchain CI uses, both read
# from .lint-toolchain.
#
# Catches the failure mode that made PR #99 fail Lint after a clean local run:
# the local golangci-lint was NEWER than CI's (v2.13.1 vs v2.10.1) and no longer
# raised gosec's G704 and G117, so "lint passes locally" meant nothing. Fixing
# the reported findings then surfaced a further one that had been masked, which
# a second CI round would have been needed to discover.
#
# The pinned linter is cached under .tools/ and reused. Pass any extra
# golangci-lint arguments through, for example:
#
#     scripts/lint.sh --build-tags live
#     scripts/lint.sh --fix

set -euo pipefail

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$REPO_ROOT"

# Read the pinned versions by parsing, never by sourcing. Sourcing would
# execute .lint-toolchain, so anyone able to edit it would gain arbitrary code
# execution in every contributor's shell and in CI. Both values then go into a
# download URL and a command line, so each is validated against a strict shape
# before use; anything unexpected stops the script rather than reaching a sink.
read_pin() {
    local key=$1
    sed -n "s/^${key}=\([^#]*\).*/\1/p" .lint-toolchain | tr -d '[:space:]' | head -n1
}

GOLANGCI_LINT_VERSION=$(read_pin GOLANGCI_LINT_VERSION)
GO_VERSION=$(read_pin GO_VERSION)

if [[ ! "$GOLANGCI_LINT_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "ERROR: GOLANGCI_LINT_VERSION in .lint-toolchain must look like v1.2.3, got '${GOLANGCI_LINT_VERSION}'" >&2
    exit 2
fi

if [[ ! "$GO_VERSION" =~ ^[0-9]+\.[0-9]+$ ]]; then
    echo "ERROR: GO_VERSION in .lint-toolchain must look like 1.26, got '${GO_VERSION}'" >&2
    exit 2
fi

TOOLS_DIR="$REPO_ROOT/.tools"
BIN="$TOOLS_DIR/golangci-lint-$GOLANGCI_LINT_VERSION"

if [[ ! -x "$BIN" ]]; then
    echo "Installing golangci-lint $GOLANGCI_LINT_VERSION into .tools ..."
    mkdir -p "$TOOLS_DIR"

    case "$(uname -s)" in
        Darwin) os=darwin ;;
        Linux)  os=linux ;;
        *)      echo "ERROR: unsupported OS $(uname -s)" >&2; exit 2 ;;
    esac

    case "$(uname -m)" in
        arm64|aarch64) arch=arm64 ;;
        x86_64|amd64)  arch=amd64 ;;
        *)             echo "ERROR: unsupported architecture $(uname -m)" >&2; exit 2 ;;
    esac

    stripped=${GOLANGCI_LINT_VERSION#v}
    name="golangci-lint-${stripped}-${os}-${arch}"
    url="https://github.com/golangci/golangci-lint/releases/download/${GOLANGCI_LINT_VERSION}/${name}.tar.gz"

    tmp=$(mktemp -d)
    trap 'rm -rf "$tmp"' EXIT

    if ! curl -sSfL "$url" | tar xz -C "$tmp" --strip-components=1; then
        echo "ERROR: failed to download $url" >&2
        exit 2
    fi

    mv "$tmp/golangci-lint" "$BIN"
    chmod +x "$BIN"
fi

# Pin the analysing toolchain too. GOTOOLCHAIN makes the go command fetch this
# release on demand, so no separate install step is needed.
export GOTOOLCHAIN="go${GO_VERSION}.0"

GOROOT=$(go env GOROOT)
export PATH="$GOROOT/bin:$PATH"

echo "golangci-lint $GOLANGCI_LINT_VERSION, Go $(go env GOVERSION)"
exec "$BIN" run --timeout=5m "$@"
