# Provider releases

The [release workflow](../.github/workflows/release.yml) publishes signed provider
artifacts for `v*` tags. A GitHub release is complete only after this workflow
succeeds; publishing the release page alone does not verify its artifacts.

## Publish a release

Before publishing, confirm that the `release` environment has the
`RELEASE_GPG_PRIVATE_KEY` and `RELEASE_PASSPHRASE` secrets required by the
[workflow](../.github/workflows/release.yml). The signing key must be registered
for this provider in the Terraform Registry; see HashiCorp's
[signing-key guidance](https://developer.hashicorp.com/terraform/registry/providers/publishing#preparing-and-adding-a-signing-key).

1. Select the intended commit from `main` and review its checks using the
   [validation scope](../DEVELOPMENT.md#validation-scope).
2. Publish a GitHub release targeting that commit, using a new `v`-prefixed
   semantic version tag, such as `v1.2.3`. Both the release name and tag must
   equal that tag. The release must be neither a draft nor a prerelease when
   the workflow checks it. Include release notes explaining user-visible
   changes, compatibility requirements, and any migration steps; the
   GoReleaser configuration disables automatic changelog generation.
3. Follow the workflow triggered by the tag push. It uses the `release`
   environment and serializes release jobs.
4. Confirm that the workflow completed, including signed-checksum verification,
   provenance attestation, and verification of the uploaded assets. If it failed,
   follow [Recover a failed release](#recover-a-failed-release).
5. Confirm that the version appears in the
   [Terraform Registry](https://registry.terraform.io/providers/cysp/contentful)
   and inspect its documentation. The GitHub workflow does not verify Registry
   indexing or rendering.

The trigger is a tag push. Publishing or editing a release for an already pushed
tag does not itself trigger this workflow; use the existing tag's workflow run
when one is available.

## Artifacts and verification

[GoReleaser](../.goreleaser.yml) builds provider archives and signs the checksum
file. It uploads the archives, Registry manifest, checksums, and signature to
the existing release. Its version is pinned in the workflow; target platforms
and artifact names are defined in the GoReleaser configuration.

The workflow verifies local checksums and the GPG signature, creates
build-provenance attestations, then checks that the uploaded asset count and
SHA-256 digests match the local artifacts. These checks establish the published
GitHub artifacts; they do not check Terraform Registry indexing.

## Recover a failed release

A failed run can leave a published release with missing or incomplete artifacts.
Inspect the failed step before treating that version as ready to use. For a
release-metadata failure, correct the release name, tag, draft, or prerelease
status to match the workflow's requirements. For a build, signing, or upload
failure, resolve the reported cause before rerunning the failed workflow.

Reruns rebuild and replace existing artifacts, then repeat all verification.
Inspect which assets were already uploaded before rerunning: a rebuilt archive
can have a different checksum even at the same commit. If the version has
already been distributed, publish a new version rather than replacing it;
HashiCorp warns that [replacing released artifacts can cause checksum errors](https://developer.hashicorp.com/terraform/registry/providers/publishing#creating-a-github-release).
Keep the tag pointing at the intended release commit; do not move a published
tag to recover a failed release.
