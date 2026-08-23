#!/usr/bin/env bash

set -euo pipefail

repository="${1:?repository is required}"
release_tag="${2:?release tag is required}"
candidate_sha="${3:?candidate SHA is required}"
expected_assets="${4:-empty}"
expected_release_id="${5:-}"
expected_asset_names=()
if (( $# > 5 )); then
  expected_asset_names=("${@:6}")
fi

case "${expected_assets}" in
  empty)
    if [[ -n "${expected_release_id}" ]]; then
      echo "::error::The empty asset state does not accept an expected release ID"
      exit 1
    fi
    if (( ${#expected_asset_names[@]} != 0 )); then
      echo "::error::The empty asset state does not accept expected asset names"
      exit 1
    fi
    ;;
  present)
    if [[ -z "${expected_release_id}" ]]; then
      echo "::error::The present asset state requires the original release ID"
      exit 1
    fi
    if (( ${#expected_asset_names[@]} == 0 )); then
      echo "::error::The present asset state requires the complete expected asset set"
      exit 1
    fi
    ;;
  *)
    echo "::error::Expected asset state must be empty or present, not ${expected_assets}"
    exit 1
    ;;
esac

matching_drafts="$(
  gh api --paginate --slurp "repos/${repository}/releases?per_page=100" |
    jq --arg release_tag "${release_tag}" '
  [
    .[][]
    | select(
        .draft == true
        and .name == $release_tag
      )
  ]
  '
)"

draft_count="$(jq 'length' <<<"${matching_drafts}")"
if [[ "${draft_count}" != "1" ]]; then
  echo "::error::Expected exactly one draft release whose title is ${release_tag}; found ${draft_count}"
  exit 1
fi

draft_id="$(jq -r '.[0].id' <<<"${matching_drafts}")"
if [[ -n "${expected_release_id}" && "${draft_id}" != "${expected_release_id}" ]]; then
  echo "::error::Release draft ${release_tag} has ID ${draft_id}, not original draft ID ${expected_release_id}"
  exit 1
fi

tag_name="$(jq -r '.[0].tag_name' <<<"${matching_drafts}")"
if [[ "${tag_name}" != "${release_tag}" ]]; then
  echo "::error::Release draft titled ${release_tag} uses tag ${tag_name}, not ${release_tag}"
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
case "${expected_assets}" in
  empty)
    if [[ "${asset_count}" != "0" ]]; then
      echo "::error::Release draft ${release_tag} already has ${asset_count} asset(s); release assets must come from the validated workflow"
      exit 1
    fi
    ;;
  present)
    if [[ "${asset_count}" == "0" ]]; then
      echo "::error::Release draft ${release_tag} has no assets after GoReleaser"
      exit 1
    fi

    actual_asset_names="$(jq -r '.[0].assets[].name' <<<"${matching_drafts}" | LC_ALL=C sort)"
    expected_asset_basenames="$({
      for expected_asset_name in "${expected_asset_names[@]}"; do
        printf '%s\n' "${expected_asset_name##*/}"
      done
    } | LC_ALL=C sort -u)"
    if [[ "${actual_asset_names}" != "${expected_asset_basenames}" ]]; then
      echo "::error::Release draft ${release_tag} assets do not match the complete GoReleaser output"
      diff \
        <(printf '%s\n' "${expected_asset_basenames}") \
        <(printf '%s\n' "${actual_asset_names}") || true
      exit 1
    fi
    ;;
esac

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  printf 'release_id=%s\n' "${draft_id}" >>"${GITHUB_OUTPUT}"
fi

echo "Release draft ${release_tag} (${draft_id}) is bound to candidate ${candidate_sha}, contains release notes, and has the expected asset state (${expected_assets})."
