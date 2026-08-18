#!/bin/bash
#
# Copyright (c) 2015-2026 MinIO, Inc.
#
# This file is part of MinIO Object Storage stack
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
#
# Bumps every `uses:` reference in .github/workflows to the commit behind the
# latest published release:
#
#   uses: owner/repo@v6              ->  uses: owner/repo@<sha> # v7.0.1
#   uses: owner/repo@<sha> # v0.5.3  ->  uses: owner/repo@<sha> # v0.6.2
#
# Everything is pinned to a hash because that is what the repository's zizmor
# policy requires; the trailing comment carries the human-readable version.
#
# It also bumps the `version:` input of golangci/golangci-lint-action, which
# pins the linter binary rather than the action.
#
# Anything this script cannot resolve is a hard error rather than a skip. The
# caller closes the Dependabot pull requests it believes this sweep replaces,
# so a quietly incomplete sweep is worse than no sweep at all. An action that
# publishes no GitHub release has to be pinned by hand.
#
# Requires the `gh` CLI, authenticated (GH_TOKEN or `gh auth login`).
# Every rewrite is reported on stdout.

set -euo pipefail

# Enable tracing if set.
[ -n "${BASH_XTRACEFD:-}" ] && set -x

WORKFLOW_DIR="$(git rev-parse --show-toplevel)/.github/workflows"

## Replacement line for each "action@ref" resolved this run, so that an action
## used by several workflows costs a single API round trip. An empty value means
## "already checked, nothing to change".
declare -A RESOLVED

function checkdeps() {
    if ! command -v gh >/dev/null 2>&1; then
        echo "error: the 'gh' CLI is required but was not found in PATH" >&2
        exit 1
    fi
    if [ -z "${GH_TOKEN:-}" ] && ! gh auth status >/dev/null 2>&1; then
        echo "error: gh is not authenticated; set GH_TOKEN or run 'gh auth login'" >&2
        exit 1
    fi
}

## Latest release tag for a repository.
##
## Deliberately no fallback to the tag list: /repos/{owner}/{repo}/tags is not
## ordered by version, so its first entry can be a nightly or a release
## candidate, and `gh api` reports a 404 the same way it reports a 502. Taking
## tags[0] because /releases/latest blipped would pin, commit and potentially
## auto-merge the wrong revision.
function latest_tag() {
    local repo="$1" tag attempt
    for attempt in 1 2 3; do
        tag="$(gh api "repos/${repo}/releases/latest" --jq '.tag_name' </dev/null 2>/dev/null || true)"
        if [ -n "$tag" ]; then
            echo "$tag"
            return 0
        fi
        sleep $((attempt * 3))
    done
    return 1
}

## Commit SHA a tag points at. Works for lightweight and annotated tags alike.
function tag_sha() {
    local repo="$1" tag="$2" sha attempt
    for attempt in 1 2 3; do
        sha="$(gh api "repos/${repo}/commits/${tag}" --jq '.sha' </dev/null 2>/dev/null || true)"
        if [ -n "$sha" ]; then
            echo "$sha"
            return 0
        fi
        sleep $((attempt * 3))
    done
    return 1
}

## Quote the characters sed would otherwise read as ERE syntax. Action names
## carry dots often enough for this to matter.
function regex_escape() {
    printf '%s' "$1" | sed -e 's/[.[\*^$+?(){}|]/\\&/g'
}

## Substitute `uses: <action>@<ref>` in a file, accepting the same whitespace the
## caller's matcher accepted, and requiring the ref to be followed by whitespace
## or end-of-line so `@v6` never matches a `@v6.0.1` line.
##
## Failing when nothing was rewritten matters: a matcher that finds a line the
## rewriter then misses would log a bump that never happened, and the caller
## would go on to close the Dependabot pull request covering that action.
function replace_ref() {
    local file="$1" action="$2" ref="$3" replacement="$4" pattern backup

    pattern="uses:[[:space:]]+$(regex_escape "$action")@$(regex_escape "$ref")([[:space:]].*)?$"
    backup="$(mktemp)"
    cp "$file" "$backup"

    sed -i -E "s|${pattern}|${replacement}|" "$file"

    if cmp -s "$file" "$backup"; then
        rm -f "$backup"
        echo "error: matched ${action}@${ref} in $(basename "$file") but rewrote nothing" >&2
        return 1
    fi
    rm -f "$backup"
}

## Rewrite the `uses:` lines of a single workflow file in place.
function bump_uses() {
    local file="$1" line action ref repo latest sha key
    ## Keys already rewritten in this file. sed substitutes every matching line
    ## at once, so a second sighting of the same action is nothing left to do
    ## rather than a rewrite that failed.
    local -A applied=()

    # Read on fd 3: `gh` runs inside this loop and would otherwise drain stdin.
    while IFS= read -r line <&3; do
        # uses: owner/repo[/subpath]@ref [# comment]
        [[ "$line" =~ uses:[[:space:]]+([^[:space:]@]+)@([^[:space:]]+) ]] || continue
        action="${BASH_REMATCH[1]}"
        ref="${BASH_REMATCH[2]}"

        # Local (./path) and container (docker://) actions are not versioned here.
        case "$action" in
        ./* | docker://*) continue ;;
        esac

        # Only semver tags and full commit pins are understood; a branch ref such
        # as @main is intentional and must not be rewritten into a tag.
        if [[ ! "$ref" =~ ^v[0-9] ]] && [[ ! "$ref" =~ ^[0-9a-f]{40}$ ]]; then
            echo "  .. ${action}@${ref} (not a version pin, left alone)"
            continue
        fi

        key="${action}@${ref}"
        if [ -n "${applied[$key]:-}" ]; then
            continue
        fi
        applied[$key]=1

        if [ -z "${RESOLVED[$key]+set}" ]; then
            # Strip any subpath: anchore/sbom-action/download-syft -> anchore/sbom-action
            repo="$(echo "$action" | cut -d/ -f1,2)"

            if ! latest="$(latest_tag "$repo")"; then
                echo "error: no published release found for ${repo}" >&2
                return 1
            fi
            if ! sha="$(tag_sha "$repo" "${latest}")"; then
                echo "error: could not resolve ${repo}@${latest} to a commit" >&2
                return 1
            fi

            if [ "$sha" = "$ref" ]; then
                RESOLVED[$key]=""
            else
                RESOLVED[$key]="uses: ${action}@${sha} # ${latest}"
            fi
        fi

        if [ -n "${RESOLVED[$key]}" ]; then
            replace_ref "$file" "$action" "$ref" "${RESOLVED[$key]}"
            echo "  -> ${action}: ${ref} -> ${RESOLVED[$key]#*@}"
        fi
    done 3<"$file"
}

## The golangci-lint binary is pinned through an action input, not an action ref.
function bump_golangci_lint() {
    local file="$1" current latest
    grep -q 'golangci/golangci-lint-action' "$file" || return 0

    current="$(sed -n 's/^[[:space:]]*version:[[:space:]]*\(v[0-9][^[:space:]]*\)[[:space:]]*$/\1/p' "$file" | head -1)"
    if [ -z "$current" ]; then
        return 0
    fi

    if ! latest="$(latest_tag golangci/golangci-lint)"; then
        echo "error: could not resolve the latest golangci-lint release" >&2
        return 1
    fi
    if [ "$latest" = "$current" ]; then
        return 0
    fi

    sed -i "s|^\([[:space:]]*version:[[:space:]]*\)$(regex_escape "$current")[[:space:]]*$|\1${latest}|" "$file"
    echo "  -> golangci-lint: ${current} -> ${latest}"
}

function main() {
    checkdeps

    local file
    for file in "${WORKFLOW_DIR}"/*.yml "${WORKFLOW_DIR}"/*.yaml; do
        [ -e "$file" ] || continue
        echo "$(basename "$file"):"
        bump_uses "$file"
        bump_golangci_lint "$file"
    done
}

main "$@"
