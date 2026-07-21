# Development

This project uses Go-native tooling entrypoints for development workflows.

## Code Generation

Run all generators from a checkout or worktree whose leaf directory is named
`terraform-provider-contentful`; `tfplugindocs` derives the provider name from
that directory name.

Code generation requires Terraform on `PATH` because the root generator formats
Terraform examples before regenerating provider documentation.

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
The diff check exits successfully only when generation leaves the working tree
unchanged:

```sh
git diff --exit-code
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

Always clear the `golangci-lint` cache before linting:

```sh
golangci-lint cache clean
golangci-lint run
```

The cache-clean step keeps lint results independent of stale analyzer state in
this repository.
