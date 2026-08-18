# Repository guidance

## Terminology and design

- Preserve established Contentful, Terraform, API, protocol, and codebase terminology. Do not rename concepts merely to sound more descriptive. Add a term only when existing vocabulary cannot express a demonstrated need, and confirm non-obvious terminology with the maintainer before applying it.
- Prefer the simplest cohesive design that preserves public contracts and explicit lifecycle boundaries. Require a demonstrated production need for new abstractions, mode flags, shallow wrappers, provider-private markers or status fields, and production-code seams added only for tests.
- When changing Terraform schemas, planning, validation, request conversion, response projection, or state publication, read and follow [Terraform value semantics](docs/design/terraform-value-semantics.md).

## Evidence and tests

- Support technical and compatibility claims with repository code, direct experiments, or primary sources, and state the verification boundary.
- Choose tests for independent behavioral evidence, not assertion count. For request and lifecycle behavior, prefer exact request and version checks plus complete lifecycle coverage; do not derive expected results from the production logic under test.

## Documentation and workflow

- Keep durable documentation current: record user-visible contracts, invariants, evidence, and limitations; omit audit inventories, cleanup chronology, and completed plans.
- Preserve unrelated worktree changes. Isolate unrelated concerns in separate worktrees and pull requests, and keep history reviewable.
- Use Conventional Commit messages with a codebase-area scope.
- After changing a schema or another input to generated code or documentation, run `go generate ./...` and inspect the generated diff.
- Before every `golangci-lint run`, run `golangci-lint cache clean`. Run focused tests for changed behavior plus proportionate broader verification.
