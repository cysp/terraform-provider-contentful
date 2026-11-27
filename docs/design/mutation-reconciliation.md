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

An ordinary timeout-only update preserves the observed `version`. An update that can rotate a write-only credential leaves `version` unknown, including when planning hides a previously managed input through `ignore_changes`. When apply needs no PUT, it can save timeouts or migrate credential storage even if the full webhook contains an unreadable credential.

For a remote change, Update validates the full request and sends the private
version as `X-Contentful-Version`. Contentful checks this precondition when
the webhook exists. If the webhook was deleted after refresh, the same PUT
can recreate it at the requested ID. See the [version observations](../research/webhooks.md#webhook-definition-versions).

The public `version` records Contentful’s `sys.version`. If a mutation response
identifies a different space, or a PUT returns a different webhook ID, recovery
state clears both versions.

The effective plan owns header membership. Omitting the map clears the collection unless a lifecycle rule retains it. Send ordinary header values, including empty strings. For an existing secret key absent from configuration and retained by a lifecycle rule, send its exact key and `secret = true` without a `value` member. Contentful retains the secret stored under that name. If the webhook was deleted after refresh, the PUT can fail because there is no stored secret to retain.

On creation or replacement, each configured header needs an initial value. Reject known null values during planning, defer unknowns, and validate the resolved values before POST.

For configured ordinary values on update, send each known planned string verbatim, including an empty string or a value retained by `ignore_changes`. If the planned value is null, require `secret = true` and omit the `value` member from the PUT. The header’s previous secret flag does not affect this decision. Check the returned headers against the plan, including the secret flag. See the [observed results for existing, new, and renamed headers](../research/webhooks.md#header-names-order-and-secret-updates).

| Attribute | Equivalence and omission |
| --- | --- |
| `active`, `name`, `url`, `http_basic_username` | Exact scalar comparison. The static `active` default constrains the plan to `true` when applied. A null username is distinct from an empty string. |
| `topics` | Nonempty ordered list. Preserve duplicate counts and reject null elements. |
| `filters` | Ignore outer order and preserve duplicate counts. Match the single selected operator, its negation if present, and all operands. Compare `in.values` without regard to order, preserving duplicates. Compare paths, equality values, and regex patterns exactly. Omission differs from an empty list. Unsupported operators make projection lossy. |
| `headers` | Match the effective plan’s header names, secret flags, and ordinary values. A null or empty map in the effective plan clears the collection. When CMA returns no headers, preserve the plan’s null or empty representation on mutation and the prior representation on refresh. An ignored header from the refreshed state participates in the full PUT. |
| `headers[*].secret` and `.value` | Project an absent flag as `false`. For a secret header whose value Contentful omits, retain only the matching planned value, even on an unrelated contradiction. For an ordinary value, a returned value takes precedence. Write-only paths and previously redacted credentials follow the confidentiality rules below. |
| `transformation` | Compare optional scalar members exactly and the body as normalized JSON, preserving array order. A null body differs from a known JSON string. |

Report differences at the affected top-level attribute. A lossy response cannot establish equivalence with a known plan. See the [Webhook observations](../research/webhooks.md) for Contentful's wire representations.

## Webhook credential responses

For ordinary credentials, use returned bytes, including an empty string. On initial import a single readable value has no write-only history. Previously redacted paths follow the rules below. An explicit JSON null clears the Basic password. A difference from a
known plan is reported after the returned state is saved for recovery.

| Operation | When Contentful omits the password |
| --- | --- |
| Create or Update | Preserve the exact known password from the effective plan, including during rotation or an unrelated response mismatch. An unknown configured password fails before the request. Omitting both Basic credentials on update clears them. |
| Refresh | Retain the previously managed value. External rotation or removal is undetectable. |
| Import | Leave the password null. Configuring credentials afterward writes them to Contentful. |

## Write-only webhook credentials

Public `headers` owns membership and the secret flag. Only `http_basic_password_wo` and `headers["name"].value_wo` supply apply-time plaintext from Config. Build a separate request model from the effective Plan and apply-time credentials. Convert that model to a full CMA request only when a mutation is needed. Response reconciliation receives the unmodified public Plan. A readable response at a protected header path can expose a secret even when the server changes its spelling or reports `secret: false`.

The provider uses the observed `sys.version` to schedule updates when a write-only change has no public value difference. The private version remains the optimistic-lock precondition. Storage-only migrations and unchanged apply-time inputs can resolve a planned update locally. A missing baseline requires a write. Header preservation on a full PUT requires an existing secret header with the same key. Basic authentication requires the password on each PUT. The [practitioner guide](../guides/secrets-and-state.md#webhook-credentials-and-secret-headers) covers migration, import, and recovery.

Each Webhook instance stores its verifiers together under the private key `write_only_secret_hashes`. The optional `basic` object holds the username and password hash. The optional `headers` object maps exact header names to individual hashes. Removing the last verifier removes the key. Webhook credentials share the [Argon2id parameters, PHC encoding and process-wide hashing slot](signing-secret-write-only.md#private-verifier-format) with signing secrets. Each record binds the exact credential bytes to its Terraform path, either `http_basic_password_wo` or `headers["name"].value_wo`. The provider encodes header names with Go string quoting and preserves their exact case. The encoded path is independent of Framework diagnostic formatting. When all prior credentials are absent or ordinary, their public state supersedes obsolete private contents. With redacted siblings, an ordinary value can supersede its own hash in a readable envelope. An invalid envelope blocks comparison because it can contain the only verifier for a sibling. A version advance by itself never acknowledges any secret. If an ephemeral credential first appears at apply and would require a PUT despite a known planned version, return an error before HTTP and require a fresh plan.

A decoded contradictory response can still identify the endpoint and supply a version for recovery. Withhold the staged verifiers. At mutation, the prior verifier paths and the write-only credentials supplied for this request identify the paths to redact. Keep their Basic username or exact secret header descriptors for the next refresh. This can discard a previous ordinary comparison value on a path that is migrating to write-only. Ordinary siblings keep their reconciled values, including the effective-plan fallback when Contentful omits their bytes. This also applies to an ordinary credential configured after an unreadable import.

On refresh, a previously redacted descriptor guards against a readable echo, including a change in header case or secret flag. A later PUT also protects a redacted credential with no verifier when the effective plan supplies no ordinary value. This includes a delegated or removed header after import or an unconfirmed first migration. Fully redacted secret headers with distinct exact spellings can coexist on import and refresh. A transport failure supplies no response checkpoint and leaves the prior public state and verifiers intact. During an ordinary-to-write-only migration, a later response that also violates Contentful’s redaction contract can expose newly submitted bytes as ordinary drift. The [recovery tests](../../internal/provider/resource_webhook_write_only_recovery_test.go) exercise uncertain responses and the [paired ordinary-credential tests](../../internal/provider/resource_webhook_ordinary_header_recovery_test.go) check sibling independence.

A successful refresh that observes removal of a redacted header, or a change to its Basic username, discards the corresponding comparison data. If the credential is recreated before any refresh observes the change, the provider still compares against its last acknowledgement.
