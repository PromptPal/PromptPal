# Changelog

## [1.16.1](https://github.com/PromptPal/PromptPal/compare/v1.16.0...v1.16.1) (2025-07-21)


### Features

* add IP tracking to webhook calls ([fb9a5df](https://github.com/PromptPal/PromptPal/commit/fb9a5df17e6eebf533dbc77f4551f36c60582007))
* add provider defaultModel to webhook payload ([c636db7](https://github.com/PromptPal/PromptPal/commit/c636db718cccb110c7ea806413dae02136a06233))
* add test environment configuration file ([302a8dc](https://github.com/PromptPal/PromptPal/commit/302a8dc7c12f9b59b4fe205e0e5ce287d14a94e3))
* change authentication from username+password to email+password ([6448dd7](https://github.com/PromptPal/PromptPal/commit/6448dd714f972d23bba2b1c2d0e7f37b4c17dcee))
* change authentication from username+password to email+password ([c2afd74](https://github.com/PromptPal/PromptPal/commit/c2afd743ef6d480a57c84a9bb533aae711dd6853))
* implement comprehensive webhook call recording ([dcdb0fc](https://github.com/PromptPal/PromptPal/commit/dcdb0fc78b1cb34100c3062b89dc0c1a1373eade))
* implement createUser GraphQL mutation with admin authorization ([e0fca20](https://github.com/PromptPal/PromptPal/commit/e0fca20be47ff7c40f0ac305ec52abd56e687092))
* implement createUser GraphQL mutation with admin authorization ([f447a02](https://github.com/PromptPal/PromptPal/commit/f447a0212b0102cbdd28e2eb8729d0f1a12ba0d0))
* implement mocked RBACService for schema tests ([5ff4365](https://github.com/PromptPal/PromptPal/commit/5ff4365353ccc3feb1f85ec4dcad8f83f868fd48))
* implement RBAC (Role-Based Access Control) system ([ed10ae4](https://github.com/PromptPal/PromptPal/commit/ed10ae4d6b11b0198cd048ea4a1b8947da078c7f))
* implement RBAC (Role-Based Access Control) system ([2ace2e6](https://github.com/PromptPal/PromptPal/commit/2ace2e6a0a0a3daf2f5cbd3160ab0d3c72348182)), closes [#110](https://github.com/PromptPal/PromptPal/issues/110)
* implement RBAC permission checking in GraphQL schema resolvers ([cf66ab0](https://github.com/PromptPal/PromptPal/commit/cf66ab07ab7c937b3061cd0126de3a463ce7ac29))
* implement RBAC permission checking in GraphQL schema resolvers ([bfcdce9](https://github.com/PromptPal/PromptPal/commit/bfcdce9a259d1161cb723ab841a1fe6001d0906b))
* implement username/password authentication system ([570365f](https://github.com/PromptPal/PromptPal/commit/570365f5a5f1a3bb625ba50b73b2debb1220500c))
* implement username/password authentication system ([39d71a3](https://github.com/PromptPal/PromptPal/commit/39d71a3d404fac5df1fd0467caeb53fc576abe9e))
* implement webhook support for onPromptFinished events ([b40ce79](https://github.com/PromptPal/PromptPal/commit/b40ce79df57a36a196e396fc7c3e07ed08579584))
* implement webhook support for onPromptFinished events ([74c2aad](https://github.com/PromptPal/PromptPal/commit/74c2aad608b5d2b49b7f91d869d0b338fd041c17))
* replace hard-coded version with versionCommit in webhook User-Agent ([ac2332d](https://github.com/PromptPal/PromptPal/commit/ac2332dc5cbf067b0379b9c7175e4ad8c812a7e7))
* **schema:** add webhook call type definitions and import ([31fd872](https://github.com/PromptPal/PromptPal/commit/31fd87227cdc07750ff53c3fe53612c9574f0747))
* update GraphQL schema to use paginated responses for roles and permissions ([525d482](https://github.com/PromptPal/PromptPal/commit/525d48299c09fba7f699eba38e79c55640cd796e))
* **webhook:** add separate update payload type and fix resolver ([4e00d2e](https://github.com/PromptPal/PromptPal/commit/4e00d2edf5bba003f7e403864b54ee0899d09b47))
* **webhook:** add webhook resolver for fetching single webhook ([3eefe67](https://github.com/PromptPal/PromptPal/commit/3eefe679dbfd1600ccc23e79622fb887372e90fe))
* **webhook:** add webhook resolver for fetching single webhook ([420e8d8](https://github.com/PromptPal/PromptPal/commit/420e8d86882aaee3fd20f72d73313867a10500a9)), closes [#120](https://github.com/PromptPal/PromptPal/issues/120)


### Bug Fixes

* add missing RBACServiceInstance global variable ([3fbae73](https://github.com/PromptPal/PromptPal/commit/3fbae73632fc7005a1852e206e6bfc40457c2d50))
* add safe type assertion for GraphQL context in CreateUser ([9a1e5cd](https://github.com/PromptPal/PromptPal/commit/9a1e5cdcb8b7b2877234745095a6be15f2177e8e))
* address remaining RBAC implementation issues from code review ([fa06a6a](https://github.com/PromptPal/PromptPal/commit/fa06a6a0520b2b6f4a39116969f3d11c2f8f1089))
* address webhook implementation review issues ([8dad49c](https://github.com/PromptPal/PromptPal/commit/8dad49ca6620a85fa0056903b58ce3de351bd52f))
* address webhook implementation security and code quality issues ([cf0de95](https://github.com/PromptPal/PromptPal/commit/cf0de95f94d618fc96bd8a9bd1bf11250749366f))
* clean up auth implementation and move passwordAuth to mutation ([5c5a325](https://github.com/PromptPal/PromptPal/commit/5c5a32530cdf7713517b4fc93a3975b7a782d325))
* fix go mod ([5f15cd0](https://github.com/PromptPal/PromptPal/commit/5f15cd0a49eef8ac691047beba57d7b586b056e1))
* **project:** fix project testcase ([0aafd9c](https://github.com/PromptPal/PromptPal/commit/0aafd9c2eeee112ccb7098dba5ddf922dbdde871))
* **project:** fix testcase ([e67a2f9](https://github.com/PromptPal/PromptPal/commit/e67a2f950464a05b3aef91afb840a5708878132b))
* remove duplicate return statements in passwordAuthHandler ([18ca192](https://github.com/PromptPal/PromptPal/commit/18ca19277f1ee4d5c6e11c4175c30bc7f6e075d2))
* remove duplicate webhook.go file and update function call ([2a9e13b](https://github.com/PromptPal/PromptPal/commit/2a9e13b4fbd866e175193e03d93ad58d9ad3ab3c))
* remove SetDurationMs and SetIsSuccess calls from webhook service ([ae3e350](https://github.com/PromptPal/PromptPal/commit/ae3e350fcd039faf31309f44d4f9109a2398b8e2))
* remove unused imports from test files ([291156d](https://github.com/PromptPal/PromptPal/commit/291156d66ff74ff3aaa7db59465eb89cef622254))
* reorder cleanup operations in test teardown ([10b08e7](https://github.com/PromptPal/PromptPal/commit/10b08e7a8adf3da928b2b550822d9d3d2a8b0ed7))
* resolve webhook test compilation issues by adding service mocks ([9a0bfd6](https://github.com/PromptPal/PromptPal/commit/9a0bfd633eb739a58b2bcb1c0438dcf2e0ee9b9f))
* **schema:** update testcase ([ea5c678](https://github.com/PromptPal/PromptPal/commit/ea5c678b821b3eabbccf134d62a92e0eb9cfaa3c))
* **service:** ignore mock files ([0afda14](https://github.com/PromptPal/PromptPal/commit/0afda148b9d361c7e44f31cc80367a7441d5c1d8))
* update GraphQL schema to return paginated results for roles and permissions ([e0c65f0](https://github.com/PromptPal/PromptPal/commit/e0c65f09a0fa1cb18f4ccc780735e19a9ceab059))
* use request context parameter instead of creating custom contexts in webhook service ([08ea000](https://github.com/PromptPal/PromptPal/commit/08ea00093bbf80fa76019b033dbfa50cb4e886b8))
* **webhook:** correct unauthorized test to return false for PermProjectEdit ([e03a07f](https://github.com/PromptPal/PromptPal/commit/e03a07f63aad2a192b985898007b706fa65a9a64))
* **webhook:** fix RBAC permission mock expectations in tests ([aca1504](https://github.com/PromptPal/PromptPal/commit/aca1504174676349090792d95948243c3f8e1a8a))
* **webhook:** fix webhook test ([f7db50b](https://github.com/PromptPal/PromptPal/commit/f7db50b274d041b7ff9b4805def5ce05cea35423))
* **webhook:** fix webhook testcase ([c68e8ed](https://github.com/PromptPal/PromptPal/commit/c68e8eda165095bd1424104a445720e53a04ecd8))
* **webhook:** remove incorrect PermProjectEdit mock in unauthorized test ([e5ee5e7](https://github.com/PromptPal/PromptPal/commit/e5ee5e7edcb68d7613be9bad6a332712756be6ff))
* **webhook:** replace JSON type with String in GraphQL schema ([4136d9a](https://github.com/PromptPal/PromptPal/commit/4136d9a751648d0200ddb36a0f288bbf7cb5659c))


### Miscellaneous Chores

* release 1.16.1 ([08afca8](https://github.com/PromptPal/PromptPal/commit/08afca8d2eb3b5a841b5e0b13831bd1e4c6338fe))

## [1.16.0](https://github.com/PromptPal/PromptPal/compare/v1.15.11...v1.16.0) (2025-07-12)


### Features

* implement user API support with optional ID parameter ([b736791](https://github.com/PromptPal/PromptPal/commit/b736791f82458ed5fd92fdb4bd3a418b4f3127c0))
* implement user API support with optional ID parameter ([f77d07a](https://github.com/PromptPal/PromptPal/commit/f77d07a036000f05ad2531bc2b556f0880de2106)), closes [#97](https://github.com/PromptPal/PromptPal/issues/97)


### Bug Fixes

* add mockery configuration and update test files ([82b4ca1](https://github.com/PromptPal/PromptPal/commit/82b4ca1f05031c17e06855f5e6230c3c882fc133))
* add mockery configuration and update test files ([55157e9](https://github.com/PromptPal/PromptPal/commit/55157e9def1b9e34d845d7b988a743fed0d749e0))
* change SPA fallback status code from 404 to 200 ([b311eaa](https://github.com/PromptPal/PromptPal/commit/b311eaaad439370b133047d22f60790ae8f82f07))
* change SPA fallback status code from 404 to 200 ([d8bfaa0](https://github.com/PromptPal/PromptPal/commit/d8bfaa07d383789e9e367bda2295f99ff756b254)), closes [#92](https://github.com/PromptPal/PromptPal/issues/92)
* **ci:** add claude setting file to let it can use tools ([e180c4f](https://github.com/PromptPal/PromptPal/commit/e180c4f62eedf2afa8781693c3c0fbad907e5b96))
* **ci:** revert to ubuntu-latest runners for workflow stability ([8a0d9ea](https://github.com/PromptPal/PromptPal/commit/8a0d9ea56e7d5d477e4792e18787ef8c20f30c88))
* **ci:** run go generate before mockery in workflows ([d17aee9](https://github.com/PromptPal/PromptPal/commit/d17aee9bbf32f8273499425784012e9a6e221f28))
* **ci:** standardize CI configuration and remove debug logging ([43b9cd1](https://github.com/PromptPal/PromptPal/commit/43b9cd1ca6e8fd2004a04d32770c91ca4561831c))
* **ci:** update CI workflow to use mockery v3 ([f58d09e](https://github.com/PromptPal/PromptPal/commit/f58d09e2d686a9ba4c36f7e9d565d9d8e662ff4d))
* **claude:** fix claude permission allow list ([43c02a6](https://github.com/PromptPal/PromptPal/commit/43c02a6c2fc93ea8493351259bcd84e0574862c8))
* **test:** standardize Redis URL format in test environments ([addc779](https://github.com/PromptPal/PromptPal/commit/addc779af6cb413825dc33cc944177c4264b125e))

## [1.15.11](https://github.com/PromptPal/PromptPal/compare/v1.15.10...v1.15.11) (2025-07-10)


### Bug Fixes

* remove artificial delay in SSE streaming and cleanup example file ([9f13f20](https://github.com/PromptPal/PromptPal/commit/9f13f20e3e86e60ad2ea64cfda38ff3b67e55727))

## [1.15.10](https://github.com/PromptPal/PromptPal/compare/v1.15.9...v1.15.10) (2025-07-10)


### Bug Fixes

* add CLAUDE.md for codebase guidance ([1357198](https://github.com/PromptPal/PromptPal/commit/1357198e1232be8ce6b276ca5156f2d1a0c39cc7))
* **ci:** update ci settings ([251714c](https://github.com/PromptPal/PromptPal/commit/251714c2a70f1c16dd6c8824d4b64bc85f687fe4))
* improve SSE streaming and optimize middleware usage ([e1b8974](https://github.com/PromptPal/PromptPal/commit/e1b89743d13ca5de1b221d38f6ce3d95997565ba))
* update dependencies and AI service implementation ([545f7a5](https://github.com/PromptPal/PromptPal/commit/545f7a54d6e9e8a1354815d52196bdaad13818f0))

## [1.15.9](https://github.com/PromptPal/PromptPal/compare/v1.15.8...v1.15.9) (2025-04-26)


### Bug Fixes

* **api:** Add event ping to confirm connection in prompt stream endpoint ([aa6e29a](https://github.com/PromptPal/PromptPal/commit/aa6e29ae1cee66baae705b8602f098741a17bd97))
* **api:** add sse header ([355217e](https://github.com/PromptPal/PromptPal/commit/355217ece0223ba00288c535e593dfd85376b8d8))

## [1.15.8](https://github.com/PromptPal/PromptPal/compare/v1.15.7...v1.15.8) (2025-04-26)


### Bug Fixes

* **db:** update db for query relations ([d343058](https://github.com/PromptPal/PromptPal/commit/d3430586360172fc1b3fb2857f88005569e2cf86))

## [1.15.7](https://github.com/PromptPal/PromptPal/compare/v1.15.6...v1.15.7) (2025-04-17)


### Bug Fixes

* **ci:** Add DB cleanup step and upgrade Codecov action to v5 in CI workflows ([c0c85bb](https://github.com/PromptPal/PromptPal/commit/c0c85bb7bfb66761cbefdab3fbba75bbf4d68bde))

## [1.15.6](https://github.com/PromptPal/PromptPal/compare/v1.15.5...v1.15.6) (2025-04-15)


### Bug Fixes

* **go:** Remove go-generics-cache dependency and related modules ([f1b2688](https://github.com/PromptPal/PromptPal/commit/f1b26883cd8827fc29715faa12daa1dd88db73c3))

## [1.15.5](https://github.com/PromptPal/PromptPal/compare/v1.15.4...v1.15.5) (2025-04-15)


### Bug Fixes

* **ci:** Add Redis service to GitHub workflow configurations ([a464052](https://github.com/PromptPal/PromptPal/commit/a46405256e3eadae916ea11680677b795e8ca28a))
* **provider:** fix provider project query ([b537eee](https://github.com/PromptPal/PromptPal/commit/b537eee31cce7eac369f29f14617602e8cf4e3d1))
* **test:** Add Redis initialization in call test setup ([f435359](https://github.com/PromptPal/PromptPal/commit/f4353594f32ccc64a5c49679ed30c78f25f6048d))

## [1.15.4](https://github.com/PromptPal/PromptPal/compare/v1.15.3...v1.15.4) (2025-04-02)


### Bug Fixes

* **provider:** support custom headers in provider ([50e358a](https://github.com/PromptPal/PromptPal/commit/50e358a2028d57c0fea973f2648d3fa85b0ce874))

## [1.15.3](https://github.com/PromptPal/PromptPal/compare/v1.15.2...v1.15.3) (2025-04-01)


### Bug Fixes

* **cache:** add redis support ([ab7c417](https://github.com/PromptPal/PromptPal/commit/ab7c4177afb064f3958672771bcf28aab167c19f))

## [1.15.2](https://github.com/PromptPal/PromptPal/compare/v1.15.1...v1.15.2) (2025-03-31)


### Bug Fixes

* **ci:** update pg in ci ([d5dcb1f](https://github.com/PromptPal/PromptPal/commit/d5dcb1fbd732bfa3fa8aecf20a20f216ee5366ea))
* **provider:** fix testcase ([798381b](https://github.com/PromptPal/PromptPal/commit/798381b9899a91309bc48178bd4391a26d073f52))

## [1.15.1](https://github.com/PromptPal/PromptPal/compare/v1.15.0...v1.15.1) (2025-03-31)


### Bug Fixes

* **err:** fix syntax err ([74b74f3](https://github.com/PromptPal/PromptPal/commit/74b74f3f586b663b2b08cd9f15683ba2342bbcd9))
* **provider:** fix testcase ([c8240c1](https://github.com/PromptPal/PromptPal/commit/c8240c14636ab54f3f8eb33f27e6d979cef617c4))
* **provider:** fix testcase ([bbe5d2a](https://github.com/PromptPal/PromptPal/commit/bbe5d2a6695d1554fac7ad5660da8234cecbaa62))
* **provider:** fix testcase ([d170ea0](https://github.com/PromptPal/PromptPal/commit/d170ea003c3ad8b966613af4c306db5ac33ec85c))
* **provider:** fix testcase ([3b48e24](https://github.com/PromptPal/PromptPal/commit/3b48e24bb1aec69306ecc9663de2324c4f8bbd32))
* **provider:** fix testcase ([3e73f28](https://github.com/PromptPal/PromptPal/commit/3e73f28f5c71922548f989b8d8b5a3cc348c84aa))

## [1.15.0](https://github.com/PromptPal/PromptPal/compare/v1.14.10...v1.15.0) (2025-03-31)


### Features

* **prompt:** use isomorphic ai client for test and run ([9688eeb](https://github.com/PromptPal/PromptPal/commit/9688eebc774204ad934ca1c190e1a07eab2f965a))


### Bug Fixes

* add provider on project and prompt ([3373c22](https://github.com/PromptPal/PromptPal/commit/3373c225fb0333b9ea8f6260a945ccd1dbde7541))
* **isomorphic:** remove clients in main.go ([bec3023](https://github.com/PromptPal/PromptPal/commit/bec30232749d454c78be39c8fe12966dbc89927a))
* **isomorphic:** remove gemini and openai clients ([72ac385](https://github.com/PromptPal/PromptPal/commit/72ac3857ebc5b1b10bf5ea02528c0a143064f656))
* **project:** fix provider assignment in project ([e479e1b](https://github.com/PromptPal/PromptPal/commit/e479e1b082ae72810354c5a28cbe4dda81dede83))
* **provider:** assign provider if valid ([489c24c](https://github.com/PromptPal/PromptPal/commit/489c24c5468647ab981f1a0f52f3a8228a416dab))
* update deps ([328d30a](https://github.com/PromptPal/PromptPal/commit/328d30aa12c38a3f185ddf5c3ba146b768bb716b))

## [1.14.10](https://github.com/PromptPal/PromptPal/compare/v1.14.9...v1.14.10) (2025-02-10)


### Bug Fixes

* **api:** disable brotli ([d3fac0e](https://github.com/PromptPal/PromptPal/commit/d3fac0e30287c585485fb18de521024466893f0d))

## [1.14.9](https://github.com/PromptPal/PromptPal/compare/v1.14.8...v1.14.9) (2025-02-10)


### Bug Fixes

* **api:** enable brotli ([9026f20](https://github.com/PromptPal/PromptPal/commit/9026f2024aee8f985a51aebee6a75ca6977cea68))

## [1.14.8](https://github.com/PromptPal/PromptPal/compare/v1.14.7...v1.14.8) (2025-02-09)


### Bug Fixes

* **api:** add test for replacer ([16b5375](https://github.com/PromptPal/PromptPal/commit/16b5375039155b463dfbc86d8bd38feca7110db8))
* **openai:** fix variables replacer ([e2cc436](https://github.com/PromptPal/PromptPal/commit/e2cc436a584075201a5e52e6890d0b23edf8c1e5))

## [1.14.7](https://github.com/PromptPal/PromptPal/compare/v1.14.6...v1.14.7) (2025-02-09)


### Bug Fixes

* **openai:** remove unused comments ([7988b07](https://github.com/PromptPal/PromptPal/commit/7988b07da29c74a10f0115101b802d42f8958ba2))

## [1.14.6](https://github.com/PromptPal/PromptPal/compare/v1.14.5...v1.14.6) (2025-02-09)


### Bug Fixes

* **deps:** update deps ([c9510c5](https://github.com/PromptPal/PromptPal/commit/c9510c51f68dc1cd1970fce1ac67efcbf6e2dbda))
* **utils:** add test for variable replacer ([dd35f0e](https://github.com/PromptPal/PromptPal/commit/dd35f0e311f795a12f32eb50791cdac7a14a9c7c))

## [1.14.5](https://github.com/PromptPal/PromptPal/compare/v1.14.4...v1.14.5) (2025-02-09)


### Bug Fixes

* **api:** add debug log for openai ([513a5c7](https://github.com/PromptPal/PromptPal/commit/513a5c770674b127a6712d160231e199e1ada813))

## [1.14.4](https://github.com/PromptPal/PromptPal/compare/v1.14.3...v1.14.4) (2025-02-09)


### Bug Fixes

* **api:** fix placeholder replacer in variable ([2d76df9](https://github.com/PromptPal/PromptPal/commit/2d76df9edcd096cbc6ecdd04effe89b1dc8ec126))

## [1.14.3](https://github.com/PromptPal/PromptPal/compare/v1.14.2...v1.14.3) (2025-02-09)


### Bug Fixes

* **ai:** add context rules file ([720e4e5](https://github.com/PromptPal/PromptPal/commit/720e4e55abb8f9f9734bdd20e36fb87757696071))
* **ai:** remove unused /v1 path prefix and support deepseek ([4fd7f30](https://github.com/PromptPal/PromptPal/commit/4fd7f301f540fd59257c80f0c66b8e9025ff0f60))
* **ai:** update rules ([f8a1fcb](https://github.com/PromptPal/PromptPal/commit/f8a1fcb99dc2afcf43cd9235fe80ff57a166901a))

## [1.14.2](https://github.com/PromptPal/PromptPal/compare/v1.14.1...v1.14.2) (2025-02-02)


### Bug Fixes

* **o3:** add o3 support ([45c9b80](https://github.com/PromptPal/PromptPal/commit/45c9b802c8dd6188624bc2a56eda574178bd1cd2))

## [1.14.1](https://github.com/PromptPal/PromptPal/compare/v1.14.0...v1.14.1) (2024-11-20)


### Bug Fixes

* **deps:** upgrade deps ([5961ffb](https://github.com/PromptPal/PromptPal/commit/5961ffb9cfa4c75e6fe77fccee0380d0e77d8b3a))

## [1.14.0](https://github.com/PromptPal/PromptPal/compare/v1.13.1...v1.14.0) (2024-09-26)


### Features

* **calls:** add user_id search for calls ([49de70d](https://github.com/PromptPal/PromptPal/commit/49de70da0226428207ed5780ad2b07bf0f3b64c6))


### Bug Fixes

* **deps:** upgrade go to latest version ([8689baf](https://github.com/PromptPal/PromptPal/commit/8689baf9dd6d60fbef990af3dbdc250e29133ad7))
* **gemini:** add more models of gemini ([bd60759](https://github.com/PromptPal/PromptPal/commit/bd60759f1a2e6de2435fdd0fe86453bbb9308c76))
* **models:** support o1 ([0bffdf4](https://github.com/PromptPal/PromptPal/commit/0bffdf4f3fb5a8d3ddd0e47394b2d57969cf7deb))

## [1.13.1](https://github.com/PromptPal/PromptPal/compare/v1.13.0...v1.13.1) (2024-08-03)


### Bug Fixes

* **deps:** update deps ([#66](https://github.com/PromptPal/PromptPal/issues/66)) ([da80159](https://github.com/PromptPal/PromptPal/commit/da8015993355d6caedd46b115661604b2a3ef253))

## [1.13.0](https://github.com/PromptPal/PromptPal/compare/v1.12.0...v1.13.0) (2024-08-02)


### Features

* **ip:** add ip data for api ([4faf217](https://github.com/PromptPal/PromptPal/commit/4faf217c9a2991facf7d462348414397525a97d9))
* **ip:** record request ip in history ([b52daa4](https://github.com/PromptPal/PromptPal/commit/b52daa4d265e05a9b46239aceb4d18fb27eef11d))


### Bug Fixes

* **docs:** update readme ([19b81ac](https://github.com/PromptPal/PromptPal/commit/19b81ac29a42c509e34b08077ed3a4396dce9a67))
* **docs:** update readme doc ([09060ff](https://github.com/PromptPal/PromptPal/commit/09060ffb3fcc6902790a7c9ce545a35950a1705a))
* **stream:** fix usage recording on streaming openai request ([ba445fa](https://github.com/PromptPal/PromptPal/commit/ba445faed30c27eb32ceeb334f913e8bbaaa442e))

## [1.12.0](https://github.com/PromptPal/PromptPal/compare/v1.11.5...v1.12.0) (2024-08-01)


### Features

* **openToken:** add ability to validate openToken ([ceb1843](https://github.com/PromptPal/PromptPal/commit/ceb1843b77057399aadeede1fadcc60161d88260))
* **openToken:** add api to update openToken ([3653c2c](https://github.com/PromptPal/PromptPal/commit/3653c2c48d82b6ef0a33ae63df0b828db7d5700d))
* **openToken:** make openToken validation server configable ([7227ed8](https://github.com/PromptPal/PromptPal/commit/7227ed85ed24eb660b5d22e4e7abdc186a9572b0))


### Bug Fixes

* **calls:** add `cached` attr to calls ([d08a78d](https://github.com/PromptPal/PromptPal/commit/d08a78dfa331de2902bdbdb1585a7759e85ac271))
* **deps:** remove sqlite support ([0a4689a](https://github.com/PromptPal/PromptPal/commit/0a4689a4f61ef3b59b6c0c36bacc14abe2ba55db))
* **openToken:** add cache when update got finished ([69af7f5](https://github.com/PromptPal/PromptPal/commit/69af7f5ce685ca3e43094903ca51f4e33a33c90b))
* **openToken:** fix render order ([e26a09a](https://github.com/PromptPal/PromptPal/commit/e26a09a605bf3350e5fb9620fcafcc6fe76f4df8))
* **openToken:** fix type name ([ffde9c1](https://github.com/PromptPal/PromptPal/commit/ffde9c1ce5889cf75f9c7c9abe29d0c04c0c6b00))
* **openToken:** remove Bearer in temporary token header ([5af7347](https://github.com/PromptPal/PromptPal/commit/5af73470b18da01954c7aba3518c7701d8c3ed57))
* **openToken:** update openToken code to make it works ([cfbdf2b](https://github.com/PromptPal/PromptPal/commit/cfbdf2b5c39182b0b8cf574c408220e2f5c65d0c))

## [1.11.5](https://github.com/PromptPal/PromptPal/compare/v1.11.4...v1.11.5) (2024-07-19)


### Features

* **api:** add cache support ([e893eb2](https://github.com/PromptPal/PromptPal/commit/e893eb205c71480c563b8a6c5378c78eb3e867d3))
* **model:** add gpt-4o-mini support ([cb95aa9](https://github.com/PromptPal/PromptPal/commit/cb95aa9cfc4f541ccfffb99e6182f00a7fc55fe5))


### Bug Fixes

* **api:** abort api request if cache hit ([9206f89](https://github.com/PromptPal/PromptPal/commit/9206f8977059d5b6325a64e4227a148473f6a183))


### Miscellaneous Chores

* release 1.11.5 ([52a2b8c](https://github.com/PromptPal/PromptPal/commit/52a2b8c4c0f587d85e7d9c3dfc3fff6f45998db3))

## [1.11.4](https://github.com/PromptPal/PromptPal/compare/v1.11.3...v1.11.4) (2024-07-15)


### Bug Fixes

* **stream:** fix streaming api ([d65b274](https://github.com/PromptPal/PromptPal/commit/d65b274a748e84810d0f7a15d4f3b0c4a27525b9))

## [1.11.3](https://github.com/PromptPal/PromptPal/compare/v1.11.2...v1.11.3) (2024-07-15)


### Bug Fixes

* **stream:** fix openai streaming response ([1387b21](https://github.com/PromptPal/PromptPal/commit/1387b212d4901d641617f4d8243714b407a406fc))

## [1.11.2](https://github.com/PromptPal/PromptPal/compare/v1.11.1...v1.11.2) (2024-07-13)


### Features

* **ci:** update docker hub readme ([4d262bb](https://github.com/PromptPal/PromptPal/commit/4d262bbf2dc28e9a9c6beec8933cf861e5d57bfc))


### Bug Fixes

* **ci:** update docker hub push tags ([c21b84a](https://github.com/PromptPal/PromptPal/commit/c21b84a81f77fd68bd529c0f0c29b8e5ec5651a8))


### Miscellaneous Chores

* release 1.11.2 ([6216df3](https://github.com/PromptPal/PromptPal/commit/6216df3995ad8ad07b070a6221abfac92292480e))

## [1.11.1](https://github.com/PromptPal/PromptPal/compare/v1.11.0...v1.11.1) (2024-07-12)


### Bug Fixes

* **ci:** fix github token ([f793cf7](https://github.com/PromptPal/PromptPal/commit/f793cf7997f66c2e5a2cc331e5a6d184606d0419))
* **docs:** update doc ([8717461](https://github.com/PromptPal/PromptPal/commit/8717461e3acf6514ac0c43f97534e9e7fca8c806))

## [1.11.0](https://github.com/PromptPal/PromptPal/compare/v1.10.1...v1.11.0) (2024-07-11)


### Features

* **stream:** add streaming api support ([75068d6](https://github.com/PromptPal/PromptPal/commit/75068d6a0b0f7461e2b80e0e3eebd849a9432a79))


### Bug Fixes

* **stream:** remove 501 response in stream api ([88c5d0f](https://github.com/PromptPal/PromptPal/commit/88c5d0fb45c1c199ecf05119f8390f7c4fe74624))
* **stream:** update stream api response ([d7c6d10](https://github.com/PromptPal/PromptPal/commit/d7c6d10e657f8d978b88adce50af3ef281946ce5))

## [1.10.1](https://github.com/PromptPal/PromptPal/compare/v1.10.0...v1.10.1) (2024-07-05)


### Bug Fixes

* **ci:** upgrade release-please to new ns ([b4d28b6](https://github.com/PromptPal/PromptPal/commit/b4d28b669005ea896777003f53fad7689e76f1ed))
* **prompt:** fix testcase ([8fedb1f](https://github.com/PromptPal/PromptPal/commit/8fedb1fe77fca812e53b71417d106f303747feb3))

## [1.10.0](https://github.com/PromptPal/PromptPal/compare/v1.9.2...v1.10.0) (2024-07-05)


### Features

* **api:** add streaming api support ([f1c145f](https://github.com/PromptPal/PromptPal/commit/f1c145f75f672bc78c6fc647b0353bc45fe577cc))
* **prompt:** add support for multiple type of variable in prompt ([cd95a93](https://github.com/PromptPal/PromptPal/commit/cd95a930bf2a0349de8fa2b5e02ef670babde01c))


### Bug Fixes

* **api:** add missing middleware ([90ab08b](https://github.com/PromptPal/PromptPal/commit/90ab08b25ccbf1a438280189452b2ae0db1890fc))
* **docs:** fix thumbnail in readme ([eb2b476](https://github.com/PromptPal/PromptPal/commit/eb2b4760e6dc7f8bdd41aee45e17528154fc761d))
* **stream:** add stream api support ([a975de6](https://github.com/PromptPal/PromptPal/commit/a975de679119ae0e33b0c2df1d0ee095e5756827))

## [1.9.2](https://github.com/PromptPal/PromptPal/compare/v1.9.1...v1.9.2) (2024-05-14)


### Bug Fixes

* **gpt4o:** enable gpt-4o pricing model ([5945054](https://github.com/PromptPal/PromptPal/commit/59450540ce3bd52605805cafd086b9b7295c39f4))

## [1.9.1](https://github.com/PromptPal/PromptPal/compare/v1.9.0...v1.9.1) (2024-05-13)


### Bug Fixes

* **ci:** fix postgres service in ci ([37d80dc](https://github.com/PromptPal/PromptPal/commit/37d80dcd7c530ea716c5a5c938c6fda813da0143))
* **ci:** update env for testing ([33d1dea](https://github.com/PromptPal/PromptPal/commit/33d1dea1c7d4d3e9bbfa3c34f76cadbd49ff0eb7))

## [1.9.0](https://github.com/PromptPal/PromptPal/compare/v1.8.0...v1.9.0) (2024-05-08)


### Features

* Add costInCents and userAgent fields to PromptCall ([8929bcf](https://github.com/PromptPal/PromptPal/commit/8929bcfe7eb74ef7d078d656a66f7e4ee8c05f99))
* Add server timing header with prompt execution duration ([d151f6c](https://github.com/PromptPal/PromptPal/commit/d151f6cf7bd2e7ad4597dddc89b9b944048875eb))
* **build:** disable CGO for release builds on Linux platform ([9920407](https://github.com/PromptPal/PromptPal/commit/99204077713bb97c5008001354829eab7b287fb4))
* Calculate and set costs for prompt input and output ([d151f6c](https://github.com/PromptPal/PromptPal/commit/d151f6cf7bd2e7ad4597dddc89b9b944048875eb))
* **db:** remove sqlite3 support ([7416aa2](https://github.com/PromptPal/PromptPal/commit/7416aa266af5133988545a912abda394052cdd04))
* Set prompt tokens based on actual count ([d151f6c](https://github.com/PromptPal/PromptPal/commit/d151f6cf7bd2e7ad4597dddc89b9b944048875eb))
* Set user agent info in prompt execution stats ([d151f6c](https://github.com/PromptPal/PromptPal/commit/d151f6cf7bd2e7ad4597dddc89b9b944048875eb))


### Bug Fixes

* **build:** remove cgo support ([19f90e6](https://github.com/PromptPal/PromptPal/commit/19f90e6e31e81bbce33432d9fc1bb07866589fa0))

## [1.8.0](https://github.com/PromptPal/PromptPal/compare/v1.7.5...v1.8.0) (2024-04-30)


### Features

* **history:** add support for prompt histories. ([8292462](https://github.com/PromptPal/PromptPal/commit/829246238b18d30194befb386e5d3958b836cbc2))
* **history:** Add tests for retrieving prompt histories ([6e91221](https://github.com/PromptPal/PromptPal/commit/6e91221122db3a5b4e37ec69ca3a80c9439281a1))


### Bug Fixes

* **history:** Fix data type in assertion for ID comparison in TestUpdatePrompt ([289501b](https://github.com/PromptPal/PromptPal/commit/289501bfbd0167543f75078f5e9a068e6e6fd32b))
* **history:** Update prompt test to include latest calls and modifier information. ([c366728](https://github.com/PromptPal/PromptPal/commit/c3667285186400b4e7895db9c5b5e9c9fd986b6d))

## [1.7.5](https://github.com/PromptPal/PromptPal/compare/v1.7.4...v1.7.5) (2024-04-13)


### Bug Fixes

* **metrics:** fix time parser in project metrics by date ([8114013](https://github.com/PromptPal/PromptPal/commit/8114013aef748991e091d0935df9f982d6444b96))

## [1.7.4](https://github.com/PromptPal/PromptPal/compare/v1.7.3...v1.7.4) (2024-04-10)


### Bug Fixes

* **deps:** upgrade deps ([a9ceb52](https://github.com/PromptPal/PromptPal/commit/a9ceb529c1a982fdb9e3bc0753321e3f0b0c6a37))

## [1.7.3](https://github.com/PromptPal/PromptPal/compare/v1.7.2...v1.7.3) (2024-03-11)


### Bug Fixes

* **docs:** fix docs and make a release ([74a51ab](https://github.com/PromptPal/PromptPal/commit/74a51abf9a8f5494ef48b7d5392a6f19a0aaf3aa))

## [1.7.2](https://github.com/PromptPal/PromptPal/compare/v1.7.1...v1.7.2) (2024-03-11)


### Bug Fixes

* **ci:** update ci script to run test in postgres ([17320cf](https://github.com/PromptPal/PromptPal/commit/17320cf93b228eee9a625afd957e618ff295becd))
* **ci:** update deps and add postgres in ci for test ([96e03f0](https://github.com/PromptPal/PromptPal/commit/96e03f02e8969229a7f48e59b342e886a3354de2))

## [1.7.1](https://github.com/PromptPal/PromptPal/compare/v1.7.0...v1.7.1) (2024-03-10)


### Bug Fixes

* **db:** add warning info for sqlite project ([ca52d34](https://github.com/PromptPal/PromptPal/commit/ca52d34203504dc2133eccea239e2bd6ac4215ce))
* **docs:** update docs ([3a7d618](https://github.com/PromptPal/PromptPal/commit/3a7d618ded502cc88afc783a6c19e3f7f2ab2f0a))

## [1.7.0](https://github.com/PromptPal/PromptPal/compare/v1.6.0...v1.7.0) (2024-03-10)


### Features

* **metric:** add prompt metrics in last 7 days ([b48cd61](https://github.com/PromptPal/PromptPal/commit/b48cd61cdd12e91b7a9c1c3c2526ba70d192362e))
* **prompt:** add prompt p50, p90, p99 metrics ([fcfda4a](https://github.com/PromptPal/PromptPal/commit/fcfda4ac9339b91609245b1d21b79aecdbfd67d8))
* **sso:** add sso settings api for ensure sso enabled ([884d285](https://github.com/PromptPal/PromptPal/commit/884d285200157383b59966b3ea206a8ee15d9b70))
* **sso:** add sso support for auth. solve [#17](https://github.com/PromptPal/PromptPal/issues/17) ([da0d708](https://github.com/PromptPal/PromptPal/commit/da0d7081da184655b5c34afdfec8e0f5644b3a7e))


### Bug Fixes

* **sso:** use oidc instead of the original google api ([541596a](https://github.com/PromptPal/PromptPal/commit/541596a02fa4211297857b62d2d72da3f7788d42))

## [1.6.0](https://github.com/PromptPal/PromptPal/compare/v1.5.1...v1.6.0) (2024-02-24)


### Features

* **db:** add gemini support ([efb0c70](https://github.com/PromptPal/PromptPal/commit/efb0c7082f2ddafe7e35fd7d9cea702a0b0f66d9))
* **gemini:** add gemini support ([6cfd442](https://github.com/PromptPal/PromptPal/commit/6cfd442d792f663cd4a2781c072a2d8fafd89947))
* **gemini:** update gemini support ([00cf12a](https://github.com/PromptPal/PromptPal/commit/00cf12a071c22c099753f3141f3a2f85f4f3d992))


### Bug Fixes

* always set base URL path to "/v1" for API requests to ensure compatibility with OpenAI API ([ecc60b5](https://github.com/PromptPal/PromptPal/commit/ecc60b58a1d3396db845c0ef98497b811fb98f3a))
* **ci:** upgrade ci action versions ([4287528](https://github.com/PromptPal/PromptPal/commit/428752871993bdb2388e31192e88a5fd5e6575e0))
* **ci:** upgrade mockery to latest ([fff2b7c](https://github.com/PromptPal/PromptPal/commit/fff2b7ccf04fb30f7aa10d5395fe717a989f790d))
* **deps:** update deps ([0bbde18](https://github.com/PromptPal/PromptPal/commit/0bbde186a7214978228fce51b83d9f97eca15aa6))
* **git:** add DS_Store to gitignore ([70c215c](https://github.com/PromptPal/PromptPal/commit/70c215cb1b4611b8a001a096ec6c394c428bc7c8))
* **openai:** fix openai fetcher ([9efc15b](https://github.com/PromptPal/PromptPal/commit/9efc15b9639b6caf1e23e83432592ede1e70d2c7))

## [1.5.1](https://github.com/PromptPal/PromptPal/compare/v1.5.0...v1.5.1) (2023-11-17)


### Bug Fixes

* **openai:** remove json format in chat api if not qualified for model ([64d307a](https://github.com/PromptPal/PromptPal/commit/64d307aad1c506b942f00e1f6709f3124a2888c7))

## [1.5.0](https://github.com/PromptPal/PromptPal/compare/v1.4.4...v1.5.0) (2023-11-16)


### Features

* **docs:** add video for readme ([91c8c12](https://github.com/PromptPal/PromptPal/commit/91c8c120a15d14984ad4aa9387a3d0fe9f2d8a31))


### Bug Fixes

* **docs:** update docs ([da8b9c3](https://github.com/PromptPal/PromptPal/commit/da8b9c37bff61b65d54522de3e63fc8c014071f1))
* **openai:** use new package and support json reply ([3a08315](https://github.com/PromptPal/PromptPal/commit/3a083155730d1af009aed6f4e46c44517b23f67b))

## [1.4.4](https://github.com/PromptPal/PromptPal/compare/v1.4.3...v1.4.4) (2023-10-29)


### Bug Fixes

* **app:** just update some information ([16d95fd](https://github.com/PromptPal/PromptPal/commit/16d95fd6fe1e89b19f843c885a42145656388821))

## [1.4.3](https://github.com/PromptPal/PromptPal/compare/v1.4.2...v1.4.3) (2023-10-28)


### Bug Fixes

* **project:** fix project maxToken data ([df370a4](https://github.com/PromptPal/PromptPal/commit/df370a446c4b391c15faf3488f56823cf3a12bf4))

## [1.4.2](https://github.com/PromptPal/PromptPal/compare/v1.4.1...v1.4.2) (2023-10-26)


### Bug Fixes

* **app:** update package's version ([44f8eb1](https://github.com/PromptPal/PromptPal/commit/44f8eb12fcd78396ee7770a8990f364c1bce3f09))

## [1.4.1](https://github.com/PromptPal/PromptPal/compare/v1.4.0...v1.4.1) (2023-10-19)


### Bug Fixes

* **app:** upgrade go version to latest ([3b0e84c](https://github.com/PromptPal/PromptPal/commit/3b0e84c161234cb09c2e1436beaee54b4cb21b2b))

## [1.4.0](https://github.com/PromptPal/PromptPal/compare/v1.3.2...v1.4.0) (2023-10-01)


### Features

* **promptcall:** add variables info when prompt debugging ([fba4c01](https://github.com/PromptPal/PromptPal/commit/fba4c014c4102e5adf0b19602596a4ef67ee9d15))

## [1.3.2](https://github.com/PromptPal/PromptPal/compare/v1.3.1...v1.3.2) (2023-10-01)


### Bug Fixes

* **auth:** ignore route api middleware tests ([4e8f6ce](https://github.com/PromptPal/PromptPal/commit/4e8f6ce2b55704edc941ff32c38fba6b3282871e))
* **prompt:** ignore auth middleware for some reason and add project to prompt ([1ce78ab](https://github.com/PromptPal/PromptPal/commit/1ce78ab29424a4b82e59dd6b68160fb6373af1d2))

## [1.3.1](https://github.com/PromptPal/PromptPal/compare/v1.3.0...v1.3.1) (2023-08-24)


### Bug Fixes

* **graphql:** fix graphql struct types and add missing http route ([5c49f3c](https://github.com/PromptPal/PromptPal/commit/5c49f3cb316934d753a8ce1256bcc26f1c22bb38))

## [1.3.0](https://github.com/PromptPal/PromptPal/compare/v1.2.0...v1.3.0) (2023-08-23)


### Features

* **graphql:** add more testcases ([14597d9](https://github.com/PromptPal/PromptPal/commit/14597d953a6ddd397dadfafdaf35a5aded2c37ce))
* **graphql:** add some tests for graphql api ([9cef56b](https://github.com/PromptPal/PromptPal/commit/9cef56bfa0706713b843a565adb396f92afee41d))
* **graphql:** update graphql test cases ([f79ff15](https://github.com/PromptPal/PromptPal/commit/f79ff1594bd283b768c0add2649dfa18c787cf3d))
* **schema:** add more test case for project and prompt in graphql ([4496ed6](https://github.com/PromptPal/PromptPal/commit/4496ed6d2c89ac41c2fd315f465059fb861b84a0))


### Bug Fixes

* **graphql:** fix auth testcase ([0555de0](https://github.com/PromptPal/PromptPal/commit/0555de09f22f1d3b309abc8e2fd1b7ad97f17d52))
* **project:** fix project args type ([55b545a](https://github.com/PromptPal/PromptPal/commit/55b545ac21ee439d88468d5bda496be87e488e87))

## [1.2.0](https://github.com/PromptPal/PromptPal/compare/v1.1.3...v1.2.0) (2023-08-06)


### Features

* **api:** support graphql as v2 api ([47d063a](https://github.com/PromptPal/PromptPal/commit/47d063af0796e1b7350fb97380ad8a2b22b5ea02))
* **calls:** add prompt calls for graphql api ([7530967](https://github.com/PromptPal/PromptPal/commit/7530967c54f40685ecb6224a4f789f53afcab5a2))
* **gql:** fix gql ([cd6c6e9](https://github.com/PromptPal/PromptPal/commit/cd6c6e974110bc83927894ff564aec3cc67c059b))
* **graphql:** update graphql api ([f8449dd](https://github.com/PromptPal/PromptPal/commit/f8449dd43c949dce8c2f6463bc37ab6761b0f371))
* **graphql:** update project and prompts list api in graphql ([48ca977](https://github.com/PromptPal/PromptPal/commit/48ca97785ff9a6c84436b2e15ee7a6c364af9a29))
* **graphql:** upgrade to graphql api ([f492455](https://github.com/PromptPal/PromptPal/commit/f492455ee36939417d02573dec0b8c745d6e1a12))
* **http:** add project, prompt and openToken graphql api ([254c4c7](https://github.com/PromptPal/PromptPal/commit/254c4c7fa5012cf866816e2d6538eb9c07f65ed6))


### Bug Fixes

* **graphql:** update user graphql api and update release tag ([e090789](https://github.com/PromptPal/PromptPal/commit/e0907892a78ed48b93bdc64487dfe2ff34c931e3))

## [1.1.3](https://github.com/PromptPal/PromptPal/compare/v1.1.2...v1.1.3) (2023-07-21)


### Bug Fixes

* **prompt:** fix prompt call recording issue and make the project of ([c4d40f1](https://github.com/PromptPal/PromptPal/commit/c4d40f12875fe278ecbb4d655f655e86c0f1d2e7))

## [1.1.2](https://github.com/PromptPal/PromptPal/compare/v1.1.1...v1.1.2) (2023-07-20)


### Bug Fixes

* **prompt:** fix history and prompt relation in db ([d182a07](https://github.com/PromptPal/PromptPal/commit/d182a0701b3ca00926953a5038d70d0e6a526530))

## [1.1.1](https://github.com/PromptPal/PromptPal/compare/v1.1.0...v1.1.1) (2023-07-18)


### Bug Fixes

* **docker:** fix docker .env mapping ([348745b](https://github.com/PromptPal/PromptPal/commit/348745b14b7fc0ec2b48d87d95b57053602e3715))

## [1.1.0](https://github.com/PromptPal/PromptPal/compare/v1.0.7...v1.1.0) (2023-07-17)


### Features

* **calls:** add function call ([3394450](https://github.com/PromptPal/PromptPal/commit/339445037e0f1ab8e2bf49b52fc7741fc73d9e78))
* **db:** support mysql and postgres ([835b877](https://github.com/PromptPal/PromptPal/commit/835b877465e0b7bf5478065ba59a3a892cff421d))
* **project:** add top prompts metrics ([c7dbaa0](https://github.com/PromptPal/PromptPal/commit/c7dbaa0a7e866ee3768c25f77e63f2b62fbae3da))
* **tests:** setup tests and add project prompts metrics api ([2f9c90b](https://github.com/PromptPal/PromptPal/commit/2f9c90bb31612537b4061e1fb561726bc99e240a))


### Bug Fixes

* **calls:** fix int convertion ([a21403b](https://github.com/PromptPal/PromptPal/commit/a21403baa44c9577d236453a849558c271642bde))
* **ci:** skip build if not a release in CI ([7a0f07f](https://github.com/PromptPal/PromptPal/commit/7a0f07f874d9e1309c0ccd546853d24ea9329390))


### Performance Improvements

* **api:** add cache for public api ([e8dacee](https://github.com/PromptPal/PromptPal/commit/e8daceeb3c19df1a902be5a59f6f87da2c717cc3))

## [1.0.7](https://github.com/PromptPal/PromptPal/compare/v1.0.6...v1.0.7) (2023-07-10)


### Bug Fixes

* **ci:** fix ci ([d743ed8](https://github.com/PromptPal/PromptPal/commit/d743ed8c861c8f5fe9eadc64ee370eb6ee7b0c76))
* **ci:** try to fix ci ([07d5a90](https://github.com/PromptPal/PromptPal/commit/07d5a908b3d5f19fdf50a09d61c595f1780ec797))

## [1.0.6](https://github.com/PromptPal/PromptPal/compare/v1.0.5...v1.0.6) (2023-07-10)


### Bug Fixes

* **ci:** fix ci ([379e02d](https://github.com/PromptPal/PromptPal/commit/379e02dacac9bf1194f7026a8b26c099e35782c3))

## [1.0.5](https://github.com/PromptPal/PromptPal/compare/v1.0.4...v1.0.5) (2023-07-10)


### Bug Fixes

* **ci:** fix ci ([bda7d58](https://github.com/PromptPal/PromptPal/commit/bda7d58e252f6519b9ff678f0910e01ec4bac3ba))
* **ci:** fix static assets ([e815abc](https://github.com/PromptPal/PromptPal/commit/e815abc11c5763724f54abde0d00b03edab9a351))

## [1.0.4](https://github.com/PromptPal/PromptPal/compare/v1.0.3...v1.0.4) (2023-07-10)


### Bug Fixes

* **ci:** fix ci to build final assets ([84ec93e](https://github.com/PromptPal/PromptPal/commit/84ec93e6b9c106da5846627b330d059777034e0e))

## [1.0.3](https://github.com/PromptPal/PromptPal/compare/v1.0.2...v1.0.3) (2023-07-09)


### Bug Fixes

* **app:** fix docker push tag name ([b66342f](https://github.com/PromptPal/PromptPal/commit/b66342f68965304daf5a028acc691a2c02047bf6))

## [1.0.2](https://github.com/PromptPal/PromptPal/compare/v1.0.1...v1.0.2) (2023-07-09)


### Bug Fixes

* **ci:** fix ci token ([838db08](https://github.com/PromptPal/PromptPal/commit/838db08918628c3ca4d693422b6a425d0c251d0d))

## [1.0.1](https://github.com/PromptPal/PromptPal/compare/v1.0.0...v1.0.1) (2023-07-09)


### Bug Fixes

* **ci:** get ready for emmbed fe assets to public folder ([d2774c2](https://github.com/PromptPal/PromptPal/commit/d2774c232b188494dccf9798b8cac48a27a4f1f3))

## 1.0.0 (2023-07-09)


### Features

* **app:** init project ([eafcdd6](https://github.com/PromptPal/PromptPal/commit/eafcdd694577a3b93e89121342b834c6b1e45471))
* **app:** init project and add LICENSE ([000467a](https://github.com/PromptPal/PromptPal/commit/000467a7f60f7564ebcfd6a0dbeb571e683daba9))
* **ci:** add ci config ([f4a306a](https://github.com/PromptPal/PromptPal/commit/f4a306aa3e768ff2eb1cd27922dbe7d85a7bed08))
* **db:** add sqlite3 driver ([1e7cdfc](https://github.com/PromptPal/PromptPal/commit/1e7cdfc41dd3d5870b577f085bde7874f62e4686))
* **db:** update db schema ([33796c8](https://github.com/PromptPal/PromptPal/commit/33796c84e8fa22e2437fc14c8e768d4458b6046d))
* **openToken:** add openToken schema ([158f0b6](https://github.com/PromptPal/PromptPal/commit/158f0b687d0531e087c65c91d358356d50117037))
* **openToken:** add openToken support ([106e19e](https://github.com/PromptPal/PromptPal/commit/106e19e02b760d21e15ce41a08b8c37ed16c04d0))
* **prompt:** update prompt ([23532a0](https://github.com/PromptPal/PromptPal/commit/23532a0ef8cc5b3e18eed8918b1936826dad3c38))
* **prompt:** update prompt test api ([522324e](https://github.com/PromptPal/PromptPal/commit/522324ec82e19d5c9addd9d57dbaf79d898a68e3))
* **routes:** add api routes ([f7ecd6b](https://github.com/PromptPal/PromptPal/commit/f7ecd6b5dd792f309c50cbd5f2b6cc2beb36cd21))
* **routes:** add basic function ([757e79a](https://github.com/PromptPal/PromptPal/commit/757e79afd931092d1ec6db30a06f76da91a637d4))
* **routes:** add routes and db operations ([5a229f9](https://github.com/PromptPal/PromptPal/commit/5a229f9ad367cbc2e3b034b6dc7a52bb0474b2ae))


### Bug Fixes

* **api:** fix api fetcher ([24dfadd](https://github.com/PromptPal/PromptPal/commit/24dfadd0deb847bf59b2291be9da1514d2fb37da))
* **api:** fix project checker ([edf2e4b](https://github.com/PromptPal/PromptPal/commit/edf2e4b9d11ab7fbc372d3766843ec724ca46806))
* **api:** make variables as public ([f90851b](https://github.com/PromptPal/PromptPal/commit/f90851bd71ccd5086db9ac2af96307bdd1b5a5c9))
* **app:** fix openToken registion of app route ([8078663](https://github.com/PromptPal/PromptPal/commit/80786639525d5d2b60d0635776b9e019e6d3832b))
* **routes:** update routes ([7ab9894](https://github.com/PromptPal/PromptPal/commit/7ab9894a07ce89c33fdbf1beaf6584683a1362b3))
