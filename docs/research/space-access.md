# Space access: Role and TeamSpaceMembership representations

Role describes permissions and policies within a space. TeamSpaceMembership relates a
team to a space and carries an `admin` Boolean and Role links. Their static
representations do not establish the behavior of permission changes.

## Addressing and representation

The pinned [Role adapter][role-adapter] reads a space Role at
`/spaces/{space_id}/roles/{role_id}`. The [TeamSpaceMembership adapter][membership-adapter]
reads a membership at
`/spaces/{space_id}/team_space_memberships/{team_space_membership_id}`. The parent
paths list the respective space-scoped collections. These detail paths use the
resource's own ID rather than the IDs in its relationship links.

The pinned [Role entity][role] declares `permissions` by application section. Several
sections accept string arrays or a scalar string; `ContentModel` is declared as an
array. Each policy has `effect`, `actions`, and `constraint`; `actions` accepts an action
array or the scalar `all`. These are property-specific declarations, not a single
interchangeable scalar/array rule for every permission.

The pinned [TeamSpaceMembership entity][membership] declares `admin`, `roles`, and
system metadata linking the team and space. Its resource ID identifies the membership;
the team and Role links identify the related resources.

## Direct read observations

Observed: 2026-09-09 (UTC), passive Role and TeamSpaceMembership collection reads, with
one detail comparison for each family.

| Resource | Observed representation |
| --- | --- |
| Role | `name`, `description`, `permissions`, `policies`, and `sys`; permission values included arrays and scalar `all`, while policy actions included arrays and scalars |
| TeamSpaceMembership | `admin`, `roles`, and `sys` |
| List/detail comparisons | Each sampled detail response exactly matched its corresponding list item |

Both collections returned the ordinary `sys`, `total`, `skip`, `limit`, and `items`
envelope. No ordering, deduplication, permission equivalence, admin transition, or
update/delete behavior follows from these static comparisons. Individual user
memberships and effective authorization were not examined.

[role]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/role.ts
[membership]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/entities/team-space-membership.ts
[role-adapter]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/role.ts
[membership-adapter]: https://github.com/contentful/contentful-management.js/blob/883e2b9dc1c76413d5c24e45f74243da699071e4/lib/adapters/REST/endpoints/team-space-membership.ts
