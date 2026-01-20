Based on the video and your existing PRD, here's an update PRD focused on migrating the search render to Lip Gloss tables:

***

# PRD: BeadsLog Search Render Migration to Lip Gloss Tables

**Version:** 1.1
**Date:** January 20, 2026
**Parent PRD:** BeadsLog Enhanced Search with Multi-Tier Suggestions v1.0
**Status:** Planned

## 🎯 Problem Statement

Current search output uses **manual box-drawing characters** for CLI tables:
```go
// Current implementation (verbose, brittle)
fmt.Println("┌─────────────────────────────────────────────────────────┐")
fmt.Println("│ 🔍 Search: \"nginx\"                                      │")
fmt.Println("├─────────────────────────────────────────────────────────┤")
// ... 50+ lines per template
```

**Issues:**
- ~100 lines of code per template (3 templates = 300 lines)
- Manual width calculations break on terminal resize
- No text wrapping for long entity names/sessions
- Conditional styling (typo ⭐, warnings ⚠️) requires string concatenation
- Maintenance burden for any layout changes

## 🎨 Desired Experience

**Declarative, CSS-like table rendering** using Lip Gloss: [youtube](https://www.youtube.com/watch?v=ss-DOiHrEjM)

```go
t := table.New().
  Headers("🔍 Search", query).
  Border(lipgloss.RoundedBorder()).
  Width(termWidth).
  Rows(rows...).
  StyleFunc(conditionalStyles)
```

**Benefits:**
- ~20 lines per template (80% reduction)
- Auto-width adaptation to terminal size
- Built-in text wrapping
- `StyleFunc` for row/column conditional formatting
- Consistent with Charm ecosystem (if using Bubble Tea elsewhere)

## 🏗️ Technical Implementation

### **New Dependency**

```bash
go get github.com/charmbracelet/lipgloss@latest
```

The table package is included in `lipgloss/table` submodule. [pkg.go](https://pkg.go.dev/github.com/charmbracelet/lipgloss/table)

### **Migration Plan**

| Template | Current Lines | Target Lines | Complexity |
|----------|---------------|--------------|------------|
| Results + Context | ~100 | ~25 | Low |
| Typo Correction | ~80 | ~20 | Low |
| No Results + Suggestions | ~70 | ~20 | Low |

### **Phase 1: Core Table Renderer (0.5 day)**

```
[ ] Add lipgloss dependency
[ ] Create render/table.go with base styles
[ ] Implement NewSearchTable(width int) helper
[ ] Define border styles (Rounded for results, Normal for errors)
```

**Base styles:**
```go
package render

import (
    "github.com/charmbracelet/lipgloss"
    "github.com/charmbracelet/lipgloss/table"
)

var (
    HeaderStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("99")).
        Align(lipgloss.Center)
    
    WarningStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("214")) // Orange
    
    SuccessStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("82"))  // Green
    
    HintStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("244")) // Gray
)
```

### **Phase 2: Template Migration (0.5 day)**

```
[ ] Migrate Template 1: RenderResultsWithContext()
[ ] Migrate Template 2: RenderTypoCorrection()
[ ] Migrate Template 3: RenderNoResults()
[ ] Add StyleFunc for conditional row highlighting
```

**Template 1 Implementation:**
```go
func RenderResultsWithContext(query string, sessions []Session, related []string, neighbors []string, width int) string {
    rows := [][]string{
        {"💡 Related entities:", strings.Join(related, ", ")},
        {"🔗 Graph neighbors:", formatNeighbors(neighbors)},
        {fmt.Sprintf("Found %d sessions:", len(sessions)), ""},
    }
    
    for i, s := range sessions {
        rows = append(rows, []string{
            fmt.Sprintf("%d. [%s]", i+1, s.Type),
            truncate(s.Title, width-20),
        })
    }
    
    return table.New().
        Headers("🔍 Search", fmt.Sprintf("\"%s\"", query)).
        Border(lipgloss.RoundedBorder()).
        BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
        Width(width).
        Rows(rows...).
        StyleFunc(func(row, col int) lipgloss.Style {
            switch {
            case row == table.HeaderRow:
                return HeaderStyle
            case row <= 2:
                return HintStyle
            default:
                return lipgloss.NewStyle().Padding(0, 1)
            }
        }).
        String()
}
```

**Template 2 Implementation:**
```go
func RenderTypoCorrection(query, corrected string, sessions []Session, width int) string {
    rows := [][]string{
        {"⚠️ No exact matches.", fmt.Sprintf("Did you mean: %s ⭐", corrected)},
        {"🔄 Auto-searching:", fmt.Sprintf("\"%s\"...", corrected)},
        {fmt.Sprintf("Found %d sessions:", len(sessions)), ""},
    }
    
    for i, s := range sessions[:min(5, len(sessions))] {
        rows = append(rows, []string{fmt.Sprintf("%d.", i+1), s.Title})
    }
    
    return table.New().
        Headers("🔍 Search", fmt.Sprintf("\"%s\"", query)).
        Border(lipgloss.RoundedBorder()).
        Width(width).
        Rows(rows...).
        StyleFunc(func(row, col int) lipgloss.Style {
            if row == 0 { return WarningStyle }
            if row == 1 { return SuccessStyle }
            return lipgloss.NewStyle()
        }).
        String()
}
```

**Template 3 Implementation:**
```go
func RenderNoResults(query string, suggestions []string, width int) string {
    rows := [][]string{
        {"⚠️ No sessions found.", ""},
        {"💡 Try these:", ""},
    }
    
    for _, s := range suggestions {
        rows = append(rows, []string{"  •", s})
    }
    
    return table.New().
        Headers("🔍 Search", fmt.Sprintf("\"%s\"", query)).
        Border(lipgloss.RoundedBorder()).
        Width(width).
        Rows(rows...).
        StyleFunc(func(row, col int) lipgloss.Style {
            if row == 0 { return WarningStyle }
            if row == 1 { return HintStyle.Bold(true) }
            return HintStyle
        }).
        String()
}
```

### **Phase 3: Terminal Width Integration (0.25 day)**

```
[ ] Add golang.org/x/term for terminal size detection
[ ] Create getTerminalWidth() helper with fallback (80)
[ ] Wire width into all render calls
```

```go
import "golang.org/x/term"

func getTerminalWidth() int {
    if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
        return w
    }
    return 80 // fallback
}
```

### **Phase 4: Testing & Cleanup (0.25 day)**

```
[ ] Unit tests for each template with mock data
[ ] Visual regression tests (golden files)
[ ] Remove old box-drawing code
[ ] Update any docs/screenshots
```

***

## 📊 Success Metrics

| Metric | Before | After |
|--------|--------|-------|
| **Lines of code (render)** | ~300 | ~75 |
| **Terminal resize support** | ❌ | ✅ |
| **Text wrapping** | Manual | Auto |
| **Style changes** | Edit strings | Edit Style vars |
| **New template effort** | ~1 hour | ~15 min |

***

## 🔧 Key Lip Gloss Concepts

From the Charm video: [youtube](https://www.youtube.com/watch?v=ss-DOiHrEjM)

1. **Static vs Interactive**: Lip Gloss is for static rendering (our use case). Bubbles is for interactive tables. We want static search output.

2. **StyleFunc**: The key function for conditional formatting. Receives `(row, col int)` and returns a `lipgloss.Style`. Use `table.HeaderRow` constant for header detection.

3. **Rows with Spread Operator**: Use `Rows(rows...)` with a `[][]string` slice. The API uses a `Data` interface internally to avoid type conversion issues.

4. **Border Styles**: `lipgloss.RoundedBorder()`, `NormalBorder()`, `ThickBorder()`, etc.

***

## 🎯 Checklist

```
[ ] go get github.com/charmbracelet/lipgloss
[ ] Create render/table.go
[ ] Define base styles (Header, Warning, Success, Hint)
[ ] Implement RenderResultsWithContext()
[ ] Implement RenderTypoCorrection()
[ ] Implement RenderNoResults()
[ ] Add terminal width detection
[ ] Write tests
[ ] Remove old render code
[ ] Update CHANGELOG
```

**Total effort:** ~1.5 days
**Risk:** Low (additive change, can keep old code as fallback)

***

**Approved:** [ ]
**Implemented:** [ ]
**Tests passing:** [ ]
