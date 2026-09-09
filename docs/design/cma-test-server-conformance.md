# CMA test-server conformance boundaries

Use this reference when changing the in-process CMA server or a test that relies
on its behavior. The server models the observable contracts that the provider
depends on. It is not intended to implement Contentful internals. Default fake
behavior should be backed by the cited CMA reference, a pinned first-party
client, or a sanitized direct observation. Retain the evidence date and
limitations when reusing an observation. Deliberate fault injection must be named
as adversarial behavior rather than presented as CMA conformance.

The first-party source links below are pinned to reviewed
`contentful-management.js` v12.15.0 commit
`cc096a337f0e1db6114e8da645d69bb6eb90f11c` so that later upstream changes do
not silently alter this evidence.

## Endpoint contracts

Each endpoint separates external evidence, default fake behavior, the provider
contract, and test coverage. A coverage statement describes what the tests
exercise; it does not independently establish CMA conformance.

### Space Enablements

**Evidence:** The [CMA
reference](https://www.contentful.com/developers/docs/references/content-management-api/space-enablements/)
and [request-values probe](../research/space-enablements.md) establish the paired-member rule
and failed-request nonmutation.

**Default fake behavior:** The fake requires both paired members with equal values
before storing the document.

**Provider boundary:** Exact validation details remain unmodelled, and the provider
keeps all enablements independently Optional+Computed.

**Coverage:** Handler and mocked lifecycle tests cover rejection, nonmutation, and
recovery.

### App signing secret

**Evidence:** [CMA app signing secret
reference](https://www.contentful.com/developers/docs/references/content-management-api/app-signing-secret/),
pinned first-party
[model](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/entities/app-signing-secret.ts#L7-L24)
and
[endpoints](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/app-signing-secret.ts#L10-L39),
and the sanitized probe in [App signing secret CMA contract](../research/app-framework/request-signing.md#direct-cma-observation).

**Default fake behavior:** PUT derives and retains only the final-four redacted suffix.
PUT and GET return that suffix, and DELETE returns no content.

**Provider boundary:** Redaction exposes only the final-four suffix; the complete
submitted value is not response data.

**Coverage:** Response tests prove that PUT and subsequent GET never return the complete
submitted value.

### Webhook topics

**Evidence:** Contentful's [Webhook configuration
documentation](https://www.contentful.com/developers/docs/extensibility/webhooks/configure-webhook/#topics)
requires at least one topic. Sanitized direct Create and Update observations established
that omitted and empty topics return 422 `ValidationFailed` without creating or changing
a Webhook.

**Default fake behavior:** The fake returns the observed missing-topic and empty-topic
validation details before storing either Create or Update requests.

**Provider boundary:** The provider requires a non-empty list while still deferring
unknown values during validation. Successful Create status is not established by these
topics-validation tests.

**Coverage:** Raw HTTP tests cover exact POST and PUT paths, status and response shapes,
and failed-request nonmutation. Mocked Terraform lifecycle tests cover local omission
and emptiness rejection before mutation plus one-topic acceptance.

### Webhook Basic password

**Evidence:** Contentful's [Webhook configuration
documentation](https://www.contentful.com/developers/docs/extensibility/webhooks/configure-webhook/#http-basic-authentication)
defines paired username/password writes and paired-null removal. Sanitized direct
Create, Update, and GET observations established that CMA accepts the password but omits
`httpBasicPassword` from every response while retaining `httpBasicUsername`.

**Default fake behavior:** The fake stores the submitted password so later request
behavior remains testable, but omits it from Create, Update, and GET responses.

**Provider boundary:** The provider restores planned or prior state only when the
response property is absent; explicit JSON null remains distinct.

**Coverage:** Raw-CMA JSON and Terraform lifecycle tests cover the omission, refresh,
update, rotation, removal, import, and `ignore_changes`.

### Content Type lifecycle and locking

**Evidence:** [CMA content types
reference](https://www.contentful.com/developers/docs/references/content-management-api/content-types/);
the first-party client passes the fetched `sys.version` to [update and
activation](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/content-type.ts#L82-L125).

**Default fake behavior:** Create returns version 1. Successful update and activation
increment the version. Activation records the pre-activation version as
`publishedVersion`. Deactivation removes publication fields and increments the version.
Stale update or activation returns 409 `VersionMismatch`. Published fields must be
omitted and activated before removal.

**Provider boundary:** Version 1 and `V+1` remain normal fake behavior, not universal
CMA guarantees. Focused HTTP adapters model arbitrary positive Create and Update
response versions, higher coherent activation responses, response loss, and
contradictory tuples without changing the normal fake. Provider activation authority is
private-state provenance, not fake state.

**Coverage:** Lifecycle tests cover exact-version recovery, refresh/no-refresh,
immediate stale-marker revocation, response loss, external transitions, 429 no-replay,
arbitrary positive draft versions, high coherent confirmation, raw requests,
deactivation, and two-phase published-field deletion.

### Entry lifecycle and locking

**Evidence:** The [CMA entries
reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/)
requires the current version for update and publish; the first-party client sends
`sys.version` for [both
operations](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L185)
and performs whole-Entry unpublish and delete without preconditions. Sanitized probes
established the [member-PUT request-selection
boundary](../research/entry-and-content-type-put-headers.md), [unpublish version
behavior](../research/entry-lifecycle.md#publish-update-and-unpublish), and the [destroy
lifecycle](../research/entry-lifecycle.md#unpublish-and-delete-preconditions).

**Default fake behavior:** Member PUT creates an absent Entry at version 1 when Content
Type is present, regardless of Version; without Content Type it returns 400 and leaves
the target absent. An existing Entry updates only with an exact Version, with or without
Content Type; a missing or stale Version returns 409 `VersionMismatch` and leaves it
unchanged. Publish and unpublish advance `sys.version`; unpublish removes
`publishedVersion`. Whole-Entry unpublish and delete ignore version and ETag headers.
Repeated unpublish returns 400 `Not published`; published delete returns 400 `Cannot
delete published`; other valid deletes remove the Entry.

**Provider boundary:** Version 1 and `V+1` remain normal fake behavior, not universal
CMA guarantees. Focused HTTP adapters model arbitrary positive Create and Update
response versions, higher coherent publication responses, response loss, and
contradictory tuples. Provider publication authority is a private integer recorded only
after validating the provider-authored draft response.

**Coverage:** Tests cover the direct member-PUT header matrix and stale/no-refresh
non-recreation, exact-version recovery, immediate stale-marker revocation, response
loss, external transitions, 429 no-replay, arbitrary positive draft versions, high
coherent confirmation, raw requests, and unconditional whole-Entry destroy sequencing.

### Taxonomy locking

**Evidence:** [CMA taxonomy
reference](https://www.contentful.com/developers/docs/references/content-management-api/taxonomy/),
first-party concept [version
headers](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/concept.ts#L38-L96),
and the sanitized live matrix in [Taxonomy version behavior](../research/taxonomy.md).

**Default fake behavior:** Create returns version 1. Successful PATCH increments the
version. PATCH and DELETE compare the supplied version. Taxonomy DELETE distinguishes
missing or zero headers (422 `ValidationFailed`) from stale positive headers (409
`VersionMismatch`) and returns 204 only for the exact version.

**Provider boundary:** Version validation remains endpoint-specific; missing or zero
headers and stale positive versions have different error classifications.

**Coverage:** Tests preserve taxonomy-specific validation details instead of applying a
generic conflict classification.

### Content Type and Entry collections

**Evidence:** The [CMA
overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/)
defines collection responses with `sys`, `total`, `skip`, `limit`, and `items`; the
first-party
[`CollectionProp`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/common-types.ts#L566-L574)
includes all five; content-type and entry list endpoints decode that type.

**Default fake behavior:** The generated response types represent all five members while
tolerating omitted pagination metadata. The fake always emits the documented metadata,
validates skip and limit, applies filtering before total and pagination, echoes the
requested skip and limit, and returns a stable ID-ordered page.

**Provider boundary:** Runtime decoding remains deliberately tolerant because the
first-party TypeScript type does not establish that its JavaScript client rejects a
structurally incomplete response.

**Coverage:** Conformance tests cover emitted metadata, filtering before pagination,
invalid bounds, stable pages, and out-of-range skip echoing. Provider pagination uses
`total` when present and otherwise continues until an empty page. ID ordering is a
deterministic fake convention, not a claim about CMA's undocumented default ordering.

### Organization teams

**Evidence:** The [User Management API
endpoint](https://www.contentful.com/developers/docs/references/user-management-api/teams/get-all-teams-for-an-organization/)
documents a paginated response with `sys`, `total`, `skip`, `limit`, and `items`. The
[User Management API pagination
contract](https://www.contentful.com/developers/docs/references/user-management-api/overview/#pagination)
defines response `skip` as the requested offset and caps `limit` at 100. The first-party
team
[`getMany`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/team.ts#L25-L31)
returns
[`CollectionProp`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/common-types.ts#L566-L574).

**Default fake behavior:** The generated response type represents all five members while
tolerating omitted pagination metadata. The fake validates the User Management API
limit, reports the requested skip even when its slice start is beyond the collection,
and returns a stable ID-ordered page.

**Provider boundary:** Provider reads remain tolerant of a missing `total` and continue
until an empty page.

**Coverage:** Client and provider tests cover decoding and pagination without metadata;
fake tests cover emitted metadata, the 100-item maximum, stable pages, and out-of-range
skip echoing. A successful team-specific out-of-range response was not established by the
retained evidence; the fake's 200 empty-page response is a fixture convention.

### Delivery API key locking

**Evidence:** [CMA update
reference](https://www.contentful.com/developers/docs/references/content-management-api/api-keys/update-a-delivery-api-key/),
first-party [API-key
adapter](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/api-key.ts#L26-L76),
and the [sanitized direct observation](../research/delivery-api-keys.md#versioning-observations).

**Default fake behavior:** Create returns `sys.version: 0`; update with version 0
returns version 1; reusing version 0 returns 409 `Conflict` with a nonempty message.

**Provider boundary:** An accepted update advances the version, so the same request
version cannot authorize a second update.

**Coverage:** Live-backed tests cover version progression and stale locking. The
`Conflict` classification is endpoint-specific and is not generalized to unrelated
resources.

### Entry response omissions

**Evidence:** The [CMA entries
reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/)
documents that the Get an entry response omits all empty fields and omits the top-level
`fields` member when no fields are set. The first-party entry adapter decodes full
entry responses for [GET and
list](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L27-L83)
and [update and
publish](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L185).
The sanitized [Entry null and omission probe](../research/entry-fields.md) observed
Create and Update accepting a raw JSON-null field and omitting it from mutation,
publish, and GET responses, while localized objects containing null remained present.
Applying the documented empty-array projection to every Entry response endpoint remains
an inference from their common representation.

**Default fake behavior:** The fake stores requests unchanged and projects raw JSON-null
fields and localized fields containing only empty arrays out of GET, list, and mutation
responses, omitting the top-level member when none remain. Localized objects containing
null remain ordinary response data. `WithOmittedEntryMutationResponseFields` is
explicitly adversarial complete mutation-response omission; later GET and list responses
retain stored nonempty fields.

**Provider boundary:** The provider treats Terraform null as request omission and raw
JSON null as a sent value. It restores either exact configured representation when CMA
omits it, but a present response value wins; localized null receives no fallback. A
whole-member omission can restore a known empty plan, but it contradicts a plan
containing meaningful fields. Individual missing keys receive only the narrow
Terraform-null, raw JSON-null, and all-empty-array fallbacks.

**Coverage:** Fake response tests cover unchanged storage and GET, list, PUT, and
publish projection for raw JSON null, localized null, and empty arrays. Provider tests
cover exact request/response restoration for Terraform null, raw JSON null, empty maps,
and all-empty arrays; response-value precedence; rejection of adversarial nonempty
whole-member omission; rejection of mixed-null fallback; and consistency errors for
partial maps missing nonempty fields. Mocked lifecycle coverage distinguishes Terraform
null, encoded JSON null, and deletion; live `TestAccEntryResourceMissingFields` covers
configured empty arrays.

### Environment creation readiness

**Evidence:** The [CMA environments
reference](https://www.contentful.com/developers/docs/references/content-management-api/environments/)
requires callers to query the environment after create and defines `queued`,
`inProgress`, `ready`, and `failed` states. The first-party environment model exposes
the returned [`sys.status`
link](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/entities/environment.ts#L9-L25).

**Default fake behavior:** Registered and newly created environments are immediately
ready in the generic fake. Focused provider handlers return the exact status values
needed to exercise readiness behavior.

**Provider boundary:** Generic asynchronous timing remains intentionally unmodelled. The
provider stops immediately on the documented terminal `failed` status, while
unrecognized future values remain pollable.

**Coverage:** Unit tests cover all four documented statuses and an unrecognized future
value. Mocked acceptance covers `queued` to `ready`, terminal `failed` after one
request, and `inProgress` remaining pollable until its configured timeout.

### 404, version conflict, and rate-limit errors

**Evidence:** [Contentful error
reference](https://www.contentful.com/developers/docs/references/errors/) requires
`sys.type: Error`, a code in `sys.id`, and a nonempty `message`; the [CMA
overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/)
defines rate-limit headers and 429 behavior.

**Default fake behavior:** Generic missing resources return 404 `NotFound`;
version-locked resources return their endpoint-appropriate 409 classification; the
limiter returns 429 `RateLimitExceeded` with second-limit, remaining, and reset headers.

**Provider boundary:** The fake describes the response contract, not provider replay
policy. Entry publication and Content Type activation lifecycle mutations deliberately
return the first 429 without transparent replay; GET and unrelated operations retain
default retry behavior.

**Coverage:** Structural tests cover type, ID, nonempty message, status, and rate-limit
headers. Generated-client and Terraform lifecycle tests independently cover
endpoint-scoped 429 no-replay and marker outcomes.

### Preview environment configuration

**Evidence:** The [preview environment contract](../research/content-preview-environments.md)
records direct CMA and Web App observations; the HTTP endpoints are undocumented.

**Default fake behavior:** The fake merges configurations by content type identity,
retains omitted and disabled configurations, and advances the version only for metadata
changes.

**Provider boundary:** The fake accepts the provider's canonical request shape; the
observed legacy create-only `contentType` alias, selected-ID deletion residue, and
delayed read-after-delete visibility are not modeled. Version locking cannot detect
concurrent configuration-only changes.

**Coverage:** Client tests cover merge, disable, and version conflicts; provider tests
cover re-enable, response identity aliases, exact delta requests, projection, recovery
state, import, and lifecycle transitions. Live observations remain dated evidence, not a
guarantee.

### Live preview variables

**Evidence:** The [variables reference](../research/live-preview-variables.md) records sanitized
direct CMA observations.

**Default fake behavior:** The fake replaces the complete document, validates payloads
against an en-US fixture and the observed Text boundary, canonicalizes arrays, checks
versions while present, recreates at version 1, and deletes unconditionally.

**Provider boundary:** Alias routing, configurable locale inventories, update metadata,
and parts of the HTTP surface remain outside the fake. Unicode code-point counting and
parent cleanup are documented mock conventions.

**Coverage:** Raw HTTP tests check stored documents, response details, and failed-write
preservation. Literal client fixtures check decoding. Mocked Terraform tests check
lifecycle behavior and recovery from service validation errors.

## External evidence index

Sanitized API observations live independently of test-server coverage:

| Concern | API reference |
| --- | --- |
| Webhook topics, defaults, headers, and Basic password | [Webhooks](../research/webhooks.md#direct-configuration-observations) |
| Editor Interface sidebar values | [Editor Interfaces](../research/editor-interfaces.md) |
| Delivery API key versions | [Delivery API key versions](../research/delivery-api-keys.md#versioning-observations) |
| Entry publication and deletion | [Entry lifecycle](../research/entry-lifecycle.md) |
| Entry and team collections | [Collections and errors](../research/collections-and-errors.md) |
| Asynchronous environment status | [Environment readiness](../research/environment-readiness.md) |

These references preserve source revisions, observation or record dates, and
unverified behavior. Updating a fake does not refresh the external evidence.

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
- Entry unpublish models the live-probed whole-Entry transitions, including
  repeated unpublish. Locale-based unpublish remains unprobed and unmodelled.
- Eventual consistency beyond environment readiness is not simulated. A focused
  handler should model a concrete observed transition before it is added to the
  generic fake.

### Live preview variables fixture details

The locale inventory is fixed to `en-US`; configurable environment locale
inventories are not modeled. The 50,000-character Text limit counts Unicode
code points. That convention reproduces the observed ASCII boundary and accepted
BMP/astral examples; combining-sequence and grapheme semantics remain unverified.
Missing-parent PUT/DELETE and cleanup after environment deletion are mock lifecycle
conventions, not observations from the variables probes.

Alias routing, response update metadata, HEAD, trailing-slash GET, and service-style
404 responses for POST/PATCH and the space-level route are not modeled. Generated
handler errors use ordinary JSON, including conflicts; independent client fixtures
cover vendor-JSON decoding. Whole non-object request bodies receive generated
decoder errors; the Terraform client sends an object envelope. Authentication
distinctions remain the shared fake's behavior.
