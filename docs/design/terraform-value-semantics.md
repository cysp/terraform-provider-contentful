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

Preview Environment mutation recovery has the same targeted endpoint-identity
exception. When a successful response contradicts the requested space or
preview environment ID, Terraform retains the requested endpoint identity and
legacy multipart ID while checkpointing every other response-derived value and
the returned `sys.version` before reporting the consistency error. For
generated-ID creation, the preview environment ID remains response-owned.

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
