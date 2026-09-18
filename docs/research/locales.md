# Locales: identity, configuration, and lifecycle

Locales are environment-level entities. Their system IDs identify resources;
mutable codes label localized content and `fallbackCode` refers to a code.
The [provider design](../design/locale.md) defines Terraform's use of these APIs.

## Sources and evidence

The [CMA reference][reference], [localization tutorial][localization], and
[settings guide][settings] describe the supported operations and constraints.
The pinned [SDK adapter][adapter] uses environment-scoped POST, GET, PUT, and
DELETE, supplies `sys.version` on PUT, and sends no version on DELETE. The
[SDK entity][entity] removes `internal_code` from wrapped objects; raw CMA
responses included it. It is not a provider input or resource identifier.

The observations below retain the findings of CMA/Terraform and independent
CDA/CPA experiments from 2026-09-09–18. They used two environments in one space,
disposable Entries and an Asset, and entry-level publishing. They are samples,
not service guarantees; this document does not establish quota accounting,
propagation deadlines, or locale-based publishing semantics.

## Identity and discovery

CMA detail and collection items were equal, including integer `sys.version`
and explicit null fallback. Codes used as IDs, missing IDs, and missing
environments returned 404. Disabling either API flag left the Locale entity
discoverable. [Alias reads](environment-aliases.md#direct-read-observations)
returned an environment link matching the addressed path.

Positive `limit`/`skip` selected pages, including an empty page past the end.
Locale requests with `order`, filters, and `select` returned 200 without the
requested effects. The [collection reference][collection] documents no sort
keys. Neither sample ordering nor general CMA pagination guidance establishes
a Locale ordering contract. Select the default by its flag, not position.
With two Locales, `limit=0` returned both with envelope limit 100. Negative
limits/offsets produced unusual slices; no general negative-index contract
was established.

## Mutation constraints and defaults

Published guidance requires unique codes, versioned updates, immutable default
status, acyclic fallbacks, and explicit null to clear a fallback. It restricts
renaming/disabling fallback targets and prohibits deleting the default locale.
Contentful remains authoritative for validation and capacity.

| Observed request | Result |
| --- | --- |
| POST with only name and code | Version 1, both API flags true, optional false, default locale's code as fallback |
| Explicit null fallback | Remained null |
| Missing, empty, self, or non-null default fallback | 422 `ValidationFailed`; no mutation |
| Duplicate code | POST 422; PUT 500 `ServerError`; no mutation |
| Default DELETE, with/without incoming fallback | 500 `ServerError`; no mutation |
| Nondefault DELETE without dependents | 204, subsequent detail 404; recreation generated a new ID |
| Complete identical PUT, including stale/future version | Response and version unchanged |
| Changed name with stale/future version | 409 `VersionMismatch`; no mutation |
| Changed name with current version | Version advanced once; repeating it did not |
| Empty/whitespace-only name | POST rejected; PUT preserved the submitted value |
| Invalid code syntax on PUT, even with stale version | 422; diagnostics specified length 2–11 and `^[a-zA-Z0-9-]{2,11}$` |

An otherwise identical PUT including `default` left status/version unchanged.
Default and nondefault renames and staged code swaps retained system IDs,
`internal_code`, and default status. No writable default-selection mechanism
was established. A referenced target must be disconnected before renaming.

## Editing and delivery flags

Observed on nondefault locales: disabling editing hid Entry/Asset translations
and Content Type default-value keys from CMA; disabling delivery hid translations
from CDA/CPA. Neither changed content versions. Re-enabling editing restored
hidden values, but writing back only visible fields permanently removed them.
Writes containing an editing-disabled code were rejected. Published content
could continue delivering the older value until republished.

CMA accepted disabling default delivery while referenced and selecting that
disabled default as fallback, an exception to the tutorial's general restriction.
Delivery propagation prevented verification of eventual default-flag effects.
Selection of a disabled nondefault fallback target was not isolated live.

## Content effects and publication

Observed code changes relabeled CMA Entry/Asset field maps, Entry locale status,
and Content Type default-value keys without content writes or version changes.
Deletion removed translations and locale publication status; recreating the code
restored neither. Response relabeling does not establish a stored migration.

With entry-level publishing, a populated fallback did not satisfy required
nondefault translations. `optional` permitted incomplete nondefault translations,
but did not relax required default translations. Disabling editing could relax
localized requirements; an entirely absent required field still failed. These
observations do not cover locale-based publishing, where the [settings guide][settings]
says the optional-locale setting is unavailable. Required-field controls used
Symbol fields without additional validations. Explicit null failed required
nondefault validation; an empty string succeeded. Content Type field `disabled`
and `omitted` flags did not reproduce the Locale editing-disabled exemption.

Absent translations resolved through fallback; explicit null and empty strings
did not. After deletion/recreation, ordinary and publication-header-enabled CDA
reads sometimes disagreed. With fallback configured, ordinary CDA returned the
fallback value while requests using `X-Contentful-Locale-Based-Publishing: true`
returned 404 without restored locale publication status. Null-fallback requests
also returned 200 with empty fields without restored status. Republishing restored
visibility in a disposable control, but these counterexamples do not establish
a universal header rule or verify the [announced backward compatibility][publishing-header].

## Cross-API propagation

CMA reads and successful mutations could disagree with CDA/CPA and creation
validation after renames. Fresh URLs, cache-miss responses, and new credentials
did not establish convergence; sampled versions could regress. Quota errors also
varied across environments and restoration attempts. These observations supply
neither a reliable delivery waiter nor a quota preflight algorithm. Distinct-code
counting fit some controls but did not explain later failures; a generic
`AccessDenied` response is not necessarily a quota error.

## Limits

Cloning, mutation responses through aliases, locale-based publishing, multi-hop
fallbacks, and disabling/deleting referenced nondefault targets were not verified
live. Exact validation precedence outside the samples is unknown. The
[test server contract](../design/cma-test-server-conformance.md#locales) distinguishes
its Locale-entity model from unmodeled content effects and service observations.

[reference]: https://www.contentful.com/developers/docs/references/content-management-api/locales/
[collection]: https://www.contentful.com/developers/docs/references/content-management-api/locales/get-all-locales-of-a-space/
[adapter]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/locale.ts
[entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/locale.ts
[localization]: https://www.contentful.com/developers/docs/tutorials/general/setting-locales/
[settings]: https://www.contentful.com/help/localization/manage-locales/
[publishing-header]: https://www.contentful.com/developers/api-changes/locale-based-publishing-request-header-for-the-content-delivery-api-and-graphql-api/
