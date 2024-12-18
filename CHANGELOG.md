# Changelog

## [2.5.1](https://github.com/chanzuckerberg/go-misc/compare/v2.5.0...v2.5.1) (2024-12-18)


### Bug Fixes

* v2 go.mod ([#1084](https://github.com/chanzuckerberg/go-misc/issues/1084)) ([a7a82ab](https://github.com/chanzuckerberg/go-misc/commit/a7a82ab59d09d3cf4a4b8c6cd751d909041daf47))

## [2.5.0](https://github.com/chanzuckerberg/go-misc/compare/v2.4.2...v2.5.0) (2024-12-18)


### Features

* Add scopes option to client ([#1082](https://github.com/chanzuckerberg/go-misc/issues/1082)) ([4125bab](https://github.com/chanzuckerberg/go-misc/commit/4125bab37eeef65bab06656da4dc5aafe4edcdf8))

## [2.4.2](https://github.com/chanzuckerberg/go-misc/compare/v2.4.1...v2.4.2) (2024-08-01)


### Bug Fixes

* make the aws mock-generation script work again ([#1079](https://github.com/chanzuckerberg/go-misc/issues/1079)) ([10bce11](https://github.com/chanzuckerberg/go-misc/commit/10bce115fd2a92aaa5468e28b60e878123351ec8))

## [2.4.1](https://github.com/chanzuckerberg/go-misc/compare/v2.4.0...v2.4.1) (2024-05-03)


### Bug Fixes

* suppress decompression error and treat uncompressed token in cache as cache miss ([#1058](https://github.com/chanzuckerberg/go-misc/issues/1058)) ([da1316d](https://github.com/chanzuckerberg/go-misc/commit/da1316d146ad857f601dd32b1709935be1b11a8c))

## [2.4.0](https://github.com/chanzuckerberg/go-misc/compare/v2.3.0...v2.4.0) (2024-04-05)


### Features

* compress token before storing in cache to allow for larger tokens to be cached ([#1047](https://github.com/chanzuckerberg/go-misc/issues/1047)) ([2cce231](https://github.com/chanzuckerberg/go-misc/commit/2cce2310ce46834e73599e569c5b02dfe5e015c7))


### Bug Fixes

* oidc-cli dependencies ([#1043](https://github.com/chanzuckerberg/go-misc/issues/1043)) ([43e3974](https://github.com/chanzuckerberg/go-misc/commit/43e397411f6e377d97be1e2e1e4d57ae19181e79))
* oidc-cli deps ([#1045](https://github.com/chanzuckerberg/go-misc/issues/1045)) ([ed196ca](https://github.com/chanzuckerberg/go-misc/commit/ed196ca9c1368a5981c9e4b3cc9f9bd46932b055))

## [2.3.0](https://github.com/chanzuckerberg/go-misc/compare/v2.2.3...v2.3.0) (2024-03-30)


### Features

* remove pkg/errors for stdlib errors ([2595678](https://github.com/chanzuckerberg/go-misc/commit/2595678e85b64b6eb394fa97aeba90ffa7e638d3))

## [2.2.3](https://github.com/chanzuckerberg/go-misc/compare/v2.2.2...v2.2.3) (2024-03-21)


### Bug Fixes

* Update invalid dependency versions in osutils and snowflake packages ([#1030](https://github.com/chanzuckerberg/go-misc/issues/1030)) ([864e76a](https://github.com/chanzuckerberg/go-misc/commit/864e76a776c639fd67ea114fc7e1b9f34a9f28d7))

## [2.2.2](https://github.com/chanzuckerberg/go-misc/compare/v2.2.1...v2.2.2) (2024-03-21)


### Bug Fixes

* Fix oidc-cli dependencies (osutil and pidlock reference invalid version numbers) ([#1027](https://github.com/chanzuckerberg/go-misc/issues/1027)) ([2389146](https://github.com/chanzuckerberg/go-misc/commit/238914650ee40f9ef103e384749be7857255d674))

## [2.2.1](https://github.com/chanzuckerberg/go-misc/compare/v2.2.0...v2.2.1) (2024-03-21)


### Bug Fixes

* Fix oidc-cli dependencies (osutil and pidlock reference invalid version numbers) ([#1027](https://github.com/chanzuckerberg/go-misc/issues/1027)) ([2389146](https://github.com/chanzuckerberg/go-misc/commit/238914650ee40f9ef103e384749be7857255d674))

## [2.2.0](https://github.com/chanzuckerberg/go-misc/compare/v2.1.0...v2.2.0) (2024-03-07)


### Features

* add support for config struct validate tags ([#1015](https://github.com/chanzuckerberg/go-misc/issues/1015)) ([e13bc83](https://github.com/chanzuckerberg/go-misc/commit/e13bc836bf68839700f75736a0c2f9fd6c0b3462))

## [2.1.0](https://github.com/chanzuckerberg/go-misc/compare/v2.0.1...v2.1.0) (2024-02-23)


### Features

* Upgrade AWS Mocks ([#1005](https://github.com/chanzuckerberg/go-misc/issues/1005)) ([543e56f](https://github.com/chanzuckerberg/go-misc/commit/543e56f1c67c9bebdb790327c7b3d5b2bbf7f752))

## [2.0.1](https://github.com/chanzuckerberg/go-misc/compare/v2.0.0...v2.0.1) (2024-02-23)


### Bug Fixes

* config v2 ([#1003](https://github.com/chanzuckerberg/go-misc/issues/1003)) ([31efa25](https://github.com/chanzuckerberg/go-misc/commit/31efa2598cea38456b86b47652bf47d3cac9464f))

## [2.0.0](https://github.com/chanzuckerberg/go-misc/compare/v1.12.0...v2.0.0) (2024-02-23)


### ⚠ BREAKING CHANGES

* break mono-package into individual packages ([#1000](https://github.com/chanzuckerberg/go-misc/issues/1000))

### Features

* break mono-package into individual packages ([#1000](https://github.com/chanzuckerberg/go-misc/issues/1000)) ([5151c5e](https://github.com/chanzuckerberg/go-misc/commit/5151c5e6a03d706156ac0a5b437875ab1600af6c))

## [1.12.0](https://github.com/chanzuckerberg/go-misc/compare/v1.11.1...v1.12.0) (2024-02-21)


### Features

* add config package ([#996](https://github.com/chanzuckerberg/go-misc/issues/996)) ([41f253a](https://github.com/chanzuckerberg/go-misc/commit/41f253a925cadd0d63025ec5b83eeb39791faefa))
