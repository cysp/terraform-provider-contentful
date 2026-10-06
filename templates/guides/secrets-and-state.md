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

The provider retries explicit rate limiting (HTTP 429) with the same request within the operation timeout. Connection errors and server errors (HTTP 5xx) are not automatically retried. A failed request can leave the remote secret changed, and a read cannot confirm the complete value. A later apply can overwrite or delete intervening changes.

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

Use cryptographically generated signing secrets and protect Terraform state, saved plans, and backups. The provider checks the API's 64-character alphabet rule; it does not reject weak, format-valid secrets.

Terraform retains an Argon2 hash for comparison with the last successfully applied secret. This adds protection against guessing, but it is not encryption or a guarantee that weak secrets cannot be recovered. Someone with access to state can test guesses. Refresh cannot detect external rotation. Historical state snapshots can still contain ordinary values saved before migration.

Managing many write-only signing secrets can require substantial memory during both plan and apply. On a memory-constrained runner, reduce Terraform's `-parallelism` for both commands. This also reduces concurrency for other resources in that operation.

## Migrating signing-secret values

To change how Terraform stores a signing secret, replace `value` with `value_wo`, or the reverse, and supply exactly the same secret Terraform last successfully applied. Plan and apply the change. Moving to `value_wo` removes the complete value from new state; moving to `value` stores it again. Neither migration rotates the remote secret. Supplying a different secret rotates it.

Remove any `ignore_changes` rule for these arguments before migrating. An ignore rule can suppress the storage change or retain an ordinary value that conflicts with `value_wo`. Migration does not remove the secret from historical state snapshots.

## Planning and rotating write-only signing secrets

Create a fresh plan when rotating a write-only secret. The secret supplied at apply can differ from the one supplied during planning. If the resource has a planned update, the provider compares the apply-time secret with the last successfully applied value and writes it only when it differs or no comparison value exists. This includes updates caused by a timeout change.

During a write-only timeout update, `created_at` and `updated_at` become unknown in the plan even when the supplied secret is unchanged. Resources that use these timestamps as replacement triggers can be replaced even though the provider sends no secret PUT and the final timestamps remain unchanged. Treat timestamps as observations, not a rotation counter or proof of secret equality.

Applying a saved **no-op** plan does not update the resource. Supplying a different ephemeral secret when applying that plan cannot rotate it.

`ignore_changes = [value_wo]` can suppress an update caused solely by changing the secret. It does not prevent rotation when another change causes an update. Do not use this rule to keep an externally managed secret unchanged during other updates. Ordinary `value` follows Terraform's usual `ignore_changes` behavior.

For example, this ignore rule can suppress an update caused only by changing `var.signing_secret`. Changing `timeouts.read` still schedules an update. If the secret supplied at apply differs from the last successfully applied value, that update rotates the secret:

```terraform
resource "contentful_webhook_signing_secret" "example" {
  space_id = "space-id"
  value_wo = var.signing_secret

  timeouts = {
    read = "1m"
  }

  lifecycle {
    ignore_changes = [value_wo]
  }
}
```

Use the sensitive ephemeral `signing_secret` variable declared above. The same behavior applies to App Signing Secrets.

## Recovering an uncertain signing-secret write

Both signing-secret resources retry explicit rate limiting (HTTP 429) with the same request within the operation timeout. Connection errors and server errors (HTTP 5xx) are not automatically retried, and writes and deletion do not follow redirects. Reads retain their normal retry behavior. Separate applies can still repeat a mutation.

After a failed write, Terraform keeps the previous successfully applied value for comparison, even if Contentful accepted the new secret. Reverting configuration to the previous secret can therefore produce no change while Contentful uses the new one. Refresh cannot resolve this ambiguity. Coordinate recovery with the consumers of the secret before rotating again.

To recover by deliberately installing a known value:

1. Prepare consumers to accept the recovery secret alongside any keys they must continue accepting during the transition. Use a secret different from the last value Terraform successfully applied. You can reuse the unacknowledged replacement if it differs from that value.
2. Supply that secret and create a fresh plan. Review that the signing-secret resource will update, then apply with the intended secret. A saved no-op plan cannot perform this recovery.
3. Verify a new Contentful-generated signature with the intended secret at the consumer, then retire superseded keys according to the consumer's rotation procedure. A redacted suffix or a subsequent no-op Terraform plan alone does not establish complete-secret equality.

Coordinate all writers during recovery. Another actor can rotate the secret after verification.

## Downgrading the provider

Before selecting a provider version that does not support `value_wo`, migrate back to ordinary `value` using the current provider:

1. Replace `value_wo` with `value`, supplying exactly the same secret Terraform last successfully applied. Remove any `ignore_changes` rule that would suppress this migration. Ordinary `value` cannot use an ephemeral variable; remove `ephemeral = true` from that input and keep `sensitive = true`.
2. Plan and apply the storage change. The complete secret is stored in state again; protect state, saved plans, and their backups accordingly.
3. Remove references to attributes unsupported by the target version. For example, v0.0.69 does not expose `created_at` or `updated_at` on signing-secret resources.
4. Select the older provider version, run `terraform init -upgrade` to update the dependency lock file, and review a fresh plan before applying.

A direct downgrade from write-only state can rotate the secret because the older provider cannot compare it with the stored hash. This behavior was verified with v0.0.69. Resolve any [uncertain write](#recovering-an-uncertain-signing-secret-write) before downgrading. Check the target version's documentation for other resource constraints.

## Importing signing secrets

Importing a webhook or app signing secret leaves `value` null because Contentful does not return the complete secret. Applying the configured value then replaces the remote secret, including during a configuration-driven import. This lets Terraform take over rotation using a value you supply.

After import, the provider has no value to compare. Applying `value_wo` therefore writes the configured value.

To leave the existing value managed outside Terraform, you can set `ignore_changes = [value]` for ordinary values. The imported value remains null and Terraform does not write the configured value while updating that resource. Reads and deletion require only the resource identifiers and the provider's management API credentials. Changing only `timeouts` sends no mutation, so none of these operations needs the unreadable secret. Creating or rotating a secret requires the complete value to install.

This setting suppresses configuration-driven rotation; it does not recover or verify the remote value. External rotation is already undetectable without it. Exactly one of `value` and `value_wo` is required in configuration, and Terraform uses it for creation or replacement, including recreation after remote deletion. [Terraform's lifecycle reference](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#ignore_changes) explains the distinction between creation and updates. Destroy still deletes the secret. Remove the setting when ready to rotate through Terraform.

## Recovering invalid signing-secret state

An **Invalid write-only secret private state** diagnostic means Terraform's stored comparison data cannot be used. In-place updates fail before writing the secret. Planning can warn about this problem and still succeed; other resources may change during apply before the signing-secret update fails.

Restore valid Terraform state before updating the resource. Explicit replacement and destruction remain available without repairing the comparison data. Destruction deletes the current remote secret.

For replacement recovery, review the plan and require destruction of the old instance before creation of its replacement (`-/+`). Both resources manage a singleton: one secret per App Definition or space. With `create_before_destroy`, deleting the old instance deletes the newly written secret, even when Terraform reports a successful apply. This lifecycle setting can propagate from dependent resources even when it is absent from the signing-secret configuration. Adjust the relevant lifecycle settings and create a fresh plan; do not apply the recovery replacement until its plan shows destroy before create. Coordinate with consumers for the interval without a secret and the subsequent installation of the configured key.

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
