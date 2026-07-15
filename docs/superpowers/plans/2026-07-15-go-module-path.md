# Go Module Path Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `go install github.com/io41/terraform-j2md/cmd/terraform-j2md@latest` install the fork and publish it as `v0.0.9-io41.2`.

**Architecture:** Rename the Go module at its declaration and every internal import boundary, then update the single documented `go install` command. Land the change through the protected `release` branch and use the existing tag-triggered GoReleaser workflow.

**Tech Stack:** Go 1.22, GitHub Actions, GoReleaser v2

## Global Constraints

- Do not change dependencies or program behavior.
- Do not rewrite historical upstream references, README badges, or `action.yml`.
- Publish exactly `v0.0.9-io41.2`, only after `release` CI passes.
- Do not force-push or move an existing tag.

---

### Task 1: Rename the Go module

**Files:**
- Modify: `go.mod:1`
- Modify: `cmd/terraform-j2md/main.go:8`
- Modify: `internal/terraform/plan.go:7`
- Modify: `test/plan_test/plan_test.go:6`
- Modify: `test/format_json_test/format_json_test.go:5`
- Modify: `test/format_json_test/format_unknown_test.go:5`
- Modify: `README.md:17`

**Interfaces:**
- Consumes: existing packages beneath `internal/format` and `internal/terraform`
- Produces: module path `github.com/io41/terraform-j2md` and install command `github.com/io41/terraform-j2md/cmd/terraform-j2md@latest`

- [ ] **Step 1: Verify the current module still has the failing identity**

Run:

```bash
go list -m
```

Expected: `github.com/reproio/terraform-j2md`.

- [ ] **Step 2: Replace the module prefix in the seven scoped files**

Change the `go.mod` declaration to:

```go
module github.com/io41/terraform-j2md
```

Change each internal Go import prefix to:

```go
github.com/io41/terraform-j2md
```

Change the README command to:

```text
% go install github.com/io41/terraform-j2md/cmd/terraform-j2md@latest
```

- [ ] **Step 3: Verify the new module identity and absence of stale code references**

Run:

```bash
go list -m
rg -n 'github\.com/reproio/terraform-j2md' go.mod cmd internal test
```

Expected: `go list -m` prints `github.com/io41/terraform-j2md`; `rg` prints nothing and exits 1.

- [ ] **Step 4: Run local validation**

Run:

```bash
go test ./...
go build -o /tmp/terraform-j2md-build ./cmd/terraform-j2md
goreleaser check
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
git diff --check
```

Expected: every command exits 0.

- [ ] **Step 5: Commit the migration**

```bash
git add go.mod README.md cmd/terraform-j2md/main.go internal/terraform/plan.go test/plan_test/plan_test.go test/format_json_test/format_json_test.go test/format_json_test/format_unknown_test.go
git commit -m "fix: use fork Go module path"
```

### Task 2: Land the change on `release`

**Files:**
- No additional file changes

**Interfaces:**
- Consumes: tested commit from Task 1
- Produces: merged `release` commit with successful required `Build` check

- [ ] **Step 1: Push the feature branch and open a PR**

```bash
git push -u origin fix/io41-module-path
gh pr create --repo io41/terraform-j2md --base release --head fix/io41-module-path --title "fix: use fork Go module path" --body "Rename the Go module and internal imports so go install resolves the io41 fork."
```

Expected: GitHub returns the new PR URL.

- [ ] **Step 2: Wait for the required check**

```bash
gh pr checks --repo io41/terraform-j2md --watch
```

Expected: `Build` succeeds.

- [ ] **Step 3: Merge without bypassing branch protection**

```bash
gh pr merge --repo io41/terraform-j2md --merge --delete-branch
```

Expected: the PR reports merged and the remote feature branch is deleted.

- [ ] **Step 4: Update and verify local `release`**

```bash
git switch release
git pull --ff-only origin release
git status --short --branch
```

Expected: local `release` matches `origin/release` with a clean worktree.

### Task 3: Publish and verify `v0.0.9-io41.2`

**Files:**
- No file changes

**Interfaces:**
- Consumes: verified `release` commit from Task 2
- Produces: immutable tag and published GitHub Release `v0.0.9-io41.2`

- [ ] **Step 1: Confirm the tag is unused, then create and push it**

```bash
git ls-remote --tags origin refs/tags/v0.0.9-io41.2
git tag v0.0.9-io41.2
git push origin v0.0.9-io41.2
```

Expected: `ls-remote` prints nothing; push creates the tag.

- [ ] **Step 2: Wait for the GoReleaser workflow**

```bash
run_id=$(gh run list --repo io41/terraform-j2md --workflow goreleaser.yml --event push --limit 1 --json databaseId,headBranch --jq '.[] | select(.headBranch == "v0.0.9-io41.2") | .databaseId')
gh run watch "$run_id" --repo io41/terraform-j2md --exit-status
```

Expected: the run for `v0.0.9-io41.2` succeeds.

- [ ] **Step 3: Verify release metadata, assets, and binary provenance**

```bash
gh release view v0.0.9-io41.2 --repo io41/terraform-j2md --json tagName,isDraft,isPrerelease,isLatest,targetCommitish,assets,url
mkdir -p /tmp/terraform-j2md-v0.0.9-io41.2
gh release download v0.0.9-io41.2 --repo io41/terraform-j2md --pattern '*Darwin_arm64.tar.gz' --pattern 'checksums.txt' --dir /tmp/terraform-j2md-v0.0.9-io41.2
```

Expected: published, non-draft, non-prerelease, Latest, targeting the merged `release` commit, with release archives and checksums.

Extract the archive and run:

```bash
tar -xzf /tmp/terraform-j2md-v0.0.9-io41.2/terraform-j2md_Darwin_arm64.tar.gz -C /tmp/terraform-j2md-v0.0.9-io41.2
/tmp/terraform-j2md-v0.0.9-io41.2/terraform-j2md --version
```

Expected: `0.0.9-io41.2 (<short release commit>)`.

- [ ] **Step 4: Verify the public Go install path**

```bash
GOBIN=/tmp/terraform-j2md-go-install go install github.com/io41/terraform-j2md/cmd/terraform-j2md@latest
go version -m /tmp/terraform-j2md-go-install/terraform-j2md
```

Expected: install exits 0 and build metadata reports module
`github.com/io41/terraform-j2md` at `v0.0.9-io41.2`.
