# Entry destroy version-locking evidence

Reviewed 2026-08-30 against Contentful's CMA reference, the current
first-party JavaScript management client v12.15.0 at commit
[`cc096a3`](https://github.com/contentful/contentful-management.js/tree/cc096a337f0e1db6114e8da645d69bb6eb90f11c),
the first-party .NET client at commit
[`0908893`](https://github.com/contentful/contentful.net/tree/090889393b58113f50af45214d92fe92045c50d5),
and a disposable live CMA probe. The live service is authoritative where it
contradicts an inference from the general CMA version-locking documentation.

## Conclusions

- The whole-Entry unpublish and delete endpoint pages show
  `X-Contentful-Version` in their cURL requests but mark it optional:
  [unpublish](https://www.contentful.com/developers/docs/references/content-management-api/entries/unpublish-an-entry/)
  and [delete](https://www.contentful.com/developers/docs/references/content-management-api/entries/delete-an-entry/).
  The first-party .NET client requires a last-known version for both methods and
  passes it to its version-header-aware DELETE helper
  ([delete](https://github.com/contentful/contentful.net/blob/090889393b58113f50af45214d92fe92045c50d5/Contentful.Core/ContentfulManagementClient.cs#L674-L688),
  [unpublish](https://github.com/contentful/contentful.net/blob/090889393b58113f50af45214d92fe92045c50d5/Contentful.Core/ContentfulManagementClient.cs#L713-L732),
  [header helper](https://github.com/contentful/contentful.net/blob/090889393b58113f50af45214d92fe92045c50d5/Contentful.Core/ContentfulClientBase.cs#L270-L281)).
  Conversely, the JavaScript client's whole-Entry
  [delete and unpublish](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L158-L209)
  omit the header, although locale-based unpublish sends it.
- A repeated live probe established that `X-Contentful-Version` is **not an
  optimistic-lock precondition** for the whole-Entry DELETE endpoints in the
  probed CMA environment. Entry unpublish and delete both accepted stale and
  zero values. Both also succeeded when the header was omitted. In particular,
  stale delete returned 204 and removed the Entry, while stale unpublish
  returned 200 and unpublished it. This contradicts applying the CMA's general
  [version-locking contract](https://www.contentful.com/developers/docs/references/content-management-api/overview/#updating-and-version-locking)
  or shared HTTP 409
  [`VersionMismatch`](https://www.contentful.com/developers/docs/references/errors/)
  description to these two endpoints.
- Successful whole-Entry unpublish returns HTTP 200 with the resulting Entry.
  Exact-current unpublish advanced `sys.version` from 2 to 3 and removed
  `sys.publishedVersion`. Stale unpublish, sent with version 2 while the Entry
  was at version 3, advanced it to version 4 and removed
  `sys.publishedVersion`. A following GET returned the same resulting version.
  The returned version remains the only sound source for a subsequent request;
  exact `+1` arithmetic is still not a documented guarantee.
- Unpublishing an existing already-unpublished Entry with its exact current
  version returned HTTP 400, `sys.id` `BadRequest`, and message `Not published`.
- Delete returned HTTP 204 with no body. After deletion, GET, delete, and
  unpublish each returned HTTP 404, `sys.id` `NotFound`, and message
  `The resource could not be found.` Contentful's
  [authentication reference](https://www.contentful.com/developers/docs/references/authentication/)
  also notes that an inaccessible resource may deliberately appear as 404, so
  destroy may treat 404 as absence but must not classify other errors that way.

## Disposable live probe

The probe ran at 2026-08-29 14:05 UTC (2026-08-30 local time) in the space and
environment configured by `terraform-provider-contentful-example`. It loaded
the management token without logging it. It created one disposable content
type and eight Entries. The canonical resource prefix was
`codex-entry-destroy-probe-20260829140508`; paths below are relative to
`/spaces/{configured-space}/environments/{configured-environment}` and `{p}` is
that exact prefix.

| Case | Exact request | Remote version before request | Result |
| --- | --- | ---: | --- |
| Stale unpublish | `DELETE /entries/{p}-stale-unpublish/published`, `X-Contentful-Version: 2` | 3, `publishedVersion: 1` | 200; returned Entry version 4 with no `publishedVersion`; GET returned the same Entry at version 4 |
| Exact-current unpublish | `DELETE /entries/{p}-current-unpublish/published`, `X-Contentful-Version: 2` | 2, `publishedVersion: 1` | 200; returned Entry version 3 with no `publishedVersion` |
| Already-unpublished unpublish | `DELETE /entries/{p}-current-unpublish/published`, `X-Contentful-Version: 3` | 3, unpublished | 400; `BadRequest`; `Not published` |
| Stale delete | `DELETE /entries/{p}-delete`, `X-Contentful-Version: 1` | 2 | 204; following GET returned 404, so the externally advanced Entry was deleted |
| Exact-current delete | `DELETE /entries/{p}-current-delete`, `X-Contentful-Version: 1` | 1 | 204; following GET returned 404 |
| Omitted-version delete | `DELETE /entries/{p}-omit-delete`, no version header | 1 | 204; cleanup verification GET returned 404 |
| Omitted-version unpublish | `DELETE /entries/{p}-omit-unpublish/published`, no version header | 2, `publishedVersion: 1` | 200; returned Entry version 3 with no `publishedVersion` |
| Zero-version delete | `DELETE /entries/{p}-zero-delete`, `X-Contentful-Version: 0` | 1 | 204; following GET returned 404 |
| Zero-version unpublish | `DELETE /entries/{p}-zero-unpublish/published`, `X-Contentful-Version: 0` | 2, `publishedVersion: 1` | 200; returned and subsequently fetched Entry version 3 with no `publishedVersion` |
| Delete after absence | `DELETE /entries/{p}-delete`, `X-Contentful-Version: 2` | absent | 404; `NotFound`; `The resource could not be found.` |
| Unpublish after absence | `DELETE /entries/{p}-delete/published`, `X-Contentful-Version: 2` | absent | 404; `NotFound`; `The resource could not be found.` |

The stale-unpublish result was first discovered in a preliminary run and then
repeated in two complete runs. The complete probe was also repeated once for
the delete, omission, zero, already-unpublished, and 404 cases with the same
statuses and state transitions.

Cleanup used a fresh GET, exact returned versions, and required unpublish or
content-type deactivation transitions. Every Entry from prefixes
`codex-entry-destroy-probe-20260829140123`,
`codex-entry-destroy-probe-20260829140254`, and
`codex-entry-destroy-probe-20260829140508` returned 404 after cleanup. Each
disposable content type was deactivated, deleted with HTTP 204, and verified by
GET 404.

## Provider consequence

Supplying the last-observed version on whole-Entry unpublish and delete does not
provide the intended concurrency protection against the live CMA behavior.
With the proposed destroy sequence, an external edit before unpublish still
allows unpublish to succeed; its returned version is then passed to delete, and
delete does not enforce that value either. An external edit between unpublish
and delete can likewise still be deleted.

The same limitation applies to any missing-private-state fallback that performs
GET followed by a versioned delete: because delete ignores the version, the GET
to mutation interval is not fenced. A present private value of zero is also not
safe merely because it is sent: the live endpoint accepted zero and performed
the destructive mutation. This also applies to JSON `null`, which the current
`optionalPrivateVersion` decoder accepts as a present Go integer zero; that
value can reach Entry destroy rather than being diagnosed as malformed private
version state.

The mock must model this observed normal CMA behavior. Rejecting omitted, stale,
or zero versions would be stricter than Contentful and may only be isolated as
explicit adversarial fault injection. The intended optimistic-lock destroy
contract cannot be claimed or tested as normal CMA behavior using these
headers; implementing a different provider-side policy requires a separate
design decision that acknowledges the remaining GET-to-mutation race.

## Remaining limitation

This is a time- and environment-specific live observation, not a documented
guarantee that Contentful will ignore the header forever or in every account.
However, it directly disproves relying on the header for current practitioner
safety. A direct refresh of the official endpoint pages during the probe was
rate-limited with HTTP 429; the first-party client sources were re-fetched and
inspected successfully, and their contradictory header choices remain as cited
above.
