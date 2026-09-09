# UI Extensions: source values

An Extension request must contain exactly one source. An empty `srcdoc` is a valid
source; an empty `src` is not. Source selection is part of the Extension request
contract on both creation and update.

## Addressing and operations

Extensions are environment-scoped. POST creates at
`/spaces/{space_id}/environments/{environment_id}/extensions`; GET, PUT, and DELETE
address an individual `.../extensions/{extension_id}`. The selected source is inside the
request's `extension` object.

## Published evidence

Contentful's [UI Extensions CMA
reference](https://www.contentful.com/developers/docs/references/content-management-api/ui-extensions/)
requires exactly one of `extension.src` or `extension.srcdoc`. The first-party
[`contentful-management.js`
source](https://github.com/contentful/contentful-management.js/blob/cc096a337f0e1db6114e8da645d69bb6eb90f11c/lib/entities/extension.ts#L54-L62)
models the same create contract with `RequireExactlyOne<...,
"src" | "srcdoc">`.

## Direct CMA observations

Creation probes observed: 2026-08-21.

Isolated live CMA `PUT` requests tested the remaining empty-value ambiguity. The
requests used an empty `fieldTypes` list and `sidebar=false`:

| Request source members | Result | Response evidence |
| --- | --- | --- |
| `"src":""` | HTTP 422 | `extension.src` failed Contentful's URL regexp. |
| `"srcdoc":""` | HTTP 201 | The response retained `"srcdoc":""`; the probe resource was then deleted with HTTP 204. |
| Both empty | HTTP 422 | The response reported both the invalid `src` URL and the source-choice constraint. |

Update probes observed: 2026-08-22.

Reversible live Update probes sent complete Extension `PUT` payloads. A payload with
neither source returned HTTP 422. Switching from `srcdoc` to `src`, and
then from `src` to `srcdoc`, each returned HTTP 200; each response contained only the
newly selected source member.

## Interpretation and limits

Source replacement must send the selected source and omit the other member. The
successful switch responses contained only the newly selected source; serializing both
the old source and its replacement is not a valid update. Explicit empty `srcdoc` is a
value, not the absence of a source.

The experiments establish source selection and switching, not constraints for all
Extension fields or a complete URL-validation grammar. Contentful recommends App
Framework for new extensibility projects; UI Extensions remain a distinct API surface.
See the [UI Extensions
reference](https://www.contentful.com/developers/docs/references/content-management-api/ui-extensions/).
