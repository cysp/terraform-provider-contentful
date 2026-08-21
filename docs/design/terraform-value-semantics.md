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

## Lifecycle ownership and plan consistency

`contentful_content_type` owns activation state: the resource represents an
activated Content Type, not an arbitrary draft definition. The lifecycle has
these deliberate reconciliation paths:

- During Update, after Terraform saves a modeled draft and activation fails,
  the provider checkpoints the returned draft and exact version. An unchanged
  retry activates that version without repeating the draft update.
- During Create, the provider likewise checkpoints the returned draft before
  returning an activation error. Terraform Core taints any resource whose
  Create returns an error, however, so the next normal plan replaces the
  resource and repeats the draft PUT. The checkpoint preserves truthful state
  for recovery and cleanup; it cannot turn a failed Create into an in-place
  Update without suppressing the activation error.
- External deactivation plans an activation-only update of the exact observed
  draft version.
- A newer external draft with no modeled drift is intentionally activated in
  place. Because no complete draft update is sent, values outside the provider
  schema survive unchanged.
- A newer external draft with modeled drift is first replaced by the complete
  provider-modeled request using the exact observed prior version. The provider
  checkpoints and activates the exact returned version; unmodeled values are
  not guaranteed to survive that complete update.

No activation path fetches and retries a newer draft. Concurrent draft changes
therefore fail Contentful's optimistic-concurrency check instead of publishing
another actor's version. The provider does not issue a confirming GET after an
activation error. A lost or otherwise ambiguous activation response is
not interpreted as success: a normal refresh establishes publication truth
before Terraform decides whether another activation is required.

`published_version` did not exist in state written by older provider versions.
A normal post-upgrade refresh projects `sys.publishedVersion` and establishes
authoritative publication state. With `-refresh=false`, Terraform decodes a
missing legacy Computed value as null, which is indistinguishable from an
observed unpublished Content Type. The provider therefore conservatively may
plan exact-version activation in that case; it does not persist a migration or
recovery marker to distinguish provenance.

After Create or Update, state must remain consistent with every known
configuration-owned value in the plan. Post-mutation state construction starts
with the complete response projection, compares each owned value against it,
and restores the exact plan representation only after every comparison is
semantically equivalent. If Contentful meaningfully contradicts the plan, the
provider emits an attribute-scoped consistency error and retains complete,
truthful response-derived recovery state; it never hides the contradiction by
copying plan values over it. Unknown and response-owned values come from the
response. Projection warnings are retained. A later Read skips this
reconciliation and projects the remote representation so meaningful remote
drift remains visible.

When a successful taxonomy mutation response disagrees with the requested
endpoint identity, recovery preserves the complete returned response except
that the requested endpoint identity and legacy ID intentionally remain the
Terraform target.

Ownership follows the schema at the individual attribute, including attributes
nested inside known objects:

- Optional values, including known null and known empty values, are
  configuration-owned.
- Optional+Computed values are response-owned when configuration omits them;
  known configured values remain configuration-owned.
- Computed-only values are response-owned.
- Plan modifiers may preserve prior state for unknown computed plans, but must
  not replace an unknown planned value for a configuration-owned attribute or
  turn known empty configuration into unknown.

Null, omission, and known empty values remain distinct throughout request and
response conversion.

### Taxonomy optimistic version locking

Taxonomy resources record Contentful `sys.version` in provider-private state,
matching the provider's other versioned resources. Update sends the prior
private-state value supplied at apply. Delete also sends that value when it is
present. A remote change for which CMA advances the resource's `sys.version` is
therefore rejected instead of being overwritten. Create, Read, and Update store
a positive version returned by Contentful, including alongside recovery state
when an Update response differs from the plan. After a successful Create or
Update with a nonpositive response version, the provider first checkpoints the
projected response state and identity, then returns an error without storing the
unusable version. An errored Create therefore has state but no private version;
an errored Update retains its prior private version until Read obtains a valid
current version.

Terraform does not supply prior private data to Delete while replacing a
tainted resource. This applies both to recovery state published alongside an
errored Create and to a resource manually marked tainted after successful
updates. Only when private version data is genuinely absent, Delete performs
one GET for the requested resource, validates the returned identity and
positive `sys.version`, and sends that version in the DELETE request. A 404
from this GET means the resource is already absent. This narrow apply-time
fallback lets tainted replacement delete resources above version `1`; a
concurrent change between GET and DELETE still produces CMA `VersionMismatch`
instead of a blind deletion. Valid private data takes the direct DELETE path
without a preliminary GET. Malformed, zero, or negative private version data
remains an error and does not trigger the fallback.

Because no prior lock token exists in the fallback case, Delete cannot detect a
remote change that occurred before its GET; it deliberately authorizes deletion
of the version current at that point. Optimistic locking then covers the
GET-to-DELETE interval.

CMA documents the version header as required, and direct concept and concept-
scheme DELETE experiments rejected both an omitted header and version `0`.

Deleting a concept can remove references from other concepts and schemes
without advancing those referencing resources' versions, so their version
headers cannot detect that cascade. Field-scoped patches avoid rewriting
unrelated cascaded fields; CMA and post-mutation response consistency checks
still govern fields the provider explicitly changes.

Version handling accepts positive integers. The repository mock and direct CMA
creation experiments returned an initial value of `1`; that observation is
modeled for service fidelity but is not a permanent schema promise. The
documented contract and raw DELETE observations are recorded in the
[taxonomy DELETE version note](../research/taxonomy-delete-version.md). Use
Contentful's established `version` and `sys.version` terminology.

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
