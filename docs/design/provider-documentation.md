# Writing provider documentation

Use schema descriptions for information someone needs while editing configuration. Put a resource's workflow in its page template and a workflow spanning resources in a guide. The [authoring map](../../DEVELOPMENT.md#documentation-authoring) identifies the source for each page.

## What belongs on a resource page

Lead with the Contentful object and its scope. Explain choices with different remote effects, especially omission versus clearing, replacement of existing objects, publication, and what import can recover. Put the consequence beside the example or attribute that triggers it. Link to HashiCorp for standard Terraform mechanics when readers need that background. Use the Contentful system ID or composite import ID by name where readers could confuse them.

Keep generated schema help useful in an editor. Describe units, defaults, and consequential constraints there. Let the generated headings communicate type and configurability. Link to a guide for shared procedures such as credential rotation or recovery after an uncertain write.

## Make examples usable

Write resource and data-source examples as snippets for an already configured provider. Use a guide for a complete workflow and identify what the reader needs before starting and what a successful run produces. Keep HCL and shell examples in `examples/`. To include HCL in a guide with `tffile`, use a `.md.tmpl` source in `templates/guides/`. The generator copies plain `.md` guides verbatim. Use consistent addresses and IDs across resource, identity import, string-ID import, and CLI examples. When an example depends on another example, link to it and name any required Contentful configuration. Use `jsonencode` for structured JSON strings.

The pinned generator reads complete resource templates. Keep specialized templates aligned with the [common layout](../../templates/resources.md.tmpl), including version labels on import alternatives. [tfplugindocs template conventions](https://github.com/hashicorp/terraform-plugin-docs/blob/v0.25.0/README.md#conventional-paths)

## Verify the claims and the rendered result

Follow [documentation verification](../../DEVELOPMENT.md#documentation-verification). Check descriptions against lifecycle code and examples against the provider schema. Keep empirical Contentful claims tied to the scope of the [API evidence](../research/README.md#evidence-and-redaction).

Use extensionless relative links within Registry pages so readers stay on their selected provider version. Use filenames with extensions in contributor docs for navigation on GitHub. Check Registry links from the generated page's location, including section anchors.

Inspect long pages, tables, and import examples with the [Registry preview](https://registry.terraform.io/tools/doc-preview) when changing their layout. The validator checks document structure and schema coverage. An example still needs Terraform validation, and a link still needs a destination. [HashiCorp provider documentation](https://developer.hashicorp.com/terraform/registry/providers/docs)
