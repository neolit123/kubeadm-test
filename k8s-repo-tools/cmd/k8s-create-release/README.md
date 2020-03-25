## Description

"k8s-create-release" is a tool for creating a GitHub release
with artifacts and release notes for Kubernetes based projects.

## Usage

Example usage:

```bash
k8s-create-release -dest=org/repo -token=<token> -release-tag=<tag> <options>
```

- See `-help` for all available options.
- `-token` must hold a valid GitHub Personal Access Token.
- `-release-tag` must be an existing SemVer tag in the `-dest` GitHub repository.
- The tool assumes that branches are versioned and formated like `<prefix>[v]MAJOR.MINOR`.
The prefix value can be controlled with the `-branch-prefix` flag.
- DRY-RUN mode for repositories is enabled by default. To disable it pass `-dry-run=false`.
- The flag `-build-command` can be used to trigger a build of a target application,
for example `-build-command "make -f somepath release"`
- The flag `-release-asset` can be used to upload artifacts to a GitHub release.
Its format is `-release-asset name=path`. Multiple instances of the flag are allowed.
- If `-release-notes-path` is used it will take priority over `-release-notes-tool-path`.

## Creating a GitHub PAT (Personal Access Token)

To obtain a PAT follow this guide:
https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line

The token must have write access for creating tags and branches for the destination directory.
Clicking all the `repo` options should suffice.

## How it works

<!--
https://textart.io/sequence

user->client: pass parameters
note left of client: check if release notes are ready
client->dest: GET the ref of the target tag
client->dest: GET all tags
note left of client: find end tag for release notes
client->releasenotestool: generate release notes
releasenotestool->client: read release notes
client->dest: POST create release for this tag
client->dest: POST upload release assets
-->

```
+-------+                            +---------+                             +-------+ +-------------------+
| user  |                            | client  |                             | dest  | | releasenotestool  |
+-------+                            +---------+                             +-------+ +-------------------+
    |                                     |                                      |               |
    | pass parameters                     |                                      |               |
    |------------------------------------>|                                      |               |
    |-----------------------------------\ |                                      |               |
    || check if release notes are ready |-|                                      |               |
    ||----------------------------------| |                                      |               |
    |                                     |                                      |               |
    |                                     | GET the ref of the target tag        |               |
    |                                     |------------------------------------->|               |
    |                                     |                                      |               |
    |                                     | GET all tags                         |               |
    |                                     |------------------------------------->|               |
    |  ---------------------------------\ |                                      |               |
    |  | find end tag for release notes |-|                                      |               |
    |  |--------------------------------| |                                      |               |
    |                                     |                                      |               |
    |                                     | generate release notes               |               |
    |                                     |----------------------------------------------------->|
    |                                     |                                      |               |
    |                                     |                                   read release notes |
    |                                     |<-----------------------------------------------------|
    |                                     |                                      |               |
    |                                     | POST create release for this tag     |               |
    |                                     |------------------------------------->|               |
    |                                     |                                      |               |
    |                                     | POST upload release assets           |               |
    |                                     |------------------------------------->|               |
    |                                     |                                      |               |
```
