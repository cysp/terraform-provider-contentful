# Provider releases

GoReleaser is pinned to `v2.18.0`. Publishing a GitHub release creates its tag
and starts the release workflow. GoReleaser signs and uploads the provider
artifacts to that existing release.

## Publishing a release

1. In GitHub, create a release named for the next `vMAJOR.MINOR.PATCH` tag,
   choose to create that same tag on publish, target `main`, write the release
   notes, and click **Publish release**.
2. Publishing the release creates the tag and starts the release workflow. The
   workflow requires an existing published release whose name and tag both
   match that version.
3. The protected `release` environment supplies the GPG signing key. GoReleaser
   builds once, signs the checksum file, and uploads the artifacts directly to
   the existing release.
4. The workflow verifies the checksums and GPG signature, creates GitHub
   build-provenance attestations, and checks every uploaded asset digest.

The separate GoReleaser configuration workflow is read-only and has no release
secrets. Signing and artifact upload run only in the protected `release`
environment, using its release-specific secrets. Release jobs are serialized
so two tags or reruns cannot upload concurrently.

The release is public while the workflow runs. If the workflow fails, it may
remain incomplete until the failed workflow is rerun successfully. GoReleaser
replaces assets from the previous attempt, and the workflow rechecks every
digest.

## Verifying a published release

Download every asset, then verify its checksum and the detached GPG signature
using the maintainer public key obtained through a trusted channel:

```bash
gh release download v0.0.63 --repo cysp/terraform-provider-contentful
shasum -a 256 --check terraform-provider-contentful_0.0.63_SHA256SUMS
gpg --verify terraform-provider-contentful_0.0.63_SHA256SUMS.sig \
  terraform-provider-contentful_0.0.63_SHA256SUMS
```

Verify GitHub provenance for each downloaded ZIP, manifest, checksum file, and
signature:

```bash
for ASSET in *.zip *_manifest.json *_SHA256SUMS *_SHA256SUMS.sig; do
  gh attestation verify "$ASSET" \
    --repo cysp/terraform-provider-contentful \
    --signer-workflow cysp/terraform-provider-contentful/.github/workflows/release.yml
done
```
