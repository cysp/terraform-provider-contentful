# Write-only arguments and ephemeral secret use cases

This note records where the provider uses or could use Terraform write-only arguments or ephemeral resources to reduce secret persistence in plan and state artifacts. The app signing secret and webhook header decisions below are implemented; the personal access token discussion remains exploratory.

## Sources

- Terraform Plugin Framework write-only arguments: <https://developer.hashicorp.com/terraform/plugin/framework/resources/write-only-arguments>
- Terraform Plugin Framework ephemeral resources: <https://developer.hashicorp.com/terraform/plugin/framework/ephemeral-resources>
- Terraform language ephemeral blocks: <https://developer.hashicorp.com/terraform/language/ephemeral>
- Contentful Personal Access Tokens: <https://www.contentful.com/help/token-management/personal-access-tokens/>
- Local Contentful OpenAPI schemas under `internal/contentful-management-go/openapi/schemas/`
- Local provider schemas under `internal/provider/`

## Current provider surface reviewed

The provider currently registers these managed resources:

- `contentful_app_definition`
- `contentful_app_signing_secret`
- `contentful_app_installation`
- `contentful_content_type`
- `contentful_delivery_api_key`
- `contentful_editor_interface`
- `contentful_environment_alias`
- `contentful_environment`
- `contentful_entry`
- `contentful_extension`
- `contentful_personal_access_token`
- `contentful_resource_provider`
- `contentful_resource_type`
- `contentful_role`
- `contentful_space_enablements`
- `contentful_tag`
- `contentful_team`
- `contentful_team_space_membership`
- `contentful_webhook`

The provider currently registers these data sources:

- `contentful_app_definition`
- `contentful_environment_status_ready`
- `contentful_marketplace_app_definition`
- `contentful_preview_api_key`

The provider currently registers these list resources:

- `contentful_content_type`
- `contentful_entry`

## Terraform semantics

Write-only arguments are for managed resource attributes that a practitioner can configure but Terraform does not persist to plan or state. Terraform Plugin Framework documentation says write-only arguments are supported in Terraform 1.11 and later, should be used for secrets such as passwords or API keys that do not need to be persisted, and are only available from configuration.

Important constraints:

- Write-only arguments are resource arguments, not data source outputs.
- Write-only values are not available from prior state during read.
- Write-only arguments cannot be used with set attributes, set nested attributes, or set nested blocks.
- A provider should be the terminal consumer of a write-only value: it should use the value in an API call or ignore it.

Ephemeral resources are different. Terraform opens an ephemeral resource during an operation and guarantees its result data is not persisted in plan or state. The language docs describe ephemeral blocks as suitable for sensitive or temporary data that should not persist. The framework docs describe ephemeral resources as external data references whose data is not persisted.

Important constraints:

- Ephemeral results can only be used in ephemeral contexts.
- Ephemeral resources are not managed resources; they do not have durable Terraform state.
- An ephemeral resource that creates a remote object must either leave an intentionally unmanaged object behind or clean it up using the ephemeral lifecycle. That is a product decision, not just a schema choice.

## Strong write-only candidate: app signing secret value

### Current provider behavior

`contentful_app_signing_secret` supports the existing sensitive `value` argument and a compatibility-preserving write-only alternative:

```go
"value": schema.StringAttribute{
	Description: "The symmetric key shared between Contentful and an app backend. Must be exactly 64 characters long and contain only alphanumeric characters (a-z, A-Z, 0-9).",
	Optional:    true,
	Sensitive:   true,
},
"value_wo": schema.StringAttribute{
	Description: "Write-only alternative to value. The symmetric key shared between Contentful and an app backend. Must be exactly 64 characters long and contain only alphanumeric characters (a-z, A-Z, 0-9).",
	Optional:    true,
	Sensitive:   true,
	WriteOnly:   true,
},
```

Source: `internal/provider/resource_app_signing_secret_schema.go`.

The local Contentful OpenAPI response schema for app signing secrets returns `redactedValue`, not the raw secret:

```yaml
AppSigningSecretData:
  type: object
  properties:
    redactedValue:
      type: string
  required:
    - redactedValue
```

Source: `internal/contentful-management-go/openapi/schemas/app-signing-secret/data.yml`.

### Evaluation

This is the cleanest write-only use case.

The secret is supplied by the practitioner and sent to Contentful. Contentful does not return the raw value on read; it returns a redacted value. Keeping the raw value in Terraform state is therefore useful only for Terraform drift comparison, not because the API can verify the value.

A write-only replacement or companion argument would align with Terraform's intended use case for secrets that the provider consumes in an API request.

### Implemented decision

The provider keeps `value` for compatibility and adds `value_wo`. Exactly one must be configured. Because Terraform cannot diff a write-only value against state, the provider stores a salted, path-separated, bounded Argon2id verifier in provider-private state. An unchanged value therefore plans no action, while a changed or unknown value marks the resource for update without persisting plaintext.

## Possible write-only candidate: secret webhook header values

### Current provider behavior

Contentful webhook headers are modeled with `key`, `value`, and `secret` in the local OpenAPI schema:

```yaml
WebhookDefinitionHeader:
  type: object
  properties:
    key:
      type: string
    value:
      type: string
    secret:
      type: boolean
  required:
    - key
```

Source: `internal/contentful-management-go/openapi/schemas/webhook-definition/headers.yml`.

The provider currently models webhook headers as a map of nested objects, each containing `value` and `secret`.

On read, the provider preserves a legacy configured value when Contentful omits the value for a secret header. A null existing value acts as the write-only confidentiality marker and is not replaced even if an API response echoes plaintext or omits the optional `secret` flag:

```go
if existingHeaderValue, existingHeaderValueOk := existingHeaderValue.GetValue(); existingHeaderValueOk {
	value.Value = existingHeaderValue.Value
	value.Secret = existingHeaderValue.Secret
	existingValueIsNull = existingHeaderValue.Value.IsNull()
}

if headerValue, ok := header.Value.Get(); ok && !existingValueIsNull {
	value.Value = types.StringValue(headerValue)
} else if !headerIsSecret {
	value.Value = types.StringNull()
}
```

Source: `internal/provider/webhook_header_value_response.go`.

### Evaluation

Secret webhook headers are a real state-exposure concern. The current state-preservation behavior is pragmatic for diff stability, but it means a secret header value remains in Terraform state if the user configured it.

A write-only path would better match Contentful's secret-header semantics when `secret = true`.

The nested design is invalid for this resource: `headers` is Optional+Computed, and the Plugin Framework rejects a Computed nested attribute containing a write-only child. Removing `Computed` would break the existing remote-state lifecycle contract.

### Implemented decision

The provider therefore keeps `headers` Optional+Computed and adds a parallel top-level `header_values_wo` map keyed by the same header names. A write-only map entry is accepted only when the matching header has `secret = true`; unmatched or nonsecret entries fail before an API request. Legacy `headers.<name>.value` remains available for compatibility and nonsecret values.

## Potential ephemeral resource: create a personal access token

### Current provider behavior

`contentful_personal_access_token` is a managed resource. Its schema exposes:

- `name` as required and replacement-forcing
- `expires_in` as optional and replacement-forcing
- `scopes` as required and replacement-forcing
- `token` as computed and sensitive

The schema description for `token` already documents the one-time nature of the value:

```go
"token": schema.StringAttribute{
	Description: "The access token for the Content Management API. This is only available immediately after creation.",
	Computed:    true,
	Sensitive:   true,
},
```

Source: `internal/provider/resource_personal_access_token_schema.go`.

Contentful's public PAT documentation says the create response contains the generated access token and that this is the only time it is shown. The same documentation says later GET endpoints return name, scope, and metadata, but not the token itself.

### Why this is not a write-only argument

The PAT token value is generated by Contentful. The practitioner does not provide the token value to the provider.

Write-only arguments solve the inverse problem: a practitioner provides a secret value and the provider sends it to the remote API without persisting it. A generated PAT's secret value is a computed result. Terraform write-only arguments cannot represent computed output values, and the framework docs state that write-only argument values are only available in configuration.

Therefore, `contentful_personal_access_token.token` is not a good write-only argument candidate.

### Why an ephemeral PAT resource could make sense

An ephemeral PAT resource could create the PAT during Terraform execution, return the generated token only as ephemeral result data, and avoid storing the token in plan or state.

The one-time retrieval behavior is not a blocker. It is the reason an ephemeral resource may be the right abstraction:

1. Terraform opens the ephemeral resource during the operation.
2. The provider calls Contentful's create PAT endpoint.
3. Contentful returns the generated token in that create response.
4. The provider returns the token as ephemeral result data.
5. Terraform allows that token only in ephemeral contexts and does not persist it to plan or state.

The provider does not need to retrieve the token later, because an ephemeral resource has no durable state that must be refreshed with the token value on future plans. Future operations that need a token would open the ephemeral resource again and create a new PAT.

### Proposed shape

A possible ephemeral resource could be named one of:

- `contentful_personal_access_token`
- `contentful_generated_personal_access_token`
- `contentful_temporary_personal_access_token`

Configuration:

- `name`
- `scopes`
- `expires_in`

Result data:

- `id`
- `token`
- `expires_at`

Lifecycle:

- Open: create the PAT and return `id`, `token`, and metadata.
- Renew: probably unsupported unless Contentful adds a token-extension API. If the token can expire during a long operation, users should set `expires_in` long enough for the run.
- Close: revoke the PAT if an `id` is available.

### Main risk

The close behavior is the hard part.

If close revokes the PAT, the resource is truly temporary and matches Terraform's ephemeral model. If close does not revoke it, each Terraform operation that opens the ephemeral resource can leave behind a real Contentful PAT that is not managed by Terraform state. That is operationally risky unless the PAT is short-lived and explicitly documented as unmanaged after creation.

Because Contentful supports expiry dates for PATs, `expires_in` should probably be required or have a strict maximum for an ephemeral PAT resource. A no-expiry ephemeral PAT would be a poor default because a failed Terraform run could leave a long-lived unmanaged credential behind.

### Useful cases

An ephemeral PAT resource could be useful when a Terraform operation needs a newly minted Contentful CMA token only during the same run, for example:

- Passing a generated CMA token into another provider's ephemeral provider configuration.
- Sending a token into a write-only argument on another resource.
- Bootstrapping an external integration that accepts a token during apply and stores it outside Terraform state.

It would not be useful for ordinary managed resources that must persist their arguments in Terraform state. Terraform will reject ephemeral values in non-ephemeral contexts.

## Non-candidates or weak candidates

### Delivery API key access token

`contentful_delivery_api_key.access_token` is computed and sensitive. The local Contentful OpenAPI schema models delivery API key `accessToken` as response data:

```yaml
ApiKeyData:
  allOf:
    - $ref: "./request-data.yml#/ApiKeyRequestData"
    - type: object
      properties:
        accessToken:
          type: string
        preview_api_key:
          $ref: "../links/preview-api-key-link.yml#/PreviewAPIKeyLink"
      required:
        - accessToken
```

Source: `internal/contentful-management-go/openapi/schemas/api-key/data.yml`.

Because Contentful returns this token as normal readable API data, it is weaker as an ephemeral resource candidate. Keeping it as sensitive computed resource state is consistent with normal Terraform behavior, unless the provider adopts a stricter "no API token in state" policy.

### Preview API key data source access token

`contentful_preview_api_key.access_token` is computed and sensitive. The local Contentful OpenAPI schema models preview API key `accessToken` as response data:

```yaml
PreviewApiKeyData:
  type: object
  properties:
    name:
      type: string
    description:
      type: string
      nullable: true
    accessToken:
      type: string
    environments:
      type: array
      items:
        $ref: "../links/environment-link.yml#/EnvironmentLink"
  required:
    - name
    - accessToken
```

Source: `internal/contentful-management-go/openapi/schemas/preview-api-key/data.yml`.

This is a state-exposure concern, but not a Contentful one-time retrieval problem. An ephemeral variant could be justified only by a provider policy that readable API tokens should not be available through normal data sources.

### App definitions, marketplace app definitions, environment readiness, content types, entries, and other configuration resources

These do not show the same one-time secret or practitioner-supplied secret semantics in the reviewed provider schemas. They are not good write-only or ephemeral candidates based on the evidence reviewed here.

## Implementation status and next step

1. App signing secret write-only support is compatibility-preserving and implemented.
2. Webhook header write-only support uses a validated parallel-map schema and is implemented.
3. Ephemeral PAT support remains a separate feature requiring an explicit lifecycle decision: revoke on close, require bounded expiry, or do not implement.

## Decisions needed

The remaining decisions concern ephemeral resources:

1. For ephemeral PATs, should the provider revoke the token on close?
2. Should an ephemeral PAT require `expires_in` and enforce a maximum TTL?
3. Should readable API tokens, such as delivery and preview API keys, remain normal sensitive values or get separate ephemeral variants under a stricter no-token-state policy?
