package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/queries"
	"github.com/steveyegge/beads/internal/storage/sqlite"
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

		promptsDir := filepath.Join(filepath.Dir(baseDir), "_prompts")

		if err := os.MkdirAll(baseDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating devlog dir: %v\n", err)
			os.Exit(1)
		}
		if err := os.MkdirAll(promptsDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating prompts dir: %v\n", err)
			os.Exit(1)
		}

		// Create _index.md
		indexPath := filepath.Join(baseDir, "_index.md")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			if err := os.WriteFile(indexPath, []byte(indexTemplate), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing _index.md: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Created %s\n", indexPath)
		}

		// Create generate-devlog.md
		promptPath := filepath.Join(promptsDir, "generate-devlog.md")
		if _, err := os.Stat(promptPath); os.IsNotExist(err) {
			if err := os.WriteFile(promptPath, []byte(promptTemplate), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing prompt: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Created %s\n", promptPath)
		}

		// Store config
		store, err := sqlite.New(rootCtx, dbPath)
		if err == nil {
			defer store.Close()
			// Always set or update if empty
			current, _ := store.GetConfig(rootCtx, "devlog_dir")
			if current == "" || current != baseDir {
				_ = store.SetConfig(rootCtx, "devlog_dir", baseDir)
			}
		}

		fmt.Println("✅ Devlog initialized. Use 'bd devlog sync' to ingest.")

		// Agent Rules Integration
		configureAgentRules()
	},
}

func configureAgentRules() {
	rules := `
## BeadsLog Workflow Protocol
This project follows a cycle of **Planning** (Forward) and **Reflection** (Backward).
- **Planning:** Coordinate future work via a dependency-aware task graph.
- **Reflection:** Capture and retrieve past context, architectural reasoning, and session history.

### The Loop:
1. **Plan:** Before starting, check tasks: ` + "`bd ready`" + ` or ` + "`bd list`" + ` to understand the goal.
2. **Context:** Before coding, check history: ` + "`bd devlog resume --last 1`" + ` or ` + "`bd devlog search \"topic\"`" + ` to avoid repeating past mistakes.
3. **Log:** At session end, use ` + "`_rules/_devlog/_generate_devlog_prompt.md`" + ` to document assumptions and outcomes.
4. **Close:** When finished, close the task: ` + "`bd close <id>`" + `.
`
	candidates := []string{
		".windsufrules",
		".cursorrules",
		".claude/rules",
		"AGENTS.md",
		"GEMINI.md",
		".github/copilot-instructions.md",
	}

	foundFile := ""
	for _, f := range candidates {
		if _, err := os.Stat(f); err == nil {
			foundFile = f
			break
		}
	}

	if foundFile == "" {
		fmt.Println("\n💡 Tip: Add these rules to your AI agent instructions:")
		fmt.Println(rules)
		return
	}

	fmt.Printf("\nFound agent configuration: %s\n", foundFile)
	fmt.Printf("Add BeadsLog workflow rules to this file? [Y/n] ")

	var response string
	fmt.Scanln(&response)
	if response == "" || strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
		f, err := os.OpenFile(foundFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}
		defer f.Close()
		if _, err := f.WriteString("\n" + rules); err != nil {
			fmt.Printf("Error writing rules: %v\n", err)
			return
		}
		fmt.Println("✅ Rules added.")
	} else {
		fmt.Println("Skipped. Here are the rules for manual addition:")
		fmt.Println(rules)
	}
}

// devlogSyncCmd updates the database from the filesystem
var devlogSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync devlogs to database",
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
				fmt.Fprintf(os.Stderr, "Error: devlog not configured. Run 'bd devlog init'\n")
				os.Exit(1)
			}
		}

		indexPath := filepath.Join(devlogDir, "_index.md")
		// Handle the case where user manually renamed/moved files or testing scenario
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			indexPath = filepath.Join(devlogDir, "index.md")
		}

		// Check mtime against last sync
		info, err := os.Stat(indexPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading index: %v\n", err)
			os.Exit(1)
		}

		// Always parse and sync, let SyncSession handle file-level hashing optimization
		// The "last_sync" metadata check is insufficient if *content* changed without mtime change (rare but possible with git)
		// or if *other* files changed. SyncSession logic is robust enough to run cheap checks.
		
		rows := parseIndexMD(indexPath)
		if rows == nil {
			fmt.Fprintf(os.Stderr, "Error parsing index or empty\n")
			return
		}

		fmt.Printf("Scanning %d sessions...\n", len(rows))
		updatedCount := 0
		for _, row := range rows {
			updated, err := SyncSession(store, row)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error syncing session %s: %v\n", row.Subject, err)
			}
			if updated {
				updatedCount++
				fmt.Printf("  Updated: %s\n", row.Subject)
			}
		}

		// Store last sync time
		_ = store.SetMetadata(rootCtx, "last_devlog_sync", info.ModTime().Format(time.RFC3339))
		
		if updatedCount > 0 {
			if !noAutoFlush {
				flushToJSONLWithState(flushState{forceDirty: true})
			}
			fmt.Printf("✅ Synced %d sessions\n", updatedCount)
		} else {
			fmt.Println("Already up to date.")
		}
	},
}

const indexTemplate = `# Development Log Index

This index provides a concise record of all development work for easy scanning and pattern recognition across sessions.

## Nomenclature Rules:
- **[fix]** - Bug fixes and error resolution
- **[feature]** - New feature implementation
- **[enhance]** - Improvements to existing functionality
- **[rationalize]** - Code cleanup and consolidation
- **[deploy]** - Deployment activities and version releases
- **[security]** - Security fixes and vulnerability patches
- **[debug]** - Troubleshooting and investigation
- **[test]** - Testing and validation activities

## Work Index

| Subject | Problems | Date | Devlog |
|---------|----------|------|---------|
| [init] Setup | Initial devlog structure setup | 2024-01-01 | [2024-01-01_setup.md](2024-01-01_setup.md) |

---

*This index is automatically updated when devlogs are created via the generation prompt. All work subjects must be referenced in this index following the established nomenclature rules.*
`

const promptTemplate = `# Prompt: Generate Chronological Debugging & Development Log

## Objective:
Analyze the entire conversation history of the current session and generate a comprehensive, chronological development log. The primary purpose is to be a transparent record of the entire problem-solving process, detailing every assumption (especially flawed ones), every action taken, the resulting outcomes, and the evidence-based corrections that led to the final solution.

## Persona:
Act as a meticulous technical writer and project manager, documenting the development journey with a focus on learning from mistakes.

## Input:
The full conversation history of the current development session.

## File Handling Logic:
1.  **Check for Existing Log:** Before generating, list the files in the "_rules/_devlog/" directory.
2.  **Identify Today's Log:** Find the most recent file. Check if its filename matches today's date (e.g., "2025-07-04_session_summary.md" or "2025-07-04_specific-title.md").
3.  **Update or Create:**
    *   **If a log for today exists:** Read that file and append the new phases from the current session to it. Do not create a new file.
    *   **If no log for today exists:** Create a new file named "_rules/_devlog/[YYYY-MM-DD]_[concise-title-separated-by-dashes].md".
        *   **Naming Convention:** The title **MUST NOT** be generic like "session_summary". It must be descriptive of the main task (e.g., "2025-07-04_csv-import-fix.md", "2025-10-12_auth-refactor-and-docs.md").
4.  **Maintain Index:** Always update the "_rules/_devlog/_index.md" file with work subjects from the current session.
    *   **If index doesn't exist:** Create the index file with the current session's work subjects.
    *   **If index exists:** Append new work subjects to the existing table.
    *   **Nomenclature Rules:** Use prefix format "[prefix]description" for subjects (e.g., "[fix]user-authentication", "[feature]csv-import", "[deploy]v4.1.0").

## Output Structure (Embedded Template):
Generate or update a single markdown file with the following structure.

---

# Comprehensive Development Log: [Briefly Describe Main Goal of the Session]

**Date:** [Current Date: YYYY-MM-DD]

### **Objective:**
To provide a complete, transparent, and chronological log of the entire development and troubleshooting process for the features worked on during this session. This document details every assumption, every action taken, the resulting errors, and the evidence-based corrections, serving as a definitive record to prevent repeating these mistakes.

---

### **Phase [X]: [Name of the First Major Task or Problem]**

**Initial Problem:** [Describe the starting problem or goal for this phase.]

*   **My Assumption/Plan #1:** [Describe the initial plan or assumption.]
    *   **Action Taken:** [Detail the specific steps taken, e.g., "Modified file X to do Y", "Ran command Z".]
    *   **Result:** [Describe the outcome. Was it a success, failure, or partial success? Include any errors or unexpected behavior.]
    *   **Analysis/Correction:** [Explain why the initial assumption was right or wrong. If wrong, what was the evidence (e.g., error message, user feedback, file inspection) that led to the correction? What was the fix?]

*(Repeat for all assumptions and plans within the phase)*

---

### **Phase [Y]: [Name of the Second Major Task or Problem]**

[Repeat the structure from the previous phase for each major part of the session.]

---

### **Final Session Summary**

**Final Status:** [Briefly describe the state of the feature(s) at the end of the session.]
**Key Learnings:**
*   [A key technical takeaway, e.g., "Electron-builder's asarUnpack is required for native addons to preserve their directory structure."]
*   [Another key learning, e.g., "Backspace handling in contenteditable requires differentiating between empty and non-empty states to provide intuitive merging vs. de-escalation."]

---

## Guidelines for Generation:
1.  **Chronological Order:** The phases must follow the order in which they occurred in the conversation.
2.  **Focus on the "Why":** Don't just list actions. Explain the reasoning behind each action (the assumption) and the analysis of the result. The goal is to capture the thought process.
3.  **Be Honest About Mistakes:** The most valuable parts of the log are the "Flawed Assumptions" or incorrect plans. Document them clearly.
4.  **Use Evidence:** When a correction is made, mention the evidence that prompted it (e.g., "The user provided 'before' and 'after' HTML that showed...", "The error message net::ERR_FILE_NOT_FOUND indicated...").
5.  **First-Person Narrative:** Write from the perspective of the AI assistant who performed the work (e.g., "My flawed assumption was...", "I modified the file...").

---

## Index Maintenance Instructions

**Index Reference:** All work subjects from this session must be referenced in the "_rules/_devlog/_index.md" file following the established nomenclature rules. The index maintains a concise record of all development work for easy scanning and pattern recognition across sessions.

### Index Structure:
` + "```markdown" + `
## Work Index

| Subject | Problems | Date | Devlog |
|---------|----------|------|---------|
| [prefix] subject-description | Brief problem description | YYYY-MM-DD | [filename.md](filename.md) |
` + "```" + `

### Subject Nomenclature:
- **[fix]** - Bug fixes and error resolution
- **[feature]** - New feature implementation
- **[enhance]** - Improvements to existing functionality
- **[rationalize]** - Code cleanup and consolidation
- **[deploy]** - Deployment activities and version releases
- **[security]** - Security fixes and vulnerability patches
- **[debug]** - Troubleshooting and investigation
- **[test]** - Testing and validation activities

### Example Subjects:
- "[rationalize] Export endpoint system" - Consolidated 5 redundant export endpoints to 2 unified endpoints
- "[fix] Vector export format detection" - Added intelligent format selection for vector vs regular tables
- "[enhance] API client export support" - Updated frontend to use rationalized export endpoints
- "[deploy] Export rationalization v4.1.141" - Successfully deployed unified export system to staging

**Important:** Each distinct work subject in a session should be listed on its own line in the index, even if multiple subjects reference the same devlog file.

**Note:** Add a reference to this index maintenance in the devlog's "Final Session Summary" section to remind users that subjects must be referenced in the "_index.md" file in case the AI assistant doesn't follow this prompt directly.
`

var devlogGraphCmd = &cobra.Command{
	Use:   "graph [entity]",
	Short: "Display entity dependency graph",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		entityName := args[0]
		depth, _ := cmd.Flags().GetInt("depth")

		// Initialize store
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		graph, err := queries.GetEntityGraph(rootCtx, store.UnderlyingDB(), entityName, depth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error querying graph: %v\n", err)
			os.Exit(1)
		}

		printGraph(graph)
	},
}

func printGraph(graph *queries.EntityGraph) {
	if len(graph.Nodes) == 0 {
		fmt.Println("No entities found.")
		return
	}

	for _, node := range graph.Nodes {
		indent := strings.Repeat("  ", node.Depth)
		marker := "└──"
		if node.Depth == 0 {
			marker = ""
		} else {
			indent = strings.Repeat("  ", node.Depth-1)
		}
		
		fmt.Printf("%s%s %s (%d)\n", indent, marker, node.Name, node.Depth)
	}
}

var devlogListCmd = &cobra.Command{
	Use:   "list",
	Short: "List devlog sessions",
	Run: func(cmd *cobra.Command, args []string) {
		sessionType, _ := cmd.Flags().GetString("type")
		
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		query := "SELECT title, date(timestamp), type FROM sessions"
		var queryArgs []interface{}
		if sessionType != "" {
			query += " WHERE type = ?"
			queryArgs = append(queryArgs, sessionType)
		}
		query += " ORDER BY timestamp DESC"

		rows, err := store.UnderlyingDB().QueryContext(rootCtx, query, queryArgs...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing sessions: %v\n", err)
			os.Exit(1)
		}
		defer rows.Close()

		for rows.Next() {
			var title, date, typ string
			if err := rows.Scan(&title, &date, &typ); err != nil {
				fmt.Fprintf(os.Stderr, "Error scanning row: %v\n", err)
				continue
			}
			fmt.Printf("[%s] %s - %s\n", date, typ, title)
		}
	},
}

var entitiesCmd = &cobra.Command{
	Use:   "entities",
	Short: "List top entities by mention count",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		rows, err := store.UnderlyingDB().QueryContext(rootCtx, "SELECT name, mention_count FROM entities ORDER BY mention_count DESC LIMIT 20")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing entities: %v\n", err)
			os.Exit(1)
		}
		defer rows.Close()

		fmt.Println("Top Entities:")
		for rows.Next() {
			var name string
			var count int
			if err := rows.Scan(&name, &count); err != nil {
				fmt.Fprintf(os.Stderr, "Error scanning row: %v\n", err)
				continue
			}
			fmt.Printf("%s (%d)\n", name, count)
		}
	},
}

var devlogShowCmd = &cobra.Command{
	Use:   "show [date/filename]",
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

		var filename string
		err = store.UnderlyingDB().QueryRowContext(rootCtx, "SELECT filename FROM sessions WHERE filename LIKE ? OR timestamp LIKE ?", "%"+target+"%", target+"%").Scan(&filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Session not found: %v\n", err)
			os.Exit(1)
		}

		// Try to find the file
		// 1. As is (absolute or relative to cwd)
		content, err := os.ReadFile(filename)
		if err != nil {
			// 2. Relative to devlog_dir
			devlogDir, _ := store.GetConfig(rootCtx, "devlog_dir")
			if devlogDir != "" {
				path := filepath.Join(devlogDir, filename)
				content, err = os.ReadFile(path)
			}
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filename, err)
			os.Exit(1)
		}
		fmt.Println(string(content))
	},
}

var devlogSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search sessions and entities",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		fmt.Printf("Searching for: %s\n\n", query)
		
		fmt.Println("Sessions:")
		rows, _ := store.UnderlyingDB().QueryContext(rootCtx, "SELECT title FROM sessions WHERE title LIKE ? OR narrative LIKE ?", "%"+query+"%", "%"+query+"%")
		foundSessions := false
		for rows != nil && rows.Next() {
			foundSessions = true
			var title string
			rows.Scan(&title)
			fmt.Printf("- %s\n", title)
		}
		if rows != nil { rows.Close() }
		if !foundSessions {
			fmt.Println("  (No sessions found)")
		}

		fmt.Println("\nEntities:")
		rows, _ = store.UnderlyingDB().QueryContext(rootCtx, "SELECT name FROM entities WHERE name LIKE ?", "%"+query+"%")
		foundEntities := false
		for rows != nil && rows.Next() {
			foundEntities = true
			var name string
			rows.Scan(&name)
			fmt.Printf("- %s\n", name)
		}
		if rows != nil { rows.Close() }
		if !foundEntities {
			fmt.Println("  (No entities found)")
		}
	},
}

var devlogImpactCmd = &cobra.Command{
	Use:   "impact [entity]",
	Short: "Show what depends on an entity",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		entityName := args[0]
		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		rows, err := store.UnderlyingDB().QueryContext(rootCtx, `
			SELECT e.name, ed.relationship 
			FROM entity_deps ed 
			JOIN entities e ON ed.from_entity = e.id 
			WHERE ed.to_entity IN (SELECT id FROM entities WHERE name = ?)
		`, entityName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer rows.Close()

		fmt.Printf("Impact of %s:\n", entityName)
		found := false
		for rows.Next() {
			found = true
			var name, rel string
			rows.Scan(&name, &rel)
			fmt.Printf("- %s (%s)\n", name, rel)
		}
		if !found {
			fmt.Println("  (No known dependencies found)")
		}
	},
}

var devlogResumeCmd = &cobra.Command{
	Use:   "resume [query]",
	Short: "Resume debugging with context",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		lastN, _ := cmd.Flags().GetInt("last")
		
		if lastN > 0 || len(args) == 0 {
			if lastN == 0 { lastN = 1 } // Default to last 1 if no arg and no flag
			
			store, err := sqlite.New(rootCtx, dbPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
				os.Exit(1)
			}
			defer store.Close()

			fmt.Printf("Resuming last %d session(s):\n\n", lastN)
			rows, err := store.UnderlyingDB().QueryContext(rootCtx, "SELECT title, narrative FROM sessions ORDER BY timestamp DESC LIMIT ?", lastN)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching sessions: %v\n", err)
				os.Exit(1)
			}
			defer rows.Close()

			for rows.Next() {
				var title, narrative string
				rows.Scan(&title, &narrative)
				fmt.Printf("=== %s ===\n%s\n\n", title, narrative)
			}
			return
		}

		query := args[0]
		fmt.Printf("Resuming context for: %s\n", query)
		// Minimal implementation: search sessions and show latest
		devlogSearchCmd.Run(cmd, args)
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
if command -v bd >/dev/null 2>&1; then
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

		fmt.Println("\nDevlog System Status")
		fmt.Println("====================")
		
		if devlogDir == "" {
			fmt.Println("Status: Not configured")
			fmt.Println("Action: Run 'bd devlog init' to set up a devlog space.")
			return
		}

		fmt.Printf("Space Directory: %s\n", devlogDir)
		fmt.Printf("Last Sync:       %s\n", lastSync)

		// Get stats
		db := store.UnderlyingDB()
		var sessionsCount, entitiesCount, relationshipsCount int
		
		_ = db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessionsCount)
		_ = db.QueryRow("SELECT COUNT(*) FROM entities").Scan(&entitiesCount)
		_ = db.QueryRow("SELECT COUNT(*) FROM entity_deps").Scan(&relationshipsCount)

		fmt.Printf("\nDatabase Statistics:\n")
		fmt.Printf("  Sessions:      %d\n", sessionsCount)
		fmt.Printf("  Entities:      %d\n", entitiesCount)
		fmt.Printf("  Relationships: %d\n", relationshipsCount)

		// Check hooks
		fmt.Printf("\nGit Hooks:\n")
		hooks := []string{"post-commit", "post-merge"}
		for _, h := range hooks {
			installed := false
			path := filepath.Join(".git/hooks", h)
			if content, err := os.ReadFile(path); err == nil {
				if strings.Contains(string(content), "bd devlog sync") {
					installed = true
				}
			}
			status := "✗"
			if installed {
				status = "✓"
			}
			fmt.Printf("  %s %s\n", status, h)
		}
		
		fmt.Println()
	},
}

func init() {
	devlogResumeCmd.Flags().IntP("last", "l", 0, "Resume last N sessions")
	devlogGraphCmd.Flags().Int("depth", 3, "Depth of graph traversal")
	devlogListCmd.Flags().String("type", "", "Filter by session type")

	devlogCmd.AddCommand(devlogInitCmd)
	devlogCmd.AddCommand(devlogSyncCmd)
	devlogCmd.AddCommand(devlogStatusCmd)
	devlogCmd.AddCommand(devlogGraphCmd)
	devlogCmd.AddCommand(devlogListCmd)
	devlogCmd.AddCommand(entitiesCmd)
	devlogCmd.AddCommand(devlogShowCmd)
	devlogCmd.AddCommand(devlogSearchCmd)
	devlogCmd.AddCommand(devlogImpactCmd)
	devlogCmd.AddCommand(devlogResumeCmd)
	devlogCmd.AddCommand(installHooksCmd)
	
	rootCmd.AddCommand(devlogCmd)
}
