# go-libp2p-bootstrap

[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)
[![Build Status](https://semaphoreci.com/api/v1/florianlenz/go-libp2p-bootstrap/branches/master/badge.svg)](https://semaphoreci.com/florianlenz/go-libp2p-bootstrap)

> bootstrapping module for libp2p

This module provides functionality to bootstrap libp2p nodes based on an interval and the amount of connected peers.

## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [API](#api)
- [Maintainers](#maintainers)
- [Contribute](#contribute)
- [License](#license)

## Install

```go
// first import the package
gx import <multihash-of-the-package>

// rewrite the import path
gx-go rw
```

## Usage

```go

conf := bootstrap.Config{
    BootstrapPeers    []string{
        "multi-address-of-peer"
        "multi-address-of-peer"
    }
    MinPeers          2
    BootstrapInterval time.Second * 5
    HardBootstrap     time.Minute
}

bootstrapper, err := bootstrap.New(host, )
if err != nil {
    panic(err)
}

bootstrap.Start()

```

## API
After creating an bootstrap instance you have the following functionality available:

- `Start` To start the bootstrapping go routines. It will also issue an initial bootstrapping.
- `Bootstrap` Does a manual bootstrap.
- `Close` Stop the bootstrapping go routines.

## Maintainers

[@florianlenz](https://github.com/florianlenz)

## Contribute

PRs accepted.

Small note: If editing the README, please conform to the [standard-readme](https://github.com/RichardLitt/standard-readme) specification.

## License

MIT © 2018 Florian Lenz
