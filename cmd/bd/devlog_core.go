package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/extractor"
	"github.com/untoldecay/BeadsLog/internal/storage/sqlite"
)

type IndexRow struct {
	ID          string // Explicit ID if present in the link (e.g. file.md?id=sess-abc)
	CommitSHA   string // Explicit SHA if present in the link (e.g. file.md?sha=abc)
	Subject     string
	Problem     string
	Author      string
	Agent       string
	Branch      string
	AuthorEmail string
	Date        string
	Filename    string
	Dir         string // Directory containing the index file
}

// SyncSession synchronizes a session from an index row into the database
// It handles creation, updates, and content hash verification
func SyncSession(store *sqlite.SQLiteStorage, row IndexRow) (bool, error) {
	db := store.UnderlyingDB()

	// 1. Session ID Resolution Strategy
	var sessionID string
	var err error
	
	if row.ID != "" {
		// Use explicit ID from index link
		sessionID = row.ID
	} else {
		// Lookup-by-Filename-and-Title Strategy (Fallback for implicit rows)
		err = db.QueryRow("SELECT id FROM sessions WHERE filename = ? AND title = ?", row.Filename, row.Subject).Scan(&sessionID)
		if err != nil {
			if err == sql.ErrNoRows {
				// New entry: Generate legacy hash-based ID
				sessionID = fmt.Sprintf("sess-%s", hashID(row.Subject+row.Date))
			} else {
				return false, fmt.Errorf("failed to lookup session: %w", err)
			}
		}
	}

	// Check if session exists and get current state
	var currentFilename, currentHash, currentAuthor, currentAuthorEmail, currentAgent, currentBranch, currentSHA string
	var currentMissing, currentGhost bool
	err = db.QueryRow("SELECT filename, file_hash, is_missing, is_ghost, COALESCE(author, ''), COALESCE(author_email, ''), COALESCE(agent, ''), COALESCE(branch, ''), COALESCE(commit_sha, '') FROM sessions WHERE id = ?", sessionID).Scan(&currentFilename, &currentHash, &currentMissing, &currentGhost, &currentAuthor, &currentAuthorEmail, &currentAgent, &currentBranch, &currentSHA)
	
	exists := err == nil
	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("failed to query session: %w", err)
	}

	// Read file content
	// If Filename is relative, resolve it relative to the index directory
	// Use base filename to prevent double-path if Filename already has prefix
	filePath := row.Filename
	if !filepath.IsAbs(filePath) && row.Dir != "" {
		filePath = filepath.Join(row.Dir, filepath.Base(row.Filename))
	}

	content, err := ioutil.ReadFile(filePath)
	var contentHash string
	var narrative string
	isMissing := false
	isGhost := currentGhost
	
	if err != nil {
		isMissing = true
		// Only warn if NOT already marked as ghost
		if !currentGhost {
			fmt.Fprintf(os.Stderr, "Missing log session, %s : %v (marking as ghost)\n", filePath, err)
			isGhost = true
		}
		narrative = row.Problem // Use problem description as fallback narrative
		// Hash only the problem
		sum := sha256.Sum256([]byte(narrative))
		contentHash = fmt.Sprintf("%x", sum)
	} else {
		isMissing = false
		isGhost = false // Recovered!
		narrative = row.Problem + "\n\n" + string(content) // Prepend problem description
		// Hash combined content
		sum := sha256.Sum256([]byte(narrative))
		contentHash = fmt.Sprintf("%x", sum)
	}

	// Determine if update is needed
	needsUpdate := !exists || 
		currentFilename != row.Filename || 
		currentHash != contentHash || 
		currentMissing != isMissing ||
		currentGhost != isGhost ||
		currentAuthor != row.Author ||
		currentAuthorEmail != row.AuthorEmail ||
		currentAgent != row.Agent ||
		currentBranch != row.Branch ||
		currentSHA != row.CommitSHA

	if !needsUpdate {
		return false, nil // No changes
	}

	// Perform update/insert
	if !exists {
		_, err = db.Exec(`
			INSERT INTO sessions (id, title, timestamp, status, type, filename, narrative, file_hash, is_missing, is_ghost, enrichment_status, author, author_email, agent, branch, commit_sha)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?)
		`, sessionID, row.Subject, parseDate(row.Date), "closed", extractType(row.Subject), row.Filename, narrative, contentHash, isMissing, isGhost, row.Author, row.AuthorEmail, row.Agent, row.Branch, row.CommitSHA)
	} else {
		_, err = db.Exec(`
			UPDATE sessions 
			SET title = ?, timestamp = ?, type = ?, filename = ?, narrative = ?, file_hash = ?, is_missing = ?, is_ghost = ?, enrichment_status = MAX(enrichment_status, 1), author = ?, author_email = ?, agent = ?, branch = ?, commit_sha = ?
			WHERE id = ?
		`, row.Subject, parseDate(row.Date), extractType(row.Subject), row.Filename, narrative, contentHash, isMissing, isGhost, row.Author, row.AuthorEmail, row.Agent, row.Branch, row.CommitSHA, sessionID)
	}

	if err != nil {
		return false, fmt.Errorf("failed to upsert session: %w", err)
	}

	// Extract and link entities (SYNC: Always ForceRegex for speed)
	result, err := extractAndLinkEntities(store, sessionID, narrative, ExtractionOptions{ForceRegex: true})
	if err == nil && result != nil && !isMissing {
		// Crystallize (Write-Back) discovered relationships to the file (from Regex)
		if err := crystallizeRelationships(filePath, result.Relationships); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to crystallize relationships to %s: %v\n", filePath, err)
		} else if len(result.Relationships) > 0 {
			// Update hash if crystallized
			newContent, err := ioutil.ReadFile(filePath)
			if err == nil {
				newNarrative := row.Problem + "\n\n" + string(newContent)
				newSum := sha256.Sum256([]byte(newNarrative))
				newContentHash := fmt.Sprintf("%x", newSum)
				_, _ = db.Exec("UPDATE sessions SET file_hash = ?, narrative = ? WHERE id = ?", newContentHash, newNarrative, sessionID)
			}
		}
	}

	return true, nil
}

func crystallizeRelationships(filename string, relationships []extractor.Relationship) error {
	if len(relationships) == 0 {
		return nil
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	contentStr := string(content)
	
	var toAppend []string
	for _, rel := range relationships {
		// Check if "From -> To" exists in the text (to avoid duplicates)
		pattern := fmt.Sprintf(`(?i)%s\s*->\s*%s`, regexp.QuoteMeta(rel.FromEntity), regexp.QuoteMeta(rel.ToEntity))
		matched, _ := regexp.MatchString(pattern, contentStr)
		if !matched {
			// Format: - From -> To (Type)
			relLine := fmt.Sprintf("- %s -> %s (%s)", rel.FromEntity, rel.ToEntity, rel.Type)
			toAppend = append(toAppend, relLine)
		}
	}
	
	if len(toAppend) > 0 {
		// Prepare content to write
		var newContent strings.Builder
		newContent.WriteString(contentStr)

		// Check if header exists
		header := "### Architectural Relationships"
		if !strings.Contains(contentStr, header) {
			// Ensure separation
			if !strings.HasSuffix(contentStr, "\n") {
				newContent.WriteString("\n")
			}
			if !strings.HasSuffix(contentStr, "\n\n") {
				newContent.WriteString("\n")
			}
			newContent.WriteString(header + "\n")
			newContent.WriteString("<!-- Format: [From Entity] -> [To Entity] (relationship type) -->\n")
		} else {
			// Header exists, make sure we end with a newline before appending
			if !strings.HasSuffix(contentStr, "\n") {
				newContent.WriteString("\n")
			}
		}
		
		for _, line := range toAppend {
			newContent.WriteString(line + "\n")
		}

		if err := os.WriteFile(filename, []byte(newContent.String()), 0644); err != nil {
			return err
		}
		fmt.Printf("  ✨ Crystallized %d new relationships to %s\n", len(toAppend), filepath.Base(filename))
	}
	return nil
}

func parseIndexMD(filename string) ([]IndexRow, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	text := string(data)
	
	// Check for common syntax corruption
	if strings.Count(text, "## Work Index") > 1 {
		return nil, fmt.Errorf("duplicate '## Work Index' headers detected (AI append error)")
	}

	dir := filepath.Dir(filename)
	lines := strings.Split(text, "\n")
	var rows []IndexRow
	inTable := false

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "| Subject | Problems |") {
			inTable = true
			continue
		}

		if inTable {
			if strings.HasPrefix(line, "|---") {
				continue
			}
			
			if !strings.HasPrefix(line, "|") {
				// We exited the table area?
				// But we are strict now: if we were in a table and see garbage, complain
				return nil, fmt.Errorf("line %d: found content after the table. The Work Index table must be the very last element in the file. Delete any footers or text below the table.", i+1)
			}

			// Critical Check: Double appends (two rows on same line)
			pipeCount := strings.Count(line, "|")
			if pipeCount > 8 {
				return nil, fmt.Errorf("line %d: malformed row (too many pipes)", i+1)
			}
			if pipeCount < 5 {
				return nil, fmt.Errorf("line %d: malformed row (missing columns)", i+1)
			}

			parts := strings.Split(line, "|")
			var row IndexRow
			var filenamePart string

			switch pipeCount {
			case 8: // 7-column format: | Subject | Problems | Author | Agent | Date | Branch | Devlog |
				filenamePart = strings.TrimSpace(parts[7])
				row = IndexRow{
					Subject: strings.TrimSpace(parts[1]),
					Problem: strings.TrimSpace(parts[2]),
					Author:  strings.TrimSpace(parts[3]),
					Agent:   strings.TrimSpace(parts[4]),
					Date:    strings.TrimSpace(parts[5]),
					Branch:  strings.TrimSpace(parts[6]),
					Dir:     dir,
				}
			case 7: // 6-column format: | Subject | Problems | Author | Agent | Date | Devlog |
				filenamePart = strings.TrimSpace(parts[6])
				row = IndexRow{
					Subject: strings.TrimSpace(parts[1]),
					Problem: strings.TrimSpace(parts[2]),
					Author:  strings.TrimSpace(parts[3]),
					Agent:   strings.TrimSpace(parts[4]),
					Date:    strings.TrimSpace(parts[5]),
					Dir:     dir,
				}
			case 6: // 5-column format: | Subject | Problems | Author | Date | Devlog |
				filenamePart = strings.TrimSpace(parts[5])
				row = IndexRow{
					Subject: strings.TrimSpace(parts[1]),
					Problem: strings.TrimSpace(parts[2]),
					Author:  strings.TrimSpace(parts[3]),
					Date:    strings.TrimSpace(parts[4]),
					Dir:     dir,
				}
			case 5: // 4-column format: | Subject | Problems | Date | Devlog |
				filenamePart = strings.TrimSpace(parts[4])
				row = IndexRow{
					Subject: strings.TrimSpace(parts[1]),
					Problem: strings.TrimSpace(parts[2]),
					Date:    strings.TrimSpace(parts[3]),
					Dir:     dir,
				}
			}

			// Extract filename from markdown link [name](file)
			if strings.Contains(filenamePart, "](") {
				start := strings.Index(filenamePart, "](") + 2
				end := strings.Index(filenamePart[start:], ")")
				if end != -1 {
					filenamePart = filenamePart[start : start+end]
				}
			}

			// Extract explicit ID and SHA if present (file.md?id=sess-abc&sha=123)
			if strings.Contains(filenamePart, "?") {
				idx := strings.Index(filenamePart, "?")
				query := filenamePart[idx+1:]
				filenamePart = filenamePart[:idx]
				
				params := strings.Split(query, "&")
				for _, p := range params {
					if strings.HasPrefix(p, "id=") {
						row.ID = strings.TrimPrefix(p, "id=")
					} else if strings.HasPrefix(p, "sha=") {
						row.CommitSHA = strings.TrimPrefix(p, "sha=")
					}
				}
			}

			// Normalize: Strip Dir prefix if present (prevent double-path bug)
			if dir != "" && strings.HasPrefix(filenamePart, dir+string(filepath.Separator)) {
				filenamePart = filenamePart[len(dir)+1:]
			}
			row.Filename = filenamePart

			rows = append(rows, row)
		}
	}
	return rows, nil
}

// extractDevlogMetadata parses a markdown file to find the # Title and ## Problem section
func extractDevlogMetadata(path string) (string, string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	lines := strings.Split(string(content), "\n")
	var subject, problem string
	var inProblem bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// 1. Get Subject (# Title)
		if subject == "" && strings.HasPrefix(trimmed, "# ") {
			subject = strings.TrimPrefix(trimmed, "# ")
			continue
		}

		// 2. Detect Problem Header (## Problem)
		if strings.HasPrefix(trimmed, "## Problem") {
			inProblem = true
			continue
		}

		// 3. Stop if another H2 starts
		if inProblem && strings.HasPrefix(trimmed, "## ") && !strings.HasPrefix(trimmed, "## Problem") {
			inProblem = false
			break
		}

		// 4. Capture first non-empty line of problem
		if inProblem && problem == "" && trimmed != "" {
			problem = trimmed
		}
	}

	return subject, problem, nil
}

// GetOrphanedFiles returns a list of .md files in the devlog directory that are NOT in the index
func GetOrphanedFiles(dir string, indexedRows []IndexRow) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	indexed := make(map[string]bool)
	for _, row := range indexedRows {
		indexed[filepath.Base(row.Filename)] = true
	}

	var orphans []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if strings.HasSuffix(name, ".md") && name != "_index.md" && name != "index.md" && name != "_generate-devlog.md" && name != "_generate-catchup.md" {
			if !indexed[name] {
				orphans = append(orphans, name)
			}
		}
	}
	return orphans, nil
}

type ExtractionOptions struct {
	ForceRegex bool
}

func extractAndLinkEntities(store *sqlite.SQLiteStorage, sessionID, text string, opts ExtractionOptions) (*extractor.ExtractionResult, error) {
	ollamaModel := ""
	if !opts.ForceRegex && config.GetBool("entity_extraction.enabled") && config.GetString("entity_extraction.primary_extractor") == "ollama" {
		ollamaModel = config.GetString("ollama.model")
	}

	pipeline := extractor.NewPipeline(ollamaModel)
	result, err := pipeline.Run(context.Background(), text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running extraction pipeline: %v\n", err)
		return nil, err
	}

	db := store.UnderlyingDB()
	
	// Store extraction log
	_, err = db.Exec(`
		INSERT INTO extraction_log (session_id, extractor, input_length, entities_found, duration_ms)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID, result.Extractor, len(text), len(result.Entities), result.Duration.Milliseconds())
	if err != nil {
		// Log but don't fail, table might not exist yet if migration failed or wasn't run
		// (though it should be run)
		// fmt.Fprintf(os.Stderr, "Warning: failed to log extraction: %v\n", err)
	}

	// 1. Process Entities
	for _, entity := range result.Entities {
		entityID := fmt.Sprintf("ent-%s", hashID(entity.Name))

		// Create/update entity
		// We use ON CONFLICT to update mention_count and potentially confidence/source if better
		// For now, we just increment mention count. 
		// Future: update confidence if new confidence > old confidence
		_, err := db.Exec(`
			INSERT INTO entities (id, name, preferred_name, type, mention_count, confidence, source)
			VALUES (?, ?, ?, ?, 1, ?, ?)
			ON CONFLICT(name) DO UPDATE SET 
				mention_count = mention_count + 1,
				source = CASE WHEN excluded.confidence > confidence THEN excluded.source ELSE source END,
				confidence = MAX(confidence, excluded.confidence),
				preferred_name = CASE WHEN excluded.confidence >= confidence THEN excluded.preferred_name ELSE preferred_name END
		`, entityID, strings.ToLower(entity.Name), entity.Name, entity.Type, entity.Confidence, entity.Source)
		
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error upserting entity: %v\n", err)
		}

		// Get the actual ID if it was an update (in case hash collision or existing name)
		var actualID string
		err = db.QueryRow("SELECT id FROM entities WHERE name = ?", strings.ToLower(entity.Name)).Scan(&actualID)
		if err == nil {
			entityID = actualID
		}
		
		// STICKY ALIAS: Check if this name is aliased to something else
		if canonicalID := resolveAlias(store, entity.Name); canonicalID != "" {
			entityID = canonicalID
		}

		// Link session -> entity
		_, err = db.Exec(`
			INSERT OR IGNORE INTO session_entities (session_id, entity_id, relevance)
			VALUES (?, ?, 'primary')
		`, sessionID, entityID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error linking session entity: %v\n", err)
		}
	}

	// 2. Process Relationships
	for _, rel := range result.Relationships {
		// Ensure both entities exist and get their IDs
		fromID := ensureEntityExists(store, rel.FromEntity)
		toID := ensureEntityExists(store, rel.ToEntity)

		// STICKY ALIAS: Respect registry
		if alias := resolveAlias(store, rel.FromEntity); alias != "" {
			fromID = alias
		}
		if alias := resolveAlias(store, rel.ToEntity); alias != "" {
			toID = alias
		}

		// Link them using IDs
		// Use ON CONFLICT to update confidence if a higher one is found
		_, err := db.Exec(`
			INSERT INTO entity_deps (from_entity, to_entity, relationship, discovered_in, confidence, source)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(from_entity, to_entity, relationship) DO UPDATE SET
				confidence = MAX(confidence, excluded.confidence),
				source = CASE WHEN excluded.confidence > confidence THEN excluded.source ELSE source END,
				discovered_in = CASE WHEN excluded.confidence > confidence THEN excluded.discovered_in ELSE discovered_in END
		`, fromID, toID, rel.Type, sessionID, rel.Confidence, rel.Source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error linking entities: %v\n", err)
		}
	}

	return result, nil
}

func ensureEntityExists(store *sqlite.SQLiteStorage, name string) string {
	db := store.UnderlyingDB()
	entityID := fmt.Sprintf("ent-%s", hashID(name))
	_, _ = db.Exec(`
        INSERT OR IGNORE INTO entities (id, name, preferred_name, type, mention_count)
        VALUES (?, ?, ?, 'component', 0)
    `, entityID, strings.ToLower(name), name)

	// In case of name collision but different ID, get the actual one
	var actualID string
	err := db.QueryRow("SELECT id FROM entities WHERE name = ?", strings.ToLower(name)).Scan(&actualID)
	if err == nil {
		return actualID
	}
	return entityID
}

func resolveAlias(store *sqlite.SQLiteStorage, name string) string {
	db := store.UnderlyingDB()
	var canonicalID string
	err := db.QueryRow("SELECT canonical_id FROM entity_aliases WHERE alias_name = ?", strings.ToLower(name)).Scan(&canonicalID)
	if err == nil {
		return canonicalID
	}
	return ""
}

func hashID(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	// Legacy behavior: use %x and take first 6 characters
	// This matches the existing database IDs.
	hex := fmt.Sprintf("%x", h.Sum32())
	if len(hex) >= 6 {
		return hex[:6]
	}
	// Fallback for very short hashes (unlikely but safe)
	return fmt.Sprintf("%06s", hex)[:6]
}

func parseDate(dateStr string) time.Time {
	layouts := []string{"2006-01-02 15:04", "2006-01-02", "Jan 2"}
	for _, layout := range layouts {
        // Handle markdown link in date column if present (e.g. [2025-01-01](...))
        if strings.Contains(dateStr, "[") && strings.Contains(dateStr, "]") {
            start := strings.Index(dateStr, "[") + 1
            end := strings.Index(dateStr, "]")
            if end > start {
                dateStr = dateStr[start:end]
            }
        }
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t
		}
	}
	return time.Now()
}

func extractType(subject string) string {
	prefixes := map[string]string{
		"fix":         "fix",
		"feature":     "feature",
		"enhance":     "enhance",
		"rationalize": "chore",
		"deploy":      "deploy",
		"security":    "security",
		"debug":       "debug",
	}
	subjectLower := strings.ToLower(subject)
	for prefix, typ := range prefixes {
		if strings.HasPrefix(subjectLower, prefix) || strings.HasPrefix(subjectLower, "["+prefix+"]") {
			return typ
		}
	}
	return "task"
}
