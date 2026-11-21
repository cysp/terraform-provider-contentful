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
| [Webhook Basic password](#basic-authentication) | Retains prior value when omitted | Null when omitted |
| [Secret webhook header](#secret-custom-headers) | Retains the value at the exact key when omitted | Null when omitted |
| [App Installation `Secret` parameter](#app-installation-secrets), read with a personal access token | Can retain prior state for an all-asterisk response | Retains mask; original value unavailable |
| [Personal Access Token](#personal-access-tokens) | Preserves token from creation | `token` is null |
| Delivery API access token | Reads token | Reads token |
| [App Key public JWK](#app-keys) | Reads public key | Reads public key; private key unavailable |

## Webhook signing secrets

[`contentful_webhook_signing_secret`](../resources/webhook_signing_secret) manages the signing secret for all webhooks in a space. Use one resource per space.

Coordinate receivers before rotating the space's secret. Contentful describes how to accept the old and new secrets during [key rotation](https://www.contentful.com/developers/docs/extensibility/webhooks/request-verification/#key-rotation).

## App signing secrets

[`contentful_app_signing_secret`](../resources/app_signing_secret) manages the signing secret for an App Definition. Use one resource per App Definition.

If a create or update loses its acknowledgement, Contentful may have installed the submitted value. A read shows only the final four characters. Terraform retains the previously acknowledged value on a failed update and records no value on a failed create. Reverting configuration to the stored value can therefore produce a no-op while the app uses a different secret. Verify a fresh signed request with the intended key. If a rotation is needed, coordinate with the app backend and other writers, apply a fresh value, and verify a new signed request.

If a delete loses its acknowledgement, check whether the App Definition still has a secret. If one exists, confirm with other writers that it should be removed before applying again. The delete has no version precondition and can remove a replacement.

## Importing signing secrets

Contentful supports one webhook signing secret per space and one app signing secret per App Definition. Creating a secret overwrites any existing value at that scope. Destroying the resource deletes the current secret, including one rotated outside Terraform. If a replacement at the same scope uses `create_before_destroy`, the final delete removes the newly written secret. To rotate in place, change `value` and apply.

After an uncertain write, Contentful can confirm that a signing secret exists but cannot return its full value. Refresh preserves the last value Terraform acknowledged, so an external rotation can go undetected.

Import leaves `value` null. Choose who will manage rotation before applying:

- **Terraform.** Configure the complete secret to install. Applying can replace the existing value in the same operation as the import.
- **An external system.** Set `ignore_changes = [value]` to preserve the existing secret on updates. Supply a valid `value` for creation or replacement, including recreation after remote deletion. Destroy still deletes the current secret. Remove `ignore_changes` when you want Terraform to rotate it.

## Webhook credentials and secret headers

### Basic authentication

For [`contentful_webhook`](../resources/webhook), Contentful normally omits the HTTP Basic password. When it does, refresh retains the previously managed value and import leaves it null. Terraform records a password returned by Contentful.

On a later update, omitting both `http_basic_username` and `http_basic_password` clears Basic authentication, including on an imported webhook. To keep Basic authentication when updating an imported webhook, supply both credentials. Terraform can read the webhook and report no changes with just the imported username. If an update would send that username alone, the provider reports a missing password before sending the request. `ignore_changes` can retain a previously managed password only when Terraform already has it in state.

### Secret custom headers

Contentful's `secret = true` flag on a custom webhook header controls how Contentful treats the value. Terraform sensitivity is configured separately: supply a sensitive Terraform expression to redact the header value from normal Terraform output. The value can still be present in plan or state data.

Terraform records returned header values. When Contentful omits a secret value, refresh retains the previously managed value at the exact header name, including its case. An external rotation can then go undetected. Import leaves an omitted value null.

When updating a webhook:

- Omit `headers` to keep the headers in state. Terraform resends known values. For a secret header with a null value, it sends only the name and `secret = true`.
- To keep an existing secret while updating other headers, use its exact name with `{ secret = true }` and omit `value`. For example, `headers = { "X-Token" = { secret = true }, "X-Format" = { value = "json" } }` changes the format header and leaves the stored token in Contentful.
- Use `ignore_changes = [headers["X-Token"]]` to retain that entry from state while managing other headers. A null secret value preserves the remote secret. A known value is sent again and can overwrite an external rotation.
- Supply `value` to set or replace a header, including when creating the webhook, renaming the key, or converting an ordinary header to a secret. A configured map removes keys that are absent from the effective plan.
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
