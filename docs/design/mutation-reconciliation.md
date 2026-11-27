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

Each header name absent from prior state needs an initial value, including on creation, replacement or rename. Reject a known absence during planning, defer unknown inputs, and validate resolved values before writing.

For configured ordinary values on update, send each known planned string verbatim, including an empty string or a value retained by `ignore_changes`. If the planned value is null, require `secret = true` and omit the `value` member from the PUT. Compare the refreshed and effective planned secret flags for each exact name retained in the same space. Warn on a known change in the preliminary plan and reject it before an in-place PUT. Contentful retains the old flag and can expose newly submitted bytes on a previously readable header. A new Webhook or a change of space has no prior header at its destination. Terraform 1.17.2 initially plans an explicit same-space `-replace` against prior state, so the preliminary warning can also appear on a valid replacement. Check the returned headers against the plan, including the secret flag. See the [observed results for existing, new, and renamed headers](../research/webhooks.md#header-names-order-and-secret-updates).

| Attribute | Equivalence and omission |
| --- | --- |
| `active`, `name`, `url` | Exact scalar comparison. The static `active` default constrains the plan to `true` when applied. |
| `http_basic_username`, `http_basic_password` | See [Webhook credential responses](#webhook-credential-responses) for submitted and delegated credentials. |
| `topics` | Nonempty ordered list. Preserve duplicate counts and reject null elements. |
| `filters` | Ignore outer order and preserve duplicate counts. Match the single selected operator, its negation if present, and all operands. Compare `in.values` without regard to order, preserving duplicates. Compare paths, equality values, and regex patterns exactly. Omission differs from an empty list. Unsupported operators make projection lossy. |
| `headers` | Match the effective plan’s header names, secret flags, and ordinary values. A null or empty map in the effective plan clears the collection. When CMA returns no headers, preserve the plan’s null or empty representation on mutation and the prior representation on refresh. An ignored header from the refreshed state participates in the full PUT. |
| `headers[*].secret` and `.value` | Project an absent flag as `false`. For a secret header whose value Contentful omits, retain an ordinary value only from an exact-name secret header in the plan (mutation) or previous state (refresh), including on an unrelated mutation contradiction. For an ordinary value, a returned value takes precedence. Write-only paths and previously redacted credentials follow the confidentiality rules below. |
| `transformation` | Compare optional scalar members exactly and the body as normalized JSON, preserving array order. A null body differs from a known JSON string. |

Report differences at the affected top-level attribute. A lossy response cannot establish equivalence with a known plan. See the [Webhook observations](../research/webhooks.md) for Contentful's wire representations.

## Webhook credential responses

Validate both Basic password forms against the effective planned username. An in-place update may retain an imported username through `ignore_changes`. Warn when such an unchanged username exceeds the limit after normalization and a known password requires a write. A replacement can still use its shorter configured username. A creation plan, including the create phase of replacement, must reject a known username without a password or a known password without a username. Check each known credential against its CMA length limit on the create plan, even if the other password form is unknown. The two password forms conflict once both have known, non-null values. When a known username changes and a write-only password is selected, reject an incomplete pair during planning.

When the visible Basic credential change is incomplete but a preliminary replacement plan could retain an ignored counterpart, issue a resource-level warning. Reject a provably invalid pair on the create plan and recheck resolved inputs before sending a PUT. An unchanged username with both password inputs null delegates the password.

Terraform can preview an update using ignored prior values before planning a replacement. Validate a newly configured username against its length limit. For either password form, reject a known oversized value when comparison with the previous ordinary password or verifier establishes a rotation, or when a fully known plan requires a PUT to change settings or headers. Matching bytes can migrate between ordinary and write-only storage without a PUT. Defer checking an ordinary password retained from state because replacement can discard an ignored value. Check mixed empty and nonempty pairs on the create plan or before a credential write, and recheck resolved lengths before sending them.

For ordinary credentials, use returned bytes, including an empty string. On initial import a single readable value has no write-only history. Previously redacted paths follow the rules below. An explicit JSON null clears the Basic password. A difference from a
known plan is reported after the returned state is saved for recovery.

| Operation | When Contentful omits the password |
| --- | --- |
| Create or Update | Preserve the exact known password from the effective plan for the same non-null username, including during rotation or an unrelated response mismatch. An unknown configured password fails before the request. When the effective plan removes both Basic credentials, the provider sends paired JSON nulls. When an update retains the same non-null username and supplies neither password, omit both fields from the PUT. For this password-free PUT, accept the returned username in Terraform’s canonical Unicode spelling. Submitted credentials and verifier bindings still require exact matches. |
| Refresh | Retain the previously managed password for the same non-null username. If the target omits both Basic fields, retain prior Basic state and any matching verifier. An explicit null username clears the credentials when no readable password conflicts. A rotation under the same username remains undetectable. |
| Import | Leave the password null. Configuring credentials afterward writes them to Contentful. |

## Write-only webhook credentials

Public `headers` owns membership and the secret flag. Only `http_basic_password_wo` and `headers["name"].value_wo` supply apply-time plaintext from Config. Build a separate request model from the effective Plan and apply-time credentials. Convert that model to a full CMA request only when a mutation is needed. Write-only password and header bytes stay out of the plan used for response reconciliation. When the receiver URL changes, submit each explicitly supplied write-only header, including one matching its verifier. Protect every secret header for which the actual PUT omitted the value, including newly delegated ordinary headers retained by a lifecycle rule. A readable response at a protected header path can expose a secret even when the server changes its spelling or reports `secret: false`.

The provider uses the observed `sys.version` to schedule updates when a write-only change has no public value difference. The private version remains the optimistic-lock precondition. Storage-only migrations and unchanged apply-time inputs can resolve a planned update locally. A missing baseline requires a write. Header preservation on a full PUT requires an existing secret header with the same key. A delegated Basic password retains any verifier bound to the same username. An otherwise consistent delegated PUT can return that username with an omitted or explicitly null password. The PUT submitted no password to acknowledge. An explicit null password in response to a write-only password submission leaves that write unconfirmed. Import supplies no verifier for it. A later write-only comparison uses the last acknowledged input. The [practitioner guide](../guides/secrets-and-state.md#webhook-credentials-and-secret-headers) covers migration, import, and recovery.

Each Webhook instance stores its verifiers together under the private key `write_only_secret_hashes`. The optional `basic` object holds the username and password hash. The optional `headers` object maps exact header names to individual hashes. Removing the last verifier removes the key. Webhook credentials share the [Argon2id parameters, PHC encoding and process-wide hashing slot](signing-secret-write-only.md#private-verifier-format) with signing secrets. Each record binds the exact credential bytes to its Terraform path, either `http_basic_password_wo` or `headers["name"].value_wo`. The provider encodes header names with Go string quoting and preserves their exact case. The encoded path is independent of Framework diagnostic formatting. When all prior credentials are absent or ordinary, their public state supersedes obsolete private contents. With redacted siblings, an ordinary value can supersede its own hash in a readable envelope. An invalid envelope blocks comparison because it can contain the only verifier for a sibling. A version advance by itself never acknowledges any secret. If an ephemeral credential first appears at apply and would require a PUT despite a known planned version, return an error before HTTP and require a fresh plan.

An otherwise consistent successful PUT acknowledges a planned Basic credential removal even if its response omits the username.

Use provider-authored text for Webhook GET, POST, PUT and DELETE errors. A remote error or decoder message can contain credentials absent from Terraform state. Keep Webhook password and header diagnostic paths in the detail without attaching their HCL source excerpts. The header map validator identifies a null entry by key in the diagnostic detail. A decoded HTTP 422 `ValidationFailed` on POST or PUT directs the practitioner to review the definition. A decoded HTTP 401 directs the practitioner to check the provider’s Contentful management access token. A decoded HTTP 409 `Conflict` on a versioned PUT calls for a fresh plan and review of remote changes. Contentful returned that response for [stale Webhook versions](../research/webhooks.md#webhook-definition-versions). A transport or response-decoding error leaves the outcome of a write uncertain.

A POST reply must identify the requested space and supply a nonempty assigned ID before the provider can record any recovery state. Otherwise, direct the practitioner to inspect the requested space before retrying.

If a PUT reply identifies another Webhook, keep all API-backed fields, identity and credential verifiers from the target’s prior state. Apply local timeout changes and clear the public and private versions. The next write requires a refreshed version.

A contradictory PUT can still identify the target Webhook and supply its version. Withhold proposed verifiers. Build recovery state using the same redacted paths that guarded the reply, including delegated headers that have no verifier.

- If the target returns a different username string, save it and discard the verifier bound to the previous username. If the target omits the username, keep the prior redacted username and its verifier when present.
- If the target explicitly returns a null username and omits the password or returns it as null, clear both public Basic fields and the Basic verifier. If the plan retained authentication, report the mismatch and keep the confirmed version. A readable password alongside a null username leaves a previously redacted credential protected.
- An explicit header list from the target invalidates the verifier for each absent name. A name still in the plan keeps its redacted descriptor to guard the next refresh. A name absent from the plan also loses its descriptor. An omitted list leaves prior redacted descriptors and their eligible verifiers in place.

An explicit ordinary value from the identified endpoint may enter recovery state when it matches the planned value and the corresponding username or exact header name and flag. This releases the old verifier and lets an ordinary refresh accept that value. Publish only prior verifiers eligible for the resulting state. An ordinary value supersedes its old verifier even if recovery removes the ordinary value during migration to write-only storage. Ordinary siblings keep their reconciled values, including the effective-plan fallback when Contentful omits their bytes. This also applies to an ordinary credential configured after an unreadable import.

On refresh, a previously redacted descriptor guards against a readable echo, including a change in header case or secret flag. A later PUT also protects a redacted credential with no verifier when the effective plan supplies no ordinary value. This includes a delegated or removed header after import or an unconfirmed first migration. Fully redacted secret headers with distinct exact spellings can coexist on import and refresh. A transport failure supplies no response checkpoint and leaves the prior public state and verifiers intact. During an ordinary-to-write-only migration, a later response that also violates Contentful’s redaction contract can expose newly submitted bytes as ordinary drift. The [recovery tests](../../internal/provider/resource_webhook_write_only_recovery_test.go) exercise uncertain responses and the [paired ordinary-credential tests](../../internal/provider/resource_webhook_ordinary_header_recovery_test.go) check sibling independence.

On successful refresh a header verifier survives only across a continuously present, redacted exact-name header. A Basic verifier survives across the same redacted username when the password is omitted. An explicit null password invalidates it. A newly observed redacted credential cannot inherit an orphan record. When no credential continues, refresh can discard even an unreadable envelope. When at least one continues, refresh can remove obsolete records from a readable envelope while preserving a malformed sibling for a subsequent Update diagnostic. If the credential is recreated before any refresh observes the change, the provider still compares against its last acknowledgement.
