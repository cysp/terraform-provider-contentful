# Live preview variables CMA contract

Live preview variables form one document per environment. The observed PUT
replaces the whole document and requires its version; the observed DELETE ignores
version preconditions. The
provider's [Live Preview variables contract](../design/terraform-value-semantics.md#live-preview-variables)
defines request validation and mutation reconciliation against these observed
endpoint behaviors.

## Evidence and scope

The [sanitized probe ledger, revision 9dbed84465f622758be08555e19a128ea461212b](https://gist.github.com/cysp/1fa4a7a837bf3220837e571ddf1499df/9dbed84465f622758be08555e19a128ea461212b)
records 141 requests on the US API host, with writes confined to one authorized
environment. Other environments and aliases were read only. The ledger is direct
observational evidence supplied by the maintainer, not a published API guarantee.
Contentful's [content preview guide](https://www.contentful.com/developers/docs/tutorials/preview/content-preview/)
describes the product feature as custom preview tokens. The variables endpoint
is separate from space-level preview platforms and environment UI configuration.

## Observed contract

`/spaces/{space}/environments/{environment}/live_preview/variables` is a
singleton. GET and PUT return HTTP 200 with `sys` and `variables`; DELETE returns
empty HTTP 204. PUT sends only `{"variables": {...}}` and replaces the entire
object, including nested locale maps. Empty objects, empty strings, null leaves,
and omitted keys remain distinct. Some array inputs were converted to objects at
the variables and locale-map levels.

`sys` contains space and environment links and a document `version`, independent
of the parent version. PUT requires `X-Contentful-Version`. An existing document
requires its exact version, advances the version even for identical content,
and rejects version 0 or stale versions with 409 `VersionMismatch`. An absent
document accepts version 0 and tested positive versions and starts at version 1.
DELETE ignores version headers and succeeds repeatedly. Recreation resets the
version, so optimistic updates do not distinguish document lifetimes.

Errors use both CMA `{sys,message,details}` and service
`{statusCode,error,message}` envelopes. Optional request IDs and validation
fields may be absent. Missing documents and parents returned CMA 404 `NotFound`;
a generic service 404 does not establish absence. Validation error details can
contain submitted values.

Alias reads echoed the supplied alias ID in `sys.environment`. Alias writes and
retargeting were not independently probed. See
Contentful's [alias concepts](https://www.contentful.com/developers/docs/concepts/environment-aliases/)
for routing behavior.

## Entitlement observation

A separate authorized read-only probe on 2026-09-08 (Australia/Sydney) returned
HTTP 200 for the parent environment and vendor JSON with HTTP 403 for variables
GET:

```json
{"statusCode":403,"error":"Forbidden","message":"previewLocalization is not enabled"}
```

This probe made no mutations. Live Terraform CRUD remains unverified; the
ledger's mutation results came from direct HTTP probes.

## Mock boundaries

The [CMA test-server conformance reference](cma-test-server-conformance.md) summarizes
the implemented lifecycle and test coverage. Mocked tests exercise provider
behavior against these fixture conventions:

- The locale inventory is fixed to `en-US`. Configurable environment locale
  inventories are not modeled.
- The 50,000-character Text limit counts Unicode code points. This reproduces
  the observed ASCII boundary and accepted BMP/astral examples; combining-sequence
  and grapheme semantics remain unverified. This counting rule is a mock
  convention, not a provider-side restriction.
- Missing-parent PUT/DELETE responses and cleanup after environment deletion
  follow mock lifecycle conventions. The ledger directly observed only the
  missing-parent GET and did not mutate environments.

Alias routing, response update metadata, HEAD, trailing-slash GET, and service-style
404 responses for POST/PATCH and the space-level route are not modeled. Generated
handler errors use ordinary JSON, including conflicts; independent client fixtures
cover vendor-JSON decoding. Whole non-object request bodies receive generated
decoder errors, while the Terraform client always sends an object envelope.
Authentication distinctions remain the shared fake's behavior.

## Unverified behavior

URL rendering, locale fallback, escaping, reserved-token collisions, EU-host
parity, environment-copy inheritance, alias writes, total document-size limits,
and maximum key lengths/counts remain unverified. The ledger does not establish
whether a throttled or ambiguous write committed. Provider retry rules are defined
in the [HTTP retry policy](../design/contentful-http-retry-policy.md).
