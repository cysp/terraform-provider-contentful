---
page_title: "Secrets and Terraform state"
description: |-
  Understand which Contentful secrets Terraform stores, which import can recover, and how to manage rotation.
---

# Secrets and Terraform state

Restrict access to Terraform state and saved plans when managing credentials. Sensitive expressions redact values in normal CLI output. Contentful controls which values it returns on refresh and import. [Terraform sensitive-data guidance](https://developer.hashicorp.com/terraform/language/manage-sensitive-data)

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

[`contentful_webhook_signing_secret`](../resources/webhook_signing_secret) manages the signing secret for all webhooks in a space. Use one resource per space.

Coordinate receivers before rotating the space's secret. Contentful describes how to accept the old and new secrets during [key rotation](https://www.contentful.com/developers/docs/extensibility/webhooks/request-verification/#key-rotation).

## App signing secrets

[`contentful_app_signing_secret`](../resources/app_signing_secret) manages the signing secret for an App Definition. Use one resource per App Definition.

## Importing signing secrets

Contentful supports one webhook signing secret per space and one app signing secret per App Definition. Creating a secret overwrites any existing value at that scope. Destroying the resource deletes the current secret, including one rotated outside Terraform. If a replacement at the same scope uses `create_before_destroy`, the final delete removes the newly written secret. To rotate in place, change `value` and apply.

After an uncertain write, Contentful can confirm that a signing secret exists but cannot return its full value. Refresh preserves the last value Terraform wrote, so an external rotation can go undetected.

Import leaves `value` null. Choose who will manage rotation before applying:

- **Terraform.** Configure the complete secret to install. Applying can replace the existing value in the same operation as the import.
- **An external system.** Set `ignore_changes = [value]` to preserve the existing secret on updates. Supply a valid `value` for creation or replacement, including recreation after remote deletion. Destroy still deletes the current secret. Remove `ignore_changes` when you want Terraform to rotate it.

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

Use `jsonencode` with sensitive Terraform inputs for credentials. A `Secret` declaration controls masking in Contentful. In an [App Installation](../resources/app_installation), omitting the parameter object can clear it. In a [UI Extension](../resources/extension), omission retains the installation values returned by Contentful, including after import.

## App Installation secrets

An [App Installation](../resources/app_installation) sends the complete `parameters` object on every write. Include each value you want to keep. Omitted keys are removed on a successful write, and Contentful can reject a missing required parameter. An empty object or an omitted `parameters` attribute can clear existing values.

### Rotation and changes outside Terraform

With a personal access token, Contentful masks Secret values with asterisks. For a top-level value made entirely of asterisks, the provider can retain the planned value after a write or the previous value on refresh. Even an ordinary string changed to literal asterisks can be hidden. This fallback applies only to returned top-level keys. [Contentful Secret installation parameters](https://www.contentful.com/developers/docs/extensibility/app-framework/app-parameters/#secret-installation-parameters)

To rotate a secret, supply its new value and apply. Updating an ordinary parameter or a `marketplace` acknowledgment resends all planned secrets, including any stale value after an external rotation. Changing only `timeouts` skips the write. With `ignore_changes = [parameters]`, another change can still resend the stored object, including old plaintext or imported masks.

### Importing an installation

Import retains the masks that Contentful returns. Supply the actual values before applying. Sending the exact current mask can preserve the existing secret, even when you meant to set a literal string of asterisks. A stale mask can become the literal secret.

If a write fails or Terraform reports an inconsistent result, inspect the installation before retrying. Contentful may have accepted the change, and a masked response cannot confirm which secret it stored.

## Personal access tokens

[`contentful_personal_access_token`](../resources/personal_access_token) stores the secret `token` returned at creation. Contentful does not return that value again, so refresh preserves the known value and import leaves it null.

If a create response was lost and no consumer has the credential, identify and revoke the unusable token and issue a new one for the consumer. If consumers already have the credential, [import the existing token](../resources/personal_access_token#import) to manage it without interrupting them.

## App keys

[`contentful_app_key`](../resources/app_key) manages public JWK material that you supply. The corresponding private key is not sent to Contentful and is not stored by this resource.

Changing the public JWK replaces the App Key. Keep the corresponding private key wherever the app signs its requests.
