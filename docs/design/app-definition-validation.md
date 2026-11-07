# App Definition configuration validation

This contract covers `contentful_app_definition` configuration validation and final
write preflight. [Terraform value semantics](terraform-value-semantics.md) owns the
shared null, unknown, ownership, and response-publication rules. The
[API evidence](../research/app-framework/schema-validation-api-evidence.md) owns
Contentful observations and source comparisons; the
[Framework evidence](app-definition-validation-framework-evidence.md) records the
local experiments that establish the lifecycle boundaries.

## Authority and public representation

Live CMA acceptance and preservation take precedence over narrower documentation,
SDK types, or CLI conventions for the tested conditions. An isolated rejection
supports a narrow check when its predicate can be identified without rejecting an
established supported representation. Accepted-but-normalized values justify a
restriction only where independently established Terraform consistency requirements
prevent faithful management. Untested forms remain uncertain, not implicitly invalid.

The public schema retains ordered nested attributes, open location/field/parameter
names, JSON-normalized default/option strings, and Optional+Computed hosting fields.
There is no schema migration or provider-private validation state. UI Extension
shares parameter representation and conversion but not App Definition semantics.
Data sources and Read/import preserve representable response values, including
historical values that would fail a write.

## Validation rules

These rules supplement existing required attributes, nonempty location/navigation
names, required field companions, and restrictions on misplaced location members.
Configuration and final Create/Update preflight use the same semantic predicates.
Paths use actual list indices `i`/`j`; `s` is `installation` or `instance`.

| Path | Rule | Evidence boundary |
| --- | --- | --- |
| `locations[i].navigation_item.path` | For a known page location, a known nonempty path starts with `/`. | Isolated page-navigation observations. No URL normalization, whitespace rule, or length limit is inferred. |
| `locations[i].field_types[j].link_type`, `.items.link_type` | A known Link target is nonempty; existing requiredness remains. | Empty Link rejection. ResourceLink and unfamiliar nonempty targets stay open. |
| `locations[i].field_types[j].items.type` | Within known entry-field and outer Array, reject Text, RichText, Integer, Number, Boolean, Date, Location, Object, and Array. | Each was isolated against an accepted Array Symbol control. ResourceLink prerequisites remain unresolved; future names stay open. |
| `locations[i]` | Reject exact repeated simple app-config objects, with null field_types and navigation_item. | Repeated isolated rejection; report the later index. No broader location-ID uniqueness is inferred. |
| `locations[i].field_types[j]` | Within entry-field, reject exact repeated complete Symbol or Array-of-Symbol definitions. | Both forms rejected against single-definition controls. Required discriminators must be known and optional companions absent; distinct Link targets remain distinct. No general type/target identity is imposed. |
| `parameters.instance` | At most eight declarations. | Eight distinct-ID controls passed; a ninth failed. Count may be known despite unknown children. No installation count limit is inferred. |
| `parameters.s[i].default` | For Symbol, Number, Boolean, Enum, and Secret, a configured default encodes a JSON string, number, or Boolean. | Cross-type primitive defaults persisted; object, array, and explicit JSON-null defaults were rejected. Terraform null means omission. Future parameter types remain service-validated. |
| `parameters.s[i].options` | For those known types, a configured options list is nonempty. | Omitted Enum options and non-Enum options persisted; empty lists failed. The reported upper bound was not tested and is not enforced locally. |
| `parameters.s[i].options[j]` | For those known types, an option is a nonempty JSON string or a single-member JSON object. | Other top-level shapes, empty strings, and zero/multiple-member objects failed. Object values may have any JSON type. Raw duplicate keys remain ambiguous and are not rejected by the new shape check. |

CMA preserved instance Secret declarations, defaults whose primitive types differ
from the declared type, nonmember Enum defaults, mixed string/object options, repeated
option values, and non-string object values. The provider accepts these forms without
coercion. Declaration acceptance does not establish secrecy or redaction of installed
parameter values.

No new hosting relation, source/bundle XOR, ID spelling/uniqueness, label applicability,
Enum membership, or typo-suggestion policy is imposed. Existing source validation and
Optional+Computed ownership remain in force. Broader uniqueness, ResourceLink support,
upper bounds, and bundle prerequisites remain service-validated.

## Lifecycle and diagnostics

Configuration validators reject independently decidable invalid values and defer only
rules whose operands are unknown. An unknown sibling does not hide a known error.
Optional null parents skip descendants. Existing schema/custom-type validation owns
requiredness, null collection elements, and JSON syntax errors.

Create and Update inspect the final Plan before native model decoding and before
POST/PUT. They reject unresolved request descendants, null collection elements, null
required parameter id/name/type values, and malformed JSON, then apply the semantic
rules. A valid Config cannot replace an invalid effective Plan. Config continues to
control only the existing hosting ownership decisions.

Effective values are not rejected during ModifyPlan: Terraform can first propose an
update using ignored prior state, then select replacement and replan creation from
Config. An unrelated update retaining invalid state can therefore plan successfully
and fail before its own write at apply. Independent resources may already have changed.
Representable irregular state remains readable and deletable with valid configuration.
An invalid configuration can block ordinary destroy even with ignore_changes; correct
Config and retry destroy without first applying a remote correction.

New errors use exact caller-rooted paths, error severity, and corrective details
without echoing defaults or option contents. Configuration syntax errors still belong
to the existing shared JSON type: its malformed-input details can echo input, and
custom-list syntax paths can be relative. This contract does not claim those separate
shared diagnostics are fixed.

## Implementation boundaries

- `resource_app_definition_locations.go` owns location semantics and existing strict
  location availability checks.
- `resource_app_definition_parameters.go` owns App Definition parameter semantics and
  final parameter availability. Its parent object validator checks known scope lists
  and declarations; Create/Update reuse it after availability checks.
- Options are `TypedList[jsontypes.Normalized]`. Traverse `Elements()` directly;
  `ToListValue` can invoke syntax validation again. A walker handling only `types.List`
  misses those children.
- The local JSON reader retains number tokens with `UseNumber`, requires complete
  consumption, and does not replace request bytes. The separate top-level member scan
  detects duplicate decoded keys before map cardinality can become an unsupported
  assumption about CMA. Shared response normalization is a different operation.
- `resource_app_definition.go` calls preflight before `Plan.Get`; shared converters,
  response projection, identity, and hosting ownership retain their existing contracts.

## Representation tradeoffs and limits

Ordered blocks are the strongest alternative for detecting unknown nested keys, but
require HCL authoring migration and lifecycle proof while still needing semantic
validation. Typed object conversion can discard misspelled keys before the provider
receives them; validators cannot report those discarded keys. A dynamic/map or raw
JSON schema could expose more keys but would lose typed guidance and require explicit
unknown/null/equivalence handling. Tagged variants and location maps also impose
identity or closed-union assumptions that the observed API does not support.

Retaining the current schema preserves existing configuration/state compatibility and
partial unknowns while catching established invalid shapes at the earliest safe
boundary. Successful validation does not certify service support for an unfamiliar
name, bundle prerequisites, every account capability, or every installed-app behavior.

## Regression boundaries

Protocol tests assert returned severity, summary, detail, and attribute paths for
known errors, with accepted live-derived controls and unknown siblings. Direct
Create/Update tests distinguish valid Config from invalid Plan, traverse custom option
lists, assert value-free rooted diagnostics and zero HTTP calls on errors, and compare
independently authored request objects for accepted representations.

Mocked Terraform tests cover import/refresh, ignored invalid values, correction,
ordinary destroy recovery, organization replacement in both orders, and tainted
replacement. Replacement tests assert planned actions, exact request payloads, mutation
order, and final absence. These establish provider behavior against fixtures, not
additional CMA acceptance. Verification commands belong in [Development](../../DEVELOPMENT.md).
