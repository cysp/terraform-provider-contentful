# CMA test-server conformance boundaries

The in-process CMA server models the observable contracts that the provider
depends on. It is not intended to implement Contentful internals. Default fake
behavior should be backed by the current CMA reference, the current first-party
client, or a sanitized direct observation. Deliberate fault injection must be
named as adversarial behavior rather than presented as CMA conformance.

The first-party source links below are pinned to reviewed
`contentful-management.js` commit
`cc096a337f0e1db6114e8da645d69bb6eb90f11c` so that later upstream changes do
not silently alter this evidence.

## Behavior matrix

| Endpoint or behavior | Authoritative evidence | Default fake behavior | Coverage and boundary |
| --- | --- | --- | --- |
| App signing secret GET, PUT, DELETE | [CMA app signing secret reference](https://www.contentful.com/developers/docs/references/content-management-api/app-signing-secret/), pinned first-party [model](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/entities/app-signing-secret.ts#L7-L24) and [endpoints](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/app-signing-secret.ts#L10-L39), and the sanitized probe in [App signing secret CMA contract](app-signing-secret.md) | PUT derives and retains only the final-four redacted suffix. PUT and GET return that suffix, and DELETE returns no content. | The implementation retains only the suffix; response tests prove that PUT and subsequent GET never return the complete submitted value. |
| Content type create, update, activate, deactivate, and version locking | [CMA content types reference](https://www.contentful.com/developers/docs/references/content-management-api/content-types/); the first-party client passes the fetched `sys.version` to [update and activation](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/content-type.ts#L82-L125) | Create returns version 1. Successful update and activation increment the version. Activation records the pre-activation version as `publishedVersion`. Deactivation removes publication fields and increments the version. Stale update or activation returns 409 `VersionMismatch`. Published fields must be omitted and activated before removal. | Lifecycle tests cover exact returned versions, stale activation, deactivation, pending drafts, and two-phase published-field deletion. |
| Entry update and publish version locking | [CMA entries reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/) requires the current version for update and publish; the first-party client sends `sys.version` for [both operations](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L185) | Create returns version 1. Update and publish compare the supplied version. Publish records the submitted draft version as `publishedVersion` and increments `sys.version`. | Lifecycle coverage rejects stale publish versions and verifies the exact-version publication transition. |
| Taxonomy concept and concept-scheme PATCH and DELETE | [CMA taxonomy reference](https://www.contentful.com/developers/docs/references/content-management-api/taxonomy/), first-party concept [version headers](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/concept.ts#L38-L96), and the sanitized live matrix in [Taxonomy DELETE version behavior](taxonomy-delete-version.md) | Create returns version 1. Successful PATCH increments the version. PATCH and DELETE compare the supplied version. Taxonomy DELETE distinguishes missing or zero headers (422 `ValidationFailed`) from stale positive headers (409 `VersionMismatch`) and returns 204 only for the exact version. | Tests preserve taxonomy-specific validation details instead of applying a generic conflict classification. |
| Offset collections for content types and entries | The [CMA overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/) defines collection responses with `sys`, `total`, `skip`, `limit`, and `items`; the first-party [`CollectionProp`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/common-types.ts#L566-L574) requires all five; content-type and entry list endpoints decode that type | The generated response types require all five members. The fake validates skip and limit, applies filtering before total and pagination, echoes the requested skip and limit, and returns a stable ID-ordered page. | Conformance tests cover required metadata, filtering before pagination, invalid bounds, stable pages, and out-of-range skip echoing. ID ordering is a deterministic fake convention, not a claim about CMA's undocumented default ordering. |
| Delivery API key create, update, and stale update | [CMA update reference](https://www.contentful.com/developers/docs/references/content-management-api/api-keys/update-a-delivery-api-key/), first-party [API-key adapter](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/api-key.ts#L26-L76), and the sanitized direct observation below | Create returns `sys.version: 0`; update with version 0 returns version 1; reusing version 0 returns 409 `Conflict` with a nonempty message. | Live-backed tests cover version progression and stale locking. The `Conflict` classification is endpoint-specific and is not generalized to unrelated resources. |
| Entry response omissions | The [CMA entries reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/) documents that unset fields can be absent; the first-party entry adapter decodes full entry responses for [GET and list](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L27-L83) and [update and publish](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L185). These sources do not establish that CMA normally omits configured, nonempty fields from mutation responses. | Default responses preserve stored fields. `WithOmittedEntryMutationResponseFields` is an explicitly adversarial provider-test mode and affects mutation responses only; later GET and list responses return stored fields. | Mutation omission is defensive coverage, not a conformance claim. GET and list behavior remains distinct from the adversarial mutation mode. |
| Environment creation readiness | [CMA environments reference](https://www.contentful.com/developers/docs/references/content-management-api/environments/) says creation is asynchronous and callers must poll until ready. The first-party environment model exposes the returned [`sys.status` link](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/entities/environment.ts#L9-L25). | Registered and newly created environments are immediately ready in the generic fake. Provider polling is covered offline with purpose-built handlers that return queued or in-progress statuses before ready. | Intentionally unmodelled in the generic fake. Adding a scheduler or timing model would not add independent evidence and would make deterministic tests less clear. |
| 404, version conflict, and rate-limit errors | [Contentful error reference](https://www.contentful.com/developers/docs/references/errors/) requires `sys.type: Error`, a code in `sys.id`, and a nonempty `message`; the [CMA overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/) defines rate-limit headers and 429 behavior | Generic missing resources return 404 `NotFound`; version-locked resources return their endpoint-appropriate 409 classification; the limiter returns 429 `RateLimitExceeded` with second-limit, remaining, and reset headers. | Structural tests cover type, ID, nonempty message, status, and rate-limit headers. |

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

A read-only entry-list probe requested `skip=999999&limit=2`. CMA returned 200
with `sys.type: Array`, echoed `skip: 999999` and `limit: 2`, and returned an
empty item list. No entry data or identifiers were printed or retained. This
establishes that the fake must echo an out-of-range skip rather than clamp it to
the collection length.

No retained live utility is needed: both ambiguities required a one-time,
sanitized structural probe, and committing a credential-aware script would add
maintenance and misuse risk without improving normal offline coverage.

## Intentionally unmodelled behavior

- The generic environment fake does not schedule queued, in-progress, ready, or
  failed transitions. Focused provider tests inject the transitions relevant to
  polling and timeout behavior.
- The fake does not reproduce Contentful's internal ordering for list endpoints.
  It uses deterministic ID order where pagination tests require stable pages.
- Cursor pagination and every endpoint-specific search/order operator are not
  implemented unless a provider path uses them.
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
