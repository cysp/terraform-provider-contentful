# Locale management

Locale uses the shared [Terraform value semantics](terraform-value-semantics.md)
and [HTTP retry policy](contentful-http-retry-policy.md). Practitioner workflows
belong in the [resource reference](../resources/locale.md); Contentful contracts
and observations belong in [Locale research](../research/locales.md).

## Mutation state and version

Create and Update send the effective Plan, including defaults and values retained
by `ignore_changes`. Omitting `fallback_code` sends an explicit null to clear the
fallback. Contentful validates codes, fallback graphs, default restrictions, and
allowances; the provider does not prefetch dependencies or coordinate separate
resources.

Mutation reconciliation retains endpoint identity, returned configuration, and
private `version` before reporting plan contradictions. Update sends the saved
version and records the returned one, which may be unchanged for a no-op. Delete
is unversioned and accepts 404.

The shared client keeps `sys.version` optional for discovery compatibility.
Managed Create/Read/Update report its absence as an error and clear the response's
private token rather than inventing a version. Terraform can persist mutation
recovery state; a [failed Read retains prior persisted state](https://developer.hashicorp.com/terraform/plugin/framework/resources/read#caveats)
and blocks normal planning until refresh succeeds.

CMA is the resource's authority. The provider does not rewrite code-dependent
content configuration, verify CDA/CPA convergence, or implement quota preflights.
[Observed propagation discrepancies](../research/locales.md#cross-api-propagation)
do not supply a reliable delivery waiter.

## Discovery

The list resource shares resource identity/schema and response projection, using
the existing streaming pagination helper without `order` or local sorting.
Terraform's result limit and consumer cancellation stop traversal. No ordering
or atomic-snapshot guarantee is added. Both flags concern content availability;
disabled Locale entities remain discoverable. Existing single/plural data sources
retain their contracts, including the plural source's ID ordering.
