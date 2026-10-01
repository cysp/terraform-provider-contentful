# Signing secret planning

App and Webhook Signing Secret resources use `created_at` and `updated_at` to
schedule updates for changed write-only inputs. Their values must also allow any
metadata that a possible PUT can return. This note explains those planning rules;
the [value contract](signing-secret-write-only.md) owns comparison, verifier
storage, and acknowledgement.

## Why timestamps can be unknown

Terraform keeps write-only arguments null in Plan and State, and their Config
values can differ between plan and apply. A write-only argument cannot itself
produce a plan difference. These are [Framework constraints](https://developer.hashicorp.com/terraform/plugin/framework/resources/write-only-arguments).

The provider marks both timestamps unknown when a secret write is possible.
This schedules Update without a separate rotation counter, and permits Contentful
to return changed timestamps. Both are necessary: observed replacements changed
`createdAt` as well as `updatedAt`, including a same-value webhook PUT. See the
[metadata evidence](../research/signing-secret-versions.md).

For existing resources, the shared plan modifier follows these rules:

| Effective inputs | Timestamp plan | Reason |
| --- | --- | --- |
| Ordinary `value` unchanged; no configured write-only input | Preserve prior values | A timeout-only update cannot write the ordinary secret |
| Ordinary `value` changed or removed | Unknown | Apply may write a secret or migrate its representation |
| Write-only input matches the comparison value and the rest of Plan equals State | Preserve prior values | No update is needed |
| Write-only input differs, is unknown, or has no comparison value | Unknown | A write may be needed |
| Ordinary Plan value is null and another change requires Update | Unknown | Apply-time Config may supply different write-only bytes, even when planning Config omitted them |

Missing API timestamps become null; they do not disable this mechanism.
Null-to-unknown can schedule Update, and the final timestamps may resolve back
to null. Malformed response timestamps are decoding errors.

A timeout-only update with unchanged `value_wo` still marks both timestamps
unknown because the apply-time secret could differ. A downstream resource using
them in `triggers_replace` can therefore be replaced even when the provider sends
no secret PUT and the final timestamps equal their prior values (including null).
This is an accepted consequence of allowing apply-time write-only inputs;
timestamps are observations, not a rotation counter or proof of secret equality.

[Timestamp tests](../../internal/provider/resource_signing_secret_timestamps_test.go)
and [timeout-update tests](../../internal/provider/resource_signing_secret_timeout_update_test.go)
cover these constraints, including unchanged ordinary values, changed write-only
values, dependent replacement with unchanged write-only bytes, and responses
with or without timestamps. The dependent-replacement test compares persisted
IDs and exact secret PUT sequences, not only the proposed plan.

## Saved plans and ignored inputs

A saved update can receive a different ephemeral secret at apply. Update must
compare that actual input, and its timestamp plan must allow the resulting PUT.
A saved no-op plan never calls Update; changing the ephemeral input when applying
it cannot rotate the secret. [Saved-plan tests](../../internal/provider/resource_signing_secret_saved_plan_test.go)
exercise both cases.

With `ignore_changes = [value_wo]`, Terraform can supply null for that argument
in planning Config and still supply the actual value during Update. The provider
therefore does not treat a null planning input as proof that no write can occur.
The timeout-update tests cover matching and different apply-time inputs under
this ignore rule. Ordinary `value` continues to use the effective Plan.

## Migration and replacement

Removing ordinary `value` while ignoring the new `value_wo` can leave both inputs
null during planning. The ordinary value's removal still schedules Update.
Apply-time Config supplies the write-only input, which is compared against the
prior ordinary value. Timestamps must allow a write if those bytes differ.

Ignoring ordinary `value` during the same migration can instead retain it in
Plan alongside configured `value_wo`. Update rejects this conflict before PUT.
The plan modifier must not reject it: Core can first propose an update using
ignored state, then select replacement and plan Create from the configuration.
The [ignored-migration tests](../../internal/provider/resource_signing_secret_ignored_migration_test.go)
cover both combinations, migration back to ordinary storage, and replacement.
Terraform's [replacement planning](https://github.com/hashicorp/terraform/blob/v1.16.0/internal/terraform/node_resource_abstract_instance.go)
defines this intermediate-plan behavior.

The same distinction applies to [invalid private state](signing-secret-write-only.md#invalid-private-state):
planning must permit replacement and destruction while Update continues to reject
invalid verifier data. Do not turn an Update precondition into an error that
blocks Core from reaching those recovery operations.
