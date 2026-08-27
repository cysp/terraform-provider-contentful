# App signing secret CMA contract

This note records the Contentful Management API behavior that the
`contentful_app_signing_secret` resource relies on. It focuses on observable
request, response, and lifecycle semantics rather than internal Contentful
implementation details.

## Authoritative sources

Contentful documents an app signing secret as exactly 64 characters matching
`^[0-9a-zA-Z+/=_-]+$`. Only one secret exists for an AppDefinition, and a PUT
rotates it by replacing the current value.

Source: [Contentful CMA App signing secret reference](https://www.contentful.com/developers/docs/references/content-management-api/app-signing-secret/).

The documented GET and PUT response examples contain `sys` and
`redactedValue`. In both examples, `redactedValue` is the final four characters
of the secret. The submitted `value` is absent. The documented DELETE response
has no content.

Sources:

- [Get the current app signing secret](https://www.contentful.com/developers/docs/references/content-management-api/app-signing-secret/get-the-current-app-signing-secret/)
- [Create or overwrite the app signing secret](https://www.contentful.com/developers/docs/references/content-management-api/app-signing-secret/create-or-overwrite-the-app-signing-secret/)
- [Remove the current app signing secret](https://www.contentful.com/developers/docs/references/content-management-api/app-signing-secret/remove-the-current-app-signing-secret/)

The first-party `contentful-management.js` model independently defines
`redactedValue` as the final four characters and the request as a 64-character
value matching the same regular expression. Its REST adapter sends the value
only in PUT request data, decodes GET and PUT as `AppSigningSecretProps`, and
decodes DELETE without resource data.

Sources, pinned to the reviewed first-party commit
`cc096a337f0e1db6114e8da645d69bb6eb90f11c`:

- [`AppSigningSecretProps` and `CreateAppSigningSecretProps`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/entities/app-signing-secret.ts#L10-L22)
- [App signing secret REST endpoints](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/app-signing-secret.ts#L8-L36)

No reviewed operation returns the complete submitted secret.

## Direct CMA observation

The public reference describes one endpoint as "create or overwrite" and shows
a 201 response, but it does not separately state the overwrite status. A
minimal probe on 2026-08-24 used a disposable AppDefinition and two synthetic
64-character secrets. The probe retained no token, AppDefinition ID, complete
secret, or response value.

| Operation | Status | Structural result |
| --- | ---: | --- |
| Create disposable AppDefinition | 201 | AppDefinition created for the probe |
| First signing-secret PUT | 201 | `sys` plus four-character `redactedValue`; it equalled the submitted value's final four characters and not the complete value |
| Overwrite signing-secret PUT | 200 | Same response structure and redaction relationship |
| GET signing secret | 200 | Same response structure and redaction relationship for the replacement value |
| DELETE signing secret | 204 | Empty response body |

The disposable signing secret and AppDefinition were deleted after the probe.

## Terraform contract

`value` remains a required, sensitive, configuration-owned value. Terraform
must have the value to create or overwrite the remote secret, while Contentful
never returns enough information to reconstruct it. Create and Update therefore
publish the planned value, and Read preserves the prior state value when GET
returns only redacted information.

Terraform's `sensitive` behavior masks routine CLI and HCP Terraform UI output;
it does not omit the value from state or saved plan files. HashiCorp documents
state and plan files as containing resource attributes and potentially sensitive
values. A local Terraform run normally stores state in a plaintext file. Remote
state storage can avoid persisting state to local disk and can provide protection
such as encryption at rest, but the protection depends on the configured
backend. Practitioners must protect access to both state and saved plan artifacts.

Sources:

- [Manage sensitive data in your configuration](https://developer.hashicorp.com/terraform/language/manage-sensitive-data)
- [State storage and locking](https://developer.hashicorp.com/terraform/language/state/backends)

The resource does not implement a write-only argument or use an ephemeral value.
HashiCorp distinguishes `sensitive`, which controls display, from ephemeral
values and write-only resource arguments, which Terraform omits from state and
plan files. This provider change documents the existing persistence contract; it
does not add write-only behavior or an external-rotation trigger.

`redactedValue` is not represented in Terraform state. It is response metadata,
not an independently manageable value, and four matching characters do not
prove that two 64-character secrets are equal. Adding it would not make drift
detection complete or make the secret recoverable.

Consequently, the provider detects remote existence through GET and deletion
through 404, but it cannot reliably detect an out-of-band secret replacement.
Even comparing the final four characters would miss replacements with the same
suffix and would not recover the configuration-owned value.

Import remains supported for identity and existence adoption. An import can
read `sys` and `redactedValue`, but it cannot populate the required configured
`value`, so the imported resource initially has a null value in state. The
practitioner must configure a valid value after import. The next apply overwrites
the remote signing secret with that configured value and stores the complete
configured replacement in resource state; import does not claim to recover or
preserve an unknown existing secret.
