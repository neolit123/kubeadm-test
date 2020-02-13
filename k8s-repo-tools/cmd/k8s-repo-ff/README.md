## Description

"k8s-repo-ff" is a tool for fast-forwarding a release branch to
the master branch of a GitHub repository.

## Usage

Example usage:

```bash
k8s-repo-sync -dest=kubernetes/kubeadm -source=kubernetes/kubernetes \
  -token <TOKEN> -min-version=v1.17.0 -output=output.json
```

- See `-help` for all available options.
- `-token` must hold a valid GitHub Personal Access Token.
- The tool assumes that branches are versioned and formated like `<prefix>[v]MAJOR.MINOR`.
The prefix value can be controlled with the `-branch-prefix` flag.
- DRY-RUN mode for repositories is enabled by default. To disable it pass `-dry-run=false`.
- `-output` writes a JSON file with the resulted merge commit and the reference for the release branch.
- The `-output` file can still be written in DRY-RUN mode.

## Creating a GitHub PAT (Personal Access Token)

To obtain a PAT follow this guide:
https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line

The token must have write access for creating tags and branches for the destination directory.
Clicking all the `repo` options should suffice.

## How it works

<!--
https://textart.io/sequence

user->client: pass parameters
dest->client: GET branches & tags
note left of client: find latest versioned branch
note left of client: find latest versioned tag for that branch
note left of client: determine if appropriate to FF
dest->client: GET compare release branch to master
client->dest: POST merge master into the release branch
note left of client: write the merge commit to disk
-->

```
+-------+                                     +---------+                                      +-------+
| user  |                                     | client  |                                      | dest  |
+-------+                                     +---------+                                      +-------+
    |                                              |                                               |
    | pass parameters                              |                                               |
    |--------------------------------------------->|                                               |
    |                                              |                                               |
    |                                              |                           GET branches & tags |
    |                                              |<----------------------------------------------|
    |             -------------------------------\ |                                               |
    |             | find latest versioned branch |-|                                               |
    |             |------------------------------| |                                               |
    |--------------------------------------------\ |                                               |
    || find latest versioned tag for that branch |-|                                               |
    ||-------------------------------------------| |                                               |
    |           ---------------------------------\ |                                               |
    |           | determine if appropriate to FF |-|                                               |
    |           |--------------------------------| |                                               |
    |                                              |                                               |
    |                                              |          GET compare release branch to master |
    |                                              |<----------------------------------------------|
    |                                              |                                               |
    |                                              | POST merge master into the release branch     |
    |                                              |---------------------------------------------->|
    |           ---------------------------------\ |                                               |
    |           | write the merge commit to disk |-|                                               |
    |           |--------------------------------| |                                               |
    |                                              |                                               |
```

## The output format

The output format is JSON and consists of:
- a non-fatal `outputError` that did not cause an exit status != 0.
If `outputError` is not `null`, the rest of the fields could be empty.
- a merge-`commit` that is a [go-github](https://github.com/google/go-github) `RepositoryCommit`.
- a `reference` (branch) that is a [go-github](https://github.com/google/go-github) `Reference`
where the merge commit was created.

Example output:

```json
{
  "outputError": "some-non-fatal-error",
  "reference": {
    "ref":"refs/heads/release-1.17",
    "url":"https://api.github.com/repos/kubernetes/kubernetes/git/refs/heads/release-1.17",
    "object":{
      "type":"commit",
      "sha":"b04b9fb3987b12045ac5b2b273f1b5b3a8a7c972",
      "url":"https://api.github.com/repos/kubernetes/kubernetes/git/commits/b04b9fb3987b12045ac5b2b273f1b5b3a8a7c972"
    },
    "node_id":"MDM6UmVmMjA1ODA0OTg6cmVsZWFzZS0xLjE3"
  },
  "commit": {
    "sha": "02a9c9f39a18ee40c37835c36c7c80e0797b0d85"
  }
}
```
