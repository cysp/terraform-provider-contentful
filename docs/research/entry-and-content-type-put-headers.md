# Entry and Content Type PUT header semantics

Creating an Entry with a specified ID and updating an Entry share a PUT path but
require different headers. In the tested requests to absent Entries, omitting
`X-Contentful-Content-Type` prevented creation despite a supplied version header.
Omitting that header on update therefore avoids recreating an Entry that another
client deleted. Deletion between a read and update was not tested separately;
that consequence follows from the observed requests to absent Entries. A version
header alone did not prevent creation of an absent Content Type.

This note separates Contentful's documented contract, pinned JavaScript client behavior,
and sanitized direct observations of the Content Management API (CMA).

## Documented contract

Contentful uses the same Entry `PUT` endpoint to create an Entry with a specified ID and
to update an existing Entry. The official [Entry
reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/)
says that Create requires `X-Contentful-Content-Type`, while Update requires the last
`X-Contentful-Version` and does not require the Content Type header. The [specified-ID
endpoint
reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/create-an-entry-with-a-specified-id/)
describes the same distinction.

The specified-ID endpoint's generated example contains both headers, so its narrative
and the first-party source are better evidence of the create/update distinction than
that one example.

Contentful's JavaScript management SDK implements that contract at commit
`cc096a337f0e1db6114e8da645d69bb6eb90f11c`. Entry
[`createWithId`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L255-L281)
sends `X-Contentful-Content-Type` and no version, while
[`update`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L155)
sends the observed `sys.version` and no Content Type header. Content Type
[`createWithId`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/content-type.ts#L71-L80)
and
[`update`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/content-type.ts#L82-L96)
likewise distinguish Create from Update by omitting or supplying the version.

## Entry direct observation

Observed: 2026-09-02.

A sanitized probe against the CMA used an existing activated Content Type, represented
below as `fixture`. The probe did not create, activate, or modify that Content Type.

Every request used the member path `PUT
/spaces/{space_id}/environments/{environment_id}/entries/{entry_id}`, the header
`Content-Type: application/vnd.contentful.management.v1+json`, and the body
`{"fields":{}}`. Every target had a randomized disposable ID and returned `404 NotFound`
to a preflight `GET`. Successful requests returned `sys.contentType.sys.id`
corresponding to `fixture`, an empty `fields` object, and no `publishedVersion`.

| Target before `PUT` | `X-Contentful-Content-Type` | `X-Contentful-Version` | `PUT` result | Follow-up `GET` |
| --- | --- | --- | --- | --- |
| Absent | `fixture` | absent | `201`, version `1` | `200`, version `1` |
| Absent | absent | `1` | `400 BadRequest`: “You should provide a content type in X-Contentful-Content-Type request header.” | `404`, still absent |
| Absent | `fixture` | `1` | `201`, version `1` | `200`, version `1` |
| Absent | absent | absent | same `400 BadRequest` | `404`, still absent |
| Absent | absent | `7` | same `400 BadRequest` | `404`, still absent |
| Absent | `fixture` | `7` | `201`, version `1` | `200`, version `1` |
| Existing disposable draft at returned version `1` | `fixture` | exact returned version `1` | `200`, version `2` | `200`, version `2` |
| Existing disposable draft at returned version `1` | absent | exact returned version `1` | `200`, version `2` | `200`, version `2` |

A separate direct collision observation sent an existing version-`1` Entry with Content
Type and no Version. CMA returned `409 VersionMismatch`, and a subsequent `GET` showed
that the Entry was unchanged.

Every failed absent-target `PUT` shown in the table returned error `sys.id`
`BadRequest`, the message quoted in the table, and null `details`. Its follow-up `GET`
returned `404 NotFound`, so the failed request did not create the target.

For the observed absent targets, Content Type presence selected creation: versions `1`
and `7` did not prevent creation when Content Type was present, and neither version
authorized creation without Content Type. For the observed existing drafts, the exact
returned version selected Update and the Content Type header made no difference. The
observations therefore distinguish the Create and Update header sets as a semantic
request boundary, not merely two equivalent encodings.

The created Entries remained unpublished. After they were deleted, subsequent `GET`
requests returned `404`.

## Content Type direct observation

Experiment date unrecorded.

A separate sanitized probe against the CMA compared version-header behavior for Content
Types:

| Target before `PUT` | Version header absent | `X-Contentful-Version: 1` |
| --- | --- | --- |
| Absent | `201`, created at version `1` | `201`, created at version `1` |
| Existing draft at version `1` | `409 VersionMismatch`; subsequent `GET` unchanged | `200`, updated to version `2` |

For the observed Content Types, target existence and exact version together
distinguished Create from Update. Supplying version `1` did not prevent creation at an
absent target.

## Limits

The Entry observation used a single isolated scope and an activated Content Type. It
exercised empty-field draft Entries, absent targets with positive version values `1` and
`7`, and existing drafts at version `1`. It did not establish regional parity or cover
published or archived Entries, nonempty field validation, or malformed, zero, negative,
or stale versions.

The Content Type observation covered absent and draft targets at version `1`. It did not
exercise the separate activation endpoint. Exact statuses, messages, and `1` to `2`
arithmetic are observed behavior rather than permanent API guarantees. The durable
compatibility evidence is the create/update header distinction supported independently
by Contentful's documentation and the pinned JavaScript SDK implementation.
