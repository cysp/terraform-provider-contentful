---
page_title: "Secrets and Terraform state"
description: |-
  Understand which Contentful secrets Terraform stores, which import can recover, and how to manage rotation.
---

# Secrets and Terraform state

Restrict access to Terraform state and saved plans when managing credentials. Sensitive expressions redact values in normal CLI output. Contentful controls which values it returns on refresh and import. [Terraform sensitive-data guidance](https://developer.hashicorp.com/terraform/language/manage-sensitive-data)

| Value | Refresh | Import |
| --- | --- | --- |
| [Webhook signing secret](#webhook-signing-secrets) | Preserves ordinary `value` or the write-only verifier | Neither secret argument has a value |
| [App signing secret](#app-signing-secrets) | Preserves ordinary `value` or the write-only verifier | Neither secret argument has a value |
| [Webhook Basic password](#basic-authentication) | Preserves ordinary password or write-only verifier | Null when omitted |
| [Secret webhook header](#secret-custom-headers) | Preserves ordinary value or write-only verifier by header name | Unreadable value is null |
| [App Installation `Secret` parameter](#app-installation-secrets), read with a personal access token | Can retain prior state for an all-asterisk response | Retains mask; original value unavailable |
| [Personal Access Token](#personal-access-tokens) | Preserves token from creation | `token` is null |
| Delivery API access token | Reads token | Reads token |
| [App Key public JWK](#app-keys) | Reads public key | Reads public key; private key unavailable |

## Webhook signing secrets

[`contentful_webhook_signing_secret`](../resources/webhook_signing_secret) manages the signing secret for all webhooks in a space. Use one resource per space.

Coordinate receivers before rotating the space's secret. Contentful describes how to accept the old and new secrets during [key rotation](https://www.contentful.com/developers/docs/extensibility/webhooks/request-verification/#key-rotation).

## App signing secrets

[`contentful_app_signing_secret`](../resources/app_signing_secret) manages the signing secret for an App Definition. Use one resource per App Definition.

After an uncertain create or update, a read shows only the final four characters of the current secret. Coordinate with the app backend and other writers and [verify a fresh signed request](#recovering-an-uncertain-signing-secret-write) before retiring a key.

If a delete loses its acknowledgement, check whether the App Definition still has a secret. If one exists, confirm with other writers that it should be removed before applying again. The delete has no version precondition and can remove a replacement.

## Write-only signing secrets

Both signing-secret resources accept exactly one of `value` or `value_wo`. With Terraform 1.11.1 or later, `value_wo` omits the secret from the planned and applied resource values. Supply it through a sensitive ephemeral variable so Terraform also omits the variable value from saved plans and state. Referencing the variable in HCL keeps the literal secret out of configuration source that Terraform can display in diagnostics. See the [App Signing Secret example](../resources/app_signing_secret#write-only-signing-secret) or the [Webhook Signing Secret example](../resources/webhook_signing_secret#write-only-signing-secret).

Use cryptographically generated signing secrets and protect Terraform state, saved plans, and backups. The provider checks the API's 64-character alphabet rule; it does not reject weak, format-valid secrets.

Terraform retains an Argon2 hash for comparison with the last successfully applied secret. This adds protection against guessing, but it is not encryption or a guarantee that weak secrets cannot be recovered. Someone with access to state can test guesses. Refresh cannot compare the complete remote secret with the last acknowledged bytes. A change to the reported update time does not identify the active key.

Each provider process hashes one signing secret at a time, so you can keep Terraform's usual parallelism for other resources. When many secrets change together, hashing can take longer. Time spent waiting counts toward the resource's create or update timeout.

## Migrating signing-secret values

To change how Terraform stores a signing secret, replace `value` with `value_wo`, or the reverse, and supply exactly the same secret Terraform last successfully applied. Plan and apply the change. Moving to `value_wo` removes the complete value from new state; moving to `value` stores it again. Neither migration rotates the remote secret. Supplying a different secret rotates it.

Remove any `ignore_changes` rule for these arguments before migrating. An ignore rule can suppress the storage change or retain an ordinary value that conflicts with `value_wo`. A saved ordinary-to-write-only migration plan includes the prior state, which still contains the complete ordinary secret even when the new input is ephemeral and the CLI redacts it. Protect that plan, earlier state snapshots, and backups. [Terraform plan representation](https://developer.hashicorp.com/terraform/internals/json-format#plan-representation)

## Planning and rotating write-only signing secrets

Create a fresh plan when rotating a write-only secret. The secret supplied at apply can differ from the one supplied during planning. If the resource has a planned update, the provider compares the apply-time secret with the last successfully applied value and writes it only when it differs or no comparison value exists. This includes updates caused by a timeout change.

If you change a timeout while using `value_wo`, Terraform plans `updated_at` as unknown because the secret can change at apply. A downstream resource using `updated_at` as a replacement trigger can therefore be replaced even when the provider skips the secret PUT and the timestamp stays the same.

Applying a saved **no-op** plan does not update the resource. Supplying a different ephemeral secret when applying that plan cannot rotate it.

`ignore_changes = [value_wo]` can suppress an update caused solely by changing the secret. It does not prevent rotation when another change causes an update. Do not use this rule to keep an externally managed secret unchanged during other updates. Ordinary `value` follows Terraform's usual `ignore_changes` behavior.

## Recovering an uncertain signing-secret write

Both signing-secret resources retry explicit rate limiting (HTTP 429) with the same request within the operation timeout. Connection errors and server errors (HTTP 5xx) are not automatically retried, and writes and deletion do not follow redirects. Reads retain their normal retry behavior. Separate applies can still repeat a mutation.

Contentful may accept a write even when its acknowledgement is lost. The provider returns the previous comparison value on a failed update and records no new value for an unacknowledged create. If Terraform saves that previous baseline, reverting configuration to it can produce a no-op while Contentful uses the new secret. After an interrupted apply or provider error, check which state your backend saved. If Terraform lost the write-only verifier, applying again can write the secret even when you supply the same bytes.

If the backend fails to save state after Contentful acknowledges a rotation, Terraform may write a local recovery file such as `errored.tfstate`. In write-only mode, the recovery file can hold the new verifier while the backend still holds the previous one. Protect the recovery file and follow Terraform's instructions for recovering the backend state before planning or applying again. [Terraform state storage and recovery](https://developer.hashicorp.com/terraform/language/state/backends#state-storage)

Refresh cannot identify the active key. Coordinate recovery with consumers before rotating again.

To recover by deliberately installing a known value:

1. Prepare consumers to accept the recovery secret alongside any keys they must continue accepting during the transition. Use a secret different from the last value Terraform successfully applied. You can reuse the unacknowledged replacement if it differs from that value.
2. Supply that secret and create a fresh plan. Confirm the plan will create or update the intended signing-secret resource, then apply with the recovery secret. A saved no-op plan cannot perform this recovery.
3. Verify a new Contentful-generated signature with the intended secret at the consumer, then retire superseded keys according to the consumer's rotation procedure. A redacted suffix or a subsequent no-op Terraform plan alone does not establish complete-secret equality.

Coordinate all writers during recovery. Another actor can rotate the secret after verification.

## Downgrading the provider

Before selecting a provider version that does not support `value_wo`, migrate back to ordinary `value` using the current provider:

1. Replace `value_wo` with `value`, supplying exactly the same secret Terraform last successfully applied. Remove any `ignore_changes` rule that would suppress this migration. Ordinary `value` cannot use an ephemeral variable; remove `ephemeral = true` from that input and keep `sensitive = true`.
2. Plan and apply the storage change. The complete secret is stored in state again; protect state, saved plans, and their backups accordingly.
3. Remove references to attributes unsupported by the target version. For example, v0.2.0 does not expose `updated_at` on signing-secret resources.
4. Select the older provider version, run `terraform init -upgrade` to update the dependency lock file, and review a fresh plan before applying.

A direct downgrade from write-only state can rotate the secret because the older provider cannot compare it with the stored hash. Resolve any [uncertain write](#recovering-an-uncertain-signing-secret-write) before downgrading. Check the target version's documentation for other resource constraints.

## Importing signing secrets

Contentful supports one webhook signing secret per space and one app signing secret per App Definition. Creating a secret overwrites any existing value at that scope. Destroying the resource deletes the current secret, including one rotated outside Terraform. If a replacement at the same scope uses `create_before_destroy`, the final delete removes the newly written secret. To rotate in place, supply the new secret through the configured argument and apply a fresh plan.

Import or refresh can confirm a signing secret exists. Contentful returns only redacted metadata. For a write with an uncertain outcome, follow the [recovery procedure](#recovering-an-uncertain-signing-secret-write).

Import leaves `value` null. Choose who will manage rotation before applying:

- **Terraform.** Configure the secret to install. Import supplies neither an ordinary value nor a verifier. Applying either argument can replace the remote secret in the same operation as the import.
- **An external system.** Set `ignore_changes = [value]` to preserve the existing secret on updates. Supply a valid `value` for creation or replacement, including recreation after remote deletion. Destroy still deletes the current secret. Remove `ignore_changes` when you want Terraform to rotate it.

## Recovering invalid signing-secret state

An **Invalid write-only secret private state** diagnostic means Terraform's stored comparison data cannot be used and no known ordinary value supplies the comparison. In-place updates fail before writing the secret. Planning can warn about this problem and still succeed; other resources may change during apply before the signing-secret update fails.

When state already contains a known ordinary `value`, updates compare against it and discard obsolete verifier data after success. Migrating to `value_wo` prepares fresh comparison data; matching secret bytes require no PUT.

Restore valid Terraform state before updating the resource. Explicit replacement and destruction remain available without repairing the comparison data. Destruction deletes the current remote secret.

For replacement recovery, review the plan and require destruction of the old instance before creation of its replacement (`-/+`). Both resources manage a singleton: one secret per App Definition or space. With `create_before_destroy`, deleting the old instance deletes the newly written secret, even when Terraform reports a successful apply. This lifecycle setting can propagate from dependent resources even when it is absent from the signing-secret configuration. Adjust the relevant lifecycle settings and create a fresh plan; do not apply the recovery replacement until its plan shows destroy before create. Coordinate with consumers for the interval without a secret and the subsequent installation of the configured key.

## Webhook credentials and secret headers

Use Terraform 1.11.1 or later and ephemeral inputs for `http_basic_password_wo` and `headers["name"].value_wo`. Use randomly generated webhook credentials. The provider stores a salted Argon2id verifier for each acknowledged write-only value. Anyone with state access can test guesses, including guesses for short passwords. The two signing-secret resources and webhook credentials share one hashing slot per provider process. [Webhook example](../resources/webhook#example-usage)

### Basic authentication

To create or replace a webhook with Basic authentication, supply `http_basic_username` and either `http_basic_password` or `http_basic_password_wo`. Changing the username also requires a password. Use a nonempty password for a nonempty username. Contentful limits the username to 50 characters and the password to 255. Some Unicode characters count as two toward these limits. Avoid colons in usernames because HTTP Basic uses the first colon to separate the username from the password. [Contentful configuration](https://www.contentful.com/developers/docs/extensibility/webhooks/configure-webhook/#http-basic-authentication) · [HTTP Basic authentication](https://www.rfc-editor.org/rfc/rfc7617.html#section-2)

Terraform normalizes Unicode strings to NFC. A configured password or secret header can therefore send different bytes from a visually identical credential created outside Terraform. During delegation Terraform omits the credential bytes from the PUT. Contentful documents retention of an omitted secret header. For Basic authentication, the observed username and version changes support retention. Verify authentication at the receiver. When taking ownership, coordinate rotation with the receiver. Random ASCII credentials avoid this normalization difference.

Reference a sensitive variable for an ordinary password or an ephemeral sensitive variable for a write-only password. Use variables for secret header values as well. Terraform diagnostics can print a source line from the configuration. In a single-line generated `.tf.json` file, that excerpt can include any inline password on the same line.

An **Incomplete Basic credentials for an in-place webhook update** warning means the visible pair would fail on an update. Supply the missing counterpart before applying. Terraform can otherwise apply prerequisites before the Webhook fails. This warning can also appear on a valid replacement when `ignore_changes` retains an old value during preliminary planning. Review the replacement’s configured pair.

An imported username can exceed Contentful’s 50-unit limit after Terraform normalizes it. If a password would be sent with that username, planning warns and an in-place update fails before the PUT. Keep both password arguments omitted to delegate the pair, or supply a shorter username and its password.

For a replacement, Terraform can validate a known pair during planning. An unknown password can resolve to an invalid value at apply, and an ephemeral password can change between plan and apply. With the default destroy-first order, this can leave the old webhook deleted and the replacement uncreated. If the receiver must remain available and can tolerate overlapping deliveries from two active webhooks, use `create_before_destroy`.

Contentful normally returns the username and omits the password. Import leaves an omitted password null. If Contentful returns a readable password on initial import, Terraform can store it in ordinary state even when configuration uses the write-only argument. Choose how to manage the imported credentials:

- **Preserve the password.** Keep the same username and leave both password arguments null. An expression that evaluates to null has the same effect as omission. On a full PUT, Terraform omits both Basic fields and checks the returned username. This also preserves the credentials when changing the receiver URL. Supply credentials for the new receiver or clear authentication if it should not use the existing pair.
- **Manage the password.** Supply either password argument. If another system owns the username, `ignore_changes = [http_basic_username]` can retain the refreshed username on updates. Password rotation sends Terraform’s Unicode spelling of that username (NFC), which can change the login for a receiver that compares the original bytes. Creation and replacement still require a configured username. For a module that must supply the password on every write, set `nullable = false` on its password variable. An ignored ordinary password retained in the effective plan can also be resent.
- **Clear Basic authentication.** Remove the username and both password arguments from the effective plan. The next PUT sends paired nulls.

Delegating the password retains its verifier for the same username. If another system rotates the password, restoring the last acknowledged write-only input can produce an empty plan. Switching back to `http_basic_password` with that same value can store it in Terraform state without a PUT. Consumers reading state can then receive a stale password. A subsequent change such as a webhook rename can send the configured password and overwrite the external rotation. To resume management, coordinate with the receiver and supply a fresh value. On refresh, an omitted password retains its prior ordinary value for the same username. If Contentful omits both Basic fields, refresh normally keeps the previous pair. A different returned username drops the previous password or verifier.

If the webhook is deleted between refresh and PUT, Contentful can recreate it at the same ID despite the version header. A password-free PUT can then create an active webhook without Basic authentication. Terraform reports the missing username after the write. Supply a password on active updates when the receiver requires credentials even in this race.

### Secret custom headers

Declare each managed header in `headers`. For a secret header, set `secret = true` and supply either its ordinary `value` or its write-only `value_wo`. Terraform keeps an independent verifier for each acknowledged write-only value. An ordinary header can coexist with write-only headers. Use a sensitive expression for an ordinary secret value if you want Terraform to redact it from CLI output. Ordinary values remain in state.

Changing any part of the receiver URL, including its path, sends each explicitly supplied `value_wo` on that PUT. This can overwrite an external rotation even when the input matches Terraform’s last acknowledged value. Omitting the value delegates that header and can carry the external credential to the new receiver. Review retained headers and Basic authentication before changing the URL. On a full webhook update, the outcome depends on the header declaration.

- Contentful keeps an existing header’s `secret` flag in both directions. Terraform warns about a known change during planning because it can preview an in-place update before resolving a replacement. A replacement can establish new flags. An in-place apply rejects the change before the Webhook PUT, even if prerequisites have already changed. To change the flag, remove the header in one apply and recreate it in a later apply, or use a new name. To rotate the value of an existing secret header, keep `secret = true` and change the value in one apply. If refresh observes a formerly readable header as secret, it drops the former value.
- A new or renamed secret header needs its value, even if the new name differs only in capitalization. Include only one spelling of each name.
- A retained secret header can be sent with its exact key and `secret = true`, omitting the value. Contentful keeps the existing remote bytes. This also works after import, when Terraform has neither the value nor a verifier. Supplying a write-only value after import installs it.
- Removing only `value_wo` while retaining the header declaration preserves the remote header and any previous verifier. A later comparison still uses the last acknowledged value. To delete the remote header, remove its key from `headers`. Omitting the map or setting `headers = {}` clears all headers unless a lifecycle rule retains them. [Contentful custom header updates](https://www.contentful.com/developers/docs/extensibility/webhooks/headers/#custom-headers)

**Upgrading a webhook with omitted headers.** Provider v0.2.x retained existing headers when the map was omitted. Before upgrading, add `lifecycle { ignore_changes = [headers] }` if another system manages the entire collection. This also preserves imported headers on subsequent updates. To delegate just one header while managing its siblings, omit that key from configuration and add it to `ignore_changes`, for example `headers["X-External"]`. Terraform keeps the refreshed header in the plan for the full webhook PUT. To manage the collection in Terraform, declare the headers to keep and review the first plan for removals.

When a lifecycle rule retains an omitted secret header, Terraform sends its exact key without a value. If you still declare the ignored key with an ordinary value, a full PUT can resend the value retained in the plan and overwrite an external rotation. If an external edit makes a previously redacted header readable, refresh reports an error and discards the response value. A subsequent update also discards an unexpected readable value for a delegated or removed redacted header.

To delegate rotation of an existing secret, remove its `value_wo` from configuration. Ignoring a configured `value_wo` can suppress a rotation-only plan, but any update triggered by another attribute can still send its apply-time value. An ignored header must have a value at creation if configuration declares it. [Terraform lifecycle rules](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#ignore_changes)

### Migration, rotation, and recovery

Switch an ordinary password or secret header to write-only storage by moving the same acknowledged bytes to the corresponding write-only argument. Switching back stores the bytes in public resource state. A storage-only migration needs no PUT. Remove conflicting `ignore_changes` rules before migrating. Old state snapshots and plans made during migration can contain the prior ordinary value.

Create a fresh plan for each rotation. A change to a write-only input alone can schedule an update through the webhook's computed `version`. Any scheduled update compares the actual apply-time inputs. Applying a saved no-op plan makes no mutation, even with different ephemeral inputs. If an apply-time credential first appears on a planned update whose version was known, make a fresh plan with the intended credential before rotating. Resource changes that could advance `version` can make it unknown in the plan, including when the final comparison skips the PUT. Ordinary arguments follow their effective Terraform plan, including ignored changes.

Refresh can observe the webhook version and header names, but cannot read back secret bytes. A lost PUT response leaves Terraform with its prior credential comparison. If a PUT reply identifies another Webhook, Terraform keeps the target’s prior settings and credentials, clears its version, and requires refresh before a further write. A decoded but contradictory response can identify the webhook and its version. Terraform keeps that recovery information and withholds new verifiers. A previous verifier survives only while it still identifies the same redacted credential. An explicit null username with no readable password from the target clears the Basic verifier even if the plan retained credentials. When the PUT response omits an ordinary password or secret header value, Terraform can retain the planned value if the response identifies the same username or the same exact-name secret header. A migration from ordinary to write-only storage can then lose its old plaintext baseline and require another PUT. Reverting to the previous inputs can produce a no-op even if Contentful installed the new values.

To re-establish agreement, coordinate with the receiver, choose fresh values different from the last acknowledged ones for the affected credentials, and apply a new plan. An existing secret header can be preserved by exact name without its bytes. Check delivery at the receiver. After a lost creation response, locate the webhook before creating another one. If a contradictory creation reply identified the webhook, Terraform can mark it tainted. Review the replacement in the next plan before applying.

A failed migration can leave no record of a newly attempted write-only path. If a later read also violates Contentful's redaction contract and returns that path as readable ordinary data, Terraform can report it as drift and store it. Check the remote header's secret flag before refreshing after such a failure. Keep state and saved plans protected during recovery.

## App and extension parameters

Use `jsonencode` with sensitive Terraform inputs for credentials. A `Secret` declaration controls masking in Contentful. In an [App Installation](../resources/app_installation), omitting the parameter object can clear it. In a [UI Extension](../resources/extension), omission retains the installation values returned by Contentful, including after import.

## App Installation secrets

With a personal access token, Contentful masks Secret values with asterisks. When a top-level parameter comes back as a string of asterisks, Terraform can retain its planned value after a write or its previous value on refresh. An ordinary string changed to literal asterisks can look like a masked secret too. [Contentful Secret installation parameters](https://www.contentful.com/developers/docs/extensibility/app-framework/app-parameters/#secret-installation-parameters)

### Rotation and changes outside Terraform

If someone rotates `apiKey` in Contentful, refresh can keep the old key in state. Changing `editor` or a `marketplace` acknowledgment then sends the whole object and can restore the old key, even with `ignore_changes = [parameters["apiKey"]]`. See [Updating parameters](../resources/app_installation#updating-parameters) for how the provider writes the planned map.

### Importing an installation

Import retains the masks that Contentful returns. Supply the actual values before applying. Sending the exact current mask can preserve the existing secret, even when you meant to set a literal string of asterisks. A stale mask can become the literal secret.

### Recovering after a failed write

Inspect the installation before retrying a failed write. Contentful may have accepted it. If Terraform can read the returned parameters but they differ from the plan after masking is handled, it reports the mismatch and keeps the returned map in state. If Terraform cannot read the parameters, a failed create can leave an installation in Contentful without a corresponding resource in state. Once the installation is readable, import it before applying again. A mask alone cannot confirm which secret Contentful stored.

## Personal access tokens

[`contentful_personal_access_token`](../resources/personal_access_token) stores the secret `token` returned at creation. Contentful does not return that value again, so refresh preserves the known value and import leaves it null.

If a create response was lost and no consumer has the credential, identify and revoke the unusable token and issue a new one for the consumer. If consumers already have the credential, [import the existing token](../resources/personal_access_token#import) to manage it without interrupting them.

## App keys

[`contentful_app_key`](../resources/app_key) manages public JWK material that you supply. The corresponding private key is not sent to Contentful and is not stored by this resource.

Changing the public JWK replaces the App Key. Keep the corresponding private key wherever the app signs its requests.
