# Entry null and omission behavior

Contentful's [Entry CMA reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/)
states that empty Entry fields, and the entire `fields` member when empty, are
omitted from responses. The official
[Content Types CMA reference](https://www.contentful.com/developers/docs/references/content-management-api/content-types/)
states that Content Type defaults apply when a field is omitted during Entry
creation, but not during Entry updates. The CMA [overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/#/introduction/updating-and-version-locking)
describes Entry updates as full-body replacement rather than merge operations.

The first-party `contentful-management.js` source at commit
[`cc096a3`](https://github.com/contentful/contentful-management.js/tree/cc096a337f0e1db6114e8da645d69bb6eb90f11c)
passes the complete Entry object to its
[`update` action](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/create-entry-api.ts#L56-L64).
Its REST adapter removes `sys`, sends the remaining body with `PUT`, and uses
the Entry version as
[`X-Contentful-Version`](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L155).
Publishing likewise sends the Entry version to the
[`/published` endpoint](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L168-L185).

On 2026-08-25, an isolated live probe used disposable resources in an existing
test space and environment. The provider binary was built from commit
`27e2830135eeeca88fbac295b3bb731bc23bf388`. A local forwarding proxy recorded
only sanitized method, path, status, header presence, and JSON member types; it
did not record authorization values or field contents.

## Live observations

The disposable Content Type had ordinary optional Symbol fields plus a
separate `defaulted` field whose default was configured for the environment's
default locale.

| Operation | Status | Request fields | Returned fields | `version` / `publishedVersion` |
| --- | ---: | --- | --- | --- |
| `PUT .../entries/{entry}` create | 201 | `title` object, `optional` object, `explicitNull` null; `defaulted` omitted | `title`, `optional`, `defaulted`; `explicitNull` omitted | `1` / absent |
| `GET .../entries/{entry}` | 200 | n/a | Same fields as create response | `1` / absent |
| `PUT .../entries/{entry}/published` | 200 | version `1` | Same fields | `2` / `1` |
| Full-body `PUT .../entries/{entry}` | 200 | Only `title` | Only `title`; prior `optional` and creation default removed | `3` / `1` |
| `GET .../entries/{entry}` | 200 | n/a | Only `title` | `3` / `1` |
| `PUT .../entries/{entry}/published` | 200 | version `3` | Only `title` | `4` / `3` |
| Full-body `PUT .../entries/{entry}` | 200 | `title` object, `explicitNull` null | Only `title`; `explicitNull` omitted | `5` / `3` |
| `GET .../entries/{entry}` | 200 | n/a | Only `title` | `5` / `3` |
| `PUT .../entries/{entry}/published` | 200 | version `5` | Only `title` | `6` / `5` |

The independent create and update probes therefore both accepted an explicit
JSON-null field and omitted it from the mutation response and subsequent GET.
The separate creation-default observation does not imply that null itself
caused defaulting: `explicitNull` had no default, while `defaulted` was omitted
from the request and had a Content Type default. The update proved that an
omitted defaulted field is removed and is not defaulted again.

A second targeted probe sent JSON null for the field that itself had the
Content Type default:

| Operation | Status | Returned fields | `version` / `publishedVersion` |
| --- | ---: | --- | --- |
| Create with `defaulted: null` | 201 | Only `title`; `defaulted` omitted | `1` / absent |
| GET after Create | 200 | Only `title` | `1` / absent |
| Publish | 200 | Only `title` | `2` / `1` |
| Update with `defaulted: null` | 200 | Only `title`; `defaulted` omitted | `3` / `1` |
| GET after Update | 200 | Only `title` | `3` / `1` |

Explicit JSON null therefore suppressed the creation default, while omission
in the first probe applied it. The two request values are observably distinct
even though CMA omitted the field from both the JSON-null response and later
GET.

A final probe distinguished raw field null from a localized field object whose
default-locale value is null. Create and Update each sent both a non-defaulted
field and the defaulted field as `{default-locale: null}`:

| Operation | Status | Returned fields | `version` / `publishedVersion` |
| --- | ---: | --- | --- |
| Create with both localized-null fields | 201 | Both field objects retained the locale key with JSON null | `1` / absent |
| GET after Create | 200 | Both localized-null fields retained | `1` / absent |
| Publish | 200 | Both localized-null fields retained | `2` / `1` |
| Update with both localized-null fields | 200 | Both localized-null fields retained | `3` / `1` |
| GET after Update | 200 | Both localized-null fields retained | `3` / `1` |

The default was not substituted in either lifecycle. A raw field value of JSON
null is therefore response-omitted, while a valid localized object containing
JSON null remains ordinary response data.

## Terraform boundary observation

The same proxy captured a `contentful_entry` Create using the exact provider
commit above. Its effective HCL `fields` map contained:

- `title = jsonencode(...)`;
- `terraformNull = null`; and
- `explicitNull = jsonencode(null)`.

The outbound `PUT` omitted `terraformNull` and included `explicitNull` with the
JSON type `null`. CMA returned HTTP 201 with `title` and the separate creation
default, omitting `explicitNull`. A subsequent GET returned the same field
shape. The provider exited with `Unexpected entry fields` because the returned
empty-field omission differed from the effective Terraform plan; it
checkpointed the response-derived draft at version `1` with no
`publishedVersion`.

Provider impact: Terraform null and encoded JSON null are distinct at the
request boundary. Terraform null means omit the field; encoded JSON null is
sent and can suppress a creation default. At the response boundary, however,
CMA can omit a field that was sent as raw JSON null. Response reconciliation
must account for that documented canonicalization and restore the exact planned
raw JSON-null representation only after verifying omission; it must not treat a
returned non-null value as equivalent to null. A localized object containing a
null locale value is retained by CMA and requires no omission fallback.
Creation defaults remain response additions for genuinely omitted fields, while
full-body updates remove omitted fields.
