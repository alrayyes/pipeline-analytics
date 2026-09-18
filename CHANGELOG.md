# Changelog

## [0.32.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.31.1...v0.32.0) (2026-09-18)


### Features

* **settings:** persist theme and filter settings per account ([#208](https://github.com/alrayyes/pipeline-analytics/issues/208)) ([e1370c6](https://github.com/alrayyes/pipeline-analytics/commit/e1370c65207ac604a94ba1b72e0dc43863b32e88))

## [0.31.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.31.0...v0.31.1) (2026-09-18)


### Bug Fixes

* **repos:** reject archived, forked, and mirror repos at registration ([#201](https://github.com/alrayyes/pipeline-analytics/issues/201)) ([81d6031](https://github.com/alrayyes/pipeline-analytics/commit/81d603192a41969d5fd46cf7ba2d680e4b380887))

## [0.31.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.30.1...v0.31.0) (2026-09-18)


### Features

* **http:** log every request, not almost nothing ([#200](https://github.com/alrayyes/pipeline-analytics/issues/200)) ([f7df62b](https://github.com/alrayyes/pipeline-analytics/commit/f7df62bbe2e98a76933f133d80dc89ea5d498174)), closes [#196](https://github.com/alrayyes/pipeline-analytics/issues/196)

## [0.30.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.30.0...v0.30.1) (2026-09-18)


### Bug Fixes

* **repos:** reject duplicate repo registration clearly, not opaquely ([#199](https://github.com/alrayyes/pipeline-analytics/issues/199)) ([2aa6272](https://github.com/alrayyes/pipeline-analytics/commit/2aa62722dcf25ab29bc3be655005c2c65e2f8c5d)), closes [#197](https://github.com/alrayyes/pipeline-analytics/issues/197)

## [0.30.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.29.2...v0.30.0) (2026-09-18)


### Features

* **auth:** add revocable API token auth alongside sessions ([#186](https://github.com/alrayyes/pipeline-analytics/issues/186)) ([6a03692](https://github.com/alrayyes/pipeline-analytics/commit/6a036926ba3d7a6d1b653bcd56537fc509b103c1))

## [0.29.2](https://github.com/alrayyes/pipeline-analytics/compare/v0.29.1...v0.29.2) (2026-09-17)


### Bug Fixes

* **docs:** scope the screenshot capture script's health-filter click ([#183](https://github.com/alrayyes/pipeline-analytics/issues/183)) ([74ca888](https://github.com/alrayyes/pipeline-analytics/commit/74ca888306b46e3c70cb8d17b71ddc3e2afc2fd0)), closes [#182](https://github.com/alrayyes/pipeline-analytics/issues/182)

## [0.29.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.29.0...v0.29.1) (2026-09-17)


### Bug Fixes

* **release:** deduplicate changelog and release notes ([#180](https://github.com/alrayyes/pipeline-analytics/issues/180)) ([adadd1b](https://github.com/alrayyes/pipeline-analytics/commit/adadd1bf1957147d5d0fc8c03f1070eb35f8c525)), closes [#179](https://github.com/alrayyes/pipeline-analytics/issues/179)

## [0.29.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.28.1...v0.29.0) (2026-09-17)


### Features

* **web:** add humans.txt ([604c13f](https://github.com/alrayyes/pipeline-analytics/commit/604c13f4dd553166585f939c86e7694ad42f7a6f)), closes [#175](https://github.com/alrayyes/pipeline-analytics/issues/175)

## [0.28.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.28.0...v0.28.1) (2026-09-17)


### Bug Fixes

* **e2e:** give every test its own isolated server instead of one shared ([debb96f](https://github.com/alrayyes/pipeline-analytics/commit/debb96f58a2e31ccbf3b813885da5e2e1e397765)), closes [#168](https://github.com/alrayyes/pipeline-analytics/issues/168)
* **nav:** collapse the header behind a menu toggle on phone widths ([2e7b290](https://github.com/alrayyes/pipeline-analytics/commit/2e7b290e2da0495ba9875f30cb274fd46382d28d)), closes [#171](https://github.com/alrayyes/pipeline-analytics/issues/171)
* **web:** add @types/node for the e2e helpers' node: imports ([7890e6a](https://github.com/alrayyes/pipeline-analytics/commit/7890e6ac18971c155de141bb94f49a56440ef418))

## [0.28.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.27.1...v0.28.0) (2026-09-17)


### Features

* **openapi:** document pagination on GET /api/repos ([#165](https://github.com/alrayyes/pipeline-analytics/issues/165)) ([75d7743](https://github.com/alrayyes/pipeline-analytics/commit/75d774312b266e106bb8dbdd263cdf0fea40cc79))

## [0.27.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.27.0...v0.27.1) (2026-09-17)


### Bug Fixes

* **docker:** bump pinned ca-certificates to 20260909-r0 ([#167](https://github.com/alrayyes/pipeline-analytics/issues/167)) ([5b61c8a](https://github.com/alrayyes/pipeline-analytics/commit/5b61c8a204d8d946d16bf67015110b77d89c2c2c))

## [0.27.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.26.0...v0.27.0) (2026-09-17)


### Features

* **dashboard-ui:** add a cross-pipeline overview of unhealthy steps ([#159](https://github.com/alrayyes/pipeline-analytics/issues/159)) ([38f0aac](https://github.com/alrayyes/pipeline-analytics/commit/38f0aac444cd490b734e1fb4e149cd7d14a1bf28)), closes [#150](https://github.com/alrayyes/pipeline-analytics/issues/150)

## [0.26.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.25.0...v0.26.0) (2026-09-17)


### Features

* **dashboard-ui:** filter and sort the Pipelines list by repo, health, and last run ([#158](https://github.com/alrayyes/pipeline-analytics/issues/158)) ([ec70ca3](https://github.com/alrayyes/pipeline-analytics/commit/ec70ca38c372813585d1197b38dd477080b4bc2d))

## [0.25.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.24.1...v0.25.0) (2026-09-17)


### Features

* **dashboard-ui:** add a GitHub API rate-limit insights page ([#161](https://github.com/alrayyes/pipeline-analytics/issues/161)) ([5d55751](https://github.com/alrayyes/pipeline-analytics/commit/5d5575162f058e1808092b66e50517cea3f56e1b)), closes [#160](https://github.com/alrayyes/pipeline-analytics/issues/160)

## [0.24.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.24.0...v0.24.1) (2026-09-17)


### Bug Fixes

* **dashboard-ui:** truncate long step names instead of widening the Steps table ([#156](https://github.com/alrayyes/pipeline-analytics/issues/156)) ([91ef4e6](https://github.com/alrayyes/pipeline-analytics/commit/91ef4e63cbb02eebb1c1913747fe643bc2009c89)), closes [#148](https://github.com/alrayyes/pipeline-analytics/issues/148)

## [0.24.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.23.2...v0.24.0) (2026-09-17)


### Features

* **dashboard-ui:** add a persistent all/GitHub/Forgejo filter, group both list pages by forge ([#143](https://github.com/alrayyes/pipeline-analytics/issues/143)) ([92beeef](https://github.com/alrayyes/pipeline-analytics/commit/92beeefd97683641f5bfe95310dc84be65f6c0a6)), closes [#140](https://github.com/alrayyes/pipeline-analytics/issues/140)

## [0.23.2](https://github.com/alrayyes/pipeline-analytics/compare/v0.23.1...v0.23.2) (2026-09-17)


### Bug Fixes

* **dashboard-ui:** click through the unhealthy-only default before capturing screenshots ([#137](https://github.com/alrayyes/pipeline-analytics/issues/137)) ([3c4dfae](https://github.com/alrayyes/pipeline-analytics/commit/3c4dfae832f6652dd8153e4a21532a67975596ae)), closes [#136](https://github.com/alrayyes/pipeline-analytics/issues/136)
* **dashboard-ui:** fix horizontal overflow on the Pipelines list page ([#139](https://github.com/alrayyes/pipeline-analytics/issues/139)) ([2c778c1](https://github.com/alrayyes/pipeline-analytics/commit/2c778c1e6a3477631baa88718775231efb52952e)), closes [#135](https://github.com/alrayyes/pipeline-analytics/issues/135)

## [0.23.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.23.0...v0.23.1) (2026-09-16)


### Bug Fixes

* **forgejo:** tolerate the jobs-listing response shape mismatch ([#131](https://github.com/alrayyes/pipeline-analytics/issues/131)) ([653b422](https://github.com/alrayyes/pipeline-analytics/commit/653b4228807c392b05ba6401c966fdaa043bc749))

## [0.23.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.22.1...v0.23.0) (2026-09-16)


### Features

* **repos:** exclude forks, mirrors, and archived repos from the picker ([#125](https://github.com/alrayyes/pipeline-analytics/issues/125)) ([62c239b](https://github.com/alrayyes/pipeline-analytics/commit/62c239bbbd875b3e8d53bf0d0850c411336aebe6))

## [0.22.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.22.0...v0.22.1) (2026-09-16)


### Bug Fixes

* **forgejo:** make webhook events and reconciliation match a real instance ([#124](https://github.com/alrayyes/pipeline-analytics/issues/124)) ([449e2b3](https://github.com/alrayyes/pipeline-analytics/commit/449e2b3b819c4a91bd5b7cc6f4ac78e09e8d1790))

## [0.22.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.21.1...v0.22.0) (2026-09-16)


### Features

* **dashboard-ui:** one-click saved token, exclude already-tracked repos ([#118](https://github.com/alrayyes/pipeline-analytics/issues/118)) ([739f6f6](https://github.com/alrayyes/pipeline-analytics/commit/739f6f61898d919fee9abd1a529aaa06ac11d0bc))

## [0.21.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.21.0...v0.21.1) (2026-09-16)


### Bug Fixes

* **dashboard-ui:** fix horizontal page overflow on the pipeline detail and nav ([#115](https://github.com/alrayyes/pipeline-analytics/issues/115)) ([8b8d3ce](https://github.com/alrayyes/pipeline-analytics/commit/8b8d3ce6657eda075bc0acafb11de69318f3b6ff))

## [0.21.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.20.0...v0.21.0) (2026-09-16)


### Features

* **dashboard-ui:** color the performance-history charts, wire up their tooltip ([#111](https://github.com/alrayyes/pipeline-analytics/issues/111)) ([e5e09f7](https://github.com/alrayyes/pipeline-analytics/commit/e5e09f772e8987d8d59d999fa0ea75877110c883)), closes [#110](https://github.com/alrayyes/pipeline-analytics/issues/110)

## [0.20.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.19.0...v0.20.0) (2026-09-16)


### Features

* **dashboard-ui:** connect a token once, then multi-select repos to follow ([#108](https://github.com/alrayyes/pipeline-analytics/issues/108)) ([ed1c78c](https://github.com/alrayyes/pipeline-analytics/commit/ed1c78ca3d6665f4a5f245c1b5dd6a02ea08ee68)), closes [#103](https://github.com/alrayyes/pipeline-analytics/issues/103)

## [0.19.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.18.0...v0.19.0) (2026-09-16)


### Features

* **dashboard-ui:** group the Pipelines list by repository ([#106](https://github.com/alrayyes/pipeline-analytics/issues/106)) ([63ebd51](https://github.com/alrayyes/pipeline-analytics/commit/63ebd51a4f436bdc66fc5979eb9570ddaa42edf6)), closes [#102](https://github.com/alrayyes/pipeline-analytics/issues/102)

## [0.18.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.17.0...v0.18.0) (2026-09-16)


### Features

* **dashboard-ui:** color-code health status and default to unhealthy-only ([#104](https://github.com/alrayyes/pipeline-analytics/issues/104)) ([08a8295](https://github.com/alrayyes/pipeline-analytics/commit/08a8295cdef30251b9c9bbeeeb34e71e662888a6)), closes [#101](https://github.com/alrayyes/pipeline-analytics/issues/101)

## [0.17.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.16.1...v0.17.0) (2026-09-16)


### Features

* **dashboard-ui:** make the dashboard installable as a PWA ([#99](https://github.com/alrayyes/pipeline-analytics/issues/99)) ([def3669](https://github.com/alrayyes/pipeline-analytics/commit/def3669f6e1437c86649aff6be67dce928377d49)), closes [#77](https://github.com/alrayyes/pipeline-analytics/issues/77)

## [0.16.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.16.0...v0.16.1) (2026-09-16)


### Bug Fixes

* embed the real git tag in version, not goreleaser's bare semver ([#97](https://github.com/alrayyes/pipeline-analytics/issues/97)) ([30eefc7](https://github.com/alrayyes/pipeline-analytics/commit/30eefc7124989b57d0f0ba41a1c3cb2d8bcda2a3)), closes [#95](https://github.com/alrayyes/pipeline-analytics/issues/95)

## [0.16.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.15.2...v0.16.0) (2026-09-16)


### Features

* add a configurable log level ([#94](https://github.com/alrayyes/pipeline-analytics/issues/94)) ([0e088d3](https://github.com/alrayyes/pipeline-analytics/commit/0e088d385aed894f97aa929553583a1ebe07cca6)), closes [#75](https://github.com/alrayyes/pipeline-analytics/issues/75)

## [0.15.2](https://github.com/alrayyes/pipeline-analytics/compare/v0.15.1...v0.15.2) (2026-09-16)


### Bug Fixes

* **ingestion:** paginate repo discovery, exclude archived/forked repos ([#92](https://github.com/alrayyes/pipeline-analytics/issues/92)) ([c12db45](https://github.com/alrayyes/pipeline-analytics/commit/c12db45b0fa47632abfc5ab4cbba3adb1be596e7)), closes [#74](https://github.com/alrayyes/pipeline-analytics/issues/74)

## [0.15.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.15.0...v0.15.1) (2026-09-16)


### Bug Fixes

* **dashboard-ui:** reflect repo-registration status in nav and empty states ([#89](https://github.com/alrayyes/pipeline-analytics/issues/89)) ([20ff445](https://github.com/alrayyes/pipeline-analytics/commit/20ff4455f3151b26e7fc65364e67aa57fc5ceef9)), closes [#71](https://github.com/alrayyes/pipeline-analytics/issues/71)

## [0.15.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.14.0...v0.15.0) (2026-09-16)


### Features

* **dashboard-ui:** add a privacy & disclaimer page, linked from the footer ([#87](https://github.com/alrayyes/pipeline-analytics/issues/87)) ([7c520cb](https://github.com/alrayyes/pipeline-analytics/commit/7c520cba8f41e5522fa45f2441a717fcbf90a143)), closes [#78](https://github.com/alrayyes/pipeline-analytics/issues/78)

## [0.14.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.13.0...v0.14.0) (2026-09-16)


### Features

* **dashboard-ui:** replace the placeholder favicon with a real logo ([#85](https://github.com/alrayyes/pipeline-analytics/issues/85)) ([2928e41](https://github.com/alrayyes/pipeline-analytics/commit/2928e4101329c15ef571a081848b52f23bbe6c84)), closes [#76](https://github.com/alrayyes/pipeline-analytics/issues/76)

## [0.13.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.12.1...v0.13.0) (2026-09-16)


### Features

* **dashboard-ui:** offer a saved forge token when registering another repo ([#83](https://github.com/alrayyes/pipeline-analytics/issues/83)) ([8b14eba](https://github.com/alrayyes/pipeline-analytics/commit/8b14eba1b36041bf28f4dc716c2f2f09adebf5d4)), closes [#72](https://github.com/alrayyes/pipeline-analytics/issues/72)

## [0.12.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.12.0...v0.12.1) (2026-09-16)


### Bug Fixes

* **dashboard-ui:** align page containers to max-w-4xl ([#79](https://github.com/alrayyes/pipeline-analytics/issues/79)) ([8056227](https://github.com/alrayyes/pipeline-analytics/commit/8056227d6402c480588d187f0289e63272ff38cd)), closes [#69](https://github.com/alrayyes/pipeline-analytics/issues/69)

## [0.12.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.11.0...v0.12.0) (2026-09-16)


### Features

* **dashboard-ui:** render and style the release history page ([#67](https://github.com/alrayyes/pipeline-analytics/issues/67)) ([fed045c](https://github.com/alrayyes/pipeline-analytics/commit/fed045c8a8064de81149bb5583c5749e83d0a274))

## [0.11.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.10.0...v0.11.0) (2026-09-16)


### Features

* **dashboard-ui:** repo picker and per-forge token guidance ([#64](https://github.com/alrayyes/pipeline-analytics/issues/64)) ([437e4b0](https://github.com/alrayyes/pipeline-analytics/commit/437e4b0b7500f07fc7d94aa7216a6a7901f60ec5))

## [0.10.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.9.0...v0.10.0) (2026-09-16)


### Features

* add a dark mode toggle ([#61](https://github.com/alrayyes/pipeline-analytics/issues/61)) ([9c4942a](https://github.com/alrayyes/pipeline-analytics/commit/9c4942ad7782631ea7a509ce30dc31ecfc8320d6))

## [0.9.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.8.0...v0.9.0) (2026-09-16)


### Features

* repo management UI (register, list, untrack) ([#57](https://github.com/alrayyes/pipeline-analytics/issues/57)) ([1d172e2](https://github.com/alrayyes/pipeline-analytics/commit/1d172e292e8e07b42cba1e0caa97c693199e900b))

## [0.8.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.7.0...v0.8.0) (2026-09-16)


### Features

* **dashboard-ui:** repo usage view ([#50](https://github.com/alrayyes/pipeline-analytics/issues/50)) ([360797a](https://github.com/alrayyes/pipeline-analytics/commit/360797a77a4450ce29de7c2d30b543fb43347d57))

## [0.7.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.6.0...v0.7.0) (2026-09-16)


### Features

* **dashboard-ui:** step breakdown with forge deep links ([#48](https://github.com/alrayyes/pipeline-analytics/issues/48)) ([f259323](https://github.com/alrayyes/pipeline-analytics/commit/f2593231853a3a751690325f09fa6a980b491d12))

## [0.6.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.5.3...v0.6.0) (2026-09-16)


### Features

* **dashboard-ui:** pipeline detail view with trend charts ([#46](https://github.com/alrayyes/pipeline-analytics/issues/46)) ([12f8ba9](https://github.com/alrayyes/pipeline-analytics/commit/12f8ba956b6ad4fd476a524724734855e3a7715f))

## [0.5.3](https://github.com/alrayyes/pipeline-analytics/compare/v0.5.2...v0.5.3) (2026-09-16)


### Bug Fixes

* release job's bun needs to read both lockfile formats ([#42](https://github.com/alrayyes/pipeline-analytics/issues/42)) ([f59145a](https://github.com/alrayyes/pipeline-analytics/commit/f59145a9c36c4a77dba7e0cd5691496adcaf4187))

## [0.5.2](https://github.com/alrayyes/pipeline-analytics/compare/v0.5.1...v0.5.2) (2026-09-16)


### Bug Fixes

* hyphenated CLI flags weren't reachable via env vars ([#33](https://github.com/alrayyes/pipeline-analytics/issues/33)) ([c3dbe61](https://github.com/alrayyes/pipeline-analytics/commit/c3dbe61dac391a9c103fc9d61296669b5864dd75))

## [0.5.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.5.0...v0.5.1) (2026-09-16)


### Bug Fixes

* pin bun below 1.4 and downgrade lockfiles to v1 ([#31](https://github.com/alrayyes/pipeline-analytics/issues/31)) ([2f7de37](https://github.com/alrayyes/pipeline-analytics/commit/2f7de37b46d0aaf1bee16df73897d6a6923a00ed))

## [0.5.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.4.0...v0.5.0) (2026-09-15)


### Features

* **dashboard-ui:** pipeline overview page ([#29](https://github.com/alrayyes/pipeline-analytics/issues/29)) ([cb05aa4](https://github.com/alrayyes/pipeline-analytics/commit/cb05aa4f59772b666d73a4c4bad5e17076421009))
* **dashboard-ui:** WebAuthn login/registration flow ([#28](https://github.com/alrayyes/pipeline-analytics/issues/28)) ([cec2ede](https://github.com/alrayyes/pipeline-analytics/commit/cec2edec668ef3ef4675daeed251933faea05bd8))
* **metrics:** pipeline health, trend, ranking, and usage computation ([#25](https://github.com/alrayyes/pipeline-analytics/issues/25)) ([806c9a0](https://github.com/alrayyes/pipeline-analytics/commit/806c9a0fae6706a0933f1d29a564e13f1c465eef))
* **metrics:** wire pipeline and usage API handlers ([#27](https://github.com/alrayyes/pipeline-analytics/issues/27)) ([f128e98](https://github.com/alrayyes/pipeline-analytics/commit/f128e98e4dd278422cb75271adfb3d260876a603))

## [0.4.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.3.0...v0.4.0) (2026-09-15)


### Features

* **ingestion:** reconciliation polling and scheduler wiring ([#24](https://github.com/alrayyes/pipeline-analytics/issues/24)) ([f59494a](https://github.com/alrayyes/pipeline-analytics/commit/f59494a7ff0305ec4ad63761d2bf2c5bb9ed59e2))
* **ingestion:** webhook receivers and run/job/step storage ([#22](https://github.com/alrayyes/pipeline-analytics/issues/22)) ([72ad44f](https://github.com/alrayyes/pipeline-analytics/commit/72ad44fda282b93abcc19550b55783ee02586353))

## [0.3.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.2.0...v0.3.0) (2026-09-15)


### Features

* **auth:** WebAuthn login/registration and session gating ([#20](https://github.com/alrayyes/pipeline-analytics/issues/20)) ([b78a48e](https://github.com/alrayyes/pipeline-analytics/commit/b78a48e942981fd915a16e6334cf6bec21c4547a))

## [0.2.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.1.0...v0.2.0) (2026-09-15)


### Features

* **api:** author OpenAPI spec for v1 endpoints ([#12](https://github.com/alrayyes/pipeline-analytics/issues/12)) ([81b336e](https://github.com/alrayyes/pipeline-analytics/commit/81b336e3315e196235f314843f784ef489d158f8))
* **ingestion:** POST /api/repos registration with real webhook creation ([#15](https://github.com/alrayyes/pipeline-analytics/issues/15)) ([d3c2dcd](https://github.com/alrayyes/pipeline-analytics/commit/d3c2dcdf04987993825ef05a2f1c30bdd6516eef))
* **ingestion:** schema and encrypted credential storage ([#14](https://github.com/alrayyes/pipeline-analytics/issues/14)) ([7f11037](https://github.com/alrayyes/pipeline-analytics/commit/7f110374026d35e2a628aaf3083caa6e802d4bdd))
* release automation (release-please + goreleaser) and Docker image ([#17](https://github.com/alrayyes/pipeline-analytics/issues/17)) ([577da7c](https://github.com/alrayyes/pipeline-analytics/commit/577da7c89e85e8890e947fc3fced79a84eff46cb))
* sveltekit frontend, embedded, with a version footer ([#16](https://github.com/alrayyes/pipeline-analytics/issues/16)) ([c0895bf](https://github.com/alrayyes/pipeline-analytics/commit/c0895bffadffff26e112346f91b6af7cce3bf710))


### Bug Fixes

* **ci:** bump golangci-lint pin past go1.27 skew ([#11](https://github.com/alrayyes/pipeline-analytics/issues/11)) ([0f3400e](https://github.com/alrayyes/pipeline-analytics/commit/0f3400ed0044311850aa740ab07a12cb0ec78e44))
