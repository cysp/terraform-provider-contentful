# Testing

Tests use Go's `testing` package, Testify assertions, and
`terraform-plugin-testing`. The module versions in `go.mod` are authoritative.
Keep tests beside the code they exercise and use the same Contentful and
Terraform terminology as the implementation.

## Suite layout

| Location | Purpose |
| --- | --- |
| `main_test.go` | CLI behavior |
| `internal/provider/*_test.go` | Provider model, schema, protocol, and lifecycle tests |
| `internal/provider/util/*_test.go` | Provider utility behavior |
| `internal/provider/testdata/` | Terraform configuration, query, upgrade, and fuzz fixtures |
| `internal/contentful-management-go/*_test.go` | Management client and value behavior |
| `internal/contentful-management-go/testing/*_test.go` | Mock API behavior and HTTP conformance |
| `internal/contentful-management-go/integration_tests/` | Generated client against the local mock API over HTTP |

Use `Test<Subject><Behavior>` for ordinary tests. Name comparable acceptance
scenarios `TestAcc<Subject>Resource<Scenario>`,
`TestAcc<Subject>DataSource<Scenario>`, or `TestAcc<Subject>ListResource<Scenario>`.
Keep combined resource contracts named for their shared concern. Spell acronyms
as `ID`, `API`, `HTTP`, `JSON`, and `JWK`, and use `RoundTrip` consistently. Name table cases
after inputs or expected behavior so `go test -run` can select them. Use named
table fields for behavior; do not branch on the case name or derive expected
results from the implementation being tested.

Use the external `provider_test` or client test package when testing exported
behavior. Use the implementation package in `*_internal_test.go` when a test
needs unexported behavior. Do not export production symbols solely for tests.
The `_internal` suffix denotes access to unexported implementation. Ordinary
tests of exported behavior can remain in `*_test.go` or focused `*_unit_test.go`
files in the external test package. Keep focused helpers near their tests.
Group shared setup in a corresponding `*_support_test.go` file; keep substantial
fixtures, fault injection, and HTTP recorders in clearly named `*_fixture_test.go`,
`*_faults_test.go`, and `*_recorder_test.go` files. The provider acceptance harness
owns provider factories and mock/live execution setup.

## Go test structure

- Use `t.Parallel()` for independent tests and subtests. Tests that change
  process environment with `t.Setenv` must remain sequential. Live acceptance
  tests share account quotas and are serialized by the harness.
- Use `t.Context()` for work owned by a test, `t.TempDir()` for temporary files,
  and `t.Cleanup()` for helper-owned resources. Cleanup runs after subtests;
  register HTTP server cleanup as soon as the server starts. Cleanup that must
  make requests after the test ends needs its own bounded context because the
  test context is canceled before cleanup begins.
- Mark helpers accepting `*testing.T` or `testing.TB` with `t.Helper()` when
  they report a failure. Stop on prerequisites with `require.NoError`,
  `require.NotNil`, `require.Len`, or checked type assertions before accessing
  results. Use `assert` to report independent value mismatches together.
- Do not call `require`, `t.Fatal`, or `t.FailNow` in HTTP handlers or worker
  goroutines. Return or record the error and assert it in the test goroutine;
  join workers before inspecting their results.
- Prefer deterministic fixtures, explicit synchronization, and controllable
  clocks to sleeps or repeated expensive random setup. Use real generated
  identities when live account isolation requires uniqueness.
- Keep property and fuzz tests for combinations that examples cannot cover
  economically. Keep seed cases readable and expectations independent of the
  production conversion logic. Ordinary `go test` runs fuzz seeds; sustained
  fuzzing is a separate command.

These conventions follow the [Go testing package](https://pkg.go.dev/testing)
and [Go test review guidance](https://go.dev/wiki/TestComments).

## Terraform acceptance tests

Acceptance tests run Terraform CLI through `resource.Test`. Use the shared
provider harness to configure a local API server or the live Contentful
account. A provider factory creates a fresh provider instance per invocation.
Call `parallelWhenMocked(t)` for tests that support either mode; mock-only tests
can use `t.Parallel()` directly. This keeps the library's execution lifecycle
while preserving the repository's live account serialization requirement.

Choose checks for the behavior being proved:

- Use `ConfigStateChecks`, `statecheck.ExpectKnownValue`, `knownvalue`, and
  `tfjsonpath` for typed state assertions. Distinguish null, empty collections,
  strings, numbers, and booleans. Prefer exact nested values when the whole
  collection is part of the contract. Terraform already checks configured
  values against the plan; additional assertions are most useful for computed
  values, defaults, normalization, and regression boundaries.
- Use `statecheck.CompareValue(compare.ValuesSame())` across steps to prove a
  computed identity survives an update, or `CompareValuePairs` for related
  resources. Create each comparison inside its test so it cannot retain values
  from other tests.
- Use `ConfigPlanChecks` and `plancheck.ExpectResourceAction` for create,
  update, replacement, and no-op behavior. Keep checks at the phase where they
  matter: before apply for planned actions, after refresh for observed drift.
  The library already rejects an unexpected non-empty final plan.
- Verify CLI import using `ImportStateVerify` against prior applied state when
  possible; the library does not support this field for import blocks. For pre-existing remote fixtures, persist imported state and use a
  following configuration step to check its values and plan. An import step
  runs `ImportStateCheck`, not `Check` or `ConfigStateChecks`; placing config
  checks directly on that step leaves them unexecuted.
- Keep `ImportStateVerifyIgnore` narrow and explain why each value cannot be
  read during import, such as a write-only secret or a configured timeout.
  Cover the import mechanism the provider supports, including resource
  identity where applicable.
- Keep legacy `Check` callbacks when they inspect exact HTTP lifecycle
  observations or run during `RefreshState`, which does not execute
  `ConfigStateChecks`. `CheckDestroy` also uses the library's legacy callback
  type. Do not add generic adapters just to hide these supported interfaces.

A create/update scenario should prove the intended state transitions, retained
identity, and final cleanup. Test external deletion and drift where supported,
expected diagnostics for invalid input, and state upgrades with a pinned prior
provider release. `ExpectError` should identify the expected diagnostic, not
match any failure. `ExpectNonEmptyPlan` needs a specific reason; it must not
hide accidental drift. Check remote absence after destroy when deletion or
retention is the behavior under test.

HTTP request and version oracles remain separate evidence from Terraform
state. Assert literal payloads, version headers, and lifecycle ordering. A
mocked acceptance pass proves behavior against the local API model; only a
live test or primary API evidence establishes Contentful compatibility.

See HashiCorp's [state checks](https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests/state-checks/resource),
[plan checks](https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests/plan-checks),
and [test step reference](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-testing/helper/resource#TestStep).

## Terraform fixtures

Use `config.TestNameDirectory()` for a fixture owned by one test and
`config.TestStepDirectory()` when successive steps need separate files. Use
`config.StaticDirectory()` when several scenarios deliberately share a
fixture. Keep variables in `ConfigVariables` with `config.StringVariable`,
`config.BoolVariable`, and other typed constructors. For short HCL helpers,
use numbered format verbs and `%q` for quoted strings.

Rename test-name-based fixture directories with their tests. Format changed
fixtures with `terraform fmt`; query fixtures use the appropriate `.tfquery.hcl`
extension. Keep pinned external-provider declarations with upgrade fixtures.
See HashiCorp's [configuration reference](https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests/configuration).

For commands, prerequisites, CLI selection, account isolation, and CI coverage,
see [Development: Tests](../DEVELOPMENT.md#tests).
