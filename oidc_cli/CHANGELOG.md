# Changelog

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
