# Development

For a provider change, start with the [validation scope](#validation-scope). For test design, see [Testing](docs/testing.md).

## Prerequisites

Tool requirements come from [`go.mod`](go.mod) (Go and generators),
[`.golangci-lint-version`](.golangci-lint-version), and the Terraform matrix in
[the test workflow](.github/workflows/test.yml). Root documentation generation
requires Terraform on `PATH`; acceptance-test options are under [Tests](#tests).

Repository-wide guidance lives in [AGENTS.md](AGENTS.md). Use the [design index](docs/design/README.md) for resource contracts and [Reviewing changes](docs/reviewing.md) for code review.

## Review skills for coding agents

Open this repository in your coding agent. Select a skill by name if your agent supports it, or ask it to follow the linked `SKILL.md` file. For example, `Follow .agents/skills/assess-upgrade-impact/SKILL.md and check the migration instructions on this branch.`

| When to use it | Skill | Example request |
| --- | --- | --- |
| Before opening a PR, on an open PR, or after a fix | [`review-implementation-pr`](.agents/skills/review-implementation-pr/SKILL.md) | “Review my working tree before I open a PR.” Give a branch or PR URL for a review, or ask for a prompt to give an independent reviewer. Include a finding if you want it rechecked. |
| Writing or checking migration advice | [`assess-upgrade-impact`](.agents/skills/assess-upgrade-impact/SKILL.md) | “Can someone with existing state follow this guide? What happens to modules that use the changed attribute?” |
| Preparing a release, reviewing its PR, or checking publication | [`assess-release-readiness`](.agents/skills/assess-release-readiness/SKILL.md) | “Is the provider ready for another release?” Give a release PR URL to review that version. See the [release guide](docs/releasing.md#assess-release-readiness) for more examples. |

An implementation review includes the upgrade assessment when the change could affect existing users. Use the upgrade skill directly when that is your only question.

## Commit conventions

- Use Conventional Commit messages. The type describes the kind of change, the scope identifies the affected component or maintenance concern, and the subject describes the specific change.
- For components under `internal/`, use the immediate directory's exact name as the scope, such as `contentful-management-go` or `provider`. Name individual resources in the subject.
- Tests, examples, generated files, and documentation about a component use that component's scope, regardless of their location.
- For repository maintenance, use the established concern name, such as `deps`, `lint`, `generate`, `release`, or `agents`. Scope workflow changes by their purpose.
- When a cohesive change spans components, use its primary concern as the scope if one clearly owns the change; otherwise omit the scope. Split independently useful changes into separate commits where appropriate.
- For a new concern, prefer an existing directory, tool, or workflow name. Use recent history to resolve choices left open by these rules.

## Repository map

The provider schemas, Terraform lifecycle methods, and request and response models live in [internal/provider](internal/provider/). The [management client](internal/contentful-management-go/) contains the OpenAPI inputs, generated code, and [in-process CMA server](internal/contentful-management-go/testing/).

## Validation scope

Select checks by the changed behavior, using the commands below. When preparing
a code change for merge, inspect the remote checks for that revision;
[`.github/workflows/`](.github/workflows/) defines the current CI matrix.
The matrix does not require replaying every CI job locally. Live acceptance
tests require authorization to use the real Contentful account; available
credentials alone do not establish that authorization.

| Change | Local verification |
| --- | --- |
| Agent instructions or prose outside generation inputs | Review instruction consistency, links, and the final diff; run `git diff --check`. |
| Go behavior | Run tests for the affected packages and the [lint and format checks](#linting). Use the [ordinary test suite](#unit-and-local-integration-tests) and `go build .` for shared behavior or when preparing for merge. |
| Terraform planning, state, or lifecycle | Run focused mocked acceptance tests for the affected transitions in addition to the Go checks; extend coverage where existing tests do not establish the changed behavior. |
| Schema, examples, templates, OpenAPI, or other generation inputs | Follow [Code generation](#code-generation) and [Documentation verification](#documentation-verification), then run the checks for the affected behavior. |
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
| Common resource layout | `templates/resources.md.tmpl`; data-source pages use the generator defaults |
| Resource-specific workflows | `templates/resources/<name>.md.tmpl` (without the `contentful_` prefix) |
| Shared list-resource narrative | `templates/list-resources.md.tmpl` |
| Practitioner guides | `templates/guides/` |
| Provider design contracts and external-behavior evidence | `docs/design/` and `docs/research/` |
| Test conventions and verification patterns | `docs/testing.md` |
| Release process | `docs/releasing.md` |

Keep handwritten contributor pages at the top level of `docs/` or in `docs/design/` and `docs/research/`. During generation, `tfplugindocs` replaces `docs/index.md` and clears every directory it manages, even if the provider has no pages of that kind. [Directories managed by the pinned generator](https://github.com/hashicorp/terraform-plugin-docs/blob/v0.25.0/internal/provider/generate.go#L102-L115)

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

After changing a schema or any other generation input, run full generation from the repository root with Terraform on `PATH`. Review the tracked changes and any new files listed by Git.

```sh
go generate ./...
git status --short --untracked-files=all
```

For a quicker iteration, `go generate .` formats examples and builds the Registry pages. `go generate ./internal/contentful-management-go` rebuilds the client from its OpenAPI definition and `ogen.yml` and can run without Terraform.

Use named OpenAPI schema references for nested request and response values with the same contract. Reusing inline objects through `allOf` generates a separate Go type for each parent. Keep separate schemas when requiredness, nullability, or writable fields differ, and keep conversions that normalize responses.

To check reproducibility, repeat full generation and the status check in a clean checkout containing your intended changes. The status output should be empty. [Generation CI](.github/workflows/gogenerate.yml) enforces a clean result.

## Tests

### Unit and local integration tests

With `TF_ACC` unset, run the normal unit and local integration test suite:

```sh
go test -race -shuffle=on -count=1 -timeout=10m ./...
```

Run a focused package or test while iterating:

```sh
go test ./internal/provider -run '^TestModelTypesPreserveNullAndUnknown$' -count=1 -timeout=5m
```

Use the current test name from the source. A successful command with
`[no tests to run]` has not checked the intended behavior.

### Mocked acceptance tests

Run a focused mocked lifecycle test, then broaden to the mocked acceptance suite
when the change requires it:

```sh
TF_ACC=1 TF_ACC_MOCKED=1 go test ./internal/provider -run '^TestAccRoleResourceCreateUpdateDelete$' -count=1 -timeout=5m
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
as shown under [Property and fuzz campaigns](#property-and-fuzz-campaigns).

### Property and fuzz campaigns

Rapid properties run with the ordinary tests. To increase their iteration budget
locally, use `RAPID_CHECKS` so packages that do not import Rapid still accept the
ordinary suite command:

```sh
RAPID_CHECKS=1000 go test -race -shuffle=on -count=1 -timeout=10m ./...
```

Use an anchored name to fuzz one target without repeating ordinary tests:

```sh
go test ./internal/provider -run '^$' -fuzz '^FuzzExtensionModelRoundTrip$' \
  -fuzztime 30s -fuzzminimizetime 10s -parallel 2 -timeout 3m
```

The [ordinary test workflow](.github/workflows/test.yml) sets the PR property
budget. The [campaign workflow](.github/workflows/fuzz.yml) owns the selected
fuzz targets, time and worker limits, and the larger scheduled/manual budgets.
CI calls it for short campaigns on PRs and main; its daily schedule and manual
dispatch run longer fuzz campaigns and the provider ordinary suite with more
Rapid checks. Campaigns do not enable acceptance tests or receive Contentful
credentials.

Failed jobs retain logs and generated reproductions as artifacts for the
retention period configured in those workflows.
Rapid prints a seed and writes `testdata/rapid/<test>/*.fail` relative to the
tested package. Restore a failure file there and rerun the named test, or use
`-rapid.seed=<seed>` with `go test ./internal/provider -run '^TestName$' -count=1`.
Go fuzzing writes failures beneath `testdata/fuzz/<target>/`; restore the corpus
entry beneath `internal/provider/` and use the `-run` command printed in the log.
Review minimized failures and retain useful cases as regression tests or corpus
entries. A campaign that passes establishes only the inputs it exercised.

### Terraform binary and test isolation

Install Terraform on `PATH` or set `TF_ACC_TERRAFORM_PATH` to an existing binary
for reproducible acceptance runs. The framework can otherwise download Terraform;
the direct CLI presentation tests require an installed binary and do not use
`TF_ACC_TERRAFORM_VERSION`. Set `TF_ACC_TERRAFORM_PATH` to select the same
installed executable for every test in the suite. Registry-upgrade
tests always use local Contentful servers but download the pinned released
provider from the Terraform registry, even with `TF_ACC_MOCKED=1`.

Mocked acceptance tests use isolated local HTTP servers. Mock-only tests always
use those servers; live-capable tests use them when `TF_ACC_MOCKED` is nonempty.
Check the affected test before treating an unset `TF_ACC_MOCKED` as evidence that
it ran live.
The live-only App Key test skips in mocked mode. Live-capable tests serialize access to shared account quotas. Query tests require Terraform 1.14 or later.

### CI coverage

The [test workflow](.github/workflows/test.yml) defines the Terraform matrix and coverage flags. Main and manually dispatched CI require live credentials. PRs without the repository secret skip live tests. The matrix describes CI coverage.

The [CI workflow](.github/workflows/ci.yml) owns the required `ci` check, including short fuzz campaigns. The gate requires every prerequisite job to succeed. A skipped or cancelled prerequisite also fails the gate. Codecov and CodeQL have separate merge requirements. See [release publishing](docs/releasing.md) for the checks required before creating a tag.

### Diagnosing failures

For order-dependent failures, rerun with the `-shuffle=<seed>` printed by Go.
Add `-v` for test and subtest progress. Use `TF_LOG=DEBUG` or
`TF_ACC_PERSIST_WORKING_DIR=1` for focused acceptance debugging; logs and
retained Terraform state may contain credentials. To verify that a changed
assertion executes, temporarily give it an impossible expected value, observe
the intended failure, then restore it.

### Test conventions

See [Testing](docs/testing.md) for naming, layout, fixtures, Go test practices,
and Terraform state, plan, import, and lifecycle checks.

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

Run `golangci-lint fmt` to apply the configured formatting, including gofumpt.

## Vulnerability checks

For a local vulnerability check, run the `go run` command from the
[govulncheck workflow](.github/workflows/govulncheck.yml) at the repository root.
The workflow pins the scanner version; update that pin explicitly when upgrading
the scanner. Each run still consults the current Go vulnerability database.

## Releases

See [Provider releases](docs/releasing.md) for the publishing and verification
workflow.
