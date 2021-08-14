# ketall
[![Build Status](https://travis-ci.com/corneliusweig/ketall.svg?branch=master)](https://travis-ci.com/corneliusweig/ketall)
[![Go Report Card](https://goreportcard.com/badge/corneliusweig/ketall)](https://goreportcard.com/report/corneliusweig/ketall)
[![LICENSE](https://img.shields.io/github/license/corneliusweig/ketall.svg)](https://github.com/corneliusweig/ketall/blob/master/LICENSE)
[![Releases](https://img.shields.io/github/release-pre/corneliusweig/ketall.svg)](https://github.com/corneliusweig/ketall/releases)
<!-- [![Code Coverage](https://codecov.io/gh/corneliusweig/ketall/branch/master/graph/badge.svg)](https://codecov.io/gh/corneliusweig/ketall) -->

Kubectl plugin to show really all kubernetes resources

## Intro
For a complete overview of all resources in a kubernetes cluster, `kubectl get all --all-namespaces` is not enough, because it simply does not show everything.
This helper lists _really_ all resources the cluster has to offer.

## Demo
![ketall demo](doc/demo.gif "ketall demo")

## Examples
Get all resources...
- ... excluding events (this is hardly ever useful)
  ```bash
  ketall
  ```

- ... _including_ events
  ```bash
  ketall --exclude=
  ```

- ... created in the last minute
  ```bash
  ketall --since 1m
  ```
  This flag understands typical human-readable durations such as `1m` or `1y1d1h1m1s`.

- ... in the default namespace
  ```bash
  ketall --namespace=default
  ```

- ... at cluster level
  ```bash
  ketall --only-scope=cluster
  ```

- ... using list of cached server resources
  ```bash
  ketall --use-cache
  ```
  Note that this may fail to show __really__ everything, if the http cache is stale.

- ... and combine with common `kubectl` options
  ```bash
  KUBECONFIG=otherconfig ketall -o name --context some --namespace kube-system --selector run=skaffold
  ```

Also see [Usage](doc/USAGE.md).

## Installation
There are several ways to install `ketall`. The recommended installation method is via `krew`.

### Via krew
Krew is a `kubectl` plugin manager. If you have not yet installed `krew`, get it at
[https://github.com/kubernetes-sigs/krew](https://github.com/kubernetes-sigs/krew).
Then installation is as simple as
```bash
kubectl krew install get-all
```
The plugin will be available as `kubectl get-all`, see [doc/USAGE](doc/USAGE.md) for further details.

### Binaries
When using the binaries for installation, also have a look at [doc/USAGE](doc/USAGE.md).

#### Linux
```bash
curl -Lo ketall.gz https://github.com/corneliusweig/ketall/releases/download/v1.3.8/ketall-amd64-linux.tar.gz && \
  gunzip ketall.gz && chmod +x ketall && mv ketall $GOPATH/bin/
```

#### OSX
```bash
curl -Lo ketall.gz https://github.com/corneliusweig/ketall/releases/download/v1.3.8/ketall-amd64-darwin.tar.gz && \
  gunzip ketall.gz && chmod +x ketall && mv ketall $GOPATH/bin/
```

#### Windows
[https://github.com/corneliusweig/ketall/releases/download/v1.3.8/ketall-amd64-windows.zip](https://github.com/corneliusweig/ketall/releases/download/v1.3.8/ketall-amd64-windows.zip)

### From source

#### Build on host

Requirements:
 - go 1.16 or newer
 - GNU make
 - git

Compiling:
```bash
export PLATFORMS=$(go env GOOS)
make all   # binaries will be placed in out/
```

#### Build in docker
Requirements:
 - docker

Compiling:
```bash
mkdir ketall && chdir ketall
curl -Lo Dockerfile https://raw.githubusercontent.com/corneliusweig/ketall/master/Dockerfile
docker build . -t ketall-builder
docker run --rm -v $PWD:/go/bin/ --env PLATFORMS=$(go env GOOS) ketall-builder
docker rmi ketall-builder
```
Binaries will be placed in the current directory.

## Future
- additional arguments could be used to filter the result set

### Credits
Idea by @ahmetb https://twitter.com/ahmetb/status/1095374856156196864
