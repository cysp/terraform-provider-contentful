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
| [Webhook Basic password](#basic-authentication) | Retains prior value when omitted | Null when omitted |
| [Secret webhook header](#secret-custom-headers) | Retains the value at the exact key when omitted | Null when omitted |
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

### Basic authentication

For [`contentful_webhook`](../resources/webhook), Contentful normally omits the HTTP Basic password. When it does, refresh retains the previously managed value and import leaves it null. Terraform records a password returned by Contentful.

On a later update, omitting both `http_basic_username` and `http_basic_password` clears Basic authentication, including on an imported webhook. To keep Basic authentication when updating an imported webhook, supply both credentials. Terraform can read the webhook and report no changes with just the imported username. If an update would send that username alone, the provider reports a missing password before sending the request. `ignore_changes` can retain a previously managed password only when Terraform already has it in state.

### Secret custom headers

Contentful's `secret = true` flag on a custom webhook header controls how Contentful treats the value. Terraform sensitivity is configured separately: supply a sensitive Terraform expression to redact the header value from normal Terraform output. The value can still be present in plan or state data.

Terraform records returned header values. When Contentful omits a secret value, refresh keeps any value already in state under the exact header name, including its case. An external rotation can then go undetected. Import leaves an omitted value null.

Supply an initial `value` for each configured header when creating or replacing a webhook.

When updating a webhook:

- Omit `headers` to keep the existing headers when changing another webhook setting. Terraform resends ordinary header values. For secret headers, it sends the exact name and secret flag so Contentful can retain the current value, including one rotated outside Terraform.
- To manage an existing header’s name while another system rotates its value, use its exact name and set `secret = true`. Omit `value` or set it to `null`. For example, `headers = { "X-Token" = { secret = true }, "X-Format" = { value = "json" } }` keeps the stored token and sets the format header. To rotate the token, set its `value` to the new token.
- To let another system manage the entire secret header while Terraform manages its siblings, omit its key from configuration and add `ignore_changes = [headers["X-Token"]]` for that exact key. On update, Terraform sends the key and secret flag so Contentful retains the stored value, including an external rotation. If configuration still declares the ignored key, Terraform can resend the value retained in its plan.
- Supply a value when adding a header under a new name, including a change in case. When changing an ordinary header to secret, set `secret = true` and supply the value you want Contentful to store. You can use the current value. If the planned value is null, Contentful leaves the header ordinary and Terraform reports the difference.
- A configured map removes headers absent from the plan. To remove all custom headers, set `headers = {}` and remove any `ignore_changes` rule that retains the map or one of its keys.

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
