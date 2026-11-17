# Reconciling Role, Editor Interface, and Webhook writes

After Create or Update, compare the complete response with the effective plan. Restore the plan's exact representation only if every constrained value is equivalent. On a mismatch, keep response-derived values and diagnose the affected top-level attribute. Keep projection warnings alongside errors. Read uses the remote representation, with the credential fallbacks below.

The requested endpoint stays fixed and `timeouts` always comes from the plan. Publish state and handle the private version under the [endpoint identity policy](terraform-value-semantics.md#mutation-endpoint-identity) before returning consistency diagnostics. A conflicting known composite `id` gets its own error.

| Resource | Create | Update |
| --- | --- | --- |
| Role | POST to the space. Contentful assigns `role_id`. | PUT to the planned space and role IDs with the private version. |
| Editor Interface | Versioned PUT to the interface identified by space, environment, and Content Type. | PUT to the same endpoint using the [activation count](#editor-interface-version-preconditions). |
| Webhook | POST to the space. Contentful assigns `webhook_id`. | Versioned PUT to the planned space and webhook IDs. |

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

| Attribute | Equivalence and omission |
| --- | --- |
| `active`, `name`, `url`, `http_basic_username` | Exact scalar comparison. The static `active` default constrains the plan to `true` when applied. A null username is distinct from an empty string. |
| `topics` | Nonempty ordered list. Preserve duplicate counts and reject null elements. |
| `filters` | Ignore outer order and preserve duplicate counts. Match the single selected operator, its negation if present, and all operands. Compare `in.values` without regard to order, preserving duplicates. Compare paths, equality values, and regex patterns exactly. Omission differs from an empty list. Unsupported operators make projection lossy. |
| `headers` | Match map keys and complete header objects. The plan can preserve prior headers on Update because omitting the CMA member can clear them. An explicit empty map clears the headers. An unknown plan on Create takes the response. |
| `headers[*].secret` and `.value` | Project an absent flag as `false`. For a secret header whose value Contentful omits, retain only the matching planned value, even on an unrelated contradiction. A returned value takes precedence. |
| `transformation` | Compare optional scalar members exactly and the body as normalized JSON, preserving array order. A null body differs from a known JSON string. |

Report differences at the affected top-level attribute. A lossy response cannot establish equivalence with a known plan. See the [Webhook observations](../research/webhooks.md) for Contentful's wire representations.

## Webhook Basic password

Contentful accepts `httpBasicPassword` on Webhook Create and Update but omits
the property from successful mutation and GET responses. The generated client
distinguishes absence (`Set` is false) from explicit JSON `null` (`Set` and
`Null` are true). Only absence receives the password fallback; explicit null
and returned values remain response truth.

| Operation | When Contentful omits the password |
| --- | --- |
| Create or Update | Preserve the exact known password from the effective plan, including during rotation or an unrelated response mismatch. An unknown configured password fails before the request. Omitting both Basic credentials on update clears them. |
| Refresh | Retain the previously managed value. External rotation or removal is undetectable. |
| Import | Leave the password null. Configuring credentials afterward writes them to Contentful. |

When the password is present, use the returned value, even explicit null. A difference from a known plan is a consistency error at `http_basic_password`.
