# Specified-ID PUT version-header semantics

## Contract

Contentful uses the same `PUT` endpoint to create an Entry or Content Type with
a specified ID and to update the existing resource at that ID. The official
[Entry reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/create-an-entry-with-a-specified-id/)
and
[Content Type reference](https://www.contentful.com/developers/docs/references/content-management-api/content-types/create-a-content-type-with-a-specified-id/)
require `X-Contentful-Version` when updating the existing resource.

Contentful's first-party management SDK implements the same distinction at
commit `cc096a337f0e1db6114e8da645d69bb6eb90f11c`. Entry
[`createWithId`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L255-L281)
and Content Type
[`createWithId`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/content-type.ts#L71-L80)
omit `X-Contentful-Version`; their
[`Entry.update`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L155)
and
[`ContentType.update`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/content-type.ts#L82-L96)
methods send the observed `sys.version`.

## Direct observation

Sanitized probes against the standard global CMA endpoint observed the same
request selection for Entries and Content Types:

| Target before `PUT` | Version header absent | `X-Contentful-Version: 1` |
| --- | --- | --- |
| Absent | Both: HTTP 201; created at version 1 | Both: HTTP 201; created at version 1 |
| Existing draft at version 1 | Both: HTTP 409 `VersionMismatch`; subsequent `GET` unchanged | Both: HTTP 200; updated to version 2 |

The Entry create requests included `X-Contentful-Content-Type`; that header did
not change the version-header behavior.

## Provider invariants

- Creating `contentful_entry` with a configured `entry_id` omits
  `X-Contentful-Version`. A collision must fail without mutating, publishing, or
  adopting the existing Entry.
- Creating `contentful_content_type` with its configured `content_type_id` omits
  `X-Contentful-Version`. A collision must fail without mutating, activating, or
  adopting the existing Content Type.
- Updating either managed resource sends its exact observed `sys.version`.

A preflight read cannot guarantee create-only behavior because another writer
could create the ID between the read and the `PUT`.

## Limits

The direct probes covered one CMA account, space, and environment through the
standard global endpoint. They covered absent targets and draft targets at
version 1, not the EU endpoint, other versions, or published or archived
resources. The Content Type probes exercised the draft `PUT`, not the separate
activation endpoint. The exact status codes and `1` to `2` version arithmetic
are observed behavior; the durable compatibility requirement is the
create/update version-header distinction supported by the official references
and SDK.
