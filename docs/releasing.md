# Provider releases

Merging the Release Please PR approves the version, source changes, and release
notes and authorizes automatic publication. Ordinary code merges update the
proposal; they do not publish a version. The [design](design/release-workflow.md)
records the policy and verification boundaries.

## One-time setup and rollout

The companion configuration in
[chikachow/infrastructure](https://github.com/chikachow/infrastructure), under
`github/cysp`, manages release environments and tag rules. Apply it through the
normal infrastructure review process. It imports the existing `release`
environment and tag creation rule and creates separate tag mutation protection.
Do not recreate the signing environment or remove its existing secrets.

1. Create a dedicated GitHub App and install it only on
   `cysp/terraform-provider-contentful`. Grant repository Contents, Issues, and
   Pull requests **write**, Administration **read**, and the implicit Metadata
   read permission. No webhook or callback is required for this workflow.
2. Configure the infrastructure variable
   `terraform_provider_contentful_release_app` with its numeric `app_id` and
   string `client_id`. Apply the reviewed configuration. The App may bypass tag
   creation restrictions only; `v*` updates and deletion have no bypass.
3. Set `RELEASE_APP_PRIVATE_KEY` directly in the `release-preparation`
   environment using GitHub's secret UI or `gh secret set --env
   release-preparation`. Keep the private key out of Terraform variables/state.
   Infrastructure supplies `RELEASE_APP_CLIENT_ID`. Preparation is restricted to
   the `main` branch, with no additional approval gate.
4. Confirm that the tag-only `release` environment retains
   `RELEASE_GPG_PRIVATE_KEY` and `RELEASE_PASSPHRASE`. The key must be
   [registered with the Terraform Registry](https://developer.hashicorp.com/terraform/registry/providers/publishing#preparing-and-adding-a-signing-key).
   Confirm live acceptance credentials are available on `main` pushes.
5. Merge the provider workflow change, then enable immutable releases in this
   repository's settings. The pinned Terraform GitHub provider cannot manage
   this setting. An administrator can use the official endpoint:

   ```sh
   gh api --method PUT repos/cysp/terraform-provider-contentful/immutable-releases
   gh api repos/cysp/terraform-provider-contentful/immutable-releases
   ```

   Preparation deliberately fails until `.enabled` is true. Enablement does not
   retroactively make older releases immutable. Keep this setting enabled while
   publication is active; the publisher confirms immutability after publishing.
6. Run the Release Please workflow on `main` and inspect its first proposal.
   The bootstrap manifest starts from `0.0.70`; if another release was published
   before rollout, update it to the last published version before enabling the
   workflow. Review the next version and full notes before approving it.

Before rollout is considered complete, verify a real proposal, its App-created
PR checks, tag/draft creation, exact-commit publication checks, artifact signing,
immutable publication, and Registry indexing. Local fixtures cannot prove
installation permissions, environment eligibility, API propagation, or Registry
webhook delivery. Do not create disposable version tags on this provider as a
test: tagged versions are permanent.

## Approve a release

1. Review the Release Please PR on `main`: source changes, `.release-please-manifest.json`,
   `CHANGELOG.md`, and the notes in the PR body. Features, fixes, breaking changes,
   and maintenance all propose a patch increment until the maintainer explicitly
   changes the policy. For an exceptional version, use Release Please's
   [`Release-As` override](https://github.com/googleapis/release-please/blob/v17.6.0/README.md#how-do-i-change-the-version-number)
   and review the resulting proposal.
2. Decide whether the proposal includes the intended changes from `main`.
   Include migration guidance and dependency/tooling changes in the notes.
   Resolve any earlier approved draft before merging this release PR.
3. Merge the release PR using the repository's normal merge method. Release
   Please creates the tag and draft at its final merged commit. The tag starts
   the publisher; the publisher waits for successful checks on that exact commit,
   including live acceptance tests. No second human publication step is required.
4. Follow the `release` run. It builds and signs with GoReleaser, verifies local
   and uploaded artifacts, attests them, and publishes the draft as immutable.
5. Check the separate Registry availability result and inspect the version's
   documentation in the [Terraform Registry](https://registry.terraform.io/providers/cysp/contentful).
   The workflow polls indexing for five minutes but does not inspect rendering.

## Recover a failed release

Inspect the failed step and the existing tag/draft before taking action. Never
move, delete, or reuse a `v*` tag. Never replace published artifacts.

- **Preparation failed before creating a tag:** resolve configuration or token
  permissions, then rerun Release Please on `main`. Check for an existing draft
  before assuming the failed action created nothing.
- **Tag exists, draft is missing:** inspect the Release Please run and merged
  release PR. Retry preparation. If upstream recovery does not recreate the
  draft, a maintainer may create a draft for that existing tag with name equal
  to tag, no prerelease flag, and notes copied exactly from the approved PR.
  The publisher never invents approval or creates a release.
- **Draft exists, but output or trigger was lost:** rerun the tag's publisher, or
  dispatch `release.yml` with the **tag as the workflow ref**:

  ```sh
  gh workflow run release.yml --repo cysp/terraform-provider-contentful --ref v0.0.71
  ```

  Substitute the existing candidate's tag. Dispatching on `main` and merely
  checking out a tag cannot access the tag-only signing environment.
- **CI failed or live tests skipped:** fix the cause and rerun the relevant
  checks on the same release commit, then rerun publication. A code fix requires
  abandoning this version and approving a higher version from `main`. Checks on
  a newer commit do not satisfy this candidate's gate.
- **Build, signing, upload, or attestation failed:** repair the external cause
  and rerun the same tag while its release remains a draft. Draft artifacts may
  be rebuilt and replaced; all verification runs again before publication.
- **Publication response was lost:** inspect the same release ID. If published
  and immutable, the rerun is read-only and only checks Registry availability.
  A published mutable release fails closed; do not rebuild it.
- **Candidate must be abandoned:** first cancel any active/queued publisher for
  that tag, then delete its draft release while retaining its tag. Record the
  abandonment on the merged release PR. Use a higher version. Review and carry
  forward the complete notes and changelog from the last *published* version:
  Release Please can use the abandoned tag as its comparison boundary, so do
  not assume those changes will appear automatically in the next proposal.
- **Registry is delayed or failed:** investigate Registry ingestion separately.
  Preserve the immutable GitHub release; rerunning publication cannot repair
  documentation or webhook delivery by replacing artifacts.

Workflow reruns execute the workflow from their original commit. Old versions
created before this design still have the old publisher, which replaces assets;
**do not rerun those legacy release jobs**. Resolve a distributed-version problem
with a new release.

## Validate workflow changes

Run `node --test .github/scripts/release.test.mjs` and the pinned GoReleaser's
`goreleaser check`. The release-checks workflow runs both on PRs and `main`.
The fixtures cover approval/source mismatch, failed or skipped checks, draft
ordering, partial uploads, changed notes, lost responses, and read-only reruns.
They do not build/sign real provider artifacts or publish a release.

Actionlint 1.7.12 does not yet understand GitHub's `concurrency.queue` key (also
used by the live acceptance job); lint with that specific syntax diagnostic
excluded until the validator supports it. Queueing preserves pending release
runs instead of replacing a pending candidate during overlapping triggers.
