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

Content Type publication metadata is observation, not activation authority.
`contentful_content_type` activates only the exact draft returned by a
successful Create or modeled Update, during that same resource operation.
Read, import, refresh, legacy state, external deactivation, and external drafts
never cause activation by themselves.

A modeled Update uses the exact current Contentful version and the effective
Terraform plan. This includes refreshed values protected by `ignore_changes`.
After Contentful accepts the draft, the provider checkpoints its returned state
and version before activating that exact version. If activation fails, the
truthful draft state remains, but a later unchanged operation does not activate
it. The practitioner must activate it manually or make another
Terraform-managed Content Type change after resolving the error.

During Create, the provider likewise checkpoints the returned draft before
returning an activation error. Terraform Core taints a resource whose Create
returns an error, so the next normal plan replaces it.

No activation path fetches or retries a newer draft. Concurrent draft changes
therefore fail Contentful's optimistic-concurrency check instead of publishing
another actor's version.

A nominally successful activation response is accepted only when its
`sys.publishedVersion` equals the exact version sent in the activation request
and its returned `sys.version` has the active one-version-newer relationship.
Likewise, a successful draft PUT must return the exact expected revision
(version 1 on Create, or the requested prior version plus one on Update) and an
unpublished or pending-draft lifecycle tuple before that version can be used as
the activation lock token.
The provider checkpoints the complete returned response before reporting a
contradiction, so recovery state remains truthful even when apply fails.

`published_version` did not exist in state written by older provider versions.
A normal post-upgrade refresh projects `sys.publishedVersion`. With
`-refresh=false`, Terraform decodes the missing legacy Computed value as null,
but publication remains observational and the unchanged transition is a no-op.

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
every HTTP method. Neither source establishes whether a mutation committed, so
mutation retries retain accepted residual ambiguity. The evidence, deadline,
and backoff contracts are recorded in
[Contentful HTTP retry policy](contentful-http-retry-policy.md).

### Entry publication recovery and partial field ownership

`contentful_entry` owns publication only for a draft written by that resource
operation. After an exact, plan-consistent Update draft response has been
projected into the Update response, the provider stores that draft version as an
optional pending publication marker in resource private state immediately before
publishing it. If Terraform persists the returned resource state and resource
private state after Update publication fails, an unchanged later apply publishes
exactly that version without repeating the draft PUT. This is not a durable
mid-RPC checkpoint: process or IPC loss before Terraform persists the response
can lose the marker. A typed Publish response clears the marker, including when
its lifecycle tuple is contradictory and the provider returns an error.

Create checkpoints the same truthful draft state and version in resource private
state before publishing, but does not store the Update pending publication
marker. If Create publication fails, Terraform taints the resource and the next
normal plan replaces it; the checkpoint remains available for cleanup and does
not authorize in-place publication recovery.

The pending publication marker is deliberately not inferred from remote
lifecycle metadata. An external pending draft, an imported unpublished entry,
and legacy state without the marker are not published merely because
`sys.publishedVersion` is absent or older than `sys.version`. Recovery remains
authorized only while the observed current version equals the pending
publication version and `publishedVersion` is absent or strictly older than that
version. Read clears the marker when the current version differs or when
`publishedVersion` is equal to or newer than the pending publication version.
Observational tolerance of unusual positive lifecycle tuples never broadens
mutation authority. No publication path fetches and retries another actor's
newer version.

A sanitized live whole-Entry unpublish returned a newer `version` with no
`publishedVersion`, and a subsequent GET preserved that tuple. Read therefore
revokes any older pending publication authority through the same exact-version
boundary; it does not treat absence of publication alone as renewed authority.

An ambiguous draft-mutation failure does not authorize publication. The client
does not replay the mutation automatically, and the provider does not infer draft
ownership from a later GET whose fields happen to match the plan. If the request
committed but its response was lost, refresh may therefore expose a matching
unpublished draft without a pending publication marker. This fail-closed
limitation avoids publishing an indistinguishable concurrent editor version.

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

Entry metadata tags and taxonomy concepts retain their public List schema, while
the provider compares them without regard to order while preserving duplicate
multiplicity. A configuration-only reorder is therefore a representation-only
Update: Terraform stores the desired list representation without writing or
publishing an Entry.

Contentful may apply Content Type default values to fields omitted during Entry
creation. A Create response may therefore add response-only field keys,
including a value for a Terraform-null key omitted from the request, but every
sent field must still be present and semantically equal after the recognized
empty-array and raw-JSON-null omission fallbacks above. Update uses full-body
replacement and Contentful does not apply defaults to Entry updates, so an
Update response must have exactly the sent field keys and values after those
same fallbacks. A Create publication response may include creation defaults only
when it returns the complete normal tuple:
`publishedVersion` equal to the version sent and `version` exactly one greater.
Update and recovery publication responses must have the exact effective-plan
field keys. Every response with a missing or different `publishedVersion`, or with
a lower or higher `version`, must also have exact field keys so that unrelated
response additions are not silently adopted.

After those checks succeed, mutation state restores the exact effective-plan
`fields`, `metadata`, and `timeouts`; identity and lifecycle values remain
response-owned where the plan does not already establish an endpoint identity.
A later Read remains authoritative and can expose Contentful-added defaults,
which `ignore_changes` can preserve through a subsequent managed update. When
Contentful contradicts a planned field, metadata, content type, or endpoint
identity, the provider reports an error and checkpoints the representable
response `fields` and `metadata`, while retaining the requested immutable content
type and endpoint identity as the recovery target. An anomalously high current
version after Publish with an otherwise coherent publication tuple remains a
warning, but its lifecycle state and exact-key response checkpoint are preserved
rather than described as a complete returned Entry.

An Update draft response must return the exact expected next version before the
provider grants publication authority. Create has no prior version from which
to prove an increment and accepts any positive, plan-consistent returned draft
version. Either draft response may omit `publishedVersion` or contain a positive
older publication (`publishedVersion < version`); equal, future, nonpositive, or
unknown publication values are contradictory. Only an Update draft version
becomes a pending publication marker.

Publish must report the submitted version as `publishedVersion`. The repeatedly
observed normal response has `version` equal to the submitted version plus one.
Any other positive current version is representable, checkpointed, and warned
about rather than rejected solely for its arithmetic. That observational
tolerance does not broaden field ownership: Create publication may accept
response-only creation defaults only for the complete normal tuple, while an
anomalous Create publication and every Update or recovery publication require
exact effective-plan fields. Missing or different `publishedVersion` and
nonpositive current versions remain errors. Read likewise preserves positive,
representable lifecycle tuples and warns about unusual ordering instead of
treating undocumented arithmetic as permanently invalid. The computed
`published_version` attribute exposes the authoritative CMA value; `version` in
resource private state always records the same response version as the published
Terraform state.

### Provider-private optimistic-lock barrier

For resources whose Contentful mutations require a version stored only in
provider-private state, inability to read or decode that version is a terminal
apply diagnostic. The provider must return before issuing any Contentful
mutation; a missing or malformed lock token must never degrade to version zero
or an unlocked request. The one deliberate exception is the taxonomy Delete
fallback below, where Terraform can omit private data for a tainted replacement
and the provider obtains a positive current version with a GET before deleting.

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
