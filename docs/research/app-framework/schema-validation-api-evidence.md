# App Definition validation evidence

This appendix records AppDefinition validation evidence from published documentation,
first-party clients, and authorized CMA experiments. **The live results below supersede
the earlier documentation-led P1/P3–P6 proposal where they disagree.** Observed CMA
acceptance and rejection now govern write validation; documentation and
client behavior remain separately identified evidence. The [configuration reference](configuration.md)
retains earlier probes without changing their original scope. Shared [evidence and
redaction conventions](../README.md#evidence-and-redaction) apply.

## Sources and evidence boundaries

Published pages were inspected on 2026-11-07. The App Parameters guide was also
successfully rechecked during refinement. A fresh retrieval of the CMA App Definitions
overview failed (the web tool could not open it; direct HTTPS returned 429), so its
contract statements below retain the earlier inspection's evidence boundary. Source
comparisons and the serialization experiment use these immutable revisions; they are
not claims about whichever dependency versions a later CLI installation resolves:

- [contentful-management.js][cma-commit]:
  `dfe576c9f320da5bf4732b6e8891f250c59d68b7`.
- [create-contentful-app][cca-commit]:
  `96190a2babbc2af6fc687795871ab0a49aa25664`.
- [App SDK (`ui-extensions-sdk`)][app-sdk-commit]:
  `224182ee69b16d901d0509e6e7daa6e947a38bd3`.

| Evidence | What it establishes |
| --- | --- |
| Published CMA and App Framework pages | The stated public contract; several enumerations and examples are incomplete or inconsistent with other first-party sources. |
| Management SDK types | Representable TypeScript shapes. They do not establish server validation, clearing, normalization, or completeness. |
| SDK REST adapter | Create copies and POSTs the input. Update copies it, removes `sys`, and sends `X-Contentful-Version`. Neither operation validates nested locations or parameters. [Adapter][sdk-adapter], [raw transport][sdk-raw]. |
| App Scripts prompts and transformations | Constraints and defaults chosen by the CLI. The local serialization experiment below follows their output through the SDK to a stubbed Axios adapter. This establishes emitted JSON, not CMA acceptance. |
| Upstream tests | Inspected source assertions, not tests executed during this study. App Scripts tests mock prompts or CMA; the SDK includes an integration test creating a Secret declaration, but no current run result was inspected. [CLI tests][cli-parameter-tests], [creation tests][cli-create-tests], [SDK integration test][sdk-integration]. |
| Earlier CMA probes | Exact tested cases recorded in [nested location configuration](configuration.md#app-definition-nested-location-configuration) and [installation parameter replacement](configuration.md#installation-parameter-replacement). |
| Authorized live CMA matrix | The [controlled PUT and GET results below](#authorized-cma-validation-experiments) establish preserved or rejected representations for one organization and the exact tested conditions. They supersede conflicting documentation-led predicates, not every untested service behavior. |

The generated [create][cma-create] and [update][cma-update] operation pages expose a
generic string-keyed request map. That absence of a detailed generated schema is not
evidence that arbitrary JSON is accepted. The separate [CMA overview][cma-definition]
and [App Parameters guide][parameters] carry the substantive constraints.

## Authorized CMA validation experiments

The [sanitized request/result artifact](schema-validation-live-results.json) retains
both completed runs: **96 cases, 52 preserved 200 responses, and 44 rejected 422
responses**. It removes identifier values and `sys`, replaces synthetic names with
stable placeholders, and preserves JSON value types, array order, controls, responses,
GETs, and error details. Its bodies are evidence rather than replay-ready requests;
the declaration IDs used by the eight/nine-entry controls were distinct.

The primary matrix began at **2026-11-07 13:25:09 UTC** (2026-11-08 in
Australia/Sydney). It used `api.contentful.com`, the example project's existing CMA
credential, and organization fingerprint **`sha256:609f4317b96b`** (a truncated hash,
not an organization identifier). Organization metadata did not expose entitlements;
none are inferred. The user explicitly authorized disposable-resource experiments.
One newly created AppDefinition was **never installed**, used synthetic declaration
values, `src=https://example.invalid/app`, no bundle, and app-config locations except
for the location cases. No existing installed definition was changed.

`POST /organizations/{organization_id}/app_definitions` returned **201**. All 58
matrix variants were subsequent **PUTs** to that definition, using
`Content-Type: application/vnd.contentful.management.v1+json` and no version header.
Before each variant, a separate valid-control PUT and GET both had to return 200 and
preserve the full expected control body before the harness continued. Each variant
had an immediate GET. Successful writes had a second GET after at least 0.35 seconds;
all 24 successful variants preserved the request, mutation response, immediate GET,
and later GET. All 34 rejected variants returned **422** and left the preceding
control unchanged, including its metadata in the harness's comparison. An independent
canonical-JSON comparison of the retained request/control and GET bodies confirmed
the outcomes without treating Boolean false as numeric zero. No case was normalized.

The cleanup DELETE returned before the immediate verification GET timed out. Its
numeric response status was not retained; a separate subsequent GET returned **404**,
confirming absence. Do not upgrade this record into an observed DELETE 204.

A confirmation run began at **2026-11-07 13:28:22 UTC**, using a second newly created,
never-installed definition under the same organization, host, credential, and hosting
conditions. Creation returned **201**. Its 38 PUT cases comprised **28 preserved 200**
and **10 rejected 422** results, each preceded by a successful valid control. Successful
cases were read immediately and again after at least **2 seconds**; all remained
unchanged. Rejections retained their controls. Cleanup returned **DELETE 204**, then
**GET 404**. The independent JSON-type-strict comparison covered all 96 cases in both
runs and found no classification discrepancy.

The following table is the retained synthetic observation matrix. `installation[0]`
means a declaration with `id:"setting"`, `name:"Setting"`, and the named `type`;
unmentioned optional members were omitted. The literal JSON distinctions matter.

| Primary run: changed request dimension | Observed status and persisted result |
| --- | --- |
| Number default `"42"` or `false`, from control default `42` | **200**, each value and JSON type preserved. |
| Boolean default `"false"` or `0`, from control default `false` | **200**, each value and JSON type preserved. |
| Symbol default `42` or `false`, from control default `"text"` | **200**, each value and JSON type preserved. |
| Number default `{}`, `[]`, or null; Boolean/Symbol default `{}` or null | **422**, unchanged control. Errors report a Boolean/string/number union at `parameters.installation[0].default`, independent of the declared type in these cases. |
| Enum with options omitted, from control `options:["light"]` and no default | **200**, options omitted in both GETs. The harness case named `P4 Enum null` sent omission, **not** an explicit JSON-null options value. |
| Enum `options:["light"]` | **200**, preserved. |
| `options:[]` on Enum, Symbol, Number, Boolean, or Secret | **422**, unchanged control; `size` error at the options list reports minimum 1 and maximum 1000. The maximum was not exercised by this matrix. |
| `options:["light"]` on Symbol, Number, Boolean, or Secret | **200**, preserved on every tested non-Enum type. |
| Enum options `[{"light":"Light"}]`, `["light",{"dark":"Dark"}]`, or `["light","light"]` | **200**, preserved, including mixed forms, ordering, and duplicates. |
| Enum options `[{}]` or `[{"light":"Light","dark":"Dark"}]` | **422**, unchanged control; errors at the element report an object with exactly one property as an alternative to string. |
| Enum options `[""]` | **422**, unchanged control; string alternative reports minimum 1 and maximum 255. The maximum was not exercised by this matrix. |
| Enum options `[false]`, `[42]`, `[null]`, or `[[1]]` | **422**, unchanged control; neither supported top-level option alternative matched. |
| Enum options `[{"light":1}]` | **200**, numeric map value preserved. A string-valued-label restriction would reject an observed usable representation. |
| Enum options `["light"]`, default `"dark"`, `42`, or `""` | **200**, each nonmember/other-primitive/empty default preserved. |
| Enum options `[{"light":"Light"}]`, default `"Light"` or `"light"` | **200**, both displayed label and key preserved. Membership was not enforced in either case. |
| Enum options `["light"]`, default null | **422**, unchanged control; default primitive-union error. |
| Enum options `[]`, default `"light"` | **422** for options-list size, **not** evidence for a default-membership rule. |
| Array field item Symbol or Link Entry | **200**, preserved, from an Array Symbol control. |
| Array field item Text, RichText, Integer, Number, Boolean, Date, Location, Object, or Array | **422**, unchanged Array Symbol control. Each invalid item was isolated without `linkType`, removing the earlier compound-input ambiguity. |
| Array field item ResourceLink with `linkType:"Contentful:Entry"` | **422**, unchanged control. This establishes the tested form's failure, not a target/entitlement conclusion. |
| Two simple `{location:"app-config"}` objects | **422**, `unique` at `locations`; one app-config control preserved. |
| Exact duplicate Symbol field definitions; exact duplicate Array-of-Symbol definitions | **422**, `unique` at `locations[0].fieldTypes`; respective single-definition controls preserved. |
| Link Entry and Link Asset in the same fieldTypes list | **200**, both preserved as distinct definitions. |

The confirmation run repeated the six cross-type Number/Boolean/Symbol primitive
default cases, omitted Enum options, Number options, numeric option-map value, two
Enum nonmembership cases, and repeated simple app-config. It reproduced every result.
It then isolated the following additional cases:

| Confirmation request dimension | Observed status and persisted result |
| --- | --- |
| Enum single-key option object with value `false`, null, `[]`, `{}`, or `""` | **200**, each JSON value preserved; together with numeric and string controls this rules out the proposed string-label restriction. It does not promise every future object size or nested value is accepted. |
| Secret or Enum default `42`, `false`, or `"text"` | **200**, preserved for both types. The Enum cases omitted options. |
| Secret or Enum default null or `{}` | **422**, unchanged respective default-omitted controls. |
| Instance declaration type changed from Symbol to Secret | **200**, Secret persisted in both GETs. This falsifies the proposed installation-only rejection; no installation or runtime secrecy behavior was tested. |
| Nine instance declarations versus eight distinct-ID controls | **422**, unchanged eight-entry control; a size diagnostic reports the maximum of eight. |
| Instance Number default `"42"` | **200**, string preserved. |
| Instance Number default null or `{}` | **422**, unchanged control. |
| Instance Symbol options `["light"]` versus `[]` | **200** with nonempty options preserved; **422** for the empty list. |
| Instance Enum options `[{"light":1}]` versus `[{"light":"Light","dark":"Dark"}]` | **200** with the single-property numeric value preserved; **422** for the two-property object. |
| Instance Enum with options omitted; with `["light"]` and default `"dark"` | **200**, omission and nonmember default respectively preserved. |

The rejection bodies expose useful expected-schema fragments, including default
string length 255, options count 1–1000, nonempty string options up to 255, and
one-property option objects. They corroborate the exercised failures but are not
boundary tests of every reported limit. Arbitrary numeric ranges, Unicode length
measurement, every operation/scope combination, and raw duplicate JSON keys require
their own evidence before additional restrictions are inferred.
The observations concern declaration persistence, not generated form rendering or
validation of installed parameter values.

### Effect on the earlier proposal

The earlier documentation-led policy is **superseded**, rather than left as an
alternative implementation mode:

- Do not reject Secret in the instance scope. The confirmation run accepted and
  preserved that declaration. Documentation about Secret installation-value protection
  does not establish instance secrecy or justify rejecting the observed declaration.
- Do not require defaults to match the declaration's type. The cross-type primitive
  acceptances are preserved service representations, not coercion. The tested object,
  array, and explicit-null failures instead support a broad primitive-shape check for
  known types, corroborated across both scopes; Terraform null/omission remains a
  separate representation. Unknown future type names stay outside new semantic policy.
- Do not require options merely because type is Enum, or forbid them on other known
  types. Omission and non-Enum options were preserved. Present-empty options are a
  different, rejected case.
- Do not require option object values to be strings, homogeneous option forms, or
  unique option values. Numeric values, mixed forms, and duplicates were preserved.
  The narrow tested failures are unsupported top-level option shapes, empty string
  options, and zero/multiple-property objects. Duplicate raw object keys were not
  exercised; do not decide cardinality after silently collapsing those occurrences.
- Do not enforce Enum default membership or key-versus-label matching. The isolated
  contrary cases succeeded unchanged. Keep supported authoring examples as guidance,
  not a substitute for the observed endpoint behavior.
- The instance count boundary remains supported: eight passed, nine failed. The new
  isolated Array-item rejections support the tested contextual rules; exact app-config
  repetition remains rejected. The field duplicate examples support checks for their
  exact complete forms, not broader type/target uniqueness.

The [validation contract][proposal] owns the resulting implementation predicates and their
unknown-value, lifecycle, and compatibility rules. These live observations supply
their service evidence; they do not authorize silent normalization or broader claims
about untested accounts, future API versions, installed apps, or UI behavior.

## Locations, field types, and hosting

| Concern | Source comparison and limit of the evidence |
| --- | --- |
| Location vocabulary | The CMA overview lists six locations; the App Locations guide lists seven including `home`; the SDK adds `experience-toolbar` and internal-only `agent`. The recorded CMA probes additionally accepted `experience-sidebar` and `component-sidebar`. None of those lists alone is a complete public server enum. [CMA overview][cma-definition], [guide][locations], [SDK][sdk-definition], [observations](configuration.md#app-definition-nested-location-configuration). |
| `entry-field.fieldTypes` | The overview requires at least one. The SDK models an array, and the CLI prompts require a nonempty selection. The overview's location-table shorthand shows an object, while its complete example shows the array; the array agrees with the SDK and recorded CMA behavior. [Overview][cma-definition], [SDK][sdk-definition], [CLI][cli-locations]. |
| Field-type alternatives | The published list and CLI offer scalar `Symbol`, `Text`, `RichText`, `Integer`, `Number`, `Boolean`, `Date`, `Location`, `Object`; Link to Entry/Asset; and Array of Symbol or Link to Entry/Asset. The SDK additionally declares ResourceLink and Array-of-ResourceLink. The [live matrix](#authorized-cma-validation-experiments) now isolates the recognized invalid Array item names, with Symbol/Link controls; no accepted AppDefinition ResourceLink form is established here. [Overview][cma-definition], [CLI][cli-locations], [SDK union][sdk-field-types], [earlier observations](configuration.md#app-definition-nested-location-configuration). |
| Nested member applicability | SDK types associate `fieldTypes` with `entry-field` and `navigationItem` with `page`. Earlier probes distinguish rejected misplaced members from accepted members that the server drops. A claim that every extra member causes CMA rejection would contradict those observations. [SDK][sdk-definition], [observations](configuration.md#app-definition-nested-location-configuration). |
| Navigation | SDK `NavigationItem` has string `name` and `path`. CLI prompts impose name length 1–40, path length at most 512, a leading slash, and no ASCII space. Earlier CMA probes accepted a space-only name and slash-leading paths containing a space suffix, query, or fragment. The no-space CLI rule is therefore stricter than recorded CMA behavior. The 40/512 limits have only CLI evidence in the reviewed sources. [SDK][sdk-definition], [CLI][cli-navigation], [observations](configuration.md#app-definition-nested-location-configuration). |
| Uniqueness and ordering | Earlier duplicate field definitions and repeated `app-config` locations were rejected; valid reordered distinct locations were preserved. The [live matrix](#authorized-cma-validation-experiments) adds exact repeated app-config, Symbol, and Array-of-Symbol request evidence, and preserves distinct Link Entry/Asset definitions. This does not make uniqueness type-only or establish every future-form equality rule. [SDK union][sdk-field-types], [earlier observations](configuration.md#app-definition-nested-location-configuration). |
| Hosting | The overview allows `src` plus a bundle containing Functions only. A blanket mutual exclusion is incompatible with that contract. `src` is documented as HTTPS except localhost. The earlier probes establish the tested source/location dependencies, including name-only creation and removing the source with locations. Bundle-content validation was not exercised. [Overview][cma-definition], [observations](configuration.md#app-definition-nested-location-configuration). |
| CLI creation workflow | App Scripts prepends `dialog`, supplies `http://localhost:3000` when any other location exists, and otherwise leaves `src` undefined. Its transformation adds `fieldTypes` only for `entry-field` and optional navigation only for `page`. Mocked tests assert source omission for dialog-only input. These choices corroborate the distinction between dialog and rendered frontend locations without proving all server dependencies. [Settings][cli-settings], [request builder][cli-create], [transformation][cli-transform], [tests][cli-create-tests]. |

The SDK create type excludes `bundle`, while the update entity includes it. Its runtime
POST adapter does not remove a supplied bundle. This type distinction is not proof
that CMA rejects bundle selection on creation. A public contract or authorized
experiment would be needed to establish that boundary. [Types][sdk-definition],
[adapter][sdk-adapter].

The independently implemented [App SDK location constants][app-sdk-locations] also
include `entry-field-sidebar` and `asset-sidebar`. Its [runtime factory][app-sdk-api]
selects API features by location, falls back for unknown locations, and exposes
parameters supplied in the connection message. Its [parameter interface][app-sdk-parameters]
uses generic maps. This is evidence about consuming installed apps, not validating
AppDefinition writes; it does not establish an exhaustive server enum. The controlled
matrix separately records option-map and default-membership behavior.

## Parameter declarations

The [App Parameters guide][parameters] documents required `id`, `name`, and `type`;
letter/digit/underscore IDs; Boolean, Symbol, Number, Enum, plus installation-only
Secret; type-matching defaults; and Enum defaults drawn from options. Options are
Enum-only, represented by strings or value-to-label maps. Labels describe Enum's
`empty` and Boolean's `true`/`false`. It states a maximum of eight instance declarations
and a default of `false` for `required`.

The guide's default-type, options-applicability, and Enum-membership statements conflict
with the controlled preserved values above. The following comparison retains each
source's actual claim; it no longer treats the guide as the provider's rejection policy.

| Candidate constraint | Evidence and unresolved behavior |
| --- | --- |
| Declaration type/scope | The guide calls Secret installation-only, but the confirmation run accepted and preserved instance Secret. Keep that documented/runtime disagreement explicit; declaration persistence alone says nothing about secrecy or rendering in the instance scope. [Guide][parameters], [live matrix](#authorized-cma-validation-experiments). |
| Nonempty names | Required `name` establishes presence, not a documented string-length or trimming rule. The CLI rejects falsy input but accepts whitespace. No CMA boundary for empty or whitespace-only parameter names was established. [Guide][parameters], [CLI][cli-parameters]. |
| Letter-first ID | The CLI uses `^[a-zA-Z][a-zA-Z0-9_]*$`; the published character restriction does not separately specify a leading-letter rule. SDK IDs are plain strings. Leading digits/underscores, length, reserved IDs, and non-ASCII interpretations have no CMA acceptance/rejection evidence in this study. [CLI][cli-parameters], [SDK][sdk-parameters]. |
| Unique IDs | The CLI checks uniqueness independently within instance and installation arrays. It does not check across those arrays. Its mocked tests exercise duplicate recovery. Server uniqueness and case sensitivity remain unverified. [CLI][cli-parameter-uniqueness], [tests][cli-parameter-tests]. |
| Default type | SDK `default` is a broad Boolean/string/number union, not a discriminated union tied to `type`. The CLI copies textual defaults unchanged; its mocked Number test expects `"42"`, and the local capture confirms that wire type. Separately, [live writes](#authorized-cma-validation-experiments) preserved cross-type primitives and rejected the tested nonprimitive/null defaults. The guide's declared-type match is therefore not an observed CMA enforcement rule. [SDK][sdk-parameters], [CLI][cli-parameters], [test][cli-number-test]. |
| Missing versus empty default | The CLI emits explicit `""` for blank non-Secret input and omits default for Secret. SDK controls preserve `false`, `0`, and null while omitting undefined. The live matrix separately preserved an empty Enum default outside its option values. Omission, empty string, and JSON null must retain their distinct wire representations. [CLI][cli-parameters], [serialization experiment](#cli-to-sdk-serialization-experiment), [live matrix](#authorized-cma-validation-experiments). |
| Explicit null default | The SDK excludes null; the guide does not specify clearing. Live PUTs rejected explicit null for all five known installation types and for instance Number, preserving their prior controls. That establishes those update cases, not every type/scope/operation combination. [SDK][sdk-parameters], [guide][parameters], [live matrix](#authorized-cma-validation-experiments). |
| Enum option form and membership | The guide requires Enum options and membership; the SDK permits strings or string-valued maps; the CLI imposes homogeneous forms but bypasses membership for blank defaults. Live writes instead preserved omitted Enum options, non-Enum options, mixed forms, duplicate values, numeric map values, and nonmember defaults. They rejected empty lists, empty option strings, and zero/multiple-property maps. Distinguish those exact endpoint observations from all three clients' assumptions. [Guide][parameters], [SDK][sdk-parameters], [CLI][cli-parameters], [live matrix](#authorized-cma-validation-experiments). |
| Labels | The guide describes rendering keys, without an explicit prohibition on other type/key combinations. The SDK allows all three optional string keys on every parameter type. The CLI asks for labels only for Boolean/Enum and emits empty label strings when inputs are blank. Neither establishes that CMA rejects misplaced or empty labels rather than preserving or normalizing them. Document the supported UI keys without inferring a hard applicability validator or a routine warning. [Guide][parameters], [SDK][sdk-parameters], [CLI][cli-parameters]. |
| Secret | The guide, SDK integration fixture, and older service observations cover installation Secret. The CLI skips its default prompt, while the SDK type does not prohibit defaults. The new matrix preserved Secret nonempty options and string/number/Boolean defaults, and rejected null/object defaults. Declaration metadata is not protected installed-value storage. Labels and runtime consequences remain separate questions. [Guide][parameters], [CLI][cli-parameters], [SDK][sdk-parameters], [integration test][sdk-integration], [scoped Secret evidence][secret-evidence], [live matrix](#authorized-cma-validation-experiments). |
| Limits | Eight instance declarations is published and now confirmed by eight/nine controls. Live error details report option-count and string-length bounds, but a reported bound is distinct from an exercised boundary. No installation-declaration count or declaration-document byte limit is established. The 16/32 kB disagreement for installation **values** is not a limit on AppDefinition declaration arrays. [Guide][parameters], [live matrix](#authorized-cma-validation-experiments), [installation size evidence](configuration.md#appdefinition-and-appinstallation-data). |

Secret protection concerns installed parameter **values**, not automatic confidentiality
of all AppDefinition declaration metadata. The guide describes redacted personal-token
and App SDK reads, with raw values available to App Identities, Events, and Functions.
The existing [Secret discussion](configuration.md#appdefinition-and-appinstallation-data)
retains those access and recovery limitations. [Guide][parameters].

The separately pinned [Secret installation evidence][secret-evidence] records accepted
optional Secret declarations, a change to `required:true`, and visibility changes after
switching a declaration to Symbol or removing it. It also records installation values
remaining free-form when the declaration array is absent or `[]`. These observations
support preserving an empty installation-declaration list and distinguish declarations
from installed values. They do **not** establish defaults/options acceptance: that study
explicitly excludes Secret defaults and complete declaration validation. Its rejection
of a numeric installed Symbol value cannot be transplanted into a claim about the
AppDefinition default endpoint. Declaration changes can also expose existing installed
plaintext. App Definition validation adds no installation reads or automatic declaration
rewrites. The pinned source delimits those independent installation observations.

### CLI-to-SDK serialization experiment

The source trace closes the gap between a mocked prompt assertion and the outgoing
JSON body. At the pinned revisions, the following path contains no parameter-default
normalization:

1. The default prompt has no `filter`. Inquirer returns the submitted string (blank
   becomes `""`), and its default filter is identity. App Scripts' `validateDefault`
   returns validation status, not a replacement value. The Number branch checks
   `isNumber(Number(input))`; this even accepts a nonnumeric string because Lodash's
   `isNumber` includes NaN. [Input prompt][inquirer-input], [filter/validation][inquirer-base],
   [CLI validation][cli-parameters].
2. `buildAppParameterSettings` copies `parameter.default`; only the labels undergo
   `JSON.parse`. `buildAppDefinitionSettings` assigns those arrays to `parameters`,
   and the interactive entry point passes the settings directly to creation.
   [Construction][cli-parameters], [settings][cli-settings], [entry point][cli-entry].
3. `createAppDefinition` places the same instance/installation arrays into `body` and
   calls `client.appDefinition.create`. SDK `createClient` chooses the plain client;
   its wrapper forwards `payload` unchanged to the adapter and endpoint dispatcher.
   [CLI create][cli-create], [SDK client][sdk-client], [plain binding][sdk-plain],
   [wrapper][sdk-wrap], [dispatcher][sdk-dispatch].
4. The AppDefinition REST endpoint deep-copies the payload and calls `raw.post`.
   `RestAdapter` and SDK Core construct an Axios instance without a custom request-body
   transform. Axios's normal object transform uses JSON serialization, then invokes
   its transport adapter. A nested string remains a string; undefined object members
   disappear; null remains present. [Endpoint][sdk-adapter], [raw HTTP][sdk-raw],
   [REST setup][sdk-rest-setup], [Core HTTP setup][core-http], [Core options][core-options],
   [Axios serialization][axios-defaults], [adapter boundary][axios-dispatch].

On 2026-11-07, a local Node 24.18.0 harness transpiled the pinned sources with
TypeScript 5.6.3 and executed that path. Inquirer 8.2.7's input-value handling and the
actual prompt `when`/`filter`/`validate` functions were exercised with supplied answers;
terminal interaction, credential lookup, organization selection, and credential caching
were stubbed. SDK `createClient`, the plain client/wrapper, REST adapter, AppDefinition
and Organization endpoints, and raw transport ran from the pinned source. The endpoint
registry contained only those two exercised endpoints; unrelated SDK exports were stubbed.
SDK Core 10.0.0, Axios 1.18.0, and fast-copy 4.1.1 matched the SDK lockfile; Lodash 4.18.1
and Inquirer 8.2.7 matched the CLI package. The CLI declares a management SDK range,
so this is an explicit composition of the two pinned revisions, not a reproduction of
an arbitrary published CLI dependency installation. [SDK lockfile][sdk-lock],
[CLI dependencies][cli-package].

The capture replaced only Axios's transport adapter, which receives `config.data`
**after** the real request transform. Synthetic GET/POST responses let the creation
workflow complete. HTTP, HTTPS, socket creation, and fetch were disabled; the run made
zero network attempts. Literal expected values were checked against parsed captured
bodies. All names, IDs, and tokens were synthetic. The 15 captures established:

| Supplied input | Captured declaration member |
| --- | --- |
| CLI Number text `42`, `0`, or `not-a-number` | Respectively `"default":"42"`, `"default":"0"`, `"default":"not-a-number"`. |
| CLI Boolean text `false` | `"default":"false"`. Blank label input also emitted `"labels":{"true":"","false":""}`. |
| Blank Number, Boolean, or Symbol input | `"default":""`; blank did not omit the member. |
| CLI Enum text `light`, options `light,dark` | `"default":"light","options":["light","dark"]`. |
| Blank Enum input, with `light,dark` or blank options | `"default":""`, with respectively `["light","dark"]` or `[]` options; blank bypassed membership validation. |
| Secret prompt | No `default` member. |
| Direct SDK controls: native `0`, native `false`, null, undefined | Respectively JSON `0`, `false`, `null`, and omitted `default`. These controls bypassed CLI prompts. |

For reproduction, fetch the cited revisions, transpile the exercised modules without
changing their bodies, substitute prompt answers and peripheral CLI I/O, and inject an
Axios adapter returning synthetic organization/create responses. Assert that its
`config.data` is a string before parsing it; a mock of `client.appDefinition.create`
or `http.post` alone stops too early to prove serialization. Use the dependency versions
above and the table's literal expectations. This is client behavior evidence only:
there was no CMA HTTP request, server validation, stored resource, or successful live
CLI creation. This local experiment alone did not falsify the documented type-match or
Enum-membership enforcement hypothesis. The later [live matrix](#authorized-cma-validation-experiments)
independently did so for the tested cases, without coercion.

The provider's JSON-string representation follows a different path. The existing
[Terraform experiment][framework-native] shows Core converting native `default = 42`
to attribute contents `42`. The provider then places those contents in `jx.Raw`, so
the payload is JSON number `42`; `default = "42"` has the same raw JSON contents.
`default = jsonencode("42")` instead supplies contents `"42"` and therefore a JSON
string. CLI text `42` is a JavaScript string property serialized normally, so it matches
the latter wire type. Likewise, native HCL `false`/`jsonencode(false)` produce a JSON
Boolean while `jsonencode("false")` and CLI text `false` produce JSON strings. This
distinction follows the observed Core conversion and the [provider raw conversion][provider-parameter-request];
the SDK experiment did not execute Terraform or the provider.

### Historical policy hypothesis

Before live writes were authorized, the proposal treated the guide's Secret scope,
default-type, Enum applicability, string labels, and membership statements as hard configuration
policy, while acknowledging possible preserved-value counterexamples. The authorized
matrix produced those counterexamples. The [live consequences above](#effect-on-the-earlier-proposal)
replace that policy; the source trace and client captures remain valid historical
evidence. No warning or compatibility toggle should retain the superseded checks.

### Read-only CMA observation

On 2026-11-07, an authorized organization-scoped
`GET /organizations/{organization_id}/app_definitions?limit=100` returned 200. Sanitized
inspection found installation declarations with a Boolean JSON default and a
string-valued labels object; Enum declarations with string-list options and no default;
and Symbol declarations with `required` present as a Boolean or optional members absent.
No values, tenant identifiers, or raw responses are retained here.

This establishes those persisted shapes in one returned page. It does not establish
creation/update acceptance, coercion or normalization, rejection cases, completeness of
the collection, or availability for another organization. That earlier read-only phase
performed no CMA writes; the subsequent authorized matrix is recorded separately above.
Earlier mutation observations remain scoped to their original experiments in
[configuration](configuration.md).

## Remaining contract questions

Live evidence now resolves several earlier documentation/client disagreements, but it
is not a complete executable server schema. Cross-account behavior, future type names,
all operation/scope combinations, raw duplicate-key semantics, label persistence,
precise numeric and Unicode limits, and untested declaration limits remain separate
questions. Preserve accepted representations without coercion and tie every new hard
predicate to its observed failure, with a valid control. First-party type definitions
and error-schema fragments may guide a targeted experiment; neither silently broadens
its result. UI form behavior and installed-value validation remain distinct from
AppDefinition declaration persistence.

[cma-commit]: https://github.com/contentful/contentful-management.js/commit/dfe576c9f320da5bf4732b6e8891f250c59d68b7
[cca-commit]: https://github.com/contentful/create-contentful-app/commit/96190a2babbc2af6fc687795871ab0a49aa25664
[app-sdk-commit]: https://github.com/contentful/ui-extensions-sdk/commit/224182ee69b16d901d0509e6e7daa6e947a38bd3
[app-sdk-locations]: https://github.com/contentful/ui-extensions-sdk/blob/224182ee69b16d901d0509e6e7daa6e947a38bd3/lib/locations.ts#L3-L15
[app-sdk-api]: https://github.com/contentful/ui-extensions-sdk/blob/224182ee69b16d901d0509e6e7daa6e947a38bd3/lib/api.ts#L40-L90
[app-sdk-parameters]: https://github.com/contentful/ui-extensions-sdk/blob/224182ee69b16d901d0509e6e7daa6e947a38bd3/lib/types/api.types.ts#L85-L93
[cma-definition]: https://www.contentful.com/developers/docs/references/content-management-api/app-definitions/
[cma-create]: https://www.contentful.com/developers/docs/references/content-management-api/app-definitions/create-a-new-app-definition/
[cma-update]: https://www.contentful.com/developers/docs/references/content-management-api/app-definitions/update-an-app-definition/
[locations]: https://www.contentful.com/developers/docs/extensibility/app-framework/locations/
[parameters]: https://www.contentful.com/developers/docs/extensibility/app-framework/app-parameters/
[sdk-definition]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/lib/entities/app-definition.ts#L20-L104
[sdk-field-types]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/lib/entities/field-type.ts#L1-L17
[sdk-parameters]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/lib/entities/widget-parameters.ts#L1-L18
[sdk-adapter]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/lib/adapters/REST/endpoints/app-definition.ts#L48-L73
[sdk-raw]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/lib/adapters/REST/endpoints/raw.ts#L33-L59
[sdk-integration]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/test/integration/app-definition-integration.test.ts#L40-L73
[cli-locations]: https://github.com/contentful/create-contentful-app/blob/96190a2babbc2af6fc687795871ab0a49aa25664/packages/contentful--app-scripts/src/location-prompts.ts#L3-L60
[cli-navigation]: https://github.com/contentful/create-contentful-app/blob/96190a2babbc2af6fc687795871ab0a49aa25664/packages/contentful--app-scripts/src/location-prompts.ts#L62-L107
[cli-settings]: https://github.com/contentful/create-contentful-app/blob/96190a2babbc2af6fc687795871ab0a49aa25664/packages/contentful--app-scripts/src/create-app-definition/build-app-definition-settings.ts#L18-L48
[cli-create]: https://github.com/contentful/create-contentful-app/blob/96190a2babbc2af6fc687795871ab0a49aa25664/packages/contentful--app-scripts/src/create-app-definition/create-app-definition.ts#L90-L108
[cli-create-tests]: https://github.com/contentful/create-contentful-app/blob/96190a2babbc2af6fc687795871ab0a49aa25664/packages/contentful--app-scripts/src/create-app-definition/create-app-definition.test.ts#L121-L143
[cli-transform]: https://github.com/contentful/create-contentful-app/blob/96190a2babbc2af6fc687795871ab0a49aa25664/packages/contentful--app-scripts/src/create-type-safe-locations.ts#L3-L31
[cli-parameters]: https://github.com/contentful/create-contentful-app/blob/96190a2babbc2af6fc687795871ab0a49aa25664/packages/contentful--app-scripts/src/create-app-definition/build-app-parameter-settings.ts#L5-L169
[cli-parameter-uniqueness]: https://github.com/contentful/create-contentful-app/blob/96190a2babbc2af6fc687795871ab0a49aa25664/packages/contentful--app-scripts/src/create-app-definition/build-app-parameter-settings.ts#L184-L192
[cli-parameter-tests]: https://github.com/contentful/create-contentful-app/blob/96190a2babbc2af6fc687795871ab0a49aa25664/packages/contentful--app-scripts/src/create-app-definition/build-app-parameter-settings.test.ts#L119-L204
[cli-number-test]: https://github.com/contentful/create-contentful-app/blob/96190a2babbc2af6fc687795871ab0a49aa25664/packages/contentful--app-scripts/src/create-app-definition/build-app-parameter-settings.test.ts#L72-L117
[cli-entry]: https://github.com/contentful/create-contentful-app/blob/96190a2babbc2af6fc687795871ab0a49aa25664/packages/contentful--app-scripts/src/create-app-definition/index.ts#L1-L10
[cli-package]: https://github.com/contentful/create-contentful-app/blob/96190a2babbc2af6fc687795871ab0a49aa25664/packages/contentful--app-scripts/package.json
[sdk-client]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/lib/index.ts#L85-L117
[sdk-plain]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/lib/plain/plain-client.ts#L499-L507
[sdk-wrap]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/lib/plain/wrappers/wrap.ts#L57-L84
[sdk-dispatch]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/lib/adapters/REST/make-request.ts#L10-L39
[sdk-rest-setup]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/lib/adapters/REST/rest-adapter.ts#L37-L72
[sdk-lock]: https://github.com/contentful/contentful-management.js/blob/dfe576c9f320da5bf4732b6e8891f250c59d68b7/package-lock.json
[core-http]: https://github.com/contentful/contentful-sdk-core/blob/9e3a5bee7346e70ecd2720222e0f8ba4143ddac5/src/create-http-client.ts#L25-L75
[core-options]: https://github.com/contentful/contentful-sdk-core/blob/9e3a5bee7346e70ecd2720222e0f8ba4143ddac5/src/create-default-options.ts#L77-L100
[axios-defaults]: https://github.com/axios/axios/blob/2d06f96e8602c2db13b65a26340ee4a1bbc0b61f/lib/defaults/index.js#L23-L107
[axios-dispatch]: https://github.com/axios/axios/blob/2d06f96e8602c2db13b65a26340ee4a1bbc0b61f/lib/core/dispatchRequest.js#L34-L48
[inquirer-input]: https://github.com/SBoudrias/Inquirer.js/blob/9c5259a889d37205829c2255171ef74c6336f7f3/packages/inquirer/lib/prompts/input.js#L74-L79
[inquirer-base]: https://github.com/SBoudrias/Inquirer.js/blob/9c5259a889d37205829c2255171ef74c6336f7f3/packages/inquirer/lib/prompts/base.js#L95-L119
[framework-native]: ../../design/app-definition-validation-framework-evidence.md#what-the-alternatives-actually-change
[provider-parameter-request]: https://github.com/cysp/terraform-provider-contentful/blob/8db2a7bc70666b6c1b1b7011c2e688373eb8ed2b/internal/provider/app_definition_parameter_request.go#L57-L80
[proposal]: ../../design/app-definition-validation.md
[secret-evidence]: https://github.com/cysp/terraform-provider-contentful/blob/739a1be8a4a36ae2a8eb69b58ea2df3cc93ae1df/docs/research/app-framework/secret-installation-parameters.md#required-declarations-and-definition-changes
