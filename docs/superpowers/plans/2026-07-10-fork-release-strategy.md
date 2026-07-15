# Fork Release Strategy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish `release` as io41's protected fork line, publish collision-free GoReleaser binaries, and preserve `master` as a fast-forward-only upstream mirror.

**Architecture:** Bootstrap `release` from the synchronized upstream mirror, merge the existing drift branch without changing it, then apply the smallest CI, GoReleaser, and CLI changes directly during the one-time unprotected setup. Push only after local validation, wait for the first successful `Build` check, then make `release` the default branch and enable branch/tag rulesets before creating the first fork tag.

**Tech Stack:** Git, GitHub CLI/API and repository rulesets, GitHub Actions, Go 1.22 standard library (`flag`, `os/exec`, `testing`), GoReleaser v2, `yq`.

---

## Scope and current state

This is one integrated release-bootstrap project: branch topology, CI, version stamping, repository protection, and the first release depend on each other and should not be split into separate plans.

State verified on 2026-07-10:

- `origin/master` and `reproio/master` point to `8b52702`; local `master` is four commits behind.
- `support-terraform-drift-detection` points to `bd094b0`, exactly two commits ahead of `origin/master`; upstream PR `reproio#46` is open.
- `release`, fork tags, fork releases, rulesets, and check histories do not exist.
- GitHub's default branch is `master`; repository `GITHUB_TOKEN` defaults are `write` and PR approval is enabled.
- Fork workflow approval already equals `first_time_contributors`.
- Latest upstream release tag is `v0.0.9`; first fork tag is therefore `v0.0.9-io41.1`.

One required compatibility correction extends the approved design: GoReleaser 2.17 rejects the current version-0 config and two deprecated keys before it can release anything. Task 2 adds `version: 2` and performs only the two required key renames alongside `release.prerelease: false`.

## File map

- Modify `.github/workflows/go.yml`: run the existing `Build` job for pushes and pull requests targeting both long-lived branches.
- Modify `.github/workflows/goreleaser.yml`: trigger only for numeric `vX.Y.Z-io41.N` tags; retain the job-scoped `contents: write` permission.
- Modify `.goreleaser.yml`: opt into schema version 2, replace two rejected deprecated properties, and force fork releases to be normal/latest releases.
- Modify `cmd/terraform-j2md/main.go`: expose GoReleaser's existing `main.Version` and `main.Revision` linker values through `--version` without disturbing drift flags.
- Create `cmd/terraform-j2md/main_test.go`: one subprocess test proves linker injection, output, exit status, and early exit without stdin.
- No new runbook, release script, dependency, module-path change, signing, package manager, or scheduled sync. The approved design document remains the manual upstream-sync runbook.

## References

- Design: `docs/superpowers/specs/2026-07-08-fork-release-strategy-design.md`
- [GitHub workflow filter syntax](https://docs.github.com/en/actions/writing-workflows/workflow-syntax-for-github-actions)
- [GitHub ruleset creation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/creating-rulesets-for-a-repository)
- [GitHub ruleset rules](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets)
- [GitHub Actions permissions API](https://docs.github.com/en/rest/actions/permissions?apiVersion=2026-03-10)
- [GoReleaser v2 configuration errors](https://www.goreleaser.com/resources/errors/version/)
- [GoReleaser release settings](https://goreleaser.com/customization/publish/scm/)

### Task 1: Bootstrap the local fork line

**Files:**
- Commit: `docs/superpowers/plans/2026-07-10-fork-release-strategy.md`
- Preserve unchanged: `support-terraform-drift-detection`
- Create branches: local `release`
- Add remote: `upstream`

- [ ] **Step 1: Verify identity, repository, branch, and clean input state**

Run:

```bash
gh auth status --active
gh repo view io41/terraform-j2md --json defaultBranchRef,viewerPermission,isFork,parent
git status --short --branch
git remote -v
git rev-parse support-terraform-drift-detection
gh api repos/reproio/terraform-j2md/pulls/46 --jq '{state,head:.head.sha,url:.html_url}'
```

Expected:

- `viewerPermission` is `ADMIN`, `isFork` is `true`, and default branch is `master`.
- Current branch is `docs/fork-release-strategy`; only this plan may be uncommitted.
- `origin` is `git@github.com:io41/terraform-j2md.git`; no `upstream` remote or `release` branch exists.
- Local drift branch and PR head both equal `bd094b0596959ecdf561e0dabe8582326f5976f1`; PR state is `open`.

Stop on any mismatch. Do not force-update any ref.

- [ ] **Step 2: Commit the implementation plan before switching branches**

Run:

```bash
git add docs/superpowers/plans/2026-07-10-fork-release-strategy.md
git commit -m "docs: add fork release implementation plan"
git status --short
```

Expected: commit succeeds and final command prints nothing.

- [ ] **Step 3: Add a fetch-only upstream remote and fetch all refs/tags**

Run:

```bash
git remote add upstream https://github.com/reproio/terraform-j2md.git
git remote set-url --push upstream DISABLED
git fetch --prune origin
git fetch --prune --tags upstream
git remote -v
test "$(git rev-parse origin/master)" = "$(git rev-parse upstream/master)"
```

Expected: `upstream` fetch URL is GitHub HTTPS, push URL is `DISABLED`, and the final comparison exits 0. If it fails, stop: `origin/master` is not a pristine mirror.

- [ ] **Step 4: Fast-forward local `master` without rewriting it**

Run:

```bash
git switch master
git merge --ff-only origin/master
git branch --set-upstream-to=origin/master master
test "$(git rev-parse master)" = "$(git rev-parse upstream/master)"
```

Expected: `master` fast-forwards to `8b52702` (or a newer identical origin/upstream tip), then the comparison exits 0. If `--ff-only` fails, stop and inspect divergence; never force-push.

- [ ] **Step 5: Create `release` and merge the unchanged drift branch**

Run:

```bash
git branch release origin/master
git switch release
git merge --no-ff support-terraform-drift-detection -m "Merge drift detection into fork release line"
test "$(git rev-parse support-terraform-drift-detection)" = "bd094b0596959ecdf561e0dabe8582326f5976f1"
git merge-base --is-ancestor support-terraform-drift-detection release
```

Expected: clean merge commit; both tests exit 0. On conflict, run `git merge --abort` and stop rather than changing the PR branch.

- [ ] **Step 6: Merge the approved design and plan into `release`**

Run:

```bash
git merge --no-ff docs/fork-release-strategy -m "Merge fork release strategy docs"
git merge-base --is-ancestor docs/fork-release-strategy release
git log --graph --decorate --oneline -10
```

Expected: clean merge; the history shows separate drift and documentation merges on `release`. Keep `release` local until Tasks 2-4 pass so its first push starts CI with the corrected workflow.

### Task 2: Prepare CI and GoReleaser for fork releases

**Files:**
- Modify: `.github/workflows/go.yml:5`
- Modify: `.github/workflows/go.yml:7`
- Modify: `.github/workflows/goreleaser.yml:6`
- Modify: `.goreleaser.yml:1`
- Modify: `.goreleaser.yml:28`
- Modify: `.goreleaser.yml:30`
- Modify: `.goreleaser.yml:34`

- [ ] **Step 1: Run focused baseline checks and confirm they fail**

Run:

```bash
yq -e '.on.push.branches == ["release", "master"] and .on.pull_request.branches == ["release", "master"]' .github/workflows/go.yml
yq -e '.on.push.tags == ["v[0-9]+.[0-9]+.[0-9]+-io41.[0-9]+"]' .github/workflows/goreleaser.yml
goreleaser check
```

Expected:

- Both `yq` assertions exit 1 because current filters cover only upstream names.
- `goreleaser check` exits nonzero and reports configuration version `0`, `snapshot.name_template`, and `archives.format_overrides.format`.

These config-only one-line changes do not justify a new test framework; the same executable assertions become the acceptance check.

- [ ] **Step 2: Add `release` to both CI branch filters**

Replace the trigger in `.github/workflows/go.yml` with:

```yaml
on:
  push:
    branches: [ release, master ]
  pull_request:
    branches: [ release, master ]
```

Leave the existing `Build` job unchanged.

- [ ] **Step 3: Restrict the release workflow to fork tags**

Replace the trigger in `.github/workflows/goreleaser.yml` with:

```yaml
on:
  push:
    tags:
      - "v[0-9]+.[0-9]+.[0-9]+-io41.[0-9]+"
```

Keep this existing job-level permission unchanged:

```yaml
permissions:
  contents: write
```

- [ ] **Step 4: Make the smallest valid GoReleaser v2 update**

Apply exactly these changes in `.goreleaser.yml`:

```diff
+version: 2
 project_name: terraform-j2md
@@
     format_overrides:
       - goos: Windows
-        format: zip
+        formats: [zip]
 release:
-  prerelease: auto
+  prerelease: false
@@
 snapshot:
-  name_template: "{{ .Tag }}-next"
+  version_template: "{{ .Version }}-next"
```

Do not migrate unrelated workflow actions or change artifact names.

- [ ] **Step 5: Validate workflow values and GoReleaser schema**

Run:

```bash
yq -e '.on.push.branches == ["release", "master"] and .on.pull_request.branches == ["release", "master"]' .github/workflows/go.yml
yq -e '.on.push.tags == ["v[0-9]+.[0-9]+.[0-9]+-io41.[0-9]+"]' .github/workflows/goreleaser.yml
yq -e '.version == 2 and .release.prerelease == false and .snapshot.version_template == "{{ .Version }}-next" and .archives[0].format_overrides[0].formats == ["zip"]' .goreleaser.yml
goreleaser check
git diff --check
```

Expected: all commands exit 0; GoReleaser reports `1 configuration file(s) validated`.

- [ ] **Step 6: Commit the release configuration atomically**

Run:

```bash
git add .github/workflows/go.yml .github/workflows/goreleaser.yml .goreleaser.yml
git commit -m "ci: prepare fork release workflows"
```

Expected: one commit containing only the three release/config files.

### Task 3: Expose linker-stamped version information

**Files:**
- Create: `cmd/terraform-j2md/main_test.go`
- Modify: `cmd/terraform-j2md/main.go:11`
- Modify: `cmd/terraform-j2md/main.go:16`

- [ ] **Step 1: Write the failing subprocess test**

Create `cmd/terraform-j2md/main_test.go`:

```go
package main

import (
	"os/exec"
	"testing"
)

func TestVersion(t *testing.T) {
	cmd := exec.Command(
		"go", "run",
		"-ldflags=-X main.Version=0.0.9-io41.1 -X main.Revision=abc1234",
		".", "--version",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run --version: %v\n%s", err, output)
	}
	if got, want := string(output), "0.0.9-io41.1 (abc1234)\n"; got != want {
		t.Errorf("--version = %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run the test and verify the missing flag fails**

Run:

```bash
go test ./cmd/terraform-j2md -run '^TestVersion$' -count=1 -v
```

Expected: FAIL; subprocess output includes `flag provided but not defined: -version` and `exit status 2`.

- [ ] **Step 3: Add linker variables and early `--version` exit**

After the drift merge from Task 1, replace `cmd/terraform-j2md/main.go` with:

```go
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/reproio/terraform-j2md/internal/terraform"
)

var (
	Version, Revision string
	escapeHTML        = true
	showDrift         = true
)

func main() {
	noEscapeHTML := flag.Bool("no-escape-html", false, "prevent <, >, and & from being escaped in JSON strings")
	noDrift := flag.Bool("no-drift", false, "omit the drift detection section from the output")
	showVersion := flag.Bool("version", false, "print version")
	flag.Parse()
	if *showVersion {
		fmt.Printf("%s (%s)\n", Version, Revision)
		return
	}
	if *noEscapeHTML {
		escapeHTML = false
	}
	if *noDrift {
		showDrift = false
	}
	os.Exit(run())
}

func run() int {
	planData, err := terraform.NewPlanData(os.Stdin, escapeHTML, showDrift)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot parse input as Terraform plan JSON: %v", err)
		return 1
	}
	if err = planData.Render(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "cannot render: %v", err)
		return 1
	}
	return 0
}
```

The early return is required: `--version` must not parse Terraform JSON from stdin. Empty development values remain allowed by the approved design; GoReleaser supplies both values for releases.

- [ ] **Step 4: Format and run the focused test**

Run:

```bash
gofmt -w cmd/terraform-j2md/main.go cmd/terraform-j2md/main_test.go
go test ./cmd/terraform-j2md -run '^TestVersion$' -count=1 -v
```

Expected: PASS.

- [ ] **Step 5: Verify the exact release-style output and all Go tests**

Run:

```bash
go run -ldflags='-X main.Version=0.0.9-io41.1 -X main.Revision=abc1234' ./cmd/terraform-j2md --version
make test
git diff --check
```

Expected version output:

```text
0.0.9-io41.1 (abc1234)
```

Expected: all tests pass and `git diff --check` prints nothing.

- [ ] **Step 6: Commit code and test together**

Run:

```bash
git add cmd/terraform-j2md/main.go cmd/terraform-j2md/main_test.go
git commit -m "feat: expose build version"
```

Expected: one commit containing only CLI version behavior and its test.

### Task 4: Validate and publish the bootstrapped `release` branch

**Files:**
- Verify: all tracked files
- Generate ignored artifacts: `dist/`
- Push branch: `origin/release`

- [ ] **Step 1: Run the complete local gate**

Run:

```bash
make test
make build
goreleaser check
goreleaser release --snapshot --clean
git diff --exit-code -- go.mod go.sum
git diff --check
git status --short
```

Expected:

- Tests, ordinary build, GoReleaser validation, and snapshot build pass.
- `go.mod` and `go.sum` remain unchanged after GoReleaser's `go mod tidy` hook.
- `dist/` is ignored; final status is empty.

- [ ] **Step 2: Verify topology and the untouched upstream PR branch**

Run:

```bash
test "$(git branch --show-current)" = "release"
test "$(git rev-parse support-terraform-drift-detection)" = "bd094b0596959ecdf561e0dabe8582326f5976f1"
git merge-base --is-ancestor origin/master release
git merge-base --is-ancestor support-terraform-drift-detection release
gh api repos/reproio/terraform-j2md/pulls/46 --jq '{state,head:.head.sha,url:.html_url}'
git log --graph --decorate --oneline -12
```

Expected: all Git tests exit 0; PR remains open with head `bd094b0596959ecdf561e0dabe8582326f5976f1`.

- [ ] **Step 3: Push `release` once, after it contains working CI**

Run:

```bash
git push --set-upstream origin release
```

Expected: new remote branch `origin/release`; no force update.

- [ ] **Step 4: Wait for the first `release` build**

Run after GitHub indexes the push:

```bash
RUN_ID="$(gh run list --repo io41/terraform-j2md --workflow go.yml --branch release --event push --limit 1 --json databaseId --jq '.[0].databaseId')"
test -n "$RUN_ID"
gh run watch "$RUN_ID" --repo io41/terraform-j2md --exit-status
gh api repos/io41/terraform-j2md/commits/release/check-runs --jq '.check_runs[] | {name,status,conclusion,app:.app.slug}'
```

Expected: `Build` completes with `success` from `github-actions`. If lookup returns an empty ID, rerun only the lookup after GitHub finishes indexing. Stop on a failed run; do not enable protection around a failing branch.

### Task 5: Lock down GitHub and switch the default branch

**Files:**
- No repository file changes
- Modify GitHub repository settings for `io41/terraform-j2md`

- [ ] **Step 1: Reduce default `GITHUB_TOKEN` privileges**

This mutates repository Actions settings and requires admin rights.

Run:

```bash
gh api --method PUT -H "X-GitHub-Api-Version: 2026-03-10" repos/io41/terraform-j2md/actions/permissions/workflow -f default_workflow_permissions=read -F can_approve_pull_request_reviews=false
gh api repos/io41/terraform-j2md/actions/permissions/workflow
```

Expected response:

```json
{"default_workflow_permissions":"read","can_approve_pull_request_reviews":false}
```

The release job still has explicit, narrowly scoped `contents: write`.

- [ ] **Step 2: Verify fork approval and workflow trust boundaries**

Run:

```bash
gh api repos/io41/terraform-j2md/actions/permissions/fork-pr-contributor-approval
rg -n 'pull_request_target' .github/workflows
rg -n 'secrets\.' .github/workflows
```

Expected:

- Approval API returns `{"approval_policy":"first_time_contributors"}`.
- `pull_request_target` search has no matches and exits 1.
- Secret search finds only `GITHUB_TOKEN` in `.github/workflows/goreleaser.yml`; build/test workflow uses no secrets.

- [ ] **Step 3: Make `release` the GitHub default and update local remote HEAD**

Run:

```bash
gh repo edit io41/terraform-j2md --default-branch release
git remote set-head origin --auto
gh repo view io41/terraform-j2md --json defaultBranchRef --jq .defaultBranchRef.name
git ls-remote --symref origin HEAD
```

Expected: both GitHub and `origin/HEAD` resolve to `release`.

- [ ] **Step 4: Create the branch ruleset in GitHub's repository UI**

Open `https://github.com/io41/terraform-j2md/settings/rules`, choose **New ruleset → New branch ruleset**, and set exactly:

- Ruleset name: `Protect release and mirror`
- Enforcement status: `Active`
- Bypass list: user `io41`, mode `Always allow`
- Target branches: include `release` and `master`
- Enable `Restrict updates`
- Enable `Restrict deletions`
- Enable `Require a pull request before merging`; leave optional review subsettings unchanged
- Enable `Require status checks to pass`; add check `Build` from GitHub Actions
- Enable `Block force pushes`
- Leave `Require linear history` disabled because both fork integration and upstream sync use merge commits

Save the ruleset. `Always allow` is intentional: the approved upstream-sync runbook directly fast-forwards `master` and merges it into `release`; only repository owner `io41` currently has that bypass.

- [ ] **Step 5: Create the release-tag ruleset in GitHub's repository UI**

On the same settings page, choose **New ruleset → New tag ruleset**, and set exactly:

- Ruleset name: `Protect release tags`
- Enforcement status: `Active`
- Bypass list: user `io41`, mode `Always allow`
- Target tags: include pattern `v*`
- Enable `Restrict creations`
- Enable `Restrict updates`
- Enable `Restrict deletions`

Save the ruleset. Only the bypassed repository owner can create, move, or delete release tags.

- [ ] **Step 6: Verify all GitHub controls**

Run:

```bash
gh ruleset list --repo io41/terraform-j2md --no-parents
gh ruleset check release --repo io41/terraform-j2md
gh ruleset check master --repo io41/terraform-j2md
gh api repos/io41/terraform-j2md/rulesets --jq '.[] | {id,name,target,enforcement,conditions}'
gh api repos/io41/terraform-j2md/actions/permissions/workflow
gh api repos/io41/terraform-j2md/actions/permissions/fork-pr-contributor-approval
git ls-remote --heads origin release master support-terraform-drift-detection
gh api repos/reproio/terraform-j2md/pulls/46 --jq '{state,head:.head.sha,url:.html_url}'
```

Expected:

- Two active rulesets exist with the exact names above; branch checks show their applicable rules.
- Actions defaults are `read` and cannot approve PRs; fork approval is `first_time_contributors`.
- All three remote branches exist; upstream PR is still open at `bd094b0`.

### Task 6: Cut and verify the first fork release

**Files:**
- No tracked file changes
- Create Git tag and GitHub Release: `v0.0.9-io41.1`

- [ ] **Step 1: Prove the upstream base and release commit are correct**

Run:

```bash
git fetch --prune --tags upstream
git fetch origin
UPSTREAM_BASE="$(git tag --merged upstream/master --list 'v[0-9]*' --sort=-version:refname | head -1)"
test "$UPSTREAM_BASE" = "v0.0.9"
git switch release
git merge --ff-only origin/release
test -z "$(git status --porcelain)"
test "$(git rev-parse release)" = "$(git rev-parse origin/release)"
make test
```

Expected: base comparison, cleanliness, remote-tip comparison, and tests all pass. If upstream base has moved, stop and revise this plan with the actual upstream tag before creating any fork tag.

- [ ] **Step 2: Prove the tag is unused locally and remotely**

Run:

```bash
VERSION=v0.0.9-io41.1
test -z "$(git tag --list "$VERSION")"
test -z "$(git ls-remote --tags origin "refs/tags/$VERSION")"
```

Expected: both checks exit 0 and print nothing.

- [ ] **Step 3: Create and push the protected release tag**

Run:

```bash
VERSION=v0.0.9-io41.1
git tag "$VERSION"
git show --no-patch --decorate "$VERSION"
git push origin "refs/tags/$VERSION"
```

Expected: tag points at the exact `origin/release` tip and push succeeds through `io41`'s tag-ruleset bypass.

- [ ] **Step 4: Wait for GoReleaser**

Run after GitHub indexes the tag push:

```bash
VERSION=v0.0.9-io41.1
RUN_ID="$(gh run list --repo io41/terraform-j2md --workflow goreleaser.yml --branch "$VERSION" --event push --limit 1 --json databaseId --jq '.[0].databaseId')"
test -n "$RUN_ID"
gh run watch "$RUN_ID" --repo io41/terraform-j2md --exit-status
```

Expected: GoReleaser workflow completes successfully. If lookup returns an empty ID, rerun only the lookup after indexing finishes.

**Failure rule:** once the tag push succeeds, never move, delete, or reuse that published tag. Fix `release`, increment the fork counter, and retry as `v0.0.9-io41.2`.

- [ ] **Step 5: Verify release metadata, Latest behavior, and assets**

Run:

```bash
VERSION=v0.0.9-io41.1
gh release view "$VERSION" --repo io41/terraform-j2md --json tagName,isDraft,isPrerelease,assets,url --jq '{tagName,isDraft,isPrerelease,assets:[.assets[].name],url}'
gh release view --repo io41/terraform-j2md --json tagName --jq .tagName
```

Expected:

- Tag is `v0.0.9-io41.1`.
- `isDraft` and `isPrerelease` are both `false`.
- Latest release is `v0.0.9-io41.1`.
- Assets include `checksums.txt` and platform archives.

- [ ] **Step 6: Verify checksum and released binary provenance**

Run on the current Darwin/arm64 host:

```bash
VERSION=v0.0.9-io41.1
VERIFY_DIR="$(mktemp -d)"
gh release download "$VERSION" --repo io41/terraform-j2md --pattern 'terraform-j2md_Darwin_arm64.tar.gz' --pattern checksums.txt --dir "$VERIFY_DIR"
grep 'terraform-j2md_Darwin_arm64.tar.gz' "$VERIFY_DIR/checksums.txt" | (cd "$VERIFY_DIR" && shasum -a 256 -c -)
tar -xzf "$VERIFY_DIR/terraform-j2md_Darwin_arm64.tar.gz" -C "$VERIFY_DIR"
EXPECTED_VERSION="0.0.9-io41.1 ($(git rev-parse --short "$VERSION"))"
test "$("$VERIFY_DIR/terraform-j2md" --version)" = "$EXPECTED_VERSION"
```

Expected: checksum reports `OK`; final exact version assertion exits 0.

## Ongoing upstream sync after bootstrap

Do not run this during bootstrap unless `upstream/master` has moved. Future syncs use the approved no-force sequence:

```bash
git fetch upstream
git switch master
git merge --ff-only upstream/master
git push origin master

git switch release
git merge master
git push origin release
```

Resolve conflicts only on `release`. Never commit directly to `master`, force-push either long-lived branch, rename the Go module, or modify `support-terraform-drift-detection` to maintain the fork.
