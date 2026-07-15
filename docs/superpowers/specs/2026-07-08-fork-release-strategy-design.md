# Fork Release Strategy — Design

**Date:** 2026-07-08
**Repo:** `io41/terraform-j2md` (fork of `reproio/terraform-j2md`)
**Status:** Approved, ready for implementation plan

## Problem

Upstream (`reproio/terraform-j2md`) may or may not accept our drift-detection
contribution (PR `reproio#46`). We want to:

1. Keep the upstream PR branch open and unchanged.
2. Maintain our own long-lived fork that ships releases regardless of upstream's
   decision.
3. Pull upstream changes in over time without rewriting shared history.
4. Publish our own versioned binaries that never collide with upstream's
   versions.
5. Ensure no external contributor can trigger a release or reach repository
   secrets.

## Consumption model

Binaries downloaded from the fork's GitHub Releases (built by GoReleaser). Not
consumed as a GitHub Action ref, not via `go install`. This is why releases and
a collision-free version scheme matter.

## Branch topology

| Branch | Role | Rules |
|---|---|---|
| `master` | Pristine mirror of `reproio:master` | Never commit directly; fast-forward only from upstream |
| `release` | **Default** branch — the fork line (`master` + our patches) | All fork work lands here; releases are cut from here |
| `support-terraform-drift-detection` | Backing branch for upstream PR `reproio#46` → `reproio:master` | Left as-is; cross-repo; unaffected by this work |

Remotes:
- `origin` → `git@github.com:io41/terraform-j2md.git` (the fork)
- `upstream` → `https://github.com/reproio/terraform-j2md.git` (fetch-only) — to be added

## Bootstrap (one-time)

```sh
# release starts from the current upstream mirror
git branch release origin/master

# land the drift feature onto the fork line
git checkout release
git merge --no-ff support-terraform-drift-detection

git push -u origin release

# make release the default branch on GitHub
gh repo edit io41/terraform-j2md --default-branch release

# add the upstream remote for future syncs
git remote add upstream https://github.com/reproio/terraform-j2md.git
```

Note: if upstream later merges PR `reproio#46`, the equivalent commits arrive via
`master` on the next sync; the `git merge master` step below reconciles them
(expected to be a clean or trivial merge).

## Upstream-sync runbook (no force-push, ever)

```sh
git fetch upstream
git checkout master
git merge --ff-only upstream/master   # master stays pristine; fails loudly if not FF
git push origin master

git checkout release
git merge master                      # resolve any conflicts HERE only
git push origin release
```

## CI — `.github/workflows/go.yml` (build + test)

- Trigger on push + pull_request to `[release, master]` (currently `[master]`).
- **Fork-PR safety is built in:** a `pull_request` event from a fork runs with a
  read-only `GITHUB_TOKEN` and receives **no repository secrets** — GitHub
  withholds them. `go.yml` uses no secrets regardless, so there is nothing for an
  external PR to exfiltrate or abuse.

## Releases — `.github/workflows/goreleaser.yml` + `.goreleaser.yml`

### Version scheme

`v<upstream-base>-io41.<N>` — e.g. `v0.0.9-io41.1`.

- `<upstream-base>` = the latest upstream version the fork currently sits on.
- `<N>` = fork release counter; increment per fork release, reset to `1` when
  `<upstream-base>` moves.
- Collision-free with upstream's `v0.0.x` line; provenance is explicit.

### Trigger

Widen the tag filter so it fires **only** on fork tags, never on any upstream
`vX.Y.Z` tag that might exist in the fork:

```yaml
on:
  push:
    tags:
      - "v[0-9]+.[0-9]+.[0-9]+-io41.[0-9]+"
```

(Equivalent simpler glob: `"v*-io41.*"`.)

### GoReleaser config

Semver treats the `-io41.N` suffix as a prerelease, and `release.prerelease:
auto` would flag these GitHub Releases as "Pre-release" (so "Latest" never points
to them). Override it:

```yaml
release:
  prerelease: false
```

### Cutting a release

```sh
git checkout release
git tag v0.0.9-io41.1
git push origin v0.0.9-io41.1   # triggers goreleaser → binaries + GitHub Release on the fork
```

Only accounts with push access can push tags, so external contributors cannot
trigger a release.

## Version stamping — `--version` flag

`.goreleaser.yml` already injects `-X main.Version={{.Version}}` and
`-X main.Revision={{.ShortCommit}}`, but `cmd/terraform-j2md/main.go` declares no
such variables, so the linker flags are silently dropped and released binaries
cannot report their version.

Add to `main.go`:

- `var Version, Revision string` (package-level, set by ldflags at release build).
- A `--version` flag that prints `Version` (+ `Revision`) and exits 0.

This lets a fork build identify itself (which `-io41.N` a user is running).
Development builds (no ldflags) print an empty/"dev" value — acceptable.

## Security controls (one-time GitHub setup)

The "external folks cannot trigger" requirement is met by a combination of
GitHub's built-in fork rules and explicit repo settings:

1. **Tag ruleset** (or legacy tag protection rule) matching `v*` — only
   maintainers/admins may create release tags. This is the primary guard on
   releases.
2. **Branch ruleset** on `release` and `master`: require PR + passing `go.yml`
   status check, block force-push and deletion, restrict who can push directly.
3. **Actions → Fork pull request workflows:** keep "Require approval for
   first-time contributors" (the public-repo default).
4. **Default `GITHUB_TOKEN` permissions = read** at the repository level. The
   goreleaser job keeps its explicit `permissions: contents: write` (least
   privilege, scoped to the one job that needs it).
5. **Never** introduce `pull_request_target` with a checkout of untrusted PR
   code (it runs in the base-repo context with secrets). Not used today; keep it
   that way.

## Known non-goal

`go install github.com/io41/terraform-j2md/...` will not work because the
`go.mod` module path is `github.com/reproio/terraform-j2md`. Irrelevant to the
binary-download consumption model. Renaming the module path is intentionally out
of scope (it would complicate merges from upstream for no benefit here).

## Out of scope

- Automated/scheduled upstream syncing (the runbook is run manually as needed).
- Homebrew tap, signing, or other distribution channels.
- Renaming the Go module path.
- Any change to the drift feature itself (handled in PR `reproio#46`).

## Summary of file changes for the implementation plan

1. `.github/workflows/go.yml` — add `release` to push/PR branch filters.
2. `.github/workflows/goreleaser.yml` — widen tag trigger to the `-io41.N`
   pattern.
3. `.goreleaser.yml` — add `release.prerelease: false`.
4. `cmd/terraform-j2md/main.go` — add `Version`/`Revision` vars and a
   `--version` flag.
5. Git/GitHub operations (branch creation, default-branch change, remote add,
   rulesets) — sequenced per the Bootstrap and Security sections; these are
   operational steps, not code, and some require repo-admin rights in the GitHub
   UI/API.
