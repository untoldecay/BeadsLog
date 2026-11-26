# Comprehensive Development Log: Release channels stable and dev (BeadsLog-70m)

**Date:** 2026-11-25

## Problem
No way to publish a pre-release of a branch without risking stable users.
bump-version.sh only accepts X.Y.Z, cut-release.sh commits the bump on the current
branch, `bd upgrade` read the version from main and installed `@main`, and version
comparison ranked `0.66.0-rc.1` above `0.66.0`.

### **Phase 1: Design**

*   **Action Taken:** Two channels, stable and dev (user decision: named dev, not rc). Versions are vX.Y.Z and vX.Y.Z-dev.N tags.
*   **Result:** A dev version merged into main would be offered to every stable user, so dev cuts must never put their bump on a branch.

### **Phase 2: Cutting**

*   **Action Taken:** `scripts/cut-dev.sh` (via `cut-release.sh --dev`) bumps version.go in a throwaway worktree with hooks off, builds and checks the binary, and pushes only the tag. A rejected push removes the local tag. GitHub pre-release uses `--prerelease --latest=false` and is optional.

### **Phase 3: Upgrading**

*   **Action Taken:** New `internal/channel`: Compare (x/mod/semver with a dotted-number fallback), Latest, FetchLatest (`git ls-remote --tags`, no API rate limit), and the per-machine choice in the user config directory. IsNewer and doctor CompareVersions now use Compare.
*   **Action Taken:** `bd upgrade` installs the exact tag, retries with GOPROXY=direct for fresh tags, works outside a repository, and offers the older stable version when going back from a dev build.
*   **Action Taken:** release.yml ignores dev tags.

### **Phase 4: Tests**

*   **Result:** Unit tests in internal/channel; E2E Test 36 cuts real tags against a throwaway remote and records what `go install` would run. E2E 36/36; no new Go failures. Dry run against the real remote: 0.66.0-dev.1.

## Final Session Summary

**Final Status:** Done on feature/dev-channel (stacked on fix/beads-redirects). No real tag pushed.
**Key Learnings:**
*   **Tag-only commits isolate pre-releases:** the branch that will be merged never learns the dev version.
*   **bd 0.65.0 cannot follow dev:** the first dev build is installed by hand.

### Architectural Relationships
- upgradeCmd -> FetchLatest (uses)
- IsNewer -> channel.Compare (uses)
- CompareVersions -> channel.Compare (uses)
- cut-release.sh -> cut-dev.sh (delegates to)
