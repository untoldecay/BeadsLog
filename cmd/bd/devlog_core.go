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
	Subject  string
	Problem  string
	Date     string
	Filename string
	Dir      string // Directory containing the index file
}

// SyncSession synchronizes a session from an index row into the database
// It handles creation, updates, and content hash verification
func SyncSession(store *sqlite.SQLiteStorage, row IndexRow) (bool, error) {
	sessionID := fmt.Sprintf("sess-%s", hashID(row.Subject+row.Date))
	db := store.UnderlyingDB()

	// Check if session exists and get current state
	var currentFilename, currentHash string
	var currentMissing bool
	err := db.QueryRow("SELECT filename, file_hash, is_missing FROM sessions WHERE id = ?", sessionID).Scan(&currentFilename, &currentHash, &currentMissing)
	
	exists := err == nil
	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("failed to query session: %w", err)
	}

	// Read file content
	// If Filename is relative, resolve it relative to the index directory
	filePath := row.Filename
	if !filepath.IsAbs(filePath) && row.Dir != "" {
		filePath = filepath.Join(row.Dir, row.Filename)
	}

	content, err := ioutil.ReadFile(filePath)
	var contentHash string
	var narrative string
	isMissing := false
	
	if err != nil {
		isMissing = true
		// If file doesn't exist, we can still create the session record but warn
		fmt.Fprintf(os.Stderr, "Missing log session, %s : %v\n", filePath, err)
		narrative = row.Problem // Use problem description as fallback narrative
		// Hash only the problem
		sum := sha256.Sum256([]byte(narrative))
		contentHash = fmt.Sprintf("%x", sum)
	} else {
		narrative = row.Problem + "\n\n" + string(content) // Prepend problem description
		// Hash combined content
		sum := sha256.Sum256([]byte(narrative))
		contentHash = fmt.Sprintf("%x", sum)
	}

	// Determine if update is needed
	needsUpdate := !exists || currentFilename != row.Filename || currentHash != contentHash || currentMissing != isMissing

	if !needsUpdate {
		return false, nil // No changes
	}

	// Perform update/insert
	if !exists {
		_, err = db.Exec(`
			INSERT INTO sessions (id, title, timestamp, status, type, filename, narrative, file_hash, is_missing, enrichment_status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1)
		`, sessionID, row.Subject, parseDate(row.Date), "closed", extractType(row.Subject), row.Filename, narrative, contentHash, isMissing)
	} else {
		_, err = db.Exec(`
			UPDATE sessions 
			SET title = ?, timestamp = ?, type = ?, filename = ?, narrative = ?, file_hash = ?, is_missing = ?, enrichment_status = MAX(enrichment_status, 1)
			WHERE id = ?
		`, row.Subject, parseDate(row.Date), extractType(row.Subject), row.Filename, narrative, contentHash, isMissing, sessionID)
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
			// A valid row has exactly 5 pipes (starts with |, ends with |, 3 internal)
			pipeCount := strings.Count(line, "|")
			if pipeCount > 5 {
				return nil, fmt.Errorf("line %d: malformed row (too many pipes, likely multiple sessions merged into one line)", i+1)
			}
			if pipeCount < 5 {
				return nil, fmt.Errorf("line %d: malformed row (missing columns, expected 4 columns)", i+1)
			}

			parts := strings.Split(line, "|")
			if len(parts) >= 5 {
				filenamePart := strings.TrimSpace(parts[4])
				// Extract filename from markdown link [name](file)
				if strings.Contains(filenamePart, "](") {
					start := strings.Index(filenamePart, "](") + 2
					end := strings.Index(filenamePart[start:], ")")
					if end != -1 {
						filenamePart = filenamePart[start : start+end]
					}
				}

				rows = append(rows, IndexRow{
					Subject:  strings.TrimSpace(parts[1]),
					Problem:  strings.TrimSpace(parts[2]),
					Date:     strings.TrimSpace(parts[3]),
					Filename: filenamePart,
					Dir:      dir,
				})
			}
		}
	}
	return rows, nil
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
			INSERT INTO entities (id, name, type, mention_count, confidence, source)
			VALUES (?, ?, ?, 1, ?, ?)
			ON CONFLICT(name) DO UPDATE SET 
				mention_count = mention_count + 1,
				source = CASE WHEN excluded.confidence > confidence THEN excluded.source ELSE source END,
				confidence = MAX(confidence, excluded.confidence)
		`, entityID, entity.Name, entity.Type, entity.Confidence, entity.Source)
		
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error upserting entity: %v\n", err)
		}

		// Get the actual ID if it was an update (in case hash collision or existing name)
		var actualID string
		err = db.QueryRow("SELECT id FROM entities WHERE name = ?", entity.Name).Scan(&actualID)
		if err == nil {
			entityID = actualID
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

		// Link them using IDs
		_, err := db.Exec(`
			INSERT OR IGNORE INTO entity_deps (from_entity, to_entity, relationship, discovered_in)
			VALUES (?, ?, ?, ?)
		`, fromID, toID, rel.Type, sessionID)
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
        INSERT OR IGNORE INTO entities (id, name, type, mention_count)
        VALUES (?, ?, 'component', 0)
    `, entityID, name)

	// In case of name collision but different ID, get the actual one
	var actualID string
	err := db.QueryRow("SELECT id FROM entities WHERE name = ?", name).Scan(&actualID)
	if err == nil {
		return actualID
	}
	return entityID
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
	layouts := []string{"2006-01-02", "Jan 2"}
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
