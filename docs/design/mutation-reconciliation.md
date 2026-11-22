# Reconciling Role, Editor Interface, and Webhook writes

After Create or Update, compare the complete response with the effective plan. Restore the plan's exact representation only if every constrained value is equivalent. On a mismatch, keep response-derived values and diagnose the affected top-level attribute. Keep projection warnings alongside errors. Read uses the remote representation, with the credential fallbacks below.

The requested endpoint stays fixed and `timeouts` always comes from the plan. Publish state and handle the private version under the [endpoint identity policy](terraform-value-semantics.md#mutation-endpoint-identity) before returning consistency diagnostics. A conflicting known composite `id` gets its own error.

| Resource | Create | Update |
| --- | --- | --- |
| Role | POST to the space. Contentful assigns `role_id`. | PUT to the planned space and role IDs with the private version. |
| Editor Interface | Versioned PUT to the interface identified by space, environment, and Content Type. | PUT to the same endpoint using the [activation count](#editor-interface-version-preconditions). |
| Webhook | POST to the space. Contentful assigns `webhook_id`. An empty returned ID prevents state publication. Inspect the space before retrying the POST. | Persist timeout-only changes locally. Changes to webhook fields require a versioned PUT to the planned space and webhook IDs. |

Create retains the server-assigned Role or Webhook ID even if it contradicts a known plan. Update retains requested IDs on a contradiction. Webhook request conversion consults configuration for Optional+Computed omission. All three resources compare the response against the effective plan, including values retained by `ignore_changes`.

## Role mutation decisions

Role resource and data-source models use native maps, slices, and policy pointers
with Framework scalar values. Nil containers and pointers represent null;
allocated empty containers remain known empty values. Whole-container and
whole-policy unknown values fail with attribute-path diagnostics during model
decoding. Unknown scalar children remain representable and are rejected by
request conversion before mutation. Role has no model-decoding planning hook;
adding one would require a representation that preserves deferred containers.

Compare names and nullable descriptions exactly. Apply these rules to Role collections before restoring plan order:

| Attribute | Equivalence |
| --- | --- |
| `permissions` | Match map keys exactly. Ignore action order within each list and preserve duplicate counts. |
| `policies` | Ignore policy order and preserve duplicate counts. Match effects exactly. Compare actions as unordered lists, projecting Contentful's scalar `"all"` as `["all"]`. Compare constraints as normalized JSON, preserving array order. |

An unrepresentable permission or policy variant makes the projection lossy. Keep the response and report the warning and a consistency error on the parent attribute.

## Editor Interface version preconditions

Create sends `X-Contentful-Version: 1` plus the provider's counted Content Type
activations. This permits management of the initial automatically created
interface, including after a subsequent activation in the same apply. Update
uses the private `version` last observed by Read or an interface mutation, plus
the same count. A version conflict fails without fetching a newer version or
retrying; adopting an existing modified interface requires import. A mismatch
alone does not establish its cause or whether import is necessary.

The count is in-memory, scoped to one provider configuration, and keyed by
`space_id`, `environment_id`, and `content_type_id`. Resources changing together
must use the same provider configuration and a Terraform dependency. Separate
aliases do not share counts, even with identical credentials. Read and successful
Create/Update checkpoint the actual response version and reset the count before
reporting mutation consistency errors. Restart discards the count; refresh
supplies a new observed baseline. Counting does not schedule or order writes.

Only activation of an already published Content Type advances an existing
interface. Determine this from the known, response-derived `published_version`
**before activation**, including recovery through Update. An unpublished draft
has no interface, so activation creates version 1 and contributes no offset.
Count the increment after checkpointing the activation response and before
reporting consistency or activation diagnostics; failed activation responses do
not reach this checkpoint.

Direct CMA experiments establish the lifecycle expectations used by the tests:

| Content Type operation | Editor Interface observation |
| --- | --- |
| Create an unpublished draft | GET returns 404. |
| First activation | Interface starts at version 1. |
| Update the draft before activation | Interface version is unchanged. |
| Activate an already published type, including unchanged content or description-only, field-name, and added-field changes | Interface version increments once. |
| Deactivate | GET returns 404. |
| Reactivate after deactivation | Default interface at version 1; previous widget settings are absent. |

The mock follows these observations. Read removes a missing interface from
Terraform state; this grants no authority to reactivate an externally deactivated
Content Type. Editor Interface Delete only relinquishes Terraform ownership.

## Editor Interface mutation decisions

Compare `editor_layout`, `controls`, `group_controls`, and `sidebar` in order. A null list and an empty list are distinct. Match layout group IDs, names, nested group and field selections, and field IDs exactly. A top-level field layout item cannot be represented by the schema. Report a lossy projection at `editor_layout`.

For controls and sidebar widgets, compare the identifiers and namespace exactly and compare `settings` as normalized JSON. A known `sidebar[].disabled` constrains the response, including the default `false`. Contentful has returned explicit `false` when sent and omitted the property when absent from the request. Report mismatches at the enclosing list.

## Webhook mutation decisions

When the remotely managed values in the effective plan match prior state,
planning preserves the prior `version`. Update saves the planned `timeouts`
without sending a PUT, including when Contentful has redacted credentials.

For a remote change, Update validates the full request and sends the private
version as `X-Contentful-Version`. Contentful checks this precondition when
the webhook exists. If the webhook was deleted after refresh, the same PUT
can recreate it at the requested ID. See the [version observations](../research/webhooks.md#webhook-definition-versions).

The public `version` records Contentful’s `sys.version`. If a mutation response
identifies a different space, or a PUT returns a different webhook ID, recovery
state clears both versions.

When configuration omits `headers`, build the PUT from the effective plan. Send ordinary header values, including empty strings. For each secret header, send its exact key and `secret = true` without a `value` member. Contentful retains the secret stored under that name. If the webhook was deleted after refresh, the PUT can fail because there is no stored secret to retain.

On creation or replacement, each configured header needs an initial value. Reject known null values during planning, defer unknowns, and validate the resolved values before POST.

On update with a configured map, send each known planned string verbatim, including an empty string or a value retained by `ignore_changes`. If the planned value is null, require `secret = true` and omit the `value` member from the PUT. The header’s previous secret flag does not affect this decision. Check the returned headers against the plan, including the secret flag. See the [observed results for existing, new, and renamed headers](../research/webhooks.md#header-names-order-and-secret-updates).

| Attribute | Equivalence and omission |
| --- | --- |
| `active`, `name`, `url`, `http_basic_username` | Exact scalar comparison. The static `active` default constrains the plan to `true` when applied. A null username is distinct from an empty string. |
| `topics` | Nonempty ordered list. Preserve duplicate counts and reject null elements. |
| `filters` | Ignore outer order and preserve duplicate counts. Match the single selected operator, its negation if present, and all operands. Compare `in.values` without regard to order, preserving duplicates. Compare paths, equality values, and regex patterns exactly. Omission differs from an empty list. Unsupported operators make projection lossy. |
| `headers` | Match map keys and complete header objects. The plan can preserve prior headers on Update because omitting the CMA member can clear them. An explicit empty map clears the headers. An unknown plan on Create takes the response. |
| `headers[*].secret` and `.value` | An absent secret flag means `false`. For returned and omitted values, see [credential responses](#webhook-credential-responses). |
| `transformation` | Compare optional scalar members exactly and the body as normalized JSON, preserving array order. A null body differs from a known JSON string. |

Report differences at the affected top-level attribute. A lossy response cannot establish equivalence with a known plan. See the [Webhook observations](../research/webhooks.md) for Contentful's wire representations.

## Webhook credential responses

Use any Basic password or header value Contentful returns, including an empty
string. An explicit JSON null clears the Basic password. A difference from a
known plan is reported after the returned state is saved for recovery.

Contentful normally omits the Basic password and secret header values. On
omission, retain the known value from the effective plan after a write or from
prior state on refresh. Import leaves omitted values null. For a secret header,
look up the exact Terraform key, regardless of its previous secret flag. Case
variants are distinct keys in the [webhook definition](../research/webhooks.md#header-names-order-and-secret-updates).
