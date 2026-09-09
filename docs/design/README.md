# Provider design and evidence

These documents are for maintainers changing provider behavior or test fixtures.
For build and verification commands, use [Development](../../DEVELOPMENT.md). For
practitioner documentation, start with the [provider page](../index.md).
Contentful's independent API contracts and observations are indexed in
[Contentful API research](../research/README.md).

## Design contracts

| Change | Read |
| --- | --- |
| Schemas, planning, request conversion, response projection, or state publication | [Terraform value semantics](terraform-value-semantics.md) |
| HTTP retries, deadlines, or mutation recovery | [Contentful HTTP retry policy](contentful-http-retry-policy.md) |
| Entry publication recovery | [Entry publication contract](terraform-value-semantics.md#entry-publication-ownership-and-partial-field-ownership) and [Terraform lifecycle evidence](entry-publication-evidence.md) |
| Preview environment representation and requests | [Configuration and reconciliation](content-preview-environments.md) |
| CMA test-server behavior or tests that rely on its defaults | [CMA test-server conformance boundaries](cma-test-server-conformance.md) |
| Practitioner and contributor documentation | [Provider documentation practices](provider-documentation.md) |

The value-semantics note starts with shared rules and includes a
[resource reading map](terraform-value-semantics.md#reading-map). Apply each named
resource contract to that resource; special cases do not establish a general
provider policy.

## Evidence boundaries

API research retains published sources, sanitized observations, and limitations
independently of provider implementation. Use the
[research reading map](../research/README.md#reading-map) to find endpoint
behavior and the [evidence conventions](../research/README.md#evidence-and-redaction)
to distinguish a published contract, first-party client behavior, direct
observation, and interpretation.

Design documents state provider policy; research documents state what the
external evidence establishes. Mock behavior alone does not establish Contentful
behavior. The [test-server conformance reference](cma-test-server-conformance.md)
connects API evidence to fixture behavior and coverage. Deliberate faults belong
in explicit adversarial cases rather than undocumented default behavior.
