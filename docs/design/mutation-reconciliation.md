# Role, Editor Interface, and Webhook reconciliation

Use the shared [value and identity rules](terraform-value-semantics.md) alongside
these resource-specific contracts.

## Webhook Basic password

Contentful accepts `httpBasicPassword` on Webhook Create and Update but omits
the property from successful mutation and GET responses. The generated client
distinguishes absence (`Set` is false) from explicit JSON `null` (`Set` and
`Null` are true). Only absence receives the password fallback; explicit null
and returned values remain response truth.

Mutation projection restores a known non-null effective Plan password only
when the response omits the property. Read uses prior State under the same
condition; import leaves it null. Explicit null and returned values take
precedence. The fallback survives a contradiction in another attribute because
Contentful supplied no competing password value.

Omitting both Basic credentials sends null for both and clears authentication.
A retained password does not verify remote equality or expose external rotation.
See [Webhook observations](../research/webhooks.md#basic-authentication-password).

## Role mutation decisions

Role resource and data-source models use native maps, slices, and policy pointers
with Framework scalar values. Nil containers and pointers represent null;
allocated empty containers remain known empty values. Whole-container and
whole-policy unknown values fail with attribute-path diagnostics during model
decoding. Unknown scalar children remain representable and are rejected by
request conversion before mutation. Role has no model-decoding planning hook;
adding one would require a representation that preserves deferred containers.

Role permissions, policy order, and policy action order compare without ordering
but preserve duplicate multiplicity. Policy constraints compare as normalized
JSON; arrays remain ordered. Contentful scalar `"all"` maps to Terraform `["all"]`.
An unsupported permission or policy variant makes projection lossy and prevents
restoring the planned representation. Report errors at `permissions` or `policies`;
nested policy differences belong to the parent `policies` diagnostic.

All other API-backed values use the shared exact comparison and endpoint identity
rules. Only after every comparison succeeds may reconciliation restore the exact
Plan representation for all API-backed attributes. Read does not run mutation
reconciliation.

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

`editor_layout`, `controls`, `group_controls`, and `sidebar` are ordered. Compare
nested structure and scalar values exactly, with JSON semantics for widget
`settings`. Top-level field layout items cannot be represented by the schema;
losing one emits a warning and prevents equivalence with a known Plan.

Report a mismatch at the enclosing attribute, including nested scalar, layout,
or settings differences. Omitted `sidebar[].disabled` normally plans `false`;
that known default is constrained like an explicitly configured value. See the
[sidebar observations](../research/editor-interfaces.md#sidebar-update-observations).

## Webhook mutation decisions

| Attribute | Comparison or omission rule |
| --- | --- |
| `topics` | Ordered comparison; duplicates remain meaningful. |
| `filters` | Outer order and `in.values` order are insignificant, but duplicate multiplicity matters. Each element must project exactly one supported operator. |
| `headers` | Exact map and object comparison after secret-value fallback. A known planned map is sent even if Config omits it; raw omission on Update can clear headers. `{}` explicitly clears them. |
| `headers[*].value` | If Contentful omits a secret header's value, retain the corresponding planned value, even when another attribute contradicts the Plan. A returned value takes precedence. |
| `headers[*].secret` | Exact boolean comparison; an absent response flag projects as `false`. |
| `transformation` | Exact nullable scalars and JSON semantics for `body`. |
| `http_basic_password` | The absence-only fallback described above. |

All other API-backed values use exact comparison. Defaults for `active` and
`headers[*].secret`, and values retained by `UseStateForUnknown`, constrain
reconciliation whenever the effective Plan is known. Lossy projection or a
meaningful difference retains the complete response and adds an error at the
owning top-level attribute; no otherwise valid API-backed Plan overlay is applied.

Read projects remote values with only the documented secret fallbacks. See
[Webhook evidence](../research/webhooks.md#direct-configuration-observations).
