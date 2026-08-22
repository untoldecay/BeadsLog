package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/storage"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

// soloDevlogDir is the separate, git-excluded devlog space used in solo mode so
// the committed team devlog dir is never touched (BeadsLog-9vd).
const soloDevlogDir = "_rules/_devlog-solo"

// teamDevlogDir is the conventional shared devlog dir.
const teamDevlogDir = "_rules/_devlog"

// gitExcludePath returns the path to .git/info/exclude, creating the info dir.
func gitExcludePath() (string, error) {
	out, err := runGit("rev-parse", "--git-dir")
	if err != nil {
		return "", fmt.Errorf("not a git repository")
	}
	gitDir := strings.TrimSpace(out)
	if err := os.MkdirAll(filepath.Join(gitDir, "info"), 0755); err != nil {
		return "", err
	}
	return filepath.Join(gitDir, "info", "exclude"), nil
}

// addGitExcludePatterns adds patterns to .git/info/exclude (idempotent, exact
// line match). Returns the patterns actually added.
func addGitExcludePatterns(patterns []string) ([]string, error) {
	path, err := gitExcludePath()
	if err != nil {
		return nil, err
	}
	var content string
	if b, err := os.ReadFile(path); err == nil { // #nosec G304 - path from git-dir
		content = string(b)
	}
	var toAdd []string
	for _, p := range patterns {
		if !containsExactPattern(content, p) {
			toAdd = append(toAdd, p)
		}
	}
	if len(toAdd) == 0 {
		return nil, nil
	}
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += "\n# Beads solo mode (bd init --solo)\n"
	for _, p := range toAdd {
		content += p + "\n"
	}
	// #nosec G306 - git exclude is a plain config file
	return toAdd, os.WriteFile(path, []byte(content), 0644)
}

// removeGitExcludePatterns removes exactly-matching pattern lines from
// .git/info/exclude, leaving the user's other excludes intact.
func removeGitExcludePatterns(patterns []string) error {
	path, err := gitExcludePath()
	if err != nil {
		return err
	}
	b, err := os.ReadFile(path) // #nosec G304 - path from git-dir
	if err != nil {
		return nil // nothing to remove
	}
	drop := make(map[string]bool, len(patterns))
	for _, p := range patterns {
		drop[p] = true
	}
	var kept []string
	for _, line := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(line)
		if drop[t] || t == "# Beads solo mode (bd init --solo)" {
			continue
		}
		kept = append(kept, line)
	}
	out := strings.TrimRight(strings.Join(kept, "\n"), "\n") + "\n"
	// #nosec G306 - git exclude is a plain config file
	return os.WriteFile(path, []byte(out), 0644)
}

// transitionToSolo repoints the devlog space to the excluded soloDevlogDir and
// excludes .beads/ + the solo dir. With continuity, the team's committed
// devlogs are COPIED into the solo dir so the graph keeps team history; the
// team dir itself is never modified. Returns the number of devlogs carried over.
func transitionToSolo(ctx context.Context, store storage.Storage, continuity bool) (carried int, err error) {
	if err := os.MkdirAll(soloDevlogDir, 0750); err != nil {
		return 0, err
	}
	if continuity {
		carried, err = copyDevlogs(teamDevlogDir, soloDevlogDir)
		if err != nil {
			return 0, err
		}
	}
	if err := store.SetConfig(ctx, "devlog_dir", soloDevlogDir); err != nil {
		return carried, err
	}
	// Hide the whole beads footprint from git in one exclude write: .beads/, the
	// solo devlog dir, the _rules scaffolding, and any untracked agent file
	// carrying the protocol block. Agents read these from disk, so solo stays
	// fully functional; the team just never sees beads (BeadsLog-9vd).
	footprintPatterns, trackedAgents := beadsFootprintExcludes()
	patterns := append([]string{".beads/", soloDevlogDir + "/"}, footprintPatterns...)
	if _, err := addGitExcludePatterns(patterns); err != nil {
		return carried, err
	}
	// .git/info/exclude only hides UNTRACKED files. Already-tracked files (an
	// established repo) need skip-worktree so local mutations don't leak: the
	// .beads/ state (kf7) and the protocol block in the team's own agent files.
	_ = setSkipWorktree(gitTrackedBeadsFiles(), true)
	_ = setSkipWorktree(trackedAgents, true)
	return carried, nil
}

// transitionToTeam reverses solo mode: publishes NEW solo devlogs into the team
// dir (files not already there), repoints devlog_dir back, removes the solo
// excludes, and clears the local-only config. Returns published file count.
func transitionToTeam(ctx context.Context, store storage.Storage) (published int, err error) {
	if _, statErr := os.Stat(soloDevlogDir); statErr == nil {
		published, err = publishSoloDevlogs(soloDevlogDir, teamDevlogDir)
		if err != nil {
			return 0, err
		}
	}
	if err := store.SetConfig(ctx, "devlog_dir", teamDevlogDir); err != nil {
		return published, err
	}
	// Clear skip-worktree first so the tracked .beads/ files + agent protocol
	// blocks are visible to git again — rejoining the team deliberately
	// re-exposes local issue state to merge (per-issue LWW), which the caller
	// warns about (kf7).
	_ = setSkipWorktree(gitTrackedBeadsFiles(), false)
	_, trackedAgents := beadsFootprintExcludes()
	_ = setSkipWorktree(trackedAgents, false)
	// Remove every possible footprint pattern (fixed dirs + all candidate agent
	// files); patterns not present are ignored.
	unExclude := append([]string{".beads/", soloDevlogDir + "/"}, soloFootprintDirs...)
	unExclude = append(unExclude, footprintFiles...)
	if err := removeGitExcludePatterns(unExclude); err != nil {
		return published, err
	}
	// Drop the per-machine override entirely so the committed config.yaml
	// (team defaults) is authoritative again — nothing to un-write since solo
	// settings never touched config.yaml in invisible mode (BeadsLog-9vd).
	if localPath, err := config.LocalYamlConfigPath(); err == nil {
		_ = os.Remove(localPath)
	}
	// Remove the now-redundant solo devlog dir: its contents were either
	// published to the team dir or carried from it. Left behind (and no longer
	// excluded) it would show as untracked clutter, at risk of being committed.
	_ = os.RemoveAll(soloDevlogDir)
	return published, nil
}

// copyDevlogs copies every *.md from src into dst (skipping ones already there).
func copyDevlogs(src, dst string) (int, error) {
	entries, err := os.ReadDir(src)
	if err != nil {
		return 0, nil // no team devlogs to carry
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		target := filepath.Join(dst, e.Name())
		if _, err := os.Stat(target); err == nil {
			continue
		}
		b, err := os.ReadFile(filepath.Join(src, e.Name())) // #nosec G304 - src is repo dir
		if err != nil {
			continue
		}
		if err := os.WriteFile(target, b, 0600); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// publishSoloDevlogs moves solo *.md files that don't yet exist in the team dir
// into it (the carried-over copies already exist there and are skipped).
func publishSoloDevlogs(solo, team string) (int, error) {
	entries, err := os.ReadDir(solo)
	if err != nil {
		return 0, nil
	}
	if err := os.MkdirAll(team, 0750); err != nil {
		return 0, err
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		target := filepath.Join(team, e.Name())
		if _, err := os.Stat(target); err == nil {
			continue // already in the team dir (a carried-over copy)
		}
		b, err := os.ReadFile(filepath.Join(solo, e.Name())) // #nosec G304 - solo is repo dir
		if err != nil {
			continue
		}
		if err := os.WriteFile(target, b, 0600); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// runGit runs a git command and returns its stdout.
func runGit(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output() // #nosec G204 - fixed args
	return string(out), err
}

// gitTrackedBeadsFiles returns the git-tracked files under .beads/ (issues.jsonl,
// config.yaml, aliases.jsonl, …). In an established team-beads repo these are
// already committed, so .git/info/exclude — which only affects UNTRACKED files —
// cannot hide the solo user's local mutations. Empty slice if none are tracked.
func gitTrackedBeadsFiles() []string {
	out, err := runGit("ls-files", "-z", ".beads/")
	if err != nil {
		return nil
	}
	var files []string
	for _, f := range strings.Split(strings.TrimRight(out, "\x00"), "\x00") {
		if f != "" {
			files = append(files, f)
		}
	}
	return files
}

// setSkipWorktree flips the skip-worktree bit on tracked files so git ignores
// local modifications to them (they won't be staged, committed or pushed) —
// git's native mechanism for locally diverging a tracked file. Passing skip
// false clears it, re-exposing the file to git.
//
// ponytail: skip-worktree can stall a `git rebase`/`pull` that also touches the
// file; solo runs with daemon auto-sync OFF so the daemon can't race it (the
// footgun in memory git-rebase-vs-bd-daemon). Upgrade path if that bites: move
// solo beads to a separate untracked DB/JSONL path instead of hiding the tracked one.
func setSkipWorktree(files []string, skip bool) error {
	if len(files) == 0 {
		return nil
	}
	flag := "--skip-worktree"
	if !skip {
		flag = "--no-skip-worktree"
	}
	_, err := runGit(append([]string{"update-index", flag}, files...)...)
	return err
}

// soloFootprintDirs are the bd scaffolding dirs (beyond .beads/) that pollute a
// team which doesn't use beads. Only untracked instances are actually hidden —
// .git/info/exclude leaves tracked files alone, so a team that committed these
// (i.e. does use beads) keeps them.
var soloFootprintDirs = []string{"_rules/_orchestration/", "_rules/_devlog/"}

// gitIsTracked reports whether path is tracked by git.
func gitIsTracked(path string) bool {
	out, err := runGit("ls-files", "--", path)
	return err == nil && strings.TrimSpace(out) != ""
}

// footprintFiles are the shared config/agent files bd injects beads content
// into: the agent instruction files plus .gitattributes (the JSONL merge
// driver). Each is hidden only if it actually carries a bd marker.
var footprintFiles = append(append([]string{}, Candidates...), ".gitattributes")

// agentFileHasProtocol reports whether a footprint file carries content bd
// injected: the <beads_protocol> block or pre-onboard trap in agent files, or
// the merge driver in .gitattributes.
func agentFileHasProtocol(path string) bool {
	b, err := os.ReadFile(path) // #nosec G304 - fixed candidate names
	if err != nil {
		return false
	}
	s := string(b)
	if path == ".gitattributes" {
		return strings.Contains(s, "merge=beads")
	}
	return strings.Contains(s, ProtocolStartTag) ||
		strings.Contains(s, LegacyProtocolStartTag) ||
		strings.Contains(s, BootstrapTrap)
}

// beadsFootprintExcludes returns the exclude patterns for the non-.beads/ beads
// footprint (scaffolding dirs + untracked agent files carrying the protocol
// block) and the tracked agent files that must instead be skip-worktree'd. This
// is what keeps beads from polluting a team that doesn't use it (BeadsLog-9vd):
// agents read these files from disk regardless of git, so hiding them from git
// leaves the solo setup fully functional while the team never sees it.
func beadsFootprintExcludes() (patterns, trackedAgents []string) {
	patterns = append(patterns, soloFootprintDirs...)
	for _, f := range footprintFiles {
		if !agentFileHasProtocol(f) {
			continue
		}
		if gitIsTracked(f) {
			trackedAgents = append(trackedAgents, f)
		} else {
			patterns = append(patterns, f)
		}
	}
	return patterns, trackedAgents
}

// protocolCommittedFiles returns agent files whose committed (HEAD) version
// already contains the protocol block — skip-worktree can't hide what's already
// in history, so the caller warns the user to strip it.
func protocolCommittedFiles() []string {
	var committed []string
	for _, f := range footprintFiles {
		if !agentFileHasProtocol(f) || !gitIsTracked(f) {
			continue
		}
		if out, err := runGit("show", "HEAD:"+f); err == nil &&
			(strings.Contains(out, ProtocolStartTag) ||
				strings.Contains(out, LegacyProtocolStartTag) ||
				strings.Contains(out, BootstrapTrap) ||
				strings.Contains(out, "merge=beads")) {
			committed = append(committed, f)
		}
	}
	return committed
}

func printSoloDevlogNote(carried int, continuity bool) {
	if continuity {
		fmt.Printf("%s Continuity: %d team devlog(s) carried into %s (team dir untouched)\n",
			ui.RenderPass("✓"), carried, ui.RenderAccent(soloDevlogDir))
	} else {
		fmt.Printf("%s Fresh solo graph: new devlogs go to %s\n", ui.RenderPass("✓"), ui.RenderAccent(soloDevlogDir))
	}
}
