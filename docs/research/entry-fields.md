# Entry fields: defaults, null, and omission

Entry field omission, a raw JSON-null field, and a localized object containing null have
different effects. In direct CMA observations, omission applied a Content Type default
during creation; raw JSON null suppressed that default and was omitted from responses;
localized null remained present. Full-body updates removed omitted fields without
reapplying creation defaults.

## Published evidence

Contentful's [Entry CMA
reference](https://www.contentful.com/developers/docs/references/content-management-api/entries/)
states that the Get an entry response omits empty Entry fields and omits the entire
`fields` member when empty. The official [Content Types CMA
reference](https://www.contentful.com/developers/docs/references/content-management-api/content-types/)
states that Content Type defaults apply when a field is omitted during Entry creation,
but not during Entry updates. The CMA
[overview](https://www.contentful.com/developers/docs/references/content-management-api/overview/#updating-content)
describes Entry updates as full-body replacement rather than merge operations.

The first-party `contentful-management.js` source at commit
[`cc096a3`](https://github.com/contentful/contentful-management.js/tree/cc096a337f0e1db6114e8da645d69bb6eb90f11c)
passes the complete Entry object to its [`update`
action](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/create-entry-api.ts#L56-L64).
Its REST adapter deep-copies the Entry, removes `sys`, and sends the remaining body with
`PUT`; see the [adapter](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/adapters/REST/endpoints/entry.ts#L133-L155).
The [PUT header reference](entry-and-content-type-put-headers.md) covers Create/Update
selection; [Entry lifecycle](entry-lifecycle.md) covers publication and version behavior.

## Live observations

Observed: 2026-08-25.

An isolated live probe used disposable resources in an existing test scope. A forwarding
proxy recorded only sanitized method, path, status, header presence, and JSON member
types; it did not record authorization values or field contents.

The disposable Content Type had ordinary optional Symbol fields plus a separate
`defaulted` field whose default was configured for the environment's default locale.

| Operation | Status | Request fields | Returned fields | `version` / `publishedVersion` |
| --- | ---: | --- | --- | --- |
| `PUT .../entries/{entry_id}` create | 201 | `title` object, `optional` object, `explicitNull` null; `defaulted` omitted | `title`, `optional`, `defaulted`; `explicitNull` omitted | `1` / absent |
| `GET .../entries/{entry_id}` | 200 | n/a | Same fields as create response | `1` / absent |
| `PUT .../entries/{entry_id}/published` | 200 | version `1` | Same fields | `2` / `1` |
| Full-body `PUT .../entries/{entry_id}` | 200 | Only `title` | Only `title`; prior `optional` and creation default removed | `3` / `1` |
| `GET .../entries/{entry_id}` | 200 | n/a | Only `title` | `3` / `1` |
| `PUT .../entries/{entry_id}/published` | 200 | version `3` | Only `title` | `4` / `3` |
| Full-body `PUT .../entries/{entry_id}` | 200 | `title` object, `explicitNull` null | Only `title`; `explicitNull` omitted | `5` / `3` |
| `GET .../entries/{entry_id}` | 200 | n/a | Only `title` | `5` / `3` |
| `PUT .../entries/{entry_id}/published` | 200 | version `5` | Only `title` | `6` / `5` |

The independent create and update probes therefore both accepted an explicit JSON-null
field and omitted it from the mutation response and subsequent GET. The separate
creation-default observation does not imply that null itself caused defaulting:
`explicitNull` had no default, while `defaulted` was omitted from the request and had a
Content Type default. The update proved that an omitted defaulted field is removed and
is not defaulted again.

A second targeted probe sent JSON null for the field that itself had the Content Type
default:

| Operation | Status | Returned fields | `version` / `publishedVersion` |
| --- | ---: | --- | --- |
| Create with `defaulted: null` | 201 | Only `title`; `defaulted` omitted | `1` / absent |
| GET after Create | 200 | Only `title` | `1` / absent |
| Publish | 200 | Only `title` | `2` / `1` |
| Update with `defaulted: null` | 200 | Only `title`; `defaulted` omitted | `3` / `1` |
| GET after Update | 200 | Only `title` | `3` / `1` |

Explicit JSON null therefore suppressed the creation default, while omission in the
first probe applied it. The two request values are observably distinct even though CMA
omitted the field from both the JSON-null response and later GET.

A final probe distinguished raw field null from a localized field object whose
default-locale value is null. Create and Update each sent both a non-defaulted field and
the defaulted field as `{default-locale: null}`:

| Operation | Status | Returned fields | `version` / `publishedVersion` |
| --- | ---: | --- | --- |
| Create with both localized-null fields | 201 | Both field objects retained the locale key with JSON null | `1` / absent |
| GET after Create | 200 | Both localized-null fields retained | `1` / absent |
| Publish | 200 | Both localized-null fields retained | `2` / `1` |
| Update with both localized-null fields | 200 | Both localized-null fields retained | `3` / `1` |
| GET after Update | 200 | Both localized-null fields retained | `3` / `1` |

The default was not substituted in either lifecycle. A raw field value of JSON null is
therefore response-omitted, while a valid localized object containing JSON null remains
ordinary response data.

## Interpretation and limits

These experiments distinguish the JSON wire values themselves. A client language's null
or missing-value sentinel must be mapped deliberately to either member omission or a
sent JSON value. A returned omission does not reveal which request representation
produced it, and does not establish that a present non-null value is equivalent to null.

The null probes used optional Symbol fields, including one with a default. They do not
establish every field type's validation or every empty-array and empty-object response
shape. The published GET description covers more empty-value shapes than the directly
exercised raw-null and localized-null cases. Applying the same projection to every
mutation and collection endpoint remains an inference where no matching experiment is
retained.
