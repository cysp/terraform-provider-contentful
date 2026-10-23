# Release workflow

## Contract

An ordinary merge to `main` may update a Release Please proposal. Merging the
release PR approves its version, source changes, and notes and authorizes CI to
publish automatically. There is no second mandatory publication approval.
Maintainers decide whether the proposal is fresh enough before merging it; the
publisher verifies the final merged commit, including rebase merges.

```text
main changes → Release Please PR → maintainer merges
  → App creates v* tag and draft → checks on that exact main commit
  → GoReleaser builds and signs → verify assets → attest → publish immutable release
  → check Terraform Registry availability
```

- Release Please proposes patch increments, including for feature and breaking
  commits, until the maintainer explicitly changes this policy. An explicit
  `Release-As` override remains a reviewed maintainer decision.
- Notes come from Conventional Commits. All configured maintenance categories,
  including dependency updates, remain visible. Review notes for compatibility
  guidance; classification does not establish compatibility.
- Only `main` is supported. Fixes and hotfixes follow that branch.
- Only one approved, unfinished release is supported. The publisher blocks a
  later draft while an earlier approved draft remains. Maintainers must not
  merge another release PR until the earlier draft is published or abandoned.
- A failed draft can be retried. An abandoned tagged version is consumed: retain
  the tag, remove the draft, and carry the unpublished changes and notes into a
  higher version. Never move or reuse a release tag.
- Once published, the release must be immutable. Reruns report publication and
  Registry availability without rebuilding or replacing artifacts.

## Ownership and identity

Release Please owns proposal/version/changelog generation and draft creation;
GoReleaser owns archives, checksums, signing, and uploading to that draft. The
repository's [release helper](../../.github/scripts/release.mjs) owns approval,
CI evidence, artifact verification, and the publication transition. It uses
Node built-ins; it does not implement semantic versioning or build packaging.
There are no provider runtime or Terraform schema changes.

A dedicated App is installed only on this repository. Its token is available
only in the `main`-restricted `release-preparation` environment. Contents, pull
requests, and issues write permissions support Release Please; administration
read checks immutable-release configuration. Its ruleset bypass permits tag
creation only. Separate rules prevent updating or deleting `v*` tags, including
for that App. The `release` signing environment remains tag-only. The publisher
uses its workflow token, not the App private key.

The tag trigger is independent of Release Please's action outputs. Preparation
waits briefly for tag-before-draft ordering and recovery finds an existing draft
by tag. It never creates a replacement release. The helper checks the actual
remote tag, checkout SHA, main ancestry, exact final merge SHA of a same-repository
Release Please PR, root manifest version, release ID, name, and approved notes.
A later draft does not block its predecessor, avoiding a recovery deadlock.

## Publication evidence

The publisher requires successful `main` push runs at the exact release SHA for
provider tests (including the live acceptance step), lint, generation,
vulnerability checks, and release workflow checks. It rejects skipped jobs and
uses the latest run and current attempt. It also requires Codecov's project
status and CodeQL analyses for Go and Actions at that SHA. GitHub-enriched SARIF establishes which findings
belong to the release commit. Their linked alerts supply severity and maintainer
dismissal state: error-level or high/critical security findings block unless
dismissed. A finding fixed only on newer `main` still blocks this release.
Missing/running checks wait for up to
45 minutes; failed checks stop immediately. Publication rechecks all evidence.

GoReleaser retains the approved notes and uploads only to an existing draft.
The workflow checks the GPG signature and local checksums, exact checksum
coverage of archives and Registry manifest, and exact remote asset names,
counts, sizes, upload states, and SHA-256 digests. It attests those artifacts
before publishing and checks the resulting immutable release and assets again.
A lost publish response is reconciled by reading the same release ID.

GitHub publication and Registry indexing are separate outcomes. Registry delay
must never cause a published release to be rebuilt. Rendered documentation still
requires inspection.

## Why this design

| Alternative | Tradeoff |
| --- | --- |
| Manual tag/release with GoReleaser | Smallest automation, but version and notes approval remain informal and recovery depends on operator discipline. |
| Release Please plus GoReleaser (selected) | Reviewable version/notes/source approval with established upstream tools and a bounded repository-owned publication check. |
| Fully automatic semantic release | Removes the explicit release decision and needs a different version policy. |
| Changelog fragments | More precise per-change editorial input, with contributor overhead that is not required here. |
| Dedicated custom release bot or cyspbot extension | Adds another deployed service to the release path; no current requirement justifies it. |

## Upstream contracts

- Release Please's [manifest configuration](https://github.com/googleapis/release-please/blob/v17.6.0/docs/manifest-releaser.md)
  and [patch strategy](https://github.com/googleapis/release-please/blob/v17.6.0/src/versioning-strategies/always-bump-patch.ts)
  provide the proposal policy. The pinned action bundles that library version.
- Its [release creation](https://github.com/googleapis/release-please/blob/v17.6.0/src/manifest.ts)
  performs API operations before all labels/comments/outputs finish. Recovery
  therefore cannot depend solely on `release_created`.
- GoReleaser's [GitHub implementation](https://github.com/goreleaser/goreleaser/blob/v2.18.3/internal/client/github.go)
  finds an existing draft by name and does not turn a published release back
  into a draft. The preflight helper rejects ambiguous names and published
  mutable releases before invoking it.
- GitHub documents [immutable releases](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases#immutable-releases)
  and [workflow event suppression for workflow tokens](https://docs.github.com/en/actions/how-tos/writing-workflows/choosing-when-your-workflow-runs/triggering-a-workflow).
  The App token permits the tag event to start publication.
- HashiCorp's [provider publication contract](https://developer.hashicorp.com/terraform/registry/providers/publishing)
  requires signed checksums and the Registry manifest. Replacing distributed
  artifacts can produce checksum failures.

See the [runbook](../releasing.md) for setup, approval, recovery, and verification.
