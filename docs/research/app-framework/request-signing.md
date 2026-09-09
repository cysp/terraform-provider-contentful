# App request signing and verification

AppSigningSecret is a symmetric secret addressed by its organization and AppDefinition.
PUT creates or replaces the secret; GET and PUT return only its final four characters.
AppSignedRequest produces headers for a frontend request, and backend verification
checks signed request provenance.

See the [App Framework API inventory](README.md#api-inventory) for the related endpoint
families and [App Identity](identity.md) for asymmetric keys and access tokens. The
signing-secret sources and direct probe below are separate from the [shared App Framework
study](README.md#scope-and-evidence). AppSignedRequest and verification behavior here
comes from the cited documentation and toolkit, without an end-to-end request test.

## AppSigningSecret addressing and operations

The singleton path is
`/organizations/{organization_id}/app_definitions/{app_definition_id}/signing_secret`.
GET reads the redacted representation, PUT sends `{"value": "<secret>"}`, and DELETE
removes it. The placeholder is not a valid example secret.

## AppSigningSecret wire contract and sources

Contentful documents an app signing secret as exactly 64 characters matching
`^[0-9a-zA-Z+/=_-]+$`. Only one secret exists for an AppDefinition, and a PUT rotates it
by replacing the current value.

Source: [Contentful CMA App signing secret
reference](https://www.contentful.com/developers/docs/references/content-management-api/app-signing-secret/).

The documented GET and PUT response examples contain `sys` and `redactedValue`. In both
examples, `redactedValue` is the final four characters of the secret. The submitted
`value` is absent. The documented DELETE response has no content.

Sources:

- [Get the current app signing secret](https://www.contentful.com/developers/docs/references/content-management-api/app-signing-secret/get-the-current-app-signing-secret/)
- [Create or overwrite the app signing secret](https://www.contentful.com/developers/docs/references/content-management-api/app-signing-secret/create-or-overwrite-the-app-signing-secret/)
- [Remove the current app signing secret](https://www.contentful.com/developers/docs/references/content-management-api/app-signing-secret/remove-the-current-app-signing-secret/)

The first-party `contentful-management.js` model independently defines `redactedValue`
as the final four characters and the request as a 64-character value matching the same
regular expression. Its `sys` declaration includes `organization` and `appDefinition`
links and excludes the signing secret's own `id` and `version`. Its REST adapter sends
the value only in PUT request data, decodes GET and PUT as `AppSigningSecretProps`, and
decodes DELETE without resource data.

Sources, pinned to the reviewed first-party commit
`cc096a337f0e1db6114e8da645d69bb6eb90f11c`:

- [`AppSigningSecretSys`, `AppSigningSecretProps`, and `CreateAppSigningSecretProps`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/entities/app-signing-secret.ts#L7-L24)
- [App signing secret REST endpoints](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/app-signing-secret.ts#L8-L36)

No reviewed operation returns the complete submitted secret.

## Direct CMA observation

The public reference describes one endpoint as "create or overwrite" and shows a 201
response, but it does not separately state the overwrite status. A minimal probe on
2026-08-24 used a disposable AppDefinition and two synthetic 64-character secrets. The
probe retained no complete secret or response value.

| Operation | Status | Structural result |
| --- | ---: | --- |
| Create disposable AppDefinition | 201 | AppDefinition created for the probe |
| First signing-secret PUT | 201 | `sys` plus four-character `redactedValue`; it equalled the submitted value's final four characters and not the complete value |
| Overwrite signing-secret PUT | 200 | Same response structure and redaction relationship |
| GET signing secret | 200 | Same response structure and redaction relationship for the replacement value |
| DELETE signing secret | 204 | Empty response body |

## Reconstruction and rotation limits

The complete value cannot be recovered from a GET or mutation response. A matching
four-character suffix cannot establish equality between two secrets; an external
replacement with the same suffix is indistinguishable by this read. Backup or migration
therefore requires the original value from another source, or an intentional replacement
coordinated with consumers.

Contentful documents one active signing secret per app definition. Rotating it changes
the secret used for subsequent signed app events; backend verification must account for
the change. The observations do not establish an overlap period, propagation delay, or
treatment of already queued events.

## AppSignedRequest and verification

AppSignedRequest requires `method` and `path`, with optional string-valued `headers` and
an optional string `body`. JSON must be serialized to those exact body bytes before
signing.
Its SDK method union is GET/PUT/POST/DELETE/PATCH/HEAD; OPTIONS acceptance is not
established by that declaration. The path excludes scheme, hostname, and port. [Request
entity][signed-entity], [reference][signed-reference].

The result's `additionalHeaders` includes `x-contentful-signature`,
`x-contentful-signed-headers`, `x-contentful-timestamp`, and space/environment/user ID
headers. `sys` links definition, space, and environment, without id/version in the SDK.
Creating this result requires an installation and signing secret; it does not deliver
the HTTP request. Space users can request signatures, so signed provenance does not
replace backend authorization for the requested operation. [Response
entity][signed-entity], [reference][signed-reference].

Verification uses HMAC-SHA256 over the canonical method, path, selected headers, and
exact body. The signature, signed-header list, timestamp, and applicable space,
environment, and actor context must be checked together. The toolkit default TTL is 30
seconds; zero disables age checking. Timestamp checking alone is not persistent
duplicate suppression. [Verification guide][verification], [signer][signer],
[verifier][verifier].

[signed-reference]: https://www.contentful.com/developers/docs/references/content-management-api/app-signed-request/
[verification]: https://www.contentful.com/developers/docs/extensibility/app-framework/request-verification/
[signer]: https://github.com/contentful/node-apps-toolkit/blob/64fa31b6b2223cd1c8b1798fa540e8aad5e2d319/src/requests/sign-request.ts
[verifier]: https://github.com/contentful/node-apps-toolkit/blob/64fa31b6b2223cd1c8b1798fa540e8aad5e2d319/src/requests/verify-request.ts
[signed-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/app-signed-request.ts
