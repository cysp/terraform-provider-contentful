# Provider design and evidence

These documents are for maintainers changing provider behavior or test fixtures.
For build and verification commands, use [Development](../../DEVELOPMENT.md). For
practitioner documentation, start with the [provider page](../index.md).

## Design contracts

| Change | Read |
| --- | --- |
| Schemas, planning, request conversion, response projection, or state publication | [Terraform value semantics](terraform-value-semantics.md) |
| HTTP retries, deadlines, or mutation recovery | [Contentful HTTP retry policy](contentful-http-retry-policy.md) |
| CMA test-server behavior or tests that rely on its defaults | [CMA test-server conformance boundaries](../research/cma-test-server-conformance.md) |

The value-semantics note starts with shared rules and includes a
[resource reading map](terraform-value-semantics.md#reading-map). Apply each named
resource contract to that resource; its special cases do not establish a general
provider policy.

## Endpoint evidence

The research notes retain published sources, sanitized observations, and their
limitations. Use them to understand why a design contract exists and what needs
to be checked when the API or client changes.

| Concern | Evidence |
| --- | --- |
| Entry and Content Type request selection | [PUT headers](../research/entry-and-content-type-put-headers.md) |
| Entry publication, field values, and deletion | [Publication lifecycle](../research/entry-publication-lifecycle-evidence.md), [null and omission](../research/entry-null-and-omission.md), [unpublish versions](../research/entry-unpublish-version.md), and [destroy lifecycle](../research/entry-destroy-lifecycle.md) |
| Taxonomy locking | [Taxonomy versions](../research/taxonomy-version.md) |
| Credentials | [App signing secret](../research/app-signing-secret.md) and [Delivery API key environments](../research/delivery-api-key-environments.md) |
| App Event subscriptions, App Actions, deployment, and related resources | [App Framework resource model, APIs, data, and behavior](../research/app-event-resources.md) |
| Extensions and Space Enablements | [Extension source values](../research/extension-source-values.md) and [Space Enablements values](../research/space-enablement-values.md) |
| Content preview | [Preview environments](../research/content-preview-environments.md) and [Live Preview variables](../research/live-preview-variables.md) |
| Webhooks, Editor Interfaces, collections, and shared mock behavior | [CMA test-server conformance boundaries](../research/cma-test-server-conformance.md) |

## Evidence boundaries

Design contracts describe intended provider behavior. External-behavior claims
must retain their provenance: primary Contentful documentation, direct
observations under `docs/research/`, test-server behavior, or unresolved
assumptions. Mock behavior alone does not establish Contentful behavior.

An observation date records when an experiment ran; a documentation edit does
not refresh that evidence. Preserve the observed scope, source revisions, and
unverified cases. Keep historical provider results clearly identified so they
cannot be mistaken for current behavior.

The [CMA test-server conformance reference](../research/cma-test-server-conformance.md)
connects external evidence to fake behavior and test coverage. Deliberate faults
belong in explicit adversarial cases rather than undocumented default behavior.
