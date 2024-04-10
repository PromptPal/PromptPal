# Changelog

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
