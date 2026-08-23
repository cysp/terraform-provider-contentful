#!/usr/bin/env bash

set -euo pipefail

repository="${1:?repository is required}"
release_tag="${2:?release tag is required}"
candidate_sha="${3:?candidate SHA is required}"

matching_drafts="$(
  gh api --paginate --slurp "repos/${repository}/releases?per_page=100" |
    jq --arg release_tag "${release_tag}" '
  [
    .[][]
    | select(
        .draft == true
        and .tag_name == $release_tag
        and .name == $release_tag
      )
  ]
  '
)"

draft_count="$(jq 'length' <<<"${matching_drafts}")"
if [[ "${draft_count}" != "1" ]]; then
  echo "::error::Expected exactly one draft release whose tag and title are ${release_tag}; found ${draft_count}"
  exit 1
fi

target_commitish="$(jq -r '.[0].target_commitish' <<<"${matching_drafts}")"
if [[ "${target_commitish}" != "${candidate_sha}" ]]; then
  echo "::error::Release draft ${release_tag} targets ${target_commitish}, not candidate ${candidate_sha}"
  exit 1
fi

if ! jq -e '((.[0].body // "") | test("\\S"))' >/dev/null <<<"${matching_drafts}"; then
  echo "::error::Release draft ${release_tag} has empty release notes"
  exit 1
fi

asset_count="$(jq '.[0].assets | length' <<<"${matching_drafts}")"
if [[ "${asset_count}" != "0" ]]; then
  echo "::error::Release draft ${release_tag} already has ${asset_count} asset(s); release assets must come from the validated workflow"
  exit 1
fi

echo "Release draft ${release_tag} is bound to candidate ${candidate_sha} and contains release notes."
