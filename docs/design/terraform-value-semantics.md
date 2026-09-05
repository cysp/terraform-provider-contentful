# Terraform value semantics

Status: current provider design. This note defines the provider's request,
response, and state-publication boundaries. The implementation currently uses
`terraform-plugin-framework` v1.19.0.

## Value states

Terraform values have three states: known, null, and unknown. A known empty
string, list, set, or map is known and non-null; it is not omission and must not
be rewritten as null or unknown.

| Value state | Boundary meaning |
| --- | --- |
| Known, non-null | May be converted to an API value or written to state. |
| Null | Absence. The schema and API contract decide whether this means omission, clearing, or an explicit null. |
| Unknown | Deferred until a later Terraform phase. It must not be read through a primitive accessor. |

These rules also apply to collection elements and object attributes. Conversion
must inspect both the container and each child before reading primitive values.
Diagnostics for invalid children must retain their exact attribute paths.

## Request conversion

Configuration validators generally defer null or unknown values that Terraform
may resolve during planning. Lifecycle request conversion is the strict boundary
before remote mutation. It must:

- preserve framework values until their state is known;
- apply an explicit null policy: omit, clear, or send an intentional null;
- accept known empty values when the schema permits them;
- reject any unresolved value needed by the remote request;
- reject invalid null children with path-specific errors; and
- require exactly one known, non-null alternative for a closed request union.

Resource Create and Update configuration is known when the lifecycle method
runs. For Optional+Computed attributes, request conversion uses configuration
only to distinguish ownership/presence; payloads are always constructed from
the plan:

- Config null + Plan unknown: omission can be valid because the attribute is
  response-owned.
- Config null + Plan known: send the known planned value, including a default or
  a value preserved from prior state.
- Config known + Plan known: send the planned value.
- Config known + Plan unknown: fail closed. Never substitute Config for Plan.

The configuration-owned Optional+Computed case is the production-reachable
ownership hazard addressed by this rule. Required and Optional-only request
values are also rejected when unknown as defensive converter hardening, even
though Framework lifecycle guarantees normally prevent those values from
reaching Create or Update unresolved.

All conversion diagnostics are appended to the lifecycle response. If any are
errors, the lifecycle returns before invoking the Contentful client with a
mutation payload containing unresolved values. Harmless required-known identity
or version parameter objects may already have been constructed; a partially
converted payload is never sent.

### Taxonomy Optional+Computed collection ownership

For both taxonomy concepts and concept schemes, configuration determines
durable ownership of the affected Optional+Computed collections. During an
actual mutation, omitted collections remain unknown unless a known plan or
default explicitly constrains them. Known plans are never discarded: they are
sent and constrain the returned state. Clean no-change plans remain stable;
response values fill unknown plans.

The ownership rules are:

- Omitted collections are response-owned, so state may reflect Contentful.
- Any known configured value, including an explicit empty collection, is
  configuration-owned and must be preserved after Create or Update.
- Explicit empty collections are accepted input and remain distinct from
  omission.

For a concept scheme request, omitted `top_concept_ids` and `concept_ids` are
left absent on the wire. Explicit empty collections are sent as `[]`. The
response schema still requires both collections; current CMA observations show
them as present empty arrays after an omitted request.

The affected concept collections are `alt_labels`, `hidden_labels`,
`notations`, `broader_concept_ids`, and `related_concept_ids`. The affected
scheme collections are `top_concept_ids` and `concept_ids`.

## Response projection

Response conversion projects remote data into the Terraform schema; it does not
reapply request validation. A representable but semantically irregular response
should preserve as much structure as the schema safely allows, including parent
collections, element positions, recognized alternatives, and sibling values.
Lossy projection emits a path-specific warning. The projected value may be
invalid for a later request, in which case strict request conversion rejects it.

Response conversion must never introduce unknown values into state. The
following remain errors because they cannot be safely represented:

- invalid JSON or failed JSON normalization;
- a framework model/schema mismatch;
- failure to encode state or identity; and
- a missing identity where resource identity publication is required.

The generated OpenAPI client remains a stricter, earlier boundary. A payload
that cannot be decoded through its closed enums or discriminated unions may fail
before provider response conversion runs. Provider lenience applies only after
the generated client has produced a typed value that the Terraform schema can
represent, wholly or partially.

For concept `alt_labels` and `hidden_labels`, Contentful may currently add
known-empty entries for locales present in `pref_label`. After Create or Update,
the provider preserves the configured representation only for that narrow
response-added canonicalization. On refresh it preserves the prior state
representation for the same canonicalization; a first Read with no prior value,
including import, takes the complete response. It does not treat arbitrary
missing and empty locales as equivalent: nonpreferred empty or nonempty values,
list ordering, and remote omission of a previously present empty value remain
meaningful. Read and import therefore expose meaningful remote labels and drift.

Omission remains valid and delegates current CMA canonicalization. When the
provider explicitly sends `altLabels` or `hiddenLabels`, it supplies an array
for each `pref_label` locale because CMA rejects explicit maps missing a
preferred-locale entry.

For the Optional localized maps `note`, `change_note`, `definition`,
`editorial_note`, `example`, `history_note`, and `scope_note` on a concept, and
`definition` on a concept scheme, CMA currently canonicalizes an explicitly
sent `{}` to `null`. This is a narrow exception to the usual distinction
between null and known empty values: after a mutation, the provider publishes
the planned `{}` only after verifying a returned `null`; on Read, it preserves
prior `{}` only for the same returned-null case. Import or prior null with a
returned null remains null; otherwise Read and import publish the remote map,
including nonempty drift. This is an observed CMA behavior, not a documented
CMA guarantee, so its focused mock and lifecycle tests are upgrade guards.

The repository mock models only the directly observed effective `en-US`
taxonomy locale behavior for `prefLabel`, `altLabels`, `hiddenLabels`, and all
observed localized concept and scheme fields:
unsupported locales are discarded when `en-US` is present, and requests without
an effective `en-US` preferred label, or with an explicitly supplied label map
without an effective `en-US` entry, are rejected. It does
not model secondary taxonomy locale enablement or provider-side locale
validation.

### Webhook Basic password

Contentful accepts `httpBasicPassword` on Webhook Create and Update but omits
the property from successful mutation and GET responses. The generated client
distinguishes absence (`Set` is false) from explicit JSON `null` (`Set` and
`Null` are true). Only absence receives the password fallback; explicit null
and returned values remain response truth.

| Operation | Config | Effective Plan | Prior State | CMA password response | Request password | Terraform password state | Diagnostic |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Create | configured | known | none | absent | Plan value | exact Plan value | none |
| Create or Update | configured | unknown | any | not reached | no request | unchanged | attribute error |
| Read/refresh | any | n/a | known managed value | absent | none | exact prior value | none; equality is not established |
| Read/import | omitted or configured | n/a | imported null | absent | none | null | none |
| Read | any | n/a | any | explicit null or returned value | none | response value | none |
| Ordinary Update or rotation | configured | known | any | absent | effective Plan value | exact Plan value | none |
| Update with `ignore_changes` | changed | prior value retained | known | absent | effective Plan value | exact effective Plan value | none |
| Credential removal | explicit null for both Basic credentials | null for both | known | absent | null for both | null | none |
| Update after omission | omitted | null | known or null | absent | null | null | none |
| Post-import ownership | configured | known | imported null | absent | effective Plan value | exact Plan value | intentional Update |
| Mutation contradiction | configured | known | any | explicit null or different returned value | effective Plan value | response value | Terraform consistency error |

Mutation state starts from the complete CMA response. Only an absent password
property restores a known, non-null effective Plan value. Read similarly
preserves only a known, non-null prior state value; import has no such value and
therefore leaves the password null. The fallback never copies Config or another
planned attribute and never introduces an unknown value into state.

Contentful does not expose the stored password. Preserving prior Terraform
state therefore prevents false drift but is not evidence that the remote
password is equal. Refresh cannot detect another actor rotating or removing
the password. Import likewise cannot discover or claim ownership of it; later
password configuration is an intentional update.

## Lifecycle ownership and plan consistency

Content Type publication metadata is observation, not activation authority.
`contentful_content_type` activates only the exact draft returned by a
successful Create or modeled Update. After the complete truthful draft state
and optimistic-lock version are checkpointed and the response identity,
positive version, draft tuple, and plan consistency are validated, private
state records that exact version as pending activation authority. Read, import,
legacy state, external deactivation, external drafts, matching configuration,
and field equality never create that authority.

A modeled Update uses the exact current Contentful version and the effective
Terraform plan. This includes refreshed values protected by `ignore_changes`.
After Contentful accepts the draft, the provider checkpoints its returned state
and version before recording the pending marker and activating that exact
version. Confirmed activation checkpoints the returned state/version and clears
the marker. An explicit or ambiguous activation failure retains the truthful
draft and marker; an unchanged later operation is deliberately planned as an
Update and activates only the marked version without another draft PUT.

Read preserves the marker only while the current `sys.version` exactly equals
the marker and the observed publication tuple exactly equals the checkpointed
draft tuple. Observing the marked
version already activated clears the marker without another activation.
Observing any other current version or publication state revokes authority and
does not mutate. With refresh disabled, recovery still submits only the marked
version; `VersionMismatch` revokes the marker and never causes a GET-and-retry
against a newer version.

During Create, once the provider has checkpointed and marked the exact returned
draft, an unconfirmed activation is reported as a warning. This lets Terraform
retain the truthful untainted draft and schedule exact-version recovery without
repeating the Create PUT. Failures before the draft is validated and marked
remain errors and grant no activation authority.

No activation path fetches or retries a newer draft. Concurrent draft changes
therefore fail Contentful's optimistic-concurrency check instead of publishing
another actor's version.

A nominally successful activation response is accepted only when its
`sys.publishedVersion` equals the exact version sent in the activation request
and its returned `sys.version` is positive and greater. The normally observed
one-version-newer response receives no special treatment; any greater current
version confirms activation and never becomes new authority. Likewise, a
successful draft PUT may return any positive exact version on Create or Update.
Its `publishedVersion` must be known or null and, when present, non-negative and
less than the returned version before that version can be used as the activation
lock token. The provider checkpoints the complete returned response before
reporting a contradiction and revokes authority, so state remains truthful even
when apply fails.

`published_version` did not exist in state written by older provider versions.
A normal post-upgrade refresh projects `sys.publishedVersion`. With
`-refresh=false`, Terraform decodes the missing legacy Computed value as null,
but no pending activation marker exists, so publication remains observational
and the unchanged transition is a no-op.

After Create or Update, state must remain consistent with every known effective
Plan value. Post-mutation state construction starts with the complete response
projection, compares each constrained API-backed value against it, and restores
the exact Plan representation only after every comparison is semantically
equivalent. If Contentful meaningfully contradicts the Plan, the provider emits
an attribute-scoped consistency error and retains complete, truthful
response-derived recovery state; it never hides the contradiction by copying
Plan values over it. Unknown values come from the response. Projection warnings
are retained. A later Read skips this reconciliation and projects the remote
representation so meaningful remote drift remains visible.

When a successful taxonomy mutation response disagrees with the requested
endpoint identity, recovery preserves the complete returned response except
that the requested endpoint identity and legacy ID intentionally remain the
Terraform target.

Apply-time authority follows the effective Plan at the individual attribute,
including attributes nested inside known objects:

- Optional values, including known null and known empty values, are
  Plan-constrained.
- Configuration omission permits an Optional+Computed value to be
  response-owned only while its effective Plan remains unknown. A known value
  supplied by a default, a plan modifier, prior state, or `ignore_changes` is
  Plan-constrained for that apply.
- Computed-only values are response-owned.
- Plan modifiers may preserve prior state for unknown computed plans. Once they
  do, the resulting known Plan value is constrained and must not be replaced by
  a different response value.

Null, omission, and known empty values remain distinct throughout request and
response conversion.

Terraform 1.15.8 with `terraform-plugin-framework` v1.19.0 was exercised with
the provider-side comparison temporarily removed so contradictory mutation
state reached Terraform Core directly. Core rejected known Plan values produced
by static defaults and `UseStateForUnknown` for Webhook `active`, `headers`, and
`headers[*].secret`, and Editor Interface `sidebar[*].disabled`. A contradictory
Create also retained the returned value in tainted recovery state. These are
version-specific observations, so the upgrade triggers below still apply.
Sanitized direct CMA observations for the affected Webhook and Editor Interface
values are recorded in
[CMA test-server conformance boundaries](../research/cma-test-server-conformance.md).

The tables below use these common rules:

- `ignore_changes` may replace changed Config with prior State in the effective
  Plan; reconciliation always uses that effective Plan and never copies Config.
- A known null or empty Plan is compared as a real value. An unknown Plan takes
  the response. A required value that remains null or unknown cannot reach a
  successful mutation because request conversion fails first.
- A semantically equivalent response restores the exact Plan representation.
  A meaningful difference or unsupported/lossy projection retains the complete
  response projection and adds the listed attribute error. Any such mismatch
  prevents every otherwise valid API-backed Plan overlay for that resource.
- Endpoint identity and provider-local `timeouts` are outside that atomic
  overlay. Endpoint identity remains the mutation target; `timeouts` has no CMA
  projection and always retains its effective Plan representation.

The production callers differ only where the endpoint supplies identity or the
request needs Config to decide omission. All six publish response-derived state
and private version before appending consistency diagnostics, allowing
Terraform to retain recovery state when apply reports the error.

| Resource caller | Mutation and identity | Inputs to request conversion | Reconciliation boundary |
| --- | --- | --- | --- |
| Role Create | `POST` to the planned `space_id`; CMA allocates `role_id`. | Effective Plan. | Unknown created identity comes from response; all known Plan values are compared. |
| Role Update | `PUT` to planned `space_id/role_id` with prior private version. | Effective Plan. | Both endpoint identities are pinned; all known Plan values are compared. |
| Editor Interface Create | Versioned `PUT` to planned `space_id/environment_id/content_type_id`. | Effective Plan. | All identity is endpoint-owned and pinned; all known Plan values are compared. |
| Editor Interface Update | Versioned `PUT` to the same planned endpoint, including any version offset. | Effective Plan. | Same reconciliation as Create; prior private version affects only the request lock. |
| Webhook Create | `POST` to planned `space_id`; CMA allocates `webhook_id`. | Effective Plan values; Config is consulted only for Optional+Computed request omission. | Reconciliation receives the effective Plan, not Config. Unknown created identity and unknown Optional+Computed values come from response. |
| Webhook Update | Versioned `PUT` to planned `space_id/webhook_id`. | Effective Plan values; Config is consulted only for Optional+Computed request omission. | Endpoint identity is pinned. Defaults, `UseStateForUnknown`, and `ignore_changes` constrain every known Plan value. |

Role, Editor Interface, and Webhook Read callers do not use mutation
reconciliation. They project the current response, retain only documented
write-only fallbacks, and thereby expose meaningful remote drift.

### Role mutation decisions

| Attribute | Schema and identity | Config, effective Plan, and prior State | Response projection and equivalence | State and diagnostic |
| --- | --- | --- | --- | --- |
| `id` | Computed legacy identity | Config cannot set it. Create Plan is unknown; Update may preserve prior State. | Derived as `space_id/role_id`; no independent CMA field. | Use the endpoint-derived identity. If a known Plan `id` conflicts with that identity, retain the endpoint identity and add an `id` error. |
| `space_id` | Required endpoint identity | Config must be known and non-null. Replacement and `ignore_changes` are already reflected in Plan; prior State otherwise has no authority. | Exact comparison with `sys.space.sys.id`. | Always retain the requested endpoint value. A different response adds a `space_id` error and cannot retarget the resource. |
| `role_id` | Computed response-owned identity on Create; endpoint identity on Update | Config cannot set it. Create Plan is unknown; Update normally preserves prior State. | Exact comparison with `sys.id` when Plan is known. | Unknown Plan takes the created response ID. Known mismatch retains the endpoint ID and adds a `role_id` error. |
| `name` | Required | Config and effective Plan must be known; prior State matters only through `ignore_changes`. Empty is a known value. | Exact scalar comparison. | Equivalent response restores Plan. Difference retains response and adds a `name` error. |
| `description` | Optional | Omitted Config produces known null; explicit empty remains empty. Effective Plan, including prior State selected by `ignore_changes`, is authoritative. | Exact nullable scalar comparison; null and empty are distinct. | Equivalent response restores Plan. Difference retains response and adds a `description` error. |
| `permissions` | Required map of lists | Null/unknown cannot reach mutation. Empty map and empty action lists are known. Effective Plan may come from `ignore_changes`. | Map keys are exact; each action list is unordered while duplicate multiplicity is preserved. Unsupported permission variants make projection lossy. | Equivalent response restores the exact map/list Plan representation. Difference adds a `permissions` error; lossy projection retains its warning and adds the same scoped consistency error. |
| `policies` | Required list of objects | Null/unknown cannot reach mutation; an empty list is known. Effective Plan may come from `ignore_changes`. | Policy order is insignificant while duplicate multiplicity is preserved. Each policy matches only when all nested attributes below are equivalent. Unsupported policy variants make projection lossy. | Equivalent response restores exact Plan order and representation. Difference adds a `policies` error; lossy projection also retains its warning. |
| `policies[].actions` | Required list | Null/unknown cannot reach mutation; empty is known when accepted by validation. | Unordered string comparison preserving duplicates; Contentful scalar `"all"` and Terraform `["all"]` are the established request/response projection. | Participates in the parent `policies` decision and diagnostic. |
| `policies[].constraint` | Optional normalized JSON string | Omitted is null; known JSON, including `{}` or `[]`, remains distinct from null. | JSON semantic comparison normalizes representation; array order remains meaningful. Unsupported JSON projection prevents equivalence. | Exact Plan JSON text is restored only when equivalent; otherwise the complete policies response remains with a `policies` error. |
| `policies[].effect` | Required | Must be known and non-null; empty remains a known scalar if it passes earlier validation. | Exact scalar comparison. | Participates in the parent `policies` decision and diagnostic. |
| `timeouts` | Optional provider-local object | Omitted is null; configured children and values selected by `ignore_changes` are known Plan values. | No CMA projection or comparison. | Always retain exact Plan. No mutation-consistency diagnostic. |
| `timeouts.create`, `timeouts.read`, `timeouts.update`, `timeouts.delete` | Optional provider-local strings | Omitted children are null; configured durations remain exact Plan strings. | Parsed only for the corresponding operation; never returned by CMA. | Retained inside `timeouts`; invalid durations fail validation before mutation. |

### Editor Interface version preconditions

Editor Interface Create sends `X-Contentful-Version: 1` plus the configured
provider's counted Content Type activations. This permits management of an
initial automatically created interface, including after a subsequent Content
Type activation earlier in the same apply. It does not fetch a newer version to
adopt an existing modified interface; deliberate adoption requires import.
A mismatch establishes only that the version precondition failed, not who
changed the interface or whether import is the only remedy.

Update requires the private `version` last observed by Read or a successful
interface mutation, plus the same activation count. It sends the effective
Plan with that version and fails on `VersionMismatch` without a fresh GET or
retry. This preserves optimistic concurrency for edits after refresh or saved
plan creation. Read and successful Create/Update publish the actual response
`sys.version` and reset the count after publishing state; mutation consistency
diagnostics are reported after that checkpoint.

The counter is in-memory and scoped to one configured provider, keyed by
`space_id`, `environment_id`, and `content_type_id`. It is not Content Type
`sys.version`, persistent Terraform state, or evidence about other writers.
Resources changing together must use the same provider configuration and a
Terraform dependency, normally a reference to the Content Type resource.
Separate aliases do not exchange offsets, even with identical credentials and
correct dependency ordering. An Editor Interface update can therefore conflict
after a Content Type activation through another alias. Review a refreshed plan
before applying again; an interface whose initial version cannot be accounted
for may require deliberate import. A provider restart discards the count;
refresh supplies the next observed baseline. The counter neither orders
resources nor causes an otherwise unplanned Editor Interface update.

Activation of an unpublished Content Type creates interface version 1; it does
not increment the offset from that initial base. Activation of an already
published Content Type advances the existing interface once. Determine this
from the known **pre-activation** checkpoint's response-derived
`published_version`, including when recovering a pending activation through
Update. The post-activation value cannot distinguish these cases. Count an
existing interface's increment after checkpointing the activation response and
before reporting consistency or activation diagnostics. Failed activation
responses do not reach that checkpoint and contribute no count.

Direct CMA experiments establish these mock and lifecycle test oracles:

| Content Type operation | Editor Interface observation |
| --- | --- |
| Create an unpublished draft | GET returns 404. |
| First activation | Interface starts at version 1. |
| Update the draft before activation | Interface version is unchanged. |
| Activate a published type after description-only, field-name, or added-field changes | Interface version increments once. |
| Activate again with no content changes | Interface version increments once. |
| Deactivate | GET returns 404. |
| Reactivate after deactivation | Default interface at version 1; previous widget settings are absent. |

Recovery after a rejected initial activation is covered separately by mocked
Terraform tests, with and without refresh. The live first-activation observation
supplies the independent version-1 expectation. These observations describe the
tested CMA behavior, not a guarantee of all future server behavior. Deactivation removes the mock interface; it does not
clear unrelated field-publication history. Editor Interface Read removes a
missing resource from Terraform state. This does not grant authority to
reactivate an externally deactivated Content Type. Editor Interface Delete
only relinquishes Terraform ownership and makes no CMA deletion request.

### Editor Interface mutation decisions

| Attribute | Schema and identity | Config, effective Plan, and prior State | Response projection and equivalence | State and diagnostic |
| --- | --- | --- | --- | --- |
| `id` | Computed legacy identity | Config cannot set it. Create Plan is unknown; Update may preserve prior State. | Derived as `space_id/environment_id/content_type_id`; no independent CMA field. | Use the endpoint-derived identity. A conflicting known Plan adds an `id` error. |
| `space_id`, `environment_id`, `content_type_id` | Required endpoint identity | Each Config value must be known and non-null. Replacement and `ignore_changes` are already reflected in Plan. | Exact comparison with the corresponding `sys` link. | Always retain each requested endpoint value. A different response adds an error at that identity path and cannot retarget the resource. |
| `editor_layout` | Optional ordered list | Omitted is null; explicit `[]` is distinct. A known Plan, including one selected by `ignore_changes`, is authoritative; unknown takes response. | Ordered, exact structural comparison. The schema cannot represent top-level field layout items, so such a response is lossy. | Equivalent response restores exact Plan. Difference adds an `editor_layout` error. Lossy projection retains its warning and adds the same scoped consistency error when Plan is known. |
| `editor_layout[].group` | Required object | Required whenever the parent element is present. | Exact object structure. | Participates in `editor_layout`; mismatch is reported at `editor_layout`. |
| `editor_layout[].group.group_id`, `editor_layout[].group.name` | Required strings | Must be known and non-null; empty is still a known scalar. | Exact comparison. | Participates in `editor_layout`. |
| `editor_layout[].group.items` | Required ordered list | Must be known; `[]` is valid and distinct from null. | Ordered structural comparison. | Participates in `editor_layout`. |
| `editor_layout[].group.items[].field`, `editor_layout[].group.items[].group` | Optional exclusive objects | Exactly one must be configured. Null marks the unselected alternative; unknown cannot reach request conversion. | Exact selected-alternative and structure comparison. | Participates in `editor_layout`; unsupported alternatives cannot manufacture equivalence. |
| `editor_layout[].group.items[].field.field_id` | Required string | Must be known and non-null. | Exact comparison. | Participates in `editor_layout`. |
| `editor_layout[].group.items[].group.group_id`, `editor_layout[].group.items[].group.name` | Required strings | Must be known and non-null. | Exact comparison. | Participates in `editor_layout`. |
| `editor_layout[].group.items[].group.items` | Required ordered list | Must be known; `[]` is distinct from null. | Ordered structural comparison. | Participates in `editor_layout`. |
| `editor_layout[].group.items[].group.items[].field` | Required object | Must be present and known. | Exact structure comparison. | Participates in `editor_layout`. |
| `editor_layout[].group.items[].group.items[].field.field_id` | Required string | Must be known and non-null. | Exact comparison. | Participates in `editor_layout`. |
| `controls` | Optional ordered list | Omitted is null; `[]` is known. Unknown takes response; known Plan may reflect `ignore_changes`. | Ordered element comparison; nested `settings` uses JSON semantics. | Equivalent response restores exact Plan. Difference or lossy projection adds a `controls` error and retains response. |
| `controls[].field_id` | Required string | Must be known and non-null. | Exact comparison. | Participates in `controls`. |
| `controls[].widget_namespace`, `controls[].widget_id` | Optional strings | Omitted is null; empty remains distinct. | Exact nullable scalar comparison. | Participates in `controls`. |
| `controls[].settings` | Optional normalized JSON string | Omitted is null; known JSON remains distinct from null. | JSON semantic comparison; array order remains meaningful. | Equivalent response restores exact Plan JSON representation; otherwise `controls` fails atomically. |
| `group_controls` | Optional ordered list | Omitted is null; `[]` is known. Unknown takes response; known Plan may reflect `ignore_changes`. | Ordered element comparison; nested `settings` uses JSON semantics. | Equivalent response restores exact Plan. Difference or lossy projection adds a `group_controls` error and retains response. |
| `group_controls[].group_id` | Required string | Must be known and non-null. | Exact comparison. | Participates in `group_controls`. |
| `group_controls[].widget_namespace`, `group_controls[].widget_id` | Optional strings | Omitted is null; empty remains distinct. | Exact nullable scalar comparison. | Participates in `group_controls`. |
| `group_controls[].settings` | Optional normalized JSON string | Omitted is null; known JSON remains distinct from null. | JSON semantic comparison. | Equivalent response restores exact Plan JSON representation; otherwise `group_controls` fails atomically. |
| `sidebar` | Optional ordered list | Omitted is null; `[]` is known. Unknown takes response; known Plan may reflect `ignore_changes`. | Ordered element comparison; nested `settings` uses JSON semantics and `disabled` is exact. | Equivalent response restores exact Plan. Difference or lossy projection adds a `sidebar` error and retains response. |
| `sidebar[].widget_namespace`, `sidebar[].widget_id` | Required strings | Must be known and non-null. | Exact comparison. | Participates in `sidebar`. |
| `sidebar[].settings` | Optional normalized JSON string | Omitted is null; known JSON remains distinct from null. | JSON semantic comparison. | Equivalent response restores exact Plan JSON representation; otherwise `sidebar` fails atomically. |
| `sidebar[].disabled` | Optional+Computed with static `false` default | Omission normally yields known `false`, so prior State or a differing response cannot override it during apply. Only an actually unknown effective Plan is response-owned. | Exact boolean comparison. CMA has been observed to omit the property when it was not sent and to return explicit `false` when sent by the provider. | Equivalent response restores Plan. Difference retains response and adds a `sidebar` error. |
| `timeouts` | Optional provider-local object | Omitted is null; configured children and `ignore_changes` results are Plan-owned. | No CMA projection. | Always retain exact Plan without a consistency diagnostic. |
| `timeouts.create`, `timeouts.read`, `timeouts.update` | Optional provider-local strings | Omitted children are null; configured durations remain exact Plan strings. | Parsed only for the corresponding operation. | Retained inside `timeouts`; invalid durations fail before mutation. |

### Webhook mutation decisions

| Attribute | Schema and identity | Config, effective Plan, and prior State | Response projection and equivalence | State and diagnostic |
| --- | --- | --- | --- | --- |
| `id` | Computed legacy identity | Config cannot set it. Create Plan is unknown; Update may preserve prior State. | Derived as `space_id/webhook_id`; no independent CMA field. | Use endpoint-derived identity. A conflicting known Plan adds an `id` error. |
| `space_id` | Required endpoint identity | Must be known and non-null. Replacement and `ignore_changes` are already reflected in Plan. | Exact comparison with `sys.space.sys.id`. | Always retain requested endpoint value. Difference adds a `space_id` error and cannot retarget the resource. |
| `webhook_id` | Computed response-owned identity on Create; endpoint identity on Update | Config cannot set it. Create Plan is unknown; Update normally preserves prior State. | Exact comparison with `sys.id` when Plan is known. | Unknown Plan takes created response ID. Known mismatch retains endpoint ID and adds a `webhook_id` error. |
| `active` | Optional+Computed with static `true` default | Omitted Config normally yields known `true`; explicit Config, prior State selected by planning, and `ignore_changes` likewise produce the effective Plan. Only an actually unknown Plan is response-owned. | Exact boolean comparison. CMA has also been observed to default an omitted raw request member to `true`; the provider normally sends its known default explicitly. | Equivalent response restores Plan. Difference retains response and adds an `active` error. |
| `name`, `url` | Required strings | Must be known and non-null; empty remains a known value. Prior State matters only through `ignore_changes`. | Exact scalar comparison. | Equivalent response restores Plan. Difference retains response and adds an error at `name` or `url`. |
| `topics` | Required non-empty ordered list of strings | Null, unknown, empty, or null-element values fail validation or request conversion before mutation. The known non-empty effective Plan, including a value selected by `ignore_changes`, is authoritative. | Exact ordered comparison; duplicates remain meaningful. | Equivalent response restores the exact effective Plan. Difference retains response and adds a `topics` error. |
| `filters` | Optional unordered list of operator objects | Omitted is null; explicit `[]` is distinct. Unknown takes response; known Plan may reflect `ignore_changes`. | Outer order is insignificant while duplicate multiplicity is preserved. Exactly one recognized operator must project for each element. Unknown operators make projection lossy. | Equivalent response restores exact Plan order and representation. Difference adds a `filters` error; lossy projection retains its warning and adds the scoped consistency error when Plan is known. |
| `filters[].not`, `filters[].equals`, `filters[].in`, `filters[].regexp` | Optional exclusive objects | Exactly one must be selected; unselected alternatives are null. Unknown cannot reach request conversion. | Selected operator and all nested values must match. | Participates in `filters`; unsupported or multiple response operators cannot establish equivalence. |
| `filters[].not.equals`, `filters[].not.in`, `filters[].not.regexp` | Optional exclusive objects | Exactly one nested negated operator must be selected. | Same operator-specific rules as their non-negated forms. | Participates in `filters`. |
| `filters[].equals.doc`, `filters[].equals.value`, `filters[].not.equals.doc`, `filters[].not.equals.value` | Required strings | Must be known and non-null. | Exact comparison. | Participates in `filters`. |
| `filters[].in.doc`, `filters[].not.in.doc` | Required strings | Must be known and non-null. | Exact comparison. | Participates in `filters`. |
| `filters[].in.values`, `filters[].not.in.values` | Required lists | Must be known; `[]` is distinct from null. | Unordered string comparison preserving duplicate multiplicity. | Participates in `filters`; equivalent response restores the exact Plan order. |
| `filters[].regexp.doc`, `filters[].regexp.pattern`, `filters[].not.regexp.doc`, `filters[].not.regexp.pattern` | Required strings | Must be known and non-null. | Exact comparison. | Participates in `filters`. |
| `http_basic_username` | Optional string | Omitted is null; empty is known. Effective Plan, including `ignore_changes`, is authoritative. | Exact nullable scalar comparison. | Equivalent response restores Plan. Difference retains response and adds an `http_basic_username` error. |
| `http_basic_password` | Optional sensitive write-only API value | Omitted and explicit null remain null; configured, rotated, prior-State, and `ignore_changes` cases use the effective Plan. Unknown configured input fails before mutation. | Apply the established password table above: only property absence permits the narrow known non-null Plan fallback; explicit null or a returned value is response truth. | Absent property restores the known non-null Plan even when another attribute mismatches. Explicit contradiction retains response and adds an `http_basic_password` error. Read continues to use prior-state preservation without claiming remote equality. |
| `headers` | Optional+Computed map | Omitted Create commonly leaves Plan unknown and therefore response-owned. On Update, `UseStateForUnknown`, explicit Config, or `ignore_changes` may make Plan known and constrained. Preserving a known prior value avoids CMA's observed behavior of clearing headers when an Update omits the member; `{}` remains the explicit clear operation. Known null and `{}` remain distinct. | Map keys and complete header objects compare exactly after the established secret-value fallback. CMA has returned `[]` for omitted raw Create and Update members. | Unknown Plan takes response. Equivalent known Plan restores its exact representation. Difference adds a `headers` error. |
| `headers[*].value` | Required string | Must be known for configured headers. Prior State may supply a secret value through the known `headers` Plan. | Exact comparison when CMA returns it. For a secret header whose value CMA omits, response projection uses only the corresponding Plan value as the established narrow fallback. | Fallback value is retained even during another contradiction because CMA supplies no competing value. A returned different value causes the parent `headers` error. |
| `headers[*].secret` | Optional+Computed with static `false` default and `UseStateForUnknown` | Omission inside a configured header normally yields known `false`; prior State may be preserved when planning leaves it unknown. Any known effective Plan is constrained. | Exact boolean comparison; response projection treats an absent flag as `false`. CMA has returned an ordinary header value with the flag absent and has echoed explicit `false`. | Difference causes the parent `headers` error; equivalence restores exact Plan. |
| `transformation` | Optional object | Omitted is null; a known object may contain null children. Unknown takes response; `ignore_changes` may select prior State. | Object comparison uses exact scalars and JSON semantics for `body`. | Equivalent response restores exact Plan. Difference or lossy projection adds a `transformation` error and retains response. |
| `transformation.method`, `transformation.content_type`, `transformation.include_content_length` | Optional scalars | Omitted children are null; empty strings and explicit booleans are known. | Exact nullable scalar comparison. | Participates in `transformation`. |
| `transformation.body` | Optional normalized JSON string | Omitted is null; known JSON remains distinct from null. | JSON semantic comparison; array order remains meaningful. | Equivalent response restores exact Plan JSON text; otherwise `transformation` fails atomically. |
| `timeouts` | Optional provider-local object | Omitted is null; configured children and `ignore_changes` results are Plan-owned. | No CMA projection. | Always retain exact Plan without a consistency diagnostic. |
| `timeouts.create`, `timeouts.read`, `timeouts.update`, `timeouts.delete` | Optional provider-local strings | Omitted children are null; configured durations remain exact Plan strings. | Parsed only for the corresponding operation. | Retained inside `timeouts`; invalid durations fail before mutation. |

### CMA transport retry safety

The provider applies one transport retry policy to CMA calls across resources,
data sources, and list operations. This is a provider-wide safety boundary, not
resource-specific behavior.

GET, HEAD, and OPTIONS do not carry provider mutations, so replay cannot
duplicate a provider write. They retain retries after retryable transport
failures and retryable server responses. POST, PUT, PATCH, and DELETE are not
transparently replayed after transport failures or ordinary retryable 5xx
responses. In those cases, the provider cannot establish the remote outcome:
after the request was sent, Contentful may have committed the mutation even
though the provider received a transport failure or 5xx instead of a usable
success response. Automatic replay could therefore repeat a committed mutation
or reuse a now-stale optimistic-lock version. An ordinary 5xx establishes
neither commitment nor rejection.
This policy addresses ambiguous observation, not whether every CMA mutation is
inherently non-idempotent.

Contentful documents `429 Too Many Requests` as rate limiting and tells clients
to wait before making another request; its first-party management SDK also
retries 429 responses. The provider follows that Contentful-specific policy for
every HTTP method by default. Entry Create, specified-ID Create, Update, and
Publish and Content Type Create, Update, and Activate are the narrow exception:
they return the first 429 without transparent replay because the response cannot
establish mutation commitment or exact-version authority. The evidence,
deadline, and backoff contracts are recorded in
[Contentful HTTP retry policy](contentful-http-retry-policy.md).

### Entry specified-ID request selection

The Entry specified-ID endpoint serves both Create and Update, so the provider
selects the operation through its request headers. Create sends
`X-Contentful-Content-Type` and no `X-Contentful-Version`. A collision must fail
without mutating, publishing, or adopting the existing Entry. Update sends the
exact prior `sys.version` in `X-Contentful-Version` and omits
`X-Contentful-Content-Type`.

If an Entry disappears after refresh, or refresh is skipped and Terraform state
is stale, the Update request therefore fails instead of recreating the absent
Entry and proceeding toward publication. A preflight existence read cannot
provide this boundary because another writer can change target existence
between the read and the `PUT`. The Contentful contract and direct CMA evidence
are recorded in
[Entry and Content Type PUT header semantics](../research/entry-and-content-type-put-headers.md).

### Entry publication ownership and partial field ownership

`contentful_entry` publishes only the exact draft returned by the same Create or
Terraform-managed Update. It checkpoints a successful, plan-consistent draft
response and its `sys.version` before sending that version to Publish. Import,
Read, refresh, an external draft, external unpublish, and prior Terraform state
are observations only and never grant publication authority.

After the complete truthful draft state and optimistic-lock version are
checkpointed and the response identity, positive version, draft tuple, and
plan consistency are validated, private state records only that exact version
as pending publication authority. The provider attempts Publish with that
version in the same operation. Confirmed publication checkpoints the returned
state/version and clears the marker. An explicit or ambiguous publication
failure retains the truthful draft and marker; an unchanged later operation is
deliberately planned as an Update and publishes only the marker version without
another draft PUT. During Create, an unconfirmed Publish after this checkpoint
is reported as a warning so Terraform retains the untainted draft and marker;
failures before that boundary remain errors and grant no publication authority.

Read preserves the marker only while current `sys.version` exactly equals the
marker and the observed publication tuple exactly equals the checkpointed draft
tuple. Observing the marker as
published clears it without replay. A different current version or publication
state revokes authority without mutation. With refresh disabled, recovery sends
only the marker version; `VersionMismatch` revokes it and never causes a fetch
and publication of a newer version.

An ambiguous draft-mutation failure likewise grants no publication authority.
The client does not replay the mutation automatically, and the provider does not
infer draft ownership from a later GET whose fields happen to match the plan.
This fail-closed rule avoids publishing an indistinguishable concurrent editor
version.

Within an Entry `fields` map, Terraform null means request omission. A known
JSON null produced by `jsonencode(null)` remains a value and is sent as JSON
null. CMA omits that raw JSON-null field from mutation and read responses, so
the provider restores its exact planned or prior representation only when the
response omits it; a present response value wins. A localized object such as
`{"en-US":null}` remains ordinary response data and receives no fallback.
Change detection compares the effective plan and prior state after the same
request projection. Adding or removing an omitted Terraform-null key may
therefore require a state-only Update, but it does not write or publish an
Entry; the provider stores the effective plan representation in Terraform state
while retaining the prior response-owned `published_version`. The direct
evidence is recorded in
[Entry null and omission behavior](../research/entry-null-and-omission.md).

Entry Read is authoritative for `fields`, including additions and removals,
except for configured request omissions and Contentful's observed empty-field
canonicalization. When Contentful omits a key whose prior value is Terraform
null or raw JSON null, the provider retains that exact representation; a
returned value always takes precedence and exposes drift. When Contentful
omits the entire `fields` member for a known empty map, or omits a localized
field whose prior value contains only empty arrays, the provider retains that
known empty value. It does not extend fallback to localized null, other scalar
or object values, or nonempty arrays, so a remote removal of meaningful data
remains observable.

Terraform's effective plan, after lifecycle processing such as `ignore_changes`,
is the mutation request and consistency boundary. This lets ignored remote map
paths survive a later managed update while ensuring an ignored-only external
draft remains untouched. Mutation responses may restore a known empty map when
the entire `fields` member is omitted, but omission cannot stand in for a nonempty
effective plan. They may also restore planned Terraform-null, raw JSON-null, or
localized all-empty-array fields when the response omits those individual keys.
Other partial or changed maps remain response truth and a plan contradiction.

Entry metadata tags and taxonomy concepts retain their public List schema.
Configured values must be unique. Comparison ignores order but remains
multiplicity-sensitive for response/state data so irregular remote duplicates
are not hidden. A configuration-only reorder changes only Terraform's stored
list representation; it does not write or publish an Entry.

Contentful may apply Content Type default values to fields omitted during Entry
creation. A Create response may therefore add response-only field keys,
including a value for a Terraform-null key omitted from the request, but every
sent field must still be present and semantically equal after the recognized
empty-array and raw-JSON-null omission fallbacks above. Update uses full-body
replacement and Contentful does not apply defaults to Entry updates, so an
Update response must have exactly the sent field keys and values after those
same fallbacks. Create publication responses retain the same creation-default
projection used for the Create draft response; Update and recovery publication
responses require the exact effective-plan field keys. Field ownership does not
depend on publication-version arithmetic.

After those checks succeed, mutation state restores the exact effective-plan
`fields`, `metadata`, and `timeouts`; identity and lifecycle values remain
response-owned where the plan does not already establish an endpoint identity.
A later Read remains authoritative and can expose Contentful-added defaults,
which `ignore_changes` can preserve through a subsequent managed update. When
Contentful contradicts a planned field, metadata, content type, or endpoint
identity, the provider reports an error and checkpoints the representable
response `fields` and `metadata`, while retaining the requested immutable content
type and endpoint identity. The complete returned response remains the truthful
checkpoint, while response-only field additions cannot become
configuration-owned values or new publication authority.

A Create or Update draft response grants authority to any positive exact
returned version in a plan-consistent response. Either response may omit
`publishedVersion` or contain a non-negative older publication
(`publishedVersion < version`); equal, future, negative, or unknown publication
values are contradictory.

Publish must report the submitted version as `publishedVersion`. The repeatedly
observed normal response has `version` equal to the submitted version plus one,
but the provider applies no policy to that arithmetic: any greater current
version confirms publication. A current version equal to or below the published
version is contradictory. Missing or different `publishedVersion` and
nonpositive current versions remain errors. Read cannot use a malformed or
changed tuple to preserve publication authority. The computed
`published_version` attribute exposes the authoritative CMA value; `version` in
resource private state always records the same response version as the published
Terraform state.

### Provider-private optimistic-lock barrier

For resources whose Contentful mutations require a version stored only in
provider-private state, inability to read or decode that version is a terminal
apply diagnostic. The provider must return before issuing any Contentful
mutation; a missing or malformed lock token must never degrade to version zero
or an unlocked request. Successfully decoded integer values are forwarded
without provider-side range validation, leaving Contentful to determine their
validity. This preserves Go's established integer decoding of JSON `null` as
zero. Taxonomy Delete is the one absence-handling exception: when Terraform
omits private data for a tainted replacement, the provider obtains the current
version with a GET before deleting.

Pending Entry publication and Content Type activation markers are a narrower
mutation-authority boundary, not ordinary optimistic-lock forwarding. A marker
must decode to a positive exact draft version; zero, negative, malformed, or
missing marker data cannot authorize Publish or Activate.

Publish or Activate confirmation requires `publishedVersion` equal to the
submitted exact version and a positive current `version` greater than it. The
normally observed `V+1` tuple receives no special treatment. Any greater current
version is checkpointed truthfully and accepted, but it never replaces the
marker or becomes new mutation authority. A missing or wrong
`publishedVersion`, or current `version <= V`, is contradictory and revokes
authority after the complete response is checkpointed.

All Entry Create, specified-ID Create, Update, and Publish requests and Content
Type Create, Update, and Activate requests disable transparent HTTP replay,
including for 429, transport failures, and 5xx. Read and unrelated CMA calls
retain their existing retry behavior.

### Entry destroy lifecycle

Entry destroy consumes no private `version`, performs no recovery GET, and
sends neither `X-Contentful-Version` nor `If-Match`. It always requests
whole-Entry unpublish before delete. A 404 response is benign, as is an exact
HTTP 400 response with `sys.id` `BadRequest` and message `Not published`;
destroy continues to delete after either result. Any other unpublish error
stops destroy before delete. Because Contentful does not enforce preconditions
on these operations, an Entry changed outside Terraform can still be deleted.
This policy is limited to Entry destroy; Entry update, Entry publish, and other
endpoints are out of scope.

See [Entry destroy lifecycle evidence](../research/entry-destroy-lifecycle.md).

### Taxonomy optimistic version locking

Taxonomy resources record Contentful `sys.version` in provider-private state,
matching the provider's other versioned resources. Update sends the prior
private-state value supplied at apply. Delete also sends that value when it is
present. A remote change for which CMA advances the resource's `sys.version` is
therefore rejected instead of being overwritten. Create, Read, and Update store
the integer version returned by Contentful, including alongside recovery state
when an Update response differs from the plan.

Terraform does not supply prior private data to Delete while replacing a
tainted resource. This applies both to recovery state published alongside an
errored Create and to a resource manually marked tainted after successful
updates. Only when private version data is genuinely absent, Delete performs
one GET for the requested resource and sends its `sys.version` in the DELETE
request. A 404 from this GET means the resource is already absent. This narrow
apply-time fallback lets tainted replacement delete resources above version
`1`; a concurrent change between GET and DELETE still produces CMA
`VersionMismatch` instead of a blind deletion. Valid private data takes the
direct DELETE path without a preliminary GET. Malformed private version data
remains an error and does not trigger the fallback; decoded zero and negative
values are sent to CMA.

Because no prior lock token exists in the fallback case, Delete cannot detect a
remote change that occurred before its GET; it deliberately authorizes deletion
of the version current at that point. Optimistic locking then covers the
GET-to-DELETE interval.

CMA documents the version header as required. Direct concept and concept-scheme
PATCH and DELETE experiments rejected omitted, zero, negative, and stale
versions without changing or deleting the resource. This observed positive,
exact-current constraint is enforced by CMA rather than duplicated as provider
policy.

Deleting a concept can remove references from other concepts and schemes
without advancing those referencing resources' versions, so their version
headers cannot detect that cascade. Field-scoped patches avoid rewriting
unrelated cascaded fields; CMA and post-mutation response consistency checks
still govern fields the provider explicitly changes.

The repository mock and direct CMA creation experiments returned an initial
version of `1`; that observation is modeled for service fidelity but is not a
permanent schema promise. The documented contract and raw observations are
recorded in the
[taxonomy version note](../research/taxonomy-version.md). Use Contentful's
established `version` and `sys.version` terminology.

### Delivery API key environments

`delivery_api_key.environments` has a deliberate forward-compatibility policy:

- Config null with a null or unknown Plan omits the request member and asks
  Contentful to choose its default;
- a known Plan is serialized even when Config is null, so a default or
  state-preserved value remains the request source;
- Config known with Plan unknown is rejected before Contentful I/O;
- known empty serializes an explicit `[]`; and
- response conversion reflects the environment links returned by Contentful
  without inventing a default identifier.

Contentful has previously canonicalized both omission and `[]` to a non-empty
default environment. When that happens, an explicitly configured empty list
cannot equal the post-mutation state, and Terraform may report an inconsistent
result. Rejecting empty lists would avoid that symptom only by making current
service behavior a permanent provider restriction. The provider instead keeps
empty input valid and treats the canonicalization conflict as a known
limitation. The supporting observations and their date are recorded in the
[Delivery API key environments note](../research/delivery-api-key-environments.md).

### Extension sources

Every Contentful Extension mutation request requires exactly one of
`extension.src` or `extension.srcdoc`. HCL may omit both for an imported or
otherwise response-owned source; planning preserves the sole prior-state source
in that case. An empty `src` is invalid, while an explicitly empty `srcdoc` is
accepted and round-trips. The provider does not silently rewrite either
explicit value to omission. The first-party contract and live CMA observations
are recorded in the
[Extension source values note](../research/extension-source-values.md).

### Space enablements

All four Space Enablements attributes are independent Optional+Computed values.
Request conversion sends each known Plan value, including explicit `false` and
values preserved from prior state, and omits a response-owned null or unknown
Plan value. A configuration-owned unknown Plan value is rejected before I/O.

Contentful currently validates the presence and equality of
`cross_space_links` and `space_templates`, but the CMA request schema represents
all four members as optional. The provider does not make that current server
policy a Terraform configuration invariant: it sends the planned members
without inferring a sibling or requiring equality, and surfaces any CMA
validation response. The first-party contract and live CMA observations are
recorded in the [Space Enablements request values note](../research/space-enablement-values.md).

## Diagnostics and local publication

Warnings from a representable response projection accompany the published
result. Encoding errors must not publish half of a state/identity pair: both
values are encoded into staged copies, and both are assigned only after both
encodings succeed. This is local publication only; it does not make remote API
operations or private provider data transactional.

An error diagnostic does not universally prohibit returning state. Terraform
can persist state returned with an error so that a provider can record remote
effects completed before a later failure. Any such intermediate state must be
truthful and is a resource-local lifecycle decision because the resource owns
the meaning of its remote operations. The generic local publication boundary
does not infer remote success or construct recovery state.

List results follow the same local rule. When a resource representation is
requested, identity and resource encoding are staged and published together;
warnings accompany a complete result, while encoding errors publish neither.
An identity-only result remains valid when the request did not ask for the
resource representation.

## Upgrade triggers

Re-check these policies and their focused tests when any of the following
change:

- `terraform-plugin-framework`, Terraform Core, or the provider protocol;
- schema flags, defaults, plan modifiers, or resource identity schemas;
- generated OpenAPI enums, discriminators, or generator behavior;
- Contentful's null, empty, or defaulting behavior; or
- a resource gains a new multi-request lifecycle or private-state contract.

## Primary sources

- HashiCorp, [accessing values and lifecycle guarantees](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values#when-can-a-value-be-unknown-or-null).
- HashiCorp, [validation](https://developer.hashicorp.com/terraform/plugin/framework/validation), [diagnostics and error state](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics#how-errors-affect-state), and [plan consistency](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#terraform-data-consistency-rules).
- HashiCorp, [list resources](https://developer.hashicorp.com/terraform/plugin/framework/list-resources) and [resource identity](https://developer.hashicorp.com/terraform/plugin/framework/resources/identity).
- Framework v1.19.0 [`tfsdk.State`](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.19.0/tfsdk/state.go), [`tfsdk.ResourceIdentity`](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.19.0/tfsdk/resource_identity.go), and [Create handling](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.19.0/internal/fwserver/server_createresource.go).
- The provider's local [editor-layout union](../../internal/contentful-management-go/openapi/schemas/editor-interface/layout-items.yml) and [taxonomy metadata enum](../../internal/contentful-management-go/openapi/schemas/content-type/metadata-taxonomy-item.yml), which illustrate closed generated-client alternatives.
