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
// entityMeta is the per-entity detail shown in the viewer's click panel.
type entityMeta struct {
	Subject string `json:"subject"`
	Date    string `json:"date"`
	Snippet string `json:"snippet"`
}

// buildEntityMeta maps each entity name to its most recent (non-ghost) devlog
// session — the subject, date, and a short narrative snippet — for the viewer's
// click panel. Keyed by the same displayed name the graph nodes use
// (preferred_name || name). Best-effort: returns nil on error.
func buildEntityMeta(ctx context.Context, db *sql.DB) map[string]entityMeta {
	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(e.preferred_name, e.name) AS name, s.title, s.timestamp, s.narrative
		FROM entities e
		JOIN session_entities se ON se.entity_id = e.id
		JOIN sessions s ON s.id = se.session_id
		WHERE COALESCE(s.is_ghost, 0) = 0
		ORDER BY s.timestamp DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	meta := make(map[string]entityMeta)
	for rows.Next() {
		var name, title, ts, narrative string
		if rows.Scan(&name, &title, &ts, &narrative) != nil {
			continue
		}
		if _, seen := meta[name]; seen {
			continue // rows are newest-first, so the first per name is the latest
		}
		date := ts
		if len(ts) >= 10 {
			date = ts[:10]
		}
		meta[name] = entityMeta{Subject: title, Date: date, Snippet: snippetOf(narrative, 220)}
	}
	return meta
}

// snippetOf returns a single-line, length-capped excerpt of s.
func snippetOf(s string, max int) string {
	s = strings.TrimSpace(strings.Join(strings.Fields(s), " "))
	if len(s) > max {
		s = strings.TrimSpace(s[:max]) + "…"
	}
	return s
}

func writeGraphHTML(path string, exports []graphExport, meta map[string]entityMeta) error {
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
	metaOut := make(map[string]entityMeta)
	for _, n := range nodes {
		nodeList = append(nodeList, n)
		if meta != nil {
			if m, ok := meta[n.ID]; ok {
				metaOut[n.ID] = m
			}
		}
	}
	data, err := json.Marshal(map[string]interface{}{"nodes": nodeList, "links": links})
	if err != nil {
		return err
	}
	metaJSON, _ := json.Marshal(metaOut)
	// Title lists the focused entities, but only when there are a few — a
	// whole-graph export has one "root" per source entity, and joining all of
	// them produced a screen-filling wall of names.
	label := "BeadsLog Graph"
	if len(roots) > 0 && len(roots) <= 4 {
		label += ": " + strings.Join(roots, ", ")
	}
	title, _ := json.Marshal(label)

	html := strings.NewReplacer(
		"__DATA__", string(data),
		"__TITLE__", string(title),
		"__NODEMETA__", string(metaJSON),
	).Replace(graphHTMLTemplate)

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
  body { margin: 0; background: #1e1e2e; font-family: -apple-system, sans-serif; overflow: hidden; }
  #title { position: absolute; top: 12px; left: 16px; color: #cdd6f4; z-index: 1; font-size: 14px; opacity: 0.8; max-width: 55%; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  #legend { position: absolute; bottom: 12px; left: 16px; color: #6c7086; z-index: 1; font-size: 11px; }
  #stats { position: absolute; bottom: 12px; right: 14px; color: #6c7086; z-index: 1; font-size: 11px; }
  #toolbar { position: absolute; top: 10px; right: 12px; z-index: 2; display: flex; gap: 8px; align-items: center; }
  #toolbar input[type=text] { background: #313244; border: 1px solid #45475a; color: #cdd6f4; border-radius: 6px; padding: 4px 8px; font-size: 12px; width: 150px; outline: none; }
  #toolbar button, #toolbar label { background: #313244; border: 1px solid #45475a; color: #cdd6f4; border-radius: 6px; padding: 4px 8px; font-size: 12px; cursor: pointer; user-select: none; }
  #panel { position: absolute; top: 0; right: -340px; width: 300px; height: 100%; background: rgba(24,24,37,0.96); color: #cdd6f4; z-index: 3; padding: 16px 18px; box-sizing: border-box; overflow-y: auto; transition: right .18s ease; box-shadow: -2px 0 14px #0007; font-size: 13px; }
  #panel.open { right: 0; }
  #panel h3 { margin: 4px 0 2px; color: #89b4fa; word-break: break-all; font-size: 15px; }
  #panel .deg { color: #a6adc8; font-size: 11px; margin-bottom: 8px; }
  #panel .rel { margin-top: 12px; color: #f9e2af; font-size: 10px; text-transform: uppercase; letter-spacing: .06em; }
  #panel a.nb { display: block; padding: 3px 0; color: #cdd6f4; text-decoration: none; cursor: pointer; border-bottom: 1px solid #2a2a3a; word-break: break-all; }
  #panel a.nb:hover { color: #89b4fa; }
  #panel a.nb span { color: #6c7086; float: right; margin-left: 8px; }
  #panel .pclose { float: right; cursor: pointer; color: #6c7086; font-size: 20px; line-height: 1; }
  #panel .devlog { margin-top: 14px; padding: 10px; background: #11111b; border-radius: 6px; border-left: 3px solid #cba6f7; }
  #panel .devlog .dl-sub { color: #cba6f7; font-size: 12px; font-weight: 600; }
  #panel .devlog .dl-date { color: #6c7086; font-size: 10px; margin: 2px 0 6px; }
  #panel .devlog .dl-snip { color: #a6adc8; font-size: 11px; line-height: 1.45; }
  #searchcount { color: #6c7086; font-size: 11px; min-width: 34px; text-align: center; }
</style>
</head>
<body>
<div id="title"></div>
<div id="toolbar">
  <input type="text" id="search" placeholder="search entity…" autocomplete="off" />
  <span id="searchcount"></span>
  <button id="fit" title="Fit graph to screen">Fit</button>
  <label title="Hide nodes with fewer connections (tames big graphs)">min&nbsp;<input type="range" id="degSlider" min="0" max="10" value="0" style="vertical-align:middle;width:70px" /> <span id="degVal">0</span></label>
  <label><input type="checkbox" id="coToggle" checked /> co-occ</label>
</div>
<div id="legend">solid = explicit dependency &nbsp;·&nbsp; dashed = co-occurrence &nbsp;·&nbsp; hover/click to focus &nbsp;·&nbsp; Enter cycles search</div>
<div id="stats"></div>
<div id="panel"></div>
<div id="graph"></div>
<script>
var data = __DATA__;
var title = __TITLE__;
var nodeMeta = __NODEMETA__;
document.getElementById('title').textContent = title;
var palette = ['#f38ba8', '#89b4fa', '#a6e3a1', '#f9e2af', '#cba6f7', '#94e2d5'];

// Compute degree + neighbor lists in-browser from the raw links (ids are strings here).
var degree = {}, neighbors = {};
data.links.forEach(function(l) {
  var s = l.source, t = l.target;
  degree[s] = (degree[s] || 0) + 1;
  degree[t] = (degree[t] || 0) + 1;
  (neighbors[s] = neighbors[s] || []).push({ id: t, rel: l.label, dir: 'out' });
  (neighbors[t] = neighbors[t] || []).push({ id: s, rel: l.label, dir: 'in' });
});
// Anchor labels: only the top hubs by degree stay labeled at rest (keeps big graphs clean).
var hubIds = {};
Object.keys(degree).sort(function(a, b) { return degree[b] - degree[a]; }).slice(0, 12)
  .forEach(function(id) { hubIds[id] = true; });

// Highlight state. Hover is transient; click sets a persistent focus. The
// active node + its neighbor set are recomputed only on change, so per-frame
// rendering is O(1) — essential for large graphs to stay smooth.
var hoverId = null, focusId = null, selId = null, search = '', showCo = true, minDeg = 0, frozen = false;
var activeCur = null, activeSet = {};
function recomputeActive() {
  activeCur = hoverId || focusId;
  activeSet = {};
  if (activeCur) { (neighbors[activeCur] || []).forEach(function(n) { activeSet[n.id] = true; }); }
}
function endId(e) { return (e && e.id !== undefined) ? e.id : e; }
function linkTouches(l, aid) { return endId(l.source) === aid || endId(l.target) === aid; }
function nodeShown(id) { return (degree[id] || 0) >= minDeg; }

var Graph = ForceGraph()(document.getElementById('graph'))
  .graphData(data)
  .backgroundColor('#1e1e2e')
  .cooldownTime(6000)
  .autoPauseRedraw(false) // draw-only repaints (physics is frozen after layout) → reliable highlight, cheap
  .nodeVal(function(n) { return Math.max(2, (degree[n.id] || 0)); })
  .nodeLabel(function(n) { return n.id + ' (' + (degree[n.id] || 0) + ')'; })
  .nodeVisibility(function(n) { return nodeShown(n.id); })
  .nodeCanvasObject(function(node, ctx, scale) {
    var deg = degree[node.id] || 0;
    var r = Math.max(2, Math.sqrt(deg + 1) * 1.7);
    var active = !activeCur || node.id === activeCur || activeSet[node.id];
    ctx.globalAlpha = active ? 1 : 0.12;
    ctx.beginPath();
    ctx.arc(node.x, node.y, r, 0, 2 * Math.PI);
    ctx.fillStyle = (node.id === selId) ? '#f9e2af'
      : (search && node.id.toLowerCase().indexOf(search) >= 0) ? '#a6e3a1'
      : palette[node.group % palette.length];
    ctx.fill();
    var isHub = hubIds[node.id];
    var labelIt = active && (isHub || node.id === selId
      || (activeCur && (node.id === activeCur || activeSet[node.id]))
      || (search && node.id.toLowerCase().indexOf(search) >= 0)
      || scale > 2.2);
    if (labelIt) {
      ctx.font = (isHub ? 4.5 : 3.5) + 'px sans-serif';
      ctx.textAlign = 'center';
      ctx.fillStyle = '#cdd6f4';
      ctx.fillText(node.id, node.x, node.y + r + 4);
    }
    ctx.globalAlpha = 1;
  })
  .linkVisibility(function(l) {
    return (showCo || !l.dashed) && nodeShown(endId(l.source)) && nodeShown(endId(l.target));
  })
  .linkColor(function(l) {
    if (!activeCur) return '#45475a';
    return linkTouches(l, activeCur) ? '#89b4fa' : '#26263a';
  })
  .linkWidth(function(l) {
    if (!activeCur) return 0.6;
    return linkTouches(l, activeCur) ? 1.6 : 0.4;
  })
  .linkLabel('label')
  .linkDirectionalArrowLength(function(l) { return l.dashed ? 0 : 3; })
  .linkLineDash(function(l) { return l.dashed ? [2, 2] : null; })
  .onNodeHover(function(node) {
    hoverId = node ? node.id : null;
    recomputeActive();
    document.getElementById('graph').style.cursor = node ? 'pointer' : '';
  })
  .onNodeClick(function(node) { focusId = node.id; recomputeActive(); showPanel(node.id); })
  .onNodeDrag(function(n) { n.fx = n.x; n.fy = n.y; })
  .onNodeDragEnd(function(n) { n.fx = n.x; n.fy = n.y; })
  .onBackgroundClick(function() { focusId = null; recomputeActive(); closePanel(); })
  // Once the initial layout settles, pin every node so dragging one node
  // doesn't reheat the whole n-body simulation — the fix for large graphs
  // locking up on drag.
  .onEngineStop(function() {
    if (frozen) return;
    frozen = true;
    Graph.graphData().nodes.forEach(function(n) { n.fx = n.x; n.fy = n.y; });
  });

function nodeById(id) {
  var ns = Graph.graphData().nodes;
  for (var i = 0; i < ns.length; i++) { if (ns[i].id === id) return ns[i]; }
  return null;
}

function showPanel(id) {
  selId = id;
  var nb = (neighbors[id] || []).slice();
  nb.sort(function(a, b) { return (degree[b.id] || 0) - (degree[a.id] || 0); });
  var byRel = {};
  nb.forEach(function(n) { (byRel[n.rel] = byRel[n.rel] || []).push(n); });
  var h = '<div class="pclose" onclick="closePanel()">&times;</div>';
  h += '<h3>' + id + '</h3>';
  h += '<div class="deg">' + (degree[id] || 0) + ' connection(s)</div>';
  Object.keys(byRel).forEach(function(rel) {
    h += '<div class="rel">' + (rel || 'related') + '</div>';
    byRel[rel].forEach(function(n) {
      h += '<a class="nb" data-id="' + n.id + '">' + (n.dir === 'out' ? '&rarr; ' : '&larr; ') + n.id + ' <span>' + (degree[n.id] || 0) + '</span></a>';
    });
  });
  var m = nodeMeta[id];
  if (m && m.subject) {
    h += '<div class="devlog"><div class="rel" style="margin-top:0">last devlog</div>';
    h += '<div class="dl-sub">' + esc(m.subject) + '</div>';
    if (m.date) { h += '<div class="dl-date">' + esc(m.date) + '</div>'; }
    if (m.snippet) { h += '<div class="dl-snip">' + esc(m.snippet) + '</div>'; }
    h += '</div>';
  }
  var p = document.getElementById('panel');
  p.innerHTML = h;
  p.classList.add('open');
  var links = p.querySelectorAll('a.nb');
  for (var i = 0; i < links.length; i++) {
    links[i].onclick = function() {
      var nn = nodeById(this.getAttribute('data-id'));
      if (nn) { Graph.centerAt(nn.x, nn.y, 500); Graph.zoom(5, 500); showPanel(nn.id); }
    };
  }
}
function closePanel() { document.getElementById('panel').classList.remove('open'); selId = null; }

function esc(s) {
  return String(s).replace(/[&<>"]/g, function(c) {
    return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[c];
  });
}

// Search cycles through ALL matching entities: type to filter, Enter for next.
var matches = [], matchIdx = -1;
function updateSearchCount() {
  var el = document.getElementById('searchcount');
  el.textContent = search ? (matches.length ? (matchIdx + 1) + '/' + matches.length : '0/0') : '';
}
function gotoMatch(i) {
  if (!matches.length) return;
  matchIdx = ((i % matches.length) + matches.length) % matches.length;
  var n = matches[matchIdx];
  Graph.centerAt(n.x, n.y, 500);
  Graph.zoom(5, 500);
  updateSearchCount();
}
var searchEl = document.getElementById('search');
searchEl.addEventListener('input', function(e) {
  search = e.target.value.trim().toLowerCase();
  matches = search ? Graph.graphData().nodes.filter(function(n) { return n.id.toLowerCase().indexOf(search) >= 0; }) : [];
  matchIdx = -1;
  if (matches.length) { gotoMatch(0); } else { updateSearchCount(); }
});
searchEl.addEventListener('keydown', function(e) {
  if (e.key === 'Enter') { e.preventDefault(); gotoMatch(matchIdx + (e.shiftKey ? -1 : 1)); }
});
document.getElementById('fit').addEventListener('click', function() { Graph.zoomToFit(500, 40); });
document.getElementById('coToggle').addEventListener('change', function(e) {
  showCo = e.target.checked;
  Graph.linkVisibility(Graph.linkVisibility()); // re-set accessor to force a refresh
});
document.getElementById('degSlider').addEventListener('input', function(e) {
  minDeg = parseInt(e.target.value, 10) || 0;
  document.getElementById('degVal').textContent = minDeg;
  Graph.nodeVisibility(Graph.nodeVisibility()); // re-apply node + link visibility
});

document.getElementById('stats').textContent = data.nodes.length + ' nodes · ' + data.links.length + ' links';
setTimeout(function() { Graph.zoomToFit(400, 50); }, 400);
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
		if err := writeGraphHTML(htmlPath, exports, buildEntityMeta(ctx, db)); err != nil {
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
