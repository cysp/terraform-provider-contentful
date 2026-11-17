# Taxonomy values and version locking

Use the shared [value and identity rules](terraform-value-semantics.md) alongside
these resource-specific contracts.

## Taxonomy Optional+Computed collection ownership

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

## Taxonomy response canonicalization

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

## Taxonomy optimistic version locking

Taxonomy resources record Contentful `sys.version` in provider-private state,
matching the provider's other versioned resources. Update sends the prior
private-state value supplied at apply. Delete also sends that value when it is
present. A remote change for which CMA advances the resource's `sys.version` is
therefore rejected instead of being overwritten. Create, Read, and Update store
the integer version returned for a confirmed endpoint, including alongside
recovery state when configuration values differ from the plan. Identity
contradictions follow the [shared version policy](terraform-value-semantics.md#mutation-endpoint-identity).

Terraform does not supply prior private data to Delete while replacing a
tainted resource. This applies both to recovery state published alongside an
errored Create and to a resource manually marked tainted after successful
updates. Only when private version data is genuinely absent, Delete performs
one GET for the requested resource and validates its organization and resource
identity before using its `sys.version` in the DELETE request. A contradictory
response stops deletion and retains state for a subsequent confirmed read. A 404
from this GET means the resource is already absent. This narrow
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
[taxonomy version note](../research/taxonomy.md). Use Contentful's
established `version` and `sys.version` terminology.
