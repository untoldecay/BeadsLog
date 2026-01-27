package main

import (
	"context"
	"path/filepath"
	"strings"
	"io/ioutil"
	"crypto/sha256"
	"fmt"

	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/storage/sqlite"
)

// EnrichmentWorkerLogger is a subset of the daemon logger used by the enrichment worker
type EnrichmentWorkerLogger interface {
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// ProcessEnrichmentQueue finds sessions that have only been processed by regex
// and enriches them with AI extraction in the background.
func ProcessEnrichmentQueue(ctx context.Context, store *sqlite.SQLiteStorage, log EnrichmentWorkerLogger) error {
	// Only run if enrichment is enabled in config
	if !config.GetBool("entity_extraction.enabled") || !config.GetBool("entity_extraction.background_enrichment") {
		return nil
	}

	db := store.UnderlyingDB()

	// Find one session at a time to process
	var id, title, filename, narrative string
	err := db.QueryRowContext(ctx, `
		SELECT id, title, filename, narrative 
		FROM sessions 
		WHERE enrichment_status = 1 
		ORDER BY timestamp DESC 
		LIMIT 1
	`).Scan(&id, &title, &filename, &narrative)

	if err != nil {
		return nil // No pending work or error
	}

	log.Info("background enrichment starting", "session", id, "title", title)

	// Run AI extraction (ForceRegex = false)
	// Note: we use context.Background() for the actual extraction to ensure it completes 
	// even if the daemon loop cycle context is cancelled, though we should respect ctx for DB.
	result, err := extractAndLinkEntities(store, id, narrative, ExtractionOptions{ForceRegex: false})
	if err != nil {
		log.Error("enrichment failed", "session", id, "error", err)
		// Mark as failed (status 3) so we don't loop on it
		_, _ = db.ExecContext(ctx, "UPDATE sessions SET enrichment_status = 3 WHERE id = ?", id)
		return err
	}

	// Resolve absolute path for crystallization
	filePath := filename
	if !filepath.IsAbs(filePath) {
		devlogDir, _ := store.GetConfig(ctx, "devlog_dir")
		if devlogDir == "" {
			devlogDir = "_rules/_devlog"
		}
		filePath = filepath.Join(devlogDir, filename)
	}

	// Crystallize to disk
	if err := crystallizeRelationships(filePath, result.Relationships); err != nil {
		log.Warn("crystallization failed", "path", filePath, "error", err)
	} else if len(result.Relationships) > 0 {
		// If we modified the file, update the hash in DB so sync doesn't re-trigger
		newContent, err := ioutil.ReadFile(filePath)
		if err == nil {
			// We need the subject/date to re-hash if we wanted, but here we just update file_hash
			// SyncSession uses Problem + Content for narrative hash.
			// Since we don't have row.Problem here easily, we'll try to find it or just update based on new narrative.
			
			// Extract problem from existing narrative (it's the first part)
			parts := strings.SplitN(narrative, "\n\n", 2)
			problem := ""
			if len(parts) > 0 {
				problem = parts[0]
			}
			
			newNarrative := problem + "\n\n" + string(newContent)
			sum := sha256.Sum256([]byte(newNarrative))
			newContentHash := fmt.Sprintf("%x", sum)
			
			_, _ = db.ExecContext(ctx, "UPDATE sessions SET file_hash = ?, narrative = ? WHERE id = ?", newContentHash, newNarrative, id)
		}
	}

	// Mark as finished (Update status to 2)
	_, err = db.ExecContext(ctx, "UPDATE sessions SET enrichment_status = 2 WHERE id = ?", id)
	if err != nil {
		log.Error("failed to update enrichment status", "session", id, "error", err)
		return err
	}

	log.Info("background enrichment complete", "session", id, "entities", len(result.Entities), "rels", len(result.Relationships))
	return nil
}
