# UI Extensions: configuration

An Extension request must contain exactly one source. An empty `srcdoc` is a valid
source; an empty `src` is not. Source selection is part of the Extension request
contract on both creation and update.

## Addressing and operations

Extensions are environment-scoped. POST creates with a generated ID at
`/spaces/{space_id}/environments/{environment_id}/extensions`. PUT creates with a
specified ID or updates at `.../extensions/{extension_id}`; GET and DELETE use that
same individual address. Updates require `X-Contentful-Version`. The selected source
is inside the request's `extension` object. See the [create or update
reference](https://www.contentful.com/developers/docs/references/content-management-api/ui-extensions/create-or-update-an-extension/).

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

The source experiments establish selection and switching, not constraints for all
Extension fields or a complete URL-validation grammar. Contentful recommends App
Framework for new extensibility projects; UI Extensions remain a distinct API surface.
See the [UI Extensions
reference](https://www.contentful.com/developers/docs/references/content-management-api/ui-extensions/).


## Field types and parameter definitions

The [UI Extensions CMA reference](https://www.contentful.com/developers/docs/references/content-management-api/ui-extensions/)
lists scalar field definitions, Link definitions with Entry or Asset targets,
and Array definitions with Symbol or Link items. Its parameter link currently
redirects to the [App Parameters guide](https://www.contentful.com/developers/docs/extensibility/app-framework/app-parameters/),
which describes type-matched defaults and Enum choices. The redirect alone does
not establish identical Extension and App Definition validation.

Direct CMA probes on 2026-11-08 created uniquely named disposable Extensions with
empty `srcdoc`. All fourteen published field definitions were accepted together.
An Array missing items, a Link missing its target, Symbol with items, Array of
Number, and an unknown field type were rejected with HTTP 422.

Parameter controls and rejection probes established the following storage behavior:

- Symbol and Secret string defaults, Number numeric defaults, and Boolean false
  were accepted; tested defaults with a different JSON type were rejected.
  Object, array, and JSON-null defaults were rejected for the tested Symbol
  declaration.
- Enum required nonempty options. String options and single-member objects with
  string labels were accepted. Mixed forms, empty string options, empty or
  multi-member objects, and non-string labels were rejected. Empty string labels
  were accepted.
- Enum defaults matched option values, including object keys for labeled options.
  Nonmember defaults and a display label used as the default were rejected.
- Symbol, Number, Boolean, and Secret with options were rejected. Instance
  controls also confirmed Number default typing and Enum membership/form rules.
  A Secret instance parameter was accepted;
  the provider therefore does not prohibit it. This says nothing about secret
  redaction or access guarantees.

Accepted fixtures were preserved on immediate and subsequent GETs; rejected
fixture IDs remained absent. Every created fixture was deleted and cleanup
verified with GET returning HTTP 404. These probes establish definition storage,
not UI rendering or SDK behavior. Default/type and Enum restrictions are specific
to Extensions and must not be inferred for App Definitions or App Actions.

The provider leaves unknown type names and link targets to Contentful. Structural
validation is not a promise of service support. Parameter counts, identifier
uniqueness, label rendering, and complete service validation remain outside this
local validation contract.
