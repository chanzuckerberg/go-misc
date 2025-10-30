# Changelog

## [3.1.0](https://github.com/chanzuckerberg/go-misc/compare/oidc_cli-v3.0.2...oidc_cli-v3.1.0) (2025-07-25)


### Features

* Update go-jose v4 and other dependencies ([#1124](https://github.com/chanzuckerberg/go-misc/issues/1124)) ([a8bee75](https://github.com/chanzuckerberg/go-misc/commit/a8bee751b630dabef14caa0620640195e9c5f2d7))

## [3.0.2](https://github.com/chanzuckerberg/go-misc/compare/oidc_cli-v3.0.1...oidc_cli-v3.0.2) (2025-07-25)


### Misc

* Upgrade go-oidc to v3 ([#1121](https://github.com/chanzuckerberg/go-misc/issues/1121)) ([fdcca18](https://github.com/chanzuckerberg/go-misc/commit/fdcca18bb4d19ea5c9e8bd16d4d8bfbeb4a535a8))

## [3.0.1](https://github.com/chanzuckerberg/go-misc/compare/oidc_cli-v3.0.0...oidc_cli-v3.0.1) (2025-07-25)


### Misc

* **deps:** bump github.com/go-jose/go-jose/v4 from 4.0.0 to 4.0.5 in /oidc_cli ([#1119](https://github.com/chanzuckerberg/go-misc/issues/1119)) ([7ecd3fb](https://github.com/chanzuckerberg/go-misc/commit/7ecd3fb4d37438fcdf57a370e38655171ab3334c))

## [3.0.0](https://github.com/chanzuckerberg/go-misc/compare/oidc_cli-v2.4.3...oidc_cli-v3.0.0) (2025-07-25)


### ⚠ BREAKING CHANGES

* Upgrade go-jose to v4 ([#1117](https://github.com/chanzuckerberg/go-misc/issues/1117))

### Bug Fixes

* Upgrade go-jose to v4 ([#1117](https://github.com/chanzuckerberg/go-misc/issues/1117)) ([98ac346](https://github.com/chanzuckerberg/go-misc/commit/98ac34687d667f3ff0ee63279550e2a1973018d0))

## [2.4.3](https://github.com/chanzuckerberg/go-misc/compare/oidc_cli-v2.4.2...oidc_cli-v2.4.3) (2025-07-21)


### Misc

* **deps:** bump golang.org/x/oauth2 from 0.25.0 to 0.27.0 in /oidc_cli ([#1115](https://github.com/chanzuckerberg/go-misc/issues/1115)) ([e397641](https://github.com/chanzuckerberg/go-misc/commit/e397641da8e395950c5670958abb3e060fb02435))

## [2.4.2](https://github.com/chanzuckerberg/go-misc/compare/oidc_cli-v2.4.1...oidc_cli-v2.4.2) (2025-04-16)


### Misc

* bump golang.org/x/crypto to 0.35.0 ([#1107](https://github.com/chanzuckerberg/go-misc/issues/1107)) ([9956e3b](https://github.com/chanzuckerberg/go-misc/commit/9956e3b2797acf329133cacbe33bab2a7df82ee8))

## [2.4.1](https://github.com/chanzuckerberg/go-misc/compare/oidc_cli-v2.4.0...oidc_cli-v2.4.1) (2025-01-30)


### Bug Fixes

* Go JOSE vulnerable to Improper Handling of Highly Compressed Data (Data Amplification) ([#1094](https://github.com/chanzuckerberg/go-misc/issues/1094)) ([836b377](https://github.com/chanzuckerberg/go-misc/commit/836b37716c58f6d70888db6e7b28af984a487a09))

## [2.4.0](https://github.com/chanzuckerberg/go-misc/compare/oidc_cli-v2.3.0...oidc_cli-v2.4.0) (2025-01-10)


### Features

* Remove refresh token if token is too big ([#1092](https://github.com/chanzuckerberg/go-misc/issues/1092)) ([23d4604](https://github.com/chanzuckerberg/go-misc/commit/23d4604bb218629b26f4fc2cf97a9b418c865146))
* Upgrade keyring package to latest ([#1093](https://github.com/chanzuckerberg/go-misc/issues/1093)) ([3eda425](https://github.com/chanzuckerberg/go-misc/commit/3eda425a903a4464730ab294806aa8f5ba7169e2))


### Bug Fixes

* Go JOSE vulnerable to Improper Handling of Highly Compressed Data (Data Amplification) ([#1090](https://github.com/chanzuckerberg/go-misc/issues/1090)) ([1b28605](https://github.com/chanzuckerberg/go-misc/commit/1b28605532373fa7688fcab875b45503ae393563))

## [2.3.0](https://github.com/chanzuckerberg/go-misc/compare/oidc_cli-v2.2.1...oidc_cli-v2.3.0) (2024-12-20)


### Features

* Update golang.org/x/net/html ([#1086](https://github.com/chanzuckerberg/go-misc/issues/1086)) ([96a6253](https://github.com/chanzuckerberg/go-misc/commit/96a62530abd701abcfa79ea0740ef6ef1980fa08))

## [2.2.1](https://github.com/chanzuckerberg/go-misc/compare/oidc_cli-v2.2.0...oidc_cli-v2.2.1) (2024-12-18)


### Bug Fixes

* v2 go.mod ([#1084](https://github.com/chanzuckerberg/go-misc/issues/1084)) ([a7a82ab](https://github.com/chanzuckerberg/go-misc/commit/a7a82ab59d09d3cf4a4b8c6cd751d909041daf47))

## [2.2.0](https://github.com/chanzuckerberg/go-misc/compare/oidc_cli-v2.1.1...oidc_cli-v2.2.0) (2024-12-18)


### Features

* Add scopes option to client ([#1082](https://github.com/chanzuckerberg/go-misc/issues/1082)) ([4125bab](https://github.com/chanzuckerberg/go-misc/commit/4125bab37eeef65bab06656da4dc5aafe4edcdf8))

## [2.1.1](https://github.com/chanzuckerberg/go-misc/compare/oidc_cli-v2.1.0...oidc_cli-v2.1.1) (2024-05-03)


### Bug Fixes

* suppress decompression error and treat uncompressed token in cache as cache miss ([#1058](https://github.com/chanzuckerberg/go-misc/issues/1058)) ([da1316d](https://github.com/chanzuckerberg/go-misc/commit/da1316d146ad857f601dd32b1709935be1b11a8c))

## [2.1.0](https://github.com/chanzuckerberg/go-misc/compare/oidc_cli-v2.0.1...oidc_cli-v2.1.0) (2024-04-05)


### Features

* compress token before storing in cache to allow for larger tokens to be cached ([#1047](https://github.com/chanzuckerberg/go-misc/issues/1047)) ([2cce231](https://github.com/chanzuckerberg/go-misc/commit/2cce2310ce46834e73599e569c5b02dfe5e015c7))


### Bug Fixes

* oidc-cli dependencies ([#1043](https://github.com/chanzuckerberg/go-misc/issues/1043)) ([43e3974](https://github.com/chanzuckerberg/go-misc/commit/43e397411f6e377d97be1e2e1e4d57ae19181e79))
* oidc-cli deps ([#1045](https://github.com/chanzuckerberg/go-misc/issues/1045)) ([ed196ca](https://github.com/chanzuckerberg/go-misc/commit/ed196ca9c1368a5981c9e4b3cc9f9bd46932b055))

## [2.0.1](https://github.com/chanzuckerberg/go-misc/compare/oidc_cli-v2.0.0...oidc_cli-v2.0.1) (2024-03-21)


### Bug Fixes

* Fix oidc-cli dependencies (osutil and pidlock reference invalid version numbers) ([#1027](https://github.com/chanzuckerberg/go-misc/issues/1027)) ([2389146](https://github.com/chanzuckerberg/go-misc/commit/238914650ee40f9ef103e384749be7857255d674))

## [2.0.0](https://github.com/chanzuckerberg/go-misc/compare/oidc_cli-v1.12.0...oidc_cli-v2.0.0) (2024-02-23)


### ⚠ BREAKING CHANGES

* break mono-package into individual packages ([#1000](https://github.com/chanzuckerberg/go-misc/issues/1000))

### Features

* break mono-package into individual packages ([#1000](https://github.com/chanzuckerberg/go-misc/issues/1000)) ([5151c5e](https://github.com/chanzuckerberg/go-misc/commit/5151c5e6a03d706156ac0a5b437875ab1600af6c))
* Push oidc implementation down to exclude aws-sdk-go from depenedencies ([#532](https://github.com/chanzuckerberg/go-misc/issues/532)) ([fad5836](https://github.com/chanzuckerberg/go-misc/commit/fad5836ca8b86dc6b8496f66919a35378a3ef115))
