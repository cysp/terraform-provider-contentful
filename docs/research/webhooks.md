# Webhooks: configuration, delivery, and observability

WebhookDefinition configures space-scoped HTTP delivery without an app installation.
Configuration, signing settings, delivery attempts, and health use distinct API paths.
App Event subscriptions have a separate contract.

## Scope and evidence

Published sources reviewed: 2026-09-09. First-party source links identify their inspected
revisions. The direct configuration probes below do not establish event delivery. See
[App Events](app-framework/events.md) for the separate subscription and delivery
contract, and the [research conventions](README.md#evidence-and-redaction) for evidence
classes and redaction.

## Addressing and configuration

`W = /spaces/{space_id}/webhook_definitions/{webhook_id}` below.

WebhookDefinition is space-scoped, has its own ID, and needs no app installation. It
supports headers and secret headers, HTTP Basic authentication, filters, transformations,
`active: false`, and wildcard topics such as `*.*`, `Entry.*`, and `*.save`. Those differ
from AppEventSubscription's HTTP target configured on an AppDefinition and its exact topics. A
wildcard may include future event types. [Webhook entity][webhook-entity],
[filters][webhook-filters].

| Resource / operation | Methods and paths |
| --- | --- |
| WebhookDefinition collection | GET/POST `/spaces/{space_id}/webhook_definitions` |
| Definition read / specified-ID create / update / delete | GET/PUT/DELETE `W` |
| Call summaries | GET `/spaces/{space_id}/webhooks/{webhook_id}/calls` |
| Call detail | GET `/spaces/{space_id}/webhooks/{webhook_id}/calls/{call_id}` |
| Health | GET `/spaces/{space_id}/webhooks/{webhook_id}/health` |
| WebhookSigningSecret | GET/PUT/DELETE `/spaces/{space_id}/webhook_settings/signing_secret` |

Definition paths use `webhook_definitions`; observability paths use `webhooks`. The SDK
sends `X-Contentful-Version` for updates. The signing secret is one space-level setting
affecting **all webhooks in that space**, unlike an AppSigningSecret's one-definition
scope. Its wire contract and evidence limits are recorded in
[Webhook signing secret API evidence](webhook-signing-secret.md). It was not changed
during the configuration experiments. [Webhook adapter][webhook-sdk], [Webhook security][webhook-security].

The SDK still contains GET/PUT/DELETE for the space singleton
`/webhook_settings/retry_policy` with `maxRetries`, but marks it deprecated because its
EAP ended and removal is planned for the next major version. This is legacy source
surface, not evidence of a supported new resource or account entitlement. [Adapter
deprecations][webhook-sdk].

## Delivery and observability

Under Contentful's documented default retry policy, a webhook delivery that receives
HTTP 429 or 5xx may be retried up to two additional times, approximately 30 seconds apart.
Recipient timeout is at most 30 seconds and timed-out requests are not retried.
Duplicate delivery can occur, and the documented deduplication header is
`X-Contentful-Idempotency-Key`. These documented behaviors apply to webhooks. No inspected
App Event source directly establishes identical behavior for external HTTP app targets. No
ordering guarantee was identified. [Webhook delivery][webhook-overview].

Webhook activity logs retain up to 500 entries, dropping the oldest; this is a count
limit, not a guaranteed duration. Call details truncate request bodies at 500 kB and
response bodies at 200 kB. Health describes recent calls. No public replay/resend
operation was established. These webhook APIs do not establish an equivalent App Event
history endpoint. [Activity log][webhook-activity], [call reference][webhook-calls].

## Related event families

Related emitter families must also be kept distinct. Release events describe the release
container; ReleaseAction, BulkAction, and ScheduledAction describe specific operations.
An `execute` event can report a failed operation as well as a successful one: the
outcome must be inspected. App Events list these topic families, but this semantic
account comes from the action-event guide, not a live execution test. [Action
events][action-events], [AppEventSubscription reference][event-reference].

Two similarly named products do not fill App Event contract gaps:

- Bulk Content Operations is a newer asynchronous job API, distinct from
  editor-facing BulkAction. Its announcement describes job status webhooks and
  seven-day status/export retention; those are not webhook activity-log
  retention limits. No separate Bulk Content Operation topic family appeared
  in the observed 88-topic App Event allowlist. [Announcement][bulk-operations].
- The Live Events dashboard monitors incoming Personalization SDK Track,
  Component, Identify, and Page events. It is not documented as an
  AppEventSubscription delivery dashboard. [Announcement][live-events].

## Direct configuration observations

Experiment dates unrecorded; topic, default, header, and Basic password results retained
by 2026-09-03.

### Topic validation

A disposable Webhook probe submitted omitted, empty, and one-element `topics` to an
existing disposable test space. Each invalid request was followed by a read that
confirmed no Create or Update occurred.

| Request | Status | Structural observation |
| --- | ---: | --- |
| `POST /spaces/{space_id}/webhook_definitions` with omitted `topics` | 422 | `ValidationFailed`; error name `invalid_type`, path `topics`, expected array but received undefined; no matching Webhook was created |
| `POST /spaces/{space_id}/webhook_definitions` with `topics: []` | 422 | `ValidationFailed`; error name `topics`, empty path, `Topics cannot be empty`; no matching Webhook was created |
| `POST /spaces/{space_id}/webhook_definitions` with `topics: ["Entry.publish"]` | 200 | One-topic configuration was accepted and returned unchanged |
| `PUT /spaces/{space_id}/webhook_definitions/{webhook_id}` with omitted `topics` | 422 | Same missing-topic validation shape as Create; the existing Webhook remained unchanged |
| `PUT /spaces/{space_id}/webhook_definitions/{webhook_id}` with `topics: []` | 422 | Same empty-topic validation shape as Create; the existing Webhook remained unchanged |
| `DELETE /spaces/{space_id}/webhook_definitions/{webhook_id}` | 204 | The disposable probes were removed |

### Defaults and headers

Separate disposable Webhook probes recorded only response-member presence, selected
Boolean values, and whether an ordinary test header value matched the request.

| Request | Status | Structural observation |
| --- | ---: | --- |
| Create with `active` and `headers` omitted | 200 | Response contained `active: true` and `headers: []` |
| Update with `active` omitted | 200 | Response contained `active: true` |
| Update after storing a header, with `headers` omitted | 200 | Response contained `headers: []`; the prior header was absent |
| Create with a header whose `secret` member was omitted | 200 | Response omitted `secret` but retained the ordinary header value unchanged |
| Create with a header whose `secret` member was `false` | 200 | Response contained Boolean `secret: false` and retained the ordinary value |

### Basic authentication password

Contentful's [Webhook configuration
reference](https://www.contentful.com/developers/docs/extensibility/webhooks/configure-webhook/#http-basic-authentication)
defines paired username/password writes and paired-null removal. Retained Create,
Update, and GET observations found that CMA accepted the password but omitted
`httpBasicPassword` from every response while retaining `httpBasicUsername`. The record
contains no password or username values.

Omission is a read limitation: it cannot reconstruct the submitted password or prove
that a stored password still equals a previous value. Explicit response null and an
absent property must not be assumed equivalent without evidence.

## Basic authentication field presence and version changes

On 2026-11-26 (UTC), raw JSON requests exercised inactive disposable Webhooks.
Every PUT to an existing ID used its current version. Every successful POST or PUT
returned 200 and was followed by GET. Both responses omitted
`httpBasicPassword` and agreed on the username, name, and version.

| Operation | Observed result |
| --- | --- |
| POST with a nonempty pair, then rename with both fields omitted | POST returned the username at version 0. PUT retained it at version 1. Resending the original pair with the new name kept version 1. Resending both fields with the same username and a different password advanced the version to 2. |
| PUT a nonempty username with the password omitted, null, or empty, or the reverse | 422 `ValidationFailed`. The message was `Validation error` and `details.errors` contained one entry with an empty `path`, the required field as its `name`, and a detail saying why it was required. GET returned the previous username and version. When the rejected PUT also attempted a rename, GET returned the original name. |
| PUT one null and omit the other field on a nonempty pair | 422 with the omitted field as the error name and detail `HTTP Basic Auth username and password must both be provided when setting to null`. GET retained the version and username. |
| PUT paired nulls on a nonempty pair | The version advanced once and both PUT and GET included an explicit null username. Subsequent lone-null writes on a cleared pair preserved the version. |
| Collection POST or specified-ID PUT to an unused ID | A lone nonempty string in either direction returned the same field-specific 422. GET on each rejected specified ID returned 404. Omitting both, supplying only a null password, or sending paired nulls created a definition at version 0 with an explicit null username. |

Empty strings also differed from null. Starting with a nonempty pair, a lone
empty password advanced the version and retained the username. A lone empty
username advanced the version and returned an empty username. After either
write, setting the **empty** field alone to null succeeded and advanced the
version. Setting the **nonempty** field alone to null returned 422. Two empty
strings were accepted. A null and an empty string were accepted together in
both directions.

The version comparison supports the inference that omitting both fields retained
the password. The probes triggered no deliveries, and the redacted responses
cannot establish which password a delivery would use.

### Credential limits

Follow-up probes on two disposable inactive Webhooks tested the documented
50-character username and 255-character password limits. A specified-ID PUT
accepted ASCII strings at both limits. Adding one character to either field
returned 422 `ValidationFailed` on update, and GET retained the prior name,
username, and version. Repeated U+00E9 fit at both limits. A username of 25
supplementary characters and a password of 127 supplementary characters plus
one ASCII character were accepted together. A username of 26 or a password of
128 supplementary characters was rejected independently. These boundaries are
consistent with counting UTF-16 code units. Collection POST also accepted the ASCII
boundary pair and returned 422 for each overlong field. The control was deleted
(204) and GET confirmed absence (404). When both fields exceeded the limits,
`details.errors` listed the username then the password, each with an empty
`path` and a field-specific length message. An oversized field submitted alone
also produced a missing-counterpart error before its length error. CMA accepted a colon in the
username, although the [HTTP Basic encoding](https://www.rfc-editor.org/rfc/rfc7617.html#section-2)
reserves the first colon as the separator.

## Existing-configuration read observations

Observed: 2026-09-09 (UTC), passive collection reads and two definition-detail
comparisons. Secret headers had exactly `key` and `secret`, without `value`, matching
the [documented secret-header redaction][webhook-headers]. Definitions with a Basic
authentication username omitted the password. Both sampled detail responses exactly
matched their list items.

The sample contained no repeated header names, including case variants.

## Webhook definition versions

On 2026-11-21 (UTC), POST and specified-ID PUT each created an inactive webhook
with HTTP 200 and `sys.version: 0`. On unused IDs, PUT created the definition
when `X-Contentful-Version` was omitted, set to 0, or set to 7.

An existing webhook required the exact version, including at version 0.
An identical PUT and a PUT retaining an existing secret by its exact header
name both returned 200 with the same version. Changing the name or rotating
the secret advanced the version by one. Missing, stale, and future versions
returned 409 `Conflict`, including on otherwise identical requests. GET
confirmed the stored versions. All disposable definitions were deleted (204)
and their absence confirmed by GET (404).

## Header names, order, and secret updates

On 2026-11-21 (UTC), disposable inactive webhook definitions were created by POST
and specified-ID PUT. Each successful write was read back by ID. Each
definition was deleted (204) and its absence confirmed by GET (404).

| Request | Configuration API observation |
| --- | --- |
| Repeat the exact header name | POST returned 422 `ValidationFailed` with identical or different ordinary values. A versioned PUT with different values returned 422 in either order. GET retained the preceding configuration and version after each rejected PUT. |
| Supply two names differing only in case | POST, PUT, and GET returned both spellings. Each ordinary value stayed with its supplied key. With two secret values, each response returned both keys with `secret: true` and omitted both values. |
| Reverse the order of two ordinary headers | PUT and GET returned the new order. The version increased by one even though the names and values were unchanged. This also held for case variants. |
| Send a new header with `secret: true` and omit its value | POST returned 422 `ValidationFailed` with the missing field at `headers.0.value`. PUT to two unused IDs returned 500 `ServerError`. Neither definition was created. PUT to an existing ID returned 409 with a wrong version and 422 `ValidationFailed` with the current version. The latter response contained a string at `details.errors` naming the header. GET confirmed the version was unchanged. |
| Omit several header values on creation | POST listed each missing `headers[i].value` in request order. PUT to an unused ID returned one 422 `ValidationFailed` response listing the headers missing both `secret` and `value`, even when another header supplied `secret: true` without a value. With just `secret: false` and no value, PUT returned the same 500 as for `secret: true`. GET confirmed the IDs remained unused. |
| Supply an empty string as a secret header value | POST and specified-ID PUT both created the definition at version 0. |
| Omit a stored secret value on PUT | The exact spelling succeeded (200) with the same version. Changing only the spelling or supplying a new name without a value failed (422), leaving the definition and version unchanged. When both spellings had been configured separately as secret, omitting both values at their exact keys succeeded (200) with the same version. |
| Change an ordinary header to secret at the same key, omitting its value | PUT returned 200. Both the PUT response and GET still contained a `value` member and `secret: false`. The version stayed the same. |

These observations concern the webhook definition. The hooks were inactive, so
the experiments did not measure the header order, casing, or secret values a
recipient receives.

### Header validation order and omitted values on PUT

A collection POST on 2026-11-27 omitted both `topics` and the first header’s `value`. CMA returned 422 `ValidationFailed` with two `invalid_type` issues, at `headers[0].value` followed by `topics`. The reply had no version. A collection GET found no definition with the inactive probe’s unique name and URL.

On 2026-11-26 (UTC), raw requests on inactive Webhooks distinguished three
stages of validation. POST, specified-ID creation PUT, and existing-ID PUT
all returned 422 `ValidationFailed` for exact duplicate header keys. For
`X-A, X-B, X-B, X-A, X-B`, each repeat produced a `headers` issue with
path `["key"]`, in order X-B, X-A, X-B. When the request also supplied a nonempty Basic
username with a null password, the missing-password issue followed those
three. Adding a header with neither value nor secret flag produced only the
POST missing-value or PUT invalid-union issue at that index. GET found no
definition at the rejected specified ID. After each rejected existing-ID
PUT, GET returned the complete preceding public definition, including its
original name and version.

At version 0, PUT with version 7 still returned 422 for duplicate keys, a
header with neither member, an incomplete Basic pair, an oversized username,
or a lone null username on a stored nonempty pair. In contrast, a new header
with an explicit false flag and no value returned 409 `Conflict` with
version 7 and 422 with version 0. At the current version the error contained
a string at `details.errors` naming the missing header.

A PUT to an existing definition supplied `secret: false` and omitted the
values for two already-stored headers differing only in case, one secret and
one readable. PUT and GET returned the original flags, the readable value
and no secret value, still at version 0. Resending the original bytes kept
version 0. Changing just the secret bytes advanced to 1. This supports
retention of the hidden value.

With `secret: true` and an explicit JSON null `value`, POST returned 422
`ValidationFailed` with `invalid_type` at `headers.0.value` (expected
string, received null). PUT to an unused ID returned 500 `ServerError`, and
GET returned 404. PUT to an existing exact-name secret returned 200 and a
redacted descriptor. In a version control with the sibling header and name
held constant, sending null and then resending the original bytes both kept
version 0. Changing just the secret bytes advanced to 1. GET agreed with
each successful PUT.

The six disposable definitions were deleted (204) and GET confirmed absence
(404). The probes exercised CMA configuration without triggering deliveries.

A subsequent probe on another inactive Webhook checked explicit null with
different secret flags. On specified-ID creation PUT, an absent flag produced
422 `invalid_union`, while either Boolean produced 500. On an existing
definition, an absent flag produced 422 for stored and new names. With either
Boolean, both stored names (one readable, one secret) succeeded and a new name
returned 422 with the missing-value message. GET confirmed the original
definition after each rejection. Cleanup returned 204 and GET returned 404.

## Observation limits

These configuration probes do not establish every secret-header projection, filter
expression, transformation, signing rotation, concurrency outcome, or successful
delivery. The delivery behavior above comes from published webhook sources; it is not
inferred from an App Event subscription or the local test server.

[event-reference]: https://www.contentful.com/developers/docs/references/content-management-api/app-event-subscriptions/
[webhook-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/webhook.ts
[webhook-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/webhook.ts
[webhook-filters]: https://www.contentful.com/developers/docs/extensibility/webhooks/filters/
[webhook-headers]: https://www.contentful.com/developers/docs/extensibility/webhooks/headers/
[webhook-security]: https://www.contentful.com/developers/docs/references/content-management-api/webhook-security/
[webhook-overview]: https://www.contentful.com/developers/docs/extensibility/webhooks/overview/
[webhook-activity]: https://www.contentful.com/developers/docs/extensibility/webhooks/activity-log/
[webhook-calls]: https://www.contentful.com/developers/docs/references/content-management-api/webhook-calls/
[action-events]: https://www.contentful.com/developers/docs/extensibility/webhooks/action-events/
[bulk-operations]: https://www.contentful.com/blog/bulk-content-operations/
[live-events]: https://www.contentful.com/developers/changelog/monitor-your-events-in-the-live-events-dashboard/

## Transformation request vocabulary

The published [transformation contract](https://www.contentful.com/developers/docs/extensibility/webhooks/transformations/)
limits methods to `POST`, `GET`, `PUT`, `PATCH`, and `DELETE`. It lists management
JSON, JSON, and form URL-encoded content types, each with or without the exact
`; charset=utf-8` suffix. Omission selects the service default. The body can be
arbitrary JSON, and URLs may contain JSON-pointer templates.

These are published constraints, not live observations of every spelling or
server configuration. Future service additions require corresponding provider
write-validation updates; readable responses remain unrestricted.
