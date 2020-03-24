## Description

"k8s-latest-version" is a tool for obtaining the latest SemVer
from a list of tags.

## Usage

Example usage:

```bash
git tag | k8s-latest-version -branch=release-1.17 -branch-prefix=release-
```

- See `-help` for all available options.
- The command only accepts input through STDIN.
- Passing a `-branch` such as `release-1.17` would mean obtaining the latest
`v1.17*` tag, as long as `-branch-prefix` is equal to `release-`.
- Not passing a branch means taking the latest tag from the whole list.
- The result goes to STDOUT but the command also writes extra details to STDERR.
