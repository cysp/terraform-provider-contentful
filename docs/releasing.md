# Release the provider

## Publish a GitHub release

1. Merge ordinary changes to `main`. When ready to prepare a release, run
   [Prepare release PR](../.github/workflows/prepare-release-pr.yml) on `main`:

   ```sh
   gh workflow run prepare-release-pr.yml --repo cysp/terraform-provider-contentful --ref main
   ```

   Release Please creates or refreshes the release PR from the current `main`.
   Ordinary merges do not create or update release PRs. If `main` advances before
   merging the release PR, run Prepare release PR again so its notes include those
   changes. This workflow does not publish a version; the release PR runs normal CI.
2. Review the release PR's version, changes, `CHANGELOG.md`, and notes. Add any
   compatibility or migration guidance. Before v1.0.0, marked breaking changes
   increment the minor version. Other releases, including compatible features,
   increment the patch version. Maintenance and dependency changes appear in the
   notes. Use a
   [Release-As override](https://github.com/googleapis/release-please/blob/v17.6.0/README.md#how-do-i-change-the-version-number)
   to select another version. Publish or abandon any earlier unpublished version
   first.
3. Merge the release PR to approve publication. The [CI workflow](../.github/workflows/ci.yml)
   checks that commit, then Release Please creates the version tag and a draft
   GitHub release and dispatches the [publisher](../.github/workflows/release.yml)
   on the tag.
4. After CI succeeds, watch the separate `Publish release` run. GoReleaser builds
   and uploads the artifacts to the draft. The workflow verifies the signed checksums
   against the local files, attests build provenance, and publishes the GitHub release.
   It trusts successful uploads. Remote byte comparisons and GitHub release
   immutability are outside its checks.
5. Before announcing the release, confirm its version and rendered documentation
   in the [Terraform Registry](https://registry.terraform.io/providers/cysp/contentful).

Only release PRs merged into `main` can publish, including hotfixes. The `ci` gate checks the merged commit before Release Please creates its tag and draft. The publisher runs against that tag and preserves the approved notes. Later merges can proceed while publication runs.

## Recover a failed release

| Failure | Recovery |
| --- | --- |
| Release PR preparation | Inspect any existing release PR, then rerun Prepare release PR on `main`. Review the refreshed version and notes. |
| CI | Fix the cause and rerun failed jobs on the original CI run. Release creation remains downstream of its checks. |
| Release creation | Inspect the tag and draft. Rerun the original release creation job if the draft has not been created; otherwise follow the manual recovery below. |
| Build, signing, upload, or attestation | Fix the cause and rerun failed jobs on the tag-context publisher. Draft assets are rebuilt and verified again. |
| Dispatch failed or its response was lost | Check for an existing publisher run before dispatching manually. Rerunning release creation may not dispatch again after Release Please changes the PR's labels. |
| GitHub publication failed or its response was lost | Inspect the release: retry only if still a draft. If published, check the Registry instead of rebuilding. |
| Registry indexing | Investigate ingestion separately; preserve the published GitHub release. |

Before a manual publisher dispatch, resolve the version tag to its exact
commit and confirm the CI dependencies listed above succeeded for it. Use the
original CI run, or dispatch `ci.yml` on the tag and wait for success. Manual
publisher dispatch does not enforce this prerequisite. Also restore any missing
draft with its name equal to the tag, no prerelease flag, and the approved PR's
notes: GoReleaser can create a draft but cannot restore those notes.

Then dispatch the publisher, substituting the unpublished version's existing tag:

```sh
gh workflow run release.yml --repo cysp/terraform-provider-contentful --ref v0.0.71
```

After confirming GitHub publication, remove `autorelease: pending` from the
release PR if it remains. Release Please can create the draft before failing to
update the PR's labels; a merged PR with that label prevents Release Please from
opening the next release PR.

GitHub rules protect version tags from updates and deletion, including while the release is a draft. Never move or reuse a tag or replace published assets. Once a tag exists,
a source or workflow fix needs a higher version: reruns use the tagged workflow.
Do not rerun legacy release jobs from before this workflow; they replace assets
on published releases. When recovering an older unpublished version, check
GitHub's Latest designation: publication leaves Latest selection unset and can
change it.

To abandon an unpublished version, cancel its active and queued CI and publication
runs, delete only its draft release, remove `autorelease: pending` from the release
PR if present, and record abandonment there.

Keep the tag and choose a higher version. Mark the abandoned version as unpublished
in the next changelog and carry changes and notes forward from the last
published release; Release Please may otherwise use the abandoned tag as its
comparison boundary.

## Setup and maintenance

- [Infrastructure](https://github.com/chikachow/infrastructure), `github/cysp`,
  manages required checks, tag protections, and the GitHub Actions environments:
  main-only `release-preparation` and tag-only `release`. Coordinate required-check
  names with workflow changes.
- Install the chikachow-owned [`cyspbot-releaser`](https://github.com/apps/cyspbot-releaser)
  App on the repository with Contents, Issues, and Pull requests write permissions.
  The [broker deployment](https://github.com/chikachow/github-app-token-broker-deploy)
  owns its credentials and issuance policy; do not store its private key in
  Terraform or Actions. Authorize this repository's `ci.yml` caller for `push`
  and `workflow_dispatch`, and `prepare-release-pr.yml` for `workflow_dispatch`.
  Both must call `release-please.yml` on `main`, in `release-preparation`.
  Deploy that policy before enabling preparation. The
  [token step](../.github/workflows/release-please.yml) specifies
  the endpoint and broker-requested scopes for the installation access token.
- Keep `RELEASE_GPG_PRIVATE_KEY` and `RELEASE_PASSPHRASE` in `release`, and
  [register the signing key](https://developer.hashicorp.com/terraform/registry/providers/publishing#preparing-and-adding-a-signing-key)
  with the Registry. Main and manually dispatched CI require live acceptance
  credentials; see [Tests](../DEVELOPMENT.md#tests).
- When bootstrapping, set the Release Please version manifest,
  `.release-please-manifest.json`, to the last published version. Inspect the first
  release PR and its checks before approving a release.

For GoReleaser configuration changes, run the pinned `goreleaser check` command
from the [configuration check workflow](../.github/workflows/goreleaser-check.yml).
Local checks do not establish real OIDC issuance, signing, attestation delivery,
GitHub publication, or Registry ingestion; verify those during rollout.
