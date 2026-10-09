# Changelog

## [0.71.3](https://github.com/alrayyes/pipeline-analytics/compare/v0.71.2...v0.71.3) (2026-10-09)


### Bug Fixes

* **ci:** run Go 1.27.2 to clear the standard-library advisories ([#534](https://github.com/alrayyes/pipeline-analytics/issues/534)) ([05bcf6b](https://github.com/alrayyes/pipeline-analytics/commit/05bcf6b3d995d8303ae2dfac37a06ab36b2df3e1))

## [0.71.2](https://github.com/alrayyes/pipeline-analytics/compare/v0.71.1...v0.71.2) (2026-10-09)


### Performance Improvements

* **web:** stop the stylesheet and font blocking first render ([#531](https://github.com/alrayyes/pipeline-analytics/issues/531)) ([e379479](https://github.com/alrayyes/pipeline-analytics/commit/e379479c90087a55267e3acd7d9d3b8021bd338e))

## [0.71.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.71.0...v0.71.1) (2026-10-08)


### Performance Improvements

* **web:** serve the dashboard's assets cached and compressed ([#526](https://github.com/alrayyes/pipeline-analytics/issues/526)) ([96aa4d2](https://github.com/alrayyes/pipeline-analytics/commit/96aa4d20c74685bcb6c34474c8c0af01143b1a1c))

## [0.71.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.70.1...v0.71.0) (2026-10-08)


### Features

* **web:** link the footer to the GitHub repo and the license ([#518](https://github.com/alrayyes/pipeline-analytics/issues/518)) ([86b6b15](https://github.com/alrayyes/pipeline-analytics/commit/86b6b15cd022cc191dc124c03c02e85e490f7aba))

## [0.70.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.70.0...v0.70.1) (2026-10-06)


### Bug Fixes

* **scripts:** give the runs mock the actions the page now reads ([#514](https://github.com/alrayyes/pipeline-analytics/issues/514)) ([8b348e2](https://github.com/alrayyes/pipeline-analytics/commit/8b348e296deda9787062759091bb2f9af9f09b96)), closes [#513](https://github.com/alrayyes/pipeline-analytics/issues/513)

## [0.70.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.69.0...v0.70.0) (2026-10-06)


### Features

* **api:** bound request bodies and request fields ([#439](https://github.com/alrayyes/pipeline-analytics/issues/439)) ([8c863e3](https://github.com/alrayyes/pipeline-analytics/commit/8c863e37a24879b49a403d4b219e77ee00e53c31))
* **api:** drain /readyz on shutdown and cache its result ([#428](https://github.com/alrayyes/pipeline-analytics/issues/428)) ([4f78103](https://github.com/alrayyes/pipeline-analytics/commit/4f78103bdc8fb2e24f46194c32c1b33725f85610))
* **api:** include the job id on run steps ([#450](https://github.com/alrayyes/pipeline-analytics/issues/450)) ([0db96c3](https://github.com/alrayyes/pipeline-analytics/commit/0db96c3c8477175159b23b6b09b0f91d46573fb9))
* **api:** list the actions a run offers ([#484](https://github.com/alrayyes/pipeline-analytics/issues/484)) ([799b9a3](https://github.com/alrayyes/pipeline-analytics/commit/799b9a3727cf2fbaeae783127acd054151b65ad1))
* **api:** list the branches that have runs ([#459](https://github.com/alrayyes/pipeline-analytics/issues/459)) ([ae4e3b1](https://github.com/alrayyes/pipeline-analytics/commit/ae4e3b155da387ceb9b305dec84a2b0f29a40007))
* **api:** re-run and cancel a run on GitHub ([#478](https://github.com/alrayyes/pipeline-analytics/issues/478)) ([534f498](https://github.com/alrayyes/pipeline-analytics/commit/534f498318132f7dcdccf0b87a5e4cb6bc7ba499))
* **api:** return the settings defaults with the settings ([#453](https://github.com/alrayyes/pipeline-analytics/issues/453)) ([17aae1d](https://github.com/alrayyes/pipeline-analytics/commit/17aae1d756a0ad528b955d714eee067d0fa91843))
* **api:** save forge tokens for registration ([#464](https://github.com/alrayyes/pipeline-analytics/issues/464)) ([b3d9126](https://github.com/alrayyes/pipeline-analytics/commit/b3d9126b7bb11dad23d11b6f9f6dfb4e37eec2b3))
* **api:** scope the telemetry reads to one branch ([#458](https://github.com/alrayyes/pipeline-analytics/issues/458)) ([e8e5836](https://github.com/alrayyes/pipeline-analytics/commit/e8e583648141e3ffdcc93a039985ffa4b66083be))
* **api:** serve a job's log tail ([#455](https://github.com/alrayyes/pipeline-analytics/issues/455)) ([8f90e62](https://github.com/alrayyes/pipeline-analytics/commit/8f90e62b9353232cbca048973b59ab43de361173))
* **ingestion:** read a job's log tail from its forge ([#451](https://github.com/alrayyes/pipeline-analytics/issues/451)) ([e2b7181](https://github.com/alrayyes/pipeline-analytics/commit/e2b718138614e31a8ed1f42130b07eec958dd513))
* **mcp:** add the get_job_log tool ([#456](https://github.com/alrayyes/pipeline-analytics/issues/456)) ([911a843](https://github.com/alrayyes/pipeline-analytics/commit/911a843cfb1665173b0b1e904e27ace779a97424))
* **web:** cancel a queued or running run from the Runs view ([#489](https://github.com/alrayyes/pipeline-analytics/issues/489)) ([b7b1ef7](https://github.com/alrayyes/pipeline-analytics/commit/b7b1ef70f3e4f366b5ead399c6def923fa78f480))
* **web:** re-run a concluded run from the Runs view ([#485](https://github.com/alrayyes/pipeline-analytics/issues/485)) ([ba12145](https://github.com/alrayyes/pipeline-analytics/commit/ba12145f49ccd6f42332d5ec374b7dc625d04c68))
* **web:** save forge tokens in settings and use them to register ([#465](https://github.com/alrayyes/pipeline-analytics/issues/465)) ([0f56f0f](https://github.com/alrayyes/pipeline-analytics/commit/0f56f0ff89ea9998939e4dc8d5084f78262a4507))
* **web:** say which token permissions a run action needs ([#505](https://github.com/alrayyes/pipeline-analytics/issues/505)) ([6bd7b17](https://github.com/alrayyes/pipeline-analytics/commit/6bd7b177e4bc47e59b3579464bf875631ff4d730))
* **web:** say which token permissions a run action needs ([#505](https://github.com/alrayyes/pipeline-analytics/issues/505)) ([#509](https://github.com/alrayyes/pipeline-analytics/issues/509)) ([70e5d05](https://github.com/alrayyes/pipeline-analytics/commit/70e5d053e81b3ff5319b41890bd02bc7d29754a0))
* **web:** scope the telemetry views to one branch ([#460](https://github.com/alrayyes/pipeline-analytics/issues/460)) ([deab211](https://github.com/alrayyes/pipeline-analytics/commit/deab2117162cb941e23d741ea817d3a6a39e17aa))
* **web:** show a failed step's log on the run page ([#508](https://github.com/alrayyes/pipeline-analytics/issues/508)) ([c7b130c](https://github.com/alrayyes/pipeline-analytics/commit/c7b130cde3651e4ec52ed643444f8fbd4541d77a))


### Bug Fixes

* bump layerchart ([#474](https://github.com/alrayyes/pipeline-analytics/issues/474)) ([37a8ae9](https://github.com/alrayyes/pipeline-analytics/commit/37a8ae9f7cbdc507a6d1ede34860231d0054be11))
* bump modernc.org/sqlite ([#435](https://github.com/alrayyes/pipeline-analytics/issues/435)) ([c63f098](https://github.com/alrayyes/pipeline-analytics/commit/c63f09869adc38c3a8173b6a037e15766e156597))
* **ci:** read the frontend coverage from the folder the artifact keeps ([#491](https://github.com/alrayyes/pipeline-analytics/issues/491)) ([2cd7884](https://github.com/alrayyes/pipeline-analytics/commit/2cd788498948f8a17f7576d853fd0bb17cc37e80))
* **forgejo:** skip a null run, job or step in a response ([#447](https://github.com/alrayyes/pipeline-analytics/issues/447)) ([4a395df](https://github.com/alrayyes/pipeline-analytics/commit/4a395df17f14a5c1bf0dc80991366a234da4fe69))
* **github:** give each client a connection pool of its own ([#499](https://github.com/alrayyes/pipeline-analytics/issues/499)) ([f03c481](https://github.com/alrayyes/pipeline-analytics/commit/f03c4814f8fd7e1356cb54e8648131da9fce8f88))
* **web:** clear the transitive audit advisories ([#487](https://github.com/alrayyes/pipeline-analytics/issues/487)) ([d74083d](https://github.com/alrayyes/pipeline-analytics/commit/d74083da9c21a8192644bbb288503b4b79809ddd))
* **web:** honour prefers-reduced-motion ([#502](https://github.com/alrayyes/pipeline-analytics/issues/502)) ([782e8a7](https://github.com/alrayyes/pipeline-analytics/commit/782e8a7ffa6b4b6e4e7601c4743dba9be71fbed0))
* **web:** move to SvelteKit 3 and adapter-static 4 ([#493](https://github.com/alrayyes/pipeline-analytics/issues/493)) ([1dc96d1](https://github.com/alrayyes/pipeline-analytics/commit/1dc96d18ca594015ab0949a223eac116a2a0a497))

## [0.69.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.68.0...v0.69.0) (2026-10-03)


### Features

* **web:** redirect the Steps page to the flaky view ([#409](https://github.com/alrayyes/pipeline-analytics/issues/409)) ([ac1193a](https://github.com/alrayyes/pipeline-analytics/commit/ac1193ab70be6b83ced33fb9bd5fd1180d4a1824))

## [0.68.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.67.0...v0.68.0) (2026-10-03)


### Features

* **web:** add the flaky tests view ([#407](https://github.com/alrayyes/pipeline-analytics/issues/407)) ([a84b388](https://github.com/alrayyes/pipeline-analytics/commit/a84b388d7e2167c5bfa45d07eb0f2476524382ab))

## [0.67.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.66.0...v0.67.0) (2026-10-03)


### Features

* **api:** serve the flaky steps list ([#405](https://github.com/alrayyes/pipeline-analytics/issues/405)) ([053e497](https://github.com/alrayyes/pipeline-analytics/commit/053e49777d18e309e72f73b454fb29294289f6d0))

## [0.66.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.65.0...v0.66.0) (2026-10-03)


### Features

* **api:** specify the flaky steps list ([#400](https://github.com/alrayyes/pipeline-analytics/issues/400)) ([98685b7](https://github.com/alrayyes/pipeline-analytics/commit/98685b7265303fea0af62affd6a89ff8174ef3ec))

## [0.65.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.64.0...v0.65.0) (2026-10-03)


### Features

* **mcp:** add tools for run and failure data ([#402](https://github.com/alrayyes/pipeline-analytics/issues/402)) ([cce3e30](https://github.com/alrayyes/pipeline-analytics/commit/cce3e306de92eef7f4dc5e675fb42a29f8076504))

## [0.64.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.63.0...v0.64.0) (2026-10-03)


### Features

* **web:** add a bottom tab bar for phone widths ([#396](https://github.com/alrayyes/pipeline-analytics/issues/396)) ([af62ef1](https://github.com/alrayyes/pipeline-analytics/commit/af62ef167935c58619c929659538bf176e49e7e1))

## [0.63.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.62.0...v0.63.0) (2026-10-03)


### Features

* **api:** report the window the failure insights cover ([#395](https://github.com/alrayyes/pipeline-analytics/issues/395)) ([5552bf0](https://github.com/alrayyes/pipeline-analytics/commit/5552bf0b94ba21001a85f68a2b01c3ac4877e069))

## [0.62.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.61.0...v0.62.0) (2026-10-02)


### Features

* **api:** report which credentials can be revoked ([#393](https://github.com/alrayyes/pipeline-analytics/issues/393)) ([293310f](https://github.com/alrayyes/pipeline-analytics/commit/293310f22bcf4427da165550cd7485fccd956db4))

## [0.61.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.60.0...v0.61.0) (2026-10-02)


### Features

* **api:** filter and sort the pipeline list across every pipeline ([#391](https://github.com/alrayyes/pipeline-analytics/issues/391)) ([4218366](https://github.com/alrayyes/pipeline-analytics/commit/421836692c20d35d83ff47b207ed18510ebd31ff))

## [0.60.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.59.0...v0.60.0) (2026-10-02)


### Features

* **web:** add the failure overview as the landing page ([#389](https://github.com/alrayyes/pipeline-analytics/issues/389)) ([8e11327](https://github.com/alrayyes/pipeline-analytics/commit/8e11327fb35beb9da8e1f62aa15c4d2739f8b7b8))

## [0.59.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.58.0...v0.59.0) (2026-10-02)


### Features

* **api:** report a normalized outcome on every run and step ([#380](https://github.com/alrayyes/pipeline-analytics/issues/380)) ([0153a3e](https://github.com/alrayyes/pipeline-analytics/commit/0153a3e5f9e54fd53dd43af5f6abd77f05524ab3))

## [0.58.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.57.0...v0.58.0) (2026-10-02)


### Features

* **api:** report failure shares and a category breakdown ([#382](https://github.com/alrayyes/pipeline-analytics/issues/382)) ([94f5f38](https://github.com/alrayyes/pipeline-analytics/commit/94f5f381e8a20d22b90dd5ac2f32420b51be2e08)), closes [#381](https://github.com/alrayyes/pipeline-analytics/issues/381)
* **web:** add the root-cause diagnostics view ([#383](https://github.com/alrayyes/pipeline-analytics/issues/383)) ([7f8ca46](https://github.com/alrayyes/pipeline-analytics/commit/7f8ca463f70935f6635209af399cbf23679567e3))
* **web:** link the footer's version to the release history ([#385](https://github.com/alrayyes/pipeline-analytics/issues/385)) ([c1b485e](https://github.com/alrayyes/pipeline-analytics/commit/c1b485ed20bec94adc3841130f3aa44022f16dbd))

## [0.57.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.56.0...v0.57.0) (2026-10-02)


### Features

* **web:** add the runs view with stage progression ([#378](https://github.com/alrayyes/pipeline-analytics/issues/378)) ([bc3004a](https://github.com/alrayyes/pipeline-analytics/commit/bc3004a1a32d1aa3fa2b9108d63df840dba3fbf3))


### Bug Fixes

* **docker:** create /data owned by the container user ([#377](https://github.com/alrayyes/pipeline-analytics/issues/377)) ([5f4fb90](https://github.com/alrayyes/pipeline-analytics/commit/5f4fb908b760df85c32867b44e1c5658ffd8f835)), closes [#363](https://github.com/alrayyes/pipeline-analytics/issues/363)

## [0.56.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.55.0...v0.56.0) (2026-10-02)


### Features

* **web:** load the release history from a bundled JSON file ([#370](https://github.com/alrayyes/pipeline-analytics/issues/370)) ([a7ba1f2](https://github.com/alrayyes/pipeline-analytics/commit/a7ba1f2cf13c817682e2e13a6bab56da47f87c29)), closes [#368](https://github.com/alrayyes/pipeline-analytics/issues/368)

## [0.55.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.54.0...v0.55.0) (2026-10-02)


### Features

* **metrics:** list runs newest first with their steps ([#361](https://github.com/alrayyes/pipeline-analytics/issues/361)) ([3bdb62b](https://github.com/alrayyes/pipeline-analytics/commit/3bdb62b7fbb085a54d115f15a1535b9003588335))

## [0.54.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.53.0...v0.54.0) (2026-10-02)


### Features

* **api:** serve the failure insights ([#365](https://github.com/alrayyes/pipeline-analytics/issues/365)) ([a4c688f](https://github.com/alrayyes/pipeline-analytics/commit/a4c688f2a45100b3bdb73014540b49dbf7aaffa4))
* **health:** report readiness from the container health check ([#364](https://github.com/alrayyes/pipeline-analytics/issues/364)) ([fc9227c](https://github.com/alrayyes/pipeline-analytics/commit/fc9227cb47938b1bece8487519633c72701c78f5))
* **metrics:** compute mean time to recovery ([#353](https://github.com/alrayyes/pipeline-analytics/issues/353)) ([fd055ca](https://github.com/alrayyes/pipeline-analytics/commit/fd055ca4e93fafee3b9c666c1023cb443898a7fe))
* **metrics:** group failed steps and compute the flaky-step ratio ([#360](https://github.com/alrayyes/pipeline-analytics/issues/360)) ([ff66069](https://github.com/alrayyes/pipeline-analytics/commit/ff66069f4f2d35dcb72aab987e03fd0c5799f1e3))
* **settings:** persist the telemetry window per account ([#357](https://github.com/alrayyes/pipeline-analytics/issues/357)) ([c7583b7](https://github.com/alrayyes/pipeline-analytics/commit/c7583b784020204af624eb2f64f161d16f4e6dab))
* **web:** add the telemetry window store ([#366](https://github.com/alrayyes/pipeline-analytics/issues/366)) ([2412b8b](https://github.com/alrayyes/pipeline-analytics/commit/2412b8bc7b26469e723e970b1e2767f07bd5d8db))
* **web:** add typed fetchers for failure insights and the run list ([#356](https://github.com/alrayyes/pipeline-analytics/issues/356)) ([0d5a0af](https://github.com/alrayyes/pipeline-analytics/commit/0d5a0af569c54cf2ee451ebe6c28fd43d7a5c77e))

## [0.53.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.52.0...v0.53.0) (2026-10-02)


### Features

* **metrics:** compute pass rate, its change and top failing pipelines over a time window ([#350](https://github.com/alrayyes/pipeline-analytics/issues/350)) ([5fdc583](https://github.com/alrayyes/pipeline-analytics/commit/5fdc58323d5b131778bbd33644d994e70c0fdcec)), closes [#335](https://github.com/alrayyes/pipeline-analytics/issues/335)

## [0.52.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.51.0...v0.52.0) (2026-10-02)


### Features

* **web:** map run and step status to tones and stage progress ([#351](https://github.com/alrayyes/pipeline-analytics/issues/351)) ([3aa9e83](https://github.com/alrayyes/pipeline-analytics/commit/3aa9e8304ebd7cd137effd559c05ae11d4fb991d)), closes [#335](https://github.com/alrayyes/pipeline-analytics/issues/335)

## [0.51.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.50.0...v0.51.0) (2026-10-02)


### Features

* **ingestion:** record run commit metadata during reconciliation ([#341](https://github.com/alrayyes/pipeline-analytics/issues/341)) ([ab7b70b](https://github.com/alrayyes/pipeline-analytics/commit/ab7b70b3c5f2eb0ea6b59b48058015bc56474739)), closes [#335](https://github.com/alrayyes/pipeline-analytics/issues/335)

## [0.50.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.49.0...v0.50.0) (2026-10-02)


### Features

* add a configurable log level ([#94](https://github.com/alrayyes/pipeline-analytics/issues/94)) ([0e088d3](https://github.com/alrayyes/pipeline-analytics/commit/0e088d385aed894f97aa929553583a1ebe07cca6)), closes [#75](https://github.com/alrayyes/pipeline-analytics/issues/75)
* add a dark mode toggle ([#61](https://github.com/alrayyes/pipeline-analytics/issues/61)) ([9c4942a](https://github.com/alrayyes/pipeline-analytics/commit/9c4942ad7782631ea7a509ce30dc31ecfc8320d6))
* **api:** author OpenAPI spec for v1 endpoints ([#12](https://github.com/alrayyes/pipeline-analytics/issues/12)) ([81b336e](https://github.com/alrayyes/pipeline-analytics/commit/81b336e3315e196235f314843f784ef489d158f8))
* **api:** spec the flaky-run drill-down endpoints ([#217](https://github.com/alrayyes/pipeline-analytics/issues/217)) ([feb11bb](https://github.com/alrayyes/pipeline-analytics/commit/feb11bb3750b6d9f735178c62ac4033fb1cc13ab))
* **api:** specify failure insights and run list endpoints ([#337](https://github.com/alrayyes/pipeline-analytics/issues/337)) ([ecd33e7](https://github.com/alrayyes/pipeline-analytics/commit/ecd33e7cd9e54ec6f959203d0f3dc67da92e3161))
* **auth:** add revocable API token auth alongside sessions ([#186](https://github.com/alrayyes/pipeline-analytics/issues/186)) ([6a03692](https://github.com/alrayyes/pipeline-analytics/commit/6a036926ba3d7a6d1b653bcd56537fc509b103c1))
* **auth:** support registering more than one passkey ([#213](https://github.com/alrayyes/pipeline-analytics/issues/213)) ([8a6c072](https://github.com/alrayyes/pipeline-analytics/commit/8a6c072f3ea2c7f8c1f8dcf8b66563ba1e76ba43))
* **auth:** WebAuthn login/registration and session gating ([#20](https://github.com/alrayyes/pipeline-analytics/issues/20)) ([b78a48e](https://github.com/alrayyes/pipeline-analytics/commit/b78a48e942981fd915a16e6334cf6bec21c4547a))
* **cmd:** add a healthcheck subcommand for the container's own HEALTHCHECK ([#317](https://github.com/alrayyes/pipeline-analytics/issues/317)) ([7cfc049](https://github.com/alrayyes/pipeline-analytics/commit/7cfc0497c7eb09bfe9b5dcff62f192a4eb3e906b)), closes [#316](https://github.com/alrayyes/pipeline-analytics/issues/316)
* **dashboard-ui:** add a cross-pipeline overview of unhealthy steps ([#159](https://github.com/alrayyes/pipeline-analytics/issues/159)) ([38f0aac](https://github.com/alrayyes/pipeline-analytics/commit/38f0aac444cd490b734e1fb4e149cd7d14a1bf28)), closes [#150](https://github.com/alrayyes/pipeline-analytics/issues/150)
* **dashboard-ui:** add a GitHub API rate-limit insights page ([#161](https://github.com/alrayyes/pipeline-analytics/issues/161)) ([5d55751](https://github.com/alrayyes/pipeline-analytics/commit/5d5575162f058e1808092b66e50517cea3f56e1b)), closes [#160](https://github.com/alrayyes/pipeline-analytics/issues/160)
* **dashboard-ui:** add a persistent all/GitHub/Forgejo filter, group both list pages by forge ([#143](https://github.com/alrayyes/pipeline-analytics/issues/143)) ([92beeef](https://github.com/alrayyes/pipeline-analytics/commit/92beeefd97683641f5bfe95310dc84be65f6c0a6)), closes [#140](https://github.com/alrayyes/pipeline-analytics/issues/140)
* **dashboard-ui:** add a privacy & disclaimer page, linked from the footer ([#87](https://github.com/alrayyes/pipeline-analytics/issues/87)) ([7c520cb](https://github.com/alrayyes/pipeline-analytics/commit/7c520cba8f41e5522fa45f2441a717fcbf90a143)), closes [#78](https://github.com/alrayyes/pipeline-analytics/issues/78)
* **dashboard-ui:** color the performance-history charts, wire up their tooltip ([#111](https://github.com/alrayyes/pipeline-analytics/issues/111)) ([e5e09f7](https://github.com/alrayyes/pipeline-analytics/commit/e5e09f772e8987d8d59d999fa0ea75877110c883)), closes [#110](https://github.com/alrayyes/pipeline-analytics/issues/110)
* **dashboard-ui:** color-code health status and default to unhealthy-only ([#104](https://github.com/alrayyes/pipeline-analytics/issues/104)) ([08a8295](https://github.com/alrayyes/pipeline-analytics/commit/08a8295cdef30251b9c9bbeeeb34e71e662888a6)), closes [#101](https://github.com/alrayyes/pipeline-analytics/issues/101)
* **dashboard-ui:** connect a token once, then multi-select repos to follow ([#108](https://github.com/alrayyes/pipeline-analytics/issues/108)) ([ed1c78c](https://github.com/alrayyes/pipeline-analytics/commit/ed1c78ca3d6665f4a5f245c1b5dd6a02ea08ee68)), closes [#103](https://github.com/alrayyes/pipeline-analytics/issues/103)
* **dashboard-ui:** filter and sort the Pipelines list by repo, health, and last run ([#158](https://github.com/alrayyes/pipeline-analytics/issues/158)) ([ec70ca3](https://github.com/alrayyes/pipeline-analytics/commit/ec70ca38c372813585d1197b38dd477080b4bc2d))
* **dashboard-ui:** group the Pipelines list by repository ([#106](https://github.com/alrayyes/pipeline-analytics/issues/106)) ([63ebd51](https://github.com/alrayyes/pipeline-analytics/commit/63ebd51a4f436bdc66fc5979eb9570ddaa42edf6)), closes [#102](https://github.com/alrayyes/pipeline-analytics/issues/102)
* **dashboard-ui:** make the dashboard installable as a PWA ([#99](https://github.com/alrayyes/pipeline-analytics/issues/99)) ([def3669](https://github.com/alrayyes/pipeline-analytics/commit/def3669f6e1437c86649aff6be67dce928377d49)), closes [#77](https://github.com/alrayyes/pipeline-analytics/issues/77)
* **dashboard-ui:** offer a saved forge token when registering another repo ([#83](https://github.com/alrayyes/pipeline-analytics/issues/83)) ([8b14eba](https://github.com/alrayyes/pipeline-analytics/commit/8b14eba1b36041bf28f4dc716c2f2f09adebf5d4)), closes [#72](https://github.com/alrayyes/pipeline-analytics/issues/72)
* **dashboard-ui:** one-click saved token, exclude already-tracked repos ([#118](https://github.com/alrayyes/pipeline-analytics/issues/118)) ([739f6f6](https://github.com/alrayyes/pipeline-analytics/commit/739f6f61898d919fee9abd1a529aaa06ac11d0bc))
* **dashboard-ui:** pipeline detail view with trend charts ([#46](https://github.com/alrayyes/pipeline-analytics/issues/46)) ([12f8ba9](https://github.com/alrayyes/pipeline-analytics/commit/12f8ba956b6ad4fd476a524724734855e3a7715f))
* **dashboard-ui:** pipeline overview page ([#29](https://github.com/alrayyes/pipeline-analytics/issues/29)) ([cb05aa4](https://github.com/alrayyes/pipeline-analytics/commit/cb05aa4f59772b666d73a4c4bad5e17076421009))
* **dashboard-ui:** render and style the release history page ([#67](https://github.com/alrayyes/pipeline-analytics/issues/67)) ([fed045c](https://github.com/alrayyes/pipeline-analytics/commit/fed045c8a8064de81149bb5583c5749e83d0a274))
* **dashboard-ui:** replace the placeholder favicon with a real logo ([#85](https://github.com/alrayyes/pipeline-analytics/issues/85)) ([2928e41](https://github.com/alrayyes/pipeline-analytics/commit/2928e4101329c15ef571a081848b52f23bbe6c84)), closes [#76](https://github.com/alrayyes/pipeline-analytics/issues/76)
* **dashboard-ui:** repo picker and per-forge token guidance ([#64](https://github.com/alrayyes/pipeline-analytics/issues/64)) ([437e4b0](https://github.com/alrayyes/pipeline-analytics/commit/437e4b0b7500f07fc7d94aa7216a6a7901f60ec5))
* **dashboard-ui:** repo usage view ([#50](https://github.com/alrayyes/pipeline-analytics/issues/50)) ([360797a](https://github.com/alrayyes/pipeline-analytics/commit/360797a77a4450ce29de7c2d30b543fb43347d57))
* **dashboard-ui:** skip the token step when one's already known ([#264](https://github.com/alrayyes/pipeline-analytics/issues/264)) ([9a0f278](https://github.com/alrayyes/pipeline-analytics/commit/9a0f2781895e2f7b35528020513860312b6d4df3)), closes [#262](https://github.com/alrayyes/pipeline-analytics/issues/262)
* **dashboard-ui:** step breakdown with forge deep links ([#48](https://github.com/alrayyes/pipeline-analytics/issues/48)) ([f259323](https://github.com/alrayyes/pipeline-analytics/commit/f2593231853a3a751690325f09fa6a980b491d12))
* **dashboard-ui:** WebAuthn login/registration flow ([#28](https://github.com/alrayyes/pipeline-analytics/issues/28)) ([cec2ede](https://github.com/alrayyes/pipeline-analytics/commit/cec2edec668ef3ef4675daeed251933faea05bd8))
* **http:** log every request, not almost nothing ([#200](https://github.com/alrayyes/pipeline-analytics/issues/200)) ([f7df62b](https://github.com/alrayyes/pipeline-analytics/commit/f7df62bbe2e98a76933f133d80dc89ea5d498174)), closes [#196](https://github.com/alrayyes/pipeline-analytics/issues/196)
* **ingestion:** POST /api/repos registration with real webhook creation ([#15](https://github.com/alrayyes/pipeline-analytics/issues/15)) ([d3c2dcd](https://github.com/alrayyes/pipeline-analytics/commit/d3c2dcdf04987993825ef05a2f1c30bdd6516eef))
* **ingestion:** reconciliation polling and scheduler wiring ([#24](https://github.com/alrayyes/pipeline-analytics/issues/24)) ([f59494a](https://github.com/alrayyes/pipeline-analytics/commit/f59494a7ff0305ec4ad63761d2bf2c5bb9ed59e2))
* **ingestion:** record run branch, SHA, message and actor from webhooks ([#340](https://github.com/alrayyes/pipeline-analytics/issues/340)) ([5a7685b](https://github.com/alrayyes/pipeline-analytics/commit/5a7685b585669dfbfea0043c96b27fe6444bbefc)), closes [#335](https://github.com/alrayyes/pipeline-analytics/issues/335)
* **ingestion:** schema and encrypted credential storage ([#14](https://github.com/alrayyes/pipeline-analytics/issues/14)) ([7f11037](https://github.com/alrayyes/pipeline-analytics/commit/7f110374026d35e2a628aaf3083caa6e802d4bdd))
* **ingestion:** webhook receivers and run/job/step storage ([#22](https://github.com/alrayyes/pipeline-analytics/issues/22)) ([72ad44f](https://github.com/alrayyes/pipeline-analytics/commit/72ad44fda282b93abcc19550b55783ee02586353))
* **mcp:** add WebMCP support to the dashboard ([#310](https://github.com/alrayyes/pipeline-analytics/issues/310)) ([f55c96d](https://github.com/alrayyes/pipeline-analytics/commit/f55c96d5d2d3858b1c916c5f70e1841638ef9515))
* **mcp:** serve pipeline metrics as MCP tools over /api/mcp ([#304](https://github.com/alrayyes/pipeline-analytics/issues/304)) ([b5611b6](https://github.com/alrayyes/pipeline-analytics/commit/b5611b6f7dfcd1e8b911a32572336084a62eec84))
* **metrics:** categorise a failed step from its name and conclusion ([#347](https://github.com/alrayyes/pipeline-analytics/issues/347)) ([00502e0](https://github.com/alrayyes/pipeline-analytics/commit/00502e0a12906e2cdb41da57d6d38b8545302809)), closes [#335](https://github.com/alrayyes/pipeline-analytics/issues/335)
* **metrics:** implement the flaky-run drill-down endpoints ([#219](https://github.com/alrayyes/pipeline-analytics/issues/219)) ([04c13b4](https://github.com/alrayyes/pipeline-analytics/commit/04c13b4c02629d3b04acbd98b613ac69830eff37))
* **metrics:** pipeline health, trend, ranking, and usage computation ([#25](https://github.com/alrayyes/pipeline-analytics/issues/25)) ([806c9a0](https://github.com/alrayyes/pipeline-analytics/commit/806c9a0fae6706a0933f1d29a564e13f1c465eef))
* **metrics:** wire pipeline and usage API handlers ([#27](https://github.com/alrayyes/pipeline-analytics/issues/27)) ([f128e98](https://github.com/alrayyes/pipeline-analytics/commit/f128e98e4dd278422cb75271adfb3d260876a603))
* **openapi:** document pagination on GET /api/pipelines ([#246](https://github.com/alrayyes/pipeline-analytics/issues/246)) ([077f3c6](https://github.com/alrayyes/pipeline-analytics/commit/077f3c643221e880e406332aeb09cd3a203adb7d))
* **openapi:** document pagination on GET /api/repos ([#165](https://github.com/alrayyes/pipeline-analytics/issues/165)) ([75d7743](https://github.com/alrayyes/pipeline-analytics/commit/75d774312b266e106bb8dbdd263cdf0fea40cc79))
* **openapi:** document pagination on GET /api/steps/unhealthy ([#259](https://github.com/alrayyes/pipeline-analytics/issues/259)) ([7c0a279](https://github.com/alrayyes/pipeline-analytics/commit/7c0a2794c53d84dc91be88ddbda04f08c11903d7))
* **pipelines:** paginate the Pipelines page ([#250](https://github.com/alrayyes/pipeline-analytics/issues/250)) ([7c3ddbe](https://github.com/alrayyes/pipeline-analytics/commit/7c3ddbe5dd4ad69d8d51f32f5fb50a95c25ce88a))
* release automation (release-please + goreleaser) and Docker image ([#17](https://github.com/alrayyes/pipeline-analytics/issues/17)) ([577da7c](https://github.com/alrayyes/pipeline-analytics/commit/577da7c89e85e8890e947fc3fced79a84eff46cb))
* repo management UI (register, list, untrack) ([#57](https://github.com/alrayyes/pipeline-analytics/issues/57)) ([1d172e2](https://github.com/alrayyes/pipeline-analytics/commit/1d172e292e8e07b42cba1e0caa97c693199e900b))
* **repos:** exclude forks, mirrors, and archived repos from the picker ([#125](https://github.com/alrayyes/pipeline-analytics/issues/125)) ([62c239b](https://github.com/alrayyes/pipeline-analytics/commit/62c239bbbd875b3e8d53bf0d0850c411336aebe6))
* **settings:** persist theme and filter settings per account ([#208](https://github.com/alrayyes/pipeline-analytics/issues/208)) ([e1370c6](https://github.com/alrayyes/pipeline-analytics/commit/e1370c65207ac604a94ba1b72e0dc43863b32e88))
* **steps:** paginate the Steps page ([#261](https://github.com/alrayyes/pipeline-analytics/issues/261)) ([6c11b7c](https://github.com/alrayyes/pipeline-analytics/commit/6c11b7c19f542b2c788a616d265d8b04a1663871))
* sveltekit frontend, embedded, with a version footer ([#16](https://github.com/alrayyes/pipeline-analytics/issues/16)) ([c0895bf](https://github.com/alrayyes/pipeline-analytics/commit/c0895bffadffff26e112346f91b6af7cce3bf710))
* **ui:** wire the app's own brand teal into the theme, not shadcn's placeholder ([#255](https://github.com/alrayyes/pipeline-analytics/issues/255)) ([8d2d3e3](https://github.com/alrayyes/pipeline-analytics/commit/8d2d3e39481175b10a1349ba6e091946db4e860c)), closes [#254](https://github.com/alrayyes/pipeline-analytics/issues/254)
* **web:** add humans.txt ([6579b70](https://github.com/alrayyes/pipeline-analytics/commit/6579b708f341da067648c84aa79bed7bd4e97edb))
* **web:** add humans.txt ([604c13f](https://github.com/alrayyes/pipeline-analytics/commit/604c13f4dd553166585f939c86e7694ad42f7a6f)), closes [#175](https://github.com/alrayyes/pipeline-analytics/issues/175)
* **web:** audit /login with Lighthouse CI, fix what it finds ([#320](https://github.com/alrayyes/pipeline-analytics/issues/320)) ([afaecc2](https://github.com/alrayyes/pipeline-analytics/commit/afaecc2704f6273d8f5ed42f49cb542963c99f8d)), closes [#319](https://github.com/alrayyes/pipeline-analytics/issues/319)
* **web:** drill down a flaky step through its failed runs ([#220](https://github.com/alrayyes/pipeline-analytics/issues/220)) ([14c19ae](https://github.com/alrayyes/pipeline-analytics/commit/14c19ae9e126e2f172ce200078b40ef6423acdb7))
* **web:** lint Tailwind class usage with @shadcn/lint via Oxlint ([#287](https://github.com/alrayyes/pipeline-analytics/issues/287)) ([234ddfe](https://github.com/alrayyes/pipeline-analytics/commit/234ddfe06ecad9ba47d26deeb974ed77fa01d4ea))


### Bug Fixes

* bump dompurify from 3.4.15 to 3.4.16 in /web ([#327](https://github.com/alrayyes/pipeline-analytics/issues/327)) ([66e812e](https://github.com/alrayyes/pipeline-analytics/commit/66e812eb0037a08273098165bb68899d061deacb))
* bump github.com/go-webauthn/webauthn from 0.18.1 to 0.18.2 ([#302](https://github.com/alrayyes/pipeline-analytics/issues/302)) ([bd96274](https://github.com/alrayyes/pipeline-analytics/commit/bd9627475f232a31977a727fba73ba70d0ff5fe8))
* bump marked from 18.0.13 to 18.0.14 in /web ([#303](https://github.com/alrayyes/pipeline-analytics/issues/303)) ([1763b5d](https://github.com/alrayyes/pipeline-analytics/commit/1763b5d636b058281afcfc756dead44bf6f110ea))
* **ci:** bump golangci-lint pin past go1.27 skew ([#11](https://github.com/alrayyes/pipeline-analytics/issues/11)) ([0f3400e](https://github.com/alrayyes/pipeline-analytics/commit/0f3400ed0044311850aa740ab07a12cb0ec78e44))
* **ci:** deploy the API docs page under /docs/api/, not the Pages site root ([#323](https://github.com/alrayyes/pipeline-analytics/issues/323)) ([03edca0](https://github.com/alrayyes/pipeline-analytics/commit/03edca065c1b532390b8860963a02d29de487c32)), closes [#322](https://github.com/alrayyes/pipeline-analytics/issues/322)
* **dashboard-ui:** align page containers to max-w-4xl ([#79](https://github.com/alrayyes/pipeline-analytics/issues/79)) ([8056227](https://github.com/alrayyes/pipeline-analytics/commit/8056227d6402c480588d187f0289e63272ff38cd)), closes [#69](https://github.com/alrayyes/pipeline-analytics/issues/69)
* **dashboard-ui:** click through the unhealthy-only default before capturing screenshots ([#137](https://github.com/alrayyes/pipeline-analytics/issues/137)) ([3c4dfae](https://github.com/alrayyes/pipeline-analytics/commit/3c4dfae832f6652dd8153e4a21532a67975596ae)), closes [#136](https://github.com/alrayyes/pipeline-analytics/issues/136)
* **dashboard-ui:** fix horizontal overflow on the Pipelines list page ([#139](https://github.com/alrayyes/pipeline-analytics/issues/139)) ([2c778c1](https://github.com/alrayyes/pipeline-analytics/commit/2c778c1e6a3477631baa88718775231efb52952e)), closes [#135](https://github.com/alrayyes/pipeline-analytics/issues/135)
* **dashboard-ui:** fix horizontal page overflow on the pipeline detail and nav ([#115](https://github.com/alrayyes/pipeline-analytics/issues/115)) ([8b8d3ce](https://github.com/alrayyes/pipeline-analytics/commit/8b8d3ce6657eda075bc0acafb11de69318f3b6ff))
* **dashboard-ui:** reflect repo-registration status in nav and empty states ([#89](https://github.com/alrayyes/pipeline-analytics/issues/89)) ([20ff445](https://github.com/alrayyes/pipeline-analytics/commit/20ff4455f3151b26e7fc65364e67aa57fc5ceef9)), closes [#71](https://github.com/alrayyes/pipeline-analytics/issues/71)
* **dashboard-ui:** truncate long step names instead of widening the Steps table ([#156](https://github.com/alrayyes/pipeline-analytics/issues/156)) ([91ef4e6](https://github.com/alrayyes/pipeline-analytics/commit/91ef4e63cbb02eebb1c1913747fe643bc2009c89)), closes [#148](https://github.com/alrayyes/pipeline-analytics/issues/148)
* **db:** dedupe pre-existing repos before migration 5's unique index ([#248](https://github.com/alrayyes/pipeline-analytics/issues/248)) ([c854f35](https://github.com/alrayyes/pipeline-analytics/commit/c854f35de7db050b55962b04e04cdce3339eccfb)), closes [#247](https://github.com/alrayyes/pipeline-analytics/issues/247)
* **docker:** bump pinned ca-certificates to 20260909-r0 ([#167](https://github.com/alrayyes/pipeline-analytics/issues/167)) ([5b61c8a](https://github.com/alrayyes/pipeline-analytics/commit/5b61c8a204d8d946d16bf67015110b77d89c2c2c))
* **docs:** dodge Vale's spell-check on a sentence-initial 'Browsable' ([#283](https://github.com/alrayyes/pipeline-analytics/issues/283)) ([3eba558](https://github.com/alrayyes/pipeline-analytics/commit/3eba5581382f34c910f1a469b26c897412b6d7a2))
* **docs:** scope the screenshot capture script's health-filter click ([#183](https://github.com/alrayyes/pipeline-analytics/issues/183)) ([74ca888](https://github.com/alrayyes/pipeline-analytics/commit/74ca888306b46e3c70cb8d17b71ddc3e2afc2fd0)), closes [#182](https://github.com/alrayyes/pipeline-analytics/issues/182)
* **docs:** serve the spec alongside the Stoplight Elements page ([#281](https://github.com/alrayyes/pipeline-analytics/issues/281)) ([af3e5b5](https://github.com/alrayyes/pipeline-analytics/commit/af3e5b52813c6588d9b7204906ecec806f51b6bd))
* **e2e:** give every test its own isolated server instead of one shared ([2fc3322](https://github.com/alrayyes/pipeline-analytics/commit/2fc3322d2e464844a95bcd2fff0cc2bfd6f55954))
* **e2e:** give every test its own isolated server instead of one shared ([debb96f](https://github.com/alrayyes/pipeline-analytics/commit/debb96f58a2e31ccbf3b813885da5e2e1e397765)), closes [#168](https://github.com/alrayyes/pipeline-analytics/issues/168)
* embed the real git tag in version, not goreleaser's bare semver ([#97](https://github.com/alrayyes/pipeline-analytics/issues/97)) ([30eefc7](https://github.com/alrayyes/pipeline-analytics/commit/30eefc7124989b57d0f0ba41a1c3cb2d8bcda2a3)), closes [#95](https://github.com/alrayyes/pipeline-analytics/issues/95)
* **forgejo:** make webhook events and reconciliation match a real instance ([#124](https://github.com/alrayyes/pipeline-analytics/issues/124)) ([449e2b3](https://github.com/alrayyes/pipeline-analytics/commit/449e2b3b819c4a91bd5b7cc6f4ac78e09e8d1790))
* **forgejo:** tolerate the jobs-listing response shape mismatch ([#131](https://github.com/alrayyes/pipeline-analytics/issues/131)) ([653b422](https://github.com/alrayyes/pipeline-analytics/commit/653b4228807c392b05ba6401c966fdaa043bc749))
* **hooks:** gate vale in pre-push, matching CI's style job ([#291](https://github.com/alrayyes/pipeline-analytics/issues/291)) ([96b176e](https://github.com/alrayyes/pipeline-analytics/commit/96b176e084f47c6dede8867f4b1ac18e48f60aeb))
* hyphenated CLI flags weren't reachable via env vars ([#33](https://github.com/alrayyes/pipeline-analytics/issues/33)) ([c3dbe61](https://github.com/alrayyes/pipeline-analytics/commit/c3dbe61dac391a9c103fc9d61296669b5864dd75))
* **ingestion:** give github/forgejo clients their own http.Client ([#306](https://github.com/alrayyes/pipeline-analytics/issues/306)) ([97a31b4](https://github.com/alrayyes/pipeline-analytics/commit/97a31b4dd21433b34e5220b7cf370152f289a43b)), closes [#305](https://github.com/alrayyes/pipeline-analytics/issues/305)
* **ingestion:** paginate repo discovery, exclude archived/forked repos ([#92](https://github.com/alrayyes/pipeline-analytics/issues/92)) ([c12db45](https://github.com/alrayyes/pipeline-analytics/commit/c12db45b0fa47632abfc5ab4cbba3adb1be596e7)), closes [#74](https://github.com/alrayyes/pipeline-analytics/issues/74)
* **nav:** collapse the header behind a menu toggle on phone widths ([115d31a](https://github.com/alrayyes/pipeline-analytics/commit/115d31ae741f885e746ea963316fc6b00a3f6d8a))
* **nav:** collapse the header behind a menu toggle on phone widths ([2e7b290](https://github.com/alrayyes/pipeline-analytics/commit/2e7b290e2da0495ba9875f30cb274fd46382d28d)), closes [#171](https://github.com/alrayyes/pipeline-analytics/issues/171)
* pin bun below 1.4 and downgrade lockfiles to v1 ([#31](https://github.com/alrayyes/pipeline-analytics/issues/31)) ([2f7de37](https://github.com/alrayyes/pipeline-analytics/commit/2f7de37b46d0aaf1bee16df73897d6a6923a00ed))
* release job's bun needs to read both lockfile formats ([#42](https://github.com/alrayyes/pipeline-analytics/issues/42)) ([f59145a](https://github.com/alrayyes/pipeline-analytics/commit/f59145a9c36c4a77dba7e0cd5691496adcaf4187))
* **release:** deduplicate changelog and release notes ([#180](https://github.com/alrayyes/pipeline-analytics/issues/180)) ([adadd1b](https://github.com/alrayyes/pipeline-analytics/commit/adadd1bf1957147d5d0fc8c03f1070eb35f8c525)), closes [#179](https://github.com/alrayyes/pipeline-analytics/issues/179)
* **repos:** reject archived, forked, and mirror repos at registration ([#201](https://github.com/alrayyes/pipeline-analytics/issues/201)) ([81d6031](https://github.com/alrayyes/pipeline-analytics/commit/81d603192a41969d5fd46cf7ba2d680e4b380887))
* **repos:** reject duplicate repo registration clearly, not opaquely ([#199](https://github.com/alrayyes/pipeline-analytics/issues/199)) ([2aa6272](https://github.com/alrayyes/pipeline-analytics/commit/2aa62722dcf25ab29bc3be655005c2c65e2f8c5d)), closes [#197](https://github.com/alrayyes/pipeline-analytics/issues/197)
* **scripts:** match query strings and the real envelope in the pipelines mock ([#289](https://github.com/alrayyes/pipeline-analytics/issues/289)) ([76f3c57](https://github.com/alrayyes/pipeline-analytics/commit/76f3c570bcc1159c3ae1bd5279fcca58b835bfc1))
* **web:** add @types/node for the e2e helpers' node: imports ([7890e6a](https://github.com/alrayyes/pipeline-analytics/commit/7890e6ac18971c155de141bb94f49a56440ef418))
* **web:** bound root layout's load() fetches so a stalled backend can't hang the page forever ([#268](https://github.com/alrayyes/pipeline-analytics/issues/268)) ([63bf8ad](https://github.com/alrayyes/pipeline-analytics/commit/63bf8ad8c6c3ae51469bb7b252122733a80835d5)), closes [#251](https://github.com/alrayyes/pipeline-analytics/issues/251)
* **web:** override devalue and basic-ftp to patched versions ([#333](https://github.com/alrayyes/pipeline-analytics/issues/333)) ([b3e6285](https://github.com/alrayyes/pipeline-analytics/commit/b3e628503c86018984f950784fe4a7b29dad16e8)), closes [#332](https://github.com/alrayyes/pipeline-analytics/issues/332)


### Performance Improvements

* **reconcile:** skip refetching jobs for unchanged completed runs ([#207](https://github.com/alrayyes/pipeline-analytics/issues/207)) ([88f4128](https://github.com/alrayyes/pipeline-analytics/commit/88f41282a6305179b0c266e5a1eb4299fd07af59)), closes [#206](https://github.com/alrayyes/pipeline-analytics/issues/206)

## [0.49.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.48.0...v0.49.0) (2026-10-02)


### Features

* **metrics:** categorise a failed step from its name and conclusion ([#347](https://github.com/alrayyes/pipeline-analytics/issues/347)) ([00502e0](https://github.com/alrayyes/pipeline-analytics/commit/00502e0a12906e2cdb41da57d6d38b8545302809)), closes [#335](https://github.com/alrayyes/pipeline-analytics/issues/335)

## [0.48.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.47.0...v0.48.0) (2026-10-02)


### Features

* **ingestion:** record run branch, SHA, message and actor from webhooks ([#340](https://github.com/alrayyes/pipeline-analytics/issues/340)) ([5a7685b](https://github.com/alrayyes/pipeline-analytics/commit/5a7685b585669dfbfea0043c96b27fe6444bbefc)), closes [#335](https://github.com/alrayyes/pipeline-analytics/issues/335)

## [0.47.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.46.2...v0.47.0) (2026-10-02)


### Features

* **api:** specify failure insights and run list endpoints ([#337](https://github.com/alrayyes/pipeline-analytics/issues/337)) ([ecd33e7](https://github.com/alrayyes/pipeline-analytics/commit/ecd33e7cd9e54ec6f959203d0f3dc67da92e3161))

## [0.46.2](https://github.com/alrayyes/pipeline-analytics/compare/v0.46.1...v0.46.2) (2026-10-02)


### Bug Fixes

* bump dompurify from 3.4.15 to 3.4.16 in /web ([#327](https://github.com/alrayyes/pipeline-analytics/issues/327)) ([66e812e](https://github.com/alrayyes/pipeline-analytics/commit/66e812eb0037a08273098165bb68899d061deacb))
* **web:** override devalue and basic-ftp to patched versions ([#333](https://github.com/alrayyes/pipeline-analytics/issues/333)) ([b3e6285](https://github.com/alrayyes/pipeline-analytics/commit/b3e628503c86018984f950784fe4a7b29dad16e8)), closes [#332](https://github.com/alrayyes/pipeline-analytics/issues/332)

## [0.46.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.46.0...v0.46.1) (2026-09-28)


### Bug Fixes

* **ci:** deploy the API docs page under /docs/api/, not the Pages site root ([#323](https://github.com/alrayyes/pipeline-analytics/issues/323)) ([03edca0](https://github.com/alrayyes/pipeline-analytics/commit/03edca065c1b532390b8860963a02d29de487c32)), closes [#322](https://github.com/alrayyes/pipeline-analytics/issues/322)

## [0.46.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.45.0...v0.46.0) (2026-09-26)


### Features

* **web:** audit /login with Lighthouse CI, fix what it finds ([#320](https://github.com/alrayyes/pipeline-analytics/issues/320)) ([afaecc2](https://github.com/alrayyes/pipeline-analytics/commit/afaecc2704f6273d8f5ed42f49cb542963c99f8d)), closes [#319](https://github.com/alrayyes/pipeline-analytics/issues/319)

## [0.45.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.44.0...v0.45.0) (2026-09-26)


### Features

* **cmd:** add a healthcheck subcommand for the container's own HEALTHCHECK ([#317](https://github.com/alrayyes/pipeline-analytics/issues/317)) ([7cfc049](https://github.com/alrayyes/pipeline-analytics/commit/7cfc0497c7eb09bfe9b5dcff62f192a4eb3e906b)), closes [#316](https://github.com/alrayyes/pipeline-analytics/issues/316)

## [0.44.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.43.0...v0.44.0) (2026-09-25)


### Features

* **mcp:** add WebMCP support to the dashboard ([#310](https://github.com/alrayyes/pipeline-analytics/issues/310)) ([f55c96d](https://github.com/alrayyes/pipeline-analytics/commit/f55c96d5d2d3858b1c916c5f70e1841638ef9515))

## [0.43.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.42.2...v0.43.0) (2026-09-25)


### Features

* **mcp:** serve pipeline metrics as MCP tools over /api/mcp ([#304](https://github.com/alrayyes/pipeline-analytics/issues/304)) ([b5611b6](https://github.com/alrayyes/pipeline-analytics/commit/b5611b6f7dfcd1e8b911a32572336084a62eec84))


### Bug Fixes

* bump github.com/go-webauthn/webauthn from 0.18.1 to 0.18.2 ([#302](https://github.com/alrayyes/pipeline-analytics/issues/302)) ([bd96274](https://github.com/alrayyes/pipeline-analytics/commit/bd9627475f232a31977a727fba73ba70d0ff5fe8))
* bump marked from 18.0.13 to 18.0.14 in /web ([#303](https://github.com/alrayyes/pipeline-analytics/issues/303)) ([1763b5d](https://github.com/alrayyes/pipeline-analytics/commit/1763b5d636b058281afcfc756dead44bf6f110ea))
* **ingestion:** give github/forgejo clients their own http.Client ([#306](https://github.com/alrayyes/pipeline-analytics/issues/306)) ([97a31b4](https://github.com/alrayyes/pipeline-analytics/commit/97a31b4dd21433b34e5220b7cf370152f289a43b)), closes [#305](https://github.com/alrayyes/pipeline-analytics/issues/305)

## [0.42.2](https://github.com/alrayyes/pipeline-analytics/compare/v0.42.1...v0.42.2) (2026-09-22)


### Bug Fixes

* **hooks:** gate vale in pre-push, matching CI's style job ([#291](https://github.com/alrayyes/pipeline-analytics/issues/291)) ([96b176e](https://github.com/alrayyes/pipeline-analytics/commit/96b176e084f47c6dede8867f4b1ac18e48f60aeb))

## [0.42.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.42.0...v0.42.1) (2026-09-22)


### Bug Fixes

* **scripts:** match query strings and the real envelope in the pipelines mock ([#289](https://github.com/alrayyes/pipeline-analytics/issues/289)) ([76f3c57](https://github.com/alrayyes/pipeline-analytics/commit/76f3c570bcc1159c3ae1bd5279fcca58b835bfc1))

## [0.42.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.41.3...v0.42.0) (2026-09-22)


### Features

* **web:** lint Tailwind class usage with @shadcn/lint via Oxlint ([#287](https://github.com/alrayyes/pipeline-analytics/issues/287)) ([234ddfe](https://github.com/alrayyes/pipeline-analytics/commit/234ddfe06ecad9ba47d26deeb974ed77fa01d4ea))

## [0.41.3](https://github.com/alrayyes/pipeline-analytics/compare/v0.41.2...v0.41.3) (2026-09-22)


### Bug Fixes

* **docs:** dodge Vale's spell-check on a sentence-initial 'Browsable' ([#283](https://github.com/alrayyes/pipeline-analytics/issues/283)) ([3eba558](https://github.com/alrayyes/pipeline-analytics/commit/3eba5581382f34c910f1a469b26c897412b6d7a2))

## [0.41.2](https://github.com/alrayyes/pipeline-analytics/compare/v0.41.1...v0.41.2) (2026-09-22)


### Bug Fixes

* **docs:** serve the spec alongside the Stoplight Elements page ([#281](https://github.com/alrayyes/pipeline-analytics/issues/281)) ([af3e5b5](https://github.com/alrayyes/pipeline-analytics/commit/af3e5b52813c6588d9b7204906ecec806f51b6bd))

## [0.41.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.41.0...v0.41.1) (2026-09-21)


### Bug Fixes

* **web:** bound root layout's load() fetches so a stalled backend can't hang the page forever ([#268](https://github.com/alrayyes/pipeline-analytics/issues/268)) ([63bf8ad](https://github.com/alrayyes/pipeline-analytics/commit/63bf8ad8c6c3ae51469bb7b252122733a80835d5)), closes [#251](https://github.com/alrayyes/pipeline-analytics/issues/251)

## [0.41.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.40.0...v0.41.0) (2026-09-21)


### Features

* **dashboard-ui:** skip the token step when one's already known ([#264](https://github.com/alrayyes/pipeline-analytics/issues/264)) ([9a0f278](https://github.com/alrayyes/pipeline-analytics/commit/9a0f2781895e2f7b35528020513860312b6d4df3)), closes [#262](https://github.com/alrayyes/pipeline-analytics/issues/262)

## [0.40.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.39.0...v0.40.0) (2026-09-21)


### Features

* **steps:** paginate the Steps page ([#261](https://github.com/alrayyes/pipeline-analytics/issues/261)) ([6c11b7c](https://github.com/alrayyes/pipeline-analytics/commit/6c11b7c19f542b2c788a616d265d8b04a1663871))

## [0.39.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.38.0...v0.39.0) (2026-09-21)


### Features

* **openapi:** document pagination on GET /api/steps/unhealthy ([#259](https://github.com/alrayyes/pipeline-analytics/issues/259)) ([7c0a279](https://github.com/alrayyes/pipeline-analytics/commit/7c0a2794c53d84dc91be88ddbda04f08c11903d7))

## [0.38.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.37.0...v0.38.0) (2026-09-20)


### Features

* **ui:** wire the app's own brand teal into the theme, not shadcn's placeholder ([#255](https://github.com/alrayyes/pipeline-analytics/issues/255)) ([8d2d3e3](https://github.com/alrayyes/pipeline-analytics/commit/8d2d3e39481175b10a1349ba6e091946db4e860c)), closes [#254](https://github.com/alrayyes/pipeline-analytics/issues/254)

## [0.37.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.36.0...v0.37.0) (2026-09-20)


### Features

* **pipelines:** paginate the Pipelines page ([#250](https://github.com/alrayyes/pipeline-analytics/issues/250)) ([7c3ddbe](https://github.com/alrayyes/pipeline-analytics/commit/7c3ddbe5dd4ad69d8d51f32f5fb50a95c25ce88a))

## [0.36.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.35.0...v0.36.0) (2026-09-20)


### Features

* **openapi:** document pagination on GET /api/pipelines ([#246](https://github.com/alrayyes/pipeline-analytics/issues/246)) ([077f3c6](https://github.com/alrayyes/pipeline-analytics/commit/077f3c643221e880e406332aeb09cd3a203adb7d))


### Bug Fixes

* **db:** dedupe pre-existing repos before migration 5's unique index ([#248](https://github.com/alrayyes/pipeline-analytics/issues/248)) ([c854f35](https://github.com/alrayyes/pipeline-analytics/commit/c854f35de7db050b55962b04e04cdce3339eccfb)), closes [#247](https://github.com/alrayyes/pipeline-analytics/issues/247)

## [0.35.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.34.0...v0.35.0) (2026-09-18)


### Features

* **metrics:** implement the flaky-run drill-down endpoints ([#219](https://github.com/alrayyes/pipeline-analytics/issues/219)) ([04c13b4](https://github.com/alrayyes/pipeline-analytics/commit/04c13b4c02629d3b04acbd98b613ac69830eff37))
* **web:** drill down a flaky step through its failed runs ([#220](https://github.com/alrayyes/pipeline-analytics/issues/220)) ([14c19ae](https://github.com/alrayyes/pipeline-analytics/commit/14c19ae9e126e2f172ce200078b40ef6423acdb7))

## [0.34.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.33.0...v0.34.0) (2026-09-18)


### Features

* **api:** spec the flaky-run drill-down endpoints ([#217](https://github.com/alrayyes/pipeline-analytics/issues/217)) ([feb11bb](https://github.com/alrayyes/pipeline-analytics/commit/feb11bb3750b6d9f735178c62ac4033fb1cc13ab))

## [0.33.0](https://github.com/alrayyes/pipeline-analytics/compare/v0.32.1...v0.33.0) (2026-09-18)


### Features

* **auth:** support registering more than one passkey ([#213](https://github.com/alrayyes/pipeline-analytics/issues/213)) ([8a6c072](https://github.com/alrayyes/pipeline-analytics/commit/8a6c072f3ea2c7f8c1f8dcf8b66563ba1e76ba43))

## [0.32.1](https://github.com/alrayyes/pipeline-analytics/compare/v0.32.0...v0.32.1) (2026-09-18)


### Performance Improvements

* **reconcile:** skip refetching jobs for unchanged completed runs ([#207](https://github.com/alrayyes/pipeline-analytics/issues/207)) ([88f4128](https://github.com/alrayyes/pipeline-analytics/commit/88f41282a6305179b0c266e5a1eb4299fd07af59)), closes [#206](https://github.com/alrayyes/pipeline-analytics/issues/206)

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
