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

Get pods from a daemonset:

```bash
kubectl d get <daemonset>
```

Or get the pod from a daemonset on a specific node

```bash
kubectl d get <daemonset> -N <node>
```

Or get all daemonset-related pods on a node.  I think of this as the equivalent
of asking systemd to list all services on a node. :)

```bash
kubectl d get -N <nodename>
```

NOTE: `-o` is not yet implemented for `get`.

Or you can delete the pod from a daemonset on a specific node

```bash
kubectl d delete <daemonset> -N <nodename>
```

You can do logs as well:

```bash
kubectl d logs <daemonset> -N <node>
```

NOTE: -N is required here.

And you can even describe:

```bash
kubectl d describe <daemonset> [-N <node>] # node optional
```

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

## Thanks

A huge thanks to Benjamin Muschko's [Writing your first kubectl plugin](https://bmuschko.com/blog/writing-your-first-kubectl-plugin/) blog post and associated [GH repo](https://github.com/bmuschko/kubectl-server-version). This was invaluable in getting me up and running.

## FAQ

**Why not just write a simple shell wrapper?**

For many reasons. First, I had that, and it's quite slow. You end up doing more queries than you need, and if your API servers are on the other side of privatelinks, it can get quite slow.

Second, I wanted a good excuse to learn some golang and get better at Kube internals.

**Aren't there already other plugins that do this?**

Not that I could find!

**Can I contribute?**

Sure, send a pull request!
