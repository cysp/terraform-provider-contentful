# Locales: resource identity and raw responses

A Locale's API resource ID and locale code have different roles. The ID addresses the
resource; `code` identifies localized content. Raw CMA responses and SDK-wrapped
objects can also expose different properties.

## Addressing and sources

The pinned [SDK adapter][adapter] reads a Locale at
`/spaces/{space_id}/environments/{environment_id}/locales/{locale_id}` and lists Locales
at the parent `/locales` path. The [CMA Locale reference][reference] describes `code`,
`name`, and `fallbackCode`; the fallback refers to a locale code rather than a resource
ID.

The pinned [SDK entity][entity] declares `internal_code` in `LocaleProps`, but
`wrapLocale` removes that property when constructing the wrapped object. This is a
client projection and does not establish omission from a raw API response.

## Direct read observations

Observed: 2026-09-09 (UTC), passive list and detail reads of an existing default Locale.
The list item and detail response were equal. The resource's `sys.id` differed from
its `code`; `internal_code` was present and `fallbackCode` was explicitly null. The
response also contained `name`, `default`, `optional`, `contentManagementApi`,
`contentDeliveryApi`, and `sys`.

The [alias comparison](environment-aliases.md#direct-read-observations) records how the
returned environment link depends on the addressed path.

## Limits

The sampled null fallback is a returned value, not evidence of how omission or null
behaves during an update. No locale rename, fallback change, reset, deletion,
cloning, or localization effect was tested. A raw `internal_code` value is not evidence
that it is writable or interchangeable with either `code` or `sys.id`.

[reference]: https://www.contentful.com/developers/docs/references/content-management-api/locales/
[adapter]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/locale.ts
[entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/locale.ts
