# UI Config: identity, configuration, and updates

UI Config stores environment-level views and editing settings. Its environment-scoped
path identifies the configuration even when the response omits `sys.id`.

## Addressing and sources

Published sources reviewed: 2026-09-10.

The [CMA GET reference][get] addresses UI Config at
`/spaces/{space_id}/environments/{environment_id}/ui_config`. The separate User UI Config
applies to the current user and was not sampled in this study. The [UI Config
reference][reference] describes both scopes.

The GET example includes `sys.id` and `sys.organization`. The pinned SDK
[`UIConfigSysProps`][entity] inherits a required `id` from [`BasicMetaSysProps`][common],
but does not declare an organization link. The [SDK adapter][adapter] addresses the
singleton using only the space and environment IDs. These source shapes differ from
the observations below; the example and type do not establish that every read returns
an ID.

## Direct read observations

Observed: 2026-09-09 (UTC), passive reads of existing configuration. A repeated read in
the selected environment returned identical JSON. A separate alias/target comparison
is recorded under [Environment aliases](environment-aliases.md#direct-read-observations).

| Response member | Observed representation |
| --- | --- |
| Top level | `publish`, `timeline`, `homeViews`, `livePreview`, `assetListViews`, `entryListViews`, and `sys` |
| `sys` | `version`, timestamps, actor links, `space`, `environment`, and `type`; no `id` or `organization` |
| `timeline` | Object containing string `mode` and object `timelineConfig` with string `entityScheduling` |
| Views | Optional property combinations included `contentTypeId`, `contentTypeIds`, `displayedFieldIds`, `order`, `roles`, `searchFilters`, and `searchText`; both content-type properties sometimes appeared together |

The current [UI Config reference][reference] documents `timeline.mode` as `disabled` or
`timeline`, and `timeline.timelineConfig.entityScheduling` as `enabled` or `disabled`.
The pinned SDK entity omits `timeline`; this is a source discrepancy, not evidence of
an undocumented server property. Both the current reference and the SDK describe
`contentTypeId` and `contentTypeIds`. The reference requires them to agree when both
are supplied. The read observations establish their coexistence, not that validation
rule's enforcement.

## Preview-mode mutation observations

On 2026-07-14, browser changes and direct CMA requests tested
`livePreview.previewMode` in a primary and a non-primary environment. Direct
requests used a personal management token.

The web app used `PUT /spaces/{space_id}/ui_config` for the primary environment
and the documented environment-explicit path for a non-primary environment. Direct
GETs reached both forms; the shortened response matched the primary environment's
mutable document. Other routing cases were not tested.

The browser changed the mode between `legacyPreview` and `livePreview` by sending
the complete shared document with `X-Contentful-Version`. It preserved unrelated
settings. The [pinned SDK update adapter][update-adapter] also copies the supplied
object, removes `sys`, and sends its version in the header. Omitted fields were not
tested, so these full-document updates leave partial-update behavior unresolved.

| Direct non-primary-environment probe | Recorded result |
| --- | --- |
| Complete-body PUT with stale version | `409 VersionMismatch` |
| Preview-mode change with current version | `200`; version advanced by one |
| Restoration using the returned version | `200`; version advanced by one; a JSON comparison of the mutable fields matched their pre-probe values |

Before restoration, a full web-app reload showed the temporary mode in the
non-primary environment and the original mode in the primary environment. The
dialog had warned that the change would affect all environments. Restoration was
also checked in the UI. This result covers this setting in these two environments.

## Interpretation and limits

UI Config has an environment-scoped singleton identity, independent of an optional
returned ID. The observed [environment link](environment-aliases.md) can identify the
alias used in the request rather than its concrete target.

The September study exercised no updates. Across both studies, nested replacement,
omitted or null branches, reset/default behavior, and changes to views or Timeline
remain unverified. Repeated reads returned equal JSON at the sampled points;
they do not establish stability between those reads or a concurrency guarantee.

[get]: https://www.contentful.com/developers/docs/references/content-management-api/ui-config/get-the-ui-config/
[reference]: https://www.contentful.com/developers/docs/references/content-management-api/ui-config/
[entity]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/ui-config.ts
[common]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/common-types.ts#L445-L453
[adapter]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/ui-config.ts

[update-adapter]: https://github.com/contentful/contentful-management.js/blob/83d6a3f5fcbc9773c21b6779fe66c7fe59349f66/lib/adapters/REST/endpoints/ui-config.ts
