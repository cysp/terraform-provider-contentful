# Signing secret values and acknowledgement

App and webhook signing-secret resources require exactly one sensitive argument:
`value` or write-only `value_wo`. Ordinary values come from the effective Plan;
write-only values come from Config because their Plan and State values are null.
Neither endpoint returns the complete secret. The reviewed evidence establishes
no usable optimistic-lock version for either endpoint. See the
[API evidence](../research/signing-secret-versions.md) and
[Terraform lifecycle evidence](signing-secret-planning.md).

## Terraform compatibility

Write-only signing secrets require Terraform 1.11.1 or later. Terraform 1.11.0
cannot reliably serialize sensitive ephemeral inputs; HashiCorp fixed that Core
bug in [1.11.1](https://github.com/hashicorp/terraform/pull/36619).

Read returns the endpoint identity even when a 404 removes resource state.
Terraform versions before 1.12 do not supply prior resource identity, so these
resources reconstruct it from their existing scope attributes before removal.
The returned resource state is null.

A provider version that predates `value_wo` cannot use its private verifier to
compare secrets. Before downgrading, use the current provider to migrate to
ordinary `value` with exactly the same secret. This restores the value to state
without a PUT. The [practitioner guide](../guides/secrets-and-state.md#downgrading-the-provider)
describes the sequence and its storage consequences. Compatibility experiments
with v0.0.69 verified migration with and without an initial refresh and downgrade
with and without restoring ordinary `value`. Ongoing tests use the locally built
provider; they do not download or execute released provider binaries.

## Comparison and representation

Compare against a known non-null prior `value`, otherwise against the `value_wo`
verifier. An unknown prior value is an error. If prior `value` is null and the
verifier is missing, an update that permits a write must send PUT. Invalid verifier
data stops Update before mutation; Read and Delete do not inspect it.
Create always writes, including recreation after remote deletion.

Equal bytes can still require representation changes: ordinary to write-only
removes plaintext and prepares a verifier; write-only to ordinary publishes
plaintext and removes the verifier. Neither migration needs a PUT. Clearing the
verifier when ordinary value becomes active prevents an obsolete verifier from
suppressing a later write. Refresh preserves the acknowledged baseline and does
not infer equality from redacted text or timestamps. Delete needs no verifier.

Private key `write_only_secret_hashes` contains a JSON array of `{path, hash}`
records. The shared helper updates or removes each path independently,
preserving other verifiers. Resources validate path ownership separately:
signing-secret resources accept only `value_wo`, while a resource with dynamic
paths can load historical records before removing paths absent from its current
configuration. Each hash uses PHC encoding with Argon2id v19,
65536 KiB memory, one iteration, four lanes, a random 16-byte salt and 32-byte
output. Input is UTF-8 path, NUL, then the exact value. Before comparing stored
data, validate its fixed Argon2 parameters, salt and key encoding and lengths,
non-empty unique paths, and 1 MiB size limit. Unknown JSON fields and unsupported
parameters are rejected. Future format changes require an explicit compatibility
policy. Verifier diagnostics contain fixed failure categories
without private data. Update rejects invalid verifier data even when a known
ordinary value would suffice for comparison. Protect Terraform state: the stored
hash can be used to test guesses of the secret without contacting Contentful.

Keep the existing hash when the write-only secret is unchanged. Generate a new
hash with a fresh salt when it changes. If generating the hash fails, keep the
previous hash. Argon2 cannot be interrupted once running; cancellation is checked
before and after the calculation.

The memory parameter applies to each hash calculation, not to the whole provider.
A rotation can perform both a comparison and a new hash, and concurrent resource
operations and garbage-collection timing increase peak resident memory. Terraform
parallelism controls this concurrency; it does not change the stored hash format.

## Mutation boundary

When writing a secret:

1. Prepare the hash changes before sending PUT.
2. Check that PUT completed without a transport error and that its response
   decodes successfully and identifies the requested resource.
3. Encode the new Terraform state and resource identity.
4. Save the hash changes, then put the new state and identity in the response to
   Terraform.

If the write cannot be confirmed, retain the previous state and hash. Contentful
might nevertheless have accepted the new secret. Reverting configuration to the
previous secret can then produce no change in Terraform, leaving the new secret
in Contentful.

Because `value_wo` is absent from plan and state, the provider must signal its
changes through another attribute. When the configured secret differs from the
previous value, or cannot yet be compared, the provider marks `created_at` and
`updated_at` unknown so Terraform schedules Update.

If another change already requires Update, the provider also leaves these
timestamps unknown, because apply-time Config can supply a changed write-only
secret even when effective planning Config contains null. During Update, the
provider compares the supplied secret with the previous value. It writes only
when they differ or no previous value is available; otherwise, it keeps the
previous timestamps.

Missing response timestamps become null; malformed timestamps cause a decoding
error. Applying a saved no-op plan does not call Update, so supplying a different
ephemeral secret cannot trigger a write. Rotation requires a fresh plan.

The provider does not automatically retry webhook signing-secret PUT requests.
App signing-secret PUT requests follow the
[shared rate-limit retry policy](contentful-http-retry-policy.md).
Errors and logs redact exact matches of the secret being sent and any previous
secret stored in `value`.

## Ignored changes

Ordinary `value` is read from the effective Plan and follows Terraform's
`ignore_changes` behavior. Write-only `value_wo` is read from apply-time Config
whenever Update runs, including when Terraform supplied null for that input
during planning. The provider compares those bytes against the acknowledged
value without tracking whether Terraform ignored them during planning.

Consequently, `ignore_changes = [value_wo]` can suppress an update caused solely
by a secret configuration change, but it does not suppress rotation when another
change causes Update. A matching value skips PUT; a different value or missing
baseline selects PUT. Create and replacement use the configured secret.

## Verification boundaries

Provider behavior is covered by:

- [Mutation acknowledgement tests](../../internal/provider/resource_signing_secret_protocol_test.go):
  request payloads and retention of prior state after unconfirmed writes.
- [Saved-plan tests](../../internal/provider/resource_signing_secret_saved_plan_test.go):
  apply-time ephemeral inputs and saved no-op plans.
- [Timeout-update tests](../../internal/provider/resource_signing_secret_timeout_update_test.go):
  ignored write-only inputs that match or differ during saved updates, with
  present or absent API timestamps, and ignored inputs in saved no-op plans.
- [Recovery tests](../../internal/provider/resource_signing_secret_recovery_test.go):
  a committed write with a lost response, retention of the prior Terraform
  baseline, and deliberate recovery through a fresh apply.
- [Lifecycle tests](../../internal/provider/resource_signing_secret_write_only_test.go):
  ordinary/write-only representation changes and exact write sequences using
  the locally built provider.
- [Verifier tests](../../internal/provider/write_only_secret_internal_test.go):
  format compatibility, path handling, and preparation/publication failures.

The [endpoint experiments](../research/signing-secret-versions.md)
record observed Contentful behavior.
