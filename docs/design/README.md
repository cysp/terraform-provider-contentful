# Provider design

## Design contracts

| When changing | Read |
| --- | --- |
| Schemas, planning, request conversion, or state publication | [Terraform value semantics](terraform-value-semantics.md) |
| Entry publication or Content Type activation | [Draft versions and recovery](entry-and-content-type-lifecycle.md) |
| Role, Editor Interface, or Webhook updates | [Mutation reconciliation and version preconditions](mutation-reconciliation.md) |
| Taxonomy Concepts or Concept Schemes | [Collection ownership and version locking](taxonomy.md) |
| App Definitions, App Actions, or Extensions | [App and Extension configuration](app-configuration.md) |
| App Installation parameters | [Parameter masking and complete-object updates](app-installation-secret-parameters.md) |
| App and webhook signing-secret values, comparison, or state publication | [Signing secret values and acknowledgement](signing-secret-write-only.md) and [Signing secret planning](signing-secret-planning.md) |
| Webhook Signing Secrets | [Space-wide signing secret ownership](webhook-signing-secret.md) |
| Preview environments or live preview variables | [Preview platform configuration](content-preview-environments.md) and [variables document ownership](content-preview-environments.md#live-preview-variables) |
| Delivery API key environments or Space Enablements | [Key environment selection](terraform-value-semantics.md#delivery-api-key-environments) and [enablement request values](terraform-value-semantics.md#space-enablements) |
| Locale identity or version handling | [Locale lifecycle](locale.md) |
| Configuration data sources | [Lookup identity, alias context, and pagination](configuration-data-sources.md) |
| HTTP retries, deadlines, or recovery after a failed request | [Contentful HTTP retry policy](contentful-http-retry-policy.md) |
| In-process Contentful API fixtures | [CMA test-server conformance](cma-test-server-conformance.md) |
| Registry documentation | [Documentation practices](provider-documentation.md) |

## Evidence boundaries

Design documents record provider decisions. Use [Contentful API research](../research/README.md) to distinguish published contracts, first-party client behavior, and dated service observations. [Test-server conformance](cma-test-server-conformance.md) records fixture defaults and synthetic failures.
