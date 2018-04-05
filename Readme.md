# LibP2P Bootstrap bundle
[![Build Status](https://semaphoreci.com/api/v1/florianlenz/go-libp2p-bootstrap/branches/master/badge.svg)](https://semaphoreci.com/florianlenz/go-libp2p-bootstrap)

## Install

1. You first need to import the project by running `gx import HASH`. The latest version can be found in the `.gx/lastpubver`.
2. Import it as you are used to and don't forget the use `gx-go rw` to replace the import path of the repo with the local path.

## Development

1. Clone this repo
2. Run `make deps`
3. Run `make install`

- Use `make deps_hack` to rewrite the import paths
- Use `make deps_hack_revert` to revert `deps_hack`