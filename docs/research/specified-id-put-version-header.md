# Entry specified-ID PUT version-header semantics

## Contract

Contentful uses the same `PUT` endpoint to create an Entry with a specified ID
and to update the existing Entry at that ID. The official
[Entry reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/create-an-entry-with-a-specified-id/)
requires `X-Contentful-Version` when updating an existing Entry.

Contentful's first-party management SDK implements the same distinction at
commit `cc096a337f0e1db6114e8da645d69bb6eb90f11c`. Its Entry
[`createWithId`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L255-L281)
method omits `X-Contentful-Version`; its Entry
[`update`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L155)
method sends the Entry's observed `sys.version`.

## Direct observation

Sanitized probes against the standard global CMA endpoint observed this Entry
request selection:

| Target before `PUT` | Version header absent | `X-Contentful-Version: 1` |
| --- | --- | --- |
| Absent | HTTP 201; created at version 1 | HTTP 201; created at version 1 |
| Existing draft at version 1 | HTTP 409 `VersionMismatch`; subsequent `GET` unchanged | HTTP 200; updated to version 2 |

The create requests included `X-Contentful-Content-Type`. That header did not
change the version-header behavior.

## Provider invariant

Creating with a configured `entry_id` is create-only: the provider omits
`X-Contentful-Version`. Updating an existing managed Entry sends its exact
observed `sys.version`. A specified-ID Create collision must fail without
mutating, publishing, or adopting the existing Entry. A preflight read cannot
provide this guarantee because another writer could create the ID between the
read and the `PUT`.

## Limits

The direct Entry probes covered one CMA account, space, and environment through
the standard global endpoint. They covered absent targets and draft targets at
version 1, not the EU endpoint, other versions, or published or archived
Entries. The exact status codes and `1` to `2` version arithmetic are observed
behavior; the durable compatibility requirement is the create/update
version-header distinction supported by the official reference and SDK.
