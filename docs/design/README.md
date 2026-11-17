# Provider design

Read the contract for the behavior you are changing. Commands live in
[Development](../../DEVELOPMENT.md); Contentful contracts and observations live in
[API research](../research/README.md).

## Design contracts

| Concern | Contract |
| --- | --- |
| Schemas, request conversion, response projection, identity, and state | [Shared value semantics](terraform-value-semantics.md) |
| HTTP retries and deadlines | [HTTP retry policy](contentful-http-retry-policy.md) |
| Content Type activation and Entry publication, recovery, and fields | [Content publication](content-publication.md) |
| Role, Editor Interface, and Webhook comparison rules | [Mutation reconciliation](mutation-reconciliation.md) |
| Taxonomy collection ownership, canonicalization, and locking | [Taxonomy](taxonomy.md) |
| App Definition and Extension validation; App Action identity | [App configuration](app-configuration.md) |
| App Installation parameter masking and complete-object writes | [Secret parameters](app-installation-secret-parameters.md) |
| Preview platforms and live preview variables | [Content preview](content-preview-environments.md) |
| Locale management and discovery | [Locales](locale.md) |
| Data-source identity, aliases, and pagination | [Configuration data sources](configuration-data-sources.md) |
| Webhook signing secret ownership and acknowledgement | [Webhook signing secret](webhook-signing-secret.md) |
| In-process CMA server fixtures and limitations | [Test-server conformance](cma-test-server-conformance.md) |

## Evidence boundaries

Design documents state provider policy. Research records published API contracts,
pinned client behavior, and scoped service observations using the
[research conventions](../research/README.md#evidence-and-redaction).
Mock tests prove behavior against fixtures; they do not establish Contentful
compatibility. Keep deliberate faults separate from normal fixture behavior.

For documentation changes, use [Writing provider documentation](../documentation.md).
