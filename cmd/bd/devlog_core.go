package main

import (
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
			INSERT INTO sessions (id, title, timestamp, status, type, filename, narrative, file_hash, is_missing)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, sessionID, row.Subject, parseDate(row.Date), "closed", extractType(row.Subject), row.Filename, narrative, contentHash, isMissing)
	} else {
		_, err = db.Exec(`
			UPDATE sessions 
			SET title = ?, timestamp = ?, type = ?, filename = ?, narrative = ?, file_hash = ?, is_missing = ?
			WHERE id = ?
		`, row.Subject, parseDate(row.Date), extractType(row.Subject), row.Filename, narrative, contentHash, isMissing, sessionID)
	}

	if err != nil {
		return false, fmt.Errorf("failed to upsert session: %w", err)
	}

	// Extract and link entities
	extractAndLinkEntities(store, sessionID, narrative)

	return true, nil
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
				continue
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

func extractAndLinkEntities(store *sqlite.SQLiteStorage, sessionID, text string) {
	entityPatterns := []*regexp.Regexp{
		regexp.MustCompile(`[A-Z][a-z]+(?:[A-Z][a-z]+)+`), // CamelCase
		regexp.MustCompile(`(?i)(modal|hook|endpoint|migration|service)`),
		regexp.MustCompile(`[a-z]+-[a-z]+`), // kebab-case
	}

	db := store.UnderlyingDB()
	seen := make(map[string]bool)

	for _, pat := range entityPatterns {
		matches := pat.FindAllString(text, -1)
		for _, match := range matches {
			if len(match) > 3 && !seen[match] {
				entityID := fmt.Sprintf("ent-%s", hashID(match))
				matchLower := strings.ToLower(match)

				// Create/update entity
				_, err := db.Exec(`
                    INSERT INTO entities (id, name, type, mention_count)
                    VALUES (?, ?, 'component', 1)
                    ON CONFLICT(name) DO UPDATE SET mention_count = mention_count + 1
                `, entityID, matchLower)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error upserting entity: %v\n", err)
				}

				// Get the actual ID if it was an update (in case hash collision or existing name)
				var actualID string
				err = db.QueryRow("SELECT id FROM entities WHERE name = ?", matchLower).Scan(&actualID)
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

				seen[match] = true
			}
		}
	}

	// Phase 2: Extract explicit relationships
	// Looking for pattern: "- EntityA -> EntityB (relationship)"
	relPat := regexp.MustCompile(`(?m)^\s*-\s+([a-zA-Z0-9\-_]+)\s+->\s+([a-zA-Z0-9\-_]+)(?:\s+\(([^)]+)\))?`)
	relMatches := relPat.FindAllStringSubmatch(text, -1)

	for _, match := range relMatches {
		if len(match) >= 3 {
			fromName := strings.ToLower(strings.TrimSpace(match[1]))
			toName := strings.ToLower(strings.TrimSpace(match[2]))
			relType := "depends_on"
			if len(match) > 3 && match[3] != "" {
				relType = strings.TrimSpace(match[3])
			}

			// Ensure both entities exist and get their IDs
			fromID := ensureEntityExists(store, fromName)
			toID := ensureEntityExists(store, toName)

			// Link them using IDs
			_, err := db.Exec(`
                INSERT OR IGNORE INTO entity_deps (from_entity, to_entity, relationship, discovered_in)
                VALUES (?, ?, ?, ?)
            `, fromID, toID, relType, sessionID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error linking entities: %v\n", err)
			}
		}
	}
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
	return fmt.Sprintf("%x", h.Sum32())[:6]
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
