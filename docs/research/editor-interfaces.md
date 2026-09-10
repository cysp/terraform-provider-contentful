# Editor Interfaces: response identity and sidebar values

An Editor Interface configures editing controls for a Content Type. Collection items
and detail responses can differ in system metadata. Sidebar omission and explicit
Boolean values also have distinct response representations.

## Scope and sources

The [Editor Interface CMA
reference](https://www.contentful.com/developers/docs/references/content-management-api/editor-interface/)
describes the environment-scoped configuration at
`/spaces/{space_id}/environments/{environment_id}/content_types/{content_type_id}/editor_interface`.
This study covers response identity and sidebar-member observations, not the complete
Editor Interface schema, widget behavior, or layout contract.

## Collection and detail representations

The [collection reference][collection] addresses
`/spaces/{space_id}/environments/{environment_id}/editor_interfaces` and illustrates a
`sys`, `total`, and `items` envelope without `skip` or `limit`. The [detail
example][detail] returns `sys.id: "default"`. The pinned [SDK adapter][adapter] uses
ordinary `CollectionProp<EditorInterfaceProps>` for the collection and the same entity
type for detail; those annotations do not establish identical wire representations.

Observed: 2026-09-09 (UTC), a passive collection read and two detail comparisons, one
with an explicit sidebar and one without. The collection returned the documented
three-member envelope. Its items omitted `sys.id` and linked their Content Types through
`sys.contentType`. Each sampled detail added only `sys.id: "default"` to its list item;
all other JSON, including versions, was equal.

The space, environment, and Content Type determine the addressed Editor Interface.
Neither a missing list-item ID nor the shared detail ID supplies a separate global
identity. Query-parameter pagination behavior was not tested.

Some returned interfaces omitted `sidebar`; none in the sample contained `groupControls`
or `editorLayout`. Each linked a sampled Content Type, and every returned control's
field ID existed in that model. These static matches do not establish cleanup after
model edits or app removal, or behavior for custom layouts.

## Sidebar update observations

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

[collection]: https://www.contentful.com/developers/docs/references/content-management-api/editor-interface/get-all-editor-interfaces-of-a-space/
[detail]: https://www.contentful.com/developers/docs/references/content-management-api/editor-interface/get-the-editor-interface/
[adapter]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/editor-interface.ts
