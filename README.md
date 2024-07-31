# go-misc
[![codecov](https://codecov.io/gh/chanzuckerberg/go-misc/branch/master/graph/badge.svg)](https://codecov.io/gh/chanzuckerberg/go-misc)

**Please note**: If you believe you have found a security issue, _please responsibly disclose_ by contacting us at [security@chanzuckerberg.com](mailto:security@chanzuckerberg.com).

----

This is a collection of Go libraries and projects used by other projects within CZI.

## Sub-projects

### aws
An AWS client that aims to standardize the way we mock and write aws tests

If you recently changed laptops and struggled with new go dependencies, use [this page to help](https://stackoverflow.com/questions/42614380/go-install-not-working-with-zsh). Full link: https://stackoverflow.com/questions/42614380/go-install-not-working-with-zsh

### config
A utility for loading configuration from environment variables and files. See the [config](config/README.md) package for more information.

### lambda
A collection of lambda functions

### kmsauth
A port of [python-kmsauth](https://github.com/lyft/python-kmsauth) to go

### slack
A slack client that aims to standardize the way we interact and write tests for slack

### ver
Code for handling versions in go programs.

## Contributing
Contributions and ideas are welcome! Please don't hesitate to open an issue or send a pull request.

