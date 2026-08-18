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
# Bumps every `uses:` reference in .github/workflows to the latest published
# release, preserving whatever pin style each line already uses:
#
#   uses: owner/repo@v6                 ->  @v7                  (major only)
#   uses: owner/repo@v0.9               ->  @v0.10               (major.minor)
#   uses: owner/repo@v1.0.2             ->  @v2.0.0              (full tag)
#   uses: owner/repo@<sha> # v6.0.1     ->  @<sha> # v7.0.1      (commit pin)
#
# It also bumps the `version:` input of golangci/golangci-lint-action, which
# pins the linter binary rather than the action.
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

## Latest release tag for a repository. Falls back to the most recent tag for
## repositories that publish tags without GitHub releases.
function latest_tag() {
    local repo="$1" tag
    tag="$(gh api "repos/${repo}/releases/latest" --jq '.tag_name' </dev/null 2>/dev/null || true)"
    if [ -z "$tag" ]; then
        tag="$(gh api "repos/${repo}/tags" --jq '.[0].name' </dev/null 2>/dev/null || true)"
    fi
    echo "$tag"
}

## Commit SHA a tag points at. Works for lightweight and annotated tags alike.
function tag_sha() {
    local repo="$1" tag="$2"
    gh api "repos/${repo}/commits/${tag}" --jq '.sha' </dev/null 2>/dev/null || true
}

## Narrow a full semver tag down to the precision the existing pin used, so that
## a floating `@v6` stays floating and a `@v1.0.2` stays fully qualified.
function match_precision() {
    local current="$1" latest="$2"
    case "$current" in
    v[0-9]*.[0-9]*.[0-9]*) echo "$latest" ;;
    v[0-9]*.[0-9]*) echo "$latest" | cut -d. -f1,2 ;;
    v[0-9]*) echo "$latest" | cut -d. -f1 ;;
    *) echo "$latest" ;;
    esac
}

## Rewrite the `uses:` lines of a single workflow file in place.
function bump_uses() {
    local file="$1" line action ref comment repo latest narrowed sha key summary

    # Read on fd 3: `gh` runs inside this loop and would otherwise drain stdin.
    while IFS= read -r line <&3; do
        # uses: owner/repo[/subpath]@ref [# comment]
        [[ "$line" =~ uses:[[:space:]]+([^[:space:]@]+)@([^[:space:]]+)(.*)$ ]] || continue
        action="${BASH_REMATCH[1]}"
        ref="${BASH_REMATCH[2]}"
        comment="${BASH_REMATCH[3]}"

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
        if [ -n "${RESOLVED[$key]+set}" ]; then
            # Resolved while walking an earlier workflow; still rewrite this file.
            if [ -n "${RESOLVED[$key]}" ]; then
                replace_ref "$file" "$action" "$ref" "${RESOLVED[$key]}"
                echo "  -> ${action}: ${ref} (see above)"
            fi
            continue
        fi
        RESOLVED[$key]=""

        # Strip any subpath: anchore/sbom-action/download-syft -> anchore/sbom-action
        repo="$(echo "$action" | cut -d/ -f1,2)"

        latest="$(latest_tag "$repo")"
        if [ -z "$latest" ]; then
            echo "  ?? ${action}@${ref} (no releases or tags found, skipped)"
            continue
        fi

        if [[ "$ref" =~ ^[0-9a-f]{40}$ ]]; then
            sha="$(tag_sha "$repo" "$latest")"
            if [ -z "$sha" ]; then
                echo "  ?? ${action}@${ref} (could not resolve ${latest}, skipped)"
                continue
            fi
            if [ "$sha" = "$ref" ]; then
                continue
            fi
            RESOLVED[$key]="uses: ${action}@${sha} # ${latest}"
            summary="  -> ${action}: ${ref:0:7} -> ${sha:0:7} (${latest})"
        else
            narrowed="$(match_precision "$ref" "$latest")"
            if [ "$narrowed" = "$ref" ]; then
                continue
            fi
            RESOLVED[$key]="uses: ${action}@${narrowed}${comment}"
            summary="  -> ${action}: ${ref} -> ${narrowed}"
        fi

        replace_ref "$file" "$action" "$ref" "${RESOLVED[$key]}"
        echo "$summary"
    done 3<"$file"
}

## Substitute every `uses: <action>@<ref>` in a file. The ref must be followed by
## end-of-line or whitespace so that `@v6` never matches a `@v6.0.1` line.
function replace_ref() {
    local file="$1" action="$2" ref="$3" replacement="$4"
    sed -i "s|uses: ${action}@${ref}\([[:space:]].*\)\?$|${replacement}|" "$file"
}

## The golangci-lint binary is pinned through an action input, not an action ref.
function bump_golangci_lint() {
    local file="$1" current latest
    grep -q 'golangci/golangci-lint-action' "$file" || return 0

    current="$(sed -n 's/^[[:space:]]*version:[[:space:]]*\(v[0-9][^[:space:]]*\)[[:space:]]*$/\1/p' "$file" | head -1)"
    if [ -z "$current" ]; then
        return 0
    fi

    latest="$(latest_tag golangci/golangci-lint)"
    if [ -z "$latest" ] || [ "$latest" = "$current" ]; then
        return 0
    fi

    sed -i "s|^\([[:space:]]*version:[[:space:]]*\)${current}[[:space:]]*$|\1${latest}|" "$file"
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
