# Provider releases

The [release workflow](../.github/workflows/release.yml) runs on pushes of `v*`
tags. It requires an existing GitHub release that is neither a draft nor a
prerelease, with both its name and tag matching the pushed tag.

The workflow uses the `release` environment and serializes release jobs.
[GoReleaser](../.goreleaser.yml) builds the provider archives, signs the checksum
file, and uploads the archives, Registry manifest, checksums, and signature to
the existing release. Its version is pinned in the workflow.

The workflow verifies checksums and the GPG signature, creates build-provenance
attestations, and verifies the uploaded asset count and digests. A published
release can therefore remain incomplete if the workflow fails. Reruns replace
existing artifacts and repeat these checks.
