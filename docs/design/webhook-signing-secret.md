# Webhook signing secret lifecycle

`contentful_webhook_signing_secret` manages the space-wide WebhookSigningSecret
singleton. Its `space_id` is the resource identity and import ID; Terraform `id`
equals `space_id`.

The resource accepts GET 200, PUT 200/201, and DELETE 204 success. GET/PUT responses
must contain `sys.type: WebhookSigningSecret`, a Space link, and a string
`redactedValue`. The resource checks that the returned space matches the request.
The client models no secret-specific `sys.id`, timestamps, users, or version.

[API research](../research/webhook-signing-secret.md) records the external evidence
and its limitations. [Terraform value semantics](terraform-value-semantics.md)
and the [HTTP retry policy](contentful-http-retry-policy.md) supply shared rules.
[Test-server conformance](cma-test-server-conformance.md#webhook-signing-secret)
documents fixture behavior and coverage.

## Ownership and value publication

Create deliberately PUTs the supplied value, replacing an existing singleton.
Mutations are unconditional, with no version state or precondition parameters in
the client or mock. The [endpoint experiment](../research/webhook-signing-secret.md#direct-observations)
records the evidence for this contract. Destroy deletes
the current singleton, even after external rotation or an import that retained
no complete secret. Changing the space replaces the resource. Multiple resources
can target the same space and overwrite or delete one another's secret. A
same-space `create_before_destroy` replacement can delete the newly written secret.

The required sensitive `value` persists in Terraform state and saved plans.
The [OpenAPI request schema](../../internal/contentful-management-go/openapi/schemas/webhook-signing-secret/request-data.yml)
defines the format constraint; configuration validation and request conversion
use its generated validator with fixed safe diagnostics. Configuration validation
defers null/unknown. Mutation conversion rejects null, unknown, and invalid known
values, including the known-empty string. It never substitutes configuration for
the effective plan.

| Operation | Remote action | Published value |
| --- | --- | --- |
| Create | PUT effective planned value | Exact sent value after a valid acknowledged response |
| Read | GET and validate entity/space identity | Exact known prior value; imported null stays null |
| Update with changed effective value | PUT effective planned value | Exact sent value after a valid acknowledged response |
| Timeout-only Update or ignored value change | No PUT | Prior value, including imported null; planned timeouts |
| Import | Set space identity, then GET | Null; redacted metadata cannot recover the value |
| Configure after import | PUT if effective value changes from null | Exact sent value after acknowledgement |
| Read HTTP 404 with typed `NotFound` | No parent query | Remove resource; parent or singleton absence both qualify |
| Delete | DELETE; accept 204 or typed 404 `NotFound` | Remove resource |

`redactedValue` is never stored as `value`, compared to a suffix, logged, or
published as another Terraform attribute. Read verifies presence and identity;
it does not verify that the retained bytes match Contentful. External rotation
with either an equal or different suffix does not trigger automatic repair.
The successful PUT response acknowledges the request; its redacted field is
not an independent equality proof.

CLI import alone does not rotate. An import block with a configured value can
rotate during the same apply. `ignore_changes = [value]` preserves the imported
null and prevents that PUT. The Update equality check skips request conversion
for unchanged values, including null/null; GET and DELETE need no secret value.
The setting does not affect Create: creation and replacement require the complete
configured value. The [import guide](../guides/secrets-and-state.md#importing-signing-secrets)
explains the ownership choice. Ordinary timeout changes preserve the remote secret
and local value.

## Failures and confidentiality

No version or conditional headers are sent. The request-context retry
opt-out covers every PUT/DELETE attempt, including 429 and redirects in both
provider-owned HTTP client layers. GET retains normal finite deadline/retry
behavior. There is no read-after-write recovery: presence cannot prove that
requested secret bytes were stored. An ambiguous DELETE also returns an error;
a subsequent refresh may observe absence.

Read and Delete treat HTTP 404 as absence only with a decoded Contentful `NotFound`
error. Other or malformed 404 responses remain errors. The endpoint experiments
observed this typed error for GET/DELETE of an absent secret and GET/PUT/DELETE
under an absent parent.

Malformed, unexpected, or identity-contradicting success responses are errors.
Failed Create publishes no resource state. Failed Update/Read/Delete preserve
prior state and identity. This is not proof that a mutation failed remotely.
Import can establish presence after an uncertain Create, but cannot confirm
which value was stored. Later deliberate rotation or destroy can affect another
actor's intervening changes. These operations do not provide at-most-once semantics
across separate applies or arbitrary transports.

The resource's CRUD log calls emit operation names without request or response
fields. Shared HTTP retry logs can include method, status, retry ordinal, and
timing. Resource diagnostics use the same upstream error formatter and literal
secret redaction as App Signing Secret. Exact occurrences of known secret values
are replaced with `***`: the planned value for Create, the prior value for Read
and Delete, and both values for Update. Other upstream detail is retained.
Encoded, partial, or unknown secret values are outside this redaction policy.
Schema sensitivity does not remove secrets from stored Terraform state, plans,
or external tooling.
