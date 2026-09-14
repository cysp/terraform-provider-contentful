# Locale management

Locale follows the shared [value semantics](terraform-value-semantics.md) and
[HTTP retry policy](contentful-http-retry-policy.md). Mutation reconciliation
retains endpoint identity, returned values, and private `version` before reporting
plan contradictions. Read reflects remote drift. Update sends the saved version;
Delete is unversioned and treats 404 as success.

The shared response keeps `sys.version` optional to preserve existing discovery
reads. Create and Update retain returned state but clear any old private version
and report an error when the response omits it. Clearing the token prevents a
later update from reusing a version that does not correspond to that response.

Read also reports a missing version as an error. On a failed refresh,
[Terraform retains the prior persisted state](https://developer.hashicorp.com/terraform/plugin/framework/resources/read#caveats)
and private data; it does not save the returned values or the cleared token.
The refresh error stops normal planning and applying until a successful refresh.

Contentful's [Locale reference](https://www.contentful.com/developers/docs/references/content-management-api/locales/)
and [Update reference](https://www.contentful.com/developers/docs/references/content-management-api/locales/update-a-locale/)
document null fallbacks, immutable default status, code/deletion restrictions, and
versioned updates. Practitioner contracts are in the resource schema descriptions.
Mock tests establish provider behavior, not live CMA conformance; the mock omits
localized-content deletion and some service validation rules.
