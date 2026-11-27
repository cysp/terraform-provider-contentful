# Signing secret values and acknowledgement

This contract applies to App and Webhook Signing Secret resources. Each accepts
exactly one sensitive argument: ordinary `value` from the effective Plan, or
write-only `value_wo` from Config. Contentful does not return the complete secret,
so the provider compares inputs with the last acknowledged value in Terraform
state. This comparison cannot detect external rotation or resolve an unconfirmed
write. The [API evidence](../research/signing-secret-versions.md) establishes no
usable optimistic-lock version for either endpoint.

Practitioner workflows belong in the [secrets and state guide](../guides/secrets-and-state.md).
[Planning rules](signing-secret-planning.md) explain how the provider schedules
updates while preserving Terraform's plan constraints.

## Comparison and representation

A known non-null prior `value` is the authoritative comparison value. When it is
null, compare against the private `value_wo` verifier. An unknown prior value is
an error. A missing verifier supplies no evidence of equality: an update that
permits a write must send PUT. An obsolete verifier must never override a known
ordinary value.

Equal bytes can still require a state change:

| Transition | State result | Remote action |
| --- | --- | --- |
| Ordinary to write-only | Remove plaintext and prepare a verifier | No PUT |
| Write-only to ordinary | Store plaintext and remove the verifier | No PUT |
| Unchanged write-only value | Keep the existing verifier and salt | No PUT |
| Changed value or missing comparison value | Prepare the new representation | PUT |

Clearing the verifier when ordinary value becomes active prevents it from
suppressing a later write. Refresh preserves the acknowledged comparison value;
it cannot infer equality from redacted text or the update timestamp. Create always writes,
including replacement and recreation after remote deletion, and prepares its own
state. The [lifecycle tests](../../internal/provider/resource_signing_secret_write_only_test.go)
cover representation changes and the resulting requests.

## Private verifier format

A write-only secret verifier lets the provider compare supplied bytes with the value represented in Terraform private state at one resource-relative attribute path. Each signing resource instance stores its verifier under the private key `write_only_secret_hashes`.
The private value is a JSON string containing the PHC-encoded hash. An absent
key supplies no comparison value. When comparison depends on private state,
the reader rejects an empty or malformed string, other JSON types, trailing
data, and verifier data larger than 1 MiB. The resource instance identifies
the singleton, so the stored value needs no resource ID or attribute name.

The verifier uses the memory-constrained profile from
[RFC 9106 §4](https://www.rfc-editor.org/rfc/rfc9106.html#section-4).
Each Argon2id v19 calculation uses 64 MiB of memory, three passes, four lanes,
a random 16-byte salt, and a 32-byte tag. The provider also compares secrets
during unchanged plans, so the workspace matters across many resources. Three
passes raise the cost of testing guesses against state when a caller supplies
a predictable, format-valid key. The input is the UTF-8 string
`value_wo`, one NUL byte, and the exact secret bytes. The reader verifies the
parameter encoding and decoded salt and output lengths before hashing.
A changed write-only value gets a fresh salt. Preparation precedes PUT so a hash
failure cannot follow a successful remote write.

Anyone with state access can test guesses offline. Use cryptographically
generated signing secrets and protect state, saved plans, and backups.

## Memory use

One semaphore per provider process admits one Argon2 calculation at a time
across both signing-secret resources and [Webhook definition credentials](mutation-reconciliation.md#write-only-webhook-credentials). Comparison and preparation acquire it
separately. An operation can be cancelled while waiting, and queue time counts
toward its deadline. An active calculation holds the slot until it returns and
then checks for cancellation. HTTP operations and state publication run outside
the slot.

The four lanes share the 64 MiB workspace. The Go runtime can retain allocations,
and separate provider processes each have their own slot. In a one-pass
signing-secret fixture with 50 secrets and Terraform 1.17.1 at parallelism 10,
serializing hashing reduced median peak provider RSS for saved rotation applies from
1,382 to 228 MiB. Median elapsed time increased from 1.068 to 1.602 seconds.
The [benchmark record at f6fc3264](https://github.com/cysp/terraform-provider-contentful/commit/f6fc32644f1771aa8305e3db0117d9be2193f3c5)
gives hardware, run counts, and the workload. RSS and elapsed time depend on
the workload and runtime.

## Invalid private state

A known ordinary value supplies the authoritative comparison without decoding
obsolete verifier contents. Successful updates remove those contents when keeping
ordinary storage, or replace them with a fresh verifier when migrating to
write-only storage. Matching bytes still require no PUT. Removal or replacement
is staged until state publication, so a failed mutation or encoding leaves the
prior private data intact.

Without a known ordinary value, Update rejects invalid verifier data before
mutation. Read and Delete do not inspect the verifier. During planning, invalid
verifier contents produce a warning and an unknown `updated_at` when comparison of a
configured write-only value depends on that verifier. Private-data access errors
and cancellation remain errors.

Terraform can plan an update before selecting replacement or destruction. The
planning warning lets Core reach those operations; it does not permit an
in-place update to proceed with invalid comparison data. An apply can therefore
fail after other resources have changed. Replacement establishes new state
without repairing the old verifier. The [recovery tests](../../internal/provider/resource_signing_secret_corrupt_verifier_test.go)
exercise rejected updates, replacement, and destruction through Terraform CLI.

## Mutation boundary

When writing a secret:

1. Prepare verifier changes before sending PUT.
2. Require a PUT response that has no transport error, decodes successfully,
   and identifies the requested resource. For an App Signing Secret, require
   `redactedValue` to match the final four characters of the submitted key.
   A mismatch contradicts the PUT. A matching suffix alone cannot verify the
   complete key. The Webhook Signing Secret response has an opaque redaction.
3. Encode the new resource state and identity.
4. Publish the verifier, then the encoded state and identity.

The provider returns no new state for an unacknowledged Create. For a failed
Update it returns the prior public state, identity, and verifier. Contentful may
already have accepted the submitted bytes. When Terraform persists the prior
baseline, reverting the configuration to it can produce a no-op while consumers
receive signatures made with a different key.

The provider can control only its RPC response. Killing a synthetic Framework
1.19.0 provider during Update on Terraform CLI 1.11.1 and 1.17.1 retained the
prior public attributes but lost the previous private data. With write-only
storage, losing the verifier leaves no comparison value, so a later apply can
issue another PUT even for unchanged bytes. In Core 1.11.1 an [RPC error](https://github.com/hashicorp/terraform/blob/v1.11.1/internal/plugin6/grpc_provider.go#L539-L546)
returns before the client receives private data. Core [falls back to the previous
public value](https://github.com/hashicorp/terraform/blob/v1.11.1/internal/terraform/node_resource_abstract_instance.go#L2392-L2397)
and [takes private data from the response](https://github.com/hashicorp/terraform/blob/v1.11.1/internal/terraform/node_resource_abstract_instance.go#L2575-L2582).
The experiment used a local backend. Check the state actually persisted after
a process or backend failure. A separate [Core identity fix](https://github.com/hashicorp/terraform/pull/37396)
addresses identity loss after apply errors in Terraform 1.12.0–1.12.2.

[Protocol tests](../../internal/provider/resource_signing_secret_protocol_test.go)
check the returned state, identity, and private data.
[Publication tests](../../internal/provider/write_only_secret_internal_test.go)
cover local encoding and verifier failures.
[Lost-response tests](../../internal/provider/resource_signing_secret_recovery_test.go)
check the attributes and private baseline saved by the CLI with a local backend,
then deliberately recover with a different secret.

App and webhook signing-secret PUT and DELETE requests retry explicit 429
responses with the same request body within the operation deadline. Transport
failures and ordinary 5xx responses stop without replay. Redirects are rejected
in both provider HTTP clients. GET follows the [shared retry policy](contentful-http-retry-policy.md).
This boundary does not provide at-most-once behavior across separate applies.
The provider redacts literal occurrences of the submitted bytes and any prior
ordinary value from upstream error details. Encodings, substrings, and secrets
unknown to the provider fall outside that policy. Core can quote HCL source in
diagnostics, so the resource examples reference variables instead of embedding
literal secrets.

## Ignored changes

Ordinary `value` follows the effective Plan. Whenever Update runs, `value_wo`
comes from apply-time Config and is compared normally, even if an ignore rule
excluded it during planning. Matching bytes skip PUT; different bytes or a
missing comparison value select PUT. A saved no-op plan does not call Update.
See [planning and ignored inputs](signing-secret-planning.md) for timestamp
constraints, migration conflicts, and the tests that preserve these boundaries.

## Terraform compatibility

Write-only signing secrets require Terraform 1.11.1 or later. Terraform 1.11.0
cannot reliably serialize sensitive ephemeral inputs; HashiCorp fixed that Core
bug in [1.11.1](https://github.com/hashicorp/terraform/pull/36619).

Terraform versions before 1.12 do not supply prior resource identity. When a
typed 404 removes the resource, Read returns null state and preserves a supplied
identity or reconstructs the missing identity from the scope attributes.

A provider predating `value_wo` cannot compare secrets using the verifier. Migrate
to ordinary `value` with the same acknowledged secret before downgrading; the
[practitioner guide](../guides/secrets-and-state.md#downgrading-the-provider)
explains the storage consequences. The [earlier signing-only compatibility check](https://github.com/cysp/terraform-provider-contentful/blob/f6fc32644f1771aa8305e3db0117d9be2193f3c5/docs/design/signing-secret-write-only.md#terraform-compatibility)
observed no PUT when migrating back before downgrading to v0.0.69, with or
without an initial refresh. Direct downgrade from write-only state selected PUT.
That result is specific to the builds and CLI used for that check.

The linked tests exercise the locally built provider and synthetic HTTP fixtures.
They establish provider behavior, not live Contentful conformance. Dated service
observations remain in the [API research](../research/signing-secret-versions.md).
