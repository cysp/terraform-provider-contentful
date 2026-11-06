# Changelog

Release Please maintains entries for releases prepared through a release PR.
Entries through 0.0.70 preserve the existing
[GitHub release notes](https://github.com/cysp/terraform-provider-contentful/releases);
dates are UTC publication dates.

## [0.0.71](https://github.com/cysp/terraform-provider-contentful/compare/v0.0.70...v0.0.71) (2026-11-06)

**Compatibility:** The provider now rejects HTTP redirects for mutation requests.
If you use a custom API endpoint or reverse proxy, configure it to handle writes
directly, without redirecting them. Set the provider's `url` to the final API
endpoint where appropriate.


### Features

* **release:** prepare and publish releases with cyspbot-releaser ([06a3460](https://github.com/cysp/terraform-provider-contentful/commit/06a3460ef577a8c29d55a7e6d1c5ca20c1cf6ecc))
* **release:** prepare release PRs on demand ([2af3fe6](https://github.com/cysp/terraform-provider-contentful/commit/2af3fe684ffcfe3235c1176ea10fe464722686b1))


### Bug fixes

* **contentful-management-go:** detach mock response snapshots ([ad5e562](https://github.com/cysp/terraform-provider-contentful/commit/ad5e562eb313a774b11e30e34cd7265411847d78))
* **contentful-management-go:** enforce Extension mock versions ([510c803](https://github.com/cysp/terraform-provider-contentful/commit/510c803175de60e9751b4a13e0e959eb7bf19f1a))
* **contentful-management-go:** export transport client result type ([9040651](https://github.com/cysp/terraform-provider-contentful/commit/9040651393f6392018d5b7141d9afaa30da9e6bd))
* **contentful-management-go:** reject updates to missing mock roles ([3e39942](https://github.com/cysp/terraform-provider-contentful/commit/3e39942887563ef9308e13517afa9a278ad17abe))
* **provider:** enforce planned content type taxonomy in responses ([025cf75](https://github.com/cysp/terraform-provider-contentful/commit/025cf752242143c437e0b8b5271e195769ca359c))
* **provider:** keep personal access token create state known ([2d59b03](https://github.com/cysp/terraform-provider-contentful/commit/2d59b03f634382079abe1b3fb0222fa6021d1211))
* **provider:** reject contradictory managed read identities ([51eb73b](https://github.com/cysp/terraform-provider-contentful/commit/51eb73b32f64baf2c89a9a0e03c22ad7df1920c8))
* **provider:** reject redirects for every mutation ([56d4232](https://github.com/cysp/terraform-provider-contentful/commit/56d42324b4f24ba74b3822e6d1b9af32d9e5b161))
* **provider:** revoke stale content type activation authority ([7ca7032](https://github.com/cysp/terraform-provider-contentful/commit/7ca70327450bcbfb32d293e157557a0cc0bde38f))
* **provider:** validate environment readiness endpoint identity ([3b977d5](https://github.com/cysp/terraform-provider-contentful/commit/3b977d56057705a5f7da10fb3b9e648c7ec4ddbd))


### Documentation

* **agents:** clarify outcome and test evidence requirements ([96bd8e9](https://github.com/cysp/terraform-provider-contentful/commit/96bd8e968550f2614da8012b898614e7d25d97fc))
* **agents:** evaluate designs against required outcomes ([75f2121](https://github.com/cysp/terraform-provider-contentful/commit/75f21216d9d93897a4701687ce12d6e4322f998e))
* **agents:** refine design exceptions and route commit conventions ([384b387](https://github.com/cysp/terraform-provider-contentful/commit/384b3871bba4b3975512afc65c4dddcb6973eb81))
* **agents:** route work to authoritative documentation ([7f79951](https://github.com/cysp/terraform-provider-contentful/commit/7f79951b13d98015147449a6c22bd427ce3e0f30))
* **provider:** align signing secret redirect guidance ([17a8b5d](https://github.com/cysp/terraform-provider-contentful/commit/17a8b5dde8d61bb63fda31ab6edd9345491b2353))
* **provider:** explain direct mutation endpoint requirement ([17fa3ee](https://github.com/cysp/terraform-provider-contentful/commit/17fa3ee7f341244536e2f5d9cf7c9e37dfefcbe8))


### Continuous integration

* **deps:** retire redundant Go dependency submission workflow ([57a5b0e](https://github.com/cysp/terraform-provider-contentful/commit/57a5b0e6b1f01cc1f0b020ff2c07284f73dafca1))
* **release:** orchestrate checks through reusable workflows ([34086e8](https://github.com/cysp/terraform-provider-contentful/commit/34086e8d722befde6522f99ee88fdc8da108e536))


### Tests

* **provider:** construct unknown taxonomy responses structurally ([0da91a0](https://github.com/cysp/terraform-provider-contentful/commit/0da91a0a65572b3984674b925416960cab9820c3))
* **provider:** pass owning contexts through test helpers ([84961b1](https://github.com/cysp/terraform-provider-contentful/commit/84961b1f989fb519748e0bfe6c399991667336bd))
* **provider:** reproduce content type taxonomy activation mismatch ([6ae56e1](https://github.com/cysp/terraform-provider-contentful/commit/6ae56e1439ed1990d8a1b88bb5fb21592a47e4f9))


### Refactoring

* **provider:** share app definition discovery parameter schema ([642b0eb](https://github.com/cysp/terraform-provider-contentful/commit/642b0ebd9fad4e7a34fe31df17875002fb3d8ddf))
* **release:** select explicit Release Please operations ([88c5860](https://github.com/cysp/terraform-provider-contentful/commit/88c58606bbf9943639c5e275b9e96f97cd70e7bf))


### Maintenance

* **deps:** bump indirect dependencies ([affc6e7](https://github.com/cysp/terraform-provider-contentful/commit/affc6e714c745ea995f023c0bd917bb02596a999))
* **deps:** bump indirect dependencies ([a2b8df5](https://github.com/cysp/terraform-provider-contentful/commit/a2b8df5bbf3ffef7746b692a23cccb59f59d599c))

## [0.0.70](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.70) (2026-10-21)

### What's Changed

* test(provider): encode Live Preview Variables response fixtures by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/772
* test(provider): use model responses and literal App Action request expectations by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/771
* docs(agents): clarify test and documentation review criteria by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/773
* test: encode JSON fixtures from Go objects by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/775
* test(provider): reject ignored import assertions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/774
* fix(provider): retry rate-limited mutations consistently by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/776
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/756
* chore: update Go and streamline development tooling by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/781
* chore(deps): bump goreleaser/goreleaser-action from 7.2.3 to 7.3.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/779
* fix(deps): select indirect requirements from go.mod by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/782
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/777
* chore(deps): bump github.com/hashicorp/terraform-json from 0.28.0 to 0.29.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/778
* fix(provider): reconcile App Installation parameters and avoid timeout-only writes by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/780
* docs(agents): define a reusable change review process by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/783


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.69...v0.0.70

## [0.0.69](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.69) (2026-09-29)

### What's Changed

* chore(deps): bump codecov/codecov-action from 7.1.0 to 7.1.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/759
* chore(release): bump GoReleaser to 2.18.2 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/762
* fix(provider)!: expose App Definition data source items as an object by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/763
* fix(provider): validate App Definition locations before writes by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/764
* feat(provider): allow specified App Action IDs on create by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/767
* feat(provider): add Role discovery data sources by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/765
* feat(provider): add Content Type discovery data sources by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/766
* docs(provider): correct Content Type discovery guidance by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/769


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.68...v0.0.69

## [0.0.68](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.68) (2026-09-22)

### What's Changed

* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/750
* feat(provider): manage locales and expose locale list resource by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/512
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/751
* chore(deps): bump codecov/codecov-action from 7.0.0 to 7.1.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/752
* test(provider): consolidate lifecycle and mutation recovery tests by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/757
* feat(provider): manage App Actions and expose discovery data sources by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/755
* test(provider): verify applied imports and Locale selection guards by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/760
* docs(provider): correct signing secret coverage reference by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/761


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.67...v0.0.68

## [0.0.67](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.67) (2026-09-14)

### What's Changed

* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/748
* feat(provider): add space webhook signing secrets by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/745
* feat(provider): add configuration discovery data sources by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/747
* feat(provider): add app event subscription resource by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/746
* chore(lint): check gofumpt separately in CI by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/749


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.66...v0.0.67

## [0.0.66](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.66) (2026-09-13)

### What's Changed

* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/738
* docs(research): add App Framework resource and API reference by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/739
* docs(research): organize durable Contentful API references by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/740
* docs(research): incorporate observed Contentful response contracts by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/742
* test: strengthen and standardize provider test coverage by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/744
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/741
* docs(provider): align documentation with behavior and reader tasks by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/743


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.65...v0.0.66

## [0.0.65](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.65) (2026-09-08)

### What's Changed

* ci(test): update Terraform minor version coverage by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/732
* feat(live-preview-variables): manage environment variables documents by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/734
* fix(provider): harden lifecycle updates, logging, and state upgrades by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/735
* docs: improve provider and contributor documentation by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/736
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/733


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.64...v0.0.65

## [0.0.64](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.64) (2026-09-06)

### What's Changed

* fix(provider): report cancelled environment readiness waits by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/712
* docs: explain environment readiness and timeouts by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/714
* ci(generate): make generation reproducible by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/713
* fix(personal-access-token): preserve token on timeout update by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/715
* docs(agents): align guidance with Sol and Astra by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/716
* refactor(contentful-management-go): remove unused nullable integer conversions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/717
* refactor(contentful-management): collapse transport client wrappers by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/719
* refactor(provider): share provider data assignment policy by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/718
* test(provider): cover opaque JSON normalization contracts by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/720
* fix(editor-interface): account for first activation during recovery by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/728
* test(provider): standardise suite execution and conventions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/729
* test(provider): verify import and lifecycle contracts by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/730
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/731
* docs: clarify provider behavior and repository workflows by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/721
* feat(preview-environment): manage content preview platforms by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/576


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.63...v0.0.64

## [0.0.63](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.63) (2026-09-04)

### What's Changed

* fix(provider): fail closed during configuration by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/588
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/611
* chore(deps): bump github.com/hashicorp/terraform-plugin-log from 0.10.0 to 0.11.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/612
* fix(app-installation): reject unresolved request values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/599
* fix(list): reject unresolved configuration values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/617
* fix(entry): reject unresolved request values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/614
* fix(role): reject unresolved request values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/615
* fix(app): reject unresolved request values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/616
* fix(content-type): reject unresolved field request values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/618
* fix(content-type): reject unresolved request unions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/605
* refactor(provider): share request scalar conversion by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/606
* fix(team): reject unresolved request values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/607
* fix(webhook): reject unresolved request values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/608
* fix(editor-interface): reject unresolved request values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/609
* fix(webhook): reject unresolved filter values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/610
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/613
* chore(deps): bump actions/attest-build-provenance from 4.1.1 to 4.2.2 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/619
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/620
* chore(deps): bump chikachow/cyspbot-app-token-action from 0.0.10 to 0.0.12 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/622
* chore(deps): bump github.com/ogen-go/ogen from 1.23.0 to 1.24.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/621
* fix(provider): handle unknown values safely by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/579
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/626
* chore(deps): bump github.com/stretchr/testify from 1.11.1 to 1.12.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/627
* fix(webhook): validate filter operators during planning by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/624
* fix(provider): preserve opaque JSON number precision by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/623
* fix(provider): omit secret-bearing log payloads by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/495
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/630
* docs(agents): add repository-wide coding guidance by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/629
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/632
* chore(deps): bump github.com/stretchr/testify from 1.12.0 to 1.12.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/633
* fix(provider): reject unresolved mutation request values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/631
* fix(taxonomy): preserve lifecycle ownership and remote drift by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/628
* fix(provider): harden content type activation and version locking by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/625
* fix(contentful-management-go): make test stores zero-value safe by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/634
* test(taxonomy): exclude rate-limited recorder attempts by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/642
* docs(provider): synchronize surface documentation by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/638
* fix(provider/app-signing-secret): align CMA semantics by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/635
* fix(contentful-management-go): align test server conformance by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/639
* fix(environment-status): stop polling failed environments by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/643
* fix(release): preserve clean artifact provenance by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/644
* fix(provider): avoid replaying ambiguous mutations by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/646
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/645
* fix(entry): harden publication lifecycle by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/640
* build(go): require Go 1.26.6 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/655
* feat(cli): report provider build identity by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/662
* docs(provider): qualify Contentful 429 retry evidence by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/657
* docs(provider/app-signing-secret): disclose state persistence by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/658
* fix(space-enablement): align CMA fake behavior by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/656
* fix(entry): reject specified ID creation collisions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/659
* fix(content-type): reject specified ID creation collisions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/663
* docs(provider): document ambiguous generated-ID creation by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/661
* fix(content-type): activate only provider-authored drafts by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/664
* fix(entry): publish only newly written drafts by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/665
* fix(provider): align private version handling by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/666
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/667
* fix(entry): align destroy with CMA lifecycle by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/670
* ci(release): cordon signing secrets by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/660
* refactor(provider): resolve deferred cleanup findings by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/672
* fix(release): fail closed before publication by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/671
* fix(provider): recover exact publication versions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/673
* ci(test): remove Contentful credentials from non-live tests by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/641
* fix(role): validate "all" permission values and policy actions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/675
* refactor(provider): use Contentful version terminology by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/676
* refactor(provider): simplify normalized JSON naming by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/678
* refactor(extension): name configuration directly by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/679
* refactor(provider): align API key terminology by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/677
* fix(webhook): report unsupported response filter properties by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/680
* fix(provider): reconcile mutation responses with plans by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/674
* fix(extension): migrate legacy source state by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/683
* fix(provider): protect secret-capable values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/684
* docs(readme): remove duplicate provider inventories by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/689
* refactor(provider): simplify resource identity handling by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/685
* refactor(provider): centralize resource timeout policy by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/690
* refactor(provider): simplify CMA response classification by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/686
* refactor(provider): remove unused scaffolding by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/687
* fix(provider): correct typed value equality by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/688
* docs(development): document development workflow by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/532
* fix(provider): improve configuration diagnostics by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/534
* fix(team): configure default description by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/552
* test(provider): guard private version representatives by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/691
* fix(entry): prevent updates from recreating deleted entries by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/692
* fix(webhook): report unsupported term response properties by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/693
* feat(provider): log CMA response retry decisions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/694
* chore(lint): update golangci-lint by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/696
* docs(entry): repair destroy evidence links by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/697
* fix(webhook): preserve basic password state by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/695
* fix(entry): reject duplicate metadata links by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/698
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/682
* fix(webhook): require at least one topic by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/699
* fix(provider): reconcile complete mutation responses by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/700
* refactor(provider): consolidate request string policies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/702
* refactor(provider): simplify infallible response projection by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/703
* refactor(provider): standardize identifier plan modifiers by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/704
* refactor(provider): simplify terminal lifecycle control flow by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/705
* fix(provider): clarify import diagnostics by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/708
* fix(provider): correct app definition text by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/707
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/706
* fix(provider): log delete operations by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/709
* fix(provider): log list operations by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/710
* fix(release): restore published release trigger by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/711


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.62...v0.0.63

## [0.0.62](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.62) (2026-08-02)

### What's Changed

* chore(deps): bump chikachow/cyspbot-app-token-action from 0.0.9 to 0.0.10 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/583
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/584
* chore(deps): bump actions/checkout from 7.0.0 to 7.0.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/585
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/586
* chore(deps): bump golangci/golangci-lint-action from 9.2.1 to 9.3.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/562
* refactor(schema): use framework state plan modifiers by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/591
* fix(schema): validate configurable collection children by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/595
* fix(entry): handle omitted response fields safely by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/592
* fix(provider): make multipart imports transactional by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/590
* fix(taxonomy): stop after conversion errors by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/593
* chore(deps): bump github.com/go-faster/errors from 0.7.1 to 0.8.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/596
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/587
* fix(provider): reject invalid request list elements by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/597
* refactor(webhook): consolidate response filter decoding by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/594
* fix(environment-status): publish converted state atomically by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/589
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/598


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.61...v0.0.62

## [0.0.61](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.61) (2026-07-20)

### What's Changed

* fix(ci): skip live acceptance without secrets by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/582
* feat(teams): add organization teams data source by @christophersjchow in https://github.com/cysp/terraform-provider-contentful/pull/580
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/578

### New Contributors

* @christophersjchow made their first contribution in https://github.com/cysp/terraform-provider-contentful/pull/580

**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.60...v0.0.61

## [0.0.60](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.60) (2026-07-19)

### What's Changed

* chore(deps): bump actions/setup-go from 6.5.0 to 7.0.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/577
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/574
* feat(app-key): manage caller-supplied Contentful app keys by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/565


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.59...v0.0.60

## [0.0.59](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.59) (2026-07-12)

### What's Changed

* test(provider): improve test coverage by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/559
* chore(deps): bump actions/attest-build-provenance from 4.1.0 to 4.1.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/560
* chore(deps): bump goreleaser/goreleaser-action from 7.2.2 to 7.2.3 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/561
* chore(deps): bump chikachow/cyspbot-app-token-action from 0.0.5 to 0.0.6 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/564
* chore(deps): bump chikachow/cyspbot-app-token-action from 0.0.6 to 0.0.8 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/566
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/563
* fix(contentful-management-go/testing): match content type lifecycle semantics by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/569
* chore(deps): bump chikachow/cyspbot-app-token-action from 0.0.8 to 0.0.9 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/570
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/568
* chore(deps): bump golang.org/x/sync from 0.21.0 to 0.22.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/567
* chore(deps): bump github.com/ogen-go/ogen from 1.22.0 to 1.23.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/571
* fix(provider): reconcile content type metadata plans by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/572
* feat(taxonomy): manage concepts and concept schemes by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/573


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.58...v0.0.59

## [0.0.58](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.58) (2026-06-28)

### What's Changed

* chore(deps): bump github.com/ogen-go/ogen from 1.20.3 to 1.22.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/547
* chore(deps): bump chikachow/cyspbot-app-token-action from 0.0.4 to 0.0.5 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/555
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/553
* chore(deps): bump actions/setup-go from 6.4.0 to 6.5.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/554
* refactor(contentful-management): normalize response content types by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/558
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/556


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.57...v0.0.58

## [0.0.57](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.57) (2026-06-20)

### What's Changed

* chore(deps): bump codecov/codecov-action from 6.0.1 to 7.0.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/546
* chore(deps): bump golang.org/x/sync from 0.20.0 to 0.21.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/545
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/544
* feat(contentful-management-go): accept application/json responses when getting teams by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/551
* chore(deps): bump actions/checkout from 6.0.3 to 7.0.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/549
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/548


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.56...v0.0.57

## [0.0.56](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.56) (2026-06-08)

### What's Changed

* chore(deps): bump hashicorp/setup-terraform from 4.0.0 to 4.0.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/513
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/511
* test(provider/entry): cover localised entry fields by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/514
* chore(deps): bump actions/create-github-app-token from 3.1.1 to 3.2.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/516
* chore(deps): bump step-security/harden-runner from 2.19.1 to 2.19.3 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/519
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/518
* chore(deps): bump codecov/codecov-action from 6.0.0 to 6.0.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/520
* chore(.github/workflows): use cyspbot-action for dependency updates by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/524
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/521
* chore(deps): bump github.com/go-git/go-git/v5 from 5.19.0 to 5.19.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/522
* chore(deps): bump goreleaser/goreleaser-action from 7.2.1 to 7.2.2 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/523
* chore(.github/workflows): remove step-security harden-runner by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/525
* chore(deps): bump golangci/golangci-lint-action from 9.2.0 to 9.2.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/527
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/526
* chore(github/workflows): update cyspbot action owner by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/530
* chore(github/workflows): update cyspbot app token action by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/531
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/529
* chore(deps): bump chikachow/cyspbot-app-token-action from 0.0.3 to 0.0.4 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/538
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/537
* ci(golangci): update lint binary by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/539
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/540
* chore(deps): bump actions/checkout from 6.0.2 to 6.0.3 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/543
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/541


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.55...v0.0.56

## [0.0.55](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.55) (2026-05-10)

### What's Changed

* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/502
* chore(deps): bump step-security/harden-runner from 2.19.0 to 2.19.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/503
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/504
* chore(lint): update golangci-lint by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/505
* fix(provider): paginate list resource results by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/499
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/506
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/507
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/509
* test(contentful-management-go/testing): use localised entry field values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/510
* fix(provider): include null timeouts in list resources responses by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/508


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.54...v0.0.55

## [0.0.54](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.54) (2026-05-01)

### What's Changed

* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/492
* chore(deps): bump github.com/hashicorp/terraform-plugin-testing from 1.15.0 to 1.16.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/491
* chore(deps): bump goreleaser/goreleaser-action from 7.1.0 to 7.2.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/493
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/494
* fix(provider/preview_api_key): report failure to retrieve preview api keys by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/496
* chore(deps): bump pgregory.net/rapid from 1.2.0 to 1.3.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/500
* feat(provider): apply a small minimum duration to resource timeouts read from state by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/498
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/501


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.53...v0.0.54

## [0.0.53](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.53) (2026-04-24)

### What's Changed

* chore(deps): bump go to 1.26 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/434
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/438
* chore(deps): bump github.com/ogen-go/ogen from 1.17.0 to 1.19.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/437
* chore(deps): bump goreleaser/goreleaser-action from 6.4.0 to 7.0.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/441
* chore(deps): bump hashicorp/setup-terraform from 3.1.2 to 4.0.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/442
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/440
* chore(deps): bump github.com/hashicorp/terraform-plugin-go from 0.29.0 to 0.30.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/444
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/445
* chore(deps): bump github.com/ogen-go/ogen from 1.19.0 to 1.20.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/443
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/448
* chore(deps): bump actions/attest-build-provenance from 3.2.0 to 4.1.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/449
* chore(deps): bump actions/setup-go from 6.2.0 to 6.3.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/446
* chore(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.17.0 to 1.18.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/450
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/453
* chore(deps): bump github.com/ogen-go/ogen from 1.20.0 to 1.20.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/452
* chore(deps): bump crazy-max/ghaction-import-gpg from 6.3.0 to 7.0.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/451
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/454
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/455
* chore(deps): bump golang.org/x/sync from 0.19.0 to 0.20.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/457
* chore(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.18.0 to 1.19.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/458
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/456
* chore(deps): bump actions/create-github-app-token from 2.2.1 to 3.0.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/462
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/461
* chore(deps): bump codecov/codecov-action from 5.5.2 to 5.5.3 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/464
* [StepSecurity] Apply security best practices by @stepsecurity-app[bot] in https://github.com/cysp/terraform-provider-contentful/pull/468
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/465
* chore(deps): bump codecov/codecov-action from 5.5.3 to 6.0.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/466
* chore(deps): bump github.com/ogen-go/ogen from 1.20.1 to 1.20.2 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/467
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/469
* chore(deps): bump actions/setup-go from 6.3.0 to 6.4.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/470
* chore(deps): bump step-security/harden-runner from 2.16.0 to 2.16.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/472
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/473
* chore(deps): bump github.com/ogen-go/ogen from 1.20.2 to 1.20.3 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/474
* chore(deps): bump step-security/harden-runner from 2.16.1 to 2.18.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/479
* chore(deps): bump actions/create-github-app-token from 3.0.0 to 3.1.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/476
* chore(deps): bump peter-evans/create-pull-request from 8.1.0 to 8.1.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/477
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/478
* chore(github/workflows): allow queuing of testacc jobs by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/484
* chore(deps): bump step-security/harden-runner from 2.18.0 to 2.19.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/481
* chore(deps): bump goreleaser/goreleaser-action from 7.0.0 to 7.1.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/483
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/480
* chore(deps): bump github.com/hashicorp/terraform-plugin-docs from 0.24.0 to 0.25.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/482
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/485
* feat(contentful-management-go/testing): implement rate limiting by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/486
* feat(contentful-management-go/testing): default to a 5/s rate limit in tests by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/487
* feat(provider): respect contentful rate limit headers by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/488
* feat(provider): implement timeouts by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/490
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/489
* test(provider): parallelise acceptance tests when mocked by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/439

### New Contributors

* @stepsecurity-app[bot] made their first contribution in https://github.com/cysp/terraform-provider-contentful/pull/468

**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.52...v0.0.53

## [0.0.52](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.52) (2026-02-16)

### What's Changed

* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/424
* docs: create readme by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/425
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/426
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/428
* fix(provider/app_signing_secret): use correct log identifiers by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/429
* feat(provider/environment): expose environment status by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/431
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/433
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/435
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/436
* feat(provider/environment): introduce a data source which can be used to wait for an environment to be ready by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/432


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.51...v0.0.52

## [0.0.51](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.51) (2026-02-01)

### What's Changed

* test(contentful-management-go): implement access token validation, refactor integration test to use local test server by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/409
* chore(deps): use go 1.25 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/410
* chore(github/workflows): bump golangci-lint to v2.7.2 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/411
* test(provider): organise entry test data by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/412
* fix(provider/content-type): preserve null metadata annotations, test round-tripping through request by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/413
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/408
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/414
* chore(deps): bump actions/setup-go from 6.1.0 to 6.2.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/416
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/415
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/417
* chore(deps): bump actions/checkout from 6.0.1 to 6.0.2 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/419
* chore(deps): bump peter-evans/create-pull-request from 8.0.0 to 8.1.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/418
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/420
* ref(contentful-management-go): remove unused created-at and updated-at timestamp fields by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/423
* chore(deps): bump actions/attest-build-provenance from 3.1.0 to 3.2.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/421


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.50...v0.0.51

## [0.0.50](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.50) (2026-01-05)

### What's Changed

* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/393
* ref(provider): simplify identity model creation for multipart ids by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/396
* chore(deps): bump actions/attest-build-provenance from 3.0.0 to 3.1.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/394
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/395
* ref(provider): simplify multipart identity model construction by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/398
* ref(provider/typed): allow the zero value for null and unknown lists and maps by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/400
* fix(provider): correctly roundtrip entry model metadata by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/397
* feat(contentful-management-go/testing): remove special-cased nonexistent identifiers by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/403
* feat(provider/tag): introduce tag resource by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/402
* test(provider/entry): refactor rapid generator, model conversion by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/404
* chore(golangci-lint): disable noglobals in test files by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/406
* chore(github/workflows): run integration tests on latest terraform, limit acceptance tests steps on only acceptance tests by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/407
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/405


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.49...v0.0.50

## [0.0.49](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.49) (2025-12-16)

### What's Changed

* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/389
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/390
* feat(contentful-management-go/testing): implement optimistic locking by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/391
* fix(provider/editor_interface): maintain per-space-environment version offsets on content type changes by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/392


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.48...v0.0.49

## [0.0.48](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.48) (2025-12-11)

### What's Changed

* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/375
* chore(deps): bump peter-evans/create-pull-request from 7.0.9 to 7.0.11 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/379
* fix(provider): normalise json encoded values before saving into state by @Copilot in https://github.com/cysp/terraform-provider-contentful/pull/383
* chore(deps): bump indirect dependencies by @cyspbot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/381
* chore(deps): bump codecov/codecov-action from 5.5.1 to 5.5.2 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/385
* chore(deps): bump peter-evans/create-pull-request from 7.0.11 to 8.0.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/386

### New Contributors

* @cyspbot[bot] made their first contribution in https://github.com/cysp/terraform-provider-contentful/pull/375

**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.47...v0.0.48

## [0.0.47](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.47) (2025-12-07)

### What's Changed

* ref(provider): align naming of model variables to indicate their source by @Copilot in https://github.com/cysp/terraform-provider-contentful/pull/377
* docs(provider): add example for entry list resource by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/378
* fix(provider): ensure resource identities are updated and remove unnecessary error check by @Copilot in https://github.com/cysp/terraform-provider-contentful/pull/376

### New Contributors

* @Copilot made their first contribution in https://github.com/cysp/terraform-provider-contentful/pull/377

**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.46...v0.0.47

## [0.0.46](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.46) (2025-12-06)

### What's Changed

* chore(deps): bump actions/checkout from 5.0.1 to 6.0.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/352
* chore(deps): bump peter-evans/create-pull-request from 7.0.8 to 7.0.9 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/354
* chore(deps): bump golangci/golangci-lint-action from 9.0.0 to 9.1.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/353
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/355
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/357
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/358
* chore(github/workflows): test against terraform 1.14 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/301
* chore(deps): bump github.com/ogen-go/ogen from 1.16.0 to 1.17.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/356
* chore(deps): bump actions/checkout from 6.0.0 to 6.0.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/365
* chore(deps): bump golangci/golangci-lint-action from 9.1.0 to 9.2.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/364
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/363
* chore(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.16.1 to 1.17.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/367
* chore(deps): bump github.com/hashicorp/terraform-plugin-testing from 1.13.3 to 1.14.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/366
* docs(provider): update to current templates, document team and space membership resources and plannable import by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/368
* feat(provider/content_type): implement list resource by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/299
* ref(contentful-management-go): extract a named type for content type collection responses by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/369
* ref(contentful-management-go): enable request options by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/370
* chore(github/workflows): use cyspbot to update indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/371
* ref(provider/content_type): simplify response handling and diagnostics by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/374
* feat(provider/entry): introduce list resource by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/372
* feat(provider/entry): implement taxonomy concept linking by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/373


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.45...v0.0.46

## [0.0.45](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.45) (2025-11-20)

### What's Changed

* chore(deps): bump github.com/go-faster/jx from 1.1.0 to 1.2.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/348
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/346
* chore(deps): bump actions/checkout from 5.0.0 to 5.0.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/347
* chore(deps): bump actions/setup-go from 6.0.0 to 6.1.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/350
* feat(provider/content_type): reduce diff noise of default field values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/351


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.44...v0.0.45

## [0.0.44](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.44) (2025-11-17)

### What's Changed

* ref(provider/app_definition): rename internal resource type to remove duplication  by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/343
* feat(provider): remove deprecated app definition resource provider and resource type by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/342
* fix(github/workflows/update-indirect-dependencies): use go version from go.mod by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/344
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/345


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.43...v0.0.44

## [0.0.43](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.43) (2025-11-17)

### What's Changed

* ref(contentful-management-go): create constructors for link and sys structs by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/337
* ref(contentful-management-go): update casing of api key to match golang style by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/338
* fix(provider/content_type_counter): make concurrency safe by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/339
* ref(provider/content_type_counter): clarify member naming by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/341
* feat(provider): implement team management by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/340


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.42...v0.0.43

## [0.0.42](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.42) (2025-11-15)

### What's Changed

* chore(deps): bump golangci/golangci-lint-action from 8.0.0 to 9.0.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/326
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/325
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/327
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/328
* chore(deps): bump github.com/hashicorp/terraform-plugin-log from 0.9.0 to 0.10.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/330
* ref(content-management-go): improve schema consistency and simplification by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/332
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/329
* docs(provider): add field descriptions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/334
* chore(github/workflows): update golangci-lint to v2.6.2 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/335
* chore(deps): bump go to v1.25.4 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/336


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.41...v0.0.42

## [0.0.41](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.41) (2025-11-04)

### What's Changed

* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/319
* ref(contentful-management-go): extract schema components into separate files by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/320
* docs: update configuration examples by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/323
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/321
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/324
* feat(provider): introduce environment management by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/322


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.40...v0.0.41

## [0.0.40](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.40) (2025-10-29)

### What's Changed

* chore(deps): bump github.com/ogen-go/ogen from 1.15.2 to 1.16.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/306
* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/305
* chore(deps): bump github.com/hashicorp/terraform-plugin-framework-validators from 0.18.0 to 0.19.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/308
* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/307
* chore(deps): bump github.com/hashicorp/terraform-plugin-docs from 0.23.0 to 0.24.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/309
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/310
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/312
* feat(provider): consistently use state for unknown resource identifiers by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/313
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/314
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/315
* feat(provider): introduce element inspection and mutation by key by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/317
* feat(provider): introduce content entry resource by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/316
* feat(provider/entry): publish entries by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/318


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.39...v0.0.40

## [0.0.39](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.39) (2025-10-06)

### What's Changed

* chore(deps): bump codecov/codecov-action from 5.5.0 to 5.5.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/290
* chore(deps): bump actions/setup-go from 5.5.0 to 6.0.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/289
* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/287
* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/291
* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/292
* chore(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.15.1 to 1.16.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/294
* chore(deps): bump github.com/hashicorp/terraform-plugin-docs from 0.22.0 to 0.23.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/296
* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/293
* chore(deps): bump github.com/ogen-go/ogen from 1.14.0 to 1.15.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/297
* chore(deps): bump github.com/ogen-go/ogen from 1.15.0 to 1.15.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/298
* chore(github/workflows): test against terraform 1.13 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/300
* chore(deps): bump github.com/ogen-go/ogen from 1.15.1 to 1.15.2 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/304
* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/303
* feat(provider): set resource identities by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/302


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.38...v0.0.39

## [0.0.38](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.38) (2025-08-30)

### What's Changed

* docs(resources/contentful_resource_type): correct example resource configuration by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/288


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.37...v0.0.38

## [0.0.37](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.37) (2025-08-30)

### What's Changed

* ref(provider): consolidate model type testcases by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/284
* docs(resources): add missing resource examples by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/285
* ref(provider): rename app_definition_resource_provider and app_definition_resource_type to remove app_definition by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/286


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.36...v0.0.37

## [0.0.36](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.36) (2025-08-30)

### What's Changed

* ref(provider): dot-import provider in provider tests by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/283
* ref(provider): introduce generic typed model object by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/276
* chore(deps): bump actions/attest-build-provenance from 2.4.0 to 3.0.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/282


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.35...v0.0.36

## [0.0.35](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.35) (2025-08-29)

### What's Changed

* chore(deps): bump actions/checkout from 4.2.2 to 5.0.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/268
* chore(deps): bump goreleaser/goreleaser-action from 6.3.0 to 6.4.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/269
* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/267
* chore(github/workflows): pin actions to release commits by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/270
* chore(deps): bump github.com/hashicorp/terraform-plugin-testing from 1.13.2 to 1.13.3 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/271
* chore(deps): bump codecov/codecov-action from 5.4.3 to 5.5.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/273
* chore(deps): bump github.com/stretchr/testify from 1.10.0 to 1.11.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/274
* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/272
* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/275
* ref(contentful-management-go): remove unnecessary special-cased nonexistent resource ids by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/280
* ref(contentful-management-go): store app definitions in a global namespace by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/281
* feat(provider): introduce app signing secrets by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/279
* chore(deps): bump github.com/stretchr/testify from 1.11.0 to 1.11.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/277
* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/278


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.34...v0.0.35

## [0.0.34](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.34) (2025-08-08)

### What's Changed

* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/265
* feat(provider): introduce app definition datasource by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/21
* feat(provider): introduce marketplace app definition datasource by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/266


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.33...v0.0.34

## [0.0.33](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.33) (2025-08-07)

### What's Changed

* chore(deps): bump github.com/hashicorp/go-retryablehttp from 0.7.7 to 0.7.8 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/255
* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/256
* chore(deps): bump github.com/hashicorp/terraform-plugin-docs from 0.21.0 to 0.22.0 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/257
* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/258
* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/259
* chore(deps-dev): bump golangci-lint to v2.3.0 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/261
* ref(contentful-management-go): simplify error response handling for both json and contentful management json content types by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/262
* chore(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.15.0 to 1.15.1 by @dependabot[bot] in https://github.com/cysp/terraform-provider-contentful/pull/263
* ref(contentful-management-go): generate server handlers for test server by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/260
* chore(deps): bump indirect dependencies by @github-actions[bot] in https://github.com/cysp/terraform-provider-contentful/pull/264


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.32...v0.0.33

## [0.0.32](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.32) (2025-06-19)

### What's Changed

* fix(provider/resource_app_definition): correct type of location field items to be a single object by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/254


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.31...v0.0.32

## [0.0.31](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.31) (2025-06-19)

### What's Changed

* fix(contentful-management-go): allow arbitrary strings for app definition locations by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/253


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.30...v0.0.31

## [0.0.30](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.30) (2025-06-15)

### What's Changed

* fix(provider/resource_extension): correct optionality of field type items by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/252


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.29...v0.0.30

## [0.0.29](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.29) (2025-06-15)

### What's Changed

* feat(provider): implement ui extension management by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/213


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.28...v0.0.29

## [0.0.28](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.28) (2025-06-14)

### What's Changed

* chore(provider): extract schema definitions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/250
* feat(provider): implement resource identity by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/251


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.27...v0.0.28

## [0.0.27](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.27) (2025-06-14)

### What's Changed

* chore(deps): bump indirect dependencies by @github-actions in https://github.com/cysp/terraform-provider-contentful/pull/241
* fix(contentful-management-testserver): preserve nil values when converting maps and slices by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/248
* ref(provider): introduce use-state-for-unknown plan modifier that preserves known nulls by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/247
* chore(deps): bump github.com/hashicorp/terraform-plugin-testing from 1.13.1 to 1.13.2 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/246
* chore(deps): bump indirect dependencies by @github-actions in https://github.com/cysp/terraform-provider-contentful/pull/243
* feat(provider): implement app definition management by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/244
* chore(deps): bump actions/attest-build-provenance from 2.3.0 to 2.4.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/245


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.26...v0.0.27

## [0.0.26](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.26) (2025-06-10)

### What's Changed

* chore(deps): bump indirect dependencies by @github-actions in https://github.com/cysp/terraform-provider-contentful/pull/239
* feat(provider): extract detailed error reasons by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/113
* fix(provider/personal_access_token): preserve existing known token values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/242


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.25...v0.0.26

## [0.0.25](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.25) (2025-06-07)

### What's Changed

* chore(deps): bump indirect dependencies by @github-actions in https://github.com/cysp/terraform-provider-contentful/pull/234
* chore(provider): treat mocked acceptance tests as acceptance tests by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/236
* chore(deps): bump github.com/ogen-go/ogen from 1.13.0 to 1.14.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/238
* chore(deps): bump indirect dependencies by @github-actions in https://github.com/cysp/terraform-provider-contentful/pull/237
* feat(contentful-management-go): allow json or management response content types for errors by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/240

### New Contributors

* @github-actions made their first contribution in https://github.com/cysp/terraform-provider-contentful/pull/234

**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.24...v0.0.25

## [0.0.24](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.24) (2025-06-01)

### What's Changed

* chore(deps): bump actions/setup-go from 5.4.0 to 5.5.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/219
* ref: simplify 2xx response decoding by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/218
* chore(deps): bump github.com/hashicorp/terraform-plugin-framework-validators from 0.17.0 to 0.18.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/220
* chore(deps): bump codecov/codecov-action from 5.4.2 to 5.4.3 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/221
* chore(deps): bump github.com/hashicorp/terraform-plugin-go from 0.26.0 to 0.27.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/224
* chore(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.14.1 to 1.15.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/222
* chore(deps): bump github.com/hashicorp/terraform-plugin-testing from 1.12.0 to 1.13.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/223
* chore(deps): bump github.com/ogen-go/ogen from 1.12.0 to 1.13.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/225
* chore(deps): bump github.com/hashicorp/terraform-plugin-go from 0.27.0 to 0.28.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/229
* chore(deps): bump github.com/hashicorp/terraform-plugin-testing from 1.13.0 to 1.13.1 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/228
* ref(provider): recreate entire resource models on each refresh by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/230
* feat(provider/content_type): model content type metadata by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/227
* chore(github/workflows): introduce scheduled action to update indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/232
* chore(github/workflows/update-indirect-dependencies): ignore errors when updating modules by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/235
* feat(provider): extract detailed validation failure error messages by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/233


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.23...v0.0.24

## [0.0.23](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.23) (2025-05-08)

### What's Changed

* chore(deps): bump golangci/golangci-lint-action from 7.0.0 to 8.0.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/216
* feat(provider): manage app resource provider definitions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/217
* chore(deps): bump actions/attest-build-provenance from 2.2.3 to 2.3.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/212


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.22...v0.0.23

## [0.0.22](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.22) (2025-05-06)

### What's Changed

* chore(github/workflows/golangci): bump golangci-lint to v2.1.5 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/211
* feat(provider/content_type): support cross-space and external resource link fields by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/215


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.21...v0.0.22

## [0.0.21](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.21) (2025-04-29)

### What's Changed

* chore(deps): bump crazy-max/ghaction-import-gpg from 6.2.0 to 6.3.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/197
* chore(deps): bump goreleaser/goreleaser-action from 6.2.1 to 6.3.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/196
* chore(deps): bump codecov/codecov-action from 5.4.0 to 5.4.2 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/200
* chore(deps): bump golang.org/x/net from 0.37.0 to 0.38.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/201
* ref(provider): replace basetypes with types equivalents by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/202
* feat(provider): implement generic typed list value type by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/199
* chore(deps): bump github.com/ogen-go/ogen from 1.10.1 to 1.12.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/204
* feat(provider): implement generic typed map value type by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/205
* fix(provider/typedlist): correctly handle conversion of null and unknown list values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/207
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/208
* test(provider): test to-object-value, value-from-object roundtrips by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/209
* feat(provider/editor_interface): model editor layout structure by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/198
* chore(github/workflows/test): test with terraform 1.11 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/210


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.20...v0.0.21

## [0.0.20](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.20) (2025-03-28)

### What's Changed

* chore(deps): bump golangci/golangci-lint-action from 6.5.1 to 6.5.2 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/190
* chore(deps): bump actions/setup-go from 5.3.0 to 5.4.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/191
* chore(deps): bump github.com/hashicorp/terraform-plugin-testing from 1.11.0 to 1.12.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/192
* chore(deps): bump golangci/golangci-lint-action from 6.5.2 to 7.0.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/193
* Support space enablements by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/195


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.19...v0.0.20

## [0.0.19](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.19) (2025-03-17)

### What's Changed

* chore(deps): bump github.com/ogen-go/ogen from 1.10.0 to 1.10.1 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/174
* chore(deps): bump go to 1.24.1 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/175
* chore(deps): bump golangci/golangci-lint-action from 6.5.0 to 6.5.1 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/176
* ref(provider): derive object attr types from schema attributes by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/177
* chore(provider/content_type): remove unused to-object-value-must function by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/179
* ref(provider/content_type_field_items): derive object attr types from schema attributes by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/180
* ref(provider): introduce separate terraform-plugin-framework-reflection package by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/178
* fix(provider): implement string function for all value types by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/181
* feat(terraform-plugin-framework-reflection): make unexpected value state errors non-fatal by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/182
* doc(webhook): provide example resource configuration by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/183
* test(terraform-plugin-framework-reflection): assert behaviour given invalid value type by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/184
* chore(provider): remove unused value-type functions from type implementations by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/185
* feat(provider): remove hand-rolled panicking must functions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/187
* feat(provider/webhook): respect identifiers in responses and synthesise a standard resource id attribute by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/188


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.18...v0.0.19

## [0.0.18](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.18) (2025-03-11)

### What's Changed

* chore(deps): bump golangci/golangci-lint-action from 6.3.0 to 6.3.2 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/130
* chore(deps): bump goreleaser/goreleaser-action from 6.1.0 to 6.2.1 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/131
* chore(deps): bump golangci/golangci-lint-action from 6.3.2 to 6.3.3 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/133
* chore(deps): bump go to 1.24 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/132
* chore(github/workflows): test against terraform 1.10 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/134
* chore(deps): bump golangci/golangci-lint-action from 6.3.3 to 6.4.1 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/135
* chore(codecov): update comment style by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/137
* ref(provider): reorganise provider instantiation to allow injection of http client by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/136
* ref(provider): shorten contentful-management-go's import alias to cm by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/139
* chore(deps): bump golangci/golangci-lint-action from 6.4.1 to 6.5.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/140
* chore(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.13.0 to 1.14.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/142
* chore(codecov): include terraform codegenerated code by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/143
* ref(provider/webhook): use hand-rolled header type by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/141
* ref(provider/webhook): use hand-rolled transformation type by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/145
* ref(provider/webhook): use hand-rolled resource model by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/146
* ref(provider/webhook): reorganise by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/147
* chore(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.14.0 to 1.14.1 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/144
* ref(provider): extract test input files by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/128
* chore(deps): bump codecov/codecov-action from 5.3.1 to 5.4.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/149
* chore(deps): bump github.com/hashicorp/terraform-plugin-docs from 0.20.1 to 0.21.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/150
* ref(contentful-management-go): simplify request payload types by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/151
* feat(contentful-management-go): decode space, environment, resource links by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/152
* ref(contentful-management-go): simplify base-sys-with-space-environment by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/153
* feat(provider): introduce mockability of resource tests by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/154
* test(provider/role): implement mocked role resource test by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/155
* ref(provider/role): use hand-rolled resource model by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/156
* ref(provider/app_installation): use hand-rolled model type by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/157
* ref(provider/personal_access_token): use hand-rolled model type by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/159
* ref(provider/delivery_api_key): use hand-rolled model type by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/158
* ref(provider/preview_api_key): use hand-rolled model type by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/160
* ref(provider/content_type): use hand-rolled model types by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/161
* feat(provider/content_type): respect identifiers in responses and synthesise a standard resource id attribute by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/163
* ref(provider): use reflection to construct model values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/162
* test(provider/delivery_api_key): clean up after introduction of synthesised id attribute by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/166
* feat(provider/role_model): respect identifiers in responses and synthesise a standard resource id attribute by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/165
* ref(provider/editor_interface): use hand-rolled model types by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/164
* feat(provider): use hand-rolled provider model type, remove tfplugingen by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/168
* chore(goreleaser): update configuration for archives.formats by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/167
* ref(provider): colocate model and schema definitions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/169
* ref(provider): derive terraform attribute types from object attribute types by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/170
* chore(github/workflows/golangci): report all issues, address them by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/171
* chore(github/workflows/golangci): bump golangci-lint to v1.64.6 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/172
* chore(workflows/release): attest built artifacts by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/116
* chore(contentful-management-testserver): remove unused error response code, test user handler by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/173


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.17...v0.0.18

## [0.0.17](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.17) (2025-02-09)

### What's Changed

* Remove deprecated golangci linters, use jsoneq assertions where sensible by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/115
* chore(github/workflows): setup terraform before generating code by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/118
* chore(deps): bump golangci/golangci-lint-action from 6.1.1 to 6.2.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/117
* chore(deps): bump codecov/codecov-action from 5.1.2 to 5.2.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/121
* chore(deps): bump actions/setup-go from 5.2.0 to 5.3.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/119
* chore(deps): bump github.com/hashicorp/terraform-plugin-go from 0.25.0 to 0.26.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/120
* chore(deps): bump codecov/codecov-action from 5.2.0 to 5.3.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/122
* chore(deps): bump codecov/codecov-action from 5.3.0 to 5.3.1 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/123
* chore(deps): bump github.com/ogen-go/ogen from 1.8.1 to 1.9.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/124
* chore(deps): bump golangci/golangci-lint-action from 6.2.0 to 6.3.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/125
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/126
* chore(deps): bump github.com/ogen-go/ogen from 1.9.0 to 1.10.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/127
* feat: introduce webhook management by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/114


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.16...v0.0.17

## [0.0.16](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.16) (2024-12-21)

### What's Changed

* test(provider/preview_api_key): test reading a nonexistent api key by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/90
* chore(deps): bump codecov/codecov-action from 4.6.0 to 5.0.2 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/92
* chore(deps): bump github.com/hashicorp/terraform-plugin-testing from 1.10.0 to 1.11.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/94
* chore(deps): bump codecov/codecov-action from 5.0.2 to 5.0.7 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/95
* chore(deps): bump github.com/stretchr/testify from 1.9.0 to 1.10.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/96
* chore(deps): bump github.com/hashicorp/terraform-plugin-docs from 0.20.0 to 0.20.1 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/97
* chore(deps): bump actions/setup-go from 5.1.0 to 5.2.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/102
* chore(deps): bump codecov/codecov-action from 5.0.7 to 5.1.1 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/101
* chore(deps): bump golang.org/x/crypto from 0.29.0 to 0.31.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/103
* chore(deps): bump github.com/ogen-go/ogen from 1.6.0 to 1.8.1 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/100
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/104
* chore(deps): bump codecov/codecov-action from 5.1.1 to 5.1.2 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/107
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/106
* ref(contentful-management-go): hoist parameter declarations by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/105
* chore(codecov): configure components by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/16
* feat(provider/content_type): introduce default argument values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/109
* feat(provider/editor_interface): reduce computed arguments by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/110
* Implement editor interface layouts and groups by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/37
* feat(provider/delivery_api_key): preserve state values for stable attributes by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/111


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.15...v0.0.16

## [0.0.15](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.15) (2024-11-10)

### What's Changed

* chore(deps): bump github.com/ogen-go/ogen from 1.5.0 to 1.6.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/82
* chore(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.12.0 to 1.13.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/83
* test(provider/environment_links): cover with unit tests by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/84
* chore(deps): bump github.com/hashicorp/terraform-plugin-docs from 0.19.4 to 0.20.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/86
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/87
* chore(deps): bump goreleaser/goreleaser-action from 6.0.0 to 6.1.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/88
* feat: make opaque json fragment fields more opaque by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/89
* Implement role resource by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/85


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.14...v0.0.15

## [0.0.14](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.14) (2024-10-31)

### What's Changed

* chore(deps): bump crazy-max/ghaction-import-gpg from 6.1.0 to 6.2.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/76
* feat(provider): add missing resource deletion request logging by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/78
* chore(deps): bump github.com/hashicorp/terraform-plugin-go from 0.24.0 to 0.25.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/79
* fix(provider/delivery_api_key_resource): mark preview_api_key_id as read-only by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/80


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.13...v0.0.14

## [0.0.13](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.13) (2024-10-26)

### What's Changed

* docs(resources/personal_access_token): add example resource definition and import by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/75


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.12...v0.0.13

## [0.0.12](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.12) (2024-10-26)

### What's Changed

* chore(deps): bump actions/checkout from 4.2.1 to 4.2.2 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/70
* chore(deps): bump github.com/ogen-go/ogen from 1.4.1 to 1.5.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/69
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/71
* chore(deps): bump actions/setup-go from 5.0.2 to 5.1.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/72
* ref: name delivery api key operation ids more semantically by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/73
* feat(provider): implement personal access token management by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/74


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.11...v0.0.12

## [0.0.11](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.11) (2024-10-14)

### What's Changed

* docs(provider): document provider and entities by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/68


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.10...v0.0.11

## [0.0.10](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.10) (2024-10-12)

### What's Changed

* chore(codecov): configure coverage thresholds by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/64
* chore(deps): bump golangci/golangci-lint-action from 6.1.0 to 6.1.1 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/66
* chore(deps): bump codecov/codecov-action from 4.5.0 to 4.6.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/65
* chore(deps): bump actions/checkout from 4.2.0 to 4.2.1 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/67
* feat(provider): introduce api key management by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/42


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.9...v0.0.10

## [0.0.9](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.9) (2024-09-29)

### What's Changed

* chore(deps): bump github.com/ogen-go/ogen from 1.2.2 to 1.3.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/49
* chore(deps): bump hashicorp/setup-terraform from 3.1.1 to 3.1.2 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/50
* chore(deps): bump github.com/ogen-go/ogen from 1.3.0 to 1.4.1 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/53
* chore(deps): bump golangci to 1.61.0 by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/55
* chore(deps): bump go to 1.23.1, bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/54
* chore(deps): bump github.com/hashicorp/terraform-plugin-framework-jsontypes from 0.1.0 to 0.2.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/52
* chore(deps): bump github.com/hashicorp/terraform-plugin-go from 0.23.0 to 0.24.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/56
* chore(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.11.0 to 1.12.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/57
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/58
* chore(deps): bump actions/checkout from 4.1.7 to 4.2.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/63
* chore(deps): bump github.com/hashicorp/terraform-plugin-codegen-framework from 0.4.0 to 0.4.1 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/61
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/59
* Support installation of Contentful Marketplace Apps by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/62


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.8...v0.0.9

## [0.0.8](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.8) (2024-08-12)

### What's Changed

* chore(deps): bump transitive dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/41
* chore(deps): bump github.com/hashicorp/terraform-plugin-testing from 1.9.0 to 1.10.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/43
* fix(provider/content_type): check for errors creating request body by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/45
* fix(provider/app_installation): use correct resource name in logs by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/46
* chore(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.10.0 to 1.11.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/40
* ref(provider): reorganise generated code into provider package by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/47
* feat(provider): mark provider access token as a sensitive value by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/48


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.7...v0.0.8

## [0.0.7](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.7) (2024-08-07)

### What's Changed

* ref(contentful-management-go): reorganise util code by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/36
* ref(provider): consistently return diags from response decoding functions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/35
* feat(contentful-management-go): allow content type descriptions to be nullable in responses by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/38
* feat(contentful-management-go): allow content type display fields to be nullable in responses by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/39


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.6...v0.0.7

## [0.0.6](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.6) (2024-07-31)

### What's Changed

* feat(provider): keep track of content-type update counts, use them as offsets to editor-interface versions by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/34


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.5...v0.0.6

## [0.0.5](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.5) (2024-07-31)

### What's Changed

* ref(provider): standardise model conversion implementations by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/27
* feat(provider): use retryablehttp client by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/28
* chore(deps): bump golangci/golangci-lint-action from 6.0.1 to 6.1.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/31
* feat(provider): introduce content type resource by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/20
* ref(provider/editor_interface_resource): combine request error handling codepaths by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/30
* ref(provider/app_installation_resource): combine request error handling codepaths by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/29
* feat(contentful-management-go): decode arbitrary error details by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/32
* feat(provider/util/jx): escape string values by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/33
* ref(provider/util): support storing any json marshallable value in provider private data by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/11


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.4...v0.0.5

## [0.0.4](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.4) (2024-07-23)

### What's Changed

* feat(contentful-management-go): make error message optional by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/22
* feat(provider): introduce diagnostic logging of requests and responses by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/23
* feat(contentful-management-go): mark editor interface controls and sidebar nullable, remove defaults by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/24
* ref(contentful-management-go): tidy base sys hierarchy by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/25
* chore(deps): bump indirect dependencies by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/26


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.3...v0.0.4

## [0.0.3](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.3) (2024-07-16)

### What's Changed

* chore(codecov): ignore all generate files by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/17
* test(provider/app_installation_model): test body creation with invalid parameters by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/15
* test(provider/app_installation_resource): test importing nonexistent app installation by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/14
* chore(codecov): ignore main.go by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/18
* chore(provider/app_installation_resource): delete commented-out test assertion by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/19
* feat(provider): introduce editor_interface resource by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/7


**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.2...v0.0.3

## [0.0.2](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.2) (2024-07-12)

### What's Changed

* chore(deps): bump github.com/hashicorp/terraform-plugin-framework from 1.9.0 to 1.10.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/9
* chore(deps): bump github.com/hashicorp/terraform-plugin-testing from 1.8.0 to 1.9.0 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/10
* feat(provider/util): introduce utility for storing integers in private resource data by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/8
* chore(deps): bump actions/setup-go from 5.0.1 to 5.0.2 by @dependabot in https://github.com/cysp/terraform-provider-contentful/pull/12
* fix(provider/app_installation): implement stable encoding of parameters object by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/13

### New Contributors

* @dependabot made their first contribution in https://github.com/cysp/terraform-provider-contentful/pull/9

**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/compare/v0.0.1...v0.0.2

## [0.0.1](https://github.com/cysp/terraform-provider-contentful/releases/tag/v0.0.1) (2024-07-06)

### What's Changed

* feat: introduce contentful provider by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/1
* chore(test): collect coverage from all packages by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/4
* feat(contentful-management-go): introduce client for contentful management api by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/3
* feat(provider): introduce app_installation resource by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/5
* chore(github/workflows): configure release workflow by @cysp in https://github.com/cysp/terraform-provider-contentful/pull/6

### New Contributors

* @cysp made their first contribution in https://github.com/cysp/terraform-provider-contentful/pull/1

**Full Changelog**: https://github.com/cysp/terraform-provider-contentful/commits/v0.0.1
