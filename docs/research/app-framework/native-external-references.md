# Native external references: providers, types, and resources

ResourceProvider and ResourceType describe how an app resolves external content.
Resource responses expose external records through CMA discovery and lookup; they do not
transfer ownership of those records to Contentful.

See the [shared study scope and sources](README.md#scope-and-evidence) and [API
inventory](README.md#api-inventory). No external-resource resolution was exercised;
source types and examples do not establish complete wire behavior.

## ResourceProvider, ResourceType, and Resource

ResourceProvider, ResourceType, and Resource connect an app's Function to external
content. Configuration, environment discovery, external resolution, and delivery each
expose different data.

| Resource | Request and response data | Identity and behavior |
| --- | --- | --- |
| ResourceProvider | PUT body requires `sys: {id}`, `type: "function"`, and `function: Link<Function>`. Response adds metadata and organization/definition links, without SDK version. | A parent-addressed singleton **with its own provider ID**. Its writable `sys.id` is an explicit exception to ordinary body projections. [Entity][provider-entity], [adapter][provider-sdk]. |
| ResourceType | PUT body contains `name` and `defaultFieldMapping`, without `sys`; response adds ID and provider/definition/organization links. | Addressed by resourceTypeId; documented IDs have `{Provider}:{Type}` form. The full configuration projection has no SDK version. [Entity][type-entity], [adapter][type-sdk], [concept][resource-entities]. |
| Environment ResourceType | Collection items expose `name` and `sys`, not `defaultFieldMapping` in the SDK. Some metadata is optional for system types such as `Contentful:Entry`. | This reduced discovery projection cannot reconstruct the full app-owned mapping. [Entity][type-entity], [endpoint][environment-types]. |
| Resource | GET selects search through `query` or lookup through string-valued `sys.urn[in]`; optional locale, referencingEntryId, limit, pageNext/pagePrev. Returns `sys.type: "Resource"`, `sys.urn`, related type/provider/definition links, and mapped `fields`. | URN identity, not a CMA-managed object ID/version; cursor pagination rather than an ordinary offset collection. These are resolved external records, not create/update requests to the external system. [Entity][resource-entity], [adapter][resource-sdk]. |

In the pinned SDK, `defaultFieldMapping.title` is required; `subtitle`, `description`,
`externalUrl`, `image`, and `badge` are optional. The `image` object has `url` and
optional `altText`; `badge` has `label` and `variant`. [Mapping type][type-entity].

The Resource Entities guide describes JSON-pointer substitutions in display mapping
strings. For example, `"title": "{ /title }"` selects the external object's title, while
`"subtitle": "Reference: { /id }"` combines text with its ID. The guide and pinned
first-party example describe display projection, not a universal schema of the external
object. [Resource Entities][resource-entities],
[mapping example][mapping-example].

Response badge variants are `primary`, `negative`, `positive`, `warning`, and `secondary`
in the SDK, while the mapping's `variant` is declared as a string. The type alone does
not establish its server validation. [Mapping type][type-entity], [Resource
type][resource-entity].

The pinned ResourceType environment-list adapter accepts cursor options in its signature
but does **not** forward them to the HTTP request. The public reference illustrates a
cursor response without listing query parameters. This is an SDK forwarding gap, not
proof that the service ignores cursors; exhaustive environment discovery through that
helper remains unverified. [Adapter][type-sdk], [endpoint][environment-types].

The [supplied read-only study](README.md#source-metadata) exercised direct HTTP forward paging
at `GET E/resource_types?limit=1`. Responses contained `{sys, items, limit, pages}`,
without `total` or `skip`, and provided relative `pages.next` links. Following those
links reached terminal `pages: {}`; concatenated items exactly matched the baseline
collection, including order. This establishes forward paging for that existing
collection independently of the SDK forwarding gap. Reverse paging, cursor internals,
and stability during concurrent changes remain unverified.

ResourceType discovery is distinct from `GET E/resource_types/{resource_type_id}/resources`,
which resolves external data through a Function and may contact a third-party service.
That resolution route was not exercised by the read-only experiment. [Resource
Entities][resource-entities].

| Function invocation | Pinned toolkit event and response shape |
| --- | --- |
| `resources.search` | Event contains resourceType and limit, optional query/locale/referencingEntryId/pages.nextCursor. Returns external-object `items` and `pages: {nextCursor?}`. |
| `resources.lookup` | Event contains resourceType, lookupBy (a map of scalar arrays), limit, and optional locale/referencingEntryId/cursor context. Returns the same items/pages shape. |
| `graphql.resourcetype.mapping` | Event supplies `resourceTypes: [{resourceTypeId}]`; response wraps mappings containing resourceTypeId, graphQLQueryField, graphQLQueryArguments, and optional graphQLOutputType. |
| `graphql.query` | Event has query, isIntrospectionQuery, variables, optional operationName. Response permits data (including null), errors, and extensions. |

The `resources.search`/`resources.lookup` responses carry external data before the CMA
display projection; GraphQL responses instead define mappings or resolve delivery
queries. The external cursor `pages.nextCursor` and CMA collection links `pages.next`
are different protocol layers. Sources: [resource invocation
types][resource-function-types], [GraphQL types][function-types]. The concept guide
disagrees with itself about whether lookup limit is optional; the toolkit requires it.
No live evidence resolves missing URNs, partial results/errors, ordering, or external
paging. [Resource Entities][resource-entities].

Custom external references additionally use `graphql.field.mapping`. Delivery lookup can
batch URNs; CDA/CPA requests omit the Entry editor request's `referencingEntryId`.
Native external references are documented for Premium and above. They are not
dependencies of an ordinary HTTP AppEventSubscription. [Resource
Entities][resource-entities], [native references availability][native-references].

## Unresolved behavior

ResourceProvider ID mutability, provider/type deletion effects on existing links,
missing/duplicate URNs, partial errors, ordering, Function reference validation, and
reverse/concurrent pagination behavior and exhaustive discovery through the SDK helper.

[provider-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/resource-provider.ts
[type-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/resource-type.ts
[resource-sdk]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/resource.ts
[function-types]: https://github.com/contentful/node-apps-toolkit/blob/64fa31b6b2223cd1c8b1798fa540e8aad5e2d319/src/requests/typings/function.ts
[resource-entities]: https://www.contentful.com/developers/docs/extensibility/app-framework/resource-entities/
[native-references]: https://www.contentful.com/help/connect-content/native-external-references/
[provider-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/resource-provider.ts
[type-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/resource-type.ts
[environment-types]: https://www.contentful.com/developers/docs/references/content-management-api/native-external-references/get-all-resource-types-in-an-environment/
[resource-entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/resource.ts
[resource-function-types]: https://github.com/contentful/node-apps-toolkit/blob/64fa31b6b2223cd1c8b1798fa540e8aad5e2d319/src/requests/typings/resources.ts
[mapping-example]: https://github.com/contentful/apps/blob/8d3ecaffa406b36e0a494917a1fbb98a59c16603/examples/native-external-references-mockshop/src/tools/entities/product.json
