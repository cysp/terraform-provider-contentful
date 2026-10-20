# App Installation Secret parameters

This note defines response reconciliation for redacted App Installation parameters
and its ownership limitations. The
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

## Response contract

Retain the existing `parameters` schema and request semantics. Reconcile only an
installation's response, without a definition lookup, new configuration argument,
custom JSON equality rule, or provider-private state. The
[complete-response probe](../research/app-framework/secret-installation-parameters.md#indicators-in-the-installation-response)
found no embedded Secret classification or redaction indicator. A masked Secret and
a literal ordinary asterisk string had the same representation; mask matching alone
is therefore an inference, not a reliable type discriminator.

Create and Update use the effective Plan as the reference after a successful
response. Read uses prior State. Import has no prior plaintext to recover.
Configuration comparison keeps the existing normalized JSON semantics: masks are
not equivalent to plaintext when deciding whether a configured change needs a write.

Apply the following rule only to immediate properties of a JSON object:

| Response and reference | Result |
| --- | --- |
| Values are equal | Keep the response; no diagnostic |
| Response is a nonempty string containing only ASCII `*`; reference at the same key is a known, nonempty string containing at least one other character | Substitute that reference string; no diagnostic |
| Reference is missing, null, unknown, empty, non-string, or entirely asterisks | Keep the response |
| Response key is missing, or its value is empty, null, non-string, or readable text | Keep the response |
| Nested object or array contains asterisks | Keep the nested value without recursive substitution |

There is no mask-length comparison, UTF-16 calculation, or mismatch warning.
Published documentation does not specify the representation, and the
[client evidence](../research/app-framework/secret-installation-parameters.md#client-source-evidence)
supports opaque replay without establishing plaintext equality. Length is not a
supported failure condition.

Reconciliation changes only qualifying returned values. It never restores an
omitted object or key, removes response-added keys, or replaces the complete
response with the Plan. Other differences remain normal refresh drift or
post-write consistency failures. A failed apply does not imply remote rollback.
See [Terraform value semantics](terraform-value-semantics.md) and
[Framework Update][tf-update].

## Writes and ownership

An actual Create or Update sends the complete effective Plan, including values
retained by `ignore_changes` and imported all-asterisk values. Do not omit masks,
replace configured plaintext with masks, or add an extra GET before writing.
Omission can delete a secret; an extra read cannot remove the concurrent-write race.

An ordinary parameter update can reassert a configured secret after an invisible
external rotation. Terraform continues to own the complete effective parameter
object. A mask returned after an uncertain write never proves that write succeeded;
reconciliation adds no recovery read or retry behavior.

## Accepted limitations

Preserved state records the last intended value, not verified remote plaintext.
External Secret rotations of any length can remain invisible. An ordinary string
changed remotely to any nonempty all-asterisk string can also be hidden. These
limitations are accepted in exchange for retaining the existing interface and
avoiding failures based on undocumented mask-length assumptions.

Import keeps returned masks until a later operation supplies a usable planned
value. An imported mask does not establish previous Terraform ownership.
Forwarding a stale mask can store literal asterisks after a concurrent
length-changing rotation. A literal all-asterisk input equal to the current mask
can preserve the existing secret instead of replacing it. No delete-then-set
workaround is proposed.

This change does not make `parameters` write-only or alter sensitivity. Secrets
remain in Terraform state and require the protections described in the
[secrets guide](../guides/secrets-and-state.md).

## Implementation

The resource-specific helper in
[`app_installation_parameters_response.go`](../../internal/provider/app_installation_parameters_response.go)
takes the projected response and reference parameter values and returns the
reconciled value plus any JSON-processing diagnostics. Choosing Plan or State is
the caller's responsibility.

Create, Read, and Update call the helper after ordinary response projection and
before the existing identity/state publication boundary. Only successful mutations
reach reconciliation. Invalid JSON, unknown request values, and failed state
encoding remain errors. Reconciliation diagnostics do not include parameter values.

Raw JSON members preserve unrelated numeric tokens without converting through
`float64`. When no property qualifies for substitution, the helper retains the
projected response. Normalization preserves every other value, including number
tokens, null, empty collections, and nested structures.

## Verification

The resource-local HTTP fixture distinguishes stored values from redacted wire
responses using explicit Secret declarations and independent backing values. It
does not call the provider's reconciliation helper to construct responses or
expected results. The shared CMA mock is unchanged. Follow [Testing](../testing.md)
when extending this coverage.

| Concern | Required evidence |
| --- | --- |
| Successful writes and refresh | Plaintext and masked Create responses; masked Update; repeated refresh with no spurious plan; stable identity |
| Intended changes | Configured Secret rotation plans and sends the new plaintext; ordinary changes and removals remain visible |
| Reconciliation boundary | Same- and different-length masks; Unicode reference; ordinary literal asterisks; missing/added keys; empty/null/non-string values; nested values; no usable reference |
| JSON fidelity | Unrelated large integers, numeric forms, escaping, and object ordering survive reconciliation under existing JSON semantics |
| Import and effective Plan | Import retains masks; later configuration takes ownership; `ignore_changes` retains the effective value on a subsequent write |
| Ownership | Exact full-object request reasserts the intended secret after external rotation; Marketplace changes still write |
| Failures | Readable mutation contradictions remain errors; uncertain writes gain no success inference; no secret values in new diagnostics |

For lifecycle tests, assert exact requests and backing values separately from
Terraform state. The independent [API evidence](../research/app-framework/secret-installation-parameters.md)
supplies the observed server cases. A deliberately different-length response tests
the provider's compatibility policy; it does not claim Contentful emitted that
representation in the live experiments.

The resource documentation and secrets guide describe these ownership and
reconciliation limits for practitioners. Changes require the focused behavior,
formatting, lint, and generation checks in
[Development](../../DEVELOPMENT.md#validation-scope).

[schema]: https://github.com/cysp/terraform-provider-contentful/blob/8db2a7bc70666b6c1b1b7011c2e688373eb8ed2b/internal/provider/resource_app_installation_schema.go
[request]: https://github.com/cysp/terraform-provider-contentful/blob/8db2a7bc70666b6c1b1b7011c2e688373eb8ed2b/internal/provider/app_installation_model_request.go
[response]: https://github.com/cysp/terraform-provider-contentful/blob/8db2a7bc70666b6c1b1b7011c2e688373eb8ed2b/internal/provider/app_installation_model_response.go
[lifecycle]: https://github.com/cysp/terraform-provider-contentful/blob/8db2a7bc70666b6c1b1b7011c2e688373eb8ed2b/internal/provider/resource_app_installation.go
[mock]: https://github.com/cysp/terraform-provider-contentful/blob/8db2a7bc70666b6c1b1b7011c2e688373eb8ed2b/internal/contentful-management-go/testing/handler_app_installation.go
[tf-update]: https://developer.hashicorp.com/terraform/plugin/framework/resources/update#caveats
