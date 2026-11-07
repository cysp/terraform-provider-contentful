# App Definition validation: Framework and lifecycle evidence

This evidence supports the [App Definition validation contract](app-definition-validation.md).
Initial protocol, schema, typo, uniqueness, and replacement experiments used
`8db2a7bc70666b6c1b1b7011c2e688373eb8ed2b`. Destroy and JSON traversal experiments
used an isolated archive of
[`739a1be8a4a36ae2a8eb69b58ea2df3cc93ae1df`](https://github.com/cysp/terraform-provider-contentful/commit/739a1be8a4a36ae2a8eb69b58ea2df3cc93ae1df).
Both pin Framework v1.19.0, Framework Validators v0.19.0, Plugin Go v0.31.0,
Plugin Testing v1.16.0, and Framework JSON Types v0.2.0. CLI experiments used
Terraform v1.17.1 on darwin_arm64; the main archive used Go 1.27.3.

All experiments in this document used local protocol calls or loopback HTTP fixtures
with synthetic credentials. They establish provider/Core behavior, not live CMA
contracts. [CMA observations](../research/app-framework/schema-validation-api-evidence.md)
are recorded separately. Baseline observations below retain their original pins and
must not be read as the current validator's results.

## Framework boundaries

The [attribute-validator API](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.19.0/schema/validator/list.go#L24-L44)
provides Config, ConfigValue, and Path. The
[resource configuration API](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.19.0/resource/validate_config.go#L21-L39)
provides configuration, which can contain unknowns. The
[server implementation](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.19.0/internal/fwserver/server_validateresourceconfig.go#L53-L116)
runs resource config validators and ValidateConfig before schema validation: a resource
hook cannot assume attribute validators have established requiredness or JSON validity.

The [ModifyPlan request](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.19.0/resource/modify_plan.go#L25-L61)
contains Config, State, and Plan but no flag proving that the proposal is the final
replacement creation plan. Framework
[plan processing](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.19.0/internal/fwserver/server_planresourcechange.go#L155-L305)
applies defaults, marks omitted computed values unknown where appropriate, and then
runs attribute/resource modifiers. The replacement experiment below establishes why
semantic rejection of effective values belongs at the final-write boundary.

## Optional+Computed ownership

The pinned [UseStateForUnknown implementation](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.19.0/resource/schema/stringplanmodifier/use_state_for_unknown.go#L12-L58)
copies prior state, including null, for an existing resource with unknown Plan and
nonunknown Config. App Definition src and bundle_id use this modifier. Omitted Config
therefore cannot establish that the effective outgoing request omits a hosting value.
The ownership contract and its null/unknown cases belong in
[Terraform value semantics](terraform-value-semantics.md#request-conversion).

## Protocol validation baseline

A scratch test called the real provider's `ValidateResourceConfig` protocol
method. It built the raw value using `AppDefinitionResourceSchema` and
`tftypes.ValueFromJSONWithOpts`, then used `Plan.SetAttribute` to insert explicit
unknowns where needed. No provider configuration or HTTP client was required.
The 21 cases all met the following explicit expectations:

| Configuration case | Error count | Observed diagnostic/boundary |
| --- | ---: | --- |
| Symbol default JSON string | 0 | Valid control. |
| Symbol default JSON Boolean | 0 | No parameter type/default semantic check. |
| Symbol default JSON object | 0 | No parameter type/default semantic check. |
| Symbol default malformed JSON `{` | 1 | `parameters.installation[0].default`: `Invalid JSON String Value`. |
| Boolean default JSON string | 0 | No parameter type/default semantic check. |
| Instance Secret | 0 | No parameter scope check. |
| Installation Secret | 0 | Accepted control; this run does not establish CMA support. |
| Enum default `"dark"`, options `["light"]` | 0 | No membership check. |
| Enum Boolean option | 0 | No option shape check. |
| Enum empty options | 0 | No semantic minimum check. |
| Duplicate installation parameter IDs | 0 | No identifier uniqueness check. |
| Identical `app-config` locations | 0 | No location uniqueness check. |
| Page navigation path `relative` | 0 | Nonempty check only. |
| Array item type Integer | 0 | Baseline provider left item type support to CMA. |
| Array without items | 1 | `locations[0].field_types[0].items`: missing items. |
| Same missing-items shape, field discriminator unknown | 0 | Correctly deferred. |
| Parameter discriminator unknown | 0 | Deferred/unchecked. |
| Parameter default unknown | 0 | Deferred/unchecked. |
| Unknown field discriminator plus known dialog `field_types=[]` sibling | 1 | Error at `locations[1].field_types`; sibling still checked. |
| Locations configured, both hosting fields omitted | 0 | Configuration cannot establish effective O+C source. |
| Known src, no locations | 0 | No hosting relation validator. |

These are baseline observations, not the current validation rule set or evidence of
CMA validity. The current [contract](app-definition-validation.md) and retained protocol
tests distinguish supported representations from rejected shapes.

## Replacement regression experiment

At the initial checkout, tests imported irregular locations: `app-config` plus a `dialog` object
containing `fieldTypes: []`. Configuration supplies only valid `app-config`,
with `ignore_changes = [locations]`. The
[local lifecycle fixture](../../internal/provider/resource_app_definition_locations_lifecycle_test.go)
checks unchanged refresh, unrelated-update rejection before PUT, correction,
destroy, organization replacement in both orders, and tainted replacement.

A Go overlay added only this method to the existing resource, leaving all test
expectations unchanged:

```go
func (r *appDefinitionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
    if req.Plan.Raw.IsNull() {
        return
    }
    resp.Diagnostics.Append(validateAppDefinitionPlannedLocations(ctx, req.Plan)...)
}
```

Both organization-replacement tests failed at step 2's non-refresh plan with
`field_types cannot be configured for location "dialog". Omit this attribute.`
Tainted replacement still passed. Removing the overlay restored **all three**
replacement subtests to PASS. This demonstrates the failure of effective-plan
rejection during replacement discovery; it does not imply every replacement
route uses the same sequence.

Terraform v1.17.1 source explains the result: it
[validates original configuration before applying ignore_changes](https://github.com/hashicorp/terraform/blob/v1.17.1/internal/terraform/node_resource_abstract_instance.go#L883-L917),
[returns on diagnostics from the first proposed plan](https://github.com/hashicorp/terraform/blob/v1.17.1/internal/terraform/node_resource_abstract_instance.go#L953-L977),
then, only after replacement is selected,
[plans again with null prior state and original configuration](https://github.com/hashicorp/terraform/blob/v1.17.1/internal/terraform/node_resource_abstract_instance.go#L1104-L1157).

Consequently, keep read/import tolerant and validate effective values before
Create/Update writes. An unrelated update retaining irregular values may plan
successfully and fail at apply before its App Definition mutation; independent
resources may already have changed. This limitation is explicit in the
[resource contract](app-definition-validation.md#lifecycle-and-diagnostics).

## Destroy compatibility of new hard configuration rules

A scratch overlay on the main archive attached a hypothetical instance-Secret
prohibition to the App Definition parameters object. This was a lifecycle probe,
**not the current API policy**: subsequent live experiments accepted and preserved
instance Secret declarations. Its exact diagnostic used error severity, path
`parameters.instance[0].type`, summary `Invalid App Definition parameter configuration`,
and detail `Secret parameters are supported only in parameters.installation.`

The ten-case protocol table checked valid controls, null/unknown parents, an unknown
discriminator, and a known error beside an unrelated unknown. The CLI experiment
imported a representable instance-Secret declaration from a loopback server and changed
only Config before ordinary destroy. The server permitted GET/DELETE at the imported
address; POST/PUT were failures and every request omitted X-Contentful-Version.

| Config under the hypothetical rule | Result | DELETE | POST/PUT |
| --- | --- | ---: | ---: |
| Valid Symbol; imported state still Secret | Destroy succeeded | 1 | 0 |
| Invalid Secret | Destroy stopped with the matching error | 0 | 0 |
| Invalid Secret plus ignore_changes on parameters | Destroy stopped with the matching error | 0 | 0 |
| Correct only Config to Symbol after either failure | Destroy succeeded without an intervening update | 1 | 0 |

Removing only the overlay rule caused the three expected protocol-error cases to fail
on zero diagnostics and both invalid-config CLI cases to fail because destroy succeeded.
Restoring it restored their passes. The seven valid/null/unknown protocol controls
remained passing. These results isolate configuration validation's effect on destroy;
they do not justify the hypothetical Secret restriction.

Terraform v1.17.1 validates
[configuration before ignore_changes](https://github.com/hashicorp/terraform/blob/v1.17.1/internal/terraform/node_resource_abstract_instance.go#L875-L905).
Its [destroy-plan implementation](https://github.com/hashicorp/terraform/blob/v1.17.1/internal/terraform/context_plan.go)
includes a pre-destroy refresh/plan route. Configuration validators receive no destroy-
intent flag. The experiment establishes ordinary destroy recovery after correcting
Config; it does not establish refresh=false, targeting, block removal, or provider
downgrades as bypasses.

The retained `TestAccAppDefinitionResourceParameterDestroyRecovery` uses the empirically
rejected empty-options shape instead of that hypothetical rule. It checks valid-config
deletion of irregular imported state, invalid-config rejection, ignore_changes, and
configuration-only recovery. It also checks that an unrelated update retaining invalid
options stops before PUT.

## Nested keys can disappear before provider validation

Real `terraform validate -json` with the unmodified local provider produced:

| Input spelling | Result |
| --- | --- |
| `app-config` plus `fieldTypes = [{ type = "Symbol" }]` | Valid; zero errors. |
| Symbol field plus `linkType = "Entry"` | Valid; zero errors. |
| `app-config` plus `navigationItem = {...}` | Valid; zero errors. |
| Symbol parameter plus `defaut = true` | Valid; zero errors. |
| Top-level `locatoins = []` alongside valid `locations` | Rejected: unsupported argument. |
| Link field plus `linkType = "Entry"` | Rejected because `link_type` is absent, not because the typo is recognized. |

A local `plan -refresh=false` also succeeded for the `fieldTypes` case.
`terraform show -json plan` reported the planned location as:

```json
[{"field_types":null,"location":"app-config","navigation_item":null}]
```

This matches [Terraform's documented object conversion](https://developer.hashicorp.com/terraform/language/expressions/type-constraints#conversion-of-complex-types):
extra object attributes are discarded. The typed provider cannot diagnose keys
it never receives. Validation can detect resulting missing required companions,
but it cannot promise to catch arbitrary camelCase or misspelled nested keys.
Retain the typed schema; a different dynamic/map public representation would be
a larger compatibility change, not a small validation fix.

## What the alternatives actually change

A further scratch schema test replaced only the current `locations`
`ListNestedAttribute` with one ordered `ListNestedBlock`, retaining the same
nested attributes and validators. The complete App Definition Terraform root
types compared equal. Empty locations and a two-element page/Array-Link value
also encoded under the attribute schema and decoded under the block schema
without changing the raw value. The pinned
[block type](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.19.0/resource/schema/list_nested_block.go#L210-L220)
and [attribute type](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.19.0/resource/schema/list_nested_attribute.go#L245-L255)
are both lists of the nested object type.

This falsifies an automatic state-type-migration claim; it is not a complete
block migration experiment. One ordered block family can retain the `location`
discriminator, open names, and mixed list order. Existing configurations would
still need syntax changes, expression-produced lists would need dynamic blocks,
and explicit `locations=[]`/required-attribute authoring would need reconsideration.
Keeping inner `field_types` as nested attributes would also retain typo loss
inside those inner objects. The experiment does not establish equivalent
planning/null behavior after a full conversion of all nested members to blocks.

A separate real CLI check used the current schema with these two installation
parameters (all other optional members omitted):

```hcl
{ id = "boolean", name = "Boolean", type = "Boolean", default = true }
{ id = "number", name = "Number", type = "Number", default = 42 }
```

Both `validate -json` and `plan -refresh=false` succeeded. The plan's defaults
were the strings `"true"` and `"42"`: Terraform had already coerced the primitive
values to the existing JSON-string attribute type. Thus native Boolean/numeric
syntax alone is not a new capability of a dynamic default attribute. Native
string/object/list authoring is a distinct potential benefit; recommending
`jsonencode` consistently remains useful. This matches
[Core primitive conversion](https://developer.hashicorp.com/terraform/language/expressions/type-constraints#conversion-of-primitive-types).

## Nested uniqueness and unknown values

Pinned [listvalidator.UniqueValues](https://github.com/hashicorp/terraform-plugin-framework-validators/blob/v0.19.0/listvalidator/unique_values.go#L29-L65)
skips elements only when the element itself is unknown, then uses `Equal`.
A scratch experiment constructed known objects with unknown `type` children:

| List values | Stock validator errors |
| --- | ---: |
| `[{type=unknown}, {type=unknown}]` | 1: `Duplicate List Value` at the list root. |
| `[{type="Symbol"}, {type="Link"}]` | 0. |
| `[{type="Symbol"}, {type="Symbol"}]` | 1. |

Thus whole-object equality with unknown descendants is not sufficient evidence for
an early duplicate error. The current narrow checks require the complete tested shape
to be known. This observation does not establish a parameter-ID uniqueness policy.

## JSON traversal and parsing boundaries

These focused checks ran on the pinned main archive. They establish the custom-list
traversal and parsing behavior independently of the App Definition semantic rules.

| Input/operation | Observed result | Required design consequence |
| --- | --- | --- |
| `TypedList[jsontypes.Normalized]` containing malformed `{`, then `ToListValue` | One `Invalid JSON String Value` error at relative path `[0]`. | This adapter performs validation; do not use it as an inert semantic iterator and append its diagnostics again. |
| Same typed list via `Elements()` or `ToTerraformValue` | Original malformed value preserved; raw conversion succeeds. | Traverse raw values, then give configuration syntax and write-preflight syntax their separate owners. |
| Protocol configuration with malformed `parameters.installation[0].options[0]` | One existing syntax error with protocol path `ElementKeyInt(0)`, not the complete Terraform path. | The existing shared custom-list path limitation must not be claimed fixed by the new parent validator. |
| `json.Decoder.UseNumber()` and one Decode of `1 2` | First Decode succeeds with number token `1`; second Decode also succeeds. | One Decode does not establish complete JSON consumption. Require syntax validity/EOF before a semantic decision. |
| Number token `9007199254740993` | `json.Number` retains that exact token; second Decode returns EOF. | Avoid float64 and retain original request bytes. |
| `{"light":1,"light":"Light"}` decoded into `map[string]any` | Syntactically valid; only the last string label survives. | A map decode is not lossless evidence about all supplied members or duplicate-key semantics. |

Pinned main's
[`TypedList` implementation](https://github.com/cysp/terraform-provider-contentful/blob/739a1be8a4a36ae2a8eb69b58ea2df3cc93ae1df/internal/provider/typed_list.go)
explicitly separates raw encoding from `ToListValue`'s validating reflection
adapter. The existing
[`jsontypes.Normalized` syntax diagnostic](https://github.com/hashicorp/terraform-plugin-framework-jsontypes/blob/v0.2.0/jsontypes/normalized_value.go#L119-L137)
also includes the malformed value in its detail. Keep **new semantic** diagnostics
value-free and caller-rooted, but do not claim all existing syntax diagnostics
already satisfy those properties. Changing shared syntax diagnostics/path handling
requires an explicit scoped fix and independent App Definition/Extension regression
coverage; avoid incidental changes to the shared typed collection.

## Retained regression evidence

The executable tests retain the boundary assertions independently of the historical
scratch harnesses:

- [Parameter protocol/preflight tests](../../internal/provider/resource_app_definition_parameters_internal_test.go)
  assert actual protocol paths and corrective details, preserve accepted live-derived
  shapes, and distinguish Terraform null from JSON null. Create/Update tests use valid
  Config beside invalid Plan and assert rooted errors without echoed values and zero
  HTTP calls for unknown/null/malformed custom option-list values. Number parsing tests
  preserve large, decimal, exponent, and negative-zero tokens and require complete JSON.
- [Location tests](../../internal/provider/resource_app_definition_locations_internal_test.go)
  cover narrow duplicate shapes, distinct valid shapes, future names, and independent
  known duplicates beside unknowns.
- [Replacement tests](../../internal/provider/resource_app_definition_locations_lifecycle_test.go)
  exercise invalid imported locations and parameters separately, retained by ignore_changes
  with valid Config. They check no-op planning, both organization replacement orders,
  tainted replacement, exact payloads, request order, and final cleanup.
- [Destroy tests](../../internal/provider/resource_app_definition_parameters_lifecycle_test.go)
  use the real CLI to distinguish invalid Config from irregular State and assert zero
  DELETE/POST/PUT after rejection and deletion after configuration-only correction.

These fixtures do not establish additional CMA validity. Run the focused tests and
applicable checks using [Development](../../DEVELOPMENT.md#validation-scope); record
revision-specific results and fault-injection outcomes with the change being reviewed.
