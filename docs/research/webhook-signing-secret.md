# Webhook signing secret API evidence

WebhookSigningSecret is the space singleton at
`/spaces/{space_id}/webhook_settings/signing_secret`. It is separate from
WebhookDefinition and AppSigningSecret. The [provider lifecycle contract](../design/webhook-signing-secret.md)
defines Terraform ownership and state behavior.

## Published contract

Published sources reviewed on 2026-09-13. Direct observations are recorded
separately below.

| Source | Established behavior | Evidence limit |
| --- | --- | --- |
| [Webhook security reference][reference] | PUT creates or replaces the signing secret for all webhooks in the space | No conditional-write or ownership-check guarantee |
| [Signing guide][guide] | One secret per space; exactly 64 characters matching `^[0-9a-zA-Z+/=_-]+$` | Does not establish error payloads, defaults, or propagation timing |
| [GET reference][get] | GET at the singleton endpoint; 200 example has `redactedValue`, `sys.type: WebhookSigningSecret`, and `sys.space` | Example metadata is not a universal server guarantee |
| [PUT reference][put] | PUT accepts `value`; documented 200 example has the same redacted response shape | Live first creation returned 201 instead of the documented 200 |
| [DELETE reference][delete] | DELETE at the same endpoint; success has no content | No endpoint-specific absent-parent/absent-secret error contract |
| [Request verification guide][verification] | Receiver rotation workflow accepts old and new secrets before submitting the new secret to Contentful | Receiver overlap is not evidence of two concurrently stored CMA secrets |

GET and PUT examples have no secret-specific `sys.id`. They show a Space link,
user links, timestamps, and `sys.version: 1`. Neither operation page supplies a
version header requirement. The example's redacted suffix cannot establish
equality of complete secrets.

## First-party SDK

The inspected SDK revision is
`883e2b9dc1c76413d5c24e45f74243da699071e4`:

- [REST adapter][adapter]: `getSigningSecret`, `upsertSigningSecret`, and
  `deleteSigningSecret` use GET, PUT, and DELETE on the same singleton. PUT sends
  the supplied value; neither PUT nor DELETE supplies a version/precondition header.
- [Entity definitions][entity]: request has `value: string`, response has
  `redactedValue: string` and a Space link. The declared signing-secret sys type
  excludes `version`, despite the reference example including it.

SDK type declarations describe client expectations; they do not establish
required server metadata or support for conditional mutations.

## Direct observations

A bounded experiment on 2026-09-14 used one authorized existing space at
`/spaces/{space_id}/webhook_settings/signing_secret`. GET first established
absence. Each PUT supplied a fresh 64-character secret with a different final
four characters. Requests used bearer authentication, had no automatic retries
or redirects, and ran sequentially. GET before and after each replacement checked
presence, version metadata, and whether the redacted value changed to the sent
suffix. No complete values or secret fragments are retained here.

| Request and starting state | Response | Subsequent observation |
| --- | --- | --- |
| GET, secret absent | 404, `sys.type: Error`, `sys.id: NotFound` | No mutation |
| PUT without conditional headers, secret absent | 201, `WebhookSigningSecret` | Redacted value matched the sent suffix |
| PUT with `X-Contentful-Version: 1`, secret present; then repeat with a different value and the same header | 200 for both | Each GET returned the new suffix |
| PUT with `X-Contentful-Version: 0`, secret present | 200 | GET returned the new suffix |
| PUT with `X-Contentful-Version: 2147483647`, secret present | 200 | GET returned the new suffix |
| PUT with `If-None-Match: *`, secret present | 200 | GET returned the new suffix |
| PUT with `If-Match: "nonexistent-probe-etag"`, secret present | 200 | GET returned the new suffix |
| DELETE with `X-Contentful-Version: 0` or `2147483647`, secret present | 204 for each | GET returned 404 `NotFound` after each |
| DELETE with `If-Match: "nonexistent-probe-etag"`, secret present | 204 | GET returned 404 `NotFound` |
| PUT with `X-Contentful-Version: 0` or `If-None-Match: *`, secret absent | 201 for each | Secret created again |
| Unconditional cleanup DELETE | 204 | Final GET returned 404 `NotFound` |

Successful PUT responses omitted `sys.version` and secret-specific `sys.id`.
Subsequent GET responses supplied no version either. The observed `redactedValue`
was the final four characters. The reference
example's `sys.version: 1` neither establishes nor rules out optimistic concurrency.
Since the live responses supplied no version, the experiment cannot label `1` as
a matching or stale version.

The tested headers did not prevent replacement or deletion. In particular,
`X-Contentful-Version: 0` and `If-None-Match: *` did not provide create-only
semantics: both succeeded when absent and overwrote when present. These are
endpoint observations, not a guarantee about all possible conditional mechanisms
or future service behavior. Changed redaction establishes that a replacement
occurred; it does not independently prove equality of the complete stored secret.

### Error responses

Further bounded probes on 2026-09-14 used the same authorized space after GET
confirmed that its secret was absent. Each invalid PUT was followed by GET 404
`NotFound`; no invalid write created a secret. A synthetic parent identifier was
verified absent with GET Space before probing its signing-secret endpoint.

| Request | HTTP status / `sys.id` | Observed error shape |
| --- | --- | --- |
| GET or DELETE, secret absent | 404 / `NotFound` | Message `The resource could not be found.`; `details` is the string `WebhookSigningSecret does not exist.` |
| GET, PUT, or DELETE, parent absent | 404 / `NotFound` | Same message; `details` is an object with `type: Space` and the requested `id` |
| GET without authentication | 401 / `AccessTokenInvalid` | Message asks for an access token |
| PUT malformed JSON | 400 / `BadRequest` | Message `Invalid request payload JSON format` |
| PUT empty body | 422 / `ValidationFailed` | Message `Validation error`; `details.errors` reports `Expected object` at the root |
| PUT `{}` | 422 / `ValidationFailed` | Reports `Expected required property` and `Expected string` at path `["value"]` |
| PUT null, number, array, or object as `value` | 422 / `ValidationFailed` | Reports `Expected string` at path `["value"]` |
| PUT 63 or 65 characters | 422 / `ValidationFailed` | Reports the minimum or maximum length of 64 |
| PUT 64 characters containing `!`, or 64 non-ASCII characters | 422 / `ValidationFailed` | Reports failure to match `^[0-9a-zA-Z+/=_-]+$` |
| PUT empty string | 422 / `ValidationFailed` | Reports both minimum length and pattern failures |

Validation details can contain the submitted value, including invalid strings,
null, numbers, arrays, and objects. These responses must be treated as potentially
secret-bearing. The example above uses structural descriptions rather than
retaining submitted values. Final GET confirmed that the secret remained absent.

## Unverified behavior

Rotation propagation, webhook delivery, permission-denied responses, and error
precedence when multiple request conditions fail remain unverified. Invalid
replacement nonmutation has not been live-probed with an existing secret.
The observations do not establish conditional-write support through any untested
header or protocol.

[reference]: https://www.contentful.com/developers/docs/references/content-management-api/webhook-security/
[get]: https://www.contentful.com/developers/docs/references/content-management-api/webhook-security/get-a-webhook-signing-secret/
[put]: https://www.contentful.com/developers/docs/references/content-management-api/webhook-security/create-a-webhook-signing-secret/
[delete]: https://www.contentful.com/developers/docs/references/content-management-api/webhook-security/delete-a-webhook-signing-secret/
[guide]: https://www.contentful.com/developers/docs/extensibility/webhooks/secrets/
[verification]: https://www.contentful.com/developers/docs/extensibility/webhooks/request-verification/#key-rotation
[adapter]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/webhook.ts
[entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/webhook.ts
