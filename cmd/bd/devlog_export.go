package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/storage/sqlite"
)

// devlogExportCmd dumps the full knowledge graph (entities, edges, sessions) as
// JSON for external tools. Rows are ORDER BY'd so the output is stable /
// diffable, complementing the deterministic-extraction fix (BeadsLog-5xf).
var devlogExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the full knowledge graph (entities, edges, sessions) as JSON",
	Long: `Dump the devlog knowledge graph to JSON on stdout (or a file with -o).

  bd devlog export                 # JSON to stdout
  bd devlog export -o graph.json   # to a file

Output is stable (rows are sorted), so it diffs cleanly across runs/machines.`,
	Run: func(cmd *cobra.Command, args []string) {
		outPath, _ := cmd.Flags().GetString("output")

		store, err := sqlite.New(rootCtx, dbPath)
		if err != nil {
			FatalError("failed to open database: %v", err)
		}
		defer store.Close()
		db := store.UnderlyingDB()

		export, err := buildGraphExport(db)
		if err != nil {
			FatalError("export failed: %v", err)
		}

		data, err := json.MarshalIndent(export, "", "  ")
		if err != nil {
			FatalError("failed to encode JSON: %v", err)
		}
		if outPath != "" {
			// #nosec G306 - graph export is non-sensitive, readable by design
			if err := os.WriteFile(outPath, append(data, '\n'), 0644); err != nil {
				FatalError("failed to write %s: %v", outPath, err)
			}
			fmt.Fprintf(os.Stderr, "Exported %d entities, %d edges, %d sessions to %s\n",
				len(export.Entities), len(export.Edges), len(export.Sessions), outPath)
			return
		}
		fmt.Println(string(data))
	},
}

type graphJSONExport struct {
	ExportedAt string               `json:"exported_at"`
	Entities   []graphExportEntity  `json:"entities"`
	Edges      []graphExportEdge    `json:"edges"`
	Sessions   []graphExportSession `json:"sessions"`
}

type graphExportEntity struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	PreferredName string  `json:"preferred_name"`
	Type          string  `json:"type"`
	MentionCount  int     `json:"mention_count"`
	Confidence    float64 `json:"confidence"`
	Source        string  `json:"source"`
}

type graphExportEdge struct {
	From         string  `json:"from"`
	To           string  `json:"to"`
	Relationship string  `json:"relationship"`
	DiscoveredIn string  `json:"discovered_in"`
	Confidence   float64 `json:"confidence"`
	Source       string  `json:"source"`
}

type graphExportSession struct {
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	Title     string `json:"title"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
}

func buildGraphExport(db *sql.DB) (*graphJSONExport, error) {
	out := &graphJSONExport{ExportedAt: time.Now().UTC().Format(time.RFC3339)}

	eRows, err := db.QueryContext(rootCtx, `SELECT id, name, COALESCE(preferred_name,''),
		COALESCE(type,''), mention_count, confidence, COALESCE(source,'')
		FROM entities ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer eRows.Close()
	for eRows.Next() {
		var e graphExportEntity
		if err := eRows.Scan(&e.ID, &e.Name, &e.PreferredName, &e.Type, &e.MentionCount, &e.Confidence, &e.Source); err != nil {
			return nil, err
		}
		out.Entities = append(out.Entities, e)
	}

	dRows, err := db.QueryContext(rootCtx, `SELECT from_entity, to_entity, COALESCE(relationship,''),
		COALESCE(discovered_in,''), confidence, COALESCE(source,'')
		FROM entity_deps ORDER BY from_entity, to_entity, relationship`)
	if err != nil {
		return nil, err
	}
	defer dRows.Close()
	for dRows.Next() {
		var d graphExportEdge
		if err := dRows.Scan(&d.From, &d.To, &d.Relationship, &d.DiscoveredIn, &d.Confidence, &d.Source); err != nil {
			return nil, err
		}
		out.Edges = append(out.Edges, d)
	}

	sRows, err := db.QueryContext(rootCtx, `SELECT id, COALESCE(filename,''), COALESCE(title,''),
		COALESCE(timestamp,''), COALESCE(type,'')
		FROM sessions WHERE is_ghost = 0 ORDER BY timestamp, id`)
	if err != nil {
		return nil, err
	}
	defer sRows.Close()
	for sRows.Next() {
		var s graphExportSession
		if err := sRows.Scan(&s.ID, &s.Filename, &s.Title, &s.Timestamp, &s.Type); err != nil {
			return nil, err
		}
		out.Sessions = append(out.Sessions, s)
	}

	return out, nil
}
