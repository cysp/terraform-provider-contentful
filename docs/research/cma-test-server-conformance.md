# CMA test-server conformance boundaries

The in-process CMA server models the observable contracts that the provider
depends on. It is not intended to implement Contentful internals. Default fake
behavior should be backed by the current CMA reference, the current first-party
client, or a sanitized direct observation. Deliberate fault injection must be
named as adversarial behavior rather than presented as CMA conformance.

The first-party source links below are pinned to reviewed
`contentful-management.js` v12.15.0 commit
`cc096a337f0e1db6114e8da645d69bb6eb90f11c` so that later upstream changes do
not silently alter this evidence.

## Behavior matrix

| Endpoint or behavior | Authoritative evidence | Default fake behavior | Resolved discrepancy or intentional boundary | Coverage |
| --- | --- | --- | --- | --- |
| App signing secret GET, PUT, DELETE | [CMA app signing secret reference](https://www.contentful.com/developers/docs/references/content-management-api/app-signing-secret/), pinned first-party [model](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/entities/app-signing-secret.ts#L7-L24) and [endpoints](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/app-signing-secret.ts#L10-L39), and the sanitized probe in [App signing secret CMA contract](app-signing-secret.md) | PUT derives and retains only the final-four redacted suffix. PUT and GET return that suffix, and DELETE returns no content. | The audited fake base64-encoded the complete submitted value into `redactedValue`; the prerequisite correction retains only the suffix. | Response tests prove that PUT and subsequent GET never return the complete submitted value. |
| Content type create, update, activate, deactivate, and version locking | [CMA content types reference](https://www.contentful.com/developers/docs/references/content-management-api/content-types/); the first-party client passes the fetched `sys.version` to [update and activation](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/content-type.ts#L82-L125) | Create returns version 1. Successful update and activation increment the version. Activation records the pre-activation version as `publishedVersion`. Deactivation removes publication fields and increments the version. Stale update or activation returns 409 `VersionMismatch`. Published fields must be omitted and activated before removal. | No high-confidence default-fake discrepancy remained in this pass; the existing lifecycle model was retained. | Lifecycle tests cover exact returned versions, stale activation, deactivation, pending drafts, and two-phase published-field deletion. |
| Entry update and publish version locking | [CMA entries reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/) requires the current version for update and publish; the first-party client sends `sys.version` for [both operations](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L185) | Create returns version 1. Update and publish compare the supplied version. Publish records the submitted draft version as `publishedVersion` and increments `sys.version`. | Publish previously ignored the supplied version and could let stale provider code pass; it now rejects a stale version. | Lifecycle coverage rejects stale publish versions and verifies the exact-version publication transition. |
| Taxonomy concept and concept-scheme PATCH and DELETE | [CMA taxonomy reference](https://www.contentful.com/developers/docs/references/content-management-api/taxonomy/), first-party concept [version headers](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/concept.ts#L38-L96), and the sanitized live matrix in [Taxonomy DELETE version behavior](taxonomy-delete-version.md) | Create returns version 1. Successful PATCH increments the version. PATCH and DELETE compare the supplied version. Taxonomy DELETE distinguishes missing or zero headers (422 `ValidationFailed`) from stale positive headers (409 `VersionMismatch`) and returns 204 only for the exact version. | No new discrepancy remained after the earlier live-backed taxonomy locking correction; endpoint-specific validation behavior was retained. | Tests preserve taxonomy-specific validation details instead of applying a generic conflict classification. |
| Offset collections for content types and entries | The [CMA overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/) defines collection responses with `sys`, `total`, `skip`, `limit`, and `items`; the first-party [`CollectionProp`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/common-types.ts#L566-L574) includes all five; content-type and entry list endpoints decode that type | The generated response types represent all five members while tolerating omitted pagination metadata. The fake always emits the documented metadata, validates skip and limit, applies filtering before total and pagination, echoes the requested skip and limit, and returns a stable ID-ordered page. | The fake omitted `skip` and `limit`, and map iteration made pages unstable. Runtime decoding remains deliberately tolerant because the first-party TypeScript type does not establish that its JavaScript client rejects a structurally incomplete response. | Conformance tests cover emitted metadata, filtering before pagination, invalid bounds, stable pages, and out-of-range skip echoing. Provider pagination uses `total` when present and otherwise continues until an empty page. ID ordering is a deterministic fake convention, not a claim about CMA's undocumented default ordering. |
| Organization team collection | The [User Management API endpoint](https://www.contentful.com/developers/docs/references/user-management-api/teams/get-all-teams-for-an-organization/) documents a paginated response with `sys`, `total`, `skip`, `limit`, and `items`. The [User Management API pagination contract](https://www.contentful.com/developers/docs/references/user-management-api/overview/#pagination) defines response `skip` as the requested offset and caps `limit` at 100. The first-party team [`getMany`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/team.ts#L25-L31) returns [`CollectionProp`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/common-types.ts#L566-L574). | The generated response type represents all five members while tolerating omitted pagination metadata. The fake validates the User Management API limit, reports the requested skip even when its slice start is beyond the collection, and returns a stable ID-ordered page. | The fake clamped returned `skip` and accepted limits through 1000. Provider reads remain tolerant of a missing `total` and continue until an empty page. | Client and provider tests cover decoding and pagination without metadata; fake tests cover emitted metadata, the 100-item maximum, stable pages, and out-of-range skip echoing. A team-specific out-of-range request could not be observed on the available disposable organization, so the fake's existing 200 response with an empty page is retained without claiming direct conformance for that status and item behavior. |
| Delivery API key create, update, and stale update | [CMA update reference](https://www.contentful.com/developers/docs/references/content-management-api/api-keys/update-a-delivery-api-key/), first-party [API-key adapter](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/api-key.ts#L26-L76), and the sanitized direct observation below | Create returns `sys.version: 0`; update with version 0 returns version 1; reusing version 0 returns 409 `Conflict` with a nonempty message. | The fake left the version unchanged after update, so a repeated stale version 0 was accepted instead of conflicting. | Live-backed tests cover version progression and stale locking. The `Conflict` classification is endpoint-specific and is not generalized to unrelated resources. |
| Entry response omissions | The [CMA entries reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/) documents that unset fields can be absent; the first-party entry adapter decodes full entry responses for [GET and list](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L27-L83) and [update and publish](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L185). These sources do not establish that CMA normally omits configured, nonempty fields from mutation responses. | Default responses preserve stored fields. `WithOmittedEntryMutationResponseFields` is an explicitly adversarial provider-test mode and affects mutation responses only; later GET and list responses return stored fields. | The adversarial option previously stripped fields from GET and list responses too, turning fault injection into an invented default read behavior. | Mutation omission is defensive coverage, not a conformance claim. GET and list behavior remains distinct from the adversarial mutation mode. |
| Environment creation readiness | The [CMA environments reference](https://www.contentful.com/developers/docs/references/content-management-api/environments/) requires callers to query the environment after create and defines queued, in-progress, ready, and failed states. The first-party environment model exposes the returned [`sys.status` link](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/entities/environment.ts#L9-L25). | Registered and newly created environments are immediately ready in the generic fake. Provider polling is covered offline with purpose-built handlers that return queued or in-progress statuses before ready. | Generic asynchronous timing remains intentionally unmodelled; no scheduler or timing transformation was added. | Purpose-built provider handlers cover the observable polling and timeout transitions without making the generic fake nondeterministic. |
| 404, version conflict, and rate-limit errors | [Contentful error reference](https://www.contentful.com/developers/docs/references/errors/) requires `sys.type: Error`, a code in `sys.id`, and a nonempty `message`; the [CMA overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/) defines rate-limit headers and 429 behavior | Generic missing resources return 404 `NotFound`; version-locked resources return their endpoint-appropriate 409 classification; the limiter returns 429 `RateLimitExceeded` with second-limit, remaining, and reset headers. | Some generic 404 and version-conflict helpers could omit `message`; default messages are now structural invariants. | Structural tests cover type, ID, nonempty message, status, and rate-limit headers. |

## Sanitized direct observations

A disposable Delivery API key was created in an existing disposable test space,
updated once, sent one stale update, and deleted. Only status codes, `sys.type`,
`sys.version`, response-member presence, and the error classification were
printed. The management token, space and environment IDs, API-key ID, delivery
and preview tokens, and response bodies were neither printed nor retained.

| Request | Status | Structural observation |
| --- | ---: | --- |
| `POST /spaces/{space}/api_keys` | 201 | `sys.type: ApiKey`, `sys.version: 0`; delivery and preview token members were present but not read or recorded |
| `PUT /spaces/{space}/api_keys/{key}` with version 0 | 200 | `sys.type: ApiKey`, `sys.version: 1`; token members remained present but were not read or recorded |
| Repeated PUT with stale version 0 | 409 | `sys.type: Error`, `sys.id: Conflict`, nonempty message, no details member |
| `DELETE /spaces/{space}/api_keys/{key}` | 204 | Empty response; the disposable key was removed |

A read-only `GET /spaces/{space}/environments/{environment}/entries?skip=999999&limit=2`
probe returned 200
with `sys.type: Array`, echoed `skip: 999999` and `limit: 2`, and returned an
empty item list. No entry data or identifiers were printed or retained. This
establishes that the fake must echo an out-of-range skip rather than clamp it to
the collection length.

A read-only `GET /organizations/{organization}/teams?skip=999999&limit=1` probe
used bearer authentication against a disposable organization. Contentful
returned 403 with `sys.type: Error` and `sys.id: FeatureNotEnabled`, so the probe
did not establish the team endpoint's out-of-range success behavior. The
management token, organization ID, request ID, message, and response body were
neither printed nor retained. The published User Management API contract still
establishes that a successful collection response must report the requested
skip.

No retained live utility is needed: these were one-time, sanitized structural
probes, and committing a credential-aware script would add maintenance and
misuse risk without improving normal offline coverage.

## Intentionally unmodelled behavior

- The generic environment fake does not schedule queued, in-progress, ready, or
  failed transitions. Focused provider tests inject the transitions relevant to
  polling and timeout behavior.
- The fake does not reproduce Contentful's internal ordering for list endpoints.
  It uses deterministic ID order where pagination tests require stable pages.
- Cursor pagination is not implemented. The generic fake also does not
  interpret arbitrary entry query filters or configurable order. An exact HTTP
  request test covers the provider's forwarding of those values; fake lifecycle
  tests do not claim their result-set semantics.
- Rate limiting models the documented second-level contract only. It does not
  reproduce distributed quota accounting or retry jitter.
- Mutation-response field omission remains an adversarial mode because current
  primary evidence does not establish it as normal behavior for configured,
  nonempty entry fields.
- Eventual consistency beyond environment readiness is not simulated. A focused
  handler should model a concrete observed transition before it is added to the
  generic fake.

These boundaries keep offline tests stable while preventing the fake from
silently validating collection, locking, redaction, or error-shape assumptions
that contradict the observable CMA contract.
