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

`W = /spaces/{space_id}/webhook_definitions/{webhook_id}` below.

WebhookDefinition is space-scoped, has its own ID, and needs no app installation. It
supports headers and secret headers, basic auth, filters, transformations, `active:
false`, and wildcard topics such as `*.*`, `Entry.*`, and `*.save`. Those are different
contracts from AppEventSubscription's parent-keyed HTTP target and exact topics. A
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
scope. Its PUT takes `value`; reads return `redactedValue`. It was not changed during
the experiments. [Webhook adapter][webhook-sdk], [Webhook security][webhook-security].

The SDK still contains GET/PUT/DELETE for the space singleton
`/webhook_settings/retry_policy` with `maxRetries`, but marks it deprecated because its
EAP ended and removal is planned for the next major version. This is legacy source
surface, not evidence of a supported new resource or account entitlement. [Adapter
deprecations][webhook-sdk].

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

## Observation limits

These configuration probes do not establish every secret-header projection, filter
expression, transformation, signing rotation, concurrency outcome, or successful
delivery. Delivery guarantees above come from published webhook sources; they are not
inferred from an App Event subscription or the local test server.

[event-reference]: https://www.contentful.com/developers/docs/references/content-management-api/app-event-subscriptions/
[webhook-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/webhook.ts
[webhook-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/webhook.ts
[webhook-filters]: https://www.contentful.com/developers/docs/extensibility/webhooks/filters/
[webhook-security]: https://www.contentful.com/developers/docs/references/content-management-api/webhook-security/
[webhook-overview]: https://www.contentful.com/developers/docs/extensibility/webhooks/overview/
[webhook-activity]: https://www.contentful.com/developers/docs/extensibility/webhooks/activity-log/
[webhook-calls]: https://www.contentful.com/developers/docs/references/content-management-api/webhook-calls/
[action-events]: https://www.contentful.com/developers/docs/extensibility/webhooks/action-events/
[bulk-operations]: https://www.contentful.com/blog/bulk-content-operations/
[live-events]: https://www.contentful.com/developers/changelog/monitor-your-events-in-the-live-events-dashboard/
