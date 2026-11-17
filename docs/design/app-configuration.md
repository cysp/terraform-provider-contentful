# App and Extension configuration

Use the shared [value and identity rules](terraform-value-semantics.md) alongside
these resource-specific contracts.

## App Definition configuration

Configuration validation checks known location and parameter values. Create/Update
validate the final Plan before model decoding or mutation, including values retained
by `ignore_changes`. Read/import preserve representable responses without imposing
write restrictions. Shared UI Extension conversion does not acquire App Definition
validation rules. Location and field-type collection uniqueness is left to CMA;
the provider preserves duplicate entries in requests. Recognized field relationships
share validation with Extensions, descending only into known Array definitions.
Unrecognized field types, including ResourceLink, remain semantically opaque;
request knownness and representability checks still apply. App Definition option
cardinality follows the [shared JSON interpretation](terraform-value-semantics.md#json-normalization), without
an exception for duplicate keys. The original request JSON is not rewritten.

Validation must preserve valid replacement replanning. Terraform can first propose
an update using ignored prior state, then select replacement and replan creation from
Config. Rejecting that intermediate proposal would block a valid replacement, so
checks of effective values run before Create/Update writes. See
[Terraform replacement planning](https://github.com/hashicorp/terraform/blob/v1.17.1/internal/terraform/node_resource_abstract_instance.go#L953-L1157)
and the [replacement tests](../../internal/provider/resource_app_definition_locations_lifecycle_test.go).
An unrelated update retaining invalid state can consequently fail at apply before its
own write, after independent resources may have changed.

With the resource block still present, `terraform destroy` validates Config,
including arguments named in `ignore_changes`. Delete accepts representable
legacy state without validating it as an update or writing the definition first.
See the [destroy recovery tests](../../internal/provider/resource_app_definition_parameters_lifecycle_test.go).

## App Action category parameters

The App Action resource's `parameters` attribute represents writable Custom
parameter definitions. It is null for built-in categories; data sources expose
their complete returned definitions. Resource projection follows the response's
actual category, including after a contradictory mutation response.

## App Action identity

`app_action_id` is Optional+Computed. Create sends the known planned ID in the
POST body when configured; otherwise Contentful assigns the ID. The provider
does not restrict the ID by action type. Contentful currently accepts a
caller-supplied ID for a `function-invocation` action and rejects one for an
`endpoint` action. Update never sends an ID in its body. Removing a configured
ID retains the existing identity, and changing it requires replacement. A
Function action updated to an endpoint retains its ID, including when that ID
remains configured; a subsequent Create with that endpoint configuration would
currently be rejected by Contentful.

Create posts directly with the configured ID. Contentful may overwrite an
existing Function action with that ID; the tested conditional headers did not
prevent this. See the [App Action API evidence](../research/app-framework/actions.md).

## Extension sources

Every Contentful Extension mutation request requires exactly one of
`extension.src` or `extension.srcdoc`. HCL may omit both for an imported or
otherwise response-owned source; planning preserves the sole prior-state source
in that case. An empty `src` is invalid, while an explicitly empty `srcdoc` is
accepted and round-trips. The provider does not silently rewrite either
explicit value to omission. The first-party contract and live CMA observations
are recorded in the
[Extension configuration note](../research/ui-extensions.md).

Extension field-type and parameter-definition validation is scoped to
`extension.field_types` and `extension.parameters`; it does not validate the
resource-level installation values. Configuration validation defers unknowns.
Create and Update validate the effective Plan before decoding the native model
and sending a mutation, including definitions retained by `ignore_changes`.
Read/import preserve representable definitions without imposing request rules.
Validation is not performed during plan modification, so replacement planning
can select configuration values after an intermediate plan retaining old state.

Extension state version zero is decoded with its own historical schema and
models, including nested field types, parameter definitions, and timeouts. The
decoder does not reuse current Extension or App Definition schema builders.
Conversion preserves every value except the legacy unused empty source, which
becomes null; an explicitly empty `srcdoc` remains empty. When the current state
representation changes, update the conversion without changing the historical
decoding contract.
