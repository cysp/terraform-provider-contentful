# CMA test-server conformance boundaries

Use this reference when changing the in-process CMA server or a test that relies
on its behavior. The [test server](../../internal/contentful-management-go/testing/)
models the observable contracts that the provider depends on. It is not intended
to implement Contentful internals. Here, "fake" means that in-process server and
its stored fixtures. Default fake behavior should be backed by the cited CMA
reference, a pinned first-party
client, or a sanitized direct observation. Retain the evidence date and
limitations when reusing an observation. Deliberate fault injection must be named
as adversarial behavior rather than presented as CMA conformance.

The first-party source links below are pinned to reviewed
`contentful-management.js` v12.15.0 commit
`cc096a337f0e1db6114e8da645d69bb6eb90f11c` so that later upstream changes do
not silently alter this evidence.

## Response ownership

Every Handler endpoint returns a detached snapshot of stored fixture data,
including nested maps, slices, raw JSON, and collection items. The snapshot is
copied while holding the handler mutex, before the generated HTTP server encodes
it. Later mutations cannot change an earlier response, and direct callers cannot
change stored data by editing returned values. StoredAppInstallation and
StoredWebhookDefinition provide the same ownership guarantee for unredacted data.

The private snapshot operation copies generated DTO values without invoking JSON
codecs: optional flags, nil versus empty collections, timestamps, and intentionally
invalid fault fixtures retain their Go representation. It supports the acyclic
exported-field DTO graphs used by this server; it is not an arbitrary Go-object
cloner. Response redaction remains the endpoint's responsibility before copying.

Use SetContentTypeResponse, SetRoleResponse, SetTaxonomyConceptResponse, or
SetTaxonomyConceptSchemeResponse when a test needs an exact stored response, including an adversarial version or value.
These methods own a copy and bypass ordinary request validation and lifecycle
side effects. Ordinary SetX methods and direct request inputs may retain caller
storage; callers must not mutate that storage while the server retains it.

## Endpoint contracts

A passing mock test establishes how the provider behaves against that fixture. For a claim about Contentful, use the published contract or the scoped service observations linked under the endpoint.

### Environment aliases

Each child request resolves its environment address once. An app request uses
the target selected during authorization, even if the alias moves before the
handler runs. Management requests and direct handler calls resolve the address
while holding the handler lock.
Parent checks, version checks, reads, writes and lifecycle side effects use the
selected concrete environment in that Space. Each returned child's own
`sys.environment.sys.id` carries the address from the request, including in
collection items and mutation responses. Stored fixtures keep their concrete
environment identity. An Environment read through an alias has its own
`sys.aliasedEnvironment` link, and an EnvironmentAlias document identifies its
target in its top-level `environment` link.

This applies to Entries, Content Types, Editor Interfaces, Locales, Tags, UI
Extensions, App Installations and Live Preview Variables. The
[alias research](../research/environment-aliases.md) records which response
shapes have been observed. Applying the same child projection to the remaining
operations is a fixture inference. A later request observes a retarget
immediately in the fake. Service propagation time is unverified.

The fake rejects deletion of a [targeted environment][alias-targets] and of the
[master environment or alias][master-deletion] before changing stored data. It
uses HTTP 422 with `ValidationFailed` for these refusals. The documentation
specifies the restriction and leaves the error representation open. Deleting an
eligible environment discards its child fixtures, Content Type publication history
and issued tokens scoped to that environment. Recreating the same ID starts empty.
Clearing those tokens on deletion is a fixture lifecycle convention.

Handlers that use `master` to recognize a Space accept either a concrete
`master` or a `master` alias whose target exists in the same Space. New servers
start without either. `RegisterMasterEnvironmentAlias(spaceID)` seeds a ready
`master-1970-01-01` Environment and the `master` alias. Use it in a Space without
a concrete `master`, before seeding children to be read through `master`. Seed
those children under the returned concrete ID. `SetSpace` seeds the Space
document for discovery. `SetSpaceEnablements` seeds only the enablements.

Creating a custom alias requires a `master` alias with a concrete target in
the same Space. With a concrete `master`, a PUT at a free alias ID returns
HTTP 403 with `sys.id = EnvironmentAliasNotOptedIn` before target lookup, as
[observed in the example Space](../research/environment-aliases.md#custom-alias-creation-before-opt-in).
The fake models opt-in independently of subscription limits.

The fake reserves Environment and Environment Alias IDs in the same namespace
within a Space. A PUT using an ID occupied by the other kind returns HTTP 409
with `sys.id = Conflict` and preserves both collections. This models an ID
collision. The exact live CMA response to either PUT is unverified.

[alias-targets]: https://www.contentful.com/developers/docs/references/content-management-api/environment-aliases/#creating-environment-aliases
[master-deletion]: https://www.contentful.com/developers/docs/concepts/multiple-environments/#environments-technical-limitations

### App Definition parent scope

The [App Framework endpoint map](../research/app-framework/README.md#api-inventory)
and [native external reference representations](../research/app-framework/native-external-references.md)
identify organization and App Definition as the parent scope for app-owned
configuration. The fake uses one organization-aware App Definition lookup for
the definition and its children. Resource Type storage includes the complete
organization, App Definition, and Resource Type identity, so addressing a type
through another app cannot read, replace, or delete the original record.

Wrong-organization requests and children whose App Definition is absent return
404 without changing stored data. The exact missing-parent error message and
the fake's rejection precedence remain fixture conventions, not live observations.
Use distinct Resource Provider IDs and valid mappings when testing cross-parent access so that the response isolates the parent-scope check.

### App access tokens

The fake verifies RS256 app JWTs against the app definition's registered public
keys. A `kid` selects a single key. Without it, the fake tries each registered
key. It follows [Contentful's JWT requirements](https://www.contentful.com/developers/docs/references/content-management-api/app-access-token/),
including `typ: JWT`, a required `iat`, and at most ten minutes between `iat` and
`exp`. For an installed app, the exchange returns HTTP 201 and an opaque bearer
token with its own ten-minute expiry.

Within the installation's space and environment, the mock grants the implemented
Content Type, Editor Interface, Entry, Tag and Locale operations. It also permits
reading the app's own installation, including its Secret parameters. Updating or
deleting the installation returns HTTP 403. Reading an App Definition returns
404, and reading a Space or Environment returns 401. Other operations are denied
by default. App requests follow the shared [environment alias routing](#environment-aliases).

A standalone uninstall in the fake removes the App Installation but retains its
issued tokens until expiry. Contentful documents [loss of access on uninstall](https://www.contentful.com/help/apps/apps-across-spaces/#how-to-uninstall-an-app).
The behavior of a previously issued bearer immediately after uninstall has not
been verified against the live CMA.

See the [App Identity permission observations](../research/app-framework/identity.md#app-access-token-permissions)
for the operations exercised against Contentful. Locale writes follow the
published App Identity permissions. The test account rejected Locale creation
with both credentials.

### Locales

[Locale research](../research/locales.md) separates published constraints from
sampled behavior. The fake models Locale entities, not their effects on content.

Create generates an ID at version 1 and rejects blank names. Updates preserve
identity/default status. Identical complete PUTs retain the version, including
with stale/future headers; changed payloads require the exact version. Code
syntax is checked first. Duplicate POST returns 422, duplicate PUT and default
DELETE return the observed 500. Rejected writes preserve values and versions.

Fallback checks reject missing/cyclic targets, non-null default fallbacks, and
referenced-target renames/deletion. Delivery-disabled nondefault targets are
rejected; the observed default exception is retained. Exact errors for unprobed
nondefault restrictions are fixture conventions. Reads require an existing
parent environment.
Positive pagination and zero-limit defaulting are modeled; ID ordering remains
a fixture convention.

The fake does not model content relabeling/deletion, flag-dependent content
projection, publication validation, capacity, cloning, delivery APIs, propagation,
or negative pagination quirks. Tests needing these effects must supply explicit
HTTP responses. [HTTP tests](../../internal/contentful-management-go/testing/handler_locale_test.go)
cover the modeled transitions; provider [lifecycle](../../internal/provider/resource_locale_lifecycle_test.go)
and [recovery](../../internal/provider/resource_locale_recovery_test.go) tests
exercise Terraform integration.

### UI Extensions

**Evidence:** The [Extension endpoint reference](https://www.contentful.com/developers/docs/references/content-management-api/ui-extensions/create-or-update-an-extension/)
requires the last version for updates. The [CMA version-locking contract](https://www.contentful.com/developers/docs/references/content-management-api/overview/#updating-and-version-locking)
requires rejecting writes after an intervening version change.

**Default fake behavior:** Creation starts at version 1. Existing extensions
require the exact stored version before PUT changes any fields. Accepted PUTs
advance the version; rejected writes retain both fields and version.

**Provider boundary:** Version 1 initialization follows the published response
example. Advancing by one per accepted PUT and using HTTP 409 `VersionMismatch`
are fixture conventions consistent with other versioned endpoints, not separate
live Extension observations. Identical-payload update behavior is unverified.
The generated server still requires a version header on Create; the provider
sends zero. This does not model the documented optional Create header.

### Space Enablements

**Evidence:** The [CMA
reference](https://www.contentful.com/developers/docs/references/content-management-api/space-enablements/)
and [request-values probe](../research/space-enablements.md) establish the paired-member rule
and failed-request nonmutation.

**Default fake behavior:** The fake requires both paired members with equal values
before storing the document.

**Provider boundary:** Exact validation details remain unmodelled, and the provider
keeps all enablements independently Optional+Computed.

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

### Webhook signing secret

**Evidence:** The [API reference, pinned SDK, and direct observations](../research/webhook-signing-secret.md)
record the space singleton, first-create versus replacement status, redacted
GET/PUT responses, tested conditional headers allowing mutations, and error
status/shape observations.

**Default fake behavior:** GET returns 200; PUT returns 201 for first creation
and 200 for replacement. Both include a Space link and the final four characters
of the secret, without `sys.version`. Conditional headers do not prevent
PUT/DELETE. DELETE returns 204.
The fake stores only the redacted representation. It uses the shared
[Space parent check](#environment-aliases). A missing parent returns 404 `NotFound`. GET and
DELETE also return 404 `NotFound` when the secret is absent.
Malformed JSON returns 400 `BadRequest`; empty bodies and invalid values return
422 `ValidationFailed`. The existing decoder-error handler maps these failures
before the endpoint handler runs, as it does for Webhook topics and Live Preview
variables. The observed error statuses and absent-secret details are modeled.
The fake omits echoes of submitted values and details identifying the parent;
the fixture used for space presence and unprobed error precedence remain mock
conventions.
Shared authentication handling retains its generic invalid-token message.

**Provider boundary:** The [lifecycle contract](webhook-signing-secret.md) defines
response validation, value retention, absence handling, and mutation recovery.
Redacted metadata does not establish equality of complete secrets.

The fixture stores only a redacted suffix. To check which secret the provider submitted, inspect the literal request body as the [lifecycle tests](../../internal/provider/resource_webhook_signing_secret_test.go) do.

### Webhook definition versions

POST creates a webhook at version 0, as does PUT to an unused ID. PUT to an
existing definition requires the exact `X-Contentful-Version`. The mock compares
the stored fields before redacting secrets and increments the version only
when those fields change. See the [live version observations](../research/webhooks.md#webhook-definition-versions).

The fixture assigns a UTC creation time on POST and creation PUT and preserves it on updates. The fixture clock supplies its own fractional-second precision.

On an existing webhook, a PUT header with an explicit Boolean `secret` and no
`value` retains the stored flag and value at that exact key. If the key is
absent, the mock checks the version before returning 422 and preserves the
entire definition. With an explicit string value, the fixture preserves the
flag of an existing exact-name header. Newly supplied
bytes on a readable header remain readable. The fixture stores supplied bytes
on a secret header for request testing. Live PUT and GET responses expose only
its redacted descriptor. See the [flag and version observations](../research/webhooks.md#changing-the-secret-flag-of-an-existing-header).

POST requires a value for each header. PUT requires either a value or an explicit
secret flag. The fixture checks these request shapes first, then reports repeated
exact header keys in encounter order, followed by Basic credential errors.
These failures precede the existing definition’s version check. A PUT creating
an unused ID cannot retain a header value. The route-specific responses and
the tested error combinations are recorded in the
[header observations](../research/webhooks.md#header-names-order-and-secret-updates).

The generated POST client requires a string value on each header. A malformed
raw POST distinguishes omission from JSON null in its validation error. The
PUT client can send either form to retain a value by exact name when it
supplies a Boolean secret flag. The response decoder accepts
a string or an omitted value. Terraform delegates by omitting the member.
See the [raw null and version controls](../research/webhooks.md#header-validation-order-and-omitted-values-on-put).

### Webhook topics

For a collection POST missing a header value and `topics`, return the header issue first and the missing-topic issue second. See the [raw CMA observation](../research/webhooks.md#header-validation-order-and-omitted-values-on-put).

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

### Webhook Basic password

**Evidence:** The [raw CMA observations](../research/webhooks.md#basic-authentication-field-presence-and-version-changes)
cover field presence, validation errors, response redaction, and version controls
on POST and PUT.

**Default fake behavior:** An omitted Basic field retains its stored value on
update. Both fields default to null on creation. A nonempty string requires a
nonempty counterpart in the same request. Submitted usernames and passwords are limited
to 50 and 255 UTF-16 code units. Paired nulls clear both fields. When a
PUT sets one field to null and omits the other, the fixture rejects the request
if the field being cleared holds a nonempty string. Invalid pairs return a
field-specific 422 before the fixture applies any changes. On update, this
preserves the entire definition and its version. Successful writes and GET
include the username (possibly null or empty) and omit the password.

The fixture stores the submitted password to model subsequent writes and version
comparisons. The probes establish the redacted configuration and version. Error
precedence for a request that also violates a different constraint is a fixture
convention.

**Provider boundary:** The provider restores planned or prior state only when the
response property is absent; explicit JSON null remains distinct.

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

The provider uses `total` when present and otherwise continues until an empty page. The fixture orders by ID to make pages repeatable. Contentful’s default ordering remains unverified.

### Role updates

**Evidence:** The [Role update addressing contract](../research/space-access.md#role-update-addressing)
requires an existing Role; the update endpoint cannot create one with a specified ID.
The update example retains `description`, the pinned SDK forwards the current
version, and the shared CMA locking contract rejects stale updates.

**Default fake behavior:** Updating a missing or deleted Role returns 404
`NotFound` without adding it to the store. Create and Update preserve submitted
description values, including null, empty, and absent representations. Create
starts at version zero and each accepted Update advances the version by one;
reusing an earlier version returns 409 `VersionMismatch` without changing the
record. These exact versions, omission behavior, and error envelopes are fixture
conventions, not separately observed Role mutation behavior.

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

For an out-of-range offset the fixture returns HTTP 200 with an empty page. The [offset observations](../research/collections-and-errors.md#direct-offset-observation) leave that case open for Teams.

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

The observed `Conflict` classification applies to this endpoint. See the [versioning observations](../research/delivery-api-keys.md#versioning-observations) when changing the fixture.

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
an inference from their common representation. The [Entry acceptance test](../../internal/provider/resource_entry_test.go) exercises configured empty arrays when run live.

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

### 404, version conflict, and rate-limit errors

**Evidence:** [Contentful error
reference](https://www.contentful.com/developers/docs/references/errors/) requires
`sys.type: Error`, a code in `sys.id`, and a nonempty `message`; the [CMA
overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/)
defines rate-limit headers and 429 behavior.

**Default fake behavior:** Generic missing resources return 404 `NotFound`;
version-locked resources return their endpoint-appropriate 409 classification; the
limiter returns 429 `RateLimitExceeded` with second-limit, remaining, and reset headers.

**Provider boundary:** Endpoint-scoped rate-limit faults reject the request before
delegating to the fake. Provider tests verify retries of the same body and version
within the deadline, followed by acknowledgement or a terminal failure. These
fixtures do not establish live CMA commitment guarantees; the
[HTTP retry policy](contentful-http-retry-policy.md) records the evidence limits.

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

### Live preview variables

**Evidence:** The [variables reference](../research/live-preview-variables.md) records sanitized
direct CMA observations.

**Default fake behavior:** The fake replaces the complete document, validates payloads
against an en-US fixture and the observed Text boundary, canonicalizes arrays, checks
versions while present, recreates at version 1, and deletes unconditionally.

#### Fixture limits

The locale inventory is fixed to `en-US`; configurable environment locale
inventories are not modeled. The 50,000-character Text limit counts Unicode
code points. That convention reproduces the observed ASCII boundary and accepted
BMP/astral examples; combining-sequence and grapheme semantics remain unverified.
Missing-parent PUT/DELETE and cleanup after environment deletion are mock lifecycle
conventions, not observations from the variables probes.

Response update metadata, HEAD, trailing-slash GET, and service-style
404 responses for POST/PATCH and the space-level route are not modeled. Generated
handler errors use ordinary JSON, including conflicts; independent client fixtures
cover vendor-JSON decoding. Whole non-object request bodies receive generated
decoder errors; the Terraform client sends an object envelope. Authentication
distinctions remain the shared fake's behavior.

### App Event Subscriptions

**Evidence:** The [App Event research](../research/app-framework/events.md) records
HTTP replacement, required topics/HTTPS target, duplicate rejection, status codes,
and the parent-scoped singleton. The pinned first-party entity documents optional
filter, transformation, and handler Function links.

**Default fake behavior:** Storage is keyed by organization and App Definition;
parent existence and organization are checked. PUT replaces the complete document,
returning 201 initially and 200 afterward. GET/DELETE of absent subscriptions or
parents return 404. HTTP topics are nonempty and unique, with no closed vocabulary.
The fake accepts documented Function links without modeling deployment or entitlement.
Function omission/removal and HTTP/handler transitions use complete replacement as
an explicit fixture assumption; the [configuration observations](../research/app-framework/events.md#subscription-configuration-observations)
did not verify any successful Function configuration. Parent deletion
clears the fake's subscription record as a storage convention.

**Provider boundary:** The provider forwards the complete desired document and
checks returned values. A contradictory or lossy response cannot establish a
successful apply. The fixture does not establish Function execution, event delivery,
retry/order guarantees, or compatibility of Function invocation roles.

To verify recovery after a failed mutation, inspect the saved state before refresh can replace it. A synthetic future topic exercises provider extensibility. Contentful acceptance of such a topic remains unverified.

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
