# ketall
Kubectl plugin to show really all kubernetes resources

## Intro
For a complete overview of all resources in a kubernetes cluster, `kubectl get all --all-namespaces` is not enough, because it only lists namespaced resources.
This helper will list __all__ resources the cluster has to offer.

## Demo
![ketall demo](doc/demo.gif "ketall demo")

## Installation

To install `ketall` as a standalone tool:
```bash
GO111MODULE=on go get github.com/corneliusweig/ketall
```

TODO: should be possible via `kubectl krew`

## Build

Requirements:
 - go 1.11 or newer

Compiling:
```bash
GO111MODULE=on go build
```

## Examples

- Get all resources
  ```bash
  ketall
  ```

- Get all resources and use list of cached server resources
  ```bash
  ketall --cache
  ```

- Get all namespaced resources
  ```bash
  ketall --only-scope=namespace
  ```

- Get all cluster level resources
  ```bash
  ketall --only-scope=cluster
  ```

<!--
Need a real "get-all" command
https://github.com/kubernetes/kubectl/issues/527#issue-355158795
-->

## Future
- Currently, `ketall` requires the rights to list all server resources, which is quite unrealistic for non-admins.
Instead, `ketall` should show all resources that can be accessed.

- additional arguments should be used to filter the result set

## Credits
Idea by @ahmetb https://twitter.com/ahmetb/status/1095374856156196864
