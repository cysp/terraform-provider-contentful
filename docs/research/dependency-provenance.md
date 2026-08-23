# Dependency and toolchain provenance

## Repository policy

The main module's requirements and source-changing module directives are
recorded in [`go.mod`](../../go.mod). Inspect the selected module graph with
`go list -m all` and `go mod graph`. [`go.sum`](../../go.sum) records hashes that
authenticate downloaded module files and can retain entries for versions that
the current graph does not select. When applicable, the configured checksum
database supplies independently authenticated hashes to the Go command.

A `replace` directive, workspace file, vendored tree, temporary fork, alternate
module proxy, or CI-only source substitution is an exception: it must record its
upstream reason, use a deterministic reference, state its removal condition,
and pass generation plus semantic regression tests. None of those mechanisms is
currently present.

Go pseudo-versions are deterministic references to upstream commits, not source
substitutions. The Go command validates their revision/timestamp relationship.
See the Go documentation for
[pseudo-versions](https://go.dev/ref/mod#pseudo-versions) and
[`replace` directives](https://go.dev/ref/mod#go-mod-file-replace).

Generators are ordinary module dependencies declared in
[`tools/tools.go`](../../tools/tools.go). Ogen and Terraform Plugin Docs are
pinned to tagged upstream releases. Generated files are not patched locally;
changes belong in the schema, provider source, or documentation templates and
must reproduce with `go generate ./...`.

GitHub Actions are pinned to full commit SHAs with the corresponding release tag
recorded beside each reference. GitHub identifies a full-length commit SHA as
the only immutable action reference in its
[secure-use guidance](https://docs.github.com/en/actions/reference/security/secure-use#using-third-party-actions).
[Dependabot](../../.github/dependabot.yml) tracks both Go modules and Actions.

## Classifying exceptional-looking selections

Version spelling alone does not establish an exceptional dependency mechanism:

- A pseudo-version may be the only upstream release form, a commit after the
  latest tag, or an older version retained by minimal version selection. Record
  which case applies and the dependency edge that selects it.
- A pre-release is an ordinary upstream release when its tag and checksum are
  authentic. Record why a pre-release remains necessary and prefer a stable
  release when the introducing dependency supports one.
- A `+incompatible` suffix can describe a legitimate pre-modules major-version
  tag whose module path does not contain the major version. It is not evidence
  of a fork.
- A deprecated module can remain in the selected module graph without being
  compiled into the provider. Trace both production and test/generator imports
  before classifying it as graph-only.
- A replacement, fork, patched generated file, local module, or CI-only source
  rewrite changes provenance and is an exception even when its version appears
  ordinary.

Transitive selections should normally be advanced or removed through their
introducing dependency. Do not add a root requirement or downgrade a selected
version merely to make the graph look conventional. A direct override requires
a demonstrated compatibility, correctness, or security need and corresponding
semantic coverage.

For every genuine exception, record the upstream issue or commit, deterministic
reference, affected generated or runtime behavior, validation required after
removal, and an objective removal condition. Remove the record with the
exception; completed audit chronology does not belong in durable documentation.

## Reproducing the classification

Run these commands with the checked-in module files:

```sh
go list -m all
go mod graph
go mod why -m MODULE
go list -m -versions MODULE
go list -m -json MODULE@latest
go mod verify
```

Also verify that the tree contains no nested `go.mod`, `go.work`, vendor tree,
patch file, or dependency-source environment override, and that every Action SHA
resolves to the release tag named beside it. The repository currently has none
of those source-substitution mechanisms. Scheduled indirect dependency updates
use `go get -u` followed by `go mod tidy`; they do not replace module sources.
