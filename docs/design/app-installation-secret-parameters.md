# App Installation Secret parameters

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

The Terraform `parameters` map holds one JSON-encoded value per parameter ID.
The CMA exchanges those values as members of a single JSON object.

Create and Update use the effective Plan as the reference after a successful
response. Read uses prior State.

When a returned top-level parameter is a nonempty string of ASCII `*` characters,
use the reference value at the same key if it is a known string containing at
least one character other than `*`. Keep every other returned value as received.
The resulting map contains exactly the keys Contentful returned.

Mask length gives no documented guarantee of plaintext equality. The
[client evidence](../research/app-framework/secret-installation-parameters.md#client-source-evidence)
supports replay of redacted values without establishing plaintext equality.
Read records other differences as drift. After a successful PUT with representable
parameters, the provider saves the map after mask reconciliation and checks it
against the plan. A mismatch produces a resource-level error naming `parameters`.
This avoids including either parameter values or their configuration expression
in the diagnostic. See [Terraform value semantics](terraform-value-semantics.md).

## Writes and ownership

A PUT with a planned map sends every key, including values retained by
`ignore_changes` and imported masks. Omitting a key can delete it. A null
map omits the `parameters` member and can clear existing values.

Update compares effective Plan with prior State. When parameters are equivalent
under normalized JSON semantics and Marketplace acknowledgments are unchanged,
it publishes planned timeouts without issuing a PUT. Timeout and request
validation still run. An empty acknowledgment set omits the
`X-Contentful-Marketplace` header. Import and state upgrade both record an
empty set when acknowledgments are absent. With equivalent configuration,
upgrading the parameter object and applying this default requires no PUT,
even with refresh disabled.

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

## Verification

The [HTTP fixture](../../internal/provider/resource_app_installation_secrets_fixture_test.go)
records the complete PUT and can override the response independently of the
stored parameters. Check both when changing mask handling. A synthetic mask
of a different length tests the provider’s inference rule. For observed service
behavior, use the [API research](../research/app-framework/secret-installation-parameters.md).

[reproduction]: https://github.com/cysp/terraform-provider-contentful/blob/741b9b316a28b4491988addc07b0d6a4027d25a9/docs/design/app-installation-secret-parameters.md#provider-behavior-and-reproduction
