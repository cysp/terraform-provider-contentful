---
page_title: "Secrets and Terraform state"
description: |-
  Understand which Contentful secrets Terraform stores, which import can recover, and how to manage rotation.
---

# Secrets and Terraform state

Terraform's `sensitive` marking redacts normal CLI output but does not keep values out of state or saved plans. Restrict access to both when using resources that manage credentials. See HashiCorp's [sensitive-data guidance](https://developer.hashicorp.com/terraform/language/manage-sensitive-data) for storage and redaction behavior.

Refresh can verify a value only when Contentful returns it. Import starts without a previously stored secret, so it cannot recover values that Contentful withholds.

| Value | Refresh | Import |
| --- | --- | --- |
| App signing secret | Preserves the previously stored value; cannot detect external rotation | Leaves `value` null |
| Webhook Basic authentication password | Preserves a previously managed password; cannot detect external changes | Leaves `http_basic_password` null |
| Secret webhook header | Preserves a previously managed value for the matching header; cannot detect external changes | Leaves an unreadable header value null |
| Personal Access Token | Preserves the token returned at creation | Leaves `token` null |
| Delivery API access token | Reads the token from Contentful | Reads the token from Contentful |
| App Key public JWK | Reads the public key from Contentful | Reads the public key; cannot recover its private key |

## App signing secrets

Whenever [`contentful_app_signing_secret`](../resources/app_signing_secret) writes a secret to Contentful, it stores the complete configured `value` in Terraform state. Contentful returns only a redacted representation during reads, so refresh preserves the stored value and cannot detect a secret rotated outside Terraform.

To rotate the secret through Terraform, change `value` and apply. Changing only `timeouts` leaves the remote secret unchanged, including a secret rotated outside Terraform. With `ignore_changes = [value]`, configuration changes to `value` do not rotate an existing secret.

Import cannot recover the complete secret and leaves `value` null. Without `ignore_changes = [value]`, applying the configured value replaces the remote secret; a configuration-driven import can do this during the import apply. With that lifecycle setting, the imported value remains null, including after timeout changes.

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

The `parameters` values on [`contentful_app_installation`](../resources/app_installation) and [`contentful_extension`](../resources/extension) can contain both ordinary configuration and secrets. Defining a Contentful parameter as type `Secret` does not mark its Terraform value sensitive. Supply secrets through sensitive Terraform expressions to redact normal output; the values are still stored in state and saved plans.

## Personal access tokens

[`contentful_personal_access_token`](../resources/personal_access_token) stores the secret `token` returned at creation. Contentful does not return that value again, so refresh preserves the known value and import leaves it null.

Changing `name`, `scopes`, or `expires_in` replaces the token and revokes the old one. Review dependent consumers before applying a replacement. Changing only `timeouts` preserves the existing token; destroying the resource revokes it.

## App keys

[`contentful_app_key`](../resources/app_key) manages public JWK material that you supply. The corresponding private key is not sent to Contentful and is not stored by this resource.

Unlike the redacted secrets above, the public JWK is readable: import and refresh populate it from Contentful. Changing configured JWK material replaces the App Key rather than updating it in place; importing the public key cannot recover its corresponding private key.
