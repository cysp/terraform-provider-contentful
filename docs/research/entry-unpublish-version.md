# Entry unpublish version behavior

Reviewed 2026-08-26 against Contentful's current CMA documentation, the
first-party JavaScript management client at commit
[`cc096a3`](https://github.com/contentful/contentful-management.js/tree/cc096a337f0e1db6114e8da645d69bb6eb90f11c),
and one sanitized disposable live probe.

## Published evidence

The CMA [Entry reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/)
documents whole-Entry unpublish as `DELETE .../entries/{entry_id}/published`.
It defines `sys.version` as the current version and `sys.publishedVersion` as
the published version, but does not state the tuple returned by unpublish.

The first-party client's whole-Entry
[`unpublish`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L187-L209)
uses the same unversioned `DELETE` and decodes the response as an Entry. It does
not predict or validate the resulting version relationship. This is strong
first-party evidence for the request and response shape, but not for lifecycle
arithmetic.

## Sanitized live observation

A disposable Content Type and Entry were created in an isolated test scope.
The Entry was published, then updated once to leave a newer draft while the
older version remained published. No identifiers, field values, locale,
credentials, or raw responses were retained.

| Operation | HTTP status | Structural observation |
| --- | --- | --- |
| Create Entry | 201 | `version` was positive |
| Publish Entry | 200 | `publishedVersion` equalled the submitted version; current `version` advanced by one |
| Update Entry | 200 | current `version` advanced by one; `publishedVersion` remained positive and older |
| Unpublish Entry | 200 | current `version` advanced by one from the pending draft; `publishedVersion` was absent |
| GET Entry after unpublish | 200 | returned the same advanced version as unpublish; `publishedVersion` remained absent |

This is a direct observation of the current service, not a documented promise
that all future unpublish responses must use exact `+1` arithmetic. It does
establish that the tested whole-Entry unpublish created a lifecycle event newer
than the pending draft.

## Provider and fake-CMA consequence

After unpublish, remote `version` no longer equals the earlier draft version.
Terraform must not publish that stale draft or treat the newly observed version
as authority to publish. The provider grants publication authority only to the
Create or Update operation that received its validated draft response, so the
observed external transition neither revokes a stored marker nor grants authority
to a later operation.

The in-process fake models this observed contract: Entry unpublish returns a
200 Entry response with an advanced positive `version` and no
`publishedVersion`. Deliberately malformed or ambiguous alternatives belong in
explicit fault adapters rather than the normal handler.
