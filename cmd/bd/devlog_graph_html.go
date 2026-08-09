package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/untoldecay/BeadsLog/internal/queries"
	"github.com/untoldecay/BeadsLog/internal/ui"
)

// graphExport is one resolved entity with its explicit graph and co-occurrence
// neighbors, ready for HTML rendering.
type graphExport struct {
	Root  string
	Graph *queries.EntityGraph
	Co    []queries.CooccurrenceNode
}

type htmlGraphNode struct {
	ID    string `json:"id"`
	Group int    `json:"group"` // 0 = root, otherwise depth
	Val   int    `json:"val"`   // node size
}

type htmlGraphLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
	Dashed bool   `json:"dashed"` // co-occurrence (implicit) links
}

// writeGraphHTML renders an Obsidian-style interactive force-directed graph
// to a standalone HTML file (force-graph via CDN).
func writeGraphHTML(path string, exports []graphExport) error {
	nodes := make(map[string]htmlGraphNode)
	seenLinks := make(map[string]bool)
	var links []htmlGraphLink

	addNode := func(name string, group, val int) {
		if existing, ok := nodes[name]; !ok || group < existing.Group {
			nodes[name] = htmlGraphNode{ID: name, Group: group, Val: val}
		}
	}
	addLink := func(l htmlGraphLink) {
		key := l.Source + "|" + l.Target + "|" + l.Label
		if !seenLinks[key] && l.Source != l.Target {
			seenLinks[key] = true
			links = append(links, l)
		}
	}

	var roots []string
	for _, ex := range exports {
		roots = append(roots, ex.Root)
		addNode(ex.Root, 0, 8)

		if ex.Graph != nil {
			for _, n := range ex.Graph.Nodes {
				if n.Depth == 0 {
					continue
				}
				addNode(n.Name, n.Depth, 4)
				// Parent is the previous hop in the traversal path
				segments := strings.Split(n.Path, " → ")
				parent := ex.Root
				if len(segments) >= 2 {
					parent = segments[len(segments)-2]
				}
				addLink(htmlGraphLink{Source: parent, Target: n.Name, Label: n.Relationship})
			}
		}
		for _, co := range ex.Co {
			addNode(co.Name, 1, 4)
			addLink(htmlGraphLink{Source: ex.Root, Target: co.Name, Label: fmt.Sprintf("co-occurs ×%d", co.Count), Dashed: true})
		}
	}

	nodeList := make([]htmlGraphNode, 0, len(nodes))
	for _, n := range nodes {
		nodeList = append(nodeList, n)
	}
	data, err := json.Marshal(map[string]interface{}{"nodes": nodeList, "links": links})
	if err != nil {
		return err
	}
	title, _ := json.Marshal("BeadsLog Graph: " + strings.Join(roots, ", "))

	html := strings.NewReplacer("__DATA__", string(data), "__TITLE__", string(title)).Replace(graphHTMLTemplate)

	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, []byte(html), 0644)
}

const graphHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>BeadsLog Graph</title>
<script src="https://unpkg.com/force-graph@1"></script>
<style>
  body { margin: 0; background: #1e1e2e; font-family: -apple-system, sans-serif; }
  #title { position: absolute; top: 12px; left: 16px; color: #cdd6f4; z-index: 1; font-size: 14px; opacity: 0.8; }
  #legend { position: absolute; bottom: 12px; left: 16px; color: #6c7086; z-index: 1; font-size: 11px; }
</style>
</head>
<body>
<div id="title"></div>
<div id="legend">solid = explicit dependency &nbsp;·&nbsp; dashed = co-occurrence &nbsp;·&nbsp; drag / zoom / hover</div>
<div id="graph"></div>
<script>
const data = __DATA__;
const title = __TITLE__;
document.getElementById('title').textContent = title;
const palette = ['#f38ba8', '#89b4fa', '#a6e3a1', '#f9e2af', '#cba6f7', '#94e2d5'];
ForceGraph()(document.getElementById('graph'))
  .graphData(data)
  .backgroundColor('#1e1e2e')
  .nodeVal('val')
  .nodeLabel(n => n.id)
  .nodeCanvasObject((node, ctx, scale) => {
    const r = Math.sqrt(node.val) * 2;
    ctx.beginPath();
    ctx.arc(node.x, node.y, r, 0, 2 * Math.PI);
    ctx.fillStyle = palette[node.group % palette.length];
    ctx.fill();
    if (scale > 1.2 || node.group === 0) {
      ctx.font = (node.group === 0 ? 5 : 4) + 'px sans-serif';
      ctx.textAlign = 'center';
      ctx.fillStyle = '#cdd6f4';
      ctx.fillText(node.id, node.x, node.y + r + 5);
    }
  })
  .linkLabel('label')
  .linkColor(() => '#45475a')
  .linkDirectionalArrowLength(l => l.dashed ? 0 : 3)
  .linkLineDash(l => l.dashed ? [2, 2] : null);
</script>
</body>
</html>
`

// runFullGraph handles `bd devlog graph` with no entity: the whole graph.
// With htmlPath set, it exports every edge as an interactive force-graph
// (reusing the per-entity path over each edge source). Without it, it prints a
// compact summary — a 1000-node ASCII tree would be unreadable, so the terminal
// path deliberately summarizes rather than dumps.
func runFullGraph(ctx context.Context, db *sql.DB, htmlPath, relType string, openBrowser bool) {
	if htmlPath != "" {
		// Every entity that is the source of an edge becomes a root; depth 1 so
		// each contributes its direct edges. writeGraphHTML de-dupes into the
		// union = the full graph.
		rows, err := db.QueryContext(ctx,
			`SELECT DISTINCT COALESCE(e.preferred_name, e.name)
			 FROM entity_deps d JOIN entities e ON d.from_entity = e.id`)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading graph: %v\n", err)
			os.Exit(1)
		}
		var sources []string
		for rows.Next() {
			var n string
			if rows.Scan(&n) == nil {
				sources = append(sources, n)
			}
		}
		rows.Close()

		var exports []graphExport
		for _, name := range sources {
			g, _ := queries.GetEntityGraphExact(ctx, db, strings.ToLower(name), 1, relType)
			if g != nil && len(g.Nodes) > 1 {
				exports = append(exports, graphExport{Root: name, Graph: g})
			}
		}
		if len(exports) == 0 {
			fmt.Println("Graph is empty — no explicit dependencies recorded yet.")
			return
		}
		if err := writeGraphHTML(htmlPath, exports); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing HTML graph: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Full graph exported: %s (%d source entities)\n", htmlPath, len(exports))
		if openBrowser {
			openExportedGraph(htmlPath)
		}
		return
	}

	// Terminal summary.
	var entityCount, edgeCount int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM entities").Scan(&entityCount)
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM entity_deps").Scan(&edgeCount)

	fmt.Printf("\n=== Whole Graph ===\n")
	fmt.Printf("Entities: %d   Explicit edges: %d\n", entityCount, edgeCount)

	rows, err := db.QueryContext(ctx,
		`SELECT COALESCE(e.preferred_name, e.name) AS name, COUNT(*) AS deg
		 FROM entity_deps d JOIN entities e ON d.from_entity = e.id
		 GROUP BY d.from_entity ORDER BY deg DESC LIMIT 15`)
	if err == nil {
		defer rows.Close()
		fmt.Printf("\nMost-connected entities (by outgoing edges):\n")
		any := false
		for rows.Next() {
			var name string
			var deg int
			if rows.Scan(&name, &deg) == nil {
				any = true
				fmt.Printf("  %-32s %d\n", name, deg)
			}
		}
		if !any {
			fmt.Println("  (no explicit edges yet)")
		}
	}
	fmt.Printf("\n%s Tip: run with --html <path> to export the full interactive graph, or pass an entity to focus.\n", ui.RenderAccent("💡"))
}

// browserOpenCommand returns the platform command + args to open a path in the
// default application. Pure (goos-parameterized) so it can be unit-tested
// without launching anything.
func browserOpenCommand(goos, path string) (string, []string) {
	switch goos {
	case "darwin":
		return "open", []string{path}
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", path}
	default: // linux, bsd, etc.
		return "xdg-open", []string{path}
	}
}

// openInBrowser best-effort launches the default handler for path. Failures are
// returned so callers can hint, but never fatal — export already succeeded.
func openInBrowser(path string) error {
	name, args := browserOpenCommand(runtime.GOOS, path)
	return exec.Command(name, args...).Start() // #nosec G204 - fixed opener + local file path
}

// openExportedGraph opens an exported graph file, printing a non-fatal hint if
// the platform opener isn't available (e.g. headless server).
func openExportedGraph(path string) {
	if err := openInBrowser(path); err != nil {
		fmt.Fprintf(os.Stderr, "  (could not open browser: %v — open %s manually)\n", err, path)
	}
}
