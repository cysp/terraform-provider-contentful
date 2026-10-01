# Write-only signing-secret planning evidence

These experiments tested when Terraform calls Update for write-only signing
secrets, including with `ignore_changes` and saved plans. See
[signing secret values and acknowledgement](signing-secret-write-only.md)
for the provider's lifecycle rules.

## Experiment boundary

On 2026-09-30 a standalone experimental provider was built with the repository's
Terraform Plugin Framework v1.19.0 dependencies and exercised with Terraform
v1.16.0 on darwin/arm64.

The probe exposed optional ordinary `value`, optional sensitive write-only
`value_wo`, computed `created_at` and `updated_at`, a stable ID, and an optional
local-only attribute representing the timeout-only update distinction. It logged
provider lifecycle inputs and the decision to perform a synthetic PUT. It used
synthetic A/B/C values, synthetic timestamp strings or null, and a SHA-256 digest
solely for deterministic equality checks. The production verifier uses Argon2id.
No Contentful credentials or network requests were used.

The assertions checked lifecycle calls and decoded public/private state for the
transitions below. The probe did not exercise the Contentful resources, RFC3339
types, validators, HTTP client, or production verifier.

## Observed lifecycle behavior

| Scenario | Observed result |
| --- | --- |
| Existing `value_wo` A, unchanged configuration | No resource Update |
| `value_wo` A to B | Unknown timestamps scheduled Update; synthetic PUT selected |
| Both prior timestamps null, `value_wo` A to B | Null-to-unknown scheduled Update; final timestamps could resolve back to null |
| Ignored `value_wo` B and a local-only change | Planning Config `value_wo` was null; Update was called with Apply Config `value_wo` B |
| `ignore_changes = all` | Effective `value_wo` Config null; no Update |
| Ignore both computed timestamps while changing `value_wo` | Terraform warned that those ignores were redundant; provider could still mark timestamps unknown and return changed metadata |
| Active `value_wo` unchanged plus local-only change | Timestamps remained unknown; Update compared the actual value and skipped PUT |
| Saved local-only update: ephemeral `value_wo` A during plan, B during apply | Apply-time planning saw B; Update could select PUT without violating timestamp constraints |
| Saved ignored-`value_wo` local-only update: B during plan, C during apply | Planning continued to see null `value_wo`; Update received C |
| Unknown `value_wo` resolving to a changed value | Initial plan scheduled possible Update; resolved value selected PUT |
| Unknown `value_wo` resolving to the existing value, with no other change | Apply-time planning restored prior timestamps; Core cancelled the resource Update |
| Saved no-op plan made with ephemeral A, applied with B | No resource Update during apply; B was not detected |
| Same-byte ordinary A to `value_wo` A | Public plaintext removed and verifier stored without PUT |
| Same-byte `value_wo` A to ordinary A | Plaintext restored and verifier removed without PUT |
| `value_wo` A to ordinary A to ordinary B to `value_wo` A | Both differing-byte transitions selected PUT; obsolete verifier was not reused |
| Missing private verifier, active `value_wo` | PUT selected to establish a baseline |

Write-only values remained null in inspected final public state. These input
observations do not imply that Terraform filters write-only values during Update.
The provider's [contract](signing-secret-write-only.md#ignored-changes) compares
apply-time values whenever Update runs, without tracking ignore rules. The
[timeout-update tests](../../internal/provider/resource_signing_secret_timeout_update_test.go)
verify that changed apply-time bytes can be written even with an ignore rule,
while matching bytes skip PUT.

## Ignored migration

Start with ordinary `value = A`, then remove it, configure `value_wo = B`, and
ignore `value_wo`. Effective planning inputs contain null ordinary and write-only
values, while prior state still contains A. Actual Apply Config contains B.

The ordinary value's removal can schedule Update even though the effective
planning inputs contain no secret. Reading Apply Config then supplies B for
comparison against A. Timestamps must permit the result of that possible write;
preserving known timestamps during planning would constrain the response to the
previous metadata.

The inverse ignore combination also matters: ignoring ordinary `value` can leave
effective ordinary A alongside active `value_wo` B even though original configuration
contains only `value_wo` B. The probe deferred its conflict diagnostic until Update and
performed no PUT. With explicit replacement, Core first presented the same
intermediate conflict, then replanned Create with ordinary value null and `value_wo` B.
Creation succeeded. Rejecting the intermediate plan would block that valid
replacement.

## Selecting the comparison value

The probe deliberately combined `value = B` with an old hash of A, then supplied
`value_wo = A`. Accepting a match against either the ordinary value or the hash
incorrectly skipped PUT because the old hash matched A. Comparing against
`value = B` correctly identifies the change.

This constructed case demonstrates why a known ordinary value must take
precedence over a stored hash. See the
[design contract](signing-secret-write-only.md) for the comparison rules.

## Sources and related evidence

- HashiCorp's [write-only argument contract](https://developer.hashicorp.com/terraform/plugin/framework/resources/write-only-arguments): null planned/state values, apply-time variability, and private hash comparison.
- Terraform v1.16.0 [planning, ignores and apply inputs](https://github.com/hashicorp/terraform/blob/v1.16.0/internal/terraform/node_resource_abstract_instance.go).
- Terraform v1.16.0 [planned-value validation](https://github.com/hashicorp/terraform/blob/v1.16.0/internal/plans/objchange/plan_valid.go): computed-only attributes remain provider-owned even when an ignore rule supplies a value in effective Config.
- [Contentful signing-secret metadata observations](../research/signing-secret-versions.md): absence of usable version metadata and changing creation/update timestamps.
