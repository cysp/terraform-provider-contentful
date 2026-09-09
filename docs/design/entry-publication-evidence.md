# Terraform foundations for Entry publication recovery

The [Entry publication contract](terraform-value-semantics.md#entry-publication-ownership-and-partial-field-ownership)
is the authoritative provider policy. This reference records the Terraform
lifecycle and persistence guarantees supporting it. Contentful's independent
[Entry lifecycle reference](../research/entry-lifecycle.md) supplies API evidence;
its observed version increments are not additional provider requirements.

The Framework source below is pinned to v1.19.0, commit
`c7ac25e86333d194946fb5e3fd1114e7d101fc23`, reviewed on 2026-08-30.

## Terraform lifecycle and state

Terraform's upstream resource lifecycle defines configuration, prior state,
planned state, and new state, and requires a known planned value to be returned
unchanged after apply. Values whose result is not predictable must remain unknown
in the plan. It also distinguishes response normalization (preserve the prior
representation) from actual drift (publish the remote value into state). See the
upstream [Resource Instance Change Lifecycle](https://github.com/hashicorp/terraform/blob/8b54d8708cc34f875aa59c3163893f72bdbf498a/docs/resource-instance-change-lifecycle.md#planresourcechange)
and the Framework [plan consistency rules](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#terraform-data-consistency-rules).

The Framework's documented plan sequence is:

1. apply defaults;
2. if the resource plan differs from prior state, mark unconfigured computed
   attributes unknown;
3. run attribute plan modifiers;
4. run resource plan modifiers.

The sequence is documented under
[Plan Modification Process](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification#plan-modification-process).
In pinned v1.19.0,
[`int64planmodifier.UseStateForUnknown`](https://github.com/hashicorp/terraform-plugin-framework/blob/c7ac25e86333d194946fb5e3fd1114e7d101fc23/resource/schema/int64planmodifier/use_state_for_unknown.go#L12-L59)
copies a known prior value, including known null, only when the plan is unknown,
the resource already exists, and configuration is not unknown. Because
resource-level modification runs later, it can explicitly restore unknown for
operations whose publication result is not predictable. If it leaves a known
prior `published_version`, Update must return that exact value or Core will report
an inconsistent result.

Terraform documents `ignore_changes` as considering configured values on Create
but ignoring them on Update so that another process can share management of a
remote object. The Core developer documentation is more precise: for an existing
object, Terraform uses the corresponding prior-state value instead of the
configured value. See the user-facing
[`ignore_changes` contract](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#ignore_changes)
and upstream [planning behavior](https://github.com/hashicorp/terraform/blob/8b54d8708cc34f875aa59c3163893f72bdbf498a/docs/planning-behaviors.md#configuration-driven-behaviors).
Consequently, a provider that later performs a full-body CMA update from the
effective plan will include ignored fields using their refreshed/prior planned
values; publication applies to that complete Entry version.

Framework [private-state documentation](https://developer.hashicorp.com/terraform/plugin/framework/resources/private-state)
says private data is hidden from plans, is readable during plan/Read/Update, and
is writable during plan/Read/Update as well as Create and import. Framework
[diagnostic semantics](https://developer.hashicorp.com/terraform/plugin/framework/diagnostics#how-errors-affect-state)
say Terraform persists returned state even with an error specifically so a
provider can checkpoint successful earlier calls in a multi-call mutation. Taken
together, those contracts support cross-operation pending-version recovery when
a validated draft write succeeds but Publish is not confirmed. An ambiguous
Publish response does not prove that publication remains pending, so recovery
does not infer status or adopt a GET result: normal Read clears the marker if the
exact draft is already published and revokes it on any changed lifecycle tuple;
with refresh disabled it submits only the stale exact marker, and
`VersionMismatch` revokes it immediately.
