package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/charmbracelet/huh"
	"github.com/ollama/ollama/api"
	"github.com/untoldecay/BeadsLog/cmd/bd/doctor"
	"github.com/untoldecay/BeadsLog/internal/beads"
	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/configfile"
	"github.com/untoldecay/BeadsLog/internal/extractor"
	"github.com/untoldecay/BeadsLog/internal/git"
	"github.com/untoldecay/BeadsLog/internal/storage/sqlite"
	"github.com/untoldecay/BeadsLog/internal/syncbranch"
	"github.com/untoldecay/BeadsLog/internal/types"
	"github.com/untoldecay/BeadsLog/internal/ui"
	"github.com/untoldecay/BeadsLog/internal/utils"
)

var initCmd = &cobra.Command{
	Use:     "init",
	GroupID: "setup",
	Short:   "Initialize bd in the current directory",
	Long: `Initialize bd in the current directory by creating a .beads/ directory
and database file. Optionally specify a custom issue prefix.

With --no-db: creates .beads/ directory and issues.jsonl file instead of SQLite database.

With --from-jsonl: imports from the current .beads/issues.jsonl file on disk instead
of scanning git history. Use this after manual JSONL cleanup (e.g., bd compact --purge-tombstones)
to prevent deleted issues from being resurrected during re-initialization.

With --solo: configures local-only mode for personal use in a shared repo.
Sets sync.mode: local-only (no push, no daemon auto-sync) and lets you choose
between git-excluding beads files entirely or committing them to a local-only
branch. 'bd doctor' skips remote sync checks in this mode.

With --stealth: configures per-repository git settings for invisible beads usage:
  • .git/info/exclude to prevent beads files from being committed
  • Claude Code settings with bd onboard instruction
  Perfect for personal use without affecting repo collaborators.`,
	Run: func(cmd *cobra.Command, _ []string) {
		prefix, _ := cmd.Flags().GetString("prefix")
		quiet, _ := cmd.Flags().GetBool("quiet")
		branch, _ := cmd.Flags().GetString("branch")
		contributor, _ := cmd.Flags().GetBool("contributor")
		team, _ := cmd.Flags().GetBool("team")
		solo, _ := cmd.Flags().GetBool("solo")
		stealth, _ := cmd.Flags().GetBool("stealth")
		inline, _ := cmd.Flags().GetBool("inline")

		if solo && team {
			fmt.Fprintln(os.Stderr, "Error: --solo and --team are mutually exclusive")
			os.Exit(1)
		}
		skipMergeDriver, _ := cmd.Flags().GetBool("skip-merge-driver")
		skipHooks, _ := cmd.Flags().GetBool("skip-hooks")
		force, _ := cmd.Flags().GetBool("force")
		fromJSONL, _ := cmd.Flags().GetBool("from-jsonl")

		// Initialize config (PersistentPreRun doesn't run for init command)
		if err := config.Initialize(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to initialize config: %v\n", err)
			// Non-fatal - continue with defaults
		}

		// Safety guard: check for existing JSONL with issues
		// This prevents accidental re-initialization in fresh clones
		if !force {
			if err := checkExistingBeadsData(prefix); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		}

		// Handle stealth mode setup
		if stealth {
			if err := setupStealthMode(!quiet); err != nil {
				fmt.Fprintf(os.Stderr, "Error setting up stealth mode: %v\n", err)
				os.Exit(1)
			}

			// In stealth mode, skip git hooks and merge driver installation
			// since we handle it globally
			skipHooks = true
			skipMergeDriver = true
		}

		// Check BEADS_DB environment variable if --db flag not set
		// (PersistentPreRun doesn't run for init command)
		if dbPath == "" {
			if envDB := os.Getenv("BEADS_DB"); envDB != "" {
				dbPath = envDB
			}
		}

		// Determine prefix with precedence: flag > config > auto-detect from git > auto-detect from directory name
		if prefix == "" {
			// Try to get from config file
			prefix = config.GetString("issue-prefix")
		}

		// auto-detect prefix from first issue in JSONL file
		if prefix == "" {
			issueCount, jsonlPath, gitRef := checkGitForIssues()
			if issueCount > 0 {
				firstIssue, err := readFirstIssueFromGit(jsonlPath, gitRef)
				if firstIssue != nil && err == nil {
					prefix = utils.ExtractIssuePrefix(firstIssue.ID)
				}
			}
		}

		// auto-detect prefix from directory name
		if prefix == "" {
			// Auto-detect from directory name
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to get current directory: %v\n", err)
				os.Exit(1)
			}
			prefix = filepath.Base(cwd)
		}

		// Show logo at the very beginning (unless quiet)
		if !quiet {
			fmt.Println()
			fmt.Println(ui.RenderInitLogo())
			fmt.Println(ui.RenderBold("BeadsLog Setup Wizard"))
			fmt.Println("Quick setup will scaffold orchestration rules, devlog space, and git hooks.")
			fmt.Println()
		}

		// Normalize prefix: strip trailing hyphens
		// The hyphen is added automatically during ID generation
		prefix = strings.TrimRight(prefix, "-")

		// Create database
		// Use global dbPath if set via --db flag or BEADS_DB env var, otherwise default to .beads/beads.db
		initDBPath := dbPath
		if initDBPath == "" {
			initDBPath = filepath.Join(".beads", beads.CanonicalDatabaseName)
		}
		dbPath = initDBPath // Ensure global dbPath is set for downstream calls (e.g., devlog init)

		// Migrate old database files if they exist
		if err := migrateOldDatabases(initDBPath, quiet); err != nil {
			fmt.Fprintf(os.Stderr, "Error during database migration: %v\n", err)
			os.Exit(1)
		}

		// Determine if we should create .beads/ directory in CWD or main repo root
		// For worktrees, .beads should always be in the main repository root
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get current directory: %v\n", err)
			os.Exit(1)
		}

		// Check if we're in a git worktree
		// Guard with isGitRepo() check first - on Windows, git commands may hang
		// when run outside a git repository (GH#727)
		isWorktree := false
		if isGitRepo() {
			isWorktree = git.IsWorktree()
		}
		var beadsDir string
		if isWorktree {
			// For worktrees, .beads should be in the main repository root
			mainRepoRoot, err := git.GetMainRepoRoot()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to get main repository root: %v\n", err)
				os.Exit(1)
			}
			beadsDir = filepath.Join(mainRepoRoot, ".beads")
		} else {
			// For regular repos, use current directory
			beadsDir = filepath.Join(cwd, ".beads")
		}

		// Prevent nested .beads directories
		// Check if current working directory is inside a .beads directory
		if strings.Contains(filepath.Clean(cwd), string(filepath.Separator)+".beads"+string(filepath.Separator)) ||
			strings.HasSuffix(filepath.Clean(cwd), string(filepath.Separator)+".beads") {
			fmt.Fprintf(os.Stderr, "Error: cannot initialize bd inside a .beads directory\n")
			fmt.Fprintf(os.Stderr, "Current directory: %s\n", cwd)
			fmt.Fprintf(os.Stderr, "Please run 'bd init' from outside the .beads directory.\n")
			os.Exit(1)
		}

		initDBDir := filepath.Dir(initDBPath)

		// Convert both to absolute paths for comparison
		beadsDirAbs, err := filepath.Abs(beadsDir)
		if err != nil {
			beadsDirAbs = filepath.Clean(beadsDir)
		}
		initDBDirAbs, err := filepath.Abs(initDBDir)
		if err != nil {
			initDBDirAbs = filepath.Clean(initDBDir)
		}

		useLocalBeads := filepath.Clean(initDBDirAbs) == filepath.Clean(beadsDirAbs)

		if useLocalBeads {
			// Create .beads directory
			if err := os.MkdirAll(beadsDir, 0750); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to create .beads directory: %v\n", err)
				os.Exit(1)
			}

			// Handle --no-db mode: create issues.jsonl file instead of database
			if noDb {
				// Create empty issues.jsonl file
				jsonlPath := filepath.Join(beadsDir, "issues.jsonl")
				if _, err := os.Stat(jsonlPath); os.IsNotExist(err) {
					// nolint:gosec // G306: JSONL file needs to be readable by other tools
					if err := os.WriteFile(jsonlPath, []byte{}, 0644); err != nil {
						fmt.Fprintf(os.Stderr, "Error: failed to create issues.jsonl: %v\n", err)
						os.Exit(1)
					}
				}

				// Create empty interactions.jsonl file (append-only agent audit log)
				interactionsPath := filepath.Join(beadsDir, "interactions.jsonl")
				if _, err := os.Stat(interactionsPath); os.IsNotExist(err) {
					// nolint:gosec // G306: JSONL file needs to be readable by other tools
					if err := os.WriteFile(interactionsPath, []byte{}, 0644); err != nil {
						fmt.Fprintf(os.Stderr, "Error: failed to create interactions.jsonl: %v\n", err)
						os.Exit(1)
					}
				}

				// Create metadata.json for --no-db mode
				cfg := configfile.DefaultConfig()
				if err := cfg.Save(beadsDir); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to create metadata.json: %v\n", err)
					// Non-fatal - continue anyway
				}

				// Create config.yaml with no-db: true
				if err := createConfigYaml(beadsDir, true); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to create config.yaml: %v\n", err)
					// Non-fatal - continue anyway
				}

				// Create README.md
				if err := createReadme(beadsDir); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to create README.md: %v\n", err)
					// Non-fatal - continue anyway
				}

				if !quiet {
					fmt.Printf("\n%s bd initialized successfully in --no-db mode!\n\n", ui.RenderPass("✓"))
					fmt.Printf("  Mode: %s\n", ui.RenderAccent("no-db (JSONL-only)"))
					fmt.Printf("  Issues file: %s\n", ui.RenderAccent(jsonlPath))
					fmt.Printf("  Issue prefix: %s\n", ui.RenderAccent(prefix))
					fmt.Printf("  Issues will be named: %s\n\n", ui.RenderAccent(prefix+"-"+"<hash> (e.g., "+prefix+"-a3f2dd)"))
					fmt.Printf("Run %s to get started.\n\n", ui.RenderAccent("bd --no-db quickstart"))
				}
				return
			}

			// Create/update .gitignore in .beads directory (idempotent - always update to latest)
			gitignorePath := filepath.Join(beadsDir, ".gitignore")
			if err := os.WriteFile(gitignorePath, []byte(doctor.GitignoreTemplate), 0600); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create/update .gitignore: %v\n", err)
				// Non-fatal - continue anyway
			}

			// Ensure interactions.jsonl exists (append-only agent audit log)
			interactionsPath := filepath.Join(beadsDir, "interactions.jsonl")
			if _, err := os.Stat(interactionsPath); os.IsNotExist(err) {
				// nolint:gosec // G306: JSONL file needs to be readable by other tools
				if err := os.WriteFile(interactionsPath, []byte{}, 0644); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to create interactions.jsonl: %v\n", err)
					// Non-fatal - continue anyway
				}
			}
		}

		// Ensure parent directory exists for the database
		if err := os.MkdirAll(initDBDir, 0750); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create database directory %s: %v\n", initDBDir, err)
			os.Exit(1)
		}

		ctx := rootCtx
		store, err := sqlite.New(ctx, initDBPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to create database: %v\n", err)
			os.Exit(1)
		}

		// === CONFIGURATION METADATA (Pattern A: Fatal) ===
		// Configuration metadata is essential for core functionality and must succeed.
		// These settings define fundamental behavior (issue IDs, sync workflow).
		// Failure here indicates a serious problem that prevents normal operation.

		// Set the issue prefix in config
		if err := store.SetConfig(ctx, "issue_prefix", prefix); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to set issue prefix: %v\n", err)
			_ = store.Close()
			os.Exit(1)
		}

		// === TRACKING METADATA (Pattern B: Warn and Continue) ===
		// Tracking metadata enhances functionality (diagnostics, version checks, collision detection)
		// but the system works without it. Failures here degrade gracefully - we warn but continue.
		// Examples: bd_version enables upgrade warnings, repo_id/clone_id help with collision detection.

		// Store the bd version in metadata (for version mismatch detection)
		if err := store.SetMetadata(ctx, "bd_version", Version); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to store version metadata: %v\n", err)
			// Non-fatal - continue anyway
		}

		// Compute and store repository fingerprint
		repoID, err := beads.ComputeRepoID()
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "Warning: could not compute repository ID: %v\n", err)
			}
		} else {
			if err := store.SetMetadata(ctx, "repo_id", repoID); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to set repo_id: %v\n", err)
			} else if !quiet {
				fmt.Printf("Repository ID: %s\n", repoID[:8])
			}
		}

		// Store clone-specific ID
		cloneID, err := beads.GetCloneID()
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "Warning: could not compute clone ID: %v\n", err)
			}
		} else {
			if err := store.SetMetadata(ctx, "clone_id", cloneID); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to set clone_id: %v\n", err)
			} else if !quiet {
				fmt.Printf("Clone ID: %s\n", cloneID)
			}
		}

		// Create or preserve metadata.json for database metadata (bd-zai fix)
		if useLocalBeads {
			// First, check if metadata.json already exists
			existingCfg, err := configfile.Load(beadsDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load existing metadata.json: %v\n", err)
			}

			var cfg *configfile.Config
			if existingCfg != nil {
				// Preserve existing config
				cfg = existingCfg
			} else {
				// Create metadata, detecting JSONL filename from existing files
				cfg = configfile.DefaultConfig()
				// Check if beads.jsonl exists but issues.jsonl doesn't (legacy)
								issuesPath := filepath.Join(beadsDir, "issues.jsonl")
								beadsPath := filepath.Join(beadsDir, "beads.jsonl")
								if _, err := os.Stat(beadsPath); err == nil {
									if _, err := os.Stat(issuesPath); os.IsNotExist(err) {
										cfg.JSONLExport = "beads.jsonl" // Legacy filename
									}
								}			}
			if err := cfg.Save(beadsDir); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create metadata.json: %v\n", err)
				// Non-fatal - continue anyway
			}

			// Create config.yaml template
			if err := createConfigYaml(beadsDir, false); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create config.yaml: %v\n", err)
				// Non-fatal - continue anyway
			}

			// Create README.md
			if err := createReadme(beadsDir); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create README.md: %v\n", err)
				// Non-fatal - continue anyway
			}
		}

		// Default beads commits to a dedicated sync branch (beads-metadata) so they
		// never land on the user's work branch (develop/main). A dedicated branch
		// is safe re: GH#807 — it is never checked out in the working tree, so the
		// sync worktree can always check it out (unlike auto-picking the current
		// branch). --inline restores the old commit-to-current-branch behavior.
		// Solo/stealth manage their own posture (invisible = no branch at all,
		// local-branch = beads-local, never pushed), so leave them alone.
		//
		// GH#927: This must run AFTER createConfigYaml() so that config.yaml exists
		// and syncbranch.Set() can update it via config.SetYamlConfig() (PR#910 mechanism)
		// Only auto-engage when a remote exists — that's the only case where beads
		// can leak onto a shared work branch via push. Purely local repos keep the
		// simple inline behavior (nothing is pushed anywhere anyway).
		if branch == "" && !inline && !solo && !stealth && hasGitRemote(ctx) {
			branch = defaultSyncBranch
		}
		if branch != "" {
			if err := syncbranch.Set(ctx, store, branch); err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to set sync branch: %v\n", err)
				_ = store.Close()
				os.Exit(1)
			}
			// Also keep beads data out of the work tree's staging, so a manual
			// 'git add -A' can't land it on the work branch.
			excludeBeadsFromWorkBranch()
			if !quiet {
				fmt.Printf("  Sync branch: %s (beads commits stay off your work branches; --inline to disable)\n", branch)
			}
		}

		// Pin the resolved issue-prefix in the committed config.yaml so it is the
		// shared source of truth — decoupled from the (renameable) directory name
		// and serving as the authored signature for prefix migrations
		// (BeadsLog-bl1 / b4p). Must run after createConfigYaml() so the file
		// exists; SetYamlConfigAt updates the commented template line in place.
		if prefix != "" {
			cfgPath := filepath.Join(beadsDir, "config.yaml")
			if err := config.SetYamlConfigAt(cfgPath, "issue-prefix", prefix); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to pin issue-prefix in config.yaml: %v\n", err)
			} else if !quiet {
				fmt.Printf("  Issue prefix: %s (pinned in config.yaml)\n", prefix)
			}
		}

		// Check if git has existing issues to import (fresh clone scenario)
		// With --from-jsonl: import from local file instead of git history
		if fromJSONL {
			// Import from current working tree's JSONL file
			localJSONLPath := filepath.Join(beadsDir, "issues.jsonl")
			if _, err := os.Stat(localJSONLPath); err == nil {
				issueCount, err := importFromLocalJSONL(ctx, initDBPath, store, localJSONLPath)
				if err != nil {
					if !quiet {
						fmt.Fprintf(os.Stderr, "Warning: import from local JSONL failed: %v\n", err)
					}
					// Non-fatal - continue with empty database
				} else if !quiet && issueCount > 0 {
					fmt.Fprintf(os.Stderr, "✓ Imported %d issues from local %s\n", issueCount, localJSONLPath)
				}
			} else if !quiet {
				fmt.Fprintf(os.Stderr, "Warning: --from-jsonl specified but %s not found\n", localJSONLPath)
			}

			// Also import aliases from local file
			localAliasesPath := filepath.Join(beadsDir, "aliases.jsonl")
			if _, err := os.Stat(localAliasesPath); err == nil {
				if err := importAliasesFromJSONL(ctx, store, localAliasesPath); err == nil {
					if !quiet {
						fmt.Fprintf(os.Stderr, "✓ Imported entity aliases from local %s\n", localAliasesPath)
					}
				}
			}
		} else {
			// Default: import from git history
			issueCount, jsonlPath, gitRef := checkGitForIssues()
			if issueCount > 0 {
				if !quiet {
					fmt.Fprintf(os.Stderr, "\n✓ Database initialized. Found %d issues in git, importing...\n", issueCount)
				}

				if err := importFromGit(ctx, initDBPath, store, jsonlPath, gitRef); err != nil {
					if !quiet {
						fmt.Fprintf(os.Stderr, "Warning: auto-import failed: %v\n", err)
						fmt.Fprintf(os.Stderr, "Try manually: git show %s:%s | bd import -i /dev/stdin\n", gitRef, jsonlPath)
					}
					// Non-fatal - continue with empty database
				} else if !quiet {
					fmt.Fprintf(os.Stderr, "✓ Successfully imported %d issues from git.\n", issueCount)
				}

				// Also import aliases from git history
				aliasesPath := filepath.Join(filepath.Dir(jsonlPath), "aliases.jsonl")
				if aliasesContent, err := readFromGitRef(aliasesPath, gitRef); err == nil {
					// Use a temporary file to leverage importAliasesFromJSONL
					tmpFile, _ := os.CreateTemp("", "aliases.*.jsonl")
					tmpPath := tmpFile.Name()
					_, _ = tmpFile.Write(aliasesContent)
					_ = tmpFile.Close()
					defer os.Remove(tmpPath)

					if err := importAliasesFromJSONL(ctx, store, tmpPath); err == nil && !quiet {
						fmt.Fprintf(os.Stderr, "✓ Successfully imported entity aliases from git.\n")
					}
				}
			}
		}

		// Run contributor wizard if --contributor flag is set
		if contributor {
			if err := runContributorWizard(ctx, store); err != nil {
				fmt.Fprintf(os.Stderr, "Error running contributor wizard: %v\n", err)
				_ = store.Close()
				os.Exit(1)
			}
		}

		// Run team wizard if --team flag is set
		if team {
			if err := runTeamWizard(ctx, store); err != nil {
				fmt.Fprintf(os.Stderr, "Error running team wizard: %v\n", err)
				_ = store.Close()
				os.Exit(1)
			}
		}

		// Run solo wizard if --solo flag is set (local-only mode)
		if solo {
			excluded, err := runSoloWizard(ctx, store, stealth)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error running solo wizard: %v\n", err)
				_ = store.Close()
				os.Exit(1)
			}
			if excluded {
				// Beads files are git-excluded: hooks would fail trying to
				// 'git add' them, and the merge driver has nothing to merge.
				skipHooks = true
				skipMergeDriver = true
			}
		}

		if err := store.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close database: %v\n", err)
		}

		// Fork detection: offer to configure .git/info/exclude (GH#742)
		setupExclude, _ := cmd.Flags().GetBool("setup-exclude")
		if setupExclude {
			// Manual flag - always configure
			if err := setupForkExclude(!quiet); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to configure git exclude: %v\n", err)
			}
		} else if !stealth && isGitRepo() {
			// Auto-detect fork and prompt (skip if stealth - it handles exclude already)
			if isFork, upstreamURL := detectForkSetup(); isFork {
				if promptForkExclude(upstreamURL, quiet) {
					if err := setupForkExclude(!quiet); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to configure git exclude: %v\n", err)
					}
				}
			}
		}

		// Check if we're in a git repo and hooks aren't installed
		// Install by default unless --skip-hooks is passed
		hooksInstalledAtInit := false
		if !skipHooks && isGitRepo() && !hooksInstalled() {
			if err := installGitHooks(); err == nil {
				hooksInstalledAtInit = true
			} else if !quiet {
				fmt.Fprintf(os.Stderr, "\n%s Failed to install git hooks: %v\n", ui.RenderWarn("⚠"), err)
				fmt.Fprintf(os.Stderr, "You can try again with: %s\n\n", ui.RenderAccent("bd doctor --fix"))
			}
		}

		// Check if we're in a git repo and merge driver isn't configured
		// Install by default unless --skip-merge-driver is passed
		mergeDriverInstalledAtInit := false
		if !skipMergeDriver && isGitRepo() && !mergeDriverInstalled() {
			if err := installMergeDriver(); err == nil {
				mergeDriverInstalledAtInit = true
			} else if !quiet {
				fmt.Fprintf(os.Stderr, "\n%s Failed to install merge driver: %v\n", ui.RenderWarn("⚠"), err)
				fmt.Fprintf(os.Stderr, "You can try again with: %s\n\n", ui.RenderAccent("bd doctor --fix"))
			}
		}

				// BeadsLog: Initialize orchestration (Progressive Disclosure) and devlog space

				// Must run even in quiet mode to ensure database state is updated.

				

				var autoSyncStr, enforceDevlogStr, backgroundEnrichStr, branchTrackingStr string
				var devlogAuthor string

				autoSyncStr = "true"
				enforceDevlogStr = "true"
				backgroundEnrichStr = "false"
				branchTrackingStr = "true"

				// Auto-detect git user for devlog author default
				out, _ := exec.Command("git", "config", "user.name").Output()
				devlogAuthor = strings.TrimSpace(string(out))

				selectedModel := "llama3.2:3b" // Default

				var selectedToolNames []string

				// Interactive setup if not quiet and stdin is a TTY
				if !quiet && ui.IsTerminal() {
					fmt.Println() // Space before questions

					// Step 1: Auto-Sync
					syncForm := huh.NewForm(
						huh.NewGroup(
							huh.NewSelect[string]().
								Title("Enable Auto-Sync?").
								Description("Keeps your issue tracker and devlogs in sync automatically via git hooks.").
								Options(
									huh.NewOption("Yes, keep me in sync (Recommended)", "true"),
									huh.NewOption("No, I'll sync manually", "false"),
								).
								Value(&autoSyncStr),
						),
					)
					if err := syncForm.Run(); err != nil {
						fmt.Fprintf(os.Stderr, "Setup wizard cancelled: %v\n", err)
						os.Exit(1)
					}

					// Step 2: Enforce Devlogs
					enforceForm := huh.NewForm(
						huh.NewGroup(
							huh.NewSelect[string]().
								Title("Enforce Devlogs?").
								Description("Prevents commits unless a devlog entry is provided. Highly recommended for AI agents.").
								Options(
									huh.NewOption("Yes, enforce best practices", "true"),
									huh.NewOption("No, allow loose commits", "false"),
								).
								Value(&enforceDevlogStr),
						),
					)
					if err := enforceForm.Run(); err != nil {
						fmt.Fprintf(os.Stderr, "Setup wizard cancelled: %v\n", err)
						os.Exit(1)
					}

					// Step 3: Author Pseudonym
					authorForm := huh.NewForm(
						huh.NewGroup(
							huh.NewInput().
								Title("Devlog Author Pseudonym").
								Description("The name that will appear in your development logs.").
								Placeholder("e.g., untoldecay").
								Value(&devlogAuthor),
						),
					)
					if err := authorForm.Run(); err != nil {
						fmt.Fprintf(os.Stderr, "Setup wizard cancelled: %v\n", err)
						os.Exit(1)
					}

					// Step 4: Branch Tracking
					branchForm := huh.NewForm(
						huh.NewGroup(
							huh.NewSelect[string]().
								Title("Track Branch Context?").
								Description("Automatically record the current git branch and upstream status in devlogs.").
								Options(
									huh.NewOption("Yes, track branch info", "true"),
									huh.NewOption("No, keep it simple", "false"),
								).
								Value(&branchTrackingStr),
						),
					)
					if err := branchForm.Run(); err != nil {
						fmt.Fprintf(os.Stderr, "Setup wizard cancelled: %v\n", err)
						os.Exit(1)
					}

					// Step 5: AI Enrichment
					aiForm := huh.NewForm(
						huh.NewGroup(
							huh.NewSelect[string]().
								Title("Background AI Enrichment?").
								Description("Uses Ollama to automatically extract architectural entities in the background.").
								Options(
									huh.NewOption("Yes, enrich my graphs (Requires Ollama)", "true"),
									huh.NewOption("No, use fast Regex only", "false"),
								).
								Value(&backgroundEnrichStr),
						),
					)
					if err := aiForm.Run(); err != nil {
						fmt.Fprintf(os.Stderr, "Setup wizard cancelled: %v\n", err)
						os.Exit(1)
					}

					// Step 4: Ollama Selection (Conditional)
					if backgroundEnrichStr == "true" {
						// Ollama Selection Wizard
						ollama, err := extractor.NewOllamaExtractor("")
						if err != nil || !ollama.Available(ctx) {
							fmt.Printf("\n  %s Ollama not detected or not running.\n", ui.RenderWarn("⚠"))
							fmt.Println("    To use AI enrichment, please ensure Ollama is installed and running.")
							fmt.Println("    Refer to " + ui.RenderAccent("docs/OLLAMA_ENRICHMENT.md") + " for setup instructions.")
							fmt.Println("    Or use " + ui.RenderCommand("bd help devlog") + " for more information.")
							fmt.Printf("\n    Disabling background enrichment for now.\n")
							backgroundEnrichStr = "false"
						} else {
							models, err := ollama.ListModels(ctx)
							if err != nil || len(models) == 0 {
								// No models found, offer to pull
								pull := false
								pullForm := huh.NewForm(
									huh.NewGroup(
										huh.NewConfirm().
											Title("No Ollama models found").
											Description("Would you like to pull 'llama3.2:1b' (lightweight) now?").
											Affirmative("Yes, pull llama3.2:1b").
											Negative("No, disable AI").
											Value(&pull),
									),
								)
								if err := pullForm.Run(); err == nil && pull {
									selectedModel = "llama3.2:1b"
									fmt.Printf("  %s Pulling %s...\n", ui.RenderAccent("→"), selectedModel)
									err := ollama.PullModel(ctx, selectedModel, func(resp api.ProgressResponse) error {
										if resp.Status != "" {
											if resp.Total > 0 {
												percent := float64(resp.Completed) / float64(resp.Total) * 100
												fmt.Printf("\r  %s %s: %.1f%%          ", ui.RenderAccent("→"), resp.Status, percent)
											} else {
												fmt.Printf("\r  %s %s          ", ui.RenderAccent("→"), resp.Status)
											}
										}
										return nil
									})
									if err != nil {
										fmt.Printf("\n  %s Failed to pull model: %v\n", ui.RenderFail("✗"), err)
										backgroundEnrichStr = "false"
									} else {
										fmt.Printf("\n  %s Model pulled successfully!\n", ui.RenderPass("✓"))
									}
								} else {
									backgroundEnrichStr = "false"
								}
							} else {
								// Models found, show selection
								options := make([]huh.Option[string], len(models))
								for i, m := range models {
									options[i] = huh.NewOption(m, m)
								}

								modelForm := huh.NewForm(
									huh.NewGroup(
										huh.NewSelect[string]().
											Title("Select Ollama Model").
											Description("Choose a model for architectural entity extraction.").
											Options(options...).
											Value(&selectedModel),
									),
								)
								if err := modelForm.Run(); err != nil {
									backgroundEnrichStr = "false"
								}
							}
						}
					}

					// Step 5: Agent Instructions
					agentOptions := make([]huh.Option[string], len(AgentToolCandidates))
					for i, tool := range AgentToolCandidates {
						agentOptions[i] = huh.NewOption(tool.Name, tool.Name).Selected(true)
					}

					agentForm := huh.NewForm(
						huh.NewGroup(
							huh.NewMultiSelect[string]().
								Title("Agent Instructions").
								Description("Select which agent instruction files to generate or update.\n(Space to toggle, Enter to confirm)").
								Options(agentOptions...).
								Limit(20).
								Value(&selectedToolNames),
						),
					)

					if err := agentForm.Run(); err != nil {
						fmt.Fprintf(os.Stderr, "Setup wizard cancelled: %v\n", err)
						os.Exit(1)
					}
				}

				// Map selected tools to file paths
				var targetFiles []string

				// If nothing selected (or non-interactive default), select all tools
				if len(selectedToolNames) == 0 && (!ui.IsTerminal() || quiet) {
					for _, tool := range AgentToolCandidates {
						targetFiles = append(targetFiles, tool.Files...)
					}
				} else {
					// Map selected names to files
					for _, name := range selectedToolNames {
						for _, tool := range AgentToolCandidates {
							if tool.Name == name {
								targetFiles = append(targetFiles, tool.Files...)
								break
							}
						}
					}
				}

		autoSync := autoSyncStr == "true"
		enforceDevlog := enforceDevlogStr == "true"
		backgroundEnrich := backgroundEnrichStr == "true"

		// Persist settings
		if devlogAuthor != "" {
			_ = config.SetYamlConfig("devlog.author", devlogAuthor)
		}
		_ = config.SetYamlConfig("devlog.branch-tracking", branchTrackingStr)

		// Persist background enrich setting if enabled
		if backgroundEnrich {
			_ = config.SetYamlConfig("entity_extraction.background_enrichment", "true")
			_ = config.SetYamlConfig("ollama.model", selectedModel)
		} else {
			_ = config.SetYamlConfig("entity_extraction.background_enrichment", "false")
		}

		orchFiles := initializeOrchestration(false)
		devlogRes := initializeDevlog("_rules/_devlog", quiet, autoSync, enforceDevlog, backgroundEnrich, targetFiles)

		// Collect information for the final report
		initResult := ui.InitResult{
			DBPath:               initDBPath,
			Prefix:               prefix,
			RepoID:               repoID,
			CloneID:              cloneID,
			OrchestrationFiles:   orchFiles,
			DevlogSpaceStatus:    devlogRes.SpaceStatus,
			DevlogPromptStatus:   devlogRes.PromptStatus,
			AgentRules:           devlogRes.AgentRules,
			DevlogHooks:          devlogRes.Hooks,
			HooksInstalled:       hooksInstalledAtInit || (isGitRepo() && hooksInstalled()),
			MergeDriverInstalled: mergeDriverInstalledAtInit || (isGitRepo() && mergeDriverInstalled()),
		}		// Run bd doctor diagnostics to catch setup issues early
		doctorResult := runDiagnostics(cwd)
		for _, check := range doctorResult.Checks {
			if check.Status != statusOK {
				initResult.DoctorIssues = append(initResult.DoctorIssues, fmt.Sprintf("%s: %s", check.Name, check.Message))
			}
		}

		// Next steps
		initResult.QuickstartCommands = []string{
			"bd quickstart --tasks",
			"bd quickstart --devlog",
		}

		// Render the final unified report
		if !quiet {
			fmt.Println()
			fmt.Println(ui.RenderInitReport(initResult, ui.GetWidth()))
			fmt.Println()
		}
	},
}

func init() {
	initCmd.Flags().StringP("prefix", "p", "", "Issue prefix (default: current directory name)")
	initCmd.Flags().BoolP("quiet", "q", false, "Suppress output (quiet mode)")
	initCmd.Flags().StringP("branch", "b", "", "Git branch for beads commits (default: dedicated "+defaultSyncBranch+" branch)")
	initCmd.Flags().Bool("inline", false, "Commit beads to the current branch instead of a dedicated sync branch")
	initCmd.Flags().Bool("contributor", false, "Run OSS contributor setup wizard")
	initCmd.Flags().Bool("team", false, "Run team workflow setup wizard")
	initCmd.Flags().Bool("solo", false, "Run solo mode setup: beads data stays local, never pushed to remote")
	initCmd.Flags().Bool("stealth", false, "Enable stealth mode: global gitattributes and gitignore, no local repo tracking")
	initCmd.Flags().Bool("setup-exclude", false, "Configure .git/info/exclude to keep beads files local (for forks)")
	initCmd.Flags().Bool("skip-hooks", false, "Skip git hooks installation")
	initCmd.Flags().Bool("skip-merge-driver", false, "Skip git merge driver setup")
	initCmd.Flags().Bool("force", false, "Force re-initialization even if JSONL already has issues (may cause data loss)")
	initCmd.Flags().Bool("from-jsonl", false, "Import from current .beads/issues.jsonl file instead of git history (preserves manual cleanups)")
	rootCmd.AddCommand(initCmd)
}

// migrateOldDatabases detects and migrates old database files to beads.db
func migrateOldDatabases(targetPath string, quiet bool) error {
	targetDir := filepath.Dir(targetPath)
	targetName := filepath.Base(targetPath)

	// If target already exists, no migration needed
	if _, err := os.Stat(targetPath); err == nil {
		return nil
	}

	// Create .beads directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0750); err != nil {
		return fmt.Errorf("failed to create .beads directory: %w", err)
	}

	// Look for existing .db files in the .beads directory
	pattern := filepath.Join(targetDir, "*.db")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to search for existing databases: %w", err)
	}

	// Filter out the target file name and any backup files
	var oldDBs []string
	for _, match := range matches {
		baseName := filepath.Base(match)
		if baseName != targetName && !strings.HasSuffix(baseName, ".backup.db") {
			oldDBs = append(oldDBs, match)
		}
	}

	if len(oldDBs) == 0 {
		// No old databases to migrate
		return nil
	}

	if len(oldDBs) > 1 {
		// Multiple databases found - ambiguous, require manual intervention
		return fmt.Errorf("multiple database files found in %s: %v\nPlease manually rename the correct database to %s and remove others",
			targetDir, oldDBs, targetName)
	}

	// Migrate the single old database
	oldDB := oldDBs[0]
	if !quiet {
		fmt.Fprintf(os.Stderr, "→ Migrating database: %s → %s\n", filepath.Base(oldDB), targetName)
	}

	// Rename the old database to the new canonical name
	if err := os.Rename(oldDB, targetPath); err != nil {
		return fmt.Errorf("failed to migrate database %s to %s: %w", oldDB, targetPath, err)
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "✓ Database migration complete\n\n")
	}

	return nil
}


// readFirstIssueFromJSONL reads the first issue from a JSONL file
func readFirstIssueFromJSONL(path string) (*types.Issue, error) {
	// #nosec G304 -- helper reads JSONL file chosen by current bd command
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSONL file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// skip empty lines
		if line == "" {
			continue
		}

		var issue types.Issue
		if err := json.Unmarshal([]byte(line), &issue); err == nil {
			return &issue, nil
		} else {
			// Skip malformed lines with warning
			fmt.Fprintf(os.Stderr, "Warning: skipping malformed JSONL line %d: %v\n", lineNum, err)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading JSONL file: %w", err)
	}

	return nil, nil
}

// readFirstIssueFromGit reads the first issue from a git ref (bd-0is: supports sync-branch)
func readFirstIssueFromGit(jsonlPath, gitRef string) (*types.Issue, error) {
	output, err := readFromGitRef(jsonlPath, gitRef)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		// skip empty lines
		if line == "" {
			continue
		}

		var issue types.Issue
		if err := json.Unmarshal([]byte(line), &issue); err == nil {
			return &issue, nil
		}
		// Skip malformed lines silently (called during auto-detection)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning git content: %w", err)
	}

	return nil, nil
}


// checkExistingBeadsData checks for existing database files
// and returns an error if found (safety guard for bd-emg)
//
// Note: This only blocks when a database already exists (workspace is initialized).
// Fresh clones with JSONL but no database are allowed - init will create the database
// and import from JSONL automatically (bd-4h9: fixes circular dependency with doctor --fix).
//
// For worktrees, checks the main repository root instead of current directory
// since worktrees should share the database with the main repository.
func checkExistingBeadsData(prefix string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return nil // Can't determine CWD, allow init to proceed
	}

	// Determine where to check for .beads directory
	// Guard with isGitRepo() check first - on Windows, git commands may hang
	// when run outside a git repository (GH#727)
	var beadsDir string
	if isGitRepo() && git.IsWorktree() {
		// For worktrees, .beads should be in the main repository root
		mainRepoRoot, err := git.GetMainRepoRoot()
		if err != nil {
			return nil // Can't determine main repo root, allow init to proceed
		}
		beadsDir = filepath.Join(mainRepoRoot, ".beads")
	} else {
		// For regular repos (or non-git directories), check current directory
		beadsDir = filepath.Join(cwd, ".beads")
	}

	// Check if .beads directory exists
	if _, err := os.Stat(beadsDir); os.IsNotExist(err) {
		return nil // No .beads directory, safe to init
	}

	// Check for existing database file
	dbPath := filepath.Join(beadsDir, beads.CanonicalDatabaseName)
	if _, err := os.Stat(dbPath); err == nil {
		return fmt.Errorf(`
%s Found existing database: %s

This workspace is already initialized.

To use the existing database:
  Just run bd commands normally (e.g., %s)

To completely reinitialize (data loss warning):
  rm -rf .beads && bd init --prefix %s

Aborting.`, ui.RenderWarn("⚠"), dbPath, ui.RenderAccent("bd list"), prefix)
	}

	// Fresh clones (JSONL exists but no database) are allowed - init will
	// create the database and import from JSONL automatically.
	// This fixes the circular dependency where init told users to run
	// "bd doctor --fix" but doctor couldn't create a database (bd-4h9).

	return nil // No database found, safe to init
}
