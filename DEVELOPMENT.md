# Development

This project uses Go-native tooling entrypoints for development workflows.

## Repository guidance

[AGENTS.md](AGENTS.md) is the authoritative repository guidance;
[CLAUDE.md](CLAUDE.md) and [GEMINI.md](GEMINI.md) import it. Keep shared rules
there and command details here. Design contracts belong in `docs/design/`;
link them from the instructions with the conditions that require reading them.

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
| Schema, examples, OpenAPI, or other generation inputs | Follow the [generation requirement](AGENTS.md#documentation-and-workflow) using [Code Generation](#code-generation), plus checks for the affected behavior. |
| A claim about live Contentful behavior | Use primary documentation or an authorized live experiment; mocked tests establish provider behavior against the fixture, not CMA conformance. |

If a check cannot run, record the command and concrete blocker. Distinguish a
passing check from a skipped test or an environment failure. Local verification
does not establish that remote CI passed.

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
the generated schema sections in `docs/`. Edit their sources: the provider
schema, examples, templates, or OpenAPI input.

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
TF_ACC=1 TF_ACC_MOCKED=1 go test ./internal/provider -run '^TestAcc' -count=1 -timeout 15m
```

For authorized live Terraform acceptance tests, configure
`CONTENTFUL_MANAGEMENT_ACCESS_TOKEN` in the environment. Clear `TF_ACC_MOCKED`
for the command so an inherited mock setting cannot mask a live check:

```sh
env -u TF_ACC_MOCKED TF_ACC=1 go test ./internal/provider -run '^TestAcc' -count=1 -timeout 15m
```

All `TestAcc` tests require `TF_ACC`, including registry upgrades and tests that
invoke Terraform directly to inspect logs and terminal output. With `TF_ACC`
unset, the ordinary suite runs unit, provider protocol, local HTTP integration,
mock conformance, property tests, and fuzz seeds without invoking Terraform.
Table-driven acceptance parents can report PASS when all their subtests skip;
inspect the subtest results when checking what actually executed.
Fuzz seeds are regression examples; active fuzzing requires `-fuzz` and a bound,
for example `go test ./internal/provider -run '^$' -fuzz '^FuzzExtensionModelRoundTrip$' -fuzztime 30s`.

Install Terraform on `PATH` or set `TF_ACC_TERRAFORM_PATH` to an existing binary
for reproducible acceptance runs. The framework can otherwise download Terraform;
the direct CLI presentation tests require an installed binary. Registry-upgrade
tests always use local Contentful servers but download the pinned released
provider from the Terraform registry, even with `TF_ACC_MOCKED=1`.

Mocked acceptance tests use isolated local HTTP servers. Mock-only tests always
use those servers; live-capable tests use them when `TF_ACC_MOCKED` is set.
The live-only App Key sibling skips in mocked mode. Live-capable harness calls
serialize access to the shared account and quota; do not remove that serialization
merely to speed up tests. Query tests require Terraform 1.14 and skip on 1.13.
CI runs mocked acceptance on 1.13 and 1.14, and authorized live acceptance on 1.14
when the repository secret is available. See the workflow for exact coverage
flags: ordinary, client, mocked, and live results have separate Codecov flags.

## Linting

Install the [official golangci-lint binary](https://golangci-lint.run/docs/welcome/install/local/#binaries)
at the version recorded in [`.golangci-lint-version`](.golangci-lint-version),
then run the repository checks:

```sh
golangci-lint run
golangci-lint fmt --diff
```

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
