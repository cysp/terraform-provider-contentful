# Release validation

The repository uses two validation levels. Pull requests and `main` run without
Contentful credentials. A release tag runs the same deterministic checks again,
then enters the `release` environment to run live acceptance and publish the
exact tested commit. The environment and repository rules described below are
required non-code gates and must be configured before this workflow can be used
to publish a release.

## Validation matrix

The Terraform acceptance matrix is:

| Validation | Terraform versions | Contentful service |
| --- | --- | --- |
| Pull request and `main` mocked acceptance | 1.13, 1.14 | In-process test server |
| Release mocked acceptance | 1.13, 1.14 | In-process test server |
| Release live acceptance | 1.14 | Live CMA |

Terraform 1.13 and 1.14 are the repository's explicit mocked-acceptance matrix.
Terraform 1.14 is the live-acceptance line. These versions are not derived from
which Terraform releases are newest, and the matrix is not a claim that other
Terraform versions are incompatible.

Every pull request and `main` revision runs:

- all-package build and unit tests;
- `go test -race ./...`, including the management test server;
- mocked acceptance on both Terraform versions in the matrix;
- lint after clearing the linter cache;
- `go mod tidy` and `go generate ./...`, followed by clean-tree checks; and
- `goreleaser check` with the repository's pinned GoReleaser version.

The release workflow repeats those checks from the tag target rather than
assuming that a branch check still describes the tagged commit.

## Candidate binding and publication

Pushing a `v*` tag starts the release workflow. The workflow peels the event
SHA to a commit, verifies that the checkout is that exact commit, and verifies
that the tag is protected and the commit is reachable from `origin/main`. An
unprotected tag fails before validation. The candidate SHA and tag are written
to the workflow summary. The checkout uses `fetch-depth: 0`, which the
pinned [checkout action](https://github.com/actions/checkout/tree/3d3c42e5aac5ba805825da76410c181273ba90b1#checkout-v4)
defines as fetching all history for all branches and tags, so `origin/main` is
available to the ancestry check.

Publication and live acceptance are steps in one `publish` job. The job starts
only after deterministic validation and the mocked Terraform matrix succeed.
After the configured environment admits the job, it checks that all three
release-only credentials are non-empty, runs live acceptance, and only then
invokes GoReleaser and build-provenance attestation. Any missing credential is
therefore a release failure, never a skipped live suite.

GitHub defines `GITHUB_REF` as the pushed tag for a tag-triggered workflow and
defines `GITHUB_SHA` according to the triggering event. The explicit checkout
comparison makes that event identity visible and testable in the workflow
itself. See GitHub's [variables reference](https://docs.github.com/en/actions/reference/workflows-and-actions/variables#default-environment-variables)
and [events reference](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#push).

## Privilege boundary

Ordinary CI uses the `pull_request` event with read-only repository contents
and does not reference the Contentful credential. It does not use
`pull_request_target` or a privileged `workflow_run` that checks out pull
request content. GitHub warns that combining either privileged trigger with an
untrusted checkout can expose secrets or repository write access; see the
[secure use reference](https://docs.github.com/en/actions/reference/security/secure-use#mitigating-the-risks-of-untrusted-code-checkout).

The release workflow is triggered only by a repository tag push. Its
deterministic jobs have read-only contents permission. Only the final job has
the write permissions needed to create a release and an attestation. GitHub
sets unspecified permissions to `none`, so job-level permissions keep those
capabilities out of earlier jobs; see [workflow permission syntax](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#permissions).

The workflow references the release-only secret names
`RELEASE_CONTENTFUL_MANAGEMENT_ACCESS_TOKEN`, `RELEASE_GPG_PRIVATE_KEY`, and
`RELEASE_PASSPHRASE`. They must exist only on the `release` environment. The
distinct names are intentional: repository-level secrets remain available to
jobs that reference an environment, so reusing repository secret names would
not fail closed while the environment is absent or misconfigured.
GitHub does not make environment secrets available until the environment's
protection rules pass; see [deployments and environments](https://docs.github.com/en/actions/reference/workflows-and-actions/deployments-and-environments#environment-secrets).

A separate manually dispatched privileged workflow is intentionally not used.
GitHub permits a caller to choose the branch or tag receiving a
`workflow_dispatch`, so a tag-bound workflow plus environment protection has a
smaller trigger surface. See the [`workflow_dispatch` event](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#workflow_dispatch).

## Required repository settings

The `release` environment must be configured with:

- `RELEASE_CONTENTFUL_MANAGEMENT_ACCESS_TOKEN`, `RELEASE_GPG_PRIVATE_KEY`, and
  `RELEASE_PASSPHRASE` as environment secrets;
- selected deployment tags matching `v*`;
- a required-reviewer policy appropriate to the repository's release
  maintainers; and
- administrator bypass disabled.

Prevent self-review when an independent trusted reviewer is available. If the
release maintainer is necessarily the sole reviewer, self-review must remain
available and the repository must explicitly accept that the environment gate
does not provide independent approval. In either case, the protected `main`
and `v*` rulesets prevent untrusted pull-request code from reaching the
environment.

GitHub documents these controls in [Managing environments for deployment](https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/manage-environments#creating-an-environment).

An active tag ruleset targeting `v*` is mandatory: restrict tag creation,
updates, and deletion to the release maintainer. The workflow also refuses to
validate a tag when GitHub reports that its ref is unprotected. Protect `main`
and require the ordinary validation workflows before merge. Environment review
remains necessary because the release job intentionally executes the tagged
provider tests while holding the live credential. GitHub's [ruleset reference](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets#restrict-creations)
documents restricting matching tag creation to configured bypass actors.

The release signal is meaningful only after these repository settings are in
place. Until the release-only environment secrets are configured, the explicit
credential preflight fails before live acceptance or signing.

The `main` rulesets must require these stable workflow contexts:

- `build`;
- `race`;
- `generate`;
- `test`;
- `contentful-management-go-test`;
- `testaccmocked (1.13.*)`;
- `testaccmocked (1.14.*)`;
- `lint`; and
- `codecov/project`.

The CodeQL code-scanning rule must also be required. Together these rules make
build, unit, management-client, race, mocked acceptance, generation, lint,
coverage, and CodeQL results mandatory on `main`.

## Toolchain pins

Third-party actions are pinned by commit. GoReleaser is additionally pinned to
`v2.17.1` for both configuration validation and publication. The GoReleaser
action otherwise accepts a floating semantic-version range, so the explicit
version keeps release behavior reproducible; see the action's [version input](https://github.com/goreleaser/goreleaser-action#customizing).
GitHub Actions workflows are statically checked with actionlint `v1.7.12` in
ordinary and release validation.
