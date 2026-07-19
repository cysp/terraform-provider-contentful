# Terraform null and unknown values in provider implementations

Research date: 2026-07-19

Scope: Terraform Plugin Framework concepts and implementation guidance relevant to
this provider's pinned `terraform-plugin-framework` v1.19.0 dependency. This note
uses only HashiCorp documentation, HashiCorp-maintained source/API documentation,
and the Terraform plugin protocol documentation.

## Executive summary

Terraform values have three states: known, null, and unknown. Null means absence;
unknown means that a value exists conceptually but Terraform cannot determine it
yet. The framework's `attr.ValueState` names these states
`ValueStateKnown`, `ValueStateNull`, and `ValueStateUnknown`; provider code normally
observes them through `IsNull()` and `IsUnknown()` on framework values rather than
accessing `ValueState` directly.
([framework value API](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework@v1.19.0/attr#ValueState),
[framework data concepts](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/terraform-concepts#type-system))

The safest default is to preserve framework values (`types.String`, `types.Object`,
`types.List`, and so on) until the code has established that a value is known and,
where relevant, non-null. Go built-ins cannot represent all Terraform metadata, and
primitive accessors deliberately collapse null and unknown to zero values. For
example, `types.String.ValueString()` returns `""` for both null and unknown.
([framework data concepts](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/terraform-concepts#type-system),
[string accessors](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/string#accessing-values))

Unknown handling is phase-sensitive:

- Configuration validation must normally return early for unknown values, because
  interpolations can make any configured attribute unknown during an early
  validation pass and Terraform will validate non-computed attributes again once
  they are known.
  ([validation](https://developer.hashicorp.com/terraform/plugin/framework/validation),
  [unknown values](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/terraform-concepts#unknown-values))
- Provider `Configure` is different: required configuration can still be unknown.
  The provider must explicitly decide whether to defer useful work, warn, or return
  an error; it must never silently treat unknown as an empty credential or endpoint.
  ([provider unknown values](https://developer.hashicorp.com/terraform/plugin/framework/providers#unknown-values))
- By resource Create, Read, Update, and Delete, required attributes are known and
  non-null and optional non-computed attributes are known or null. Computed
  attributes can still be unknown in a plan, but state never contains unknown
  values.
  ([value availability guarantees](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values#when-can-a-value-be-unknown-or-null))

## 1. The value model

### 1.1 Known, null, and unknown are distinct

Null is absence, commonly caused by an omitted optional attribute. Unknown is a
deferred value, commonly caused by an expression referring to a value that is only
known after apply. Providers do not control which configuration expressions become
unknown, so handling unknown only on "likely computed" inputs is insufficient.
([null and unknown concepts](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/terraform-concepts#type-system))

The states are mutually exclusive in the framework base types:

| Terraform state | `IsNull()` | `IsUnknown()` | Safe to use value accessor |
|---|---:|---:|---:|
| known | false | false | yes |
| null | true | false | no, unless intentionally accepting the accessor's zero value |
| unknown | false | true | no |

The `attr.Value` contract defines `IsNull()` and `IsUnknown()`, and the framework
source defines the three `ValueState` constants.
([`attr.Value`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework@v1.19.0/attr#Value),
[`ValueState`](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.19.0/attr/value_state.go))

### 1.2 Framework types should be the boundary model

HashiCorp recommends framework types because ordinary Go types cannot retain
Terraform's null and unknown metadata. This is particularly important for model
fields populated by `Config.Get`, `Plan.Get`, or `State.Get`, and for nested object
fields that might be null or unknown.
([data concepts](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/terraform-concepts#type-system),
[object conversion guidance](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/object#setting-values))

Conversion to a Go primitive is appropriate only after the surrounding phase and
schema guarantee the value state, or after explicit state checks. Accessors do not
raise diagnostics: primitive accessors return zero values for null/unknown, which
can erase the distinction between omitted, deferred, and explicitly empty values.
([string accessors](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/string#accessing-values))

## 2. Schema semantics determine legal value states

For attributes, configurability has these meanings:

| Schema flags | Practitioner configuration | Provider responsibility |
|---|---|---|
| `Required` | eventually known and non-null | preserve configured value |
| `Optional` | known or null | interpret null deliberately |
| `Optional` + `Computed` | configured value, or provider-chosen value when config is null | preserve configured values; fill only the unconfigured case |
| `Computed` | cannot be configured | produce the value in provider logic |

These semantics are enforced by Terraform/framework schema validation.
([object attribute configurability](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/object#configurability))

Consequences for an audit:

1. Every call to a primitive accessor must be checked against the attribute's schema
   flags and the lifecycle phase.
2. `Optional` does not mean empty. Null should be mapped intentionally to API
   omission, a documented default, or explicit clearing.
3. `Optional` + `Computed` permits the provider to choose a value only when
   configuration is null; changing a known configured value violates data
   consistency.
   ([plan data consistency](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#terraform-data-consistency-rules))
4. Defaults belong in schema defaults when they are static planning defaults. The
   framework applies a default only when configuration is null, before it marks
   unconfigured computed values unknown.
   ([resource defaults](https://developer.hashicorp.com/terraform/plugin/framework/resources/default))

For a plain `ObjectAttribute`, configurability applies to the whole object rather
than individual members. Use nested attributes when individual child fields need
their own `Required`, `Optional`, or `Computed` behavior.
([object attribute configurability](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/object#configurability))

## 3. Availability by lifecycle phase

HashiCorp documents these safe assumptions:

| Context | Null possible? | Unknown possible? |
|---|---:|---:|
| resource CRUD required attribute | no | no |
| resource CRUD optional, non-computed attribute | yes | no |
| resource plan computed attribute | no | yes |
| resource configuration read-only computed attribute | yes | no |
| provider `Configure` required attribute | no | yes |
| Terraform state | yes | no |

The full guarantee is authoritative; outside its listed cases, provider code must
handle null and unknown itself.
([accessing values: guarantees](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values#when-can-a-value-be-unknown-or-null))

These guarantees do not make unconditional access universally safe. Shared
conversion helpers may be called from validation or plan modification, where
unknown is valid, as well as from CRUD. Audit helpers by all call sites, not only by
the field's normal CRUD use.

## 4. Primitive and aggregate access

### 4.1 Primitive accessors

Call `IsUnknown()` and `IsNull()` before `ValueString()`, `ValueBool()`,
`ValueInt64()`, and analogous accessors unless the lifecycle/schema combination
guarantees known/non-null. `String()` is only a lossy debugging representation and
must not be used as the actual value.
([string type access](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/string#accessing-values),
[`attr.Value.String`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework@v1.19.0/attr#Value))

Pointer accessors preserve null but not unknown. For example,
`ValueStringPointer()` returns `nil` for null but a pointer to `""` for unknown, so
it is not a substitute for `IsUnknown()`.
([string type access](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/string#accessing-values))

### 4.2 Lists, maps, and sets

A collection can itself be null, unknown, or known. A known collection can still
contain null or unknown elements. Therefore a check of only the container state
does not prove that its elements can safely become Go primitives.
([list access](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/list#accessing-values),
[map access](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/map#accessing-values),
[set access](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/set#accessing-values))

Recommended sequence:

1. Decide the intended behavior for a null container.
2. Return/defer/diagnose for an unknown container as appropriate to the phase.
3. Convert using `ElementsAs` into framework element types, such as
   `[]types.String` or `map[string]types.String`, when elements may carry metadata.
4. Append conversion diagnostics and stop on errors before using the result.
5. Inspect each framework element before converting to an API-native type.

HashiCorp explicitly recommends framework element types to account for unknown
elements.
([list conversion](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/list#accessing-values),
[map conversion](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/map#accessing-values))

The `allowUnhandled` boolean accepted by `ElementsAs` should normally be `false`.
In v1.19.0, `true` configures the reflection converter to replace unhandled null and
unknown values with empty Go values, deliberately discarding the distinction. Use
it only where that lossy policy is the explicit domain contract.
([v1.19.0 list implementation](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.19.0/types/basetypes/list_value.go),
[v1.19.0 map implementation](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.19.0/types/basetypes/map_value.go),
[v1.19.0 set implementation](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.19.0/types/basetypes/set_value.go))

`Elements()` returns `nil` for both a null and unknown container, so it also cannot
distinguish those states without prior checks.
([list access](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/list#accessing-values),
[map access](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/map#accessing-values))

### 4.3 Objects

Check the object itself before calling `As`. Prefer target structs whose fields are
framework types if child attributes can be null or unknown. Go primitives in a
target struct are appropriate only when every child is guaranteed known and
non-null.
([object conversion](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/object#setting-values))

When constructing values, use the state-specific constructors
(`ObjectNull`, `ObjectUnknown`, `ListNull`, `ListUnknown`, and equivalents) when
that is the intended state. Use `*Value`/`*ValueFrom` for known values, append their
diagnostics, and reserve `*ValueMust` for tests or exhaustively tested logic because
it converts diagnostics into panics.
([list constructors](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/list#setting-values),
[map constructors](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/map#setting-values),
[object constructors](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/object#setting-values))

## 5. Validation

Attribute, type, resource, data-source, and provider configuration validation must
expect unknown input. The normal custom-validator pattern is:

```go
if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
    return
}
```

Whether null returns early depends on schema and validator purpose, but unknown
normally must return without an error so interpolation remains valid. Terraform
performs basic required/type enforcement, so a content validator should not
reimplement requiredness by rejecting null unless null has a distinct invalid
meaning not already represented by the schema.
([validation guidance and example](https://developer.hashicorp.com/terraform/plugin/framework/validation),
[unknown validation behavior](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/terraform-concepts#unknown-values))

Cross-field validation must independently account for each participating value's
state. It should diagnose only relationships proven invalid from known values.
Path-aware diagnostics are preferable for attribute-specific failures.
([resource configuration validation](https://developer.hashicorp.com/terraform/plugin/framework/resources/validate-configuration),
[attribute diagnostics](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics#addattributeerror-and-addattributewarning))

Configuration validation is offline: resource/data-source `Configure` may not have
received provider data. Remote/API-backed checks belong in resource plan
modification or data-source `Read`, not configuration validation.
([resource validation](https://developer.hashicorp.com/terraform/plugin/framework/resources/validate-configuration),
[resource configure ordering](https://developer.hashicorp.com/terraform/plugin/framework/resources/configure#define-resource-configure-method))

## 6. Plan modifiers and consistency

During planning, defaults run first. If the plan differs from state, the framework
then changes null-configured computed attributes to unknown, followed by attribute
and resource plan modifiers.
([plan modification process](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#plan-modification-process))

Plan modifiers must preserve unknown configuration. The official
`UseStateForUnknown` example refuses to copy prior state when `ConfigValue` is
unknown because doing so would break interpolation. It copies state only for an
unknown plan arising from an unconfigured stable computed value.
([official modifier example](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#creating-attribute-plan-modifiers))

Use `UseStateForUnknown` only when the value is known not to change across the
operation. Copying stale state for a value the API may change produces an
inaccurate plan and can lead to an inconsistent-result error after apply.
([built-in plan modifiers](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#common-use-case-attribute-plan-modifiers),
[update recommendations](https://developer.hashicorp.com/terraform/plugin/framework/resources/update#recommendations))

Configured known or null plan values must be returned exactly in state; only
unknown planned values can be replaced with provider-observed known values.
Terraform rejects inconsistent results.
([create caveats](https://developer.hashicorp.com/terraform/plugin/framework/resources/create#caveats),
[update caveats](https://developer.hashicorp.com/terraform/plugin/framework/resources/update#caveats),
[protocol overview and lifecycle references](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol))

Plan modifiers execute for create, update, and destroy. Resource-level modifiers
must leave an entirely null destroy plan null. Modifiers nested below lists and sets
cannot assume Terraform has realigned prior-state elements after reorder/removal.
([plan operation checks and collection caveat](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#caveats))

## 7. CRUD, import, configure, and data-source practices

### Create

- Read from `req.Plan`, not `req.Config`, because defaults and plan modifiers may
  have changed the planned representation.
- Do not write unknown state.
- Preserve every known or null planned value exactly; resolve only unknown planned
  values.
- Append `Get`, conversion, and `Set` diagnostics and stop before API calls or
  dereferences when errors exist.

([create recommendations and caveats](https://developer.hashicorp.com/terraform/plugin/framework/resources/create))

### Read

- Start from prior state for identifiers and other lookup context.
- Write a complete refreshed state containing only known or null values.
- If the remote object no longer exists, call `resp.State.RemoveResource(ctx)`.
- Preserve configured representation unless a custom type's semantic equality
  explicitly defines an inconsequential normalization.

([read resources](https://developer.hashicorp.com/terraform/plugin/framework/resources/read),
[semantic equality](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/custom#semantic-equality))

### Update

- Read intended values from `req.Plan`; compare framework values with `Equal` when
  detecting changes so state metadata and Terraform collection semantics are
  retained.
- Always set response state; it must contain no unknown values.
- Preserve known/null plan values exactly and resolve only unknown plan values.

([update caveats and change detection](https://developer.hashicorp.com/terraform/plugin/framework/resources/update))

### Delete

- Read lookup data from prior state.
- Treat a remote "not found" as successful deletion.
- Do not manually preserve or set non-null response state; successful Delete is
  automatically removed from state, while an error keeps it managed.

([delete recommendations and caveats](https://developer.hashicorp.com/terraform/plugin/framework/resources/delete))

### Import

Import must set enough known state for the next Read to locate and fully refresh
the object, or return an error. When lookup requires multiple attributes, parse and
set all of them explicitly rather than relying on a single passthrough ID.
([resource import](https://developer.hashicorp.com/terraform/plugin/framework/resources/import))

### Provider and child `Configure`

After `req.Config.Get`, stop on diagnostics, then explicitly handle unknown provider
configuration before calling accessors or constructing clients. HashiCorp allows
either an error or a deferred/warning strategy depending on whether any provider
functionality remains possible, but silent zero-value conversion is not a strategy.
([provider Configure and unknown values](https://developer.hashicorp.com/terraform/plugin/framework/providers#configure-method))

Resource and data-source `Configure` must accept `ProviderData == nil`, validate
its concrete type, and lifecycle methods should guard against an unexpectedly nil
client.
([resource configure](https://developer.hashicorp.com/terraform/plugin/framework/resources/configure#define-resource-configure-method),
[data-source configure](https://developer.hashicorp.com/terraform/plugin/framework/data-sources/configure#define-data-source-configure-method))

### Data sources

Data-source `Read` receives configuration but no plan or prior state. Terraform
normally defers the call until arguments are known when dependencies require it;
the implementation should still preserve framework values through conversion,
append diagnostics, and write known/null state.
([framework data sources](https://developer.hashicorp.com/terraform/plugin/framework/data-sources#read-method),
[Terraform data-source deferral](https://developer.hashicorp.com/terraform/language/data-sources#data-source-behavior))

## 8. Diagnostics are part of correct value handling

Every `Get`, `GetAttribute`, `As`, `ElementsAs`, `*ValueFrom`, `Set`, and
`SetAttribute` diagnostic must be appended to the response. Check
`resp.Diagnostics.HasError()` before using partially converted values or continuing
to an API call. Never replace existing diagnostics.
([diagnostic append and error handling](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics#working-with-existing-diagnostics))

Use attribute-path diagnostics for value-specific errors and whole-resource
diagnostics for API/lifecycle failures. Error diagnostics do not automatically
prevent response state from being persisted, so partial-operation error paths must
return the deliberately correct state rather than assume Terraform discards it.
([diagnostic paths and state effects](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics#how-errors-affect-state))

## 9. Provider audit checklist

Search for and inspect all of the following:

- Primitive accessors: `ValueString`, `ValueBool`, `ValueInt64`,
  `ValueFloat64`, `ValueBigFloat`, pointer accessors, and `String()`.
- Aggregate access/conversion: `Elements`, `ElementsAs`, `As`, `Attributes`,
  direct indexing into object attributes, and conversions into Go
  `[]T`/`map[string]T`/struct fields.
- Lossy conversion flags: `ElementsAs(..., true)` and object `As` options that
  turn unhandled null/unknown into empty values.
- Constructors and setters: `types.*Value`, `types.*Null`, `types.*Unknown`,
  `types.*ValueFrom`, `types.*ValueMust`, `State.Set`, and
  `State.SetAttribute`.
- Diagnostics returned by every framework conversion or setter.
- Validators and plan modifiers that read a value without first handling the
  unknown case.
- Optional attributes whose null state is silently treated as a Go zero value.
- `Optional` + `Computed` attributes where provider logic overwrites known
  configuration.
- `UseStateForUnknown` on values that can actually change during update/read.
- Create/Update paths that read configuration instead of plan.
- Read/Create/Update/state-upgrade paths that can write unknown values.
- Import paths that do not seed every value required for Read.
- Provider Configure paths that turn unknown credentials/endpoints into empty
  strings.

For each finding, record:

1. Schema flags and nested element schema.
2. Execution phase and the authoritative availability guarantee for that phase.
3. Container state and possible child/element states.
4. Whether conversion diagnostics are handled before use.
5. Intended distinction among omitted, deferred, empty, and cleared.
6. Required plan/state consistency behavior.
7. A focused test that supplies null/unknown at the earliest phase where it is
   legal.

## 10. Recommended test strategy

Unit-test validators and plan modifiers directly with known, null, and unknown
framework values. For collections and objects, separately test a null container,
an unknown container, a known empty container, and known containers containing null
or unknown children. This distinction follows the framework's separate container
and element states.
([collection access semantics](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/list#accessing-values),
[validation unknown guidance](https://developer.hashicorp.com/terraform/plugin/framework/validation))

Acceptance tests should use interpolations from computed attributes to exercise
unknown configuration during planning, plus omitted optional attributes and
explicit `null` where Terraform syntax permits it. Assert that planning succeeds
without false validation errors, that apply resolves unknowns, and that the
post-apply plan is empty. Terraform's testing documentation establishes acceptance
testing as the mechanism that runs the provider through real Terraform operations.
([provider acceptance testing](https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests))

Tests for Create/Update should also assert plan/state consistency: configured known
and null values must survive unchanged, while provider code may resolve planned
unknowns. Tests for Read should cover remote disappearance and normalization/drift
behavior.
([create caveats](https://developer.hashicorp.com/terraform/plugin/framework/resources/create#caveats),
[update caveats](https://developer.hashicorp.com/terraform/plugin/framework/resources/update#caveats),
[read resources](https://developer.hashicorp.com/terraform/plugin/framework/resources/read))

## 11. Audit of this provider

### 11.1 Scope and method

The implementation audit covered all 187 non-test Go files under
`internal/provider` plus the custom value reflection package under
`internal/terraform-plugin-framework-reflection`. The mechanical inventory found:

- 468 calls that inspect or extract framework values (`IsNull`, `IsUnknown`,
  primitive accessors, `Elements`, `ElementsAs`, `As`, or `ToTerraformValue`);
- 308 `tfsdk` model fields;
- 24 model fields represented by Go pointers, slices, strings, or booleans rather
  than an `attr.Value`;
- 29 `Optional` + `Computed` schema declarations; and
- no production use of `ElementsAs(..., true)`.

Each accessor was then classified by schema flags and lifecycle phase. This is
important because the raw count is not a defect count: the documented CRUD
guarantees make direct access to required values safe, and make optional
non-computed values known-or-null.

The audit also ran the existing focused null/unknown unit tests. They passed, but
several currently assert lossy behavior described below; passing tests therefore
confirm current behavior rather than prove that behavior correct.

### 11.2 Confirmed finding: provider configuration collapses unknown values

`ContentfulProvider.Configure` checks `IsNull()` but not `IsUnknown()` before
calling `ValueString()` for both `url` and `access_token`. This loses the distinction
between an unknown value and `""`, contrary to the provider Configure guidance.

A protocol-v6 probe, run with both relevant environment variables unset, observed:

```text
unknown access token: diagnostics=1
  Failed to configure client: No access token provided
unknown URL: diagnostics=0
```

The unknown URL was silently replaced by Contentful's default endpoint. The
unknown token was rejected for the wrong reason. The temporary probe was removed
after execution.

Impact: a provider configuration that depends on another resource can silently use
the wrong endpoint during planning, or emit a misleading missing-token diagnostic
instead of identifying the unresolved dependency.

Evidence:

- `internal/provider/contentful_provider.go`, `Configure`;
- the protocol observation above; and
- HashiCorp's [provider unknown-value guidance](https://developer.hashicorp.com/terraform/plugin/framework/providers#unknown-values)
  and [official Configure tutorial](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider-configure).

### 11.3 Confirmed finding: taxonomy updates conflate unknown with empty

The taxonomy concept and concept-scheme schemas define these collections as
`Optional` + `Computed` without plan modifiers:

- concept: `alt_labels`, `hidden_labels`, `notations`,
  `broader_concept_ids`, and `related_concept_ids`;
- concept scheme: `top_concept_ids` and `concept_ids`.

On an update where some other attribute changes, Terraform marks unconfigured
computed attributes unknown. The conversion helpers `stringMap`, `stringList`, and
`stringListMap` convert both null and unknown containers into known empty Go
collections. The update code then diffs that empty request against the current
remote object and can issue a clearing patch.

This is not merely a stylistic concern. The checked-in lifecycle test uses explicit
empty maps/lists to request clearing, establishing that omission and explicit empty
are distinct user intents. Collapsing an unconfigured, unknown planned value to the
same request as explicit empty can therefore erase labels or relationships during
an unrelated update.

Evidence:

- `internal/provider/resource_taxonomy_schema.go`;
- `internal/provider/taxonomy_models.go`;
- `internal/provider/resource_taxonomy_test.go`, whose clear stage explicitly
  configures empty collections; and
- HashiCorp's [plan modification process](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#plan-modification-process).

### 11.4 Confirmed finding: known collections do not consistently validate child states

The provider generally checks whether a collection itself is null or unknown, but
several converters do not check null/unknown children before extracting them:

- team-space membership silently drops null and unknown role IDs;
- environment links diagnose an unknown ID but still append an empty-ID link, and
  convert a null ID to an empty-ID link without a diagnostic;
- content-type fields and allowed resources call `TypedObject.Value()` directly;
- editor-interface layout, controls, group controls, sidebar, and nested items call
  `TypedObject.Value()` directly;
- role policies and webhook filters call `TypedObject.Value()` directly;
- entry metadata concept/tag IDs and list-resource order/query elements use
  primitive accessors without checking element state; and
- app-installation marketplace elements silently drop null/unknown values.

A known list, set, or map can legally contain null or unknown children. For the
custom `TypedObject`, `Value()` returns the zero Go struct when the object is null
or unknown. The affected paths can therefore emit malformed empty objects, empty
identifiers, silently shortened configured collections, or inconsistent plan/state
results.

The current focused tests confirm representative behavior: unknown environment
elements become empty-ID links (with a diagnostic), null/unknown team role IDs are
silently removed, and null and unknown webhook filter containers are treated
identically. These tests should be changed to assert the intended path-aware
diagnostics and no partial request values.

This finding does not apply to conversions that use `tfsdk.ValueAs` or
`ElementsAs(..., false)`, append the diagnostics, and stop before the API call.
Those correctly reject unrepresentable child states.

Evidence:

- `internal/provider/team_space_membership_model_request.go`;
- `internal/provider/environment_links.go`;
- `internal/provider/content_type_model_request.go`;
- `internal/provider/editor_interface_model_request.go` and its nested converters;
- `internal/provider/role_model_policies_request.go`;
- `internal/provider/webhook_filters_request.go`;
- `internal/provider/entry_model_request.go`;
- `internal/provider/list_resource_entry.go`;
- `internal/provider/app_installation_model_request.go`; and
- HashiCorp's [list](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/list#accessing-values),
  [map](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/map#accessing-values),
  and [set](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/set#accessing-values)
  access guidance.

### 11.5 Confirmed finding: taxonomy Update ignores conversion errors before patching

Both taxonomy `Update` methods append diagnostics returned by `ToRequest`, but do
not check `HasError()` before building a patch and potentially calling Contentful.
A known collection containing a null child can make `ElementsAs(..., false)`
produce an error, yet the partially converted request is still used.

The Create methods have the required diagnostic barrier; the Update methods do
not. This violates HashiCorp's diagnostic handling guidance and can combine an
appropriate Terraform error with an unintended remote mutation.

Evidence:

- `internal/provider/resource_taxonomy_concept.go`, `Update`;
- `internal/provider/resource_taxonomy_concept_scheme.go`, `Update`; and
- HashiCorp's [diagnostic append and error handling guidance](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics#working-with-existing-diagnostics).

### 11.6 Reviewed patterns that are not findings

- Direct primitive access to required attributes in resource CRUD is safe under
  HashiCorp's documented lifecycle guarantees.
- Optional, non-computed primitive attributes are known or null in CRUD. Existing
  pointer conversions are safe where null deliberately means API omission or
  clearing.
- The custom typed list/map/object implementations preserve top-level null and
  unknown states through Terraform conversion.
- Production conversions consistently use `ElementsAs(..., false)`; none opt into
  the framework's lossy `allowUnhandled` behavior.
- App Key validation correctly avoids rejecting unknown values during validation
  and performs strict known-value checks before request construction.
- The content-type metadata resource-level plan reconciliation intentionally
  distinguishes omitted annotations from taxonomy preservation; its null/unknown
  behavior is covered by focused tests.
- The 24 native Go model fields are concentrated in App Definition, Extension,
  Resource Type, and Team Space Membership models. Most are read only in CRUD,
  where lifecycle guarantees make whole-attribute unknown values unavailable.
  They are a maintainability risk, not by themselves a demonstrated defect; a
  wholesale migration is not justified without a model being reused in validation
  or planning.

## 12. Improvement plan

### Phase 1: lock in failing semantics with focused tests

1. Add protocol tests for unknown provider `url` and `access_token`; assert
   path-specific unknown-value diagnostics and that no client is configured.
2. Add taxonomy lifecycle coverage that:
   - creates non-empty optional-computed collections;
   - changes only an unrelated known attribute while omitting those collections;
   - proves the collections are preserved; and
   - separately proves explicit empty collections still clear them.
3. Add table-driven converter tests for null and unknown collection elements,
   including object elements, and assert a diagnostic at the exact element path
   plus no partial API value.
4. Add taxonomy Update tests that force a request-conversion diagnostic and assert
   that no PATCH request occurs.

### Phase 2: fix provider Configure first

1. After `req.Config.Get`, check `IsUnknown()` for both provider attributes before
   reading either primitive.
2. If no internal test override supersedes the attribute, add an attribute error
   explaining that the client cannot be configured until the value is known.
3. Do not fall back to an environment variable or the default URL when an explicit
   configured value is unknown; unknown is not omission.
4. Retain the existing null behavior: null may use environment fallback, and null
   URL may ultimately use Contentful's default endpoint.

### Phase 3: make taxonomy omission and clearing distinct

1. Add list/map `UseStateForUnknown` plan modifiers, or equivalent
   resource-level plan reconciliation, to the seven optional-computed taxonomy
   collections.
2. Preserve prior refreshed state on unrelated updates when configuration is
   null; preserve an explicitly configured empty collection as empty.
3. Keep initial-create behavior explicit. If the API default is known to be empty,
   resolve unknown create plans to known empty values in plan logic or document the
   request converter's create-only defaulting precondition.
4. Stop treating unknown and null as synonyms in generic taxonomy conversion
   helpers. Give each call site an explicit policy or return framework values until
   the lifecycle-specific boundary.

### Phase 4: make aggregate conversion fail closed

1. Replace direct `TypedObject.Value()` calls at request boundaries with
   `GetValue()` plus an attribute-path error for null/unknown elements.
2. Validate every primitive collection child before calling `ValueString`,
   `ValueBool`, or another accessor. Do not append partial empty API values after an
   error.
3. For collection attributes where null children have no domain meaning, reject
   them during configuration validation as well as at request conversion.
4. Preserve deliberately distinct cases. In particular, keep entry-field
   Terraform null omission separate from a known JSON `null`, and add a test that
   documents that contract.
5. Change existing tests that currently bless silent dropping or zero-value
   conversion.

### Phase 5: enforce diagnostic barriers and safer helper contracts

1. Return immediately after taxonomy request conversion if diagnostics contain an
   error, before patch construction or network I/O.
2. Audit every `Get`, `GetAttribute`, `ElementsAs`, `ValueAs`, and state setter so
   partially converted data is never used after an error. The current audit found
   the missing pre-network barriers in the two taxonomy Update paths; retain a
   mechanical check in review to prevent recurrence.
3. Make unsafe custom accessors difficult to misuse: either remove request-side
   calls to raw `Value()`/`Elements()`, or add path-aware known-value extraction
   helpers that return diagnostics.
4. Document lifecycle preconditions on converters that intentionally rely on CRUD
   guarantees, so future reuse in validation or plan modification cannot silently
   broaden their legal input states.

### Phase 6: validate end to end

1. Run focused unit and protocol tests for every changed converter and plan
   modifier.
2. Run the complete provider unit suite.
3. Run mock acceptance tests for provider configuration, taxonomy preservation and
   clearing, content types, editor interfaces, roles, webhooks, entries, app
   installations, delivery API keys, and team-space memberships.
4. Run live acceptance only for behavior the mock cannot authoritatively establish,
   with cleanup verification. No live Contentful probe is needed for Terraform's
   value-state rules themselves; HashiCorp's framework contract is authoritative.

## 13. Implementation audit matrix

This matrix records the systematic implementation pass performed from the plan
above. "CRUD-known" refers to HashiCorp's documented guarantee that required and
optional non-computed configuration values are known (or null, when optional) by
resource Create and Update. That guarantee makes direct access at those specific
lifecycle boundaries correct; it is not a general license to use lossy accessors
in Configure, validation, planning, or aggregate-child conversion.

| Area | Value boundary reviewed | Result |
| --- | --- | --- |
| Provider configuration | `url`, `access_token`, environment fallback, default URL | Unknown configuration now produces path-specific errors before primitive access or fallback. Null retains the documented fallback behavior. Protocol tests cover both attributes. |
| App Definition | Base model, locations, field types, navigation, installation/instance parameters, JSON defaults and options | Native nested models remain CRUD-known. Framework JSON values reject unknown defaults and null/unknown option children; an unknown options collection is rejected. Optional-computed `src` and `bundle_id` intentionally remain unset on initial create and use prior state on update. |
| App Installation | Marketplace set, JSON parameters | Unknown containers and null/unknown set children now fail with attribute diagnostics. Unknown JSON parameters fail rather than being sent as omission. |
| App Key | JWK object and `x5c` children | Existing strict known-value validation is retained. Null/unknown JWK material is rejected before certificate decoding or request construction. |
| App Signing Secret | Secret request and write-only response reconciliation | Required request input is CRUD-known. Existing response logic preserves configured secret state when Contentful omits the write-only value; it does not introduce unknown state. |
| Content Type | Fields, items, validations, allowed-resource unions, metadata taxonomy unions | Required containers reject top-level null/unknown. Every typed object and required primitive child is extracted with a path-aware known-value check. Union alternatives reject missing/unknown variants. Metadata taxonomy children are handled identically. |
| Delivery API Key | Name, description, environment link list | Primitive fields are CRUD-known. Optional-computed environment links intentionally omit an unknown initial-create value; null/unknown children are rejected instead of producing empty links. Update plans preserve prior state. |
| Editor Interface | Layout, nested groups/items, controls, group controls, sidebar | Every typed object element, including recursively nested layout objects, is checked before access. Conversion diagnostics stop Create/Update before network I/O. |
| Entry | Arbitrary JSON fields, metadata concepts and tags, plan/state field reconciliation | Unknown field children and null/unknown metadata ID children are rejected. Terraform null still means omit a field, while a known normalized JSON `null` is sent as JSON null. Required field maps and CRUD identity values rely on lifecycle guarantees. |
| Environment | Name and source environment | Required/optional primitives are CRUD-known. The optional source ID retains null-as-omission semantics. |
| Environment Alias | Target environment link | Required target and identity values are CRUD-known; conversion does not run in a phase where they can be unknown. |
| Extension | Extension model, field types, parameters, App Definition parameter reuse | Native nested models are CRUD-known. Shared parameter conversion receives the App Definition JSON default/options checks. Optional-computed extension parameters intentionally omit unknown on initial create and preserve state on update. |
| Personal Access Token | Name, scopes, expiry | Required scopes now reject a null/unknown container before `ValueAs`; primitive children are decoded with `allowUnhandled=false`. |
| Resource Provider | Function link and request identity | Native/required model values are CRUD-known. No lossy aggregate access occurs. |
| Resource Type | Default field mapping, image, badge, field mappings | Native nested model values are CRUD-known. No framework aggregate child is accessed without the model decode diagnostic barrier. |
| Role | Permissions map/action lists, policy objects, action union, constraints | Top-level and nested containers reject null/unknown; policy objects and primitive children use path-aware extraction. Invalid response unions return null plus an error, never unknown state. |
| Space Enablements | Four optional-computed booleans | Unknown initial-create values intentionally leave API fields unset; all four use state-for-unknown on update. Known false remains distinct from null/unknown. |
| Tag | Identity, name, visibility | All request values are required or optional non-computed CRUD values. Null visibility is API omission; no aggregate children exist. |
| Taxonomy Concept | URI, preferred/alternative/hidden labels, notes, notations, relationship IDs | Seven optional-computed collections now preserve state on update and use an explicit empty initial-create default. Required maps and optional nullable maps/URI reject unknown. Null nullable maps remain explicit API null. Update now stops on conversion errors before patching. |
| Taxonomy Concept Scheme | URI, preferred label, definition, top/concept IDs | Uses the same explicit taxonomy policies and diagnostic barrier as concepts. Explicit empty collections still clear; omitted update configuration preserves state. |
| Team | Name and nullable description | Required name is CRUD-known. Unknown description now errors rather than becoming API null; explicit null still clears/omits according to the existing request contract. |
| Team Space Membership | Admin and role ID list | Null/unknown role children now fail instead of being silently dropped. Native required values remain CRUD-known. |
| Webhook | Required scalars, topics, credentials, filters, headers, transformation | Required and optional request primitives are checked where they cross custom conversion boundaries. Unknown collections, typed objects, union alternatives, and primitive children fail with paths. Null filters retain explicit API-null semantics. Header/transformation requests are not partially set after errors. |
| Data sources | App Definition, environment-ready status, marketplace App Definition, Preview API Key | Inputs are required data-source configuration and are known before Read under the framework contract. Response conversion diagnostics are checked before state is written. |
| List resources | Content Type and Entry | Required IDs use known-value extraction. Optional Entry filters distinguish null from unknown, validate collection children, and stop before API calls on diagnostics. |
| Import | Multipart identity parsing and state writes | Parsed identity components reject null/unknown before use. Import writes only known values to state. |
| Custom collection/object types | Terraform conversion and request-side accessors | Top-level null/unknown preservation is retained. Request converters now use `GetValue` through path-aware helpers for object children and keep `ElementsAs`/`ValueAs` in strict (`allowUnhandled=false`) mode. |
| Response/state conversion | All model-from-response paths and state setters | No API response converter intentionally creates unknown state. Conversion diagnostics are treated as barriers before `State.Set`; invalid role unions return null plus an error. Planning-only unknown values remain confined to plan reconciliation and plan modifiers. |

Representative tests cover provider protocol configuration, taxonomy lifecycle
preservation versus explicit clearing, known collection containers with
null/unknown primitive and object children, entry JSON-null semantics, and
response-conversion failures. The mechanical mutation-site audit confirmed that
every request converter returning diagnostics is followed by `HasError()` before
the Contentful client call.
