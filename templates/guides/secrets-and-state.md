---
page_title: "Secrets and Terraform state"
description: |-
  Explains how sensitive Contentful values interact with Terraform state, refresh, import, and provider-specific secret handling.
---

# Secrets and Terraform state

Terraform's `sensitive` marking controls display: it redacts values from normal plan and apply output. It does not, by itself, keep those values out of plan or state files. Treat Terraform state and saved plan files as sensitive when they contain credentials or secrets, restrict access to them, and use an appropriately secured backend. See HashiCorp's [sensitive-data guidance](https://developer.hashicorp.com/terraform/language/manage-sensitive-data) for Terraform-level storage and redaction behavior.

Contentful does not expose every secret again after it has been written. The provider therefore has different refresh and import behavior depending on what the Contentful API returns. Do not assume that all attributes marked sensitive have the same lifecycle.

## Provider credentials

Set `CONTENTFUL_MANAGEMENT_ACCESS_TOKEN` in the environment instead of writing a Contentful Management API token directly in Terraform configuration when possible. The `access_token` provider attribute is also supported and is marked sensitive, but the sensitive marking is a display control rather than a general secret-storage mechanism.

## App signing secrets

[`contentful_app_signing_secret`](../resources/app_signing_secret) stores the complete configured `value` in Terraform state after a successful Create or Update. Contentful returns only a redacted representation during later reads, so refresh preserves the previously managed value and cannot detect a replacement made outside Terraform.

A command-line import cannot recover the existing signing secret and leaves `value` null. A subsequent apply with `value` configured writes that configured replacement. A configuration-driven import can write the configured replacement during the import apply. Review the planned replacement before applying an import when preserving the remote secret matters.

## Webhook credentials and secret headers

For [`contentful_webhook`](../resources/webhook), Contentful does not return the HTTP Basic authentication password. Terraform therefore preserves a previously managed `http_basic_password` during refresh and cannot detect an out-of-band password change. Import leaves the password null.

Contentful's `secret = true` flag on a custom webhook header is separate from Terraform sensitivity. It controls how Contentful treats the header value; it does not cause Terraform to mark that value sensitive. Supply the header value from a sensitive Terraform expression when it should be redacted from normal Terraform output. The value can still be present in plan or state data.

## App keys

[`contentful_app_key`](../resources/app_key) manages caller-supplied public JWK material. The corresponding private key is not sent to Contentful and is not stored by this resource. Generate and retain the private key outside the resource, and follow the resource-specific key-rotation constraints when replacing public key material.

## Review the resource contract

Other provider resources and data sources also expose sensitive values. Check the specific resource reference before deciding how to store, import, rotate, or recover a credential: whether Contentful returns a value after creation determines what Terraform can observe during refresh, and the provider documents important exceptions at the affected attribute or resource.
