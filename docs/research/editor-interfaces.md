# Editor Interfaces: sidebar disabled values

An Editor Interface configures editing controls for a Content Type. In the retained
update observations, omitted `sidebar[].disabled` and explicit `false` produced
different response representations. Omission did not become a returned Boolean default.

## Scope and sources

The [Editor Interface CMA
reference](https://www.contentful.com/developers/docs/references/content-management-api/editor-interface/)
describes the environment-scoped configuration at
`/spaces/{space_id}/environments/{environment_id}/content_types/{content_type_id}/editor_interface`.
This study covers only the sidebar-member observation, not the complete Editor Interface
schema, widget behavior, or layout contract.

## Direct observations

Experiment date unrecorded; results retained by 2026-09-03.

A disposable Content Type was created and activated to make its Editor Interface
available. Two separate Editor Interface updates compared an omitted sidebar `disabled`
member with explicit `false`.

| Request | Status | Structural observation |
| --- | ---: | --- |
| Editor Interface update with `sidebar[].disabled` omitted | 200 | Response omitted `disabled` |
| Editor Interface update with `sidebar[].disabled: false` | 200 | Response contained Boolean `disabled: false` |

## Interpretation and limits

A client must distinguish absent `disabled` from returned `false` when retaining the
response representation. The observations do not establish a difference in rendered
sidebar behavior, or the default a future response might supply. Empty sidebars,
explicit null, other widget properties, and concurrency were not part of these probes.
