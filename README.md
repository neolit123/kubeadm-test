# kubeadm-test

A repository for testing kubeadm build tools and automation.

## Contents

### /Makefile

A Makefile with generic build / release / clean / other functionality.

### /app

A test application to build and release.

### /hack

Scripts that are executed from CI or from the `/Makefile`

### /.github/workflows

A set of GitHub workflows that are automatically executed, but can also
be executed on demand using a GitHub Personal Access Token.

### /k8s-repo-tools

A collection of tools for automatic repository synchronization,
branch fast-forward and creating releases.
