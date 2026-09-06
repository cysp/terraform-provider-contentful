---
page_title: "Secrets and Terraform state"
description: |-
  Explains how sensitive Contentful values interact with Terraform state, refresh, import, and provider-specific secret handling.
---

# Secrets and Terraform state

Terraform's `sensitive` marking redacts normal CLI output but does not keep values out of state or saved plans. See HashiCorp's [sensitive-data guidance](https://developer.hashicorp.com/terraform/language/manage-sensitive-data) for storage and redaction behavior.

Refresh and import depend on what Contentful returns, so these resources have different secret-handling behavior.

## App signing secrets

[`contentful_app_signing_secret`](../resources/app_signing_secret) stores the complete configured `value` in Terraform state after a successful Create or Update. Contentful returns only a redacted representation during later reads, so refresh preserves the previously managed value and cannot detect a replacement made outside Terraform.

A command-line import cannot recover the existing signing secret and leaves `value` null. A subsequent apply with `value` configured writes that configured replacement. A configuration-driven import can write the configured replacement during the import apply.

## Webhook credentials and secret headers

For [`contentful_webhook`](../resources/webhook), Contentful does not return the HTTP Basic authentication password. Refresh preserves a previously managed `http_basic_password` but cannot detect changes made outside Terraform. Import leaves the password null and does not change the remote credentials.

On a later update, omitting both `http_basic_username` and `http_basic_password` clears Basic authentication, including on an imported webhook. `ignore_changes` can retain previously managed credentials, but it cannot recover an unreadable imported password. To set credentials after import, configure both the username and password.

Contentful's `secret = true` flag on a custom webhook header is separate from Terraform sensitivity. It controls how Contentful treats the header value; it does not cause Terraform to mark that value sensitive. Supply the header value from a sensitive Terraform expression when it should be redacted from normal Terraform output. The value can still be present in plan or state data.

Contentful also omits secret custom-header values from responses. Refresh preserves a previously managed value for the matching header but cannot verify it or detect an out-of-band replacement. Import leaves an unreadable secret-header value null. When `headers` is omitted from configuration, later updates preserve imported headers and send an unreadable secret header without a value so Contentful retains its secret. Explicitly configuring `headers` requires values for the configured headers; `headers = {}` clears all custom headers.

## App keys

[`contentful_app_key`](../resources/app_key) manages caller-supplied public JWK material. The corresponding private key is not sent to Contentful and is not stored by this resource.

Unlike the redacted secrets above, the public JWK is readable: import and refresh populate it from Contentful. Changing configured JWK material replaces the App Key rather than updating it in place; importing the public key cannot recover its corresponding private key.
