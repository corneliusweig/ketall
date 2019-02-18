<!--
Need a real "get-all" command
https://github.com/kubernetes/kubectl/issues/527#issue-355158795
-->

# ketall
Kubectl plugin to show really all kubernetes resources

## Intro
For a complete overview of all resources in a kubernetes cluster, `kubectl get all --all-namespaces` is not enough, because it simply does not show everything.
This helper lists _really_ all resources the cluster has to offer.

## Demo
![ketall demo](doc/demo.gif "ketall demo")

## Examples
Get all resources...
- ```
  ketall
  ```

- ... at cluster level
  ```
  ketall --only-scope=cluster
  ```

- ... in some namespace
  ```
  ketall --only-scope=namespace --namespace=my-namespace
  ```

- ... using list of cached server resources
  ```
  ketall --cache
  ```
  Note that this may fail to show __really__ everything, if the cache is stale.

- ... and combine with common `kubectl` parameters
  ```
  KUBECONFIG=otherconfig ketall -o name --context some --namespace kube-system
  ```

Also see [Usage](doc/USAGE.md).

## Installation

### Via krew
```bash
kubectl krew install get-all
```

### Binaries
When using the binaries for installation, also have a look at [doc/USAGE](doc/USAGE.md).

#### Linux
```bash
curl -Lo ketall https://github.com/corneliusweig/ketall/releases/download/v0.0.1/ketall-linux-amd64 &&
  chmod +x ketall && mv ketall $GOPATH/bin/
```

#### OSX
```bash
curl -Lo ketall https://github.com/corneliusweig/ketall/releases/download/v0.0.1/ketall-darwin-amd64 &&
  chmod +x ketall && mv ketall $GOPATH/bin/
```

#### Windows
[https://github.com/corneliusweig/ketall/releases/download/v0.0.1/ketall-windows-amd64 ](https://github.com/corneliusweig/ketall/releases/download/v0.0.1/ketall-windows-amd64 )

### From source

#### Build on host

Requirements:
 - go 1.11 or newer
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
- Currently, `ketall` requires the rights to list all server resources, which is quite unrealistic for non-admins.
Instead, `ketall` should show all resources that can be accessed.

- additional arguments should be used to filter the result set

- need to verify that this works for CRD

### Credits
Idea by @ahmetb https://twitter.com/ahmetb/status/1095374856156196864
