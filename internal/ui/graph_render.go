package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/untoldecay/BeadsLog/internal/queries"
)

// BuildEntityTree constructs a lipgloss/tree for an EntityGraph
func BuildEntityTree(graph *queries.EntityGraph) *tree.Tree {
	if len(graph.Nodes) == 0 {
		return nil
	}

	// First node is the root
	rootNode := graph.Nodes[0]
	t := tree.New().Root(fmt.Sprintf("%s (%d)", rootNode.Name, rootNode.Depth))
	
	// Set styles
	t.EnumeratorStyle(lipgloss.NewStyle().Foreground(ColorAccent))
	t.RootStyle(lipgloss.NewStyle().Bold(true).Foreground(ColorAccent))

	// Track paths to trees for nesting
	nodeMap := make(map[string]*tree.Tree)
	nodeMap[rootNode.Path] = t

	// Add nodes sorted by depth (guaranteed by SQL ORDER BY)
	for i := 1; i < len(graph.Nodes); i++ {
		node := graph.Nodes[i]
		
		// Find parent from path: "A → B → C" -> parent is "A → B"
		parts := strings.Split(node.Path, " → ")
		if len(parts) < 2 {
			continue 
		}
		parentPath := strings.Join(parts[:len(parts)-1], " → ")
		
		childTree := tree.New().Root(fmt.Sprintf("%s (%d)", node.Name, node.Depth))
		childTree.EnumeratorStyle(lipgloss.NewStyle().Foreground(ColorAccent))
		nodeMap[node.Path] = childTree
		
		if parentTree, ok := nodeMap[parentPath]; ok {
			parentTree.Child(childTree)
		} else {
			t.Child(childTree)
		}
	}

	return t
}

// RenderEntityTree renders an EntityGraph using lipgloss/tree
func RenderEntityTree(graph *queries.EntityGraph) string {
	t := BuildEntityTree(graph)
	if t == nil {
		return TableHintStyle.Render("No entities found.")
	}
	return t.String()
}

// RenderGraphTable renders multiple trees inside a single structured table
func RenderGraphTable(query string, matches []struct {
	Name  string
	Graph *queries.EntityGraph
}, width int) string {
	if len(matches) == 0 {
		return TableHintStyle.Render("No graphs to display.")
	}

	rows := [][]string{}
	for _, m := range matches {
		treeStr := RenderEntityTree(m.Graph)
		// We use a 2-column layout: [Entity Name] | [Tree]
		rows = append(rows, []string{
			lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(m.Name),
			treeStr,
		})
	}

	return table.New().
		Headers("Root Entity", fmt.Sprintf("Graph Analysis for %q", query)).
		Rows(rows...).
		Border(lipgloss.RoundedBorder()).
		BorderRow(true).
		BorderStyle(lipgloss.NewStyle().Foreground(ColorMuted)).
		Width(width).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				if col == 0 {
					return TableHeaderStyle.Width(25)
				}
				return TableHeaderStyle.Width(width - 25 - 3)
			}
			style := lipgloss.NewStyle().Padding(0, 1).Align(lipgloss.Left)
			if col == 0 {
				style = style.Border(lipgloss.NormalBorder(), false, true, false, false).
					BorderForeground(ColorMuted).
					Width(25).
					PaddingTop(1) // Align with first line of tree
			}
			return style
		}).
		String()
}

// RenderCooccurrenceTable renders a list of entities frequently mentioned together
func RenderCooccurrenceTable(target string, coNodes []queries.CooccurrenceNode, width int) string {
	if len(coNodes) == 0 {
		return ""
	}

	rows := [][]string{}
	for _, node := range coNodes {
		rows = append(rows, []string{
			node.Name,
			fmt.Sprintf("%d sessions", node.Count),
		})
	}

	return table.New().
		Headers(fmt.Sprintf("💡 Co-mentioned with %q", target), "Frequency").
		Rows(rows...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ColorMuted)).
		Width(width).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return TableHeaderStyle
			}
			style := lipgloss.NewStyle().Padding(0, 1)
			if col == 1 {
				style = style.Align(lipgloss.Right).Foreground(ColorMuted)
			} else {
				style = style.Align(lipgloss.Left)
			}
			return style
		}).
		String()
}

// RenderPath renders a path between two entities
func RenderPath(path []queries.PathStep, width int) string {
	if len(path) == 0 {
		return TableHintStyle.Render("No path found.")
	}

	var parts []string
	for i, step := range path {
		entity := lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(step.EntityName)
		parts = append(parts, entity)
		
		if i < len(path)-1 {
			session := lipgloss.NewStyle().Italic(true).Foreground(ColorMuted).Render(fmt.Sprintf("[%s] %s", step.SessionID, step.SessionTitle))
			parts = append(parts, "  via "+session+"  →")
		}
	}

	return lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Width(width).
		Render(strings.Join(parts, "\n"))
}
