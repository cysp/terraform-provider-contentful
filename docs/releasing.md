# Provider releases

GoReleaser is pinned to `v2.18.0`. A release starts as a GitHub draft and is
published only after its artifacts have been signed and attested.

## Publishing a release

1. In GitHub, draft a new release named for the next `vMAJOR.MINOR.PATCH` tag,
   target `main`, write the release notes, and save it as a draft. Do not click
   **Publish release**.
2. Update local `main`, create that exact tag at the intended commit, and push
   the tag:

   ```bash
   git switch main
   git pull --ff-only
   git tag v0.0.63
   git push origin v0.0.63
   ```

3. The tag starts the release workflow. The protected `release` environment
   supplies the GPG signing key. GoReleaser builds once, signs the checksum file,
   and uploads the artifacts to the existing draft.
4. The workflow verifies the checksums and GPG signature, creates GitHub
   build-provenance attestations, and checks the uploaded asset digests. Its
   final step publishes that same draft.

The separate GoReleaser configuration workflow is read-only and has no release
secrets. Signing and publication run only in the protected `release`
environment, using its release-specific secrets. Releases are serialized so
two tags or reruns cannot publish concurrently.

If the workflow fails, the release remains a non-public draft. Fix the cause and
rerun the failed workflow. GoReleaser replaces assets from the previous attempt;
the workflow rechecks every digest before publishing. Do not publish the draft
manually after a failed run.

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
