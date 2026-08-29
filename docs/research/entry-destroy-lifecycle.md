# Entry destroy lifecycle evidence

Evidence captured 2026-08-30 against Contentful's CMA reference, first-party
JavaScript management client v12.15.0 commit
[`cc096a3`](https://github.com/contentful/contentful-management.js/tree/cc096a337f0e1db6114e8da645d69bb6eb90f11c),
first-party .NET client commit
[`0908893`](https://github.com/contentful/contentful.net/tree/090889393b58113f50af45214d92fe92045c50d5),
and disposable live CMA probes.

## Published evidence

The whole-Entry [unpublish](https://www.contentful.com/developers/docs/references/content-management-api/entries/unpublish-an-entry/)
and [delete](https://www.contentful.com/developers/docs/references/content-management-api/entries/delete-an-entry/)
references show an optional `X-Contentful-Version` header. The JavaScript client
sends no version or ETag precondition for
[unpublish](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L1123-L1168)
or [delete](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L1073-L1089).
The .NET client instead requires and sends a version for
[unpublish](https://github.com/contentful/contentful.net/blob/090889393b58113f50af45214d92fe92045c50d5/Contentful.Core/ContentfulManagementClient.cs#L713-L732)
and [delete](https://github.com/contentful/contentful.net/blob/090889393b58113f50af45214d92fe92045c50d5/Contentful.Core/ContentfulManagementClient.cs#L674-L688).

## Live observations

| Operation | Preconditions tested | Result |
| --- | --- | --- |
| Unpublish published Entry | stale, zero, or omitted `X-Contentful-Version`; stale `If-Match` | 200; all returned an unpublished Entry |
| Delete unpublished Entry | stale, zero, or omitted `X-Contentful-Version`; stale `If-Match` | 204; following GET returned 404 in every case |
| Unpublish unpublished Entry | current `X-Contentful-Version` | 400 `BadRequest`, `Not published`; no version change |
| Delete published Entry | none | 400 `BadRequest`, `Cannot delete published`; Entry remained |
| Unpublish or delete absent Entry | version header present | 404 `NotFound`, `The resource could not be found.` |

The probes establish that neither header provides a mutation precondition for
these whole-Entry endpoints. Successful unpublish returned the resulting Entry;
its version transition is recorded separately in
[Entry unpublish version behavior](entry-unpublish-version.md).

## Limitation and scope

These live observations are time- and environment-specific, not a promise that
Contentful will ignore the headers forever or in every account. Whole-Entry
unpublish and delete behavior must not be generalized to other endpoints
without endpoint-specific evidence.
