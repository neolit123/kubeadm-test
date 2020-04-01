## Workflows

### k8s-repo-ff.yaml

This is a chron job that runs the `k8s-repo-ff` tool on a daily schedule.
The `k8s-repo-ff` tool itself determines if it is appropriate to perform
a fast-forward on the the latest release branch based on the latest
tag that is in the branch.

### k8s-repo-sync.yaml

This is a chron job that should run on a higher frequency to catch
the release of new tags and branches in a "source" repository
(e.g. kubernetes/kubernetes).

It runs the `k8s-repo-sync` tool that outputs an `output.json` file.
The file contains all new branches and tags that were synchronized.

`output.json` is passed to the `/hack/ci-dispatch-release.sh` script
which extracts all tags from it and sends this same repository a
dispatch event `dispatch-k8s-create-release` for every new tag,
with the idea to create new releases from the tags.

Note that the default workflow token `secrets.GITHUB_TOKEN` cannot
be used for repository dispatch. Therefore the repository must
include a secret called `secrets.RELEASE_TOKEN` that is a Personal
Access Token of a repository maintainer.

The `dispatch-k8s-create-release` event is handled by the workflow
`k8s-create-release.yaml`.

### k8s-create-release.yaml

This workflow only triggers on a `dispatch-k8s-create-release`
event. The event can normally also trigger on tag `create|push` events
by a maintainer, however tags created by a GitHub workflow using the
default token `secrets.GITHUB_TOKEN` will not trigger this workflow.
This is a hard coded limitation of GitHub workflows.

The workflow requires a `release_tag` as part of the dispatch payload
as the target tag to create a release for. It executes the script
`/hack/ci-release.sh` which manages the creation of releases from CI.
