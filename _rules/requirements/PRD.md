# 🚀 COMPLETE DEVLOG IMPLEMENTATION PRD

**Summary**
The Devlog Beads project forks the Beads issue tracker into a graph-powered development session memory system that imports existing markdown devlogs (index.md + dated session files), automatically extracts software entities from problem descriptions, builds architectural dependency graphs between components/modal/hooks/endpoints, and enables hybrid text+graph retrieval to resume debugging with complete session history and related-context discovery across multi-session fixes/features/enhancements.

## 🎯 Phase 1: Fork & Schema (30min)

### 1. Fork Beads
```bash
git clone https://github.com/steveyegge/beads devlog
cd devlog
git remote add upstream https://github.com/steveyegge/beads
go mod tidy
```

### 2. Complete Schema (internal/db/schema.go)
```sql
-- Add after Beads' existing tables
CREATE TABLE sessions (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  timestamp DATETIME NOT NULL,
  status TEXT DEFAULT 'closed',
  type TEXT, -- fix, feature, enhance, etc.
  filename TEXT,
  narrative TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE entities (
  id TEXT PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  type TEXT DEFAULT 'component',
  first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
  mention_count INTEGER DEFAULT 1
);

CREATE TABLE session_entities (
  session_id TEXT,
  entity_id TEXT,
  relevance TEXT DEFAULT 'mentioned',
  PRIMARY KEY(session_id, entity_id),
  FOREIGN KEY(session_id) REFERENCES sessions(id),
  FOREIGN KEY(entity_id) REFERENCES entities(id)
);

CREATE TABLE entity_deps (
  from_entity TEXT,
  to_entity TEXT,
  relationship TEXT,
  discovered_in TEXT,
  PRIMARY KEY(from_entity, to_entity, relationship),
  FOREIGN KEY(from_entity) REFERENCES entities(id),
  FOREIGN KEY(to_entity) REFERENCES entities(id),
  FOREIGN KEY(discovered_in) REFERENCES sessions(id)
);
```

**Agent Task**: Replace `internal/db/schema.go` with Beads schema + above tables. Add `migrate()` function.

## 🛠️ Phase 2: Import Script (30min)

### cmd/devlog/import-md.go (COMPLETE)
```go
package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
    "time"

    "github.com/spf13/cobra"
    "github.com/steveyegge/beads/internal/db"
)

type IndexRow struct {
    Subject  string
    Problem  string
    Date     string
    Filename string
}

var importCmd = &cobra.Command{
    Use:   "import-md [index.md]",
    Short: "Import devlog index.md into sessions",
    Run: func(cmd *cobra.Command, args []string) {
        if len(args) == 0 {
            fmt.Println("Usage: devlog import-md index.md")
            os.Exit(1)
        }
        rows := parseIndexMD(args[0])
        fmt.Printf("Found %d sessions to import\n", len(rows))
        
        for _, row := range rows {
            sessionID := createSession(row)
            extractAndLinkEntities(sessionID, row.Problem)
        }
        
        db.ExportToJSONL()
        fmt.Printf("✅ Imported %d sessions, %d entities\n", len(rows), countEntities())
    },
}

func parseIndexMD(filename string) []IndexRow {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        panic(err)
    }
    
    lines := strings.Split(string(data), "\n")
    var rows []IndexRow
    inTable := false
    
    for i, line := range lines {
        line = strings.TrimSpace(line)
        if strings.Contains(line, "| Subject | Problems |") {
            inTable = true
            continue
        }
        if inTable && strings.Count(line, "|") >= 4 && !strings.HasPrefix(line, "|---") {
            parts := strings.Split(line, "|")
            if len(parts) >= 5 {
                rows = append(rows, IndexRow{
                    Subject:  strings.TrimSpace(parts[1]),
                    Problem:  strings.TrimSpace(parts[2]),
                    Date:     strings.TrimSpace(parts[3]),
                    Filename: strings.TrimSpace(parts[4]),
                })
            }
        }
    }
    return rows
}

func createSession(row IndexRow) string {
    sessionID := fmt.Sprintf("sess-%s", hashID(row.Subject+row.Date))
    
    session := map[string]interface{}{
        "id":        sessionID,
        "title":     row.Subject,
        "timestamp": parseDate(row.Date).Format(time.RFC3339),
        "status":    "closed",
        "type":      extractType(row.Subject),
        "filename":  row.Filename,
    }
    
    db.Create("sessions", session)
    return sessionID
}

func extractAndLinkEntities(sessionID, problem string) {
    entityPatterns := []*regexp.Regexp{
        regexp.MustCompile(`[A-Z][a-z]+(?:[A-Z][a-z]+)+`), // CamelCase
        regexp.MustCompile(`(?i)(modal|hook|endpoint|migration|service)`),
        regexp.MustCompile(`[a-z]+-[a-z]+`), // kebab-case
    }
    
    seen := make(map[string]bool)
    for _, pat := range entityPatterns {
        matches := pat.FindAllString(problem, -1)
        for _, match := range matches {
            if len(match) > 3 && !seen[match] {
                entityID := fmt.Sprintf("ent-%s", hashID(match))
                
                // Create/update entity
                db.Upsert("entities", map[string]interface{}{
                    "id":           entityID,
                    "name":         strings.ToLower(match),
                    "type":         "component",
                    "mention_count": db.GetInt("entities", "mention_count", "id=?", entityID) + 1,
                })
                
                // Link session → entity
                db.Create("session_entities", map[string]interface{}{
                    "session_id": sessionID,
                    "entity_id":  entityID,
                    "relevance":  "primary",
                })
                seen[match] = true
            }
        }
    }
}

func hashID(s string) string {
    h := fnv.New32a()
    h.Write([]byte(s))
    return fmt.Sprintf("%x", h.Sum32())[:6]
}

func parseDate(dateStr string) time.Time {
    layouts := []string{"2006-01-02", "Jan 2"}
    for _, layout := range layouts {
        if t, err := time.Parse(layout, dateStr); err == nil {
            return t
        }
    }
    return time.Now()
}

func extractType(subject string) string {
    prefixes := map[string]string{
        "fix":        "fix",
        "feature":    "feature", 
        "enhance":    "enhance",
        "rationalize": "chore",
        "deploy":     "deploy",
        "security":   "security",
        "debug":      "debug",
    }
    for prefix, typ := range prefixes {
        if strings.HasPrefix(strings.ToLower(subject), prefix) {
            return typ
        }
    }
    return "task"
}

func countEntities() int {
    return db.Count("entities")
}

func init() {
    rootCmd.AddCommand(importCmd)
}
```

**Agent Task**: Create this file exactly. Add `import "hash/fnv"`. Wire to `main.go`.

## 📊 Phase 3: Graph Queries (20min)

### internal/queries/graph.go
```go
func GetEntityGraph(entityName string, depth int) (*EntityGraph, error) {
    query := `
    WITH RECURSIVE graph(id, name, rel_type, depth, path) AS (
        SELECT e.id, e.name, '', 0, e.name
        FROM entities e WHERE LOWER(e.name) LIKE ? 
        
        UNION ALL 
        
        SELECT e.id, e.name, ed.relationship, g.depth+1, g.path || ' → ' || e.name
        FROM entities e 
        JOIN entity_deps ed ON e.id = ed.to_entity 
        JOIN graph g ON ed.from_entity = g.id
        WHERE g.depth < ? AND g.path NOT LIKE '%' || e.name || '%'
    )
    SELECT * FROM graph ORDER BY depth;
    `
    
    rows, err := db.Query(query, "%"+entityName+"%", depth)
    // Parse into EntityGraph struct
    return parseGraph(rows), err
}
```

### cmd/devlog/graph.go
```bash
devlog graph "manage columns" --depth 3
# Output:
# manage-columns-modal (0)
# └── usesortable-hook (1)
#     └── column-management (2)
```

## 🔍 Phase 4: CLI Commands (20min)

```bash
# Your exact workflow preserved
devlog list --type fix           # Your index.md filtered
devlog show 2025-11-29           # Full .md content
devlog search "migration"        # Text + graph

# Graph superpowers
devlog entities                  # Top entities by mention_count
devlog impact "nginx"            # What depends on nginx changes
devlog resume "mcp" --hybrid     # Session + 2-hop entity context
```

## 🚀 Phase 5: Test & Deploy (10min)

```bash
# 1. Init project
./bd init --quiet  # Becomes ./devlog init

# 2. Import your data
./devlog import-md index.md

# 3. Test graph
./devlog graph "manage columns"
./devlog list --type security

# 4. Git magic works
git add .devlog/
git commit -m "Devlog system live"
```

## 📋 Agent Checklist (Copy-Paste This)

```
[ ] 1. Fork beads → devlog
[ ] 2. Add schema tables to internal/db/schema.go  
[ ] 3. Create cmd/devlog/import-md.go (copy exact code above)
[ ] 4. Add internal/queries/graph.go
[ ] 5. Build: go build ./cmd/devlog
[ ] 6. Test: ./devlog init && ./devlog import-md index.md
[ ] 7. Verify: ./devlog list && ./devlog graph "migration"
[ ] 8. Push: git add . && git commit -m "Devlog v1.0"
```

## 🎉 Success Criteria

```
✅ 13663 chars index.md → 50+ sessions in DB
✅ "devlog graph manage-columns" shows relationships  
✅ "devlog list --type fix" matches your mental model
✅ .devlog/sessions.jsonl is git-tracked
✅ Full narratives preserved in filename links
✅ Ready for AI agents: devlog resume --json
```
