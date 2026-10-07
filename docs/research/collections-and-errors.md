# Contentful collections and error representations

Collection envelopes, pagination limits, and error classifications belong to the
endpoint and API in use. Shared client types are useful representation references; they
do not prove that every endpoint supports the same queries, ordering, or error behavior.

## Scope and sources

This reference covers Content Management API (CMA) and User Management API (UMA)
collection envelopes and common error shapes. The Entry, Content Type, and organization
team source comparison is pinned to [contentful-management.js
v12.15.0](https://github.com/contentful/contentful-management.js/tree/cc096a337f0e1db6114e8da645d69bb6eb90f11c).

## Offset collections

| Surface | Documented representation and limits |
| --- | --- |
| CMA Entries and Content Types | `sys`, `total`, `skip`, `limit`, and `items` in the collection envelope. Query and ordering support remain endpoint-specific. |
| UMA organization teams | The same named envelope members; `skip` is the requested offset and `limit` is capped at 100. |

Sources: [CMA
overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/),
[UMA
pagination](https://www.contentful.com/developers/docs/references/user-management-api/overview/#pagination),
and [organization
teams](https://www.contentful.com/developers/docs/references/user-management-api/teams/get-all-teams-for-an-organization/).
The pinned SDK's
[`CollectionProp`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/common-types.ts#L566-L574)
declares all five envelope members; its [team
adapter](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/team.ts#L25-L31)
returns that type. A TypeScript return annotation does not establish runtime rejection
of omitted metadata or a guaranteed service ordering.

### Direct offset observation

Experiment date unrecorded; results retained by 2026-08-24.

A read-only `GET
/spaces/{space_id}/environments/{environment_id}/entries?skip=999999&limit=2` probe
returned 200 with `sys.type: Array`, echoed `skip: 999999` and `limit: 2`, and returned
an empty item list. The returned offset was the requested offset, not a value clamped
to the collection length.

A successful out-of-range response from the team endpoint was not established by the
retained evidence. The documented requested-offset meaning remains independent of that
unverified status and item behavior.

## Space Role cursor migration

Checked against primary sources and read-only live requests on 2026-10-07 (UTC).

The [Space Role announcement](https://www.contentful.com/developers/api-changes/space-roles-collection-endpoints-update/)
says the space Role collection will replace `total` and `skip` with `pages`
navigation, accept `pageNext` and `pagePrev`, and retain `limit`. Its body says
February 15, 2027; its header says February 14. The scope is
`GET /spaces/{space_id}/roles`, not singular Role reads or mutations.

The [CMA cursor overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/#cursor-pagination)
shows root-relative URL strings in `pages.next` and `pages.prev`, carrying
`pageNext` and `pagePrev` opaque tokens. Forward traversal ends when `next`
is absent. It permits changing `limit` between pages. Its general opt-in
`cursor=true` example concerns Entries; the Role announcement describes an
endpoint migration and does not mention that parameter.
Neither source specifies an overlap period or the mechanics of the Role rollout.

The [Role endpoint reference](https://www.contentful.com/developers/docs/references/content-management-api/roles/get-all-roles/)
still shows an offset envelope. The SDK
[Role adapter](https://github.com/contentful/contentful-management.js/blob/1b080d46936332bce96499fc4d68532588981013/lib/adapters/REST/endpoints/role.ts#L24-L31)
passes query parameters through but still returns `CollectionProp<RoleProps>`.
Its [query type](https://github.com/contentful/contentful-management.js/blob/1b080d46936332bce96499fc4d68532588981013/lib/common-types.ts#L420-L433)
permits arbitrary keys; being able to send `cursor=true` does not establish
that the endpoint honors it. The shared
[cursor collection type](https://github.com/contentful/contentful-management.js/blob/1b080d46936332bce96499fc4d68532588981013/lib/common-types.ts#L585-L593)
omits `total`/`skip` and defines an optional `pages` object containing optional
`next` and `prev` strings.
That supports representing absent terminal navigation, but is not proof of
the eventual Role wire shape.

### Direct Role opt-in observations

On 2026-10-07, 21:43:30–21:44:08 UTC, 13 authenticated `GET` requests to
`https://api.contentful.com/spaces/{space_id}/roles` used credentials from
`terraform-provider-contentful-example` against one existing space containing
three Roles. No resources were created or modified. Every response returned HTTP
200 and the `sys`, `total`, `skip`, `limit`, and `items` envelope; none contained
`pages`.

| Query | Observed result |
| --- | --- |
| `limit=1`, with `cursor` omitted, `true`, or `false` | Same first Role; `total: 3`, `skip: 0`, `limit: 1`. |
| `cursor=true&limit=2` | First two Roles; `total: 3`, `skip: 0`, `limit: 2`. |
| `limit=100`, with `cursor` omitted or `true` | Same three Role IDs in the same order; `total: 3`, `skip: 0`, `limit: 100`. |
| `limit=1&skip=1`, with `cursor` omitted or `true` | Second Role; `total: 3`, `skip: 1`, `limit: 1`. |
| `limit=1&pageNext=<invalid>`, with `cursor` omitted or `true` | First Role; `total: 3`, `skip: 0`, `limit: 1`. |
| `cursor=true&limit=1&pagePrev=<invalid>` | First Role; `total: 3`, `skip: 0`, `limit: 1`. |

The `<invalid>` values above were synthetic non-cursor strings. Invalid string
controls for `skip` and `limit` also returned 200: invalid `skip` yielded
`skip: 0` and the first Role; invalid `limit` with `cursor=true` yielded
`limit: 0` and no items. Accepting invalid cursor parameters therefore does not
establish a cursor parser or its validation behavior.

In this space, `cursor=true` did not select cursor pagination and `skip` remained
effective. A provider that treated these responses as terminal cursor pages
would stop after the first page when the requested limit was below the Role
count. Retain offset traversal until live Role cursor behavior is established.
These observations do not establish availability for other spaces or regions,
feature enablement, or future rollout behavior. No valid cursor was returned,
so forward/backward cursor traversal and terminal cursor envelopes were not
exercised.

The provider's [Role traversal policy](../design/configuration-data-sources.md#role-traversal)
uses these published cursor conventions alongside the existing offset behavior.
Raw local fixtures establish provider behavior, not deployment of the announced change or the service's cursor encoding.

## Endpoint-specific collection observations

Observed: 2026-09-09 (UTC), passive configuration reads. These samples demonstrate
differences in returned metadata; they do not establish a shared pagination protocol.

| Collection | Observed envelope |
| --- | --- |
| Locales, Roles, TeamSpaceMemberships, WebhookDefinitions, EnvironmentAliases, and Content Types | `sys`, `total`, `skip`, `limit`, and `items` |
| [Editor Interfaces](editor-interfaces.md#collection-and-detail-representations) | `sys`, `total`, and `items`; no `skip` or `limit` |
| [Taxonomy concepts and concept schemes](taxonomy.md#read-representations-and-pagination) | `sys`, `limit`, `items`, and `pages`; no `total` or `skip` |
| WorkflowDefinitions: `GET /spaces/{space_id}/environments/{environment_id}/workflow_definitions` | `sys`, `total`, `skip`, `limit`, and empty `items` |
| AutomationDefinitions: `GET /spaces/{space_id}/environments/{environment_id}/automation_definitions` | `sys`, `limit`, and empty `items`; `pages` absent |
| AI Providers: `GET /organizations/{organization_id}/ai/providers` | `sys`, `limit`, and empty `items`; `pages` absent |
| Logging configurations: `GET /organizations/{organization_id}/logging_configurations` | `sys`, `limit`, empty `items`, and `pages: {}` |
| Environment templates: `GET /organizations/{organization_id}/environment_templates/` | `sys`, `limit`, empty `items`, and `pages: {}` |

The rows with empty `items` establish empty-response forms only. They do not characterize
object contents, nonempty pagination, credentials, activation, or other lifecycle behavior.
An absent `pages` member and a present empty object are distinct observed forms;
neither supplies a continuation URL.

The [App Framework collection account](app-framework/README.md#transport-envelopes-and-collection-queries)
links separately scoped installation/action offset probes, Function discovery
projections, FunctionLog summaries, and ResourceType cursor traversal. Related entities
can share an SDK return annotation while returning different envelopes or reduced
items. No generic collection decoder contract follows from the annotation alone.

## Error representations

The [Contentful error
reference](https://www.contentful.com/developers/docs/references/errors/) describes
`sys.type: "Error"`, an error code in `sys.id`, and `message`. The [CMA
overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/)
describes HTTP 429 and rate-limit headers, including `X-Contentful-RateLimit-Reset` in
seconds.

An HTTP status alone does not identify a resource's lifecycle outcome. The retained
[Taxonomy observations](taxonomy.md) distinguish 422 version validation from 409
`VersionMismatch`; [Delivery API key
observations](delivery-api-keys.md#versioning-observations) returned 409
`Conflict` for a stale version. Some [Live preview variables
errors](live-preview-variables.md#errors-and-alias-reads) use a separate `{statusCode,
error, message}` envelope. A generic service 404 is not sufficient evidence that a
document at a supported route is absent.

Likewise, an empty collection establishes only that the addressed request returned no
items. A denied response does not identify whether access policy, feature availability,
or another service condition caused it, and does not establish global object absence.

Validation details can vary in type and echo submitted values. The [App Action
observations](app-framework/actions.md) include both strings and arrays at
`details.errors`.

## Limits

No generic collection ordering, complete filter grammar, cursor support,
malformed-envelope acceptance policy, or cross-endpoint error union is established here.
Published throttling behavior does not independently establish whether a particular
failed mutation committed or can safely be replayed.
