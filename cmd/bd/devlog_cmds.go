package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/charmbracelet/huh"
	"github.com/untoldecay/BeadsLog/internal/beads"
	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/extractor"
	"github.com/untoldecay/BeadsLog/internal/queries"
	"github.com/untoldecay/BeadsLog/internal/storage/sqlite"
	"github.com/untoldecay/BeadsLog/internal/types"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

var devlogCmd = &cobra.Command{
	Use:   "devlog",
	Short: "Devlog management commands",
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand, check if initialized and provide guidance
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Println("Beads database not initialized. Run 'bd init' first.")
			return
		}
		defer store.Close()

		devlogDir, _ := store.GetConfig(rootCtx, "devlog_dir")
		if devlogDir == "" {
			fmt.Println("Devlog space not configured.")
			fmt.Println("\nSuggestions:")
			fmt.Println("  1. Run 'bd devlog init' to set up a new devlog space in _rules/_devlog")
			fmt.Println("  2. If you already have devlogs, run 'bd devlog init <path>' to point to them")
		} else {
			_ = cmd.Help()
		}
	},
}

// devlogInitCmd scaffolds the devlog structure
var devlogInitCmd = &cobra.Command{
	Use:   "initialize [dir]",
	Short: "Initialize devlog structure",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		baseDir := "_rules/_devlog"
		if len(args) > 0 {
			baseDir = args[0]
		}
		// In manual mode (non-wizard), we assume background enrichment off by default
		initializeDevlog(baseDir, false, true, false, false, Candidates)
	},
}

type DevlogInitResult struct {

	SpaceStatus  string

	PromptStatus string

	AgentRules   []string

	Hooks        []string

}



// writeGenesisDevlog creates the first devlog of a fresh install — a real,
// honest record of the initialization itself — and registers it in the index
// with a stable session ID, exactly like a normal 'bd devlog record'.
func writeGenesisDevlog(baseDir, indexPath string, quiet bool) {
	now := time.Now()
	date := now.Format("2006-01-02 15:04")

	author := os.Getenv("BD_ACTOR")
	if author == "" {
		out, _ := exec.Command("git", "config", "user.name").Output()
		author = strings.TrimSpace(string(out))
	}
	if author == "" {
		author = "Unknown"
	}

	repoName := ""
	if out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		repoName = filepath.Base(strings.TrimSpace(string(out)))
	}
	if repoName == "" {
		if cwd, err := os.Getwd(); err == nil {
			repoName = filepath.Base(cwd)
		} else {
			repoName = "this-repository"
		}
	}

	branch := "main"
	if out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		if b := strings.TrimSpace(string(out)); b != "" {
			branch = b
		}
	}

	// The relationship arrow endpoint must survive noise filtering, or the
	// genesis session starts incomplete (short repo names like "ui" would).
	arrowTarget := repoName
	if extractor.IsNoise(arrowTarget) {
		arrowTarget = repoName + "-repository"
	}

	content := strings.NewReplacer(
		"{{DATE}}", date,
		"{{AUTHOR}}", author,
		"{{REPO}}", repoName,
		"{{ARROW_TARGET}}", arrowTarget,
		"{{BRANCH}}", branch,
	).Replace(genesisTemplate)

	fileName := now.Format("2006-01-02") + "_beadslog-genesis.md"
	if err := os.WriteFile(filepath.Join(baseDir, fileName), []byte(content), 0644); err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Error writing genesis devlog: %v\n", err)
		}
		return
	}

	subject := "[init] BeadsLog Genesis"
	sessionID := fmt.Sprintf("sess-%s", hashID(subject+date))
	problem := fmt.Sprintf("BeadsLog initialized on %s by %s", date, author)
	row := fmt.Sprintf("| %s | %s | %s | %s | %s | %s | [%s](%s?id=%s) |\n",
		subject, problem, author, "bd init", date, branch, fileName, fileName, sessionID)

	f, err := os.OpenFile(indexPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Error registering genesis devlog: %v\n", err)
		}
		return
	}
	defer f.Close()
	_, _ = f.WriteString(row)
}

func initializeDevlog(baseDir string, quiet bool, autoSync, enforce, backgroundEnrich bool, targetAgents []string) DevlogInitResult {

	var result DevlogInitResult

	statusDir := "(Created)"

	if _, err := os.Stat(baseDir); err == nil {

		statusDir = "(Already exists)"

	} else if err := os.MkdirAll(baseDir, 0755); err != nil {

		if !quiet {

			fmt.Fprintf(os.Stderr, "Error creating devlog dir: %v\n", err)

		}

		return result

	}

	result.SpaceStatus = statusDir



	// Create _index.md
	indexPath := filepath.Join(baseDir, "_index.md")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		if err := os.WriteFile(indexPath, []byte(indexTemplate), 0644); err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "Error writing _index.md: %v\n", err)
			}
		} else {
			writeGenesisDevlog(baseDir, indexPath, quiet)
		}
	} else {
		// UPGRADE: Automatically migrate existing index if it's legacy
		migrateIndexInternal(indexPath, quiet)
	}

	// Create _generate-devlog.md in the baseDir
	promptPath := filepath.Join(baseDir, "_generate-devlog.md")
	statusPrompt := "(Created)"

	template := promptTemplateManual
	if backgroundEnrich {
		template = promptTemplateAuto
	}

	// UPGRADE: Always refresh prompt if it exists but is outdated
	// Or just always overwrite if we want to ensure latest instructions
	if _, err := os.Stat(promptPath); err == nil {
		statusPrompt = "(Updated)"
	}
	
	if err := os.WriteFile(promptPath, []byte(template), 0644); err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Error writing prompt: %v\n", err)
		}
	}

	// Create _generate-catchup.md
	catchupPath := filepath.Join(baseDir, "_generate-catchup.md")
	if err := os.WriteFile(catchupPath, []byte(catchupPromptTemplate), 0644); err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Error writing catchup prompt: %v\n", err)
		}
	}

	result.PromptStatus = statusPrompt



	// Store config

	store, err := sqlite.New(rootCtx, dbPath)

	if err != nil {

		if !quiet {

			fmt.Fprintf(os.Stderr, "  %s Warning: could not open database to save devlog configuration: %v\n", ui.RenderWarn("⚠"), err)

		}

	} else {

		defer store.Close()

		// Set only when unset — never clobber a devlog_dir already chosen this
		// run (e.g. solo mode repoints it to the excluded _devlog-solo).
		current, _ := store.GetConfig(rootCtx, "devlog_dir")

		if current == "" {

			if err := store.SetConfig(rootCtx, "devlog_dir", baseDir); err != nil && !quiet {

				fmt.Fprintf(os.Stderr, "  %s Warning: failed to save devlog_dir to database: %v\n", ui.RenderWarn("⚠"), err)

			}

		}

	}



	// Logic remains the same, but we only print if NOT quiet

	if !quiet {

		fmt.Printf("  %s Devlog space: %s %s\n", ui.RenderPass("✓"), baseDir, statusDir)

		fmt.Printf("  %s Devlog prompt: %s %s\n", ui.RenderPass("✓"), promptPath, statusPrompt)



		// Integrity Check

		indexPath := filepath.Join(baseDir, "_index.md")

		if _, err := parseIndexMD(indexPath); err != nil {

			fmt.Printf("  %s Devlog index corrupted: %s\n", ui.RenderWarn("⚠"), indexPath)

			fmt.Println("    Run 'bd devlog sync' to see error and get AI repair directive.")

		} else {

			fmt.Printf("  %s Devlog scaffolding initialized successfully\n", ui.RenderPass("✓"))

			fmt.Println("  Run 'bd devlog sync' to ingest all devlogs in the system.")

		}

		fmt.Println() // Added for spacing

	}



	// Agent Rules Integration (force enabled if called from init)

	result.AgentRules = configureAgentRules(quiet, true, targetAgents)



	// Automation Setup (handled via parameters)

	if autoSync {

		installDevlogHooks(false)

		result.Hooks = append(result.Hooks, "post-commit", "post-merge")

	}



	if enforce {

		if err := config.SetYamlConfig("devlog.enforce-on-commit", "true"); err != nil && !quiet {

			fmt.Fprintf(os.Stderr, "     Warning: failed to set devlog enforcement: %v\n", err)

		}

	} else {

		_ = config.SetYamlConfig("devlog.enforce-on-commit", "false")

	}



	return result

}



func configureAgentRules(quiet bool, forceEnable bool, targetCandidates []string) []string {
	var results []string

	bootstrapTrigger := "BEFORE ANYTHING ELSE: run 'bd onboard' and follow ALL instructions to initialize your session."

	// If no candidates specified, nothing to do
	if len(targetCandidates) == 0 {
		return results
	}

	// Find all existing agent files from the TARGET list
	var foundFiles []string
	for _, f := range targetCandidates {
		if _, err := os.Stat(f); err == nil {
			foundFiles = append(foundFiles, f)
		}
	}

	if len(foundFiles) == 0 {
		// If we specifically selected files that don't exist, create them
		// (Assume forceEnable=true if called from main setup wizard)
		if forceEnable {
			for _, file := range targetCandidates {
				f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
				if err == nil {
					f.WriteString(bootstrapTrigger + "\n")
					f.Close()
					if !quiet {
						fmt.Printf("    %s Agent instruction: %s (Created)\n", ui.RenderPass("✓"), file)
					}
					results = append(results, fmt.Sprintf("%s (Created)", file))
				}
			}
		}
		return results
	}

	// Update all found files
	for _, file := range foundFiles {
		if injectBootstrapTrigger(file, bootstrapTrigger) {
			if !quiet {
				fmt.Printf("    %s Agent instruction: %s (Updated)\n", ui.RenderPass("✓"), file)
			}
			results = append(results, fmt.Sprintf("%s (Updated)", file))
		} else {
			if !quiet {
				fmt.Printf("    %s Agent instruction: %s (Active)\n", ui.RenderPass("✓"), file)
			}
			results = append(results, fmt.Sprintf("%s (Active)", file))
		}
	}

	// Also handle any selected candidates that DON'T exist
	for _, file := range targetCandidates {
		exists := false
		for _, f := range foundFiles {
			if f == file {
				exists = true
				break
			}
		}
		if !exists && forceEnable {
			f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				f.WriteString(bootstrapTrigger + "\n")
				f.Close()
				if !quiet {
					fmt.Printf("    %s Agent instruction: %s (Created)\n", ui.RenderPass("✓"), file)
				}
				results = append(results, fmt.Sprintf("%s (Created)", file))
			}
		}
	}

	return results
}

func injectBootstrapTrigger(file, trigger string) bool {
	content, err := os.ReadFile(file)
	if err != nil {
		return false
	}

	sContent := string(content)
	
	// If it already only contains the trigger, skip
	if strings.Contains(sContent, trigger) {
		return false
	}

	// If it already contains the full protocol (New or Legacy), skip (GH#BeadsLog-976, GH#BeadsLog-061)
	if strings.Contains(sContent, ProtocolStartTag) || strings.Contains(sContent, LegacyProtocolStartTag) {
		return false
	}

	// Identify Legacy Content (Anything outside current protocol tags if they exist)
	startIndex := strings.Index(sContent, ProtocolStartTag)
	endIndex := strings.Index(sContent, ProtocolEndTag)

	var legacyContent string
	if startIndex != -1 && endIndex != -1 && endIndex > startIndex {
		preBlock := sContent[:startIndex]
		postBlock := sContent[endIndex+len(ProtocolEndTag):]
		legacyContent = strings.TrimSpace(preBlock + "\n" + postBlock)
	} else {
		legacyContent = strings.TrimSpace(sContent)
	}

	// If there's content to migrate, move it
	if legacyContent != "" {
		contextPath := "_rules/_orchestration/PROJECT_CONTEXT.md"
		
		// Ensure directory exists (idempotent)
		initializeOrchestration(false)

		// Read existing context
		existingContext, err := os.ReadFile(contextPath)
		if err == nil {
			// Check if we already migrated this content
			if !strings.Contains(string(existingContext), legacyContent) {
				header := fmt.Sprintf("\n\n## Legacy Content from %s (at init)\n", file)
				newContext := string(existingContext) + header + legacyContent
				_ = os.WriteFile(contextPath, []byte(newContext), 0644)
			}
		} else {
			// Create new
			header := fmt.Sprintf("\n\n## Legacy Content from %s (at init)\n", file)
			baseContent := restoreCodeBlocks(ProjectContextMdTemplate)
			newContext := baseContent + header + legacyContent
			_ = os.WriteFile(contextPath, []byte(newContext), 0644)
		}
	}

	// Overwrite agent file with ONLY the trigger
	newContent := trigger + "\n"
	if err := os.WriteFile(file, []byte(newContent), 0644); err != nil {
		return false
	}
	
	return true
}

func installDevlogHooks(verbose bool) {
	gitDir := ".git/hooks"
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		if verbose {
			fmt.Println("    Error: not a git repository")
		}
		return
	}

	hookContent := `#!/bin/sh
# Auto-sync devlogs to beads database
# Try local binary first, then global command
if [ -f "./bd" ]; then
    ./bd devlog sync >/dev/null 2>&1 &
elif command -v bd >/dev/null 2>&1; then
    bd devlog sync >/dev/null 2>&1 &
fi
`
	hooks := []string{"post-commit", "post-merge"}
	installed := []string{}
	for _, hook := range hooks {
		path := filepath.Join(gitDir, hook)
		// Read existing hook to check if we're already in it
		existing, _ := os.ReadFile(path)
		if strings.Contains(string(existing), "bd devlog sync") {
			installed = append(installed, hook)
			continue
		}

		// Append or create
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			if verbose {
				fmt.Printf("    Error installing %s: %v\n", hook, err)
			}
			continue
		}
		if _, err := f.WriteString("\n" + hookContent); err != nil {
			if verbose {
				fmt.Printf("    Error writing %s: %v\n", hook, err)
			}
		}
		f.Close()
		installed = append(installed, hook)
	}
	if verbose && len(installed) > 0 {
		fmt.Printf("    %s Installed hooks (%s)\n", ui.RenderPass("✓"), strings.Join(installed, ", "))
	}
}


// devlogOnboardCmd is deprecated and aliases to 'bd onboard'
var devlogOnboardCmd = &cobra.Command{
	Use:        "onboard",
	Short:      "Set up agent instruction files (Deprecated: use 'bd onboard')",
	Deprecated: "use 'bd onboard' instead",
	Run: func(cmd *cobra.Command, args []string) {
		// Forward to the new command
		onboardCmd.Run(cmd, args)
	},
}

// devlogSyncCmd updates the database from the filesystem
var devlogSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync devlogs to database (use -v for detailed logs)",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		// Find index file
		devlogDir, _ := store.GetConfig(rootCtx, "devlog_dir")
		if devlogDir == "" {
			// Try default locations
			if _, err := os.Stat("_rules/_devlog/_index.md"); err == nil {
				devlogDir = "_rules/_devlog"
			} else if _, err := os.Stat("index.md"); err == nil {
				devlogDir = "." // Fallback for testing
			} else {
				fmt.Fprintf(os.Stderr, "Error: devlog not configured. Run 'bd devlog initialize'\n")
				os.Exit(1)
			}
		}

		indexPath := filepath.Join(devlogDir, "_index.md")
		// Handle the case where user manually renamed/moved files or testing scenario
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			indexPath = filepath.Join(devlogDir, "index.md")
		}

		// Check index existence
		if _, err := os.Stat(indexPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading index: %v\n", err)
			os.Exit(1)
		}

		// UPGRADE: Automatically migrate existing index if it's legacy
		migrateIndexInternal(indexPath, quietFlag)

		rows, err := parseIndexMD(indexPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "🚨 **SYNTAX ERROR in %s**\n", indexPath)
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Println("\n🚀 **AI SYNTAX CORRECTION DIRECTIVE**")
			fmt.Println("The index file is corrupted. You MUST fix it to continue.")
			fmt.Println("1. **READ** the '_index.md' file to identify the corruption.")
			fmt.Println("2. **FIX** the table structure (one row per line, 5 pipes per row).")
			fmt.Println("3. **REMOVE** any duplicate headers or footers.")
			fmt.Println("4. **MANDATORY RE-RUN:** Execute 'bd devlog sync' immediately after fixing.")
			os.Exit(1)
		}

		if rows == nil || len(rows) == 0 {
			// Fresh devlog space: an empty table is normal, not an error
			if !quietFlag {
				fmt.Println("No sessions recorded yet — write a devlog and run 'bd devlog record --file <path>'.")
			}
			return
		}

		if verboseFlag {
			fmt.Printf("Scanning %d sessions...\n", len(rows))
		}
		updatedCount := 0
		for _, row := range rows {
			updated, err := SyncSession(store, row)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error syncing session %s: %v\n", row.Subject, err)
			}
			if updated {
				updatedCount++
				if verboseFlag {
					fmt.Printf("  Updated: %s\n", row.Subject)
				}
			}
		}

		// Store last sync time
		_ = store.SetMetadata(rootCtx, "last_devlog_sync", time.Now().Format(time.RFC3339))

		// GHOST RECONCILIATION: sessions whose index row was removed are never
		// visited by the row loop above, so their is_ghost flag goes stale.
		// If the file is also gone from disk, mark them ghosts here.
		indexed := make(map[string]bool, len(rows))
		for _, r := range rows {
			indexed[filepath.Base(r.Filename)] = true
		}
		db := store.UnderlyingDB()
		if sessRows, qErr := db.QueryContext(rootCtx, "SELECT id, filename FROM sessions WHERE is_ghost = 0"); qErr == nil {
			var staleIDs []string
			for sessRows.Next() {
				var id, fn string
				if sessRows.Scan(&id, &fn) != nil || indexed[filepath.Base(fn)] {
					continue
				}
				if _, err := os.Stat(fn); err == nil {
					continue // file exists at recorded path (orphan, adoptable via verify --fix)
				}
				if _, err := os.Stat(filepath.Join(devlogDir, filepath.Base(fn))); err == nil {
					continue // file exists in devlog dir
				}
				staleIDs = append(staleIDs, id)
			}
			sessRows.Close()
			for _, id := range staleIDs {
				_, _ = db.ExecContext(rootCtx, "UPDATE sessions SET is_ghost = 1, is_missing = 1 WHERE id = ?", id)
			}
			if len(staleIDs) > 0 && verboseFlag {
				fmt.Printf("Marked %d de-indexed session(s) with missing files as ghosts\n", len(staleIDs))
			}
		}

		// ORPHAN DETECTION
		orphans, _ := GetOrphanedFiles(devlogDir, rows)
		if len(orphans) > 0 {
			fmt.Printf("\n%s Found %d orphaned devlog file(s) not in index:\n", ui.RenderWarn("⚠️ "), len(orphans))
			for _, o := range orphans {
				fmt.Printf("   - %s\n", o)
			}
			fmt.Printf("   Run 'bd devlog verify --fix' to adopt them.\n")
		}

		if updatedCount > 0 {
			if !noAutoFlush {
				flushToJSONLWithState(flushState{forceDirty: true})
			}
			fmt.Printf("✅ Synced %d sessions\n", updatedCount)
		} else if !quietFlag {
			fmt.Println("Already up to date.")
		}

		// AUTO-ALIAS: merge separator-variant duplicates (ollama-extractor vs
		// ollamaextractor). Variant names stay in the registry; undo via unalias.
		if mergedCount, aErr := queries.AutoAliasDuplicates(rootCtx, db); aErr == nil && mergedCount > 0 {
			fmt.Printf("🔗 Auto-merged %d duplicate entity variant(s) (undo: bd devlog unalias <name>)\n", mergedCount)
		}

		// AUTO-PRUNE NOISE: sweep out generic prose-term entities so the graph
		// can't accumulate junk over time. Self-limiting: no-op once clean.
		if removed := pruneNoiseEntities(rootCtx, db); removed > 0 {
			fmt.Printf("🧹 Removed %d noise entity(ies) (generic prose terms)\n", removed)
		}

		// RE-APPLY MANUAL LINKS: entities were just (re)extracted, so now the
		// manual edges from links.jsonl can resolve their endpoints and be
		// re-inserted — this is what makes them survive a DB rebuild and arrive
		// from a teammate's clone (BeadsLog-58r).
		if bDir := beads.FindBeadsDir(); bDir != "" {
			linksPath := filepath.Join(bDir, "links.jsonl")
			if _, statErr := os.Stat(linksPath); statErr == nil {
				_ = importLinksFromJSONL(rootCtx, store, linksPath)
			}
		}

		// RECONCILIATION SUMMARY: surface repair needs instead of burying them in per-file warnings
		incomplete, ghosts := devlogHealthCounts(store.UnderlyingDB())
		if ghosts > 0 {
			fmt.Printf("✖ %d ghost session(s) in index — run 'bd devlog prune'\n", ghosts)
		}
		if incomplete > 0 {
			fmt.Printf("○ %d incomplete session(s) — run 'bd devlog verify --fix'\n", incomplete)
		}
		fmt.Printf("\n%s Tip: Use -v for detailed sync logs or --help for options.\n", ui.RenderAccent("💡"))
	},
}

// devlogHealthCounts returns sessions needing repair: incomplete (missing
// entities/relationships or containing placeholder text) and ghosts (index
// entries whose markdown file no longer exists on disk).
// pruneNoiseEntities removes every entity whose name is noise per the
// extraction filter (generic prose terms, fragments, truncations) along with
// its session links and edges. Returns the count removed. Cheap and
// self-limiting — once the backlog is clean, subsequent calls are no-ops — so
// it is safe to run automatically on each sync to keep the graph from
// accumulating junk (BeadsLog-4qu).
func pruneNoiseEntities(ctx context.Context, db *sql.DB) int {
	rows, err := db.QueryContext(ctx, "SELECT id, name FROM entities")
	if err != nil {
		return 0
	}
	var noiseIDs []string
	for rows.Next() {
		var id, name string
		if rows.Scan(&id, &name) == nil && extractor.IsNoise(name) {
			noiseIDs = append(noiseIDs, id)
		}
	}
	rows.Close()
	for _, id := range noiseIDs {
		_, _ = db.ExecContext(ctx, "DELETE FROM session_entities WHERE entity_id = ?", id)
		_, _ = db.ExecContext(ctx, "DELETE FROM entity_deps WHERE from_entity = ? OR to_entity = ?", id, id)
		_, _ = db.ExecContext(ctx, "DELETE FROM entities WHERE id = ?", id)
	}
	return len(noiseIDs)
}

func devlogHealthCounts(db *sql.DB) (incomplete, ghosts int) {
	// enrichment_status = 2 (ai_crystallized) is terminal: AI extraction already
	// ran, so a session still lacking entities/deps has nothing left to extract
	// and must not be flagged forever (BeadsLog-ftw). Stub-template sessions
	// stay flagged regardless — that's an authoring problem, not extraction.
	_ = db.QueryRow(`
		SELECT COUNT(DISTINCT id)
		FROM sessions
		WHERE (((id NOT IN (SELECT DISTINCT session_id FROM session_entities)
		OR id NOT IN (SELECT DISTINCT discovered_in FROM entity_deps))
		AND IFNULL(enrichment_status, 0) != 2)
		OR (narrative LIKE '%<!-- Describe the technical context -->%'
		OR narrative LIKE '%- [ ] Task 1%'))
		AND narrative NOT LIKE '%No architectural changes%'
	`).Scan(&incomplete)
	_ = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE is_ghost = 1").Scan(&ghosts)
	return
}

const devlogStubTemplate = `# %s
**Date:** %s
**Author:** %s

## Problem
%s

## Context
<!-- Describe the technical context and why this change is needed -->

## Work Done
- [ ] Task 1

## Architectural Relationships
<!-- Add explicit edges here: - EntityA -> EntityB (relationship) -->
<!-- No code/dependency change (docs, file moves, reports)? Write: _No architectural changes_ -->
`

const catchupPromptTemplate = `# Prompt: High-Signal Activity Catchup Summary

## Objective
Analyze the raw activity feed provided by ` + "`bd catchup`" + ` and generate a concise, human-readable technical summary that aligns the agent and the user on the project's current state.

## Persona
A meticulous Lead Software Engineer ensuring that architectural shifts and blocked paths are clearly communicated to prevent duplicate work or context loss.

---

## Context Analysis
The input for this prompt is the raw output from ` + "`bd catchup`" + `. 
Focus your summary on:
1.  **Activity by Team Member**: Group technical progress by the person/agent who performed it.
2.  **Landed Work**: What features or fixes were actually merged and validated.
3.  **Directional Shifts**: Any new epics, major task splits, or modularization attempts.
4.  **The Grave (Tombstones)**: Abandoned paths and the reasons WHY they were dropped.
5.  **Paused Context**: Parked branches and what is blocking them.

## Summary Requirements
The summary MUST follow this structured format:

### 🔄 Project Evolution Since [Date]

[One paragraph explaining the major themes of the recent work and any significant architectural drift.]

#### 👥 Activity by Team Member
- **[Author Name]**:
  - [High-level summary of their primary focus]
  - [Bullet points of specific landed or paused work]

#### ✨ Landed Features & Enhancements
- **[Component/Feature]**: [Concise technical summary of what was completed]
- **[Landed via [SessionID]]**: [Brief impact on the codebase]

#### ⚠️ Blocked & Abandoned Paths (Critical)
- **[Abandoned Scope]**: [The 'Why' - explain the specific failure or reason for rejection]
- **[Paused Scope]**: [What is blocking this work? Link to reasoning session if available]

#### 🎯 Suggested Next Steps
- [Based on ` + "`bd ready`" + `, what are the highest priority items to address now?]

## Success Criteria
- A human reading this summary should know exactly what they missed without opening individual devlog files.
- The summary must have a high signal-to-noise ratio: prioritize "Why" over "What".
- Blocked paths must be highly visible to prevent the agent from re-proposing failed approaches.
`

const indexTemplate = `# Development Log Index

> [!IMPORTANT]
> **AI AGENT INSTRUCTIONS:**
> 1. **RECORD SESSION:** Use ` + "`bd devlog record`" + ` to add new entries. NEVER edit this table manually.
> 2. **STAY AT BOTTOM:** The table remains the last element in this file.

This index provides a record of human-AI collaboration.

## Nomenclature Rules:
- **[fix]** - Bug fixes
- **[feature]** - New features
- **[enhance]** - Improvements
- **[rationalize]** - Cleanup
- **[deploy]** - Releases
- **[security]** - Security
- **[debug]** - Investigation
- **[test]** - Validation

## Work Index

| Subject | Problems | Author | Agent | Date | Branch | Devlog |
|---------|----------|--------|-------|------|--------|---------|
`

// genesisTemplate is the first, automatically generated devlog of a fresh
// install: an honest record of the memory system's own initialization that
// doubles as a format example for agents. Placeholders: {{DATE}}, {{AUTHOR}},
// {{REPO}}, {{BRANCH}}.
const genesisTemplate = `# [init] BeadsLog Genesis

**Date:** {{DATE}}

## Problem
This repository had no persistent memory. BeadsLog was initialized on {{DATE}} by {{AUTHOR}} to give agents and humans a shared record, versioned in git, of every session, decision, and architectural change in {{REPO}}.

### Details

- **Repository:** {{REPO}}
- **Branch:** {{BRANCH}}
- **Initialized by:** {{AUTHOR}}

This entry was generated automatically by 'bd devlog initialize'. It is the first session in this project's memory and serves as a format example: every future entry records a real working session following this structure — a Problem, the narrative of what happened, and the architectural relationships it touched.

### Architectural Relationships
- BeadsLog -> {{ARROW_TARGET}} (tracks)
`

const promptTemplateManual = `# Prompt: Generate Chronological Debugging & Development Log (Manual Mode)

## Objective:
Analyze the session history and generate a comprehensive development log. Document assumptions, actions, outcomes, and corrections.

## ⚠️ MANDATORY: Architectural Relationships
Since Background AI Enrichment is DISABLED, manually append relationships at the bottom:
` + "```markdown" + `
### Architectural Relationships
- EntityA -> EntityB (uses)
` + "```" + `
If the session touched no code/dependencies (docs, file moves, reports), write ` + "`_No architectural changes_`" + ` instead — this marks it complete so ` + "`verify`" + ` stops flagging it.

## persona:
Meticulous technical writer documenting the learning journey.

## ⏸ LIFECYCLE MANAGEMENT:
- **PAUSE:** If you must pivot away from a task before completion, run:
  ` + "`bd devlog pause --scope branch:name --message \"Reason for pivot\"`" + `
- **ABANDON:** If an approach is found to be flawed (memory leaks, regression, rejected PRD), run:
  ` + "`bd devlog abandon --scope branch:name --message \"Reason why this failed\"`" + `
- **WHY:** Mandatory reason strings prevent future agents from repeating your failed experiments.

## File Handling:
1.  **Check for Existing Log:** Find today's file in ` + "`_rules/_devlog/`" + `.
2.  **Update or Create:** Append to today's log or create ` + "`_rules/_devlog/[YYYY-MM-DD]_[title].md`" + `.
3.  **Record Session:** After saving the file, you MUST run:
    *   ` + "`bd devlog record --subject \"[prefix] description\" --problem \"description\" --file \"_rules/_devlog/[filename].md\"`" + `

## Output Structure:
---
# Comprehensive Development Log: [Goal]
**Date:** [YYYY-MM-DD]
...
`

const promptTemplateAuto = `# Prompt: Generate Chronological Debugging & Development Log (AI Enhanced)

## Objective:
Analyze the session history and generate a comprehensive development log.

## ✨ Background AI Active:
Background AI Enrichment is ENABLED. Focus strictly on the technical narrative.

## persona:
Meticulous technical writer documenting the learning journey.

## ⏸ LIFECYCLE MANAGEMENT:
- **PAUSE:** If you must pivot away from a task before completion, run:
  ` + "`bd devlog pause --scope branch:name --message \"Reason for pivot\"`" + `
- **ABANDON:** If an approach is found to be flawed (memory leaks, regression, rejected PRD), run:
  ` + "`bd devlog abandon --scope branch:name --message \"Reason why this failed\"`" + `
- **WHY:** Mandatory reason strings prevent future agents from repeating your failed experiments.

## File Handling:
1.  **Check for Existing Log:** Find today's file in ` + "`_rules/_devlog/`" + `.
2.  **Update or Create:** Append to today's log or create ` + "`_rules/_devlog/[YYYY-MM-DD]_[title].md`" + `.
3.  **Record Session:** After saving the file, you MUST run:
    *   ` + "`bd devlog record --subject \"[prefix] description\" --problem \"description\" --file \"_rules/_devlog/[filename].md\"`" + `

## Output Structure:
---
# Comprehensive Development Log: [Goal]
**Date:** [YYYY-MM-DD]
...
`

func refreshDevlogPrompt(store *sqlite.SQLiteStorage) {
	// Find devlog_dir
	devlogDir, _ := store.GetConfig(rootCtx, "devlog_dir")
	if devlogDir == "" {
		// Fallback to default
		if _, err := os.Stat("_rules/_devlog"); err == nil {
			devlogDir = "_rules/_devlog"
		} else {
			return
		}
	}

	promptPath := filepath.Join(devlogDir, "_generate-devlog.md")
	if _, err := os.Stat(promptPath); os.IsNotExist(err) {
		return
	}

	backgroundEnrich := config.GetBool("entity_extraction.background_enrichment")
	expectedTemplate := promptTemplateManual
	if backgroundEnrich {
		expectedTemplate = promptTemplateAuto
	}

	// Read current
	content, err := os.ReadFile(promptPath)
	if err != nil {
		return
	}

	// Update if outdated (missing Lifecycle Management)
	if !strings.Contains(string(content), "⏸ LIFECYCLE MANAGEMENT") {
		_ = os.WriteFile(promptPath, []byte(expectedTemplate), 0644)
		fmt.Printf("  ✨ Updated devlog instructions to include lifecycle management\n")
	}

	// Also ensure catchup prompt exists
	catchupPath := filepath.Join(devlogDir, "_generate-catchup.md")
	if _, err := os.Stat(catchupPath); os.IsNotExist(err) {
		_ = os.WriteFile(catchupPath, []byte(catchupPromptTemplate), 0644)
		fmt.Printf("  ✨ Created catchup summary instructions\n")
	} else {
		// Update if outdated
		catchupContent, err := os.ReadFile(catchupPath)
		if err == nil && !strings.Contains(string(catchupContent), "High-Signal Activity Catchup Summary") {
			_ = os.WriteFile(catchupPath, []byte(catchupPromptTemplate), 0644)
			fmt.Printf("  ✨ Updated catchup summary instructions\n")
		}
	}
}

var devlogGraphCmd = &cobra.Command{
	Use:   "graph [entity]",
	Short: "Display entity dependency graph (Fuzzy)",
	Long: `Display the architectural graph for a given entity.
This includes:
- Explicit Dependencies: Linked via arrows in devlogs (A -> B).
- Implicit Relationships: Inferred from frequent session co-occurrence.

With no entity, operates on the WHOLE graph:
  bd devlog graph --html out.html   Export the entire graph (interactive viz)
  bd devlog graph                   Print a summary (counts + most-connected)`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		depth, _ := cmd.Flags().GetInt("depth")
		strict, _ := cmd.Flags().GetBool("strict")
		limit, _ := cmd.Flags().GetInt("limit")
		relType, _ := cmd.Flags().GetString("type")
		htmlPath, _ := cmd.Flags().GetString("html")
		openBrowser, _ := cmd.Flags().GetBool("open")

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		db := store.UnderlyingDB()

		// No entity → whole-graph mode.
		if len(args) == 0 {
			runFullGraph(rootCtx, db, htmlPath, relType, openBrowser)
			return
		}
		term := args[0]

		var targets []queries.ResolvedEntity

		if strict {
			var id, name string
			err := db.QueryRowContext(rootCtx, "SELECT id, name FROM entities WHERE name = ?", term).Scan(&id, &name)
			if err == nil {
				targets = append(targets, queries.ResolvedEntity{ID: id, Name: name})
			}
		} else {
			// Fuzzy (Hybrid FTS + LIKE) via centralized resolver
			var err error
			targets, err = queries.ResolveEntities(rootCtx, db, term, limit)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error resolving entities: %v\n", err)
			}
		}

		if len(targets) == 0 {
			fmt.Printf("No entity found matching '%s'.\n", term)

			// Suggestions
			suggestions, _ := queries.SuggestEntities(rootCtx, db, term)
			if len(suggestions) > 0 {
				fmt.Println("\nDid you mean?")
				for _, s := range suggestions {
					fmt.Printf("- %s\n", s.Name)
				}
			}
			return
		}

		var matches []graphExport

		for _, t := range targets {
			graph, _ := queries.GetEntityGraphExact(rootCtx, db, t.Name, depth, relType)
			co, _ := queries.GetRelatedEntitiesByCooccurrence(rootCtx, db, t.ID, 5)

			if (graph != nil && len(graph.Nodes) > 0) || len(co) > 0 {
				matches = append(matches, graphExport{Root: t.Name, Graph: graph, Co: co})
			}
		}

		if htmlPath != "" {
			if len(matches) == 0 {
				fmt.Println("No graph or co-occurrence data to export.")
				return
			}
			if err := writeGraphHTML(htmlPath, matches, buildEntityMeta(rootCtx, db)); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing HTML graph: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ Interactive graph exported: %s\n", htmlPath)
			if openBrowser {
				openExportedGraph(htmlPath)
			}
			return
		}

		if len(matches) > 0 {
			for _, m := range matches {
				fmt.Printf("\n=== Entity: %s ===\n", ui.RenderAccent(m.Root))

				if m.Graph != nil && len(m.Graph.Nodes) > 1 { // More than just the root
					fmt.Println("\nExplicit Dependencies (Graph):")
					fmt.Println(ui.RenderEntityTree(m.Graph))
				}

				if len(m.Co) > 0 {
					fmt.Println("\nImplicit Relationships (Co-occurrence):")
					fmt.Println(ui.RenderCooccurrenceTable(m.Root, m.Co, ui.GetWidth()))
				}
			}
		} else {
			fmt.Println("No graph or co-occurrence data found.")
		}
		showAliasHints(rootCtx, db)
		showLinkHints(rootCtx, db)
		fmt.Printf("\n%s Tip: Use --help for depth and filtering options.\n", ui.RenderAccent("💡"))
	},
}

var devlogPathCmd = &cobra.Command{
	Use:   "path [entityA] [entityB]",
	Short: "Find historical path between two entities via devlog sessions",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		termA := args[0]
		termB := args[1]

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		db := store.UnderlyingDB()

		resolveWithSuggestions := func(term string) (string, string) {
			targets, _ := queries.ResolveEntities(rootCtx, db, term, 1)
			if len(targets) > 0 {
				return targets[0].ID, targets[0].Name
			}
			
			// Suggestions if not found
			suggestions, _ := queries.SuggestEntities(rootCtx, db, term)
			if len(suggestions) > 0 {
				fmt.Printf("No exact match for '%s'. Did you mean?\n", term)
				for _, s := range suggestions {
					fmt.Printf("- %s\n", s.Name)
				}
			} else {
				fmt.Printf("No entity found matching '%s'.\n", term)
			}
			return "", ""
		}

		idA, nameA := resolveWithSuggestions(termA)
		if idA == "" {
			return
		}

		idB, nameB := resolveWithSuggestions(termB)
		if idB == "" {
			return
		}

		fmt.Printf("Finding path from %s to %s...\n", ui.RenderAccent(nameA), ui.RenderAccent(nameB))

		path, err := queries.GetPathBetweenEntities(rootCtx, db, idA, idB)
		if err != nil {
			fmt.Printf("Error finding path: %v\n", err)
			return
		}

		fmt.Println(ui.RenderPath(path, ui.GetWidth()))
	},
}

var devlogListCmd = &cobra.Command{
	Use:   "list",
	Short: "List devlog sessions",
	Run: func(cmd *cobra.Command, args []string) {
		sessionType, _ := cmd.Flags().GetString("type")
		author, _ := cmd.Flags().GetString("author")
		preview, _ := cmd.Flags().GetBool("preview")
		
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		fields := "id, title, timestamp, type, COALESCE(author, 'Unknown')"
		if preview {
			fields += ", narrative"
		}
		query := fmt.Sprintf("SELECT %s FROM sessions", fields)
		var queryArgs []interface{}
		conditions := []string{"is_ghost = 0"}
		if sessionType != "" {
			conditions = append(conditions, "type = ?")
			queryArgs = append(queryArgs, sessionType)
		}
		if author != "" {
			conditions = append(conditions, "author LIKE ?")
			queryArgs = append(queryArgs, "%"+author+"%")
		}
		
		if len(conditions) > 0 {
			query += " WHERE " + strings.Join(conditions, " AND ")
		}
		query += " ORDER BY timestamp DESC"

		rows, err := store.UnderlyingDB().QueryContext(rootCtx, query, queryArgs...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing sessions: %v\n", err)
			os.Exit(1)
		}
		defer rows.Close()

		type listSession struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			Timestamp string `json:"timestamp"`
			Type      string `json:"type"`
			Author    string `json:"author"`
			Narrative string `json:"narrative,omitempty"`
		}
		var sessions []listSession
		for rows.Next() {
			var s listSession
			var narrative sql.NullString

			var err error
			if preview {
				err = rows.Scan(&s.ID, &s.Title, &s.Timestamp, &s.Type, &s.Author, &narrative)
			} else {
				err = rows.Scan(&s.ID, &s.Title, &s.Timestamp, &s.Type, &s.Author)
			}

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error scanning row: %v\n", err)
				continue
			}
			if narrative.Valid {
				s.Narrative = narrative.String
			}
			sessions = append(sessions, s)
		}

		if jsonOutput {
			outputJSON(sessions)
			return
		}

		for _, s := range sessions {
			// Parse and format timestamp
			displayTime := s.Timestamp
			if t, err := time.Parse(time.RFC3339, s.Timestamp); err == nil {
				displayTime = t.Local().Format("2006-01-02 15:04")
			}

			// Render standard line
			line := fmt.Sprintf("[%s] [%s] [%s] %s - %s",
				ui.RenderMuted(displayTime),
				ui.RenderAccent(s.ID),
				ui.RenderMuted(s.Author),
				ui.RenderMuted(s.Type),
				s.Title)
			fmt.Println(line)

			// Render preview if requested
			if preview && s.Narrative != "" {
				lines := strings.Split(s.Narrative, "\n")
				count := 0
				for _, l := range lines {
					trimmed := strings.TrimSpace(l)
					if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "---") {
						continue
					}
					// Clean markers
					clean := strings.ReplaceAll(trimmed, "**", "")

					fmt.Printf("      %s\n", ui.RenderMuted(clean))
					count++
					if count >= 2 { // 2 lines of content is enough for list
						break
					}
				}
			}
		}
		fmt.Printf("\n%s Tip: Use --preview for snippets, --type to filter, or --help for more.\n", ui.RenderAccent("💡"))
	},
}

var entitiesCmd = &cobra.Command{
	Use:   "entities",
	Short: "List top entities ranked by architectural significance",
	Long: `List top entities. By default, ranks by degree centrality (number of unique relationships).
Noise terms like CSS properties and common generic verbs are automatically filtered.`,
	Run: func(cmd *cobra.Command, args []string) {
		sortBy, _ := cmd.Flags().GetString("sort")
		limit, _ := cmd.Flags().GetInt("limit")
		minRels, _ := cmd.Flags().GetInt("min-rels")

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		db := store.UnderlyingDB()

		var query string
		if sortBy == "mentions" {
			query = `
				SELECT COALESCE(preferred_name, name), mention_count
				FROM entities
				WHERE name NOT IN ('used', 'using', 'service', 'component', 'implemented', 'updated', 'fixed', 'added')
				  AND name NOT LIKE 'border-%' AND name NOT LIKE 'padding-%' AND name NOT LIKE 'margin-%'
				ORDER BY mention_count DESC
				LIMIT ?`
		} else {
			// Default: sort by unique relationship count (centrality)
			query = `
				SELECT COALESCE(e.preferred_name, e.name), COUNT(DISTINCT ed.relationship || ed.to_entity || ed.from_entity) as rel_count
				FROM entities e
				JOIN entity_deps ed ON (e.id = ed.from_entity OR e.id = ed.to_entity)
				WHERE e.name NOT IN ('used', 'using', 'service', 'component', 'implemented', 'updated', 'fixed', 'added')
				  AND e.name NOT LIKE 'border-%' AND e.name NOT LIKE 'padding-%' AND e.name NOT LIKE 'margin-%'
				GROUP BY e.id
				HAVING rel_count >= ?
				ORDER BY rel_count DESC, e.mention_count DESC
				LIMIT ?`
		}
		var rows *sql.Rows
		var qErr error
		if sortBy == "mentions" {
			rows, qErr = db.QueryContext(rootCtx, query, limit)
		} else {
			rows, qErr = db.QueryContext(rootCtx, query, minRels, limit)
		}

		if qErr != nil {
			fmt.Fprintf(os.Stderr, "Error listing entities: %v\n", qErr)
			os.Exit(1)
		}
		defer rows.Close()

		type entityRow struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}
		var entityRows []entityRow
		var entities [][]string
		for rows.Next() {
			var e entityRow
			if err := rows.Scan(&e.Name, &e.Count); err != nil {
				continue
			}
			entityRows = append(entityRows, e)
			entities = append(entities, []string{e.Name, fmt.Sprintf("%d", e.Count)})
		}

		if jsonOutput {
			outputJSON(entityRows)
			return
		}

		if len(entities) > 0 {
			title := "Top Architectural Entities (by Relationships)"
			if sortBy == "mentions" {
				title = "Top Mentioned Entities"
			}
			fmt.Println(ui.RenderEntitiesTable(title, entities, ui.GetWidth()))
		} else {
			fmt.Println("No significant entities found.")
		}
		fmt.Printf("\n%s Tip: Use --sort=mentions to see raw frequency, or --limit to see more.\n", ui.RenderAccent("💡"))
	},
}

var devlogShowCmd = &cobra.Command{
	Use:   "show [id/date/filename]",
	Short: "Show session details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		// 1. Direct Lookup (ID, Filename, or UTC Timestamp prefix)
		timestampQuery := target
		if len(target) >= 10 { // At least YYYY-MM-DD
			// If input is "2026-01-27 01:00", convert to "2026-01-27T01:00" for LIKE
			if strings.Contains(target, " ") {
				parts := strings.Split(target, " ")
				if len(parts) == 2 {
					timestampQuery = parts[0] + "T" + parts[1]
				}
			}
		}

		rows, err := store.UnderlyingDB().QueryContext(rootCtx,
			"SELECT id, filename, timestamp, is_ghost FROM sessions WHERE id = ? OR filename LIKE ? OR timestamp LIKE ?",
			target, "%"+target+"%", timestampQuery+"%")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error querying sessions: %v\n", err)
			os.Exit(1)
		}

		type sessionMatch struct {
			ID, Filename, Timestamp string
			IsGhost                 bool
		}
		var matches []sessionMatch
		for rows.Next() {
			var m sessionMatch
			if err := rows.Scan(&m.ID, &m.Filename, &m.Timestamp, &m.IsGhost); err == nil {
				matches = append(matches, m)
			}
		}
		rows.Close()

		// 2. Fallback: Local Time Matching (if no direct match and target looks like a date)
		if len(matches) == 0 && len(target) >= 10 {
			// Query by date prefix to narrow down
			datePart := target[:10]
			rows, err := store.UnderlyingDB().QueryContext(rootCtx,
				"SELECT id, filename, timestamp, is_ghost FROM sessions WHERE timestamp LIKE ?",
				datePart+"%")
			if err == nil {
				for rows.Next() {
					var m sessionMatch
					if err := rows.Scan(&m.ID, &m.Filename, &m.Timestamp, &m.IsGhost); err == nil {
						// Convert DB timestamp to Local and check if it matches target
						if t, err := time.Parse(time.RFC3339, m.Timestamp); err == nil {
							localStr := t.Local().Format("2006-01-02 15:04")
							if strings.HasPrefix(localStr, target) {
								matches = append(matches, m)
							}
						}
					}
				}
				rows.Close()
			}
		}

		if len(matches) == 0 {
			fmt.Fprintf(os.Stderr, "Session not found.\n")
			os.Exit(1)
		}

		if len(matches) > 1 {
			// Check for exact ID match first
			for _, m := range matches {
				if m.ID == target {
					matches = []sessionMatch{m}
					goto found
				}
			}

			fmt.Println("Multiple sessions found:")
			for _, m := range matches {
				ts := m.Timestamp
				if t, err := time.Parse(time.RFC3339, m.Timestamp); err == nil {
					ts = t.Local().Format("2006-01-02 15:04")
				}
				fmt.Printf("- %s [%s] %s\n", ts, m.ID, m.Filename)
			}
			fmt.Println("\nPlease specify the ID or full filename.")
			return
		}

	found:
		match := matches[0]
		filename := match.Filename

		// Try to find the file
		// 1. As is (absolute or relative to cwd)
		content, err := os.ReadFile(filename)
		if err != nil {
			// 2. Relative to devlog_dir
			// Use base filename to prevent double-path if filename already has prefix
			devlogDir, _ := store.GetConfig(rootCtx, "devlog_dir")
			if devlogDir != "" {
				path := filepath.Join(devlogDir, filepath.Base(filename))
				content, err = os.ReadFile(path)
			}
		}

		if err != nil {
			if match.IsGhost {
				fmt.Fprintf(os.Stderr, "✖ Session %s is a ghost: its file '%s' no longer exists on disk.\n", match.ID, filename)
				fmt.Fprintf(os.Stderr, "  Run 'bd devlog prune' to clean ghost sessions.\n")
			} else {
				fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filename, err)
			}
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(struct {
				ID        string `json:"id"`
				Filename  string `json:"filename"`
				Timestamp string `json:"timestamp"`
				IsGhost   bool   `json:"is_ghost"`
				Content   string `json:"content"`
			}{match.ID, filename, match.Timestamp, match.IsGhost, string(content)})
			return
		}
		fmt.Println(string(content))
		fmt.Printf("\n%s Tip: Use --help to see how to match IDs or dates.\n", ui.RenderAccent("💡"))
	},
}

var devlogSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Hybrid search across devlog sessions",
	Long: `Search devlog sessions using a hybrid FTS5 ranking engine.
Matches are ranked based on BM25 relevance, phrase bonuses, term proximity (NEAR), and recency.
If high-confidence matches are absent, the system automatically falls back to individual keyword matching.

Examples:
  bd devlog search "auth bug"          # Search for phrase/terms
  bd devlog search "mcp" --preview     # Show 3-line snippets
  bd devlog search "rules" --explain   # See score breakdown`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		strict, _ := cmd.Flags().GetBool("strict")
		textOnly, _ := cmd.Flags().GetBool("text-only")
		limit, _ := cmd.Flags().GetInt("limit")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		author, _ := cmd.Flags().GetString("author")
		preview, _ := cmd.Flags().GetBool("preview")
		explain, _ := cmd.Flags().GetBool("explain")

		// Fallback to global json flag if not explicitly set on command
		if !cmd.Flags().Changed("json") {
			jsonOutput, _ = cmd.InheritedFlags().GetBool("json")
		}

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		db := store.UnderlyingDB()

		// Get current branch
		branchOut, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
		currentBranch := strings.TrimSpace(string(branchOut))

		// Tier 1: HybridSearch (Exact Match + Entity Expansion)
		opts := queries.SearchOptions{
			Query:         query,
			Author:        author,
			CurrentBranch: currentBranch,
			Limit:         limit,
			Strict:        strict,
			TextOnly:      textOnly,
			Preview:       preview,
			Explain:       explain,
		}
		response, err := queries.HybridSearch(rootCtx, db, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing search: %v\n", err)
			os.Exit(1)
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(response)
			return
		}

		if len(response.Results) > 0 || len(response.RelatedEntities) > 0 {
			// Fetch graph neighbors for the first matched entity if available
			var graphNeighbors []string
			if len(response.RelatedEntities) > 0 {
				primaryEntity := response.RelatedEntities[0]
				graph, err := queries.GetEntityGraphExact(rootCtx, db, primaryEntity, 1, "") // Depth 1 for neighbors
				if err == nil && graph != nil {
					for _, node := range graph.Nodes {
						// Only show direct neighbors (depth 1), not the root or deeper
						if node.Depth == 1 {
							rel := "related"
							if node.Relationship != "" {
								rel = node.Relationship
							}
							graphNeighbors = append(graphNeighbors, fmt.Sprintf("%s (%s)", node.Name, rel))
						}
					}
				}
			}

			// Found direct results (sessions and/or related entities)
			fmt.Println(ui.RenderResultsWithContext(query, convertSearchResultsToUI(response.Results), response.RelatedEntities, graphNeighbors, ui.GetWidth(), response.Strategy, explain))
			showAliasHints(rootCtx, db)
		showLinkHints(rootCtx, db)
			fmt.Printf("\n%s Tip: Use --limit to cap results, --preview for snippets, or --help for filters.\n", ui.RenderAccent("💡"))
			return
		}

		// Tier 2, 3, 4: Suggestions
		suggestions, err := queries.SuggestEntities(rootCtx, db, query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting suggestions: %v\n", err)
			return
		}

		if len(suggestions) > 0 {
			// Check if first suggestion is a typo correction
			first := suggestions[0]
			if first.Type == "typo" {
				// Perform auto-search
				correctedOpts := queries.SearchOptions{Query: first.Name, Limit: limit, Strict: strict, TextOnly: textOnly}
				correctedResponse, err := queries.HybridSearch(rootCtx, db, correctedOpts)

				if err == nil && len(correctedResponse.Results) > 0 {
					fmt.Println(ui.RenderTypoCorrection(query, first.Name, convertSearchResultsToUI(correctedResponse.Results), ui.GetWidth()))
					showAliasHints(rootCtx, db)
		showLinkHints(rootCtx, db)
					fmt.Printf("\n%s Tip: Use --limit to cap results, --preview for snippets, or --help for filters.\n", ui.RenderAccent("💡"))
				} else {
					fmt.Printf("No results found for corrected query '%s'.\n", first.Name)
				}
				return
			}

			// Just suggestions
			suggestionNames := make([]string, len(suggestions))
			for i, s := range suggestions {
				suggestionNames[i] = s.Name
			}
			fmt.Println(ui.RenderNoResults(query, suggestionNames, ui.GetWidth()))
			showAliasHints(rootCtx, db)
		showLinkHints(rootCtx, db)
			fmt.Printf("\n%s Tip: Not finding what you expect? Run %s to ingest latest logs, or use --help for search filters.\n", ui.RenderAccent("💡"), ui.RenderAccent("bd devlog sync"))
			return
		}

		fmt.Println(ui.RenderNoResults(query, nil, ui.GetWidth()))
		showAliasHints(rootCtx, db)
		showLinkHints(rootCtx, db)
		fmt.Printf("\n%s Tip: Not finding what you expect? Run %s to ingest latest logs, or use --help for search filters.\n", ui.RenderAccent("💡"), ui.RenderAccent("bd devlog sync"))
	},
}

var devlogImpactCmd = &cobra.Command{
	Use:   "impact [entity]",
	Short: "Show what depends on an entity (Fuzzy)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		term := args[0]
		strict, _ := cmd.Flags().GetBool("strict")
		limit, _ := cmd.Flags().GetInt("limit")

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		db := store.UnderlyingDB()

		var targets []queries.ResolvedEntity

		if strict {
			var id, name string
			err := db.QueryRowContext(rootCtx, "SELECT id, name FROM entities WHERE name = ?", term).Scan(&id, &name)
			if err == nil {
				targets = append(targets, queries.ResolvedEntity{ID: id, Name: name})
			}
		} else {
			// Fuzzy (Hybrid FTS + LIKE) via centralized resolver
			var err error
			targets, err = queries.ResolveEntities(rootCtx, db, term, limit)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error resolving entities: %v\n", err)
			}
		}

		if len(targets) == 0 {
			if jsonOutput {
				outputJSON([]struct{}{})
				return
			}
			fmt.Printf("No entity found matching '%s'.\n", term)

			// Suggestions
			suggestions, _ := queries.SuggestEntities(rootCtx, db, term)
			if len(suggestions) > 0 {
				fmt.Println("\nDid you mean?")
				for _, s := range suggestions {
					fmt.Printf("- %s\n", s.Name)
				}
			}
			return
		}

		type dependent struct {
			Name         string `json:"name"`
			Relationship string `json:"relationship"`
		}
		type impactResult struct {
			Entity     string      `json:"entity"`
			Dependents []dependent `json:"dependents"`
		}
		var results []impactResult
		for _, t := range targets {
			rows, err := db.QueryContext(rootCtx, `
				SELECT e.name, ed.relationship
				FROM entity_deps ed
				JOIN entities e ON ed.from_entity = e.id
				WHERE ed.to_entity = ?
			`, t.ID)

			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error querying deps for %s: %v\n", t.Name, err)
				continue
			}

			r := impactResult{Entity: t.Name, Dependents: []dependent{}}
			for rows.Next() {
				var d dependent
				rows.Scan(&d.Name, &d.Relationship)
				r.Dependents = append(r.Dependents, d)
			}
			rows.Close()
			results = append(results, r)
		}

		if jsonOutput {
			outputJSON(results)
			return
		}

		fmt.Printf("Impact of '%s' (%d matches):\n\n", term, len(targets))
		for _, r := range results {
			var deps []string
			for _, d := range r.Dependents {
				deps = append(deps, fmt.Sprintf("- %s (%s)", d.Name, d.Relationship))
			}
			fmt.Println(ui.RenderImpactTable(r.Entity, deps, ui.GetWidth()))
			fmt.Println()
		}
		fmt.Printf("\n%s Tip: Use --help for command options (limit, strict).\n", ui.RenderAccent("💡"))
	},
}

var devlogResumeCmd = &cobra.Command{
	Use:   "resume [query]",
	Short: "Resume debugging with context",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		lastN, _ := cmd.Flags().GetInt("last")
		author, _ := cmd.Flags().GetString("author")
		
		if lastN > 0 || len(args) == 0 {
			if lastN == 0 { lastN = 1 } // Default to last 1 if no arg and no flag
			
			store, err := sqlite.New(rootCtx, dbPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
				os.Exit(1)
			}
			defer store.Close()

			query := "SELECT title, narrative FROM sessions WHERE is_ghost = 0"
			var queryArgs []interface{}
			if author != "" {
				query += " AND author LIKE ?"
				queryArgs = append(queryArgs, "%"+author+"%")
			}
			query += " ORDER BY timestamp DESC LIMIT ?"
			queryArgs = append(queryArgs, lastN)

			rows, err := store.UnderlyingDB().QueryContext(rootCtx, query, queryArgs...)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching sessions: %v\n", err)
				os.Exit(1)
			}
			defer rows.Close()

			type resumeSession struct {
				Title     string `json:"title"`
				Narrative string `json:"narrative"`
			}
			var sessions []resumeSession
			for rows.Next() {
				var s resumeSession
				rows.Scan(&s.Title, &s.Narrative)
				sessions = append(sessions, s)
			}

			if jsonOutput {
				outputJSON(sessions)
				return
			}

			fmt.Printf("Resuming last %d session(s):\n\n", lastN)
			for _, s := range sessions {
				fmt.Printf("=== %s ===\n%s\n\n", s.Title, s.Narrative)
			}
			return
		}

		query := args[0]
		if !jsonOutput {
			fmt.Printf("Resuming context for: %s\n", query)

			// Check for proximity warnings
			store, err := sqlite.New(rootCtx, dbPath)
			if err == nil {
				defer store.Close()
				db := store.UnderlyingDB()
				warnings, _ := queries.CheckProximityWarnings(rootCtx, db, types.ScopeSession, query)
				for _, w := range warnings {
					badge := ""
					if w.State == types.StatePaused {
						badge = ui.TableWarningStyle.Render("[⏸ PAUSED]")
					} else if w.State == types.StateAbandoned {
						badge = ui.TableFailStyle.Render("[🚫 ABANDONED]")
					}
					fmt.Printf("\n⚠️  %s This work is %s: %s\n", badge, w.State, w.ShortReason)
					fmt.Printf("   Scope: %s:%s\n", w.ScopeType, w.ScopeRef)
					fmt.Printf("   Reasoning session: %s\n\n", w.FullReasonID)
				}
			}
		}

		// Search sessions and show latest (search handles --json itself)
		devlogSearchCmd.Run(cmd, args)
		if !jsonOutput {
			fmt.Printf("\n%s Tip: Use --help to customize context window size.\n", ui.RenderAccent("💡"))
		}
	},
}

// installHooksCmd installs git hooks for auto-sync
var installHooksCmd = &cobra.Command{
	Use:   "install-hooks",
	Short: "Install git hooks for auto-sync",
	Run: func(cmd *cobra.Command, args []string) {
		gitDir := ".git/hooks"
		if _, err := os.Stat(".git"); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: not a git repository\n")
			os.Exit(1)
		}

		hookContent := `#!/bin/sh
# Auto-sync devlogs to beads database
# Try local binary first, then global command
if [ -f "./bd" ]; then
    ./bd devlog sync >/dev/null 2>&1 &
elif command -v bd >/dev/null 2>&1; then
    bd devlog sync >/dev/null 2>&1 &
fi
`
		hooks := []string{"post-commit", "post-merge"}
		for _, hook := range hooks {
			path := filepath.Join(gitDir, hook)
			// Read existing hook to check if we're already in it
			existing, _ := os.ReadFile(path)
			if strings.Contains(string(existing), "bd devlog sync") {
				fmt.Printf("Hook %s already installed\n", hook)
				continue
			}

			// Append or create
			f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error installing %s: %v\n", hook, err)
				continue
			}
			if _, err := f.WriteString("\n" + hookContent); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", hook, err)
			}
			f.Close()
			fmt.Printf("Installed %s hook\n", hook)
		}
	},
}




// devlogStatusCmd shows current devlog configuration and stats
var devlogStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show devlog system status and configuration",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Println("Beads database not initialized.")
			return
		}
		defer store.Close()

		devlogDir, _ := store.GetConfig(rootCtx, "devlog_dir")
		lastSync, _ := store.GetMetadata(rootCtx, "last_devlog_sync")

		if lastSync != "" {
			if t, err := time.Parse(time.RFC3339, lastSync); err == nil {
				lastSync = t.Local().Format("2006-01-02 at 15h04m05s")
			}
		} else {
			lastSync = "(never)"
		}

		if devlogDir == "" {
			if jsonOutput {
				outputJSON(map[string]bool{"configured": false})
			} else {
				fmt.Println("\nDevlog System Status")
				fmt.Println("====================")
				fmt.Println("Status: Not configured")
				fmt.Println("Action: Run 'bd devlog initialize' to set up a devlog space.")
			}
			return
		}

		// Get stats
		db := store.UnderlyingDB()
		var sessionsCount, entitiesCount, relationshipsCount int

		_ = db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessionsCount)
		_ = db.QueryRow("SELECT COUNT(*) FROM entities").Scan(&entitiesCount)
		_ = db.QueryRow("SELECT COUNT(*) FROM entity_deps").Scan(&relationshipsCount)
		incompleteCount, ghostCount := devlogHealthCounts(db)

		var optimized, pending, failed, unenriched int
		_ = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE enrichment_status = 2").Scan(&optimized)
		_ = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE enrichment_status = 1").Scan(&pending)
		_ = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE enrichment_status = 3").Scan(&failed)
		_ = db.QueryRow("SELECT COUNT(*) FROM sessions WHERE enrichment_status IS NULL OR enrichment_status = 0").Scan(&unenriched)

		// Check hooks
		hookStatus := map[string]bool{}
		for _, h := range []string{"post-commit", "post-merge"} {
			installed := false
			if content, err := os.ReadFile(filepath.Join(".git/hooks", h)); err == nil {
				installed = strings.Contains(string(content), "bd devlog sync")
			}
			hookStatus[h] = installed
		}

		if jsonOutput {
			outputJSON(struct {
				Configured    bool            `json:"configured"`
				SpaceDir      string          `json:"space_dir"`
				LastSync      string          `json:"last_sync"`
				Sessions      int             `json:"sessions"`
				Incomplete    int             `json:"incomplete"`
				Ghosts        int             `json:"ghosts"`
				Entities      int             `json:"entities"`
				Relationships int             `json:"relationships"`
				Optimized     int             `json:"optimized"`
				Pending       int             `json:"pending"`
				Failed        int             `json:"failed"`
				Unenriched    int             `json:"unenriched"`
				Hooks         map[string]bool `json:"hooks"`
			}{true, devlogDir, lastSync, sessionsCount, incompleteCount, ghostCount,
				entitiesCount, relationshipsCount, optimized, pending, failed, unenriched, hookStatus})
			return
		}

		fmt.Println("\nDevlog System Status")
		fmt.Println("====================")
		fmt.Printf("Space Directory: %s\n", devlogDir)
		fmt.Printf("Last Sync:       %s\n", lastSync)

		fmt.Printf("\nDatabase Statistics:\n")
		fmt.Printf("  Sessions:      %d\n", sessionsCount)
		if incompleteCount > 0 {
			fmt.Printf("  ○ Incomplete:  %d (run 'bd devlog verify --fix')\n", incompleteCount)
		}
		if ghostCount > 0 {
			fmt.Printf("  ✖ Ghosts:      %d (run 'bd devlog prune')\n", ghostCount)
		}
		fmt.Printf("  Entities:      %d\n", entitiesCount)
		fmt.Printf("  Relationships: %d\n", relationshipsCount)

		// Enrichment Stats
		fmt.Printf("\nEnrichment Status (AI):\n")
		fmt.Printf("  ● Optimized: %d\n", optimized)
		if pending > 0 {
			fmt.Printf("  ○ Pending:   %d (in queue)\n", pending)
		}
		if failed > 0 {
			fmt.Printf("  ✖ Failed:    %d\n", failed)
		}
		if unenriched > 0 {
			fmt.Printf("  ○ Unenriched: %d (never queued)\n", unenriched)
		}
		// Only claim full health when nothing needs repair anywhere
		if pending == 0 && failed == 0 && unenriched == 0 && incompleteCount == 0 && ghostCount == 0 && optimized > 0 {
			fmt.Printf("  ✨ All memory optimized\n")
		} else if incompleteCount > 0 || ghostCount > 0 {
			fmt.Printf("\n⚠️  Memory health: %d incomplete, %d ghost(s) — see hints above.\n", incompleteCount, ghostCount)
		}
		fmt.Printf("\nGit Hooks:\n")
		for _, h := range []string{"post-commit", "post-merge"} {
			status := "✗"
			if hookStatus[h] {
				status = "✓"
			}
			fmt.Printf("  %s %s\n", status, h)
		}

		fmt.Printf("\n%s Tip: Use --help for detailed configuration info.\n", ui.RenderAccent("💡"))
		fmt.Println()
	},
}

var devlogResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Clear all devlog data (sessions, entities, relationships)",
	Run: func(cmd *cobra.Command, args []string) {
		if !ui.IsTerminal() {
			fmt.Println("Error: interactive confirmation required for reset. Use --force if needed (not implemented yet).")
			os.Exit(1)
		}

		confirmStr := "false"
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Reset Devlog Database?").
					Description("This will delete ALL sessions, entities, and architectural relationships from the local database. Files on disk will NOT be touched.").
					Options(
						huh.NewOption("Yes, Reset Database", "true"),
						huh.NewOption("No, Cancel", "false"),
					).
					Value(&confirmStr),
			),
		)

		if err := form.Run(); err != nil || confirmStr != "true" {
			fmt.Println("Reset cancelled.")
			return
		}

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		db := store.UnderlyingDB()
		
		tables := []string{"session_entities", "entity_deps", "sessions", "entities"}
		for _, table := range tables {
			if _, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)); err != nil {
				fmt.Fprintf(os.Stderr, "Error clearing table %s: %v\n", table, err)
			}
		}

		// Reset sync metadata so next sync works
		_ = store.SetMetadata(rootCtx, "last_devlog_sync", "")

		fmt.Println("✅ Devlog database reset.")
		fmt.Println("Run 'bd devlog sync' to re-import from files.")
	},
}

// devlogVerifyCmd identifies sessions missing metadata
// devlogExtractCmd performs immediate foreground regeneration of entities/relationships
var devlogExtractCmd = &cobra.Command{
	Use:   "extract [target]",
	Short: "Regenerate entities and relationships for a session (Foreground AI)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := ""
		if len(args) > 0 {
			target = args[0]
		}

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		db := store.UnderlyingDB()

		query := "SELECT id, title, filename, narrative FROM sessions"
		var queryArgs []interface{}
		if target != "" {
			query += " WHERE id = ? OR filename LIKE ?"
			queryArgs = append(queryArgs, target, "%"+target+"%")
		} else {
			// If no target, extract from the most recent session
			query += " ORDER BY timestamp DESC LIMIT 1"
		}

		rows, err := db.Query(query, queryArgs...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error querying sessions: %v\n", err)
			os.Exit(1)
		}
		defer rows.Close()

		found := false
		for rows.Next() {
			var id, title, filename, narrative string
			if err := rows.Scan(&id, &title, &filename, &narrative); err != nil {
				continue
			}
			found = true

			fmt.Printf("Extracting from %s (%s)...\n", id, title)
			result, err := extractAndLinkEntities(store, id, narrative, ExtractionOptions{ForceRegex: false})
			if err != nil {
				fmt.Fprintf(os.Stderr, "  ✗ Extraction failed: %v\n", err)
				continue
			}

			// Resolve path for write-back
			path := filename
			if !filepath.IsAbs(path) {
				devlogDir, _ := store.GetConfig(rootCtx, "devlog_dir")
				if devlogDir == "" { devlogDir = "_rules/_devlog" }
				path = filepath.Join(devlogDir, path)
			}

			if err := crystallizeRelationships(path, result.Relationships); err != nil {
				fmt.Printf("    Warning: failed to crystallize: %v\n", err)
			}

			// Update status to 2 (Done)
			_, _ = db.Exec("UPDATE sessions SET enrichment_status = 2 WHERE id = ?", id)
			fmt.Printf("  ✓ Extraction complete (%d entities, %d rels)\n", len(result.Entities), len(result.Relationships))
		}

		if !found {
			fmt.Println("No session found matching target.")
		}
	},
}

// devlogEnrichCmd schedules background AI enrichment
var devlogEnrichCmd = &cobra.Command{
	Use:   "enrich [target]",
	Short: "Schedule sessions for background AI enrichment (Daemon)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		target := ""
		if len(args) > 0 {
			target = args[0]
		}

		if !all && target == "" {
			fmt.Println("Please specify a [target] or use --all")
			return
		}

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		db := store.UnderlyingDB()

		var res sql.Result
		if all {
			fmt.Println("Scheduling ALL sessions for background enrichment...")
			res, err = db.Exec("UPDATE sessions SET enrichment_status = 1")
		} else {
			fmt.Printf("Scheduling session '%s' for background enrichment...\n", target)
			res, err = db.Exec("UPDATE sessions SET enrichment_status = 1 WHERE id = ? OR filename LIKE ?", target, "%"+target+"%")
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating status: %v\n", err)
			os.Exit(1)
		}

		affected, _ := res.RowsAffected()
		fmt.Printf("✓ %d sessions queued. Run 'bd status' to monitor progress.\n", affected)
	},
}

var devlogVerifyCmd = &cobra.Command{
	Use:   "verify [target]",
	Short: "Audit sessions for missing architectural metadata (entities/relationships)",
	Run: func(cmd *cobra.Command, args []string) {
		fix, _ := cmd.Flags().GetBool("fix")
		fixRegex, _ := cmd.Flags().GetBool("fix-regex")
		fixAI, _ := cmd.Flags().GetBool("fix-ai")
		idStability, _ := cmd.Flags().GetBool("id-stability")

		// Priority: fix-ai > fix-regex > fix
		forceRegex := true // Default to regex for --fix
		if fixAI {
			fix = true
			forceRegex = false
		} else if fixRegex {
			fix = true
			forceRegex = true
		}

		target := ""
		if len(args) > 0 {
			target = args[0]
		}

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		db := store.UnderlyingDB()

		devlogDir, _ := store.GetConfig(rootCtx, "devlog_dir")
		if devlogDir == "" {
			devlogDir = "_rules/_devlog"
		}
		indexPath := filepath.Join(devlogDir, "_index.md")

		// 0. Path Normalization (Fix "double path" bug in index/DB)
		if fix && target == "" {
			normalizePathsInternal(store, indexPath, false)
		}

		// 0.1 ID Stability (Backfill IDs into index links)
		if idStability && target == "" {
			backfillIDsInternal(store, indexPath, false)
		}
		// 1. Orphaned Files Detection (Disk vs Index)
		// ... (Orphan detection logic remains, but we skip it if targeting a specific session ID)
		var orphans []string
		if target == "" {
			indexedFiles := make(map[string]bool)
			if rows, err := parseIndexMD(indexPath); err == nil {
				for _, r := range rows {
					indexedFiles[filepath.Base(r.Filename)] = true
				}
			}

			files, _ := os.ReadDir(devlogDir)
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") && !strings.HasPrefix(f.Name(), "_") {
					if !indexedFiles[f.Name()] {
						orphans = append(orphans, f.Name())
					}
				}
			}

			if len(orphans) > 0 {
				fmt.Println("Found orphaned devlog files (on disk but not in _index.md):")
				for _, o := range orphans {
					fmt.Printf("- %s\n", o)
				}
				fmt.Println()
			}
		}

		// 1. Sessions marked as missing
		// ... (Skip missing check if targeting)

		// 2. Sessions without entities OR relationships (INCOMPLETE)
		// IFNULL(enrichment_status,0) != 2 — AI-crystallized sessions are
		// terminal even with zero edges (BeadsLog-ftw): AI already ran.
		query := `
			SELECT DISTINCT id, title, filename, narrative
			FROM sessions
			WHERE (id NOT IN (SELECT DISTINCT session_id FROM session_entities)
			OR id NOT IN (SELECT DISTINCT discovered_in FROM entity_deps))
			AND IFNULL(enrichment_status, 0) != 2
			AND narrative NOT LIKE '%No architectural changes%'
		`
		var queryArgs []interface{}
		
		if target != "" {
			query += " AND (id = ? OR filename LIKE ?)"
			queryArgs = append(queryArgs, target, "%"+target+"%")
		}

		rows, err := db.Query(query, queryArgs...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error querying sessions: %v\n", err)
			return
		}
		defer rows.Close()

		var incomplete []struct{ ID, Title, Filename string }
		for rows.Next() {
			var s struct{ ID, Title, Filename, Narrative string }
			rows.Scan(&s.ID, &s.Title, &s.Filename, &s.Narrative)
			incomplete = append(incomplete, struct{ ID, Title, Filename string }{s.ID, s.Title, s.Filename})
		}

		if len(incomplete) == 0 {
			if target != "" {
				fmt.Printf("✅ Session '%s' has linked entities and relationships (or doesn't exist/isn't incomplete).\n", target)
			} else {
				fmt.Println("✅ All sessions have linked entities and relationships.")
			}
			return
		}

		if !fix {
			fmt.Println("Sessions missing entities or architectural relationships:")
			for _, s := range incomplete {
				fmt.Printf("- [%s] %s (%s)\n", s.ID, s.Title, s.Filename)
			}

			fmt.Printf("\nFound %d sessions with missing metadata.\n", len(incomplete))
			if len(orphans) > 0 {
				fmt.Printf("Found %d orphaned files.\n", len(orphans))
			}
			fmt.Printf("\n%s Tip: Run 'bd devlog verify --fix' to automatically adopt orphans and backfill metadata.\n", ui.RenderAccent("💡"))
			fmt.Println("     Use '--fix-regex' for fast extraction or '--fix-ai' for high quality.")
		} else {
			// Phase 1: Adopt Orphans
			if target == "" && len(orphans) > 0 {
				fmt.Printf("Adopting %d orphaned files...\n", len(orphans))
				devlogDir, _ := store.GetConfig(rootCtx, "devlog_dir")
				if devlogDir == "" { devlogDir = "_rules/_devlog" }
				indexPath := filepath.Join(devlogDir, "_index.md")
				
				for _, o := range orphans {
					title := strings.TrimSuffix(o, ".md")
					date := time.Now().Format("2006-01-02")
					if len(o) >= 10 {
						if _, err := time.Parse("2006-01-02", o[:10]); err == nil {
							date = o[:10]
							title = strings.TrimSpace(strings.ReplaceAll(title[10:], "-", " "))
							if title == "" { title = "Adopted session" }
						}
					}
					entry := fmt.Sprintf("| [adopt] %s | Automatically adopted during verify | Unknown | Unknown | %s | N/A | [%s](%s) |\n", title, date, o, o)
					f, err := os.OpenFile(indexPath, os.O_APPEND|os.O_WRONLY, 0644)
					if err == nil {
						stat, _ := f.Stat()
						if stat.Size() > 0 {
							lastChar := make([]byte, 1)
							_, _ = f.ReadAt(lastChar, stat.Size()-1)
							if lastChar[0] != '\n' { f.WriteString("\n") }
						}
						f.WriteString(entry)
						f.Close()
						fmt.Printf("  ✓ Adopted %s\n", o)
					}
				}
				fmt.Println("  Tip: Run 'bd devlog sync' to ingest newly adopted files.")
			}

			if len(incomplete) > 0 {
				// Disclaimer
				if fixAI {
					// Check if AI is enabled
					if config.GetBool("entity_extraction.enabled") && config.GetString("entity_extraction.primary_extractor") == "ollama" {
						fmt.Printf("\n⚠️  Backfilling %d sessions with AI. This may take a while.\n", len(incomplete))
						fmt.Println("    (Use Ctrl+C to abort, or run without '--fix-ai' for instant regex results)")
						fmt.Println()
					}
				}

				fmt.Printf("Backfilling metadata for %d sessions...\n", len(incomplete))
				for _, s := range incomplete {
					var narrative string
					err := db.QueryRow("SELECT narrative FROM sessions WHERE id = ?", s.ID).Scan(&narrative)
					if err != nil {
						fmt.Printf("  ✗ Failed to retrieve narrative for %s: %v\n", s.ID, err)
						continue
					}

					fmt.Printf("  → Processing %s (%s)...\n", s.ID, s.Title)
					result, err := extractAndLinkEntities(store, s.ID, narrative, ExtractionOptions{ForceRegex: forceRegex})
					if err == nil && result != nil {
						// Resolve path for write-back
						path := s.Filename
						if !filepath.IsAbs(path) {
							// Try devlog_dir first
							devlogDir, _ := store.GetConfig(rootCtx, "devlog_dir")
							if devlogDir == "" { devlogDir = "_rules/_devlog" }
							path = filepath.Join(devlogDir, path)
						}
						
						if err := crystallizeRelationships(path, result.Relationships); err != nil {
							fmt.Printf("    Warning: failed to crystallize relationships: %v\n", err)
						}

						// Update status to mark AI done if AI was used
						if !forceRegex {
							_, _ = db.Exec("UPDATE sessions SET enrichment_status = 2 WHERE id = ?", s.ID)
						}
					}
				}

				// Honest completion report (BeadsLog-ftw): regex extraction can be
				// fully noise-filtered, leaving sessions incomplete no matter how
				// often --fix runs. Re-check and say so instead of claiming success.
				var remaining int
				_ = db.QueryRow(`
					SELECT COUNT(DISTINCT id)
					FROM sessions
					WHERE (id NOT IN (SELECT DISTINCT session_id FROM session_entities)
					OR id NOT IN (SELECT DISTINCT discovered_in FROM entity_deps))
					AND IFNULL(enrichment_status, 0) != 2
					AND narrative NOT LIKE '%No architectural changes%'
				`).Scan(&remaining)
				if remaining > 0 {
					fmt.Printf("⚠️  %d session(s) still incomplete: their extracted entities/relationships were filtered as noise.\n", remaining)
					fmt.Println("    Regex extraction cannot clear these. Instead:")
					fmt.Println("    • Run 'bd devlog verify --fix-ai' or 'bd devlog enrich --all' (AI extraction), or")
					fmt.Println("    • Add the '_No architectural changes_' marker to sessions with genuinely no edges.")
				} else {
					fmt.Println("✅ Backfill complete.")
				}
			}

			if len(orphans) == 0 && len(incomplete) == 0 {
				fmt.Println("Nothing to fix.")
			}
		}
	},
}

var devlogRecordCmd = &cobra.Command{
	Use:   "record",
	Short: "Automatically record a devlog entry in the index",
	Run: func(cmd *cobra.Command, args []string) {
		file, _ := cmd.Flags().GetString("file")
		subject, _ := cmd.Flags().GetString("subject")
		problem, _ := cmd.Flags().GetString("problem")
		author, _ := cmd.Flags().GetString("author")
		agent, _ := cmd.Flags().GetString("agent")

		if file == "" {
			fmt.Println("Error: --file is mandatory.")
			os.Exit(1)
		}

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		devlogDir, _ := store.GetConfig(rootCtx, "devlog_dir")
		if devlogDir == "" {
			devlogDir = "_rules/_devlog"
		}

		// 1. Verify file existence
		targetFile := file
		if _, err := os.Stat(targetFile); os.IsNotExist(err) {
			// Try relative to devlogDir
			relPath := filepath.Join(devlogDir, file)
			if _, err := os.Stat(relPath); os.IsNotExist(err) {
				fmt.Printf("\n%s Error: Devlog file not found: %s\n", ui.RenderFail("🚨"), file)
				fmt.Println("\n🚀 **AI ACTION REQUIRED:**")
				fmt.Println("You MUST create and write the full devlog Markdown file FIRST, then run this command.")
				fmt.Println("Follow the rules in '_rules/_devlog/_generate-devlog.md'.")
				os.Exit(1)
			}
			targetFile = relPath
		}

		indexPath := filepath.Join(devlogDir, "_index.md")

		// 2. File Collision Check
		// Prevent registering the same file twice
		if indexContent, err := os.ReadFile(indexPath); err == nil {
			fileName := filepath.Base(targetFile)
			// Check if the filename appears inside markdown link syntax e.g. [name](name)
			if strings.Contains(string(indexContent), "]("+fileName) {
				fmt.Printf("\n%s Error: File Collision Detected\n", ui.RenderFail("🚨"))
				fmt.Printf("The file '%s' is already registered in the devlog index.\n", fileName)
				fmt.Println("You MUST use a unique filename for each session.")
				os.Exit(1)
			}
		}

		// 3. Auto-Extract Metadata from file if missing
		if subject == "" || problem == "" {
			extractedSubject, extractedProblem, err := extractDevlogMetadata(targetFile)
			if err == nil {
				if subject == "" && extractedSubject != "" {
					subject = extractedSubject
				}
				if problem == "" && extractedProblem != "" {
					problem = extractedProblem
				}
			}
		}

		// 3. Mandatory checks
		if subject == "" || problem == "" {
			fmt.Println("Error: Could not extract Subject or Problem from file.")
			fmt.Println("Ensure your file has a '# Subject' and a '## Problem' section, or use --subject and --problem flags.")
			os.Exit(1)
		}

		// 4. Resolve Author/Agent
		if author == "" {
			author = config.GetString("devlog.author")
		}
		if author == "" {
			author = os.Getenv("BD_ACTOR")
		}
		if author == "" {
			out, _ := exec.Command("git", "config", "user.name").Output()
			author = strings.TrimSpace(string(out))
		}
		if author == "" {
			author = "Unknown"
		}

		if agent == "" {
			agent = os.Getenv("BD_AGENT_NAME")
		}
		if agent == "" {
			agent = "Unknown"
		}
		
		// 5. Get Date
		date := time.Now().Format("2006-01-02 15:04")

		// 6. Generate Session ID (for stability)
		sessionID := fmt.Sprintf("sess-%s", hashID(subject+date))

		// 7. Get Branch & Commit Info
		branchInfo := "N/A"
		commitSHA := ""
		if config.GetBool("devlog.branch-tracking") {
			// Get current branch
			branchOut, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
			branchName := strings.TrimSpace(string(branchOut))
			if branchName != "" {
				// Check for upstream
				upstreamOut, err := exec.Command("git", "rev-parse", "--abbrev-ref", "@{u}").Output()
				if err == nil && strings.TrimSpace(string(upstreamOut)) != "" {
					branchInfo = branchName
				} else {
					branchInfo = branchName + " (local)"
				}
			}

			// Get current commit
			commitOut, _ := exec.Command("git", "rev-parse", "HEAD").Output()
			commitSHA = strings.TrimSpace(string(commitOut))
		}

		// 8. Format Row
		fileName := filepath.Base(targetFile)
		devlogLink := fmt.Sprintf("[%s](%s?id=%s", fileName, fileName, sessionID)
		if commitSHA != "" {
			devlogLink += "&sha=" + commitSHA
		}
		devlogLink += ")"
		
		row := fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |\n", subject, problem, author, agent, date, branchInfo, devlogLink)

		// 9. Append to File
		f, err := os.OpenFile(indexPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			fmt.Printf("Error opening index: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		if _, err := f.WriteString(row); err != nil {
			fmt.Printf("Error writing to index: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Recorded session: %s\n", subject)

		// 10. Sync
		fmt.Println("→ Syncing devlog...")
		devlogSyncCmd.Run(cmd, []string{})
	},
}

var devlogAliasCmd = &cobra.Command{
	Use:   "alias [target] [alias1,alias2,...]",
	Short: "Collapse fragmented entities into a single canonical entity",
	Long: `Collapse one or more fragmented entities into a target canonical entity.
This merges all session mentions and architectural relationships in the database.

Sticky Aliasing:
The mapping is saved in a persistent registry inside the BeadsLog database. 
This means the alias will survive 'bd devlog sync' and 'verify --fix' operations.
However, because this is a database-level mapping, if you completely delete the 
.beads/beads.db file and re-initialize the workspace, the aliases will be lost.

Example:
  bd devlog alias AuthService auth-service,auth-provider`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		targetName := args[0]
		aliasNames := strings.Split(strings.Join(args[1:], ","), ",")

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		db := store.UnderlyingDB()

		// Resolve target
		targets, _ := queries.ResolveEntities(rootCtx, db, targetName, 1)
		if len(targets) == 0 {
			fmt.Printf("Error: target entity '%s' not found.\n", targetName)
			return
		}
		target := targets[0]

		var aliasIDs []string
		var resolvedAliases []queries.ResolvedEntity
		for _, name := range aliasNames {
			trimmed := strings.TrimSpace(name)
			if trimmed == "" {
				continue
			}
			aliases, _ := queries.ResolveEntities(rootCtx, db, trimmed, 1)
			if len(aliases) == 0 {
				fmt.Printf("Warning: alias entity '%s' not found, skipping.\n", trimmed)
				continue
			}
			if aliases[0].ID == target.ID {
				continue
			}
			aliasIDs = append(aliasIDs, aliases[0].ID)
			resolvedAliases = append(resolvedAliases, aliases[0])
		}

		if len(aliasIDs) == 0 {
			fmt.Println("No valid aliases found to merge.")
			return
		}

		var names []string
		for _, a := range resolvedAliases {
			names = append(names, a.Name)
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			fmt.Printf("\n%s Proposed Collapse (Dry Run):\n", ui.RenderAccent("🔍"))
			fmt.Printf("   Target: %s\n", target.Name)
			fmt.Printf("   Aliases to merge: %s\n", strings.Join(names, ", "))

			// Show impact stats
			var sessionsTotal, relsTotal int
			for _, a := range resolvedAliases {
				var sessions, rels int
				_ = db.QueryRow("SELECT COUNT(*) FROM session_entities WHERE entity_id = ?", a.ID).Scan(&sessions)
				_ = db.QueryRow("SELECT COUNT(*) FROM entity_deps WHERE from_entity = ? OR to_entity = ?", a.ID, a.ID).Scan(&rels)
				sessionsTotal += sessions
				relsTotal += rels
			}
			fmt.Printf("\nImpact:\n")
			fmt.Printf("  • %d session mentions will be moved to %s\n", sessionsTotal, target.Name)
			fmt.Printf("  • %d architectural relationships will be moved to %s\n", relsTotal, target.Name)
			fmt.Printf("\nRun without '--dry-run' to apply these changes.\n")
			return
		}

		fmt.Printf("Aliasing %s → %s...\n", strings.Join(names, ", "), target.Name)

		if err := queries.AliasEntities(rootCtx, db, target.ID, resolvedAliases); err != nil {
			fmt.Fprintf(os.Stderr, "Error aliasing entities: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ Entities collapsed successfully. Aliases are now sticky.")
		flushMetadata(rootCtx)
	},
}

var devlogUnaliasCmd = &cobra.Command{
	Use:   "unalias [alias_name]",
	Short: "Remove an alias mapping and restore the original entity on next sync",
	Long: `Remove an alias mapping from the persistent registry.

Once removed, you must run 'bd devlog verify --fix' to re-parse the 
underlying Markdown devlogs. This will extract the original entity name 
from the text and restore its session links and architectural dependencies.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		aliasName := strings.TrimSpace(args[0])

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		db := store.UnderlyingDB()

		if err := queries.UnaliasEntity(rootCtx, db, aliasName); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing alias: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Alias '%s' removed. Run 'bd devlog verify --fix' to restore the original entity links.\n", aliasName)
		flushMetadata(rootCtx)
	},
}

var devlogAliasesCmd = &cobra.Command{
	Use:   "aliases",
	Short: "Review and manage alias suggestions",
	Long: `Review pending alias suggestions and manage the suggestion queue.

Suggestions are entity pairs with both high session co-occurrence AND similar
names. Merge a pair with 'bd devlog alias <canonical> <variant>', or reject it
with 'bd devlog aliases dismiss <a> <b>' so it never resurfaces.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return devlogAliasesSuggestCmd.RunE(cmd, args)
	},
}

var devlogAliasesSuggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "List pending alias suggestions, best candidates first",
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer store.Close()
		db := store.UnderlyingDB()

		suggestions, err := queries.GetAliasSuggestions(rootCtx, db, 0.8, limit)
		if err != nil {
			return fmt.Errorf("fetching suggestions: %w", err)
		}

		if jsonOutput {
			outputJSON(suggestions)
			return nil
		}

		if len(suggestions) == 0 {
			fmt.Println("✅ No pending alias suggestions.")
			return nil
		}

		fmt.Printf("Alias suggestions (%d shown, ranked by combined score):\n\n", len(suggestions))
		for _, s := range suggestions {
			fmt.Printf("  %s ↔ %s  (overlap %.0f%%, name %.0f%%)\n", s.EntityA, s.EntityB, s.Similarity*100, s.NameSimilarity*100)
			fmt.Printf("    merge:   bd devlog alias %s %s\n", s.EntityA, s.EntityB)
			fmt.Printf("    reject:  bd devlog aliases dismiss %s %s\n\n", s.EntityA, s.EntityB)
		}
		fmt.Println("💡 Use --limit N for more, --json for machine-readable output.")
		return nil
	},
}

var devlogAliasesDismissCmd = &cobra.Command{
	Use:   "dismiss [entityA] [entityB]",
	Short: "Reject a suggestion so this pair is never suggested again",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer store.Close()

		if err := queries.DismissAliasPair(rootCtx, store.UnderlyingDB(), args[0], args[1], actor); err != nil {
			return fmt.Errorf("recording dismissal: %w", err)
		}
		fmt.Printf("✅ Dismissed: '%s' ↔ '%s' will not be suggested again.\n", args[0], args[1])
		return nil
	},
}

// linkMinCooccur is the minimum shared sessions for a relationship suggestion.
const linkMinCooccur = 4

// showLinkHints prints a one-line count of pending relationship suggestions —
// entity pairs that co-occur strongly but have no explicit edge yet.
// filterNoiseLinks drops suggestions where either endpoint is a noise entity —
// a safety net so a DB that predates the current stoplist (not yet re-pruned)
// still gets clean suggestions instead of "user ↔ uses" junk (BeadsLog-rob).
func filterNoiseLinks(sugs []queries.LinkSuggestion) []queries.LinkSuggestion {
	out := sugs[:0]
	for _, s := range sugs {
		if extractor.IsNoise(s.EntityA) || extractor.IsNoise(s.EntityB) {
			continue
		}
		out = append(out, s)
	}
	return out
}

func showLinkHints(ctx context.Context, db *sql.DB) {
	suggestions, err := queries.GetLinkSuggestions(ctx, db, linkMinCooccur, 0)
	if err != nil {
		return
	}
	suggestions = filterNoiseLinks(suggestions)
	if len(suggestions) == 0 {
		return
	}
	suffix := "ies"
	if len(suggestions) == 1 {
		suffix = "y"
	}
	fmt.Printf("\n%s %d relationship opportunit%s detected — review: %s\n",
		ui.RenderAccent("🔗"), len(suggestions), suffix,
		ui.RenderAccent("bd devlog links suggest"))
}

var devlogLinkCmd = &cobra.Command{
	Use:   "link [from] [to]",
	Short: "Create an explicit relationship edge between two entities",
	Long: `Record an explicit dependency edge (from -> to) between two existing
entities. The edge is marked as manual, so it survives re-extraction and is
exported to .beads/links.jsonl (shared like aliases).

Example:
  bd devlog link tsstackwysiwygeditor SlashCommandMenu --relationship uses`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		rel, _ := cmd.Flags().GetString("relationship")
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer store.Close()
		if err := queries.AddManualLink(rootCtx, store.UnderlyingDB(), args[0], args[1], rel); err != nil {
			return err
		}
		if rel == "" {
			rel = "relates-to"
		}
		fmt.Printf("✅ Linked: %s --%s--> %s\n", args[0], rel, args[1])
		flushMetadata(rootCtx)
		return nil
	},
}

var devlogLinksCmd = &cobra.Command{
	Use:   "links",
	Short: "Review and manage relationship (edge) suggestions",
	Long: `Review pending relationship suggestions — entity pairs that co-occur
strongly in devlogs but have no explicit edge. Promote a pair with
'bd devlog link <from> <to> --relationship uses', or reject it with
'bd devlog links dismiss <a> <b>' so it never resurfaces.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return devlogLinksSuggestCmd.RunE(cmd, args)
	},
}

var devlogLinksSuggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "List pending relationship suggestions, strongest first",
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer store.Close()
		db := store.UnderlyingDB()

		// Fetch all (limit applied after noise filtering so we don't return a
		// short list padded with junk).
		suggestions, err := queries.GetLinkSuggestions(rootCtx, db, linkMinCooccur, 0)
		if err != nil {
			return fmt.Errorf("fetching suggestions: %w", err)
		}
		suggestions = filterNoiseLinks(suggestions)
		if limit > 0 && len(suggestions) > limit {
			suggestions = suggestions[:limit]
		}
		if jsonOutput {
			outputJSON(suggestions)
			return nil
		}
		if len(suggestions) == 0 {
			fmt.Println("✅ No pending relationship suggestions.")
			return nil
		}
		fmt.Printf("Relationship suggestions (%d shown, strongest co-occurrence first):\n\n", len(suggestions))
		for _, s := range suggestions {
			fmt.Printf("  %s ↔ %s  (%d shared sessions, overlap %.0f%%)\n", s.EntityA, s.EntityB, s.CoSessions, s.Overlap*100)
			fmt.Printf("    link:    bd devlog link %s %s --relationship uses\n", s.EntityA, s.EntityB)
			fmt.Printf("    reject:  bd devlog links dismiss %s %s\n\n", s.EntityA, s.EntityB)
		}
		fmt.Println("💡 Use --limit N for more, --json for machine-readable output.")
		return nil
	},
}

var devlogLinksDismissCmd = &cobra.Command{
	Use:   "dismiss [entityA] [entityB]",
	Short: "Reject a relationship suggestion so this pair is never suggested again",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer store.Close()
		if err := queries.DismissLinkPair(rootCtx, store.UnderlyingDB(), args[0], args[1], actor); err != nil {
			return fmt.Errorf("recording dismissal: %w", err)
		}
		fmt.Printf("✅ Dismissed: '%s' ↔ '%s' will not be suggested again.\n", args[0], args[1])
		return nil
	},
}

var devlogPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove ghost sessions that no longer exist on disk",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()
		db := store.UnderlyingDB()

		// --noise: purge junk entities (truncation artifacts, phrase fragments)
		// that predate extraction-time filtering.
		if noiseFlag, _ := cmd.Flags().GetBool("noise"); noiseFlag {
			removed := pruneNoiseEntities(rootCtx, db)
			fmt.Printf("✅ Pruned %d noise entit(ies).\n", removed)
		}

		// Remove the ghosts' rows from _index.md first — otherwise the next
		// sync re-ingests them from the index and resurrects the ghosts.
		devlogDir, _ := store.GetConfig(rootCtx, "devlog_dir")
		if devlogDir == "" {
			devlogDir = "_rules/_devlog"
		}
		var ghostFiles []string
		if ghostRows, qErr := db.QueryContext(rootCtx, "SELECT filename FROM sessions WHERE is_ghost = 1"); qErr == nil {
			for ghostRows.Next() {
				var fn string
				if ghostRows.Scan(&fn) != nil {
					continue
				}
				base := filepath.Base(fn)
				// Skip if the file reappeared on disk — sync will un-ghost it
				if _, err := os.Stat(filepath.Join(devlogDir, base)); err == nil {
					continue
				}
				ghostFiles = append(ghostFiles, base)
			}
			ghostRows.Close()
		}
		removedRows := 0
		indexPath := filepath.Join(devlogDir, "_index.md")
		if content, rErr := os.ReadFile(indexPath); rErr == nil && len(ghostFiles) > 0 {
			var kept []string
			for _, line := range strings.Split(string(content), "\n") {
				dead := false
				for _, gf := range ghostFiles {
					if strings.Contains(line, "]("+gf) {
						dead = true
						break
					}
				}
				if dead {
					removedRows++
				} else {
					kept = append(kept, line)
				}
			}
			if removedRows > 0 {
				if wErr := os.WriteFile(indexPath, []byte(strings.Join(kept, "\n")), 0644); wErr != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not update %s: %v\n", indexPath, wErr)
					removedRows = 0
				}
			}
		}

		// Manual cleanup to ensure it works on databases without cascades
		_, _ = db.Exec("DELETE FROM session_entities WHERE session_id IN (SELECT id FROM sessions WHERE is_ghost = 1)")
		_, _ = db.Exec("DELETE FROM extraction_log WHERE session_id IN (SELECT id FROM sessions WHERE is_ghost = 1)")
		_, _ = db.Exec("DELETE FROM entity_deps WHERE discovered_in IN (SELECT id FROM sessions WHERE is_ghost = 1)")

		res, err := db.Exec("DELETE FROM sessions WHERE is_ghost = 1")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error pruning ghosts: %v\n", err)
			os.Exit(1)
		}

		count, _ := res.RowsAffected()
		fmt.Printf("✅ Pruned %d ghost session(s).\n", count)
		if removedRows > 0 {
			fmt.Printf("✅ Removed %d dead row(s) from _index.md\n", removedRows)
		}
		
		// Flush cleanup
		flushMetadata(rootCtx)
	},
}

var devlogPauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause work on a specific scope (branch, entity, file, task)",
	Run: func(cmd *cobra.Command, args []string) {
		handleStateChange(cmd, types.StatePaused)
	},
}

var devlogAbandonCmd = &cobra.Command{
	Use:   "abandon",
	Short: "Abandon work on a specific scope (branch, entity, file, task)",
	Run: func(cmd *cobra.Command, args []string) {
		handleStateChange(cmd, types.StateAbandoned)
	},
}

var devlogOngoingCmd = &cobra.Command{
	Use:   "ongoing",
	Short: "Mark a scope as ongoing (alive, will be resumed) — the explicit un-pause",
	Long: `Mark a scope as ongoing: work that is alive and will be resumed, just not
actively in front of anyone right now (e.g. after a context switch).
Unlike 'pause', ongoing carries no warning for other agents. Use it to
revive a previously paused scope without checking out the branch.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleStateChange(cmd, types.StateOngoing)
	},
}

func handleStateChange(cmd *cobra.Command, state types.LifecycleState) {
	scope, _ := cmd.Flags().GetString("scope")
	message, _ := cmd.Flags().GetString("message")

	if scope == "" || message == "" {
		fmt.Println("Error: --scope and --message are mandatory.")
		os.Exit(1)
	}

	scopeType, scopeRef, err := parseScope(scope)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	store, err := sqlite.New(rootCtx, dbPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	// 1. Create special devlog entry
	devlogDir, _ := store.GetConfig(rootCtx, "devlog_dir")
	if devlogDir == "" {
		devlogDir = "_rules/_devlog"
	}

	date := time.Now().Format("2006-01-02 15:04")
	subject := fmt.Sprintf("[%s] %s: %s", state, scopeType, scopeRef)
	sessionID := fmt.Sprintf("sess-state-%s-%x", state, hashID(subject+date+message))

	fileName := fmt.Sprintf("%s_%s_%s_%s.md", time.Now().Format("2006-01-02"), state, scopeType, strings.ReplaceAll(scopeRef, "/", "-"))
	filePath := filepath.Join(devlogDir, fileName)

	// Get Author
	author := config.GetString("devlog.author")
	if author == "" {
		author = os.Getenv("BD_ACTOR")
	}
	if author == "" {
		out, _ := exec.Command("git", "config", "user.name").Output()
		author = strings.TrimSpace(string(out))
	}

	// Get Agent
	agent := os.Getenv("BD_AGENT_NAME")
	if agent == "" {
		agent = "Unknown"
	}

	// Get Branch
	branchOut, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	branchName := strings.TrimSpace(string(branchOut))

	// Get Commit SHA
	commitOut, _ := exec.Command("git", "rev-parse", "HEAD").Output()
	commitSHA := strings.TrimSpace(string(commitOut))

	// Write devlog file
	content := fmt.Sprintf(`# State Change: %s

**Date:** %s
**Actor:** %s
**Scope:** %s:%s
**Reason:** %s

### **Details**
This devlog entry records an explicit state change for the specified scope.

---

### **Context**
- **Commit:** %s
- **Branch:** %s
`, strings.ToUpper(string(state)), date, author, scopeType, scopeRef, message, commitSHA, branchName)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Error writing devlog file: %v\n", err)
		os.Exit(1)
	}

	// 2. Add to _index.md
	indexPath := filepath.Join(devlogDir, "_index.md")
	devlogLink := fmt.Sprintf("[%s](%s?id=%s", fileName, fileName, sessionID)
	if commitSHA != "" {
		devlogLink += "&sha=" + commitSHA
	}
	devlogLink += ")"
	row := fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |\n", subject, message, author, agent, date, branchName, devlogLink)

	f, err := os.OpenFile(indexPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Error opening index: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err := f.WriteString(row); err != nil {
		fmt.Printf("Error writing to index: %v\n", err)
		os.Exit(1)
	}

	// 3. Update branch_states table
	branchState := types.BranchState{
		State:         state,
		ScopeType:     scopeType,
		ScopeRef:      scopeRef,
		ShortReason:   message,
		FullReasonRef: sessionID,
		Actor:         author,
		CommitSHA:     commitSHA,
		BranchRef:     branchName,
	}

	if err := store.SetBranchState(rootCtx, branchState); err != nil {
		fmt.Printf("Error updating database: %v\n", err)
		// We don't exit here because the file was already written
	}

	fmt.Printf("✓ %s work on %s:%s\n", strings.Title(string(state)), scopeType, scopeRef)
	fmt.Printf("  Reason: %s\n", message)
	fmt.Printf("  Recorded as: %s\n", sessionID)

	// 4. Sync
	fmt.Println("→ Syncing devlog...")
	devlogSyncCmd.Run(cmd, []string{})
}

func parseScope(scope string) (types.ScopeType, string, error) {
	parts := strings.SplitN(scope, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid scope format. Use <type>:<ref> (e.g. branch:feature-x)")
	}

	stype := types.ScopeType(parts[0])
	switch stype {
	case types.ScopeBranch, types.ScopeEntity, types.ScopeFile, types.ScopeTask, types.ScopeSession:
		return stype, parts[1], nil
	default:
		return "", "", fmt.Errorf("invalid scope type: %s. Valid types: branch, entity, file, task, session", parts[0])
	}
}

func migrateIndexInternal(indexPath string, quiet bool) int {
	content, err := os.ReadFile(indexPath)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inTable := false
	migratedCount := 0

	authorOut, _ := exec.Command("git", "config", "user.name").Output()
	currentAuthor := strings.TrimSpace(string(authorOut))
	if currentAuthor == "" {
		currentAuthor = "Unknown"
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "| Subject | Problems |") {
			if !strings.Contains(trimmed, "| Branch |") {
				newLines = append(newLines, "| Subject | Problems | Author | Agent | Date | Branch | Devlog |")
				inTable = true
				continue
			}
		}

		if inTable && strings.HasPrefix(trimmed, "|---") {
			if strings.Count(trimmed, "|") >= 5 && strings.Count(trimmed, "|") <= 7 {
				newLines = append(newLines, "|---------|----------|--------|-------|------|--------|---------|")
				continue
			}
		}

		if inTable && strings.HasPrefix(trimmed, "|") {
			pipes := strings.Count(trimmed, "|")
			if pipes == 7 {
				// Migrate 6 -> 7 (Missing Branch)
				parts := strings.Split(trimmed, "|")
				subj := strings.TrimSpace(parts[1])
				prob := strings.TrimSpace(parts[2])
				auth := strings.TrimSpace(parts[3])
				agent := strings.TrimSpace(parts[4])
				date := strings.TrimSpace(parts[5])
				file := strings.TrimSpace(parts[6])

				newLines = append(newLines, fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", subj, prob, auth, agent, date, "N/A", file))
				migratedCount++
				continue
			} else if pipes == 6 {
				// Migrate 5 -> 7 (Missing Agent and Branch)
				parts := strings.Split(trimmed, "|")
				subj := strings.TrimSpace(parts[1])
				prob := strings.TrimSpace(parts[2])
				auth := strings.TrimSpace(parts[3])
				date := strings.TrimSpace(parts[4])
				file := strings.TrimSpace(parts[5])

				newLines = append(newLines, fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", subj, prob, auth, "Unknown", date, "N/A", file))
				migratedCount++
				continue
			} else if pipes == 5 {
				// Migrate 4 -> 7
				parts := strings.Split(trimmed, "|")
				subj := strings.TrimSpace(parts[1])
				prob := strings.TrimSpace(parts[2])
				date := strings.TrimSpace(parts[3])
				file := strings.TrimSpace(parts[4])

				// Add 00:00 to date if only YYYY-MM-DD
				if len(date) == 10 {
					date += " 00:00"
				}

				newLines = append(newLines, fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", subj, prob, currentAuthor, "Unknown", date, "N/A", file))
				migratedCount++
				continue
			}
		}
		newLines = append(newLines, line)
	}

	if migratedCount > 0 {
		err = os.WriteFile(indexPath, []byte(strings.Join(newLines, "\n")), 0644)
		if err == nil && !quiet {
			fmt.Printf("  %s Migrated %d rows in %s to 7-column format\n", ui.RenderPass("✓"), migratedCount, indexPath)
		}
	}
	return migratedCount
}

func normalizePathsInternal(store *sqlite.SQLiteStorage, indexPath string, quiet bool) int {
	content, err := os.ReadFile(indexPath)
	if err != nil {
		return 0
	}

	devlogDir := filepath.Dir(indexPath)
	prefix := devlogDir + string(filepath.Separator)

	lines := strings.Split(string(content), "\n")
	var newLines []string
	normalizedCount := 0
	inTable := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "| Subject | Problems |") {
			inTable = true
			newLines = append(newLines, line)
			continue
		}

		if inTable && strings.HasPrefix(trimmed, "|") && !strings.HasPrefix(trimmed, "|---") {
			pipeCount := strings.Count(trimmed, "|")
			if pipeCount >= 5 {
				// We have a row. Find the last column (the Devlog link)
				parts := strings.Split(line, "|")
				lastIdx := len(parts) - 2 // The content is between the last two pipes
				fileCol := parts[lastIdx]

				if strings.Contains(fileCol, "](") {
					start := strings.Index(fileCol, "](") + 2
					end := strings.Index(fileCol[start:], ")")
					if end != -1 {
						rawPath := fileCol[start : start+end]
						if strings.HasPrefix(rawPath, prefix) {
							// Found a polluted path!
							cleanPath := rawPath[len(prefix):]
							
							// 1. Update Index Row
							newFileCol := strings.Replace(fileCol, rawPath, cleanPath, 1)
							parts[lastIdx] = newFileCol
							line = strings.Join(parts, "|")
							
							// 2. Update Database (immediately to prevent sync collision)
							db := store.UnderlyingDB()
							_, dbErr := db.Exec("UPDATE sessions SET filename = ? WHERE filename = ?", cleanPath, rawPath)
							if dbErr == nil {
								normalizedCount++
							} else if !quiet {
								fmt.Fprintf(os.Stderr, "  %s Warning: failed to update DB for %s: %v\n", ui.RenderWarn("⚠"), rawPath, dbErr)
							}
						}
					}
				}
			}
		}
		newLines = append(newLines, line)
	}

	if normalizedCount > 0 {
		err = os.WriteFile(indexPath, []byte(strings.Join(newLines, "\n")), 0644)
		if err == nil && !quiet {
			fmt.Printf("  %s Normalized %d polluted paths in %s\n", ui.RenderPass("✓"), normalizedCount, indexPath)
		}
	}
	return normalizedCount
}

func backfillIDsInternal(store *sqlite.SQLiteStorage, indexPath string, quiet bool) int {
	content, err := os.ReadFile(indexPath)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	backfilledCount := 0
	inTable := false

	db := store.UnderlyingDB()

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "| Subject | Problems |") {
			inTable = true
			newLines = append(newLines, line)
			continue
		}

		if inTable && strings.HasPrefix(trimmed, "|") && !strings.HasPrefix(trimmed, "|---") {
			parts := strings.Split(trimmed, "|")
			if len(parts) >= 5 {
				subj := strings.TrimSpace(parts[1])
				// Devlog link is always the last column
				lastIdx := len(parts) - 2
				fileCol := strings.TrimSpace(parts[lastIdx])

				if strings.Contains(fileCol, "](") && !strings.Contains(fileCol, "?id=") {
					start := strings.Index(fileCol, "](") + 2
					end := strings.Index(fileCol[start:], ")")
					if end != -1 {
						rawPath := fileCol[start : start+end]
						cleanPath := rawPath
						// Lookup ID in DB
						var sessionID string
						err := db.QueryRow("SELECT id FROM sessions WHERE filename = ? AND title = ?", cleanPath, subj).Scan(&sessionID)
						if err == nil {
							// Found! Append ID to link
							newPath := fmt.Sprintf("%s?id=%s", cleanPath, sessionID)
							newFileCol := strings.Replace(fileCol, rawPath, newPath, 1)
							parts[lastIdx] = " " + newFileCol + " "
							line = strings.Join(parts, "|")
							backfilledCount++
						}
					}
				}
			}
		}
		newLines = append(newLines, line)
	}

	if backfilledCount > 0 {
		err = os.WriteFile(indexPath, []byte(strings.Join(newLines, "\n")), 0644)
		if err == nil && !quiet {
			fmt.Printf("  %s Backfilled %d explicit IDs into %s\n", ui.RenderPass("✓"), backfilledCount, indexPath)
		}
	}
	return backfilledCount
}

var devlogMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate devlog index to new format",
	Run: func(cmd *cobra.Command, args []string) {
		formatIndex, _ := cmd.Flags().GetBool("format-index")
		if !formatIndex {
			fmt.Println("Use --format-index to upgrade _index.md to 6-column format.")
			return
		}

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		devlogDir, _ := store.GetConfig(rootCtx, "devlog_dir")
		if devlogDir == "" {
			devlogDir = "_rules/_devlog"
		}

		indexPath := filepath.Join(devlogDir, "_index.md")
		migrateIndexInternal(indexPath, false)
	},
}

var devlogAuthorsCmd = &cobra.Command{
	Use:   "authors",
	Short: "List devlog authors and metrics",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		db := store.UnderlyingDB()
		rows, err := db.Query(`
			SELECT 
				COALESCE(author, 'Unknown') as author_name, 
				COALESCE(agent, 'Unknown') as agent_name,
				MIN(timestamp) as first_seen, 
				MAX(timestamp) as last_seen, 
				COUNT(*) as session_count 
			FROM sessions 
			GROUP BY author_name, agent_name 
			ORDER BY session_count DESC
		`)
		if err != nil {
			fmt.Printf("Error querying authors: %v\n", err)
			os.Exit(1)
		}
		defer rows.Close()

		fmt.Println("\n👥 Devlog Contributors")
		fmt.Println("--------------------------------------------------------------------------------------------------")
		fmt.Printf("%-20s %-20s %-20s %-20s %-10s\n", "Author", "Agent", "First Session", "Last Session", "Count")
		fmt.Println("--------------------------------------------------------------------------------------------------")

		for rows.Next() {
			var authorName, agentName, first, last string
			var count int
			if err := rows.Scan(&authorName, &agentName, &first, &last, &count); err != nil {
				continue
			}
			// Truncate timestamp for display if it has too much precision
			if len(first) > 16 {
				first = first[:16]
			}
			if len(last) > 16 {
				last = last[:16]
			}
			fmt.Printf("%-20s %-20s %-20s %-20s %-10d\n", authorName, agentName, first, last, count)
		}
		fmt.Println()
	},
}

func init() {
	devlogResumeCmd.Flags().IntP("last", "l", 0, "Resume last N sessions")
	devlogResumeCmd.Flags().String("author", "", "Filter resume context by author")

	devlogGraphCmd.Flags().Int("depth", 3, "Depth of graph traversal")
	devlogGraphCmd.Flags().Bool("strict", false, "Disable fuzzy matching")
	devlogGraphCmd.Flags().Int("limit", 25, "Max matching entities to show")
	devlogGraphCmd.Flags().String("type", "", "Filter by relationship type (e.g., uses, contains)")
	devlogGraphCmd.Flags().String("html", "", "Export an interactive force-graph HTML file to the given path (e.g. --html output/graph.html)")
	devlogGraphCmd.Flags().Bool("open", false, "After --html export, open the file in the default browser")
	devlogPruneCmd.Flags().Bool("noise", false, "Also purge junk entities (truncation artifacts, phrase fragments)")

	devlogImpactCmd.Flags().Bool("strict", false, "Disable fuzzy matching")
	devlogImpactCmd.Flags().Int("limit", 25, "Max matching entities to show")
	
	devlogSearchCmd.Flags().Bool("strict", false, "Disable fuzzy matching and entity expansion")
	devlogSearchCmd.Flags().Bool("text-only", false, "Disable entity expansion (BM25 only)")
	devlogSearchCmd.Flags().Int("limit", 25, "Max results to return")
	devlogSearchCmd.Flags().String("author", "", "Filter search results by author")
	devlogSearchCmd.Flags().Bool("preview", false, "Show first 3 lines of narrative in results")
	devlogSearchCmd.Flags().Bool("explain", false, "Show score breakdown for matches")

	devlogListCmd.Flags().String("type", "", "Filter by session type")
	devlogListCmd.Flags().String("author", "", "Filter by author")
	devlogListCmd.Flags().Bool("preview", false, "Show snippets of the session narrative")

	entitiesCmd.Flags().String("sort", "relationships", "Sort by 'relationships' or 'mentions'")
	entitiesCmd.Flags().Int("limit", 20, "Max entities to show")
	entitiesCmd.Flags().Int("min-rels", 1, "Minimum relationships for an entity to be shown")

	devlogVerifyCmd.Flags().Bool("fix", false, "Adopt orphans and backfill missing metadata (Fast Regex only)")
	devlogVerifyCmd.Flags().Bool("fix-regex", false, "Force regex-only extraction (faster, skips AI)")
	devlogVerifyCmd.Flags().Bool("fix-ai", false, "Force AI extraction for backfilling (slow, higher quality)")
	devlogVerifyCmd.Flags().Bool("id-stability", false, "Append explicit session IDs to devlog links in _index.md")

	devlogEnrichCmd.Flags().Bool("all", false, "Schedule all sessions for enrichment")

	devlogRecordCmd.Flags().String("subject", "", "Subject of the devlog")
	devlogRecordCmd.Flags().String("problem", "", "Problem description")
	devlogRecordCmd.Flags().String("file", "", "Devlog file path")
	devlogRecordCmd.Flags().String("author", "", "Author name (overrides git config)")
	devlogRecordCmd.Flags().String("agent", "", "Agent name (overrides detected name)")

	devlogPauseCmd.Flags().String("scope", "", "Scope to pause (branch:name, entity:id, file:path, task:id, session:id)")
	devlogPauseCmd.Flags().String("message", "", "Short reason for pausing")

	devlogAbandonCmd.Flags().String("scope", "", "Scope to abandon (branch:name, entity:id, file:path, task:id, session:id)")
	devlogAbandonCmd.Flags().String("message", "", "Short reason for abandoning")

	devlogOngoingCmd.Flags().String("scope", "", "Scope to mark ongoing (branch:name, entity:id, file:path, task:id, session:id)")
	devlogOngoingCmd.Flags().String("message", "", "Short note on why/when it will resume")

	devlogMigrateCmd.Flags().Bool("format-index", false, "Upgrade _index.md to 5-column format")

	devlogAliasCmd.Flags().Bool("dry-run", false, "Show what would be collapsed without making changes")

	devlogCmd.AddCommand(devlogInitCmd)
	devlogCmd.AddCommand(devlogOnboardCmd)
	devlogCmd.AddCommand(devlogSyncCmd)
	devlogCmd.AddCommand(devlogAliasCmd)
	devlogCmd.AddCommand(devlogUnaliasCmd)
	devlogAliasesSuggestCmd.Flags().Int("limit", 10, "Maximum suggestions to show (0 = all)")
	devlogAliasesCmd.Flags().Int("limit", 10, "Maximum suggestions to show (0 = all)")
	devlogAliasesCmd.AddCommand(devlogAliasesSuggestCmd)
	devlogAliasesCmd.AddCommand(devlogAliasesDismissCmd)
	devlogCmd.AddCommand(devlogAliasesCmd)
	// Relationship (edge) suggestions + manual linking.
	devlogLinkCmd.Flags().StringP("relationship", "r", "", "Relationship label (default: relates-to)")
	devlogLinksSuggestCmd.Flags().Int("limit", 10, "Maximum suggestions to show (0 = all)")
	devlogLinksCmd.Flags().Int("limit", 10, "Maximum suggestions to show (0 = all)")
	devlogLinksCmd.AddCommand(devlogLinksSuggestCmd)
	devlogLinksCmd.AddCommand(devlogLinksDismissCmd)
	devlogCmd.AddCommand(devlogLinkCmd)
	devlogCmd.AddCommand(devlogLinksCmd)
	devlogCmd.AddCommand(devlogPruneCmd)
	devlogCmd.AddCommand(devlogStatusCmd)
	devlogCmd.AddCommand(devlogGraphCmd)
	devlogCmd.AddCommand(devlogPathCmd)
	devlogCmd.AddCommand(devlogListCmd)
	devlogCmd.AddCommand(entitiesCmd)
	devlogCmd.AddCommand(devlogShowCmd)
	devlogCmd.AddCommand(devlogSearchCmd)
	devlogCmd.AddCommand(devlogImpactCmd)
	devlogCmd.AddCommand(devlogResumeCmd)
	devlogCmd.AddCommand(installHooksCmd)
	devlogCmd.AddCommand(devlogResetCmd)
	devlogCmd.AddCommand(devlogVerifyCmd)
	devlogCmd.AddCommand(devlogExtractCmd)
	devlogCmd.AddCommand(devlogEnrichCmd)
	devlogCmd.AddCommand(devlogRecordCmd)
	devlogCmd.AddCommand(devlogPauseCmd)
	devlogCmd.AddCommand(devlogAbandonCmd)
	devlogCmd.AddCommand(devlogOngoingCmd)
	devlogCmd.AddCommand(devlogMigrateCmd)
	devlogCmd.AddCommand(devlogAuthorsCmd)
	devlogExportCmd.Flags().StringP("output", "o", "", "Write JSON to a file instead of stdout")
	devlogCmd.AddCommand(devlogExportCmd)
	devlogCmd.AddCommand(devlogWatchCmd)

	rootCmd.AddCommand(devlogCmd)
}



func convertSearchResultsToUI(results []queries.SearchResult) []ui.SearchResultItem {
	items := make([]ui.SearchResultItem, len(results))
	for i, r := range results {
		items[i] = ui.SearchResultItem{
			ID:              r.ID,
			Filename:        r.Filename,
			Title:           r.Title,
			Date:            r.Date,
			Narrative:       r.Narrative,
			Reason:          r.Reason,
			Score:           r.Score,
			BM25:            r.BM25,
			PhraseBonus:     r.PhraseBonus,
			NearBonus:       r.NearBonus,
			EntityBonus:     r.EntityBonus,
			RecencyBonus:    r.RecencyBonus,
			IsLowConfidence: r.IsLowConfidence,
			LifecycleStatus: string(r.LifecycleStatus),
			StatusReason:    r.StatusReason,
			IsValidated:     r.IsValidated,
			Author:          r.Author,
			AuthorEmail:     r.AuthorEmail,
			Agent:           r.Agent,
			Branch:          r.Branch,
		}
	}
	return items
}

// showAliasHints prints a one-line count of pending alias suggestions.
// Deliberately NOT gated on ui.IsTerminal(): agents reading piped output are
// the primary audience for graph hygiene, and the old per-pair dump both
// flooded interactive shells (thousands of lines) and was invisible in pipes.
func showAliasHints(ctx context.Context, db *sql.DB) {
	suggestions, err := queries.GetAliasSuggestions(ctx, db, 0.8, 0)
	if err != nil || len(suggestions) == 0 {
		return
	}

	plural := "y"
	if len(suggestions) > 1 {
		plural = "ies"
	}
	fmt.Printf("\n%s %d alias opportunit%s detected — review: %s\n",
		ui.RenderAccent("💡"), len(suggestions), plural, ui.RenderAccent("bd devlog aliases suggest"))
}

func flushMetadata(ctx context.Context) {
	if config.GetBool("storage.no-auto-flush") {
		return
	}
	
	store, err := sqlite.New(ctx, dbPath)
	if err != nil {
		return
	}
	defer store.Close()

	bDir := beads.FindBeadsDir()
	if bDir == "" {
		bDir = ".beads"
	}

	// 1. Export Issues (Issues + Status)
	jsonlPath := config.GetString("storage.jsonl_path")
	if jsonlPath == "" {
		jsonlPath = filepath.Join(bDir, "issues.jsonl")
	}
	if err := exportToJSONL(ctx, jsonlPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: auto-flush issues failed: %v\n", err)
	}

	// 2. Export Aliases
	aliasesPath := filepath.Join(bDir, "aliases.jsonl")
	if err := exportAliasesToJSONL(ctx, aliasesPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: auto-flush aliases failed: %v\n", err)
	}
	if err := exportLinksToJSONL(ctx, filepath.Join(bDir, "links.jsonl")); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: auto-flush links failed: %v\n", err)
	}
}
