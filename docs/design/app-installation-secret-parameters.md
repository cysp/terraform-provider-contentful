# App Installation Secret parameters

This note records the provider's response-handling limitation and the design choices
for addressing it. No new schema or reconciliation policy has been adopted. The
[API evidence](../research/app-framework/secret-installation-parameters.md) is the
authority for observed Contentful behavior; the
[secrets guide](../guides/secrets-and-state.md#app-and-extension-parameters) describes
practitioner-facing limitations.

## Provider behavior and reproduction

At provider revision `8db2a7bc70666b6c1b1b7011c2e688373eb8ed2b`, `parameters` is one
Optional normalized JSON string. Request conversion sends the complete object;
Create, Read, and Update project the returned object without reconciling Secret
masks with planned or prior values. The mock returns stored values unchanged, so
ordinary-parameter tests do not reproduce PAT redaction.

Sources at that revision: [schema][schema], [request conversion][request],
[response projection][response], [lifecycle][lifecycle], and [mock handler][mock].

A live CLI reproduction used Terraform 1.16.0 on darwin_arm64, Framework v1.19.0,
and jsontypes v0.2.0. The disposable definition declared optional `Secret` and `Symbol`
parameters. The installation configured both using `jsonencode(...)`, with a synthetic
Secret value of `synthetic-terraform-alpha`.

| Operation | Observed result |
| --- | --- |
| Apply with installation absent | Succeeded; state contained the supplied plaintext |
| Plan after creation | Proposed replacing 25 asterisks with the configured value |
| Import an installation containing that value | Succeeded; state contained the mask |
| Apply after import | Failed with `Provider produced inconsistent result after apply`: returned `.parameters` differed from the known planned string |
| Plan again | Proposed the same update |
| Set `ignore_changes = [parameters]`, then change the configured Secret | Both plans reported no changes; deliberate rotation was suppressed |

The first refresh result was observed in the plan; an ordinary plan does not itself
persist that refreshed snapshot to the original state file. The inconsistent-result
error occurs after the remote write, so an apply failure does not imply rollback.
These results describe the pinned provider revision, not every future release.

## Design constraints

The design should avoid fetching another remote resource merely to reconcile an
installation response. The [complete-response probe](../research/app-framework/secret-installation-parameters.md#indicators-in-the-installation-response)
found no embedded Secret classification or redaction indicator. A masked Secret and
a literal ordinary asterisk string had the same representation; mask matching alone
is therefore an inference, not a reliable type discriminator.

A PAT-based design can manage the desired Secret value without verifying its remote
equality. Contentful returns the same mask for different strings of the same length.
Preserving a prior value would therefore mean preserving the last intended value,
not confirming that no external rotation occurred. Missing keys and different-length
masks still provide observable evidence; neither should be indiscriminately hidden.

After a successful write, known ordinary attributes must match the effective Plan,
including lifecycle processing such as `ignore_changes`. A supported redaction may
justify narrow reconciliation; it does not justify copying the whole Plan over a
contradictory response. See [Terraform value semantics](terraform-value-semantics.md)
and [Framework Update][tf-update].

Contentful's Secret declaration and Terraform's sensitivity are independent.
`Sensitive` masks normal output but retains values in state. Write-only arguments,
available from Terraform 1.11, keep values out of plan/state but cannot themselves
trigger an update; a separate revision or equivalent trigger is needed.
[Attribute sensitivity][tf-sensitive], [write-only arguments][tf-write-only].

Import cannot recover plaintext. An observed mask establishes key presence, not
prior Terraform ownership. Replaying it is not a safe general preservation operation:
a concurrent length-changing rotation can turn it into a literal secret. Nor can a
caller reliably set literal asterisks equal to the existing mask with that PUT.
These are API ambiguities that every proposed interface must address.

## Alternatives

The following are alternatives, not an instruction to add multiple configuration
modes. Argument names illustrate possible interfaces and are not reserved contracts.

| Approach | Strongest case for | Strongest case against |
| --- | --- | --- |
| Reconcile Secret values using current AppDefinition metadata within `parameters` | Preserves existing HCL and ordinary drift; uses Contentful's authoritative classification | Adds metadata-access requirements and non-atomic definition/read dependencies |
| Explicit Secret key set alongside `parameters` | Small additive change; no mandatory extra lookup; callers can derive keys from a definition | Practitioner assertions can diverge from Contentful; sensitivity and import classification still need handling |
| Sensitive `secret_parameters` map alongside ordinary `parameters` | Clear input ownership, output masking, and normal configuration-driven rotation | Migration and overlapping-key rules; duplicates classification; import Read has no configuration with which to split unknown keys; secrets remain in state |
| Write-only Secret map plus revision | Supports ephemeral inputs without persisting values | Secret changes alone do not plan writes; ordinary updates must still assemble the complete replacement object |
| Whole parameter object as one-way input | Simple opaque-input model without per-key classification | Loses ordinary drift and readable outputs; a write-only form needs a trigger, while blindly retaining a normal Plan could conceal rejected values |
| Read using App Identity | Actual equality and external rotation become observable | Requires app-specific credentials and token exchange; unavailable to many shared-app consumers |
| Delegate the whole object with `ignore_changes` or external management | Existing mechanism with no provider change | Suppresses ordinary changes and intentional rotations; another resource update may still submit parameters |

Splitting secret and ordinary inputs need not split resource ownership: they would
still have to form one PUT. Two independent resources managing halves of the same
replacement object would overwrite each other. A write-only design must also say
whether ordinary updates reassert configured secrets even when their revision stays
unchanged; placeholder replay does not remove that decision.

Inferring Secret keys from asterisks alone is unsound: an ordinary parameter can
legitimately contain the same string. A custom JSON equality rule cannot recover
classification or hidden values. Local hashes can detect local input changes, but
cannot verify remote equality. Removing a Secret declaration is not an equivalent
fix: the experiment exposed existing plaintext to PAT reads.

## Recommendation and unresolved decisions

Prefer explicit Secret classification while retaining the existing `parameters`
object if the objective is a compatible repair without an extra resource read. An
optional set such as `secret_parameter_ids` is the smallest additive interface;
callers already managing a definition can derive it from that configuration without
a provider-side lookup. This remains a proposal, not an adopted argument.

Its limitation must be explicit: the caller asserts that the named keys are Secret.
If that assertion diverges from Contentful, a mask-like ordinary value may be
misinterpreted. A separate sensitive map is a stronger interface alternative when
clearer secret inputs and output masking justify migration, but it does not eliminate
import classification. Write-only support addresses the additional requirement to
exclude secrets from state and needs its own update-trigger contract.

Metadata-guided reconciliation remains an alternative if the extra dependency is
later accepted, rather than the default recommendation. The
[pinned SDK definition lookup][definition-adapter] is organization-scoped;
permission parity with installation management has not been established. The
provider's top-level Marketplace lookup has a separate
[evidence gap](../research/app-framework/README.md#unresolved-behavior-and-source-disagreements).
Neither that lookup nor inference from asterisks should be introduced implicitly.

Before choosing a contract, resolve:

- Which response representations permit reconciliation with effective Plan or prior
  State, including the observed UTF-16 mask lengths, readable contradictions, empty
  strings, key removal, and definition changes.
- How import and previously masked state retain key-presence evidence without
  claiming recovered plaintext, and when later writes require actual values.
- How ambiguous all-asterisk input is reported without silently preserving an
  unintended secret or introducing a non-atomic delete-then-set workaround.
- Which operations send the complete parameter object, including local-only changes
  and uncertain writes. A masked GET cannot prove that an uncertain write succeeded.

The API evidence provides independent expected cases for assessing a future change.
Mock storage and PAT projection must remain distinguishable so that tests can expose
the original mismatch rather than derive their expected behavior from the proposed
reconciliation itself. Any adopted behavior belongs in the resource contract and
practitioner documentation; this discussion should then be reconciled with that decision.

[schema]: https://github.com/cysp/terraform-provider-contentful/blob/8db2a7bc70666b6c1b1b7011c2e688373eb8ed2b/internal/provider/resource_app_installation_schema.go
[request]: https://github.com/cysp/terraform-provider-contentful/blob/8db2a7bc70666b6c1b1b7011c2e688373eb8ed2b/internal/provider/app_installation_model_request.go
[response]: https://github.com/cysp/terraform-provider-contentful/blob/8db2a7bc70666b6c1b1b7011c2e688373eb8ed2b/internal/provider/app_installation_model_response.go
[lifecycle]: https://github.com/cysp/terraform-provider-contentful/blob/8db2a7bc70666b6c1b1b7011c2e688373eb8ed2b/internal/provider/resource_app_installation.go
[mock]: https://github.com/cysp/terraform-provider-contentful/blob/8db2a7bc70666b6c1b1b7011c2e688373eb8ed2b/internal/contentful-management-go/testing/handler_app_installation.go
[definition-adapter]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/app-definition.ts
[tf-update]: https://developer.hashicorp.com/terraform/plugin/framework/resources/update#caveats
[tf-sensitive]: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/dynamic#sensitive
[tf-write-only]: https://developer.hashicorp.com/terraform/plugin/framework/resources/write-only-arguments
