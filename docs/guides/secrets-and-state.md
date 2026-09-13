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

Contentful returns only a redacted representation. Refresh preserves the complete `value` last applied by Terraform and cannot detect external rotation. To rotate the secret, change `value` and apply. Contentful's [key rotation workflow](https://www.contentful.com/developers/docs/extensibility/webhooks/request-verification/#key-rotation) describes how receivers accept the old and new secrets during rotation. Changing only `timeouts` leaves the remote secret unchanged.

### Uncertain writes

A failed request can leave the remote secret changed. The provider does not retry the mutation, and a read cannot confirm the complete value. A later apply can overwrite or delete intervening changes.

After an uncertain create, [importing without rotation](#importing-signing-secrets) can adopt an existing secret without another write. This confirms its presence, not which value Contentful stored.

## App signing secrets

Whenever [`contentful_app_signing_secret`](../resources/app_signing_secret) writes a secret to Contentful, it stores the complete configured `value` in Terraform state. Contentful returns only a redacted representation during reads, so refresh preserves the stored value and cannot detect a secret rotated outside Terraform.

To rotate the secret through Terraform, change `value` and apply. Changing only `timeouts` leaves the remote secret unchanged, including a secret rotated outside Terraform.

## Importing signing secrets

Importing a webhook or app signing secret leaves `value` null because Contentful does not return the complete secret. Applying the configured value then replaces the remote secret, including during a configuration-driven import. This lets Terraform take over rotation using a value you supply.

To leave the existing value managed outside Terraform, you can set `ignore_changes = [value]`. The imported value remains null and Terraform does not write the configured value while updating that resource. Reads and deletion require only the resource identifiers and the provider's management API credentials. Changing only `timeouts` sends no mutation, so none of these operations needs the unreadable secret. Creating or rotating a secret requires the complete value to install.

This setting suppresses configuration-driven rotation; it does not recover or verify the remote value. External rotation is already undetectable without it. `value` is still required in configuration, and Terraform uses it for creation or replacement, including recreation after remote deletion. [Terraform's lifecycle reference](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#ignore_changes) explains the distinction between creation and updates. Destroy still deletes the secret. Remove the setting when ready to rotate through Terraform.

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
