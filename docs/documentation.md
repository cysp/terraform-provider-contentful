# Writing provider documentation

Edit the [authoritative inputs](../DEVELOPMENT.md#documentation-authoring), then
[generate and validate](../DEVELOPMENT.md#documentation-verification) the Registry
pages. The repository uses the pinned [tfplugindocs conventions](https://github.com/hashicorp/terraform-plugin-docs/blob/v0.25.0/README.md#conventional-paths)
and [Registry page structure](https://developer.hashicorp.com/terraform/registry/providers/docs#headers).

## Content

Write for the task: configuration and recovery on resource pages, cross-resource
workflows in guides, implementation contracts in `docs/design/`, and external API
evidence in `docs/research/`. The README introduces the provider and directs
readers to the right reference.

Schema descriptions should explain a value's purpose, format, defaults, and
Contentful-specific omission, replacement, or import behavior. They also appear
in editor help, so retain essential constraints there. Put longer procedures in
templates. Do not repeat generated attribute lists or teach standard Terraform
syntax and lifecycle behavior; link to HashiCorp when that context is needed.

Use established Contentful terminology and plain, direct sentences. Keep a note
when it affects a configuration choice, expected remote effect, or recovery
step. Omission that deletes data, unreadable imported secrets, and automatic
publication warrant explicit guidance. Implementation chronology, audit
inventories, and repeated instructions to review a plan do not.

## Examples

Keep HCL in `examples/` and include it with `tffile`. When adding the first
include to a plain Markdown guide under `templates/guides/`, rename the source
from `.md` to `.md.tmpl` so the generator expands it. Resource and data-source
examples are snippets for an already configured provider. Use workflow guides
for complete setup or multiple dependent resources. Keep IDs and resource
addresses consistent with all import examples; import forms are alternatives.
Use synthetic values and `jsonencode` for JSON attributes. State prerequisites
that matter, such as an activated Content Type and enabled locale.

## Templates and links

`templates/resources.md.tmpl` owns the common resource layout. Resource-specific
templates are complete pages because the pinned generator has no shared partials;
keep their example, schema, and import sections aligned with the common template.
Use level-two headings for tasks readers need in Registry navigation, and retain
the generated schema hierarchy and identity schema.

Registry pages use extensionless relative links, such as `../resources/entry`,
to stay within the selected provider version. Check template links from the
*generated page's* location. Repository-only documents use `.md` links.

## Verification

Check claims against implementation and primary API evidence, retaining the
scope and dates of service observations. The Registry validator checks filenames,
front matter, and schema coverage; it cannot establish prose accuracy, example
behavior, link destinations, or generation freshness.

Validate changed examples with the local provider. Check representative pages
in the [Registry preview tool](https://registry.terraform.io/tools/doc-preview),
including long schemas, import alternatives, tables, and navigation. Documentation
is versioned with the provider; publishing changed Registry pages requires a
[release](releasing.md).
