#!/usr/bin/env bash
# Fails when any direct or transitive dependency declares a minimum Go version
# higher than this module's go.mod.
#
# Catches the failure mode that produced the v1.7.0 backlog: oauth2 / bolt /
# fortify all bumped their minimum Go to 1.25 while jirasdk was on 1.24, which
# left Dependabot PRs un-mergeable until the policy change.

set -euo pipefail

OWN_GO=$(awk '/^go [0-9]/ {print $2; exit}' go.mod)
if [[ -z "$OWN_GO" ]]; then
    echo "ERROR: cannot read 'go' directive from go.mod" >&2
    exit 2
fi

# Normalise to MAJOR.MINOR for comparison.
own_mm() { awk -F. -v v="$1" 'BEGIN{split(v,a,"."); printf "%d.%02d", a[1], a[2]}'; }
OWN_KEY=$(own_mm "$OWN_GO")

violations=0
while read -r mod; do
    [[ -z "$mod" ]] && continue
    [[ "$mod" == "$(go list -m)" ]] && continue
    info=$(go mod download -json "$mod" 2>/dev/null) || continue
    gomod=$(echo "$info" | awk -F'"' '/"GoMod":/ {print $4}')
    [[ -f "$gomod" ]] || continue
    dep_go=$(awk '/^go [0-9]/ {print $2; exit}' "$gomod")
    [[ -z "$dep_go" ]] && continue
    dep_key=$(own_mm "$dep_go")
    if [[ "$dep_key" > "$OWN_KEY" ]]; then
        echo "DRIFT: $mod requires go $dep_go (this module: $OWN_GO)"
        violations=$((violations + 1))
    fi
done < <(go list -m -f '{{.Path}}@{{.Version}}' all)

if (( violations > 0 )); then
    echo
    echo "$violations dependency Go-version drift(s) detected." >&2
    echo "Bump go.mod's 'go' directive or pin offending dependencies." >&2
    exit 1
fi

echo "OK: no dependency requires a Go version newer than $OWN_GO."
