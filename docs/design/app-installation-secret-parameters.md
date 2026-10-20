# App Installation Secret parameters

The provider reconciles redacted App Installation responses with the effective
Plan after successful writes and with prior State during refresh. This preserves
Terraform's parameter values when Contentful withholds plaintext, but cannot
verify the remote secret. Each write owns the complete planned parameter object.

The [API research](../research/app-framework/secret-installation-parameters.md)
records published contracts and scoped observations. The
[secrets guide](../guides/secrets-and-state.md#app-installation-secrets) explains
configuration, import, and rotation for practitioners.

## Rationale

The [installation response probe](../research/app-framework/secret-installation-parameters.md#indicators-in-the-installation-response)
found no embedded Secret classification or redaction indicator. A masked Secret
and a literal ordinary asterisk string had the same representation. Treating an
all-asterisk response as redaction is therefore an inference, with the limitations
below.

Without reconciliation, redacted responses can cause repeated plans and
post-write consistency errors. A [pinned Terraform reproduction][reproduction]
records these failures at provider revision
`8db2a7bc70666b6c1b1b7011c2e688373eb8ed2b`.

## Response contract

`parameters` is an optional normalized JSON string. Reconciliation applies only
to the installation response. It requires no AppDefinition lookup, separate
Secret configuration, or provider-private state. Configuration comparison uses
normalized JSON semantics: a mask is not equivalent to plaintext when deciding
whether a configured change needs a write.

Create and Update use the effective Plan as the reference after a successful
response. Read uses prior State. Import has no prior plaintext to recover.

Apply the following rule only to immediate properties of a JSON object:

| Response and reference | Result |
| --- | --- |
| Values are equal | Keep the response; no diagnostic |
| Response is a nonempty string containing only ASCII `*`; reference at the same key is a known, nonempty string containing at least one other character | Substitute that reference string; no diagnostic |
| Reference is missing, null, unknown, empty, non-string, or entirely asterisks | Keep the response |
| Response key is missing, or its value is empty, null, non-string, or readable text | Keep the response |
| Nested object or array contains asterisks | Keep the nested value without recursive substitution |

The provider does not compare mask lengths because Contentful does not document
them as an equality guarantee. The
[client evidence](../research/app-framework/secret-installation-parameters.md#client-source-evidence)
supports replay of redacted values without establishing plaintext equality.
There is no UTF-16 calculation or warning for a different-length response.

Reconciliation changes only qualifying returned values. It never restores an
omitted object or key, removes response-added keys, or replaces the complete
response with the Plan. Other differences remain normal refresh drift or
post-write consistency failures. A failed apply does not imply remote rollback.
See [Terraform value semantics](terraform-value-semantics.md).

## Writes and ownership

Each Create or Update that issues a PUT sends the complete `parameters` value
from the effective Plan, including values retained by `ignore_changes` and
imported all-asterisk values. It does not omit masks, replace plaintext with masks,
or add a GET before writing. Omitting a key can delete it; reading before writing
cannot make the operation atomic.

Update compares effective Plan with prior State. When parameters are equivalent
under normalized JSON semantics and Marketplace acknowledgments are unchanged,
it publishes planned timeouts without issuing a PUT. Timeout and request
validation still run. A parameter or Marketplace change still writes;
Marketplace acknowledgments affect the outgoing header.

An ordinary parameter update can reassert a configured secret after an invisible
external rotation. A mask returned after an uncertain write never proves that
write succeeded; reconciliation provides no recovery read or retry behavior.

## Limitations

A value preserved from Plan or State does not verify current remote plaintext.
External Secret rotations of any length can remain invisible. An ordinary string
changed remotely to a nonempty all-asterisk string can also be hidden. The policy
accepts these ambiguities to support redacted responses through the existing
parameter interface without relying on undocumented mask lengths.

Import records returned masks without recovering plaintext. A later write can
supply the value Terraform should manage. An imported mask does not establish
previous Terraform ownership. The server's observed preservation rule is separate from
the provider's response rule: sending the exact current mask can preserve a
secret, while a stale mask can store literal asterisks after a length-changing
rotation. A literal all-asterisk input equal to the current mask can preserve the
existing secret instead of replacing it. See the
[placeholder observations](../research/app-framework/secret-installation-parameters.md#replacement-deletion-and-placeholders).

`parameters` is not write-only or automatically sensitive. Secret values can be
stored in Terraform state and saved plans; see the
[secrets guide](../guides/secrets-and-state.md).

## Implementation boundaries

The [response helper](../../internal/provider/app_installation_parameters_response.go)
must preserve unrelated JSON values, including numeric tokens, nulls, empty
collections, and nested structures. A response with no qualifying substitution
is retained unchanged.

The [resource lifecycle](../../internal/provider/resource_app_installation.go)
selects Plan or State as the reference and reconciles after response projection,
before publishing identity and state. Only successful mutations reach
reconciliation. Invalid JSON, unknown request values, and failed state encoding
remain errors. Reconciliation diagnostics must not include parameter values.

## Verification

[Response tests](../../internal/provider/app_installation_parameters_response_internal_test.go)
exercise the matching boundaries and JSON fidelity.
[Lifecycle tests](../../internal/provider/resource_app_installation_secrets_test.go)
and the [failed Update test](../../internal/provider/resource_app_installation_secrets_internal_test.go)
check state and requests against independent backing values.
[Timeout lifecycle tests](../../internal/provider/resource_app_installation_timeout_test.go)
and [validation tests](../../internal/provider/resource_app_installation_timeout_internal_test.go)
check which changes issue a PUT and that validation runs before skipping it.

The [shared test server](../../internal/contentful-management-go/testing/app_installation.go)
separates stored secrets from PAT responses and models the observed mask replay
rule. The [HTTP fixture](../../internal/provider/resource_app_installation_secrets_fixture_test.go)
records requests and injects deliberate response contradictions or failures. Expected values must remain
independent of the reconciliation helper. Exact requests and backing values
establish ownership behavior separately from Terraform state assertions.
A deliberately different-length response tests provider policy; it does not
claim that Contentful emitted that representation in the live experiments.

Follow [Testing](../testing.md) and the
[validation scope](../../DEVELOPMENT.md#validation-scope) when changing this contract.

[reproduction]: https://github.com/cysp/terraform-provider-contentful/blob/741b9b316a28b4491988addc07b0d6a4027d25a9/docs/design/app-installation-secret-parameters.md#provider-behavior-and-reproduction
