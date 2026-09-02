# Entry and Content Type PUT header semantics

This note distinguishes Contentful's documented contract, pinned JavaScript
client behavior, and sanitized direct observations of the standard global
Content Management API (CMA).

## Documented contract

Contentful uses the same Entry `PUT` endpoint to create an Entry with a
specified ID and to update an existing Entry. The official
[Entry reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/)
says that Create requires `X-Contentful-Content-Type`, while Update requires the
last `X-Contentful-Version` and does not require the Content Type header. The
[specified-ID endpoint reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/create-an-entry-with-a-specified-id/)
describes the same distinction.

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

## Entry direct observation (2026-09-02)

A sanitized probe against the standard global CMA used an existing activated
Content Type, represented below as `fixture`. The probe did not create,
activate, or modify that Content Type.

Every request used the member path
`PUT /spaces/{space_id}/environments/{environment_id}/entries/{entry_id}`, the
header `Content-Type: application/vnd.contentful.management.v1+json`, and the
body `{"fields":{}}`. Every target had a randomized disposable ID and returned
`404 NotFound` to a preflight `GET`. Successful requests returned
`sys.contentType.sys.id` corresponding to `fixture`, an empty `fields` object,
and no `publishedVersion`.

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

A separate direct collision observation sent an existing version-`1` Entry
with Content Type and no Version. CMA returned `409 VersionMismatch`, and a
subsequent `GET` showed that the Entry was unchanged.

Every failed absent-target `PUT` shown in the table returned error `sys.id`
`BadRequest`, the message quoted in the table, and null `details`. Its follow-up
`GET` returned `404 NotFound`, so the failed request did not create the target.

For the observed absent targets, Content Type presence selected creation:
versions `1` and `7` did not prevent creation when Content Type was present,
and neither version authorized creation without Content Type. For the observed
existing drafts, the exact returned version selected Update and the Content
Type header made no difference. The observations therefore distinguish the
Create and Update header sets as a semantic request boundary, not merely two
equivalent encodings.

All disposable Entries created for these observations were deleted. Subsequent
`GET` requests returned `404`; none of the Entries was published, and no probe
Entry remained.

## Content Type direct observation

A separate sanitized probe against the standard global CMA compared
version-header behavior for Content Types:

| Target before `PUT` | Version header absent | `X-Contentful-Version: 1` |
| --- | --- | --- |
| Absent | `201`, created at version `1` | `201`, created at version `1` |
| Existing draft at version `1` | `409 VersionMismatch`; subsequent `GET` unchanged | `200`, updated to version `2` |

For the observed Content Types, target existence and exact version together
distinguished Create from Update. Supplying version `1` did not prevent
creation at an absent target.

## Limits

The Entry observation covered one CMA account, space, environment, activated
Content Type, and the standard global endpoint. It exercised empty-field draft
Entries, absent targets with positive version values `1` and `7`, and existing
drafts at version `1`. It did not cover the EU endpoint, published or archived
Entries, nonempty field validation, or malformed, zero, negative, or stale
versions.

The Content Type observation covered absent and draft targets at version `1`.
It did not exercise the separate activation endpoint. Exact statuses, messages,
and `1` to `2` arithmetic are observed behavior rather than permanent API
guarantees. The durable compatibility evidence is the create/update header
distinction supported independently by Contentful's documentation and the
pinned JavaScript SDK implementation.
