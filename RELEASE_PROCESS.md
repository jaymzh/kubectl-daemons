# Releasing kubectl-daemons (WIP)

Build a scratch release with:

```
goreleaser release --skip-validate --snapshot --clean
```

Or for a real release, make a tag and releaser will release
that tag.

```
version="0.0.1"
git tag -a v${version?} -m "version ${version?}" -s
git push origin --tags
goreleaser release
```
