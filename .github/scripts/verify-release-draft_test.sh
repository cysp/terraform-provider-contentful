#!/usr/bin/env bash

set -euo pipefail

repository_root="$(git rev-parse --show-toplevel)"
verifier="${repository_root}/.github/scripts/verify-release-draft.sh"
mock_bin="$(mktemp -d "${TMPDIR:-/tmp}/verify-release-draft.XXXXXX")"

cleanup() {
  rm -rf -- "${mock_bin}"
}
trap cleanup EXIT

cat >"${mock_bin}/gh" <<'EOF'
#!/usr/bin/env bash

set -euo pipefail

if [[ "${1:-}" != "api" ]]; then
  echo "unexpected gh command: $*" >&2
  exit 1
fi

printf '%s\n' "${MOCK_GH_RESPONSE:?mock GitHub API response is required}"
EOF
chmod +x "${mock_bin}/gh"

candidate_sha="1111111111111111111111111111111111111111"
empty_draft='[[{"id":123,"draft":true,"tag_name":"v1.2.3","name":"v1.2.3","target_commitish":"1111111111111111111111111111111111111111","body":"Release notes","assets":[]}]]'
populated_draft='[[{"id":123,"draft":true,"tag_name":"v1.2.3","name":"v1.2.3","target_commitish":"1111111111111111111111111111111111111111","body":"Release notes","assets":[{"name":"provider.zip"}]}]]'
published_release='[[{"id":123,"draft":false,"tag_name":"v1.2.3","name":"v1.2.3","target_commitish":"1111111111111111111111111111111111111111","body":"Release notes","assets":[{"name":"provider.zip"}]}]]'
wrong_candidate='[[{"id":123,"draft":true,"tag_name":"v1.2.3","name":"v1.2.3","target_commitish":"2222222222222222222222222222222222222222","body":"Release notes","assets":[{"name":"provider.zip"}]}]]'
wrong_tag='[[{"id":123,"draft":true,"tag_name":"v9.9.9","name":"v1.2.3","target_commitish":"1111111111111111111111111111111111111111","body":"Release notes","assets":[]}]]'
replacement_draft='[[{"id":456,"draft":true,"tag_name":"v1.2.3","name":"v1.2.3","target_commitish":"1111111111111111111111111111111111111111","body":"Release notes","assets":[{"name":"provider.zip"}]}]]'
duplicate_title='[[{"id":123,"draft":true,"tag_name":"v1.2.3","name":"v1.2.3","target_commitish":"1111111111111111111111111111111111111111","body":"Release notes","assets":[]},{"id":456,"draft":true,"tag_name":"v9.9.9","name":"v1.2.3","target_commitish":"1111111111111111111111111111111111111111","body":"Other notes","assets":[]}]]'

expect_success() {
  local description="$1"
  local response="$2"
  shift 2

  if ! PATH="${mock_bin}:${PATH}" MOCK_GH_RESPONSE="${response}" \
    "${verifier}" example/repository v1.2.3 "${candidate_sha}" "$@" >/dev/null; then
    echo "expected success: ${description}" >&2
    exit 1
  fi
}

expect_failure() {
  local description="$1"
  local response="$2"
  shift 2

  if PATH="${mock_bin}:${PATH}" MOCK_GH_RESPONSE="${response}" \
    "${verifier}" example/repository v1.2.3 "${candidate_sha}" "$@" >/dev/null 2>&1; then
    echo "expected failure: ${description}" >&2
    exit 1
  fi
}

GITHUB_OUTPUT="${mock_bin}/github-output" \
  expect_success "an empty pre-upload draft" "${empty_draft}"
if [[ "$(<"${mock_bin}/github-output")" != "release_id=123" ]]; then
  echo "expected the verified release ID as a workflow output" >&2
  exit 1
fi
expect_success "the original populated pre-publication draft" "${populated_draft}" present 123 dist/provider.zip
expect_failure "assets before the validated upload" "${populated_draft}"
expect_failure "no expected asset set" "${populated_draft}" present
expect_failure "no assets after GoReleaser" "${empty_draft}" present 123 dist/provider.zip
expect_failure "an incomplete remote asset set" "${populated_draft}" present 123 dist/provider.zip dist/checksums.txt
expect_failure "a release that was published early" "${published_release}" present 123 dist/provider.zip
expect_failure "a draft for another candidate" "${wrong_candidate}" present 123 dist/provider.zip
expect_failure "a draft for another tag" "${wrong_tag}"
expect_failure "a replacement draft" "${replacement_draft}" present 123 dist/provider.zip
expect_failure "ambiguous drafts with the GoReleaser lookup title" "${duplicate_title}"
expect_failure "an unsupported asset-state assertion" "${empty_draft}" unsupported

echo "verify-release-draft tests passed"
