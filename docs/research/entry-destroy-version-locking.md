# Entry destroy version-locking evidence

Reviewed 2026-08-29 against Contentful's current CMA reference, the current
first-party JavaScript management client v12.15.0 at commit
[`cc096a3`](https://github.com/contentful/contentful-management.js/tree/cc096a337f0e1db6114e8da645d69bb6eb90f11c),
the first-party .NET client at commit
[`0908893`](https://github.com/contentful/contentful.net/tree/090889393b58113f50af45214d92fe92045c50d5),
and the repository's earlier sanitized live observation.

## Conclusions

- Whole-Entry unpublish and delete both **support** `X-Contentful-Version`.
  Their current endpoint pages show the header in the cURL request, but mark it
  optional: [unpublish](https://www.contentful.com/developers/docs/references/content-management-api/entries/unpublish-an-entry/)
  and [delete](https://www.contentful.com/developers/docs/references/content-management-api/entries/delete-an-entry/).
  The first-party .NET client requires a last-known version for both methods and
  passes it to its version-header-aware DELETE helper
  ([delete](https://github.com/contentful/contentful.net/blob/090889393b58113f50af45214d92fe92045c50d5/Contentful.Core/ContentfulManagementClient.cs#L674-L688),
  [unpublish](https://github.com/contentful/contentful.net/blob/090889393b58113f50af45214d92fe92045c50d5/Contentful.Core/ContentfulManagementClient.cs#L713-L732),
  [header helper](https://github.com/contentful/contentful.net/blob/090889393b58113f50af45214d92fe92045c50d5/Contentful.Core/ContentfulClientBase.cs#L270-L281)).
- The current JavaScript client disagrees with that safe usage: whole-Entry
  [delete and unpublish](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L158-L209)
  omit the version header, although locale-based unpublish sends it. This and
  the reference's `Optional` marker mean that current primary sources do not
  prove that an omitted header is rejected. Tests should therefore require the
  provider's exact versioned request; a normal mock must not claim omission is
  rejected without stronger evidence.
- Successful unpublish returns HTTP 200 and an Entry. Both first-party clients
  decode it as an Entry; the JavaScript client's
  [endpoint type](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/common-types.ts#L2311-L2320)
  makes `EntryProps` the unpublish result. The current CMA page's example is
  structurally incomplete because it omits `sys`, so it does not establish a
  returned version value. The repository's sanitized
  [live observation](entry-unpublish-version.md) found that whole-Entry
  unpublish advanced `sys.version` by one, removed `sys.publishedVersion`, and a
  subsequent GET returned the same new version. Exact `+1` arithmetic is not a
  documented guarantee: the returned positive `sys.version`, not a predicted
  value or the pre-unpublish version, is the evidence available for a following
  delete.
- Contentful's [version-locking contract](https://www.contentful.com/developers/docs/references/content-management-api/overview/#updating-and-version-locking)
  says a mutation carrying an out-of-date current version is rejected. Its
  [error reference](https://www.contentful.com/developers/docs/references/errors/)
  classifies a missing or outdated current version for an Entry as HTTP 409
  `VersionMismatch`. The references do not repeat this error table on the two
  Entry DELETE endpoint pages, so the endpoint-specific classification is
  strong shared-contract evidence rather than a retained live observation.
- The error reference defines HTTP 404 `NotFound` as a missing resource or
  endpoint. Contentful's [authentication reference](https://www.contentful.com/developers/docs/references/authentication/)
  also notes that an inaccessible resource may deliberately appear as 404.
  Thus an already-absent Entry can make destroy idempotently complete, but 404
  is not evidence that a version conflict occurred and must not cause a retry
  with a refreshed version. The primary sources reviewed do not specify the
  exact response for unpublishing an existing but already-unpublished Entry.

## Provider consequence

When Terraform has a last-observed Entry version, unpublish or delete should
send that exact value and surface a stale conflict without refreshing it into
new deletion authority. A published destroy should delete only with the version
returned by successful unpublish. This protects both concurrency windows
without assuming version arithmetic.

## Evidence gaps

No local live CMA credentials were available. A minimal disposable probe would
still be needed to establish (1) whether omission is currently accepted or
rejected on each endpoint, (2) the exact stale error payload for each endpoint,
and (3) the already-unpublished response. None of those gaps prevents the
provider from sending and testing exact version fences; they do constrain what
the default mock may claim as normal CMA behavior.
