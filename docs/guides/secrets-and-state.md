---
page_title: "Secrets and Terraform state"
description: |-
  Understand which Contentful secrets Terraform stores, which import can recover, and how to manage rotation.
---

# Secrets and Terraform state

Terraform's `sensitive` marking redacts normal CLI output but does not keep values out of state or saved plans. Restrict access to both when using resources that manage credentials. See HashiCorp's [sensitive-data guidance](https://developer.hashicorp.com/terraform/language/manage-sensitive-data) for storage and redaction behavior.

Refresh can verify a value only when Contentful returns it. Import starts without a previously stored secret, so it cannot recover values that Contentful withholds. Preserving a stored value does not detect external changes to an unreadable secret.

| Value | Refresh | Import |
| --- | --- | --- |
| [Webhook signing secret](#webhook-signing-secrets) | Preserves stored value | `value` is null |
| [App signing secret](#app-signing-secrets) | Preserves stored value | `value` is null |
| [Webhook Basic password](#basic-authentication) | Preserves managed password | `http_basic_password` is null |
| [Secret webhook header](#secret-custom-headers) | Preserves managed value for matching header | Unreadable value is null |
| [App Installation `Secret` parameter](#app-and-extension-parameters), read with a personal access token | Records redacted response | Original value unavailable |
| [Personal Access Token](#personal-access-tokens) | Preserves token from creation | `token` is null |
| Delivery API access token | Reads token | Reads token |
| [App Key public JWK](#app-keys) | Reads public key | Reads public key; private key unavailable |

## Webhook signing secrets

[`contentful_webhook_signing_secret`](../resources/webhook_signing_secret) manages the signing secret for all webhooks in a space. Creating the resource replaces an existing secret; destroying it deletes the current secret, including one rotated outside Terraform. A space has at most one signing secret. Multiple Terraform resources can target it and overwrite or delete one another's value. A same-space replacement with `create_before_destroy` can delete the newly written secret.

Contentful returns only a redacted representation. Refresh preserves the complete `value` last applied by Terraform and cannot detect external rotation. To rotate the secret, change `value` and apply. Contentful's [key rotation workflow](https://www.contentful.com/developers/docs/extensibility/webhooks/request-verification/#key-rotation) describes how receivers accept the old and new secrets during rotation. With ordinary `value`, changing only `timeouts` leaves the remote secret unchanged.

A successful rotation response does not mean every webhook signer immediately uses the new secret. Keep receivers accepting both keys during the transition and confirm new-key signatures before retiring the old key. Do not treat a fixed waiting period or a matching redacted suffix as verification of that transition.

### Uncertain writes

A failed request can leave the remote secret changed. The provider does not retry the mutation, and a read cannot confirm the complete value. A later apply can overwrite or delete intervening changes.

After an uncertain create, [importing without rotation](#importing-signing-secrets) can adopt an existing secret without another write. This confirms its presence, not which value Contentful stored.

## App signing secrets

[`contentful_app_signing_secret`](../resources/app_signing_secret) stores the complete ordinary `value` in Terraform state. Contentful returns only a redacted representation during reads, so refresh preserves the stored value and cannot detect a secret rotated outside Terraform.

To rotate the secret through Terraform, change `value` and apply. With ordinary `value`, changing only `timeouts` leaves the remote secret unchanged, including a secret rotated outside Terraform.

## Write-only signing secrets

Both signing-secret resources accept exactly one of `value` or `value_wo`. Use `value_wo` with Terraform 1.11.1 or later to keep the complete value out of this resource's plan and state. Use an ephemeral input variable so Terraform does not save that input in plans or state:

```terraform
variable "signing_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

resource "contentful_webhook_signing_secret" "example" {
  space_id = "space-id"
  value_wo = var.signing_secret
}
```

The provider stores a salted hash in Terraform state to compare the configured value with the last successfully written value. Contentful does not return the complete secret, so the hash cannot detect external rotation. Protect state and historical snapshots, which can still contain ordinary values saved before migration.

Each hash calculation uses 64 MiB of working memory. Concurrent comparisons and rotations can use substantially more memory across a provider process. When managing many signing secrets on a memory-constrained runner, reduce Terraform's `-parallelism` for both planning and applying. This also reduces concurrency for other resources in that operation.

## Migrating signing-secret values

Changing from `value` to `value_wo` with exactly the same secret removes the ordinary value from new state without a remote write. Changing back with the same secret restores it to state without a remote write. A different secret replaces the remote value.

## Planning and rotating write-only signing secrets

When applying a planned update, the provider compares the supplied write-only value with the last successfully written value. This also applies to an update caused by a timeout change: a different ephemeral value at apply can rotate the secret. Applying a saved **no-op** plan does not update the resource, so changing an ephemeral input when applying that plan cannot rotate the secret. Create a fresh plan for rotation.

`ignore_changes = [value_wo]` does not prevent rotation during an update caused by another change, such as a timeout change. Terraform can exclude the secret from planning while still supplying it during apply. Whenever Update runs, the provider compares that supplied value normally: matching bytes cause no secret write; different bytes, or a missing comparison value after import, cause a write. Do not use this setting to keep an externally managed secret unchanged during other updates. Ordinary `value` continues to follow Terraform's effective plan, including `ignore_changes`.

## Recovering an uncertain signing-secret write

A failed or ambiguous write retains the previous successfully applied value for comparison. If the remote service accepted new bytes but its acknowledgement was lost, reverting configuration to the previous bytes can be a no-op even though the remote value differs. Refresh cannot resolve this ambiguity. Coordinate recovery with the consumers of the secret before making another deliberate rotation.

To recover by deliberately installing a known value:

1. Prepare consumers to accept the recovery secret alongside any keys they must continue accepting during the transition. Use a secret different from the last value Terraform successfully applied. You can reuse the unacknowledged replacement if it differs from that value.
2. Supply that secret and create a fresh plan. Review that the signing-secret resource will update, then apply with the intended secret. A saved no-op plan cannot perform this recovery.
3. Verify a new Contentful-generated signature with the intended secret at the consumer, then retire superseded keys according to the consumer's rotation procedure. A redacted suffix or a subsequent no-op Terraform plan alone does not establish complete-secret equality.

Coordinate all writers during recovery. Another actor can rotate the secret after verification.

## Downgrading the provider

Provider versions without `value_wo`, including v0.0.69, require ordinary `value`.
Remove configuration references to these resources' `created_at` and `updated_at`
before selecting v0.0.69; that version does not expose either attribute.
To avoid rewriting the remote secret when returning to v0.0.69:

1. Keep the current provider selected. Replace `value_wo` with `value`, supplying exactly the same secret Terraform last successfully applied. Ordinary `value` cannot use an ephemeral variable; remove `ephemeral = true` from that input and keep `sensitive = true`.
2. Plan and apply this representation change with the current provider. The secret is stored in state again; protect state, saved plans, and their backups accordingly.
3. Select v0.0.69, run `terraform init -upgrade` to update the dependency lock file, and review a fresh plan. With the same ordinary value, the signing-secret resource should have no remote change.

Downgrading directly from write-only state and configuring ordinary `value` schedules a PUT with v0.0.69, even if the supplied secret matches the remote secret. That provider cannot use the stored hash to compare values. If a write was unconfirmed, resolve it using the [recovery procedure](#recovering-an-uncertain-signing-secret-write) before downgrading. Other resources can have their own downgrade constraints.

## Importing signing secrets

Importing a webhook or app signing secret leaves `value` null because Contentful does not return the complete secret. Applying the configured value then replaces the remote secret, including during a configuration-driven import. This lets Terraform take over rotation using a value you supply.

After import, the provider has no value to compare. Applying `value_wo` therefore writes the configured value. If the stored hash data is invalid, the provider reports an error without writing a secret.

To leave the existing value managed outside Terraform, you can set `ignore_changes = [value]` for ordinary values. The imported value remains null and Terraform does not write the configured value while updating that resource. Reads and deletion require only the resource identifiers and the provider's management API credentials. Changing only `timeouts` sends no mutation, so none of these operations needs the unreadable secret. Creating or rotating a secret requires the complete value to install.

This setting suppresses configuration-driven rotation; it does not recover or verify the remote value. External rotation is already undetectable without it. Exactly one of `value` and `value_wo` is required in configuration, and Terraform uses it for creation or replacement, including recreation after remote deletion. [Terraform's lifecycle reference](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#ignore_changes) explains the distinction between creation and updates. Destroy still deletes the secret. Remove the setting when ready to rotate through Terraform.

## Webhook credentials and secret headers

### Basic authentication

For [`contentful_webhook`](../resources/webhook), Contentful does not return the HTTP Basic authentication password. Refresh preserves a previously managed `http_basic_password` but cannot detect changes made outside Terraform. Import leaves the password null and does not change the remote credentials.

On a later update, omitting both `http_basic_username` and `http_basic_password` clears Basic authentication, including on an imported webhook. `ignore_changes` can retain previously managed credentials, but it cannot recover an unreadable imported password. To set credentials after import, configure both the username and password.

### Secret custom headers

Contentful's `secret = true` flag on a custom webhook header controls how Contentful treats the value. Terraform sensitivity is configured separately: supply a sensitive Terraform expression to redact the header value from normal Terraform output. The value can still be present in plan or state data.

Contentful omits secret custom-header values from responses. Refresh preserves a previously managed value for the matching header but cannot verify it or detect an external replacement. Import leaves an unreadable secret-header value null.

After import, choose how Terraform should manage headers:

- Omit `headers` to preserve the imported headers on later updates. The provider sends unreadable secret headers without values, which lets Contentful retain their secrets.
- Configure `headers` with values for every header you want to manage.
- Set `headers = {}` to clear all custom headers.

## App and extension parameters

The `parameters` values on [`contentful_app_installation`](../resources/app_installation) and [`contentful_extension`](../resources/extension) can contain both ordinary configuration and secrets. Supply secrets through sensitive Terraform expressions. Defining an App Definition's installation parameter as type `Secret` does not mark its Terraform value sensitive.

Parameter omission has different effects for these resources. Follow each resource's `parameters` description before importing an existing object or removing configured values.

Contentful redacts App Installation parameters declared as `Secret` when read through the Content Management API with a personal access token. The provider records the returned parameter object on refresh and import, without preserving the original secret values. Import cannot recover those values, and redacted values returned during refresh can cause Terraform to plan further parameter updates. See [Contentful's Secret installation parameter documentation](https://www.contentful.com/developers/docs/extensibility/app-framework/app-parameters/#secret-installation-parameters).

Omission and replacement of `Secret` values have not been verified against Contentful. Terraform sensitivity controls output redaction; it does not resolve these read and update limitations.

## Personal access tokens

[`contentful_personal_access_token`](../resources/personal_access_token) stores the secret `token` returned at creation. Contentful does not return that value again, so refresh preserves the known value and import leaves it null.

Changing `name`, `scopes`, or `expires_in` replaces the token and revokes the old one. Review dependent consumers before applying a replacement. Changing only `timeouts` preserves the existing token; destroying the resource revokes it.

## App keys

[`contentful_app_key`](../resources/app_key) manages public JWK material that you supply. The corresponding private key is not sent to Contentful and is not stored by this resource.

Unlike the redacted secrets above, the public JWK is readable: import and refresh populate it from Contentful. Changing configured JWK material replaces the App Key rather than updating it in place; importing the public key cannot recover its corresponding private key.
