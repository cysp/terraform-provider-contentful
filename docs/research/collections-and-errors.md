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
