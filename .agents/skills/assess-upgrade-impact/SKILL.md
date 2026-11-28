---
name: assess-upgrade-impact
description: Assess an implementation or documentation diff for changes to what existing Terraform provider users must do or will observe on upgrade.
---

# Assess upgrade impact

Compare the PR head with its base, or use the equivalent local diff. Start with someone who already has configuration, Terraform state, or remote resources. If the change has no plausible effect on their upgrade, say so and finish.

For each migration claim, work out the configuration and steps a reader would take from the guide before consulting the tests. Compare those steps with the scenario the test actually exercises. Check both an unchanged configuration and one updated as directed. Follow changed types or values into module outputs and downstream callers.

Trace each claim to the first operation that could disprove it. An empty plan shows that Terraform proposes no resource action. To establish that a remote value survives a later write, inspect that write and its result. Check distinct configurations separately when they exercise different behavior.

Identify the provider version that created the starting state when it matters. If a fixture seeds that state directly, establish that the released provider could produce the relevant values or name the upgrade path that still lacks evidence. Distinguish what the test asserts, which checks ran, and what a fixture or the live service demonstrated. Use the repository’s [test conventions](../../../docs/testing.md) and [validation scope](../../../DEVELOPMENT.md#validation-scope) to choose a check.

When authoring a change, cover the upgrade in the appropriate test and put the steps users need in the documentation they will read. When reviewing, cite the starting version, transition, and evidence for each material claim. If a claim is unsupported, say who is affected and what check or change to the guidance would resolve it.
