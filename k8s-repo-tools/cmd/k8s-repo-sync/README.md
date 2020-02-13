## Description

"k8s-repo-sync" is a tool for synchronizing tags and branches
between GitHub repositories.

## Usage

Example usage:

```bash
k8s-repo-sync -dest=kubernetes/kubeadm -source=kubernetes/kubernetes \
  -token <TOKEN> -min-version=v1.17.0 -output=output.json
```

- See `-help` for all available options.
- `-token` must hold a valid GitHub Personal Access Token.
- `-min-version` is required to filter branches and tags older than this version.
- The tool assumes that branches are versioned and formated like `<prefix>[v]MAJOR.MINOR`.
The prefix value can be controlled with the `-branch-prefix` flag.
- DRY-RUN mode for repositories is enabled by default. To disable it pass `-dry-run=false`.
- `-output` writes a JSON file with the tags and branches that were written to
- The `-output` file can still be written in DRY-RUN mode.
the destination repository.

## Creating a GitHub PAT (Personal Access Token)

To obtain a PAT follow this guide:
https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line

The token must have write access for creating tags and branches for the destination directory.
Clicking all the `repo` options should suffice.

## How it works

<!--
https://textart.io/sequence

user->client: pass parameters
source->client: GET branches & tags
dest->client: GET branches & tags
note left of client: determine diff
client->dest: POST new branches
dest->client: GET updated branches
client->dest: POST new tags
dest->client: GET updated tags
note left of client: write diff to disk
-->

```
+-------+              +---------+                +---------+ +-------+
| user  |              | client  |                | source  | | dest  |
+-------+              +---------+                +---------+ +-------+
    |                       |                          |          |
    | pass parameters       |                          |          |
    |---------------------->|                          |          |
    |                       |                          |          |
    |                       |      GET branches & tags |          |
    |                       |<-------------------------|          |
    |                       |                          |          |
    |                       |                 GET branches & tags |
    |                       |<------------------------------------|
    |    -----------------\ |                          |          |
    |    | determine diff |-|                          |          |
    |    |----------------| |                          |          |
    |                       |                          |          |
    |                       | POST new branches        |          |
    |                       |------------------------------------>|
    |                       |                          |          |
    |                       |                GET updated branches |
    |                       |<------------------------------------|
    |                       |                          |          |
    |                       | POST new tags            |          |
    |                       |------------------------------------>|
    |                       |                          |          |
    |                       |                    GET updated tags |
    |                       |<------------------------------------|
    |---------------------\ |                          |          |
    || write diff to disk |-|                          |          |
    ||--------------------| |                          |          |
    |                       |                          |          |
```

## The output format

The output format uses the [go-github](https://github.com/google/go-github)
`Reference` object to enumerate tags and branches as Git "refs".

Example output:

```json
[
  {
    "ref":"refs/heads/release-1.17",
    "url":"https://api.github.com/repos/kubernetes/kubernetes/git/refs/heads/release-1.17",
    "object":{
      "type":"commit",
      "sha":"b04b9fb3987b12045ac5b2b273f1b5b3a8a7c972",
      "url":"https://api.github.com/repos/kubernetes/kubernetes/git/commits/b04b9fb3987b12045ac5b2b273f1b5b3a8a7c972"
    },
    "node_id":"MDM6UmVmMjA1ODA0OTg6cmVsZWFzZS0xLjE3"
  },
  {
    "ref":"refs/tags/v1.17.0",
    "url":"https://api.github.com/repos/kubernetes/kubernetes/git/refs/tags/v1.17.0",
    "object":{
      "type":"tag",
      "sha":"02a9c9f39a18ee40c37835c36c7c80e0797b0d85",
      "url":"https://api.github.com/repos/kubernetes/kubernetes/git/tags/02a9c9f39a18ee40c37835c36c7c80e0797b0d85"
    },
    "node_id":"MDM6UmVmMjA1ODA0OTg6djEuMTcuMA=="
  },
]
```
