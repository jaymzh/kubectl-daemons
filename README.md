# kubectl-daemons

## Intro

kubectl-daemons is a kubectl plugin to help with daemonset work. When running
a kubernetes infrastructure, it is common to do lots of work with deamonsets,
but the tools are optimized for other types of k8s objects. With daemonsets,
you often know a node and a DS you care about, and this plugin lets you focus
on that.

Previously you'd do things like:

```shell
kubectl get pods -o wide | grep <daemonset> | grep <node>
```

Now you can do:
```shell
kubectl d get <daemonset> -N <node>
```

## General Usage

Here's a bunch more examples:

* Get all pods from a specific daemonset: `kubectl d get <daemonset>`
* Get all pods from all daemonsets on a specific node: `kubectl d get -N <nodename>`
* Delete the pod from a daemonset on a specific node: `kubectl d delete <daemonset> -N <nodename>`

Like most kubectl commands it takes `-n <namespace>` wherever applicable.

Other commands coming soon:

* logs: `kubectl d logs <daemonset> -N <node>` (-N would be required)
* describe: `kubectl d describe <daemonset> -N <node>` (same)

## Building

The quick-n-easy way to build is:

```shell
go build -o kubectl-d
```

Then stick `kubectl-d` somewhere in your path.

## Releasing (WIP)

Build a scratch release with:

```
goreleaser release --skip-validate --snapshot --clean
```

Or for a real release, make a tag and releaser will release
that tag.

```
version="v0.0.1"
git tag -a v${version?} -m "version ${version?}" -s
git push origin --tags
goreleaser release
```
