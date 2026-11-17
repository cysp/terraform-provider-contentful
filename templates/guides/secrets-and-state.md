---
page_title: "Secrets and Terraform state"
description: |-
  Understand which Contentful secrets Terraform stores, which import can recover, and how to manage rotation.
---

# Secrets and Terraform state

Use this guide when importing or rotating Contentful credentials. For Terraform
storage and redaction, see HashiCorp's [sensitive-data guidance](https://developer.hashicorp.com/terraform/language/manage-sensitive-data).

Refresh can verify a value only when Contentful returns it. Import starts without a previously stored secret, so it cannot recover values that Contentful withholds. Preserving a stored value does not detect external changes to an unreadable secret.

| Value | Refresh | Import |
| --- | --- | --- |
| [Webhook signing secret](#webhook-signing-secrets) | Preserves stored value | `value` is null |
| [App signing secret](#app-signing-secrets) | Preserves stored value | `value` is null |
| [Webhook Basic password](#basic-authentication) | Preserves managed password | `http_basic_password` is null |
| [Secret webhook header](#secret-custom-headers) | Preserves managed value for matching header | Unreadable value is null |
| [App Installation `Secret` parameter](#app-installation-secrets), read with a personal access token | Can retain prior state for an all-asterisk response | Retains mask; original value unavailable |
| [Personal Access Token](#personal-access-tokens) | Preserves token from creation | `token` is null |
| Delivery API access token | Reads token | Reads token |
| [App Key public JWK](#app-keys) | Reads public key | Reads public key; private key unavailable |

## Webhook signing secrets

[`contentful_webhook_signing_secret`](../resources/webhook_signing_secret) manages the signing secret for all webhooks in a space. Creating the resource replaces an existing secret; destroying it deletes the current secret, including one rotated outside Terraform. A space has at most one signing secret. Multiple Terraform resources can target it and overwrite or delete one another's value. A same-space replacement with `create_before_destroy` can delete the newly written secret.

Contentful returns only a redacted representation. Refresh preserves the complete `value` last applied by Terraform and cannot detect external rotation. To rotate the secret, change `value` and apply. Contentful's [key rotation workflow](https://www.contentful.com/developers/docs/extensibility/webhooks/request-verification/#key-rotation) describes how receivers accept the old and new secrets during rotation. Changing only `timeouts` leaves the remote secret unchanged.

### Uncertain writes

The provider retries explicit rate limiting (HTTP 429) with the same request within the operation timeout. Connection errors and server errors (HTTP 5xx) are not automatically retried. A failed request can leave the remote secret changed, and a read cannot confirm the complete value. A later apply can overwrite or delete intervening changes.

After an uncertain create, [importing without rotation](#importing-signing-secrets) can adopt an existing secret without another write. This confirms its presence, not which value Contentful stored.

## App signing secrets

[`contentful_app_signing_secret`](../resources/app_signing_secret) manages an App Definition's signing secret. Change `value` to rotate it. Refresh preserves the last applied value because Contentful returns only a redacted representation. Changing only `timeouts` leaves the remote secret unchanged.

## Importing signing secrets

Import leaves `value` null. Applying a configured value replaces the remote
secret, including during a configuration-driven import.

To adopt a secret while leaving rotation outside Terraform, set
`ignore_changes = [value]`. Reads, timeout-only updates, and deletion need no
complete secret; the imported value stays null. This does not verify the stored
secret, and destroy still deletes it. `value` remains required in configuration
and is used for creation or replacement, including recreation after remote deletion. Remove `ignore_changes`
when ready to rotate through Terraform.

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

## App Installation secrets

Terraform sends the complete planned `parameters` object whenever it writes an App Installation. Include every parameter you want to retain, including secrets. Refresh cannot reliably detect secrets rotated outside Terraform. A later parameter or Marketplace update can overwrite such a rotation.

### Configuring and rotating secrets

Use `jsonencode(...)` to configure the parameter object and supply secrets through sensitive Terraform expressions, as shown in the [App Installation example](../resources/app_installation#example-usage). To rotate a secret through Terraform, change its configured value and apply.

Contentful redacts App Installation parameters declared as `Secret` when read through the Content Management API with a personal access token. See [Contentful's Secret installation parameter documentation](https://www.contentful.com/developers/docs/extensibility/app-framework/app-parameters/#secret-installation-parameters).

When a top-level parameter is returned as asterisks, the provider can retain its planned value after a successful write or its prior state value during refresh. A retained value does not confirm the current remote secret. Mask length does not establish equality, and even an ordinary string changed remotely to literal asterisks can be hidden. Missing keys and values inside nested objects or arrays are not restored.

Changing only `timeouts` does not write the installation. Changing `marketplace` acknowledgments still writes the complete planned parameter object.

A failed apply does not imply that Contentful rolled back the write. A readable response that contradicts the plan can cause an inconsistent-result error after Contentful accepted the request; a redacted response cannot confirm which secret was stored after an uncertain write.

### Importing an installation

Import retains returned masks and cannot recover the original Secret values. Before applying, configure the complete object, including the Secret values Terraform should manage. Do not copy masks into configuration as if they were recovered values.

On a successful write, omitted keys are removed, including secrets. An empty object clears optional values; Contentful can reject an object that omits a required parameter. Omitting the entire `parameters` object can clear existing values, including after import.

Sending the exact current asterisk mask can preserve the existing secret. A stale or different-length mask can instead become a literal secret. This also means that requesting a literal all-asterisk value equal to the current mask can leave the old secret in place.

### Changes outside Terraform

`ignore_changes = [parameters]` applies to the entire object. A later write, including a `marketplace` change, still sends the retained plaintext or imported masks. It cannot selectively preserve externally rotated secrets while updating ordinary parameters.

## Personal access tokens

[`contentful_personal_access_token`](../resources/personal_access_token) stores the secret `token` returned at creation. Contentful does not return that value again, so refresh preserves the known value and import leaves it null.

Changing `name`, `scopes`, or `expires_in` replaces the token and revokes the old one. Review dependent consumers before applying a replacement. Changing only `timeouts` preserves the existing token; destroying the resource revokes it.

## App keys

[`contentful_app_key`](../resources/app_key) manages public JWK material that you supply. The corresponding private key is not sent to Contentful and is not stored by this resource.

Unlike the redacted secrets above, the public JWK is readable: import and refresh populate it from Contentful. Changing configured JWK material replaces the App Key rather than updating it in place; importing the public key cannot recover its corresponding private key.
