# Entry lifecycle: publication, versions, and deletion

An Entry is created as a draft. A full-body update changes its draft; publishing is a
separate versioned operation. Deleting a published Entry requires unpublishing it first.
`sys.version` identifies the current Entry version; `sys.publishedVersion` records the
version published. Exact increments observed in experiments are distinct from documented
optimistic-locking requirements.

## Scope and evidence

This reference covers whole-Entry operations under
`/spaces/{space_id}/environments/{environment_id}/entries/{entry_id}`. Locale-based
publication is a separate feature and was not exercised here.

First-party sources are [contentful-management.js
v12.15.0](https://github.com/contentful/contentful-management.js/tree/cc096a337f0e1db6114e8da645d69bb6eb90f11c)
and the pinned .NET client cited below. Published unpublish/delete references reviewed:
2026-09-09. Observation dates are recorded with their corresponding studies.

## Operations and transitions

| Operation | Method and suffix | Effect |
| --- | --- | --- |
| Create with a specified ID | PUT to the Entry address | Creates a draft; content-type header identifies its Content Type. |
| Update | PUT to the Entry address | Replaces draft content at the supplied version. |
| Publish | PUT `.../published` | Publishes the specified current version. |
| Unpublish | DELETE `.../published` | Removes whole-Entry publication; returns an Entry. |
| Delete | DELETE to the Entry address | Removes an unpublished Entry; succeeds with no response body. |

The [PUT header reference](entry-and-content-type-put-headers.md) records the
create/update distinction, including absent-target behavior. The published sources below
and the observation tables establish the remaining transitions.

## Published contract and first-party clients

The CMA [version-locking
contract](https://www.contentful.com/developers/docs/references/content-management-api/overview/#updating-and-version-locking)
requires the current version in `X-Contentful-Version` and says Contentful rejects the
update if the version changed in between. The [PUT header
reference](entry-and-content-type-put-headers.md#documented-contract) owns the
create/update header distinction and client request behavior. [Entry
fields](entry-fields.md#published-evidence) covers full-body replacement, defaults,
and response omission.

The [Entry
reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/)
defines Publish and whole-Entry Unpublish as separate operations on `/published`.
Whole-Entry unpublish returns an Entry; its headers and direct precondition probes are
covered under [unpublish and delete preconditions](#unpublish-and-delete-preconditions).

The common-system-property table defines `sys.version` as the current version and
`sys.publishedVersion` as the published version, but specifies no arithmetic
relationship between them. See [Common resource
attributes](https://www.contentful.com/developers/docs/references/content-management-api/overview/#common-resource-attributes).

The first-party
[`publish`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L168-L185)
adapter sets `X-Contentful-Version` from `sys.version`. It and the
[`update`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L155)
and
[`unpublish`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L187-L209)
adapters return an Entry without validating its version tuple.

## Version relationships and evidence limits

| Relationship | Evidence |
| --- | --- |
| Update and publish submit the current `sys.version` | Documented locking requirement, also implemented by the first-party adapter. |
| Create returns version `1`; update returns prior version plus one | Observed in the retained probes. The reviewed narrative contract does not promise either number. |
| Publish returns `publishedVersion` equal to the submitted version | Observed in the retained probes. The reviewed generated example sends version `6` but shows `publishedVersion: 9`; that example cannot prove an equality guarantee. |
| Publish increments `version` by one | Explicit assumption in the first-party `isUpdated` helper; matches the probes but is not a general service arithmetic guarantee. |
| Unpublished changes mean `version > publishedVersion + 1` | First-party helper classification; it does not validate every mutation response before returning it. |
| Unpublish advances `version` and removes `publishedVersion` | Observed in both the unpublish response and subsequent GET. No exact increment is established as a permanent contract. |

The first-party [`isUpdated`
helper](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/plain/checks.ts#L3-L11)
is evidence of client interpretation. API response arithmetic, client state
classification, and a consumer's consistency checks are separate matters. A new version
observed by GET does not identify which client authored it or which version an earlier
ambiguous request published.

## Direct lifecycle observations

### Create and publish

Experiment date unrecorded; results retained by 2026-08-24.

A disposable Content Type and Entry were used to probe the successful Entry publication
relationship. Only status codes and the selected version fields below were recorded;
complete response bodies were not retained.

| Request | Relevant header | Status | Structural observation |
| --- | --- | ---: | --- |
| `PUT /spaces/{space_id}/environments/{environment_id}/content_types/{content_type_id}` | — | 201 | `sys.version: 1` |
| `PUT /spaces/{space_id}/environments/{environment_id}/content_types/{content_type_id}/published` | `X-Contentful-Version: 1` | 200 | The disposable Content Type was activated |
| `PUT /spaces/{space_id}/environments/{environment_id}/entries/{entry_id}` | `X-Contentful-Content-Type: {content_type_id}` | 201 | `sys.version: 1` |
| `PUT /spaces/{space_id}/environments/{environment_id}/entries/{entry_id}/published` | `X-Contentful-Version: 1` | 200 | `sys.version: 2`, `sys.publishedVersion: 1` |

### Publish, update, and unpublish

Experiment date unrecorded; results reviewed on 2026-08-26.

A disposable Content Type and Entry were created in an isolated test scope. The Entry
was published, then updated once to leave a newer draft while the older version remained
published. Complete response bodies were not retained.

| Operation | HTTP status | Structural observation |
| --- | --- | --- |
| Create Entry | 201 | `version` was positive |
| Publish Entry | 200 | `publishedVersion` equalled the submitted version; current `version` advanced by one |
| Update Entry | 200 | current `version` advanced by one; `publishedVersion` remained positive and older |
| Unpublish Entry | 200 | current `version` advanced by one from the pending draft; `publishedVersion` was absent |
| GET Entry after unpublish | 200 | returned the same advanced version as unpublish; `publishedVersion` remained absent |

In this sequence, whole-Entry unpublish advanced `version` beyond the pending draft's
version. The exact `+1` increment is an observation, not a documented guarantee for every
unpublish response.

## Unpublish and delete preconditions

Observed precondition probes: 2026-08-30.

The whole-Entry
[unpublish](https://www.contentful.com/developers/docs/references/content-management-api/entries/unpublish-an-entry/)
and
[delete](https://www.contentful.com/developers/docs/references/content-management-api/entries/delete-an-entry/)
references show an optional `X-Contentful-Version` header. The JavaScript client sends
no version or ETag precondition for
[unpublish](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L187-L209)
or
[delete](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L158-L166).
The .NET client instead requires and sends a version for
[unpublish](https://github.com/contentful/contentful.net/blob/090889393b58113f50af45214d92fe92045c50d5/Contentful.Core/ContentfulManagementClient.cs#L713-L732)
and
[delete](https://github.com/contentful/contentful.net/blob/090889393b58113f50af45214d92fe92045c50d5/Contentful.Core/ContentfulManagementClient.cs#L674-L688).

| Operation | Preconditions tested | Result |
| --- | --- | --- |
| Unpublish published Entry | stale, zero, or omitted `X-Contentful-Version`; stale `If-Match` | 200; all returned an unpublished Entry |
| Delete unpublished Entry | stale, zero, or omitted `X-Contentful-Version`; stale `If-Match` | 204; following GET returned 404 in every case |
| Unpublish unpublished Entry | current `X-Contentful-Version` | 400 `BadRequest`, `Not published`; no version change |
| Delete published Entry | none | 400 `BadRequest`, `Cannot delete published`; Entry remained |
| Unpublish or delete absent Entry | version header present | 404 `NotFound`, `The resource could not be found.` |

In the tested requests, neither header enforced a mutation precondition for these
whole-Entry endpoints. Successful unpublish returned the resulting Entry; its version
transition is recorded separately in [publish, update, and
unpublish](#publish-update-and-unpublish).

## Operational limits

The observed ignored version and ETag headers do not provide a concurrency barrier for
whole-Entry unpublish or delete. An intervening change can therefore be affected by a
later deletion request. These results must not be generalized to versioned update,
publish, Taxonomy deletion, or another endpoint.

A successful unpublish followed by a failed delete leaves an unpublished Entry; the two
HTTP requests are distinct operations. No cross-request transaction or unconditional
replay safety is established by the evidence. After an ambiguous response, a read can
show current state but cannot prove the provenance of a matching draft or reconstruct
every intervening operation.

See [Entry field values](entry-fields.md) and [Entry
metadata](entry-metadata.md) for response normalization separate from publication
versions.
