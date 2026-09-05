# Development

This project uses Go-native tooling entrypoints for development workflows.

## Code Generation

The root documentation generator passes the provider name explicitly, so all
generators work from an ordinary checkout or worktree name.

Running the root generator requires Terraform on `PATH` because it formats
Terraform examples before regenerating provider documentation. The narrower
management API client generator does not require Terraform.

Run all generators from the repository root:

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

Use the narrower commands when only one generated surface is relevant:

```sh
go generate .
go generate ./internal/contentful-management-go
```

Generated files include `internal/contentful-management-go/oas_*_gen.go` and
the generated schema sections in `docs/`. Change the provider schema, examples,
or OpenAPI input, then regenerate the derived files.

After committing generated changes, rerun the generators from a clean checkout.
Use the same cleanliness check as CI; it prints short status, including
untracked output, before failing when generation causes drift:

```sh
generated_status="$(git status --short --untracked-files=all)"
if [ -n "$generated_status" ]; then
  printf '%s\n' "$generated_status"
  exit 1
fi
```

## Tests

Run the normal unit and local integration test suite:

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

Run live Terraform acceptance tests only when you intend to use a real
Contentful account:

```sh
CONTENTFUL_MANAGEMENT_ACCESS_TOKEN=... TF_ACC=1 go test ./internal/provider -run '^TestAcc' -count=1
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
