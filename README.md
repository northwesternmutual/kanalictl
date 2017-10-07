# kanalictl

[![Travis](https://img.shields.io/travis/northwesternmutual/kanalictl/master.svg?style=flat-square)](https://travis-ci.org/northwesternmutual/kanalictl) [![Coveralls](https://img.shields.io/coveralls/northwesternmutual/kanalictl/master.svg?style=flat-square)](https://coveralls.io/github/northwesternmutual/kanalictl)

> cli tool for Kanali

# Installation

```sh
# replace 'darwin' and 'amd64' with your OS and ARCH
$ curl -O https://s3.amazonaws.com/kanalictl/release/$(curl -s https://s3.amazonaws.com/kanalictl/release/latest.txt)/darwin/amd64/kanalictl
$ chmod +x kanalictl
$ sudo mv kanalictl /usr/local/bin/kanalictl
$ kanalictl -h
```

# Local Development

Below are the steps to follow if you want to build/run/test locally. [Glide](https://glide.sh/) is a dependency.

```sh
$ mkdir -p $GOPATH/src/github.com/northwesternmutual
$ cd $GOPATH/src/github.com/northwesternmutual
$ git clone git@github.com:northwesternmutual/kanalictl.git
$ cd kanalictl
$ make kanalictl
$ ./kanalictl --help
```

# Usage

```sh
$ kanalictl [command] [subcommand] [flags]
$ kanalictl -h
```