# Go Module Path Migration — Design

**Date:** 2026-07-15
**Repo:** `io41/terraform-j2md`
**Status:** Approved scope

## Goal

Make `go install github.com/io41/terraform-j2md/cmd/terraform-j2md@latest`
build the fork, including its Terraform drift support.

## Changes

- Change the module path in `go.mod` from `github.com/reproio/terraform-j2md`
  to `github.com/io41/terraform-j2md`.
- Update every Go import that uses the old module path.
- Update the README `go install` command to use the fork path.
- Publish the next fork release as `v0.0.9-io41.2` after the change lands on
  `release` and CI passes.

## Non-goals

- Do not rewrite historical design or implementation-plan references to the
  upstream repository.
- Do not change README badges or the GitHub Action download URL; those are
  separate fork-branding and distribution concerns.
- Do not change program behavior, dependencies, or the drift feature.

## Verification

- No Go source or `go.mod` reference retains the old module path.
- `go test ./...`, `go build ./cmd/terraform-j2md`, `goreleaser check`, and
  workflow linting pass.
- The merged `release` commit passes the required GitHub Actions build.
- Pushing `v0.0.9-io41.2` creates a non-prerelease GitHub Release whose binary
  reports `0.0.9-io41.2` and whose module is installable through the fork path.
