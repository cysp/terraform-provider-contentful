# App Definition validation evidence

These observations explain why App Definition validation deliberately differs from
some published guides and first-party client constraints. The
[provider validation boundaries](../../design/terraform-value-semantics.md#app-definition-configuration)
own the implementation policy; this note records the external evidence and its limits.

## Controlled CMA observations

On **2026-11-07**, two runs against disposable, never-installed AppDefinitions in one
organization exercised **96 PUT cases: 52 accepted and preserved, 44 rejected**.
The [sanitized results](schema-validation-live-results.json) retain requests, controls,
mutation responses, immediate and later GETs, error details, and cleanup results.
They are the inspectable evidence for the table below, not replay-ready requests:
identifiers and `sys` were removed, synthetic names replaced, and JSON types/order
preserved. The eight/nine-declaration controls used distinct IDs.

Both runs used `api.contentful.com`, `src=https://example.invalid/app`, no bundle,
and `app-config` except for location cases. Entitlements were not exposed and are
not inferred. Creation used POST; each tested variant used PUT with the CMA JSON
content type and no version header. Each variant followed a successful control PUT
and GET. Accepted variants returned 200 and matched immediate and delayed GETs
(at least 0.35 seconds in the primary run, 2 seconds in confirmation). Rejected
variants returned 422 and left the control unchanged. JSON comparisons preserved
the distinction between Boolean and numeric values. No normalization was observed.
Both disposable definitions were subsequently confirmed absent.

| Tested dimension | Accepted and preserved | Rejected with control unchanged |
| --- | --- | --- |
| Installation defaults | String, number, and Boolean defaults across Symbol, Number, Boolean, Enum, and Secret, including values that differ from the declared type. | Object and explicit JSON-null defaults for all five types; Array default for Number. |
| Enum default membership | Nonmember string, number, and empty-string defaults; both a string option-object key and its displayed label. | No isolated membership rejection. |
| Options applicability | Omitted Enum options; nonempty options on Symbol, Number, Boolean, and Secret. | Empty options lists on all five types. |
| Enum option elements | Nonempty strings, single-member objects, mixed string/object lists, and repeated values. Object values included strings, numbers, Boolean, null, empty array, and empty object. | Empty strings, Boolean/number/null/array elements, and zero- or multiple-member objects. |
| Instance declarations | Secret; eight declarations with distinct IDs. | Nine declarations. |
| Instance defaults/options | Number default `"42"`; Symbol nonempty options; Enum numeric option-object value, omitted options, and nonmember default. | Number null/object defaults; Symbol empty options; Enum multiple-member option object. |
| Array field items | Symbol and Link Entry. | Text, RichText, Integer, Number, Boolean, Date, Location, Object, and Array, each isolated without `linkType`. |
| ResourceLink Array item | No accepted form established. | `ResourceLink` with `linkType: "Contentful:Entry"`; prerequisites and entitlement effects remain unresolved. |
| Repeated locations/fields | Single-definition controls; distinct Link Entry and Link Asset definitions together. | Exact repeated simple `app-config` objects; exact repeated complete Symbol and Array-of-Symbol definitions. |

The confirmation run repeated the surprising cross-type defaults, optional/non-Enum
options, numeric option-object values, nonmember defaults, and repeated app-config
results. Its additional instance and Secret cases are retained individually in the
artifact. The case named `P4 Enum null` sent **omitted options**, not explicit null.

Earlier [nested-location observations](configuration.md#app-definition-nested-location-configuration)
cover slash-leading navigation paths, empty Link targets, nested-member applicability,
location ordering, and hosting. Their original conditions remain authoritative for
those cases; this matrix does not repeat every earlier observation.

## Why published constraints are insufficient

Sources were inspected on 2026-11-07. Source-code links pin the inspected revisions;
web guides may change after that date. These comparisons distinguish documentation
and client behavior from endpoint acceptance.

| Source | Relevant discrepancy or limitation |
| --- | --- |
| [App Parameters guide][parameters] | Describes Secret as installation-only, default/type matching, required Enum options, and Enum membership. The recorded PUT/GET counterexamples prevent treating these as universal declaration-write restrictions. |
| [Management SDK parameter types][sdk-parameters] | Defaults are a Boolean/string/number union, not tied to the declared type. Option maps require string values, narrower than the preserved CMA representations. |
| [Management SDK field/location types][sdk-definition] and [field union][sdk-field-types] | Include ResourceLink forms without establishing successful server use. Location enumerations differ from guides and earlier accepted sidebar locations; no single list establishes a closed server vocabulary. |
| [Management SDK REST adapter][sdk-adapter] | Copies and sends inputs; its types do not enforce nested semantic validation at runtime. |
| [App Scripts parameter prompts][cli-parameters] | Impose ID spelling/uniqueness and option-form conventions. They copy textual defaults; the [Number test][cli-number-test] expects `"42"`. These are client choices, not server validation guarantees. |
| [App Scripts navigation prompts][cli-navigation] | Impose lengths and a no-space path rule. Earlier CMA observations preserve slash-leading paths containing spaces, queries, or fragments. |

Installed parameter values are a separate contract from declaration persistence.
Accepting an instance Secret declaration does not demonstrate runtime secrecy or
redaction. See [Secret installation parameters](secret-installation-parameters.md)
for the separately established installed-value behavior.

## Limits on inference

These results cover the recorded organization, hosting conditions, declarations, and
PUT cases. They do not establish every POST/scope combination, installed-app behavior,
UI rendering, or cross-account capability. Broader location/field equality, parameter
ID uniqueness, label applicability, future type names, bundle prerequisites, and raw
duplicate JSON-key behavior remain unverified.

Error bodies report further bounds, including default-string length 255, option-count
1–1000, and option-string length 1–255. Only the exercised boundaries support local
checks: the matrix did not test those upper bounds, numeric ranges, Unicode length
measurement, or installation declaration limits. SDK types and error-schema fragments
can guide another controlled experiment but do not broaden the observed result.

[parameters]: https://www.contentful.com/developers/docs/extensibility/app-framework/app-parameters/
[sdk-definition]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/lib/entities/app-definition.ts#L20-L104
[sdk-field-types]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/lib/entities/field-type.ts#L1-L17
[sdk-parameters]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/lib/entities/widget-parameters.ts#L1-L18
[sdk-adapter]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/lib/adapters/REST/endpoints/app-definition.ts#L48-L73
[cli-parameters]: https://github.com/contentful/create-contentful-app/blob/96190a2babbc2af6fc687795871ab0a49aa25664/packages/contentful--app-scripts/src/create-app-definition/build-app-parameter-settings.ts#L5-L192
[cli-number-test]: https://github.com/contentful/create-contentful-app/blob/96190a2babbc2af6fc687795871ab0a49aa25664/packages/contentful--app-scripts/src/create-app-definition/build-app-parameter-settings.test.ts#L72-L117
[cli-navigation]: https://github.com/contentful/create-contentful-app/blob/96190a2babbc2af6fc687795871ab0a49aa25664/packages/contentful--app-scripts/src/location-prompts.ts#L62-L107
