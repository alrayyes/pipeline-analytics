# Changelog

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
