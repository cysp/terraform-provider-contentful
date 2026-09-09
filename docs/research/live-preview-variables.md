# Live preview variables: custom preview token storage

Live preview variables form one document per environment. PUT replaces its complete
`variables` object, including nested locale maps, and requires a version header. DELETE
was observed to ignore version preconditions. Empty objects, empty strings, JSON null,
and omitted members have distinct effects.

## Scope and evidence

Contentful's [content preview
guide](https://www.contentful.com/developers/docs/tutorials/preview/content-preview/)
calls the product feature **custom preview tokens** and documents availability for
Premium customers. That product contract concerns preview URL values; it does not
document the `/live_preview/variables` HTTP endpoint. The guide's 100-platform limit
applies to preview platforms, not to this variables document. Space-level [preview
environments](content-preview-environments.md) are a separate resource.

Evidence: a sanitized structural extract of a maintainer-supplied HTTP probe record,
revision `9dbed84465f622758be08555e19a128ea461212b`. Experiment date unrecorded; results
retained by 2026-09-08. The source trace and its URL are excluded because they expose
tenant context. The extract follows the [research redaction
conventions](README.md#evidence-and-redaction).

The original requests exercised document storage in an isolated environment; other
environment and alias requests were reads. Synthetic variable names and locale
placeholders below preserve the tested relationships. They are not an example of a
tenant's configuration.

## Addressing and operations

The singleton path is
`/spaces/{space_id}/environments/{environment_id}/live_preview/variables`.

| Operation | Observed response |
| --- | --- |
| GET existing document | 200 with `sys` and `variables` |
| PUT valid document | 200 with `sys` and `variables` |
| DELETE | 204, empty body; repeated deletion also succeeded |
| GET missing document or missing parent | 404 with CMA `NotFound` |

`sys` contains space and environment links and a document `version`, independent of the
parent's version. GET and PUT returned ordinary JSON despite a vendor-JSON Accept
header. No ETag was observed, and `If-None-Match: *` still returned 200. Submitted
pagination and filter queries had no observed effect; this does not establish universal
query handling.

## Request values and normalization

PUT sends a `variables` object. Stable round trips had this shape:

```text
Document = { variables: object<string, VariableValue> }
VariableValue = string | null | object<ConfiguredLocaleCode, string | null>
```

`ConfiguredLocaleCode` means a locale configured in the addressed environment. This
describes stable observed representations, not every accepted input. Arrays were
accepted and normalized in some positions:

| Input condition | Observed result |
| --- | --- |
| Root `{}`, raw variables map, root `[]`, or root `null` | 422; required variables object absent |
| `variables` is null, string, number, or Boolean | 422; expected Object |
| `variables: {}` | 200; readable empty document |
| `variables: []` | 200; converted to `{}` |
| `variables: ["first", "second"]` | 200; converted to `{"0":"first","1":"second"}` |
| Variable string, empty string, or null | 200; preserved |
| Variable number or Boolean | 422; expected Text |
| Variable `{}` | 200; preserved |
| Variable `[]` | 200; converted to `{}` |
| Variable nonempty array | 422; numeric keys rejected as locale keys |
| Configured locale leaf string, empty string, or null | 200; preserved |
| Configured locale leaf number, Boolean, array, or object | 422; expected Text |
| Unconfigured or differently cased locale key | 422; unknown property |
| Additional root property or fabricated root `sys` | 200; ignored |
| Malformed JSON | 400; service-style invalid-payload error |

PUT replaced all variables and nested locale maps: omitted keys disappeared. An empty
document, an empty locale map, a null leaf, and an empty string remained distinct stored
forms. Acceptance of an ignored root property does not establish support for configuring
that property.

Empty variable names, spaces, punctuation, Unicode, and names resembling built-in tokens
were accepted and preserved. A 256-character key and 1,001 variables succeeded. These
are tested lower bounds, not maximum sizes or proof that such names work in URL
placeholders. A `__proto__` key returned 400 with parser-style invalid-payload text
despite syntactically valid JSON; separate `constructor` and `prototype` keys succeeded.
These outcomes do not identify the service's implementation or establish a
vulnerability.

### Text length

The service accepted 50,000 ASCII characters and rejected 50,001, including a localized
value, with a reported Text maximum of 50,000. It also accepted 50,000 accented BMP
characters and 25,001 astral characters. For those probes the bound was therefore
neither 50,000 UTF-8 bytes nor 50,000 UTF-16 code units. The evidence does not
distinguish all Unicode code-point, combining-sequence, and grapheme counting behavior.

## Versioning and deletion

PUT required `X-Contentful-Version`. The retained requests distinguish header
validation from an existing document's version conflict:

| Version input | Observed PUT result |
| --- | --- |
| Header omitted, malformed, or negative | 400 `BadRequest` |
| Existing document, exact version | 200; version advanced, even for identical content |
| Existing document, zero, stale, or future numeric version | 409 `VersionMismatch` |
| Body `sys.version` or `If-Match` supplied without the required header | Did not substitute for `X-Contentful-Version` |

Rejected validation writes did not change the document version.

An absent document accepted version 0 and tested positive versions, then returned
version 1. Recreation also returned version 1, so versions did not distinguish document
lifetimes. A version header therefore did not enforce update-only intent at an absent
address.

DELETE ignored supplied version headers and succeeded repeatedly. Those observations are
endpoint-specific; they do not establish the same behavior for PUT or other resources.
The evidence does not establish whether a throttled or ambiguous mutation committed.

## Errors and alias reads

Both CMA `{sys, message, details}` and service `{statusCode, error, message}` error
envelopes occurred. Request IDs and validation fields could be absent. Validation errors
could contain multiple entries and echo submitted values, including oversized strings;
such values must be removed from retained evidence.

The missing-document and missing-parent GET probes returned CMA 404 `NotFound`. Other
route/method errors used service-style 404 responses, so HTTP status alone does not
distinguish a missing document from an unsupported route. Missing-parent PUT/DELETE and
cleanup after deleting an environment were not probed.

Alias reads echoed the supplied alias ID in `sys.environment`. Writes through aliases
and behavior after retargeting were not independently probed. Contentful's [environment
alias
concepts](https://www.contentful.com/developers/docs/concepts/environment-aliases/)
explain alias routing; they do not establish this endpoint's write semantics.

## Unresolved behavior

URL rendering, locale fallback, escaping, reserved-token collisions, regional parity,
environment-copy inheritance, alias writes, total document-size limits, and maximum key
lengths/counts remain unverified. Locale membership and casing were exercised, but
changing the environment's locale inventory was not. Accepted storage values do not
establish successful preview URL interpolation.
