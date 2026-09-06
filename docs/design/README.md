# Provider design documentation

## Reading map

- [Terraform value semantics](terraform-value-semantics.md) defines provider-wide
  handling of Terraform values, request conversion, response projection, state
  publication, and plan consistency. Its named resource sections contain the
  narrower contracts for taxonomy, Webhook credentials, publication/activation
  recovery, and mutation reconciliation; apply those sections to the affected
  resource rather than generalizing their special cases.
- [Contentful HTTP retry policy](contentful-http-retry-policy.md) defines request
  deadlines and retry boundaries. Resource mutations can narrow the general
  policy when replay would be unsafe.

## Evidence boundaries

Design contracts describe intended provider behavior. External-behavior claims
must retain their provenance: primary Contentful documentation, direct
observations under `docs/research/`, test-server behavior, or unresolved
assumptions. Mock behavior alone does not establish Contentful behavior.

The [CMA test-server conformance matrix](../research/cma-test-server-conformance.md)
links endpoint-specific provider contracts, primary sources, sanitized
observations, fake behavior, test coverage, and evidence gaps. Follow those
sources when changing an affected contract; preserve their stated limitations.
