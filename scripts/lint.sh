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

# shellcheck source=/dev/null
source .lint-toolchain

if [[ -z "${GOLANGCI_LINT_VERSION:-}" || -z "${GO_VERSION:-}" ]]; then
    echo "ERROR: .lint-toolchain must set GOLANGCI_LINT_VERSION and GO_VERSION" >&2
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
