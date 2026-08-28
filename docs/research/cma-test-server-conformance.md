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
| Space Enablements PUT | The [CMA reference](https://www.contentful.com/developers/docs/references/content-management-api/space-enablements/) and [request-values probe](space-enablement-values.md) establish the paired-member rule and failed-request nonmutation | The fake requires both paired members with equal values before storing the document | Exact validation details remain unmodelled, and the provider keeps all enablements independently Optional+Computed | Handler and mocked lifecycle tests cover rejection, nonmutation, and recovery. |
| App signing secret GET, PUT, DELETE | [CMA app signing secret reference](https://www.contentful.com/developers/docs/references/content-management-api/app-signing-secret/), pinned first-party [model](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/entities/app-signing-secret.ts#L7-L24) and [endpoints](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/app-signing-secret.ts#L10-L39), and the sanitized probe in [App signing secret CMA contract](app-signing-secret.md) | PUT derives and retains only the final-four redacted suffix. PUT and GET return that suffix, and DELETE returns no content. | The audited fake base64-encoded the complete submitted value into `redactedValue`; the prerequisite correction retains only the suffix. | Response tests prove that PUT and subsequent GET never return the complete submitted value. |
| Content type create, update, activate, deactivate, and version locking | [CMA content types reference](https://www.contentful.com/developers/docs/references/content-management-api/content-types/); the first-party client passes the fetched `sys.version` to [update and activation](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/content-type.ts#L82-L125) | Create returns version 1. Successful update and activation increment the version. Activation records the pre-activation version as `publishedVersion`. Deactivation removes publication fields and increments the version. Stale update or activation returns 409 `VersionMismatch`. Published fields must be omitted and activated before removal. | No high-confidence default-fake discrepancy remained in this pass; the existing lifecycle model was retained. | Lifecycle tests cover exact returned versions, stale activation, deactivation, pending drafts, and two-phase published-field deletion. |
| Entry update, publish, unpublish, and version locking | [CMA entries reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/) requires the current version for update and publish; the first-party client sends `sys.version` for [both operations](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L185) and performs whole-Entry unpublish as an unversioned DELETE returning an Entry; sanitized probes observed normal publish arithmetic and [unpublish advancing beyond a pending draft while removing `publishedVersion`](entry-unpublish-version.md) | Create returns version 1. Update and publish compare the supplied version. Publish records the submitted draft version as `publishedVersion` and increments `sys.version`. Unpublish increments `sys.version`, removes `publishedVersion`, and returns the resulting Entry. | Exact version locking and version-advancing unpublish are normal fake behavior. Malformed alternatives remain explicit adversarial fault injection. | Lifecycle coverage rejects stale publish versions, verifies the live-observed publish and unpublish transitions, revokes failed-publication recovery after external unpublish, scopes recovery to provider-written drafts, and leaves external or imported pending drafts untouched. A coherent higher `sys.version` response is adversarial fault injection for response handling, not a claim about observed CMA behavior; it is representable with a warning but must satisfy exact field ownership. |
| Taxonomy concept and concept-scheme PATCH and DELETE | [CMA taxonomy reference](https://www.contentful.com/developers/docs/references/content-management-api/taxonomy/), first-party concept [version headers](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/concept.ts#L38-L96), and the sanitized live matrix in [Taxonomy version behavior](taxonomy-version.md) | Create returns version 1. Successful PATCH increments the version. PATCH and DELETE compare the supplied version. Taxonomy DELETE distinguishes missing or zero headers (422 `ValidationFailed`) from stale positive headers (409 `VersionMismatch`) and returns 204 only for the exact version. | No new discrepancy remained after the earlier live-backed taxonomy locking correction; endpoint-specific validation behavior was retained. | Tests preserve taxonomy-specific validation details instead of applying a generic conflict classification. |
| Offset collections for content types and entries | The [CMA overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/) defines collection responses with `sys`, `total`, `skip`, `limit`, and `items`; the first-party [`CollectionProp`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/common-types.ts#L566-L574) includes all five; content-type and entry list endpoints decode that type | The generated response types represent all five members while tolerating omitted pagination metadata. The fake always emits the documented metadata, validates skip and limit, applies filtering before total and pagination, echoes the requested skip and limit, and returns a stable ID-ordered page. | The fake omitted `skip` and `limit`, and map iteration made pages unstable. Runtime decoding remains deliberately tolerant because the first-party TypeScript type does not establish that its JavaScript client rejects a structurally incomplete response. | Conformance tests cover emitted metadata, filtering before pagination, invalid bounds, stable pages, and out-of-range skip echoing. Provider pagination uses `total` when present and otherwise continues until an empty page. ID ordering is a deterministic fake convention, not a claim about CMA's undocumented default ordering. |
| Organization team collection | The [User Management API endpoint](https://www.contentful.com/developers/docs/references/user-management-api/teams/get-all-teams-for-an-organization/) documents a paginated response with `sys`, `total`, `skip`, `limit`, and `items`. The [User Management API pagination contract](https://www.contentful.com/developers/docs/references/user-management-api/overview/#pagination) defines response `skip` as the requested offset and caps `limit` at 100. The first-party team [`getMany`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/team.ts#L25-L31) returns [`CollectionProp`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/common-types.ts#L566-L574). | The generated response type represents all five members while tolerating omitted pagination metadata. The fake validates the User Management API limit, reports the requested skip even when its slice start is beyond the collection, and returns a stable ID-ordered page. | The fake clamped returned `skip` and accepted limits through 1000. Provider reads remain tolerant of a missing `total` and continue until an empty page. | Client and provider tests cover decoding and pagination without metadata; fake tests cover emitted metadata, the 100-item maximum, stable pages, and out-of-range skip echoing. A team-specific out-of-range request could not be observed on the available disposable organization, so the fake's existing 200 response with an empty page is retained without claiming direct conformance for that status and item behavior. |
| Delivery API key create, update, and stale update | [CMA update reference](https://www.contentful.com/developers/docs/references/content-management-api/api-keys/update-a-delivery-api-key/), first-party [API-key adapter](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/api-key.ts#L26-L76), and the sanitized direct observation below | Create returns `sys.version: 0`; update with version 0 returns version 1; reusing version 0 returns 409 `Conflict` with a nonempty message. | The fake left the version unchanged after update, so a repeated stale version 0 was accepted instead of conflicting. | Live-backed tests cover version progression and stale locking. The `Conflict` classification is endpoint-specific and is not generalized to unrelated resources. |
| Entry response omissions | The [CMA entries reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/) documents that all empty fields are omitted from responses and that an entry with no set fields omits the top-level `fields` member. The first-party entry adapter decodes full entry responses for [GET and list](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L27-L83) and [update and publish](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L185). The sanitized [Entry null and omission probe](entry-null-and-omission.md) observed Create and Update accepting a raw JSON-null field and omitting it from mutation, publish, and GET responses, while localized objects containing null remained present. Applying the documented empty-array projection to every Entry response endpoint remains an inference from their common representation. | The fake stores requests unchanged and projects raw JSON-null fields and localized fields containing only empty arrays out of GET, list, and mutation responses, omitting the top-level member when none remain. Localized objects containing null remain ordinary response data. `WithOmittedEntryMutationResponseFields` is explicitly adversarial complete mutation-response omission; later GET and list responses retain stored nonempty fields. | The provider treats Terraform null as request omission and raw JSON null as a sent value. It restores either exact configured representation when CMA omits it, but a present response value wins; localized null receives no fallback. A whole-member omission can restore a known empty plan, but it contradicts a plan containing meaningful fields. Individual missing keys receive only the narrow Terraform-null, raw JSON-null, and all-empty-array fallbacks. | Fake response tests cover unchanged storage and GET, list, PUT, and publish projection for raw JSON null, localized null, and empty arrays. Provider tests cover exact request/response restoration for Terraform null, raw JSON null, empty maps, and all-empty arrays; response-value precedence; rejection of adversarial nonempty whole-member omission; rejection of mixed-null fallback; and consistency errors for partial maps missing nonempty fields. Mocked lifecycle coverage distinguishes Terraform null, encoded JSON null, and deletion; live `TestAccEntryResourceMissingFields` covers configured empty arrays. |
| Environment creation readiness | The [CMA environments reference](https://www.contentful.com/developers/docs/references/content-management-api/environments/) requires callers to query the environment after create and defines `queued`, `inProgress`, `ready`, and `failed` states. The first-party environment model exposes the returned [`sys.status` link](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/entities/environment.ts#L9-L25). | Registered and newly created environments are immediately ready in the generic fake. Focused provider handlers return the exact status values needed to exercise readiness behavior. | Generic asynchronous timing remains intentionally unmodelled; no scheduler or timing transformation was added. The provider stops immediately on the documented terminal `failed` status, while unrecognized future values remain pollable. | Unit tests cover all four documented statuses and an unrecognized future value. Mocked acceptance covers `queued` to `ready`, terminal `failed` after one request, and `inProgress` remaining pollable until its configured timeout. |
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
| `PUT /spaces/{space}/api_keys/{key}` with `X-Contentful-Version: 0` | 200 | `sys.type: ApiKey`, `sys.version: 1`; token members remained present but were not read or recorded |
| Repeated PUT with stale `X-Contentful-Version: 0` | 409 | `sys.type: Error`, `sys.id: Conflict`, nonempty message, no details member |
| `DELETE /spaces/{space}/api_keys/{key}` | 204 | Empty response; the disposable key was removed |

A disposable Content Type and Entry were then used to probe the successful
Entry publication relationship. The bearer token and generated resource IDs
were not printed. Only status codes and the selected version fields below were
printed; complete response bodies were not retained.

| Request | Relevant header | Status | Structural observation |
| --- | --- | ---: | --- |
| `PUT /spaces/{space}/environments/{environment}/content_types/{content-type}` | — | 201 | `sys.version: 1` |
| `PUT /spaces/{space}/environments/{environment}/content_types/{content-type}/published` | `X-Contentful-Version: 1` | 200 | The disposable Content Type was activated |
| `PUT /spaces/{space}/environments/{environment}/entries/{entry}` | `X-Contentful-Content-Type: {content-type}` | 201 | `sys.version: 1` |
| `PUT /spaces/{space}/environments/{environment}/entries/{entry}/published` | `X-Contentful-Version: 1` | 200 | `sys.version: 2`, `sys.publishedVersion: 1` |

A later disposable [whole-Entry unpublish probe](entry-unpublish-version.md)
retained only structural observations. Unpublish returned HTTP 200 with an Entry
whose `version` advanced by one beyond the pending draft and whose
`publishedVersion` was absent; a subsequent GET returned the same tuple.

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

- The generic environment fake does not schedule `queued`, `inProgress`,
  `ready`, or `failed` transitions. Focused provider tests inject a
  `queued`-to-`ready` transition and static `inProgress` and `failed` responses
  for polling, timeout, and terminal-failure behavior.
- The fake does not reproduce Contentful's internal ordering for list endpoints.
  It uses deterministic ID order where pagination tests require stable pages.
- Cursor pagination is not implemented. The generic fake also does not
  interpret arbitrary entry query filters or configurable order. An exact HTTP
  request test covers the provider's forwarding of those values; fake lifecycle
  tests do not claim their result-set semantics.
- Rate limiting models the documented second-level contract only. It does not
  reproduce distributed quota accounting or retry jitter.
- Complete mutation-response field omission remains an adversarial mode because
  current primary evidence does not establish it as normal behavior for
  configured, nonempty entry fields. Other unprobed empty representations are
  not projected by the fake or treated as equivalent by the provider.
- Entry unpublish models the observed whole-Entry transition from a published
  Entry with a pending draft. Repeated unpublish of an already-unpublished Entry
  and locale-based unpublish remain unprobed and are not treated as conformance
  claims.
- Eventual consistency beyond environment readiness is not simulated. A focused
  handler should model a concrete observed transition before it is added to the
  generic fake.

These boundaries keep offline tests stable while preventing the fake from
silently validating collection, locking, redaction, or error-shape assumptions
that contradict the observable CMA contract.
