# Terraform provider documentation practices

Provider documentation should help practitioners predict what Terraform will
configure, change, preserve, and delete. Contributor documentation should explain
how the provider implements and verifies those contracts. This document records
the primary-source basis for that distinction and the editorial practices used
in this repository. The authoring and verification commands remain in
[Development](../../DEVELOPMENT.md#documentation-authoring).

## Organize around the reader's task

HashiCorp recommends an overview, example, and configuration reference for the
provider; resource pages describe their purpose, show usage, and explain inputs
and outputs. Additional sections can cover import and timeouts. These are
content expectations, not a requirement to replace the schema headings produced
by the repository's generator.
See [HashiCorp's documentation structure](https://developer.hashicorp.com/terraform/registry/providers/docs#headers).

The [AWS contributor index](https://github.com/hashicorp/terraform-provider-aws/blob/d987ba0fd9ec754b958ca7cb18f63d1fd3522a06/docs/index.md)
identifies developers as its audience and organizes contribution work into
steps. Its separate practitioner reference demonstrates how a large provider
keeps implementation material out of configuration instructions.

For this repository, the useful separation is:

| Reader and task | Best home |
| --- | --- |
| Evaluate the provider and find setup instructions | Repository README and provider overview |
| Configure one resource or data source | Generated Registry reference |
| Complete a workflow spanning resources or commands | Practitioner guide |
| Build, test, or release the provider | Development and release documentation |
| Change lifecycle behavior or evaluate external evidence | Design and research documentation |

This mapping is an editorial recommendation for this repository. Link between
these documents when the reader needs more depth, and keep each contract in its
authoritative source. Lead with what the reader can do; reserve internal
functions, algorithms, and test machinery for contributor material.

## Make setup and authentication explicit

The [AWS provider overview](https://github.com/hashicorp/terraform-provider-aws/blob/d987ba0fd9ec754b958ca7cb18f63d1fd3522a06/website/docs/index.html.markdown#authentication-and-configuration)
lists credential sources in precedence order. This is a useful pattern because
listing supported environment variables alone leaves mixed configurations
ambiguous.

The Contentful overview should identify the API it uses, show provider
configuration, and lead readers to authentication instructions. Attribute
descriptions should state environment fallbacks and override behavior wherever
the implementation supports them. Keep credentials out of example literals and
link to the provider's state and diagnostics guidance where secrets are used.
Copy the explanatory pattern from AWS, not its authentication mechanisms.

## Describe decisions, not just attribute names

Framework attributes distinguish required, optional, optional-and-computed, and
computed-only values. Their descriptions are consumed by both documentation
generation and editor integrations. Sensitive attributes are generally masked
in output but still stored in state. These distinctions justify precise
descriptions at the schema source.
See [Framework attribute semantics](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/string#configurability),
[descriptions](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/string#description),
and [sensitivity](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/string#sensitive).

For an attribute whose behavior warrants it, explain:

- What the value controls or reports, using Contentful's established name.
- What omission means, including API defaults or observed values.
- Relevant units, formats, supported values, and conflicts.
- Whether a change replaces the resource or affects remote ownership.
- Whether an empty value clears something, and whether it differs from null.
- Any persistence or import consequence that changes how it should be used.

These are review questions, not mandatory sentences for every attribute. Check
answers against schema, conversion, and lifecycle code; an `Optional` label
alone does not establish a default. Keep longer workflows in templates and
guides so schema help remains readable.

## Use examples to explain a complete decision

The [Random password template](https://github.com/hashicorp/terraform-provider-random/blob/36122169cd489f1e8c2799c42e6be2caaa4db81c/templates/resources/password.md.tmpl)
includes its ordinary configuration from an example file, then uses focused
examples to explain import limitations and how configuration can avoid a
subsequent replacement. It demonstrates that a good example explains the next
operation, not only valid syntax.

For Contentful reference snippets, name external prerequisites such as an
existing space, enabled locales, or an activated Content Type. Keep configuration
addresses and IDs consistent with corresponding import examples. For a complete
workflow, include the provider setup, variables, commands, and expected outcome
needed to run it. Label alternatives so readers do not combine several import
forms into one configuration. Prefer `jsonencode` for structured JSON values
when that matches the provider's schema.

There is no general Registry rule that all examples must omit `terraform`,
`provider`, or `output` blocks. The right boundary depends on whether the
example is a reference snippet or a complete workflow; HashiCorp explicitly
expects a provider configuration example in the overview.
See [the provider overview format](https://developer.hashicorp.com/terraform/registry/providers/docs#index-headers).

## Explain lifecycle and import consequences

The [Google Storage Bucket reference](https://github.com/hashicorp/terraform-provider-google/blob/3fdfc360c44ff1d4a4b87cb18d073fbbfdb10f71/website/docs/r/storage_bucket.html.markdown)
describes the default and destructive effect of `force_destroy`, the accepted
import identifiers, version requirements for import forms, and the imported
state value that affects later deletion. These details connect configuration to
observable behavior across operations.

Apply that pattern to Contentful publication, activation, cloning, drift,
import, and destroy behavior. State the triggering condition, remote effect,
Terraform state result, and any recovery action. When an operation may have
committed despite an ambiguous response, tell practitioners how to inspect and
recover the object. Keep that consequence on the resource page; link to design
evidence for the underlying retry or reconciliation algorithm.

## Keep generated sources and navigation coherent

The pinned [tfplugindocs v0.25.0 reference](https://github.com/hashicorp/terraform-plugin-docs/blob/v0.25.0/README.md#conventional-paths)
uses `templates/` for source templates, `examples/` for configuration, and
`docs/` for rendered output. Template filenames omit the provider prefix;
resource and data-source example directories include it. The tool obtains
schema information from the provider and supplies `.SchemaMarkdown` to
templates. Its `tffile` helper includes Terraform files. These conventions
match this repository's [generation directives](../../main.go) and
[dependency pin](../../go.mod).

Edit descriptions, examples, or templates according to the
[authoritative-input map](../../DEVELOPMENT.md#documentation-authoring), then
regenerate. Preserve handwritten design and research files. Do not copy
generated attribute lists into manual prose or add an override template when
the default page already expresses the contract.

Registry guide titles come from front matter, and `subcategory` can group a
large navigation tree. Use recognizable tasks and established domain terms for
titles; introduce grouping only when it improves scanning.
See [Registry navigation](https://developer.hashicorp.com/terraform/registry/providers/docs#navigation-hierarchy).

## Verify accuracy separately from publication format

[tfplugindocs validation](https://github.com/hashicorp/terraform-plugin-docs/blob/v0.25.0/README.md#validate-subcommand)
checks document structure, file size and extensions, front matter, and agreement
between documented object filenames and provider schema. These checks do not
establish prose accuracy, example behavior, or generation freshness. Review
links and rendered pages; validate complete configurations with Terraform;
use focused behavior tests where a documentation claim needs verification.
Follow the repository's [documentation verification workflow](../../DEVELOPMENT.md#documentation-verification).

Registry documentation belongs to a provider version; updating a released
page requires another release. HashiCorp provides a Registry preview tool for
rendering checks. See [documentation publication](https://developer.hashicorp.com/terraform/registry/providers/docs#publishing).
The [provider publishing requirements](https://developer.hashicorp.com/terraform/registry/providers/publishing#creating-a-github-release)
specify `v`-prefixed semantic version tags and signed release assets, including
the manifest. Follow this repository's [release process](../releasing.md)
rather than deriving a second release checklist here.

The provider examples above were inspected as documentation patterns, not run
against AWS or Google Cloud. Their behavior is not evidence of Contentful
behavior. Links to example source are pinned to the inspected commits;
HashiCorp's general documentation links track its maintained guidance.
