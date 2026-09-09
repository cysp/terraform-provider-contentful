# Contentful collections and error representations

Collection envelopes, pagination limits, and error classifications belong to the
endpoint and API in use. Shared client types are useful representation references; they
do not prove that every endpoint supports the same queries, ordering, or error behavior.

## Scope and sources

This reference covers Content Management API (CMA) Entry and Content Type offset
collections, User Management API (UMA) organization teams, and common error shapes.
First-party source is pinned to [contentful-management.js
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
an empty item list. No entry data or identifiers were printed or retained. The returned
offset was the requested offset, not a value clamped to the collection length.

A successful out-of-range response from the team endpoint was not established by the
retained evidence. The documented requested-offset meaning remains independent of that
unverified status and item behavior. Tenant feature-access outcomes are excluded from
the record.

Cursor collections and specialized discovery projections have separate representations;
see [App Framework collection
queries](app-framework/README.md#transport-envelopes-and-collection-queries).

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

Validation details can vary in type and echo submitted values. The [App Action
observations](app-framework/actions.md) include both strings and arrays at
`details.errors`. Preserve the structural distinction while removing submitted tenant
data from research extracts.

## Limits

No generic collection ordering, complete filter grammar, cursor support,
malformed-envelope acceptance policy, or cross-endpoint error union is established here.
Published throttling behavior does not independently establish whether a particular
failed mutation committed or can safely be replayed.
