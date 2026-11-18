# Repository guidance

## Execution and scope

- Complete requested changes, verify them, and update the affected documentation. For reviews and plans, deliver analysis unless the user requests edits. Treat examples as evidence of the intended outcome unless the user explicitly limits the scope.
- When reviewing a pull request, commit range, or working-tree diff, follow [Reviewing changes](docs/reviewing.md).
- Resolve routine choices from repository evidence and prior agreements. Proceed on reversible assumptions within scope and flag consequential ones. Ask when an unresolved requirement materially affects scope, authorization, public contracts, or the compatibility or recovery of provider-private state. Continue independent work while waiting.

## Terminology and design

- Use established Contentful, Terraform, API, and protocol terminology. Keep Contentful `sys.version` scoped as `version`. Rename internal symbols when the gain in correctness or clarity justifies the change. Confirm non-obvious new domain terms with the maintainer.
- Compare the current design and alternatives against the same contracts and lifecycle guarantees across affected callers. Distinguish required behavior from implementation choices in code, tests, and documentation. Choose the simplest cohesive design, weighing correctness and maintainability gains against migration, regression, review, and maintenance costs.
- When adding or removing a safeguard, identify the failure and the guarantee it must preserve. Compare alternatives by complexity, operational cost, and failure modes. Explain consequential or uncertain tradeoffs with evidence and identify what is still unknown.
- Extract a helper when it expresses a coherent operation or lifecycle boundary, removes meaningful duplication, or clarifies its caller. A single caller can justify extraction on those grounds. Name helpers for their operations and keep simple one-off expressions local. A production seam is justified when the behavioral evidence it enables warrants its interface and maintenance costs.
- When changing or reviewing Terraform schemas, planning, validation, request conversion, response projection, or state publication, read and follow [Terraform value semantics](docs/design/terraform-value-semantics.md).
- When changing or reviewing HTTP retries, deadlines, or mutation recovery, read and follow the [Contentful HTTP retry policy](docs/design/contentful-http-retry-policy.md).
- For resource-specific behavior or CMA test-server behavior and defaults, use the [design documentation map](docs/design/README.md#design-contracts) to find and follow the applicable contract.

## Evidence and tests

- Ground implementation claims in repository code or direct experiments. Use primary sources or direct experiments for external behavior and compatibility claims.
- When choosing, changing, or reviewing tests, follow [Testing](docs/testing.md). For request and lifecycle changes, prefer independent checks of the exact requests and versions plus end-to-end coverage of affected transitions.
- Select checks using the [validation scope](DEVELOPMENT.md#validation-scope). Broaden or repeat passing checks only for new changes, failures, or unresolved concerns.

## Documentation and workflow

- When writing or reviewing Terraform Registry documentation, follow the [documentation practices](docs/design/provider-documentation.md). Use [Documentation authoring](DEVELOPMENT.md#documentation-authoring) to locate its source.
- Give each contract, invariant, and policy one authoritative home with its evidence and limitations. Keep audit history and completed plans with the change.
- Keep each change focused and preserve unrelated worktree changes. Use a separate worktree and pull request for concurrent work on an independent concern.
- For generated files, edit the authoritative input. When changing a schema or other generation input, follow [Code generation](DEVELOPMENT.md#code-generation).
- Check the final diff against the request and repository guidance. Report each material outcome with evidence or a specific limitation. For design work, explain the benefit of the chosen design, including a decision to retain the existing one.
- When creating, rewriting, or reviewing commit history, follow [Commit conventions](DEVELOPMENT.md#commit-conventions).
