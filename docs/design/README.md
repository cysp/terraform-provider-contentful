# Provider design documentation

These notes record current provider contracts that are too detailed for practitioner-facing Registry documentation. Read them as normative implementation guidance unless a section explicitly identifies an observation, experiment, or external source instead.

## Reading map

### Provider-wide invariants

- [Terraform value semantics](terraform-value-semantics.md) defines the provider-wide handling of known, null, and unknown Terraform values; request conversion; response projection; state publication; and plan consistency. Its opening sections are the general rules that new schema and lifecycle work must preserve.
- [Contentful HTTP retry policy](contentful-http-retry-policy.md) defines request deadline budgets and the provider-wide retry boundary for Contentful Management API traffic. Resource lifecycle mutations can deliberately narrow that general retry policy where replay would be unsafe.

### Resource-specific contracts

`terraform-value-semantics.md` also contains named sections for concrete resource families and lifecycle boundaries, including taxonomy ownership/canonicalization, Webhook credentials, publication and activation recovery, mutation reconciliation, and resource-specific decision tables. A section named for a resource or attribute is scoped to that behavior unless it explicitly states a provider-wide rule.

Use those sections when changing the affected resource, but do not generalize their special cases to unrelated resources. For example, omission, explicit empty values, write-only secret recovery, and publication authority have intentionally different contracts in different resources.

## Evidence boundaries

Keep these categories distinct when maintaining design documentation:

- **Provider guarantee:** behavior intentionally enforced by provider code and tests.
- **Documented Contentful behavior:** behavior supported by current primary Contentful documentation.
- **Direct observation:** behavior established by an authorized experiment and recorded under `docs/research/`.
- **Mock behavior:** behavior provided by the repository test server; useful for provider tests but not independent evidence of CMA behavior.
- **Assumption or compatibility boundary:** a dependency that has not been established strongly enough to state as a guarantee.

When external behavior matters to a design decision, link the relevant primary documentation or research record rather than making the design note the sole source of that external fact.
