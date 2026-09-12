# Terraform provider documentation practices

Use this guide when writing or reviewing repository and Registry documentation.
Practitioner documentation should help readers predict what Terraform will
configure, change, preserve, and delete. Contributor documentation should explain
how the provider implements and verifies those contracts. Authoring and
verification commands live in
[Development](../../DEVELOPMENT.md#documentation-authoring).

## Organize around the reader's task

| Reader and task | Best home |
| --- | --- |
| Evaluate the provider and find setup instructions | Repository README and provider overview |
| Configure one resource, data source, or list resource | Generated Registry reference |
| Complete a workflow spanning resources or commands | Practitioner guide |
| Build, test, or release the provider | Development and release documentation |
| Change lifecycle behavior or evaluate external evidence | Design and research documentation |

Lead with what the reader can do. Use plain language in setup instructions,
examples, and recovery steps. Keep precise Contentful and Terraform terms where
they distinguish behavior: an environment is different from an environment
alias, for example, and a composite Terraform identifier is different from a
Contentful system ID. Reserve internal functions, algorithms, and test machinery
for contributor documentation.

Keep each contract in its authoritative source and link to it when another page
needs more detail. Practitioner pages should explain the consequence without
requiring readers to understand the implementation. Design notes should state
current invariants and limitations; research notes should distinguish published
contracts from dated observations and interpretations.

## Make setup and authentication explicit

The provider overview should identify the API it uses, show provider
configuration, and link to authentication instructions. Describe credential
precedence so a reader knows which value wins when both configuration and an
environment variable are set. Keep credentials out of example literals and
explain state persistence where a resource handles secrets.

An overview needs enough setup to get started. A resource reference can assume
that setup and focus on its own prerequisites. HashiCorp recommends an overview,
example, and configuration reference for the provider, with usage, inputs, and
outputs for each resource. Retain the schema headings supplied by the generator.
See [HashiCorp's documentation structure](https://developer.hashicorp.com/terraform/registry/providers/docs#headers).

## Describe decisions, not just attribute names

Write short attribute contracts in schema descriptions. Explain the details
that change how a reader uses the attribute:

- What the value controls or reports, using the established Contentful name.
- What omission means, including defaults or values learned from the API.
- Units, formats, supported values, and conflicts.
- Whether a change replaces the resource or changes what Terraform manages.
- Whether an empty value clears something and how it differs from null.
- State, refresh, or import consequences.

Check each claim against schema, conversion, and lifecycle code. `Optional`
alone does not establish a default, and `Sensitive` does not mean a value is
absent from state. Do not describe a password omitted by the API as a Terraform
write-only argument unless its schema actually uses `WriteOnly`.
See [Framework attribute semantics](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/string#configurability)
and [sensitivity](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/string#sensitive).

Keep longer workflows in templates and guides so schema help remains readable.
Schema descriptions also appear in editor integrations; they should make sense
without the surrounding generated page.

## Make examples usable

Name prerequisites such as an existing space, enabled locales, or an activated
Content Type. Keep resource addresses and IDs consistent with the corresponding
import examples. Label alternative import forms so readers do not combine them
into one configuration. Prefer `jsonencode` for structured JSON strings when
that matches the schema.

Distinguish a reference snippet from a complete workflow. A complete workflow
needs provider setup, variables, commands, and an expected outcome. A reference
snippet needs enough context to fit it into an existing configuration. The
repository's example conventions are in
[Documentation authoring](../../DEVELOPMENT.md#documentation-authoring);
HashiCorp's [provider overview format](https://developer.hashicorp.com/terraform/registry/providers/docs#index-headers)
expects a provider configuration example.

## Explain lifecycle and import consequences

For publication, activation, cloning, drift, import, and destroy, state the
triggering condition, remote effect, Terraform state result, and available
recovery action. When an operation may have committed despite an ambiguous
response, explain how to inspect and recover the object. Keep the actionable
consequence on the resource page and link to design evidence for the underlying
retry or reconciliation behavior.

An import section should distinguish the Contentful object ID from the composite
Terraform ID and resource identity attributes. Explain any values that the API
cannot recover during import and what configuring them later will do. State
Terraform version requirements next to version-dependent syntax.

## Maintain generated sources and navigation

Edit descriptions, examples, or templates according to the
[authoritative-input map](../../DEVELOPMENT.md#documentation-authoring), then
regenerate. Preserve handwritten design and research files. Do not copy
generated attribute lists into manual prose or add an override template when
the default page already expresses the contract.

The pinned [tfplugindocs reference](https://github.com/hashicorp/terraform-plugin-docs/blob/v0.25.0/README.md#conventional-paths)
documents the `templates/`, `examples/`, and `docs/` conventions used by this
repository. The [generation directives](../../main.go) and
[dependency pin](../../go.mod) determine the local workflow.

Use recognizable tasks and established domain terms for page titles. Registry
guide titles come from front matter; use `subcategory` only when grouping makes
the navigation easier to scan.
See [Registry navigation](https://developer.hashicorp.com/terraform/registry/providers/docs#navigation-hierarchy).

## Verify the claims and the rendered result

[tfplugindocs validation](https://github.com/hashicorp/terraform-plugin-docs/blob/v0.25.0/README.md#validate-subcommand)
checks publication structure and agreement with the provider schema. It does
not establish prose accuracy, working examples, or generation freshness.
Review links and rendered pages, validate complete configurations with Terraform,
and use focused behavior tests when a documentation claim needs verification.
Follow the [documentation verification workflow](../../DEVELOPMENT.md#documentation-verification).

Registry documentation is versioned with the provider. Updates to a released
page require another release; the Registry preview tool can check rendering
before publication. See [documentation publication](https://developer.hashicorp.com/terraform/registry/providers/docs#publishing)
and the repository's [release process](../releasing.md).

## Supporting provider examples

These pinned examples informed the practices above. They illustrate documentation
structure, not Contentful behavior, and were not run against their services.

| Practice | Example |
| --- | --- |
| Separate contributor tasks from practitioner reference | [AWS contributor index](https://github.com/hashicorp/terraform-provider-aws/blob/d987ba0fd9ec754b958ca7cb18f63d1fd3522a06/docs/index.md) |
| Explain credential precedence | [AWS provider overview](https://github.com/hashicorp/terraform-provider-aws/blob/d987ba0fd9ec754b958ca7cb18f63d1fd3522a06/website/docs/index.html.markdown#authentication-and-configuration) |
| Connect examples and import to the next operation | [Random password template](https://github.com/hashicorp/terraform-provider-random/blob/36122169cd489f1e8c2799c42e6be2caaa4db81c/templates/resources/password.md.tmpl) |
| Explain destructive options and import consequences | [Google Storage Bucket reference](https://github.com/hashicorp/terraform-provider-google/blob/3fdfc360c44ff1d4a4b87cb18d073fbbfdb10f71/website/docs/r/storage_bucket.html.markdown) |

HashiCorp's general documentation links track its maintained guidance. Recheck
external requirements when changing the generator or release workflow.
