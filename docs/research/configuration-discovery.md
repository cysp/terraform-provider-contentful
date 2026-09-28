# Configuration discovery endpoints

These Content Management API operations address existing Spaces, Environments,
Environment Aliases, Locales, Roles, and Content Types. Evidence here is limited to
published references and pinned SDK source; direct observations are linked separately.

| Family | Detail and collection addressing | Collection evidence |
| --- | --- | --- |
| Space | `/spaces/{space_id}` and `/spaces` | [Listing reference][spaces] explicitly documents `skip`/`limit` offset pagination, optional cursor pagination, and optional `X-Contentful-Organization`. Omitted organization scope returns Spaces accessible across organizations. |
| Environment | `/spaces/{space_id}/environments/{environment_id}` and parent collection | [Listing reference][environments] shows an offset envelope and a `limit` SDK example; the family adapter forwards `PaginationQueryParams`. |
| Environment Alias | `/spaces/{space_id}/environment_aliases/{environment_alias_id}` and parent collection | [Listing reference][aliases] shows an offset envelope and a `limit` SDK example; the family adapter forwards `PaginationQueryParams`. |
| Locale | `/spaces/{space_id}/environments/{environment_id}/locales/{locale_id}` and parent collection | [Listing reference][locales] shows an offset envelope and a `limit` SDK example; the family adapter forwards `QueryParams`. |
| Role | `/spaces/{space_id}/roles/{role_id}` and parent collection | [Detail][role] and [listing][roles] references show the routes and offset envelope; the [pinned adapter][role-adapter] forwards pagination query parameters. |
| Content Type | `/spaces/{space_id}/environments/{environment_id}/content_types/{content_type_id}` and parent collection | [Detail][content-type] and [listing][content-types] references show these GET routes; the listing example includes `skip`, `limit`, `total`, and `items`. The provider's generated `GetContentTypes` client accepts `skip` and `limit`. |

The [CMA overview][overview] describes offset traversal using `skip`, `limit`,
`total`, and `items`. Family adapters corroborate the routes and query forwarding;
their TypeScript signatures alone do not establish raw server validation or order
stability. Pagination does not establish a transactionally consistent inventory.

Contentful has [announced a cursor transition for Role collections][role-cursor].
The current Role reader uses offsets and cannot traverse a cursor-only response.

Space `query` matches an exact Space ID or a partial Space name. A returned fuzzy
match therefore does not establish unique identity. The [Space detail reference][space]
and pinned entity expose `name` and `sys.organization`; optional license expansion
is separate from ordinary configuration identity. Billing and audit metadata are
not needed to identify a Space.

The [alias reference][alias-family] supports Environment detail reads through
alias IDs. Such responses keep the alias at `sys.id` and supply the target at
`sys.aliasedEnvironment`; a direct target read need not include that property.
Published alias detail/list examples spell their type `Environment Alias`;
examples do not prove which spelling every live endpoint emits. See
[provider decoding policy](../design/configuration-data-sources.md#identity-and-projection)
for supported spellings and [alias routing and observations](environment-aliases.md) for
non-atomicity and child metadata limits.

The Environment [detail example][environment] omits `name` and `sys.status`, while
its pinned entity requires both and the alias family examples show both. This
source discrepancy is unresolved by the published material. The Locales entity
includes ID, code, default, fallback and editing/delivery/optional flags. See
[Locale raw response evidence](locales.md) for ID/code differences, null fallback,
and the wrapper's removal of `internal_code`. A single static list/detail
comparison does not establish universal equivalence across these families.

Content Type detail and ordinary collection reads address the current CMA
model. The [activated collection][activated-content-types] is a distinct
`/public/content_types` endpoint that ignores changes since the last
activation. Published examples do not establish how Content Type
`sys.environment` is represented on an alias-routed response. The [environment
alias guide][alias-guide] establishes routing for environment-aware resources,
not a Content Type-specific response-link guarantee.

## First-party source

The JavaScript evidence is pinned to
[`883e2b9dc1c76413d5c24e45f74243da699071e4`](https://github.com/contentful/contentful-management.js/tree/883e2b9dc1c76413d5c24e45f74243da699071e4):
[`space.ts` adapter](https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/space.ts),
[`environment.ts` adapter](https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/environment.ts),
[`environment-alias.ts` adapter](https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/environment-alias.ts),
[`locale.ts` adapter](https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/locale.ts),
and their corresponding [entity definitions](https://github.com/contentful/contentful-management.js/tree/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities).

Ruby [`space.rb`](https://github.com/contentful/contentful-management.rb/blob/46060ea8341350ebb15753ff9a66990cefbc8e71/lib/contentful/management/space.rb)
and [`locale.rb`](https://github.com/contentful/contentful-management.rb/blob/46060ea8341350ebb15753ff9a66990cefbc8e71/lib/contentful/management/locale.rb)
independently corroborate Space addressing/name and Locale code/fallback/flags.
Their projections are client behavior, not raw response schemas.

[spaces]: https://www.contentful.com/developers/docs/references/content-management-api/spaces/get-all-spaces-an-account-has-access-to/
[space]: https://www.contentful.com/developers/docs/references/content-management-api/spaces/get-a-space/
[environments]: https://www.contentful.com/developers/docs/references/content-management-api/environments/get-all-environments-of-a-space/
[environment]: https://www.contentful.com/developers/docs/references/content-management-api/environments/get-an-environment/
[aliases]: https://www.contentful.com/developers/docs/references/content-management-api/environment-aliases/get-all-environment-aliases-of-a-space/
[alias-family]: https://www.contentful.com/developers/docs/references/content-management-api/environment-aliases/
[locales]: https://www.contentful.com/developers/docs/references/content-management-api/locales/get-all-locales-of-a-space/
[roles]: https://www.contentful.com/developers/docs/references/content-management-api/roles/get-all-roles/
[role]: https://www.contentful.com/developers/docs/references/content-management-api/roles/get-a-role/
[role-adapter]: https://github.com/contentful/contentful-management.js/blob/310c01216a2eaa5041dbac8263e9cbba47eb2fb2/lib/adapters/REST/endpoints/role.ts
[role-cursor]: https://www.contentful.com/developers/api-changes/space-roles-collection-endpoints-update/
[overview]: https://www.contentful.com/developers/docs/references/content-management-api/overview/#collection-resources-and-pagination
[content-type]: https://www.contentful.com/developers/docs/references/content-management-api/content-types/get-a-content-type/
[content-types]: https://www.contentful.com/developers/docs/references/content-management-api/content-types/get-all-content-types-of-a-space/
[activated-content-types]: https://www.contentful.com/developers/docs/references/content-management-api/content-types/get-all-activated-content-types-of-a-space/
[alias-guide]: https://www.contentful.com/developers/docs/concepts/environment-aliases/
