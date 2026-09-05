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
to the changed behavior before merge.

## Repository guidance

[AGENTS.md](AGENTS.md) is the authoritative repository guidance;
[CLAUDE.md](CLAUDE.md) and [GEMINI.md](GEMINI.md) import it. Keep shared rules
there and command details here. Design contracts belong in `docs/design/`;
link them from the instructions with the conditions that require reading them.

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
| Schema, examples, templates, OpenAPI, or other generation inputs | Follow the [generation requirement](AGENTS.md#documentation-and-workflow) using [Code Generation](#code-generation), plus checks for the affected behavior. |
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

When adding practitioner prose, keep attribute-level contracts concise in the
schema, put multi-step workflows in templates or guides, and keep implementation
algorithms and evidence details in developer documentation. Preserve
resource-specific semantics rather than applying a generic omission, retry, or
secret-handling rule where behavior differs.

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
