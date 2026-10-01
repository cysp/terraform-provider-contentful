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
it cannot infer equality from redacted text or timestamps. Create always writes,
including replacement and recreation after remote deletion, and prepares its own
state. The [lifecycle tests](../../internal/provider/resource_signing_secret_write_only_test.go)
cover representation changes and the resulting requests.

## Private verifier format

Private key `write_only_secret_hashes` contains a JSON array of `{path, hash}`
records. The shared helper updates or removes each path independently, preserving
other records. Signing-secret resources accept only the `value_wo` path.

The shared format and dynamic-path operations retain compatibility with the
captured `header_values_wo` encoding from [PR #542 at revision 1d50a229](https://github.com/cysp/terraform-provider-contentful/blob/1d50a229f30d8c85abf163e2c68e1809a3bcad3c/internal/provider/write_only_secret.go).
That encoding supports removing one header while preserving other headers' verifiers.
The [captured fixture](../../internal/provider/testdata/write_only_secret_pr542.json)
and [verifier tests](../../internal/provider/write_only_secret_internal_test.go)
define the supported encoding and path-removal behavior. This is a shared-helper
compatibility commitment to that captured encoding and those path operations,
not to future revisions of the proposal. It does not add webhook-header support
or permit header paths in signing-secret private state. Future encoding or
parameter changes require an explicit compatibility and migration decision.

Each hash uses PHC encoding with Argon2id v19, 65536 KiB memory, one iteration,
four lanes, a random 16-byte salt, and 32-byte output. Input is the UTF-8 path,
NUL, then the exact value. The reader validates fixed parameters, salt and key
encoding and lengths, non-empty unique paths, and a 1 MiB size limit. It rejects
unknown JSON fields and unsupported parameters. Stored hashes permit offline
guesses of the secret; they do not make state safe to disclose. Use cryptographically generated
signing secrets and protect Terraform state, saved plans, and backups. Argon2 adds
protection against guessing; it is not encryption or a quantified guarantee for
weak secrets. Validation still enforces only the API's 64-character alphabet
rule, so format-valid low-entropy values remain accepted.

A changed write-only value gets a new hash and salt. Hash preparation precedes
PUT so a hashing failure cannot follow a remote mutation. Cancellation is checked
before and after Argon2, which cannot be interrupted once running.

## Memory use

The fixed 64 MiB Argon2 memory parameter does not bound provider process RSS:
concurrent calculations and garbage collection can leave earlier allocations
resident. Lowering Terraform `-parallelism` for both plan and apply can reduce
memory use and may increase elapsed time; a time penalty is not inevitable.

In measurements on 2026-10-05, 50 signing secrets reached approximately 1448 MiB
peak provider RSS during saved rotation apply at parallelism 10, versus 352 MiB
at parallelism 1. These were maxima across six samples per setting (App and
Webhook resources, three repetitions each), using Terraform 1.16.0 and Go 1.27.1
on darwin/arm64, an Apple M5 with 10 CPUs and 24 GiB RAM. Each sample reports the
largest individual provider-process peak, not a sum, Terraform RSS, or machine
memory use. The synthetic local HTTP fixture omits Contentful latency and rate
limits; these observations are neither a performance guarantee nor a prescribed
parallelism setting.

The [historical measurement report](https://github.com/cysp/terraform-provider-contentful/blob/0d48ce06dcc818932521b2d2c137c14994dd7f7b/docs/design/signing-secret-memory.md)
links the runner and raw samples retained as evidence for this decision. Although
the report says no memory-limit or garbage-collection overrides were set, the
runner neither enforced nor recorded `GOGC`,
`GOMEMLIMIT`, or `GOMAXPROCS`. The samples therefore do not establish a controlled
default runtime environment or performance on other workloads and machines.

## Invalid private state

Update rejects invalid verifier data before mutation, even when a known ordinary
value would suffice for comparison. Read and Delete do not inspect the verifier.
During planning, invalid verifier contents produce a warning and unknown
timestamps when the provider inspects a configured write-only value. Private-data
access errors and cancellation remain errors.

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
   and identifies the requested resource.
3. Encode the new resource state and identity.
4. Publish the verifier, then the encoded state and identity.

If any step fails, retain the previous state and verifier. Contentful may
nevertheless have accepted the new secret. Reverting configuration to the
previous secret can then produce no change while Contentful uses the new one.
This is an acknowledgement boundary, not a transaction spanning Contentful and
Terraform. [Protocol tests](../../internal/provider/resource_signing_secret_protocol_test.go)
cover publication failures; [uncertain-write tests](../../internal/provider/resource_signing_secret_recovery_test.go)
cover a lost response and deliberate recovery with a different secret.

App and webhook signing-secret PUT and DELETE requests retry explicit 429
responses with the same request body within the operation deadline. Transport
failures and ordinary 5xx responses stop without replay. Redirects are rejected
in both provider HTTP clients. GET follows the [shared retry policy](contentful-http-retry-policy.md).
This boundary does not provide at-most-once behavior across separate applies.
Diagnostics and logs redact exact matches of the sent secret and any previous
secret stored in ordinary `value`. This does not cover encoded, partial, or
unknown secret values.

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

Terraform versions before 1.12 do not supply prior resource identity. On Read,
these resources reconstruct identity from their scope attributes, including
when a 404 removes resource state. The returned resource state is then null.

A provider predating `value_wo` cannot compare secrets using the verifier. Migrate
to ordinary `value` with the same acknowledged secret before downgrading; the
[practitioner guide](../guides/secrets-and-state.md#downgrading-the-provider)
explains the storage consequences. Compatibility checks against v0.0.69 established
that this migration avoids a PUT, with or without an initial refresh. Direct
downgrade from write-only state selected PUT. This evidence is specific to that
release; it does not establish every older version's behavior.

The linked tests exercise the locally built provider and synthetic HTTP fixtures.
They establish provider behavior, not live Contentful conformance. Dated service
observations remain in the [API research](../research/signing-secret-versions.md).
