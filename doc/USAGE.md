<!-- DO NOT MOVE THIS FILE, BECAUSE IT NEEDS A PERMANENT ADDRESS -->

# Usage
![ketall demo](demo.gif "ketall demo")

If you installed via [krew](https://github.com/GoogleContainerTools/krew) do
```bash
kubectl get-all
```

## Options

- `--only-scope=cluster` will only show cluster level resources, such as `ClusterRole`, `Namespace`, or `PersistentVolume`.
- `--only-scope=namespace` will only show namespaced resources, such as `ServiceAccount`, `Role`, `ConfigMap`, or `Endpoint`.
- `--use-cache` will consider the http cache to determine the server resources to look at. Disabled by default.
- `--exclude` will filter out the given resources (either plural names `componentstatuses` or short form `cs`). Defaults to `events` because those are rarely useful.
- ...and many standard `kubectl` options. Have a look at `kubectl get-all --help` for a full list of supported flags.
- `--verbosity` set the log level (one of debug, info, warn, error, fatal, panic).

**Hint**: If you do not have access to all resources, bulk fetching needs to be disabled. You can speed things up by explicitly excluding all resources which you may not access.

## Examples
Get all resources...
- ... excluding events (this is hardly ever useful)
  ```bash
  kubectl get-all
  ```

- ... _including_ events
  ```bash
  kubectl get-all --exclude=
  ```

- ... created in the last minute
  ```bash
  kubectl get-all --since 1m
  ```
  This flag understands typical human-readable durations such as `1m` or `1y1d1h1m1s`.

- ... at cluster level
  ```bash
  kubectl get-all --only-scope=cluster
  ```

- ... in some namespace
  ```bash
  kubectl get-all --only-scope=namespace --namespace=my-namespace
  ```

- ... using list of cached server resources
  ```bash
  kubectl get-all --use-cache
  ```
  Note that this may fail to show __really__ everything, if the http cache is stale.

- ... and combine with common `kubectl` parameters
  ```bash
  KUBECONFIG=otherconfig kubectl get-all -o name --context some --namespace kube-system
  ```

## Getting help
```bash
kubectl get-all help
```
Note that in the help, the tool is referred to as `ketall`, which is the standard name when installed as stand-alone tool.

## Completion
Completion does currently not work when used as a `kubectl` plugin. When used stand-alone, you can do
```bash
source <(ketall completion bash) # for bash users
source <(ketall completion zsh)  # for zsh users
```
Also see `ketall completion --help` for further instructions.

## Configuration
The command will look for the configuration file `ketall` (no extension) in `.` or `$HOME/.kube/`, unless overridden by the `--config` option.
The following settings can be configured:
```yaml
only-scope: cluster
namespace: default
use-cache: true
since: 1m
# only plural form or abbreviations
exclude:
- componentstatuses
- cm   # configmaps
```

## Installation

### Via krew
If you do not have `krew` installed, visit [https://github.com/GoogleContainerTools/krew](https://github.com/GoogleContainerTools/krew).
```bash
kubectl krew install get-all
```

### As `kubectl` plugin
Most users will have installed `ketall` via [krew](https://github.com/GoogleContainerTools/krew),
so the plugin is already correctly installed.
Otherwise, rename `ketall` to `kubectl-get_all` and put it in some directory from your `$PATH` variable.
Then you can invoke the plugin via `kubectl get-all`

### Standalone
Put the `ketall` binary in some directory from your `$PATH` variable. For example
```bash
sudo mv -i ketall /usr/bin/ketall
```
Then you can invoke the plugin via `ketall`
