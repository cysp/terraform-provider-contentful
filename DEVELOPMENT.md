# Development

Use this guide to change, test, and document the provider. Start with the
[validation scope](#validation-scope) for your change, then use the relevant
commands: [tests](#tests), [generation](#code-generation),
[documentation verification](#documentation-verification), or [linting](#linting).
For release publishing, see [Provider releases](docs/releasing.md).

## Prerequisites

Tool requirements come from [`go.mod`](go.mod) (Go and generators),
[`.golangci-lint-version`](.golangci-lint-version), and the Terraform matrix in
[the test workflow](.github/workflows/test.yml). Root documentation generation
requires Terraform on `PATH`; acceptance-test options are under [Tests](#tests).

## Repository guidance

[AGENTS.md](AGENTS.md) is the authoritative repository guidance;
[CLAUDE.md](CLAUDE.md) and [GEMINI.md](GEMINI.md) import it. Keep shared rules
there and command details here. Design contracts belong in `docs/design/`;
link them from the instructions with the conditions that require reading them.
See the [design documentation map](docs/design/README.md) for the distinction
between provider-wide invariants and resource-specific contracts.

## Repository map

- `internal/provider/contentful_provider.go`, `resource_*.go`, `data_source_*.go`,
  and `list_resource_*.go` define provider configuration, schemas, and Terraform
  lifecycle methods.
- `internal/provider/*_model_request.go` converts Terraform values to CMA request
  payloads; `*_model_response.go` projects CMA responses into Terraform values.
  Shared `request_*.go` and `response_*.go` helpers implement value conversion.
- `internal/provider/*_test.go` exercises those boundaries.
  `contentful_provider_testing_test.go` wires the mocked/live acceptance-test
  harness; `internal/contentful-management-go/testing/` implements the in-process
  CMA server and fixtures.
- `internal/contentful-management-go/` contains the generated Contentful
  Management API client and its OpenAPI generation inputs.
- Documentation inputs and generated outputs are mapped under
  [Documentation authoring](#documentation-authoring).

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
| Schema, examples, templates, OpenAPI, or other generation inputs | Follow the [generation requirement](AGENTS.md#documentation-and-workflow) using [Code generation](#code-generation), then follow [Documentation verification](#documentation-verification), plus checks for the affected behavior. |
| A claim about live Contentful behavior | Use primary documentation or an authorized live experiment; mocked tests establish provider behavior against the fixture, not CMA conformance. |

If a check cannot run, record the command and concrete blocker. Distinguish a
passing check from a skipped test or an environment failure. Local verification
does not establish that remote CI passed.

## Documentation authoring

Practitioner-facing Registry documentation is generated with
`terraform-plugin-docs`. Follow the [documentation practices](docs/design/provider-documentation.md)
for editorial rules and supporting evidence. Change the authoritative input for
the information being documented, then [regenerate](#code-generation) and
[verify the output](#documentation-verification):

| Documentation concern | Authoritative input |
| --- | --- |
| Provider, resource, data-source, list-resource, and attribute contracts | Schema descriptions under `internal/provider/` |
| Terraform configuration and import syntax | `examples/` |
| Provider page | `templates/index.md.tmpl` |
| Resource narrative and workflows | Existing overrides in `templates/resources/`; other resource and data-source pages use the generator defaults |
| Shared list-resource narrative | `templates/list-resources.md.tmpl` |
| Practitioner guides | `templates/guides/` |
| Provider design contracts and external-behavior evidence | `docs/design/` and `docs/research/` |
| Release process | `docs/releasing.md` |

`tfplugindocs` replaces `docs/index.md`, `docs/resources/`, `docs/data-sources/`,
`docs/list-resources/`, and `docs/guides/`. These are generated pages, not just
generated schema fragments. Keep handwritten `docs/design/`, `docs/research/`,
and `docs/releasing.md`; never clear the entire `docs/` tree to regenerate.

## Documentation verification

After regenerating with [Code generation](#code-generation), run the pinned
Registry validator:

```sh
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs \
  validate --provider-name=terraform-provider-contentful
git diff --check
```

Review the pages using the [content and rendering criteria](docs/design/provider-documentation.md#verify-the-claims-and-the-rendered-result),
which also explain the validator's limits. Check
[generation reproducibility](#code-generation) separately, then validate
runnable examples in an isolated configuration:

- Initialize and validate complete workflow examples. Select one import variant
  at a time using the [example conventions](docs/design/provider-documentation.md#make-examples-usable).
- Validate `.tfquery.hcl` configurations using `terraform validate -query`; exercise
  list behavior with `terraform query`. Query tests require Terraform 1.14 or
  later. See HashiCorp's [query workflow](https://developer.hashicorp.com/terraform/language/import/bulk)
  for the general syntax and commands.

## Code generation

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

### Unit and local integration tests

With `TF_ACC` unset, run the normal unit and local integration test suite:

```sh
go test ./...
```

Run a focused package or test while iterating:

```sh
go test ./internal/provider -run '^TestContentTypeFieldValueToTerraformValueRoundTrip$' -count=1
```

Use the current test name from the source. A successful command with
`[no tests to run]` has not checked the intended behavior.

### Mocked acceptance tests

Run a focused mocked lifecycle test, then broaden to the mocked acceptance suite
when the change requires it:

```sh
TF_ACC=1 TF_ACC_MOCKED=1 go test ./internal/provider -run '^TestAccRoleResourceCreateUpdateDelete$' -count=1
TF_ACC=1 TF_ACC_MOCKED=1 go test ./internal/provider -run '^TestAcc' -count=1 -timeout 15m
```

### Live acceptance tests

For authorized live Terraform acceptance tests, configure
`CONTENTFUL_MANAGEMENT_ACCESS_TOKEN` in the environment. Clear `TF_ACC_MOCKED`
for the command so an inherited mock setting cannot mask a live-capable check:

```sh
env -u TF_ACC_MOCKED TF_ACC=1 go test ./internal/provider -run '^TestAcc' -count=1 -timeout 15m
```

### Interpreting test results

All `TestAcc` tests require `TF_ACC`, including registry upgrades and tests that
invoke Terraform directly to inspect logs and terminal output. With `TF_ACC`
unset, the ordinary suite runs unit, provider protocol, local HTTP integration,
mock conformance, property tests, and fuzz seeds without invoking Terraform.
Table-driven acceptance parents can report PASS when all their subtests skip;
inspect the subtest results when checking what actually executed.
Fuzz seeds are regression examples; active fuzzing requires `-fuzz` and a bound,
for example `go test ./internal/provider -run '^$' -fuzz '^FuzzExtensionModelRoundTrip$' -fuzztime 30s`.

### Terraform binary and test isolation

Install Terraform on `PATH` or set `TF_ACC_TERRAFORM_PATH` to an existing binary
for reproducible acceptance runs. The framework can otherwise download Terraform;
the direct CLI presentation tests require an installed binary. Registry-upgrade
tests always use local Contentful servers but download the pinned released
provider from the Terraform registry, even with `TF_ACC_MOCKED=1`.

Mocked acceptance tests use isolated local HTTP servers. Mock-only tests always
use those servers; live-capable tests use them when `TF_ACC_MOCKED` is nonempty.
Check the affected test before treating an unset `TF_ACC_MOCKED` as evidence that
it ran live.
The live-only App Key test skips in mocked mode. Live-capable harness calls
serialize access to the shared account and quota; do not remove that serialization
merely to speed up tests. Query tests require Terraform 1.14 and skip on 1.13.

### CI coverage

The [test workflow](.github/workflows/test.yml) defines the Terraform version
matrix. CI runs the ordinary suite once, mocked acceptance tests on the two
newest stable Terraform minors, and authorized live acceptance on the newest
stable minor when the repository secret is available. The explicit minor ranges
select the latest patch in each minor; update them together through review when
a new stable minor is released, and coordinate the corresponding GitHub required
status check names before merging. This matrix defines CI coverage, not a minimum
supported Terraform version. See the workflow for the separate ordinary, client,
mocked, and live coverage flags.

### Test conventions

Name comparable acceptance scenarios `TestAcc<Subject>Resource<Scenario>`,
`TestAcc<Subject>DataSource<Scenario>`, or `TestAcc<Subject>ListResource<Scenario>`.
Keep combined resource contracts named for their shared concern. Use `Test` for
unit, protocol, and local HTTP tests, and `Fuzz` for fuzz targets; spell acronyms
as `ID`, `API`, `HTTP`, `JSON`, and `JWK`, and use `RoundTrip` consistently.
Rename `TestNameDirectory` fixtures and fuzz corpus directories with their test.

Use independent cases in tables and sequential lifecycle transitions in explicit
steps. Prefer a scenario directory and independent per-step `ConfigVariables`
for simple value changes. Keep structural changes, unknown-producing expressions,
literal lifecycle settings, and substantial nested HCL visible in separate
fixtures or concise inline configuration. Do not encode phases or a fixture
language merely to reduce directory count.

Prefer typed state and plan checks when null, empty, unknown, collection semantics,
or action timing matter. Retain API checks and phase-specific legacy hooks.
CLI imports use `ImportStateCheck` for direct imported-state assertions;
`Check` and `ConfigStateChecks` are not invoked by that import path. Use
`ImportStateVerify` when a preceding apply supplies the comparison state.
Explain verification exclusions and test those attributes separately.

## Local provider build

`go build .` produces `terraform-provider-contentful` in the repository root.
For manual Terraform use, the provider address is `cysp/contentful`; see
HashiCorp's [development overrides](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers).

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
