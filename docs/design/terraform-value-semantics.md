# Terraform value semantics

This contract is for maintainers changing schemas, planning, request conversion,
response projection, or state publication. Use the [design map](README.md) for
resource-specific contracts. `Config`, `Plan`, and `State` refer to configuration,
the effective plan after lifecycle processing, and prior state. A response-owned
value can come from Contentful only where the effective Plan permits it.

## Value states

Keep Terraform values intact until conversion: inspect containers and their
children before using primitive accessors, and retain attribute paths in errors.
Do not turn known empty values into null. Each API member needs an explicit policy
for omission, null, and empty values; the [Framework value guarantees](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values#when-can-a-value-be-unknown-or-null)
supply the lifecycle context.

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

### JSON normalization

Opaque JSON normalization uses Go's `encoding/json` behavior throughout the
provider. It validates one complete document, orders object keys, retains array
order, escapes HTML and control characters in keys and values, and preserves
number tokens with `UseNumber` so large integers, decimal precision, exponent
notation, and negative zero are not rounded through floating-point conversion.

Invalid UTF-8 and unpaired UTF-16 surrogates become the Unicode replacement
character. A valid surrogate pair following an unpaired surrogate is preserved:
`"\ud800\ud800\udc00"` normalizes to `"�𐀀"`. Object keys that become equal after
decoding follow the same last-value rule as duplicate keys. A malformed key
and an explicit replacement-character key can collide; the last value in
document order wins.

App Installation and Extension parameters, Role policy constraints, and Editor
Interface settings use the checked normalization operation and report failures
at their attribute paths. Missing response members remain Terraform null; an
explicit JSON `null` remains a known JSON string. The existing fallback
normalizers retain the original input on failure for callers that own their
validation or preserve opaque values.

## Lifecycle ownership and plan consistency

After Create or Update, state must remain consistent with every known effective
Plan value. Post-mutation state construction starts with the complete response
projection, compares each constrained API-backed value against it, and restores
the exact Plan representation only after every comparison is semantically
equivalent. If Contentful meaningfully contradicts the Plan, the provider emits
a consistency error and retains complete, truthful response-derived recovery
state; it never hides the contradiction by copying Plan values over it. Unknown
values come from the response. Projection warnings are retained. A later Read
skips this reconciliation and projects the remote
representation so meaningful remote drift remains visible.

### Mutation endpoint identity

Mutation recovery retains requested path components in state, structured identity,
and the legacy multipart ID. Server-generated IDs remain response-owned: a known
Plan value constrains consistency but does not select a POST target. An explicitly
supplied App Action creation ID or validated App Key `jwk.kid` fingerprint also
selects the recovery target, even though it travels in the request body. Singleton body IDs, Environment Alias targets, and
related-resource references do not select the addressed endpoint; their value
reconciliation follows the resource-specific contract. The requested environment
or alias remains the address; the provider does not infer alias resolution from
a contradictory response.

A version belongs to the response's identified object. An identity contradiction
therefore clears the existing private `version` after state/identity encoding
succeeds; neither the returned version nor the stale pre-mutation version may
authorize another write. A later successful refresh must confirm the
requested identity before restoring version authority. Nonidentity Plan
contradictions retain the confirmed endpoint's returned version. Entry publication
and Content Type activation authority is also revoked after an identity contradiction; no
follow-up publish or activate request uses that response's version. No path
repairs a conflict by reading and retrying a newer version.

The [local publication contract](#diagnostics-and-local-publication) defines the
state/identity encoding boundary and subsequent private-data writes. These local
guarantees do not make a remote mutation transactional.

App Signing Secret mutation responses must identify the requested singleton
before acknowledging the planned secret. A contradiction leaves prior secret
state intact on Update and publishes no new secret state on Create. The
[Webhook Signing Secret](webhook-signing-secret.md#failures-and-confidentiality)
has its own unconfirmed-write acknowledgement contract.

For Preview Environment, Config omission selects generated-ID creation without
discarding known Plan constraints. See the
[preview environment design](content-preview-environments.md#concurrency-and-recovery)
for its state-to-plan configuration delta, and the
[API research](../research/content-preview-environments.md#concurrency-and-errors)
for observed version-locking limits.

Known effective Plan values constrain apply results, including defaults,
computed values retained by plan modifiers, and values retained by `ignore_changes`.
Only unknown planned values can take an unconstrained response value. Endpoint
identity and provider-local `timeouts` remain outside the API-value overlay:
identity retains the addressed target; timeouts retain the effective Plan.

Terraform 1.15.8 with Framework v1.19.0 was exercised with provider comparisons
removed. Core rejected contradictory defaults and `UseStateForUnknown` values
for Webhook `active`, `headers`, `headers[*].secret`, and Editor Interface
`sidebar[*].disabled`. A contradictory Create retained returned values in tainted
recovery state. Recheck these observations when upgrading Core or Framework.

Resource-specific comparison rules are in [mutation reconciliation](mutation-reconciliation.md).

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

Pending Entry publication and Content Type activation have a stricter,
positive-version authority contract; see [content publication](content-publication.md).
All requests use the [HTTP retry policy](contentful-http-retry-policy.md).

### Delivery API key environments

The `environments` attribute on `contentful_delivery_api_key` has a deliberate
forward-compatibility policy:

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
[Delivery API key environments note](../research/delivery-api-keys.md).

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
recorded in the [Space Enablements request values note](../research/space-enablements.md).

## Read identity

Managed Read compares response identity with the state attributes used to address
its GET before publishing response state, ResourceIdentity, or changes to private
versions and pending lifecycle authority. A contradictory response produces an
attribute error and leaves prior state intact. App Action and App Event Subscription
perform this check before publishing their endpoint-preserving projections;
Webhook Signing Secret retains its existing explicit check.

The comparison uses addressing attributes rather than requiring an existing
ResourceIdentity, so it also applies when reading legacy state and immediately
after import. An import with a contradictory Read fails without recording the
resource as managed. The requested environment or alias remains the address;
related-resource references and singleton body IDs are not endpoint identity.
Matching responses still expose ordinary remote drift and retain established
redacted-value and narrow canonicalization rules. A supported NotFound response
still removes the resource.

Read rejects a response from an unconfirmed target. This differs from mutation
recovery, which may need to retain the addressed target and returned values to
record a completed remote effect. Neither policy establishes that a successful
response is a transactionally stable snapshot of Contentful.

## Diagnostics and local publication

Warnings from a representable response projection accompany the published
result. Encoding errors must not publish half of a state/identity pair: both
values are encoded into staged copies, and both are assigned only after both
encodings succeed. This is local publication only; it does not make remote API
operations or private provider data transactional.

For versioned responses, the provider records or revokes the private `version`
under the [endpoint identity policy](#mutation-endpoint-identity) only after
publishing the matching state and identity. An encoding error leaves the prior
private version unchanged. A later private-data write error is reported without
rolling back the published state and identity. Pending lifecycle authority and
resource-specific missing-version handling remain separate operations.

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
