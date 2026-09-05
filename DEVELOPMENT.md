# Development

This project uses Go-native tooling entrypoints for development workflows.

## Start here

Use the Go version declared by [`go.mod`](go.mod). A Terraform CLI on `PATH` is
required for documentation generation and Terraform acceptance tests. Linting
uses the version recorded in [`.golangci-lint-version`](.golangci-lint-version).

For a first local check of a code checkout:

```sh
go test ./...
go build .
```

Use the narrower checks below while iterating, then expand verification according
to the changed behavior before merge. For Terraform CLI behavior, also test a
local provider build as described in [Local provider build](#local-provider-build).

## Repository guidance

[AGENTS.md](AGENTS.md) is the authoritative repository guidance;
[CLAUDE.md](CLAUDE.md) and [GEMINI.md](GEMINI.md) import it. Keep shared rules
there and command details here. Design contracts belong in `docs/design/`;
link them from the instructions with the conditions that require reading them.
See the [design documentation map](docs/design/README.md) for the distinction
between provider-wide invariants and resource-specific contracts.

## Repository map

- `internal/provider/` contains provider schemas, Terraform lifecycle behavior,
  request conversion, response projection, and provider tests.
- `internal/contentful-management-go/` contains the generated Contentful
  Management API client and its OpenAPI generation inputs.
- `examples/` contains Terraform configuration and import examples consumed by
  `tfplugindocs`.
- `templates/` contains custom practitioner-facing Registry documentation and
  guide inputs consumed by `tfplugindocs`.
- `docs/index.md`, `docs/resources/`, `docs/data-sources/`,
  `docs/list-resources/`, and `docs/guides/` are generated Registry
  documentation. Do not edit them as authoritative sources.
- `docs/design/`, `docs/research/`, and [`docs/releasing.md`](docs/releasing.md)
  are handwritten developer documentation and are not generated Registry
  output.

## Validation scope

Select checks by the changed behavior, using the commands below. When preparing
a code change for merge, inspect the remote checks for that revision;
[`.github/workflows/`](.github/workflows/) defines the current CI matrix.
The matrix does not require replaying every CI job locally. Live acceptance
tests require authorization to use the real Contentful account; available
credentials alone do not establish that authorization.

| Change | Local verification |
| --- | --- |
| Agent instructions or prose only | Review instruction consistency, links, and the final diff; run `git diff --check`. |
| Go behavior | Run tests for the affected packages and the lint and format checks. Use `go test ./...` and `go build .` for shared behavior or when preparing for merge. |
| Terraform planning, state, or lifecycle | Run focused mocked acceptance tests for the affected transitions in addition to the Go checks; extend coverage where existing tests do not establish the changed behavior. |
| Schema, examples, templates, OpenAPI, or other generation inputs | Follow the [generation requirement](AGENTS.md#documentation-and-workflow) using [Code Generation](#code-generation), then follow [Documentation verification](#documentation-verification), plus checks for the affected behavior. |
| A claim about live Contentful behavior | Use primary documentation or an authorized live experiment; mocked tests establish provider behavior against the fixture, not CMA conformance. |

If a check cannot run, record the command and concrete blocker. Distinguish a
passing check from a skipped test or an environment failure. Local verification
does not establish that remote CI passed.

## Documentation authoring

Practitioner-facing Registry documentation is generated with
`terraform-plugin-docs`. Change the authoritative input for the kind of
information being documented, then regenerate and review the rendered output:

| Documentation concern | Authoritative input |
| --- | --- |
| Provider, resource, data-source, list-resource, and attribute contracts | Schema descriptions under `internal/provider/` |
| Terraform configuration and import syntax | `examples/` |
| Resource-specific workflows, warnings, and additional narrative | `templates/` |
| Practitioner guides | `templates/guides/` |
| Provider design contracts and external-behavior evidence | `docs/design/` and `docs/research/` |
| Release process | `docs/releasing.md` |

`tfplugindocs` rebuilds its managed Registry-documentation outputs; generation
is not limited to replacing schema fragments inside otherwise handwritten
files. Do not hand-edit generated pages to make a durable change, and do not
remove the entire `docs/` tree because it also contains handwritten developer
documentation.

Resource-page examples are normally reference snippets rather than standalone
Terraform modules. Keep them small, but make each resource directory internally
consistent: when the configuration example declares one resource address, its
identity import, string-ID import, and CLI import examples should target that
same address unless the difference is intentional and explained. Workflow guides
are the appropriate place for complete setup and multi-step examples.

When documenting identifiers, distinguish the Terraform resource identifier
from a Contentful system ID whenever both concepts exist. In particular, do not
use a generic "ID of this resource" description when `id` is composite and a
separate `*_id` attribute is the Contentful object ID.

### Practitioner writing conventions

- Open with what the provider manages, reads, lists, or waits for.
- Explain Terraform-visible behavior and required practitioner action before
  implementation mechanisms.
- State warnings as condition, consequence, and available action when an action
  exists.
- Preserve established Terraform and Contentful terminology and exact API
  identifiers.
- Keep short contracts in schema descriptions, multi-step workflows in templates
  or guides, and algorithms or evidence details in developer documentation.
- Distinguish provider guarantees, documented Contentful behavior, direct
  observations, and assumptions; do not present one category as another.
- Keep simple attributes concise. Add detail when omission, explicit empty
  values, drift, import, sensitive values, or another lifecycle distinction
  changes practitioner behavior.

## Documentation verification

Documentation checks prove different things and should not be collapsed into a
single generation-success signal.

After changing a schema, example, template, or guide input, run:

```sh
go generate ./...
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs \
  validate --provider-name=terraform-provider-contentful
git diff --check
```

The pinned `tfplugindocs` validator checks Registry documentation structure,
front matter, and schema/document correspondence. Its Registry-documentation
glob covers `docs/index.md` and the recognized Registry subdirectories; it does
not treat `docs/design/`, `docs/research/`, or `docs/releasing.md` as Registry
pages. Generation reproducibility remains a separate check: inspect tracked and
untracked output after regeneration.

Also review what the automated checks do not establish:

- Inspect rendered relative links and navigation from the pages you changed.
- Compare practitioner claims with provider behavior and, for Contentful
  behavior, primary documentation or authorized direct evidence.
- Review all configuration and import variants for the affected resource
  together so their resource addresses, variables, and composite IDs agree.
- When an example is intended to be standalone, copy it into an isolated
  Terraform configuration, supply its provider and required variables, run
  `terraform init`, and run `terraform validate`. Do not treat a resource example
  directory containing alternative import forms as one executable module.
- Treat `.tfquery.hcl` examples as Terraform query configurations. Use Terraform
  1.14 or later and the `terraform query` workflow when exercising them; ordinary
  `terraform validate` of `.tf` configuration is not a substitute for checking
  list-resource behavior.

## Code Generation

The commands below implement the
[generation requirement](AGENTS.md#documentation-and-workflow).

Running the root generator requires Terraform on `PATH` because it formats
Terraform examples before regenerating provider documentation. The narrower
management API client generator does not require Terraform.

Full generation command, run from the repository root:

```sh
go generate ./...
```

The full generation command runs the package-local `go:generate` directives:

- `go generate .` formats Terraform examples under `examples/` and regenerates
  provider documentation with `tfplugindocs`.
- `go generate ./internal/contentful-management-go` regenerates the Contentful
  Management API client from
  `internal/contentful-management-go/openapi/openapi.yml` using
  `internal/contentful-management-go/ogen.yml`.

For faster iteration on one generated surface:

```sh
go generate .
go generate ./internal/contentful-management-go
```

Generated files include `internal/contentful-management-go/oas_*_gen.go` and
the managed Registry documentation described in
[Documentation authoring](#documentation-authoring). Edit their sources rather
than the rendered output.

To check reproducibility, rerun the generators in a clean checkout containing
the intended input and generated changes. Use a separate checkout if the working
directory contains unrelated changes. The same cleanliness check as CI prints
short status, including untracked output, before failing when generation causes
drift:

```sh
generated_status="$(git status --short --untracked-files=all)"
if [ -n "$generated_status" ]; then
  printf '%s\n' "$generated_status"
  exit 1
fi
```

## Tests

With `TF_ACC` unset, run the normal unit and local integration test suite:

```sh
go test ./...
```

Run a focused package or test while iterating:

```sh
go test ./internal/provider -run TestContentTypeModelRoundTrip -count=1
```

Run mocked Terraform acceptance tests locally:

```sh
TF_ACC=1 TF_ACC_MOCKED=1 go test ./internal/provider -run '^TestAcc' -count=1
```

The acceptance-test helpers have two modes. Tests using the always-mocked helper
remain mocked regardless of `TF_ACC_MOCKED`. Tests using the mockable helper use
a local HTTP server when `TF_ACC_MOCKED` is set and the configured live
Contentful account when it is absent. Dedicated live-only tests can also skip
when mocked mode is selected. Check the affected test before treating an unset
`TF_ACC_MOCKED` value as evidence that it ran live.

For authorized live Terraform acceptance tests, configure
`CONTENTFUL_MANAGEMENT_ACCESS_TOKEN` in the environment. Clear `TF_ACC_MOCKED`
for the command so an inherited mock setting cannot mask a live-capable check:

```sh
env -u TF_ACC_MOCKED TF_ACC=1 go test ./internal/provider -run '^TestAcc' -count=1
```

Acceptance tests require a Terraform CLI on `PATH`. Mocked acceptance tests use
local HTTP test servers; they do not call Contentful, but they still exercise
the Terraform acceptance-test harness.

## Local provider build

Use Terraform development overrides when a change needs manual CLI verification
against a local provider build. Keep the override scoped to a temporary CLI
configuration rather than changing your normal Terraform installation policy.

Build the provider into a dedicated directory:

```sh
mkdir -p /tmp/terraform-provider-contentful-dev
go build -o /tmp/terraform-provider-contentful-dev/terraform-provider-contentful .
```

Create a temporary Terraform CLI configuration:

```hcl
provider_installation {
  dev_overrides {
    "cysp/contentful" = "/tmp/terraform-provider-contentful-dev"
  }
  direct {}
}
```

Point `TF_CLI_CONFIG_FILE` at that file when running `terraform plan` or another
command in an isolated test configuration. Terraform still performs normal
version selection during `terraform init`; commands after initialization use the
development override for `cysp/contentful`. Do not use a development override as
a deployment or released-provider installation mechanism.

## Linting

Install the [official golangci-lint binary](https://golangci-lint.run/docs/welcome/install/local/#binaries)
at the version recorded in [`.golangci-lint-version`](.golangci-lint-version),
then run the repository checks:

```sh
golangci-lint run
golangci-lint fmt --diff
```

## Releases

See [Provider releases](docs/releasing.md) for the publishing and verification
workflow.
