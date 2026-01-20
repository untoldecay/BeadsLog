# PRD: BeadsLog Enhanced Search with Multi-Tier Suggestions

**Version:** 1.0
**Date:** January 18, 2026
**Owner:** [Your Name]
**Status:** Planned

## 🎯 Problem Statement

Current search is **exact-match only**:
```
bd devlog search "nginix" → No results (dead end)
bd devlog search "modal" → 14 noisy results (no guidance)
bd devlog search "nginx" → 3 results (good, but no related context)
```

**Users lose:**
- Context (what's related to "nginx"?)
- Guidance (typo → dead end)
- Discovery (partial matches hidden)

## 🎨 Desired Experience

**Every search is smart, contextual, and actionable.**

```
$ bd devlog search "nginix"
No exact matches. Did you mean: nginx (distance: 1)? ⭐
🔄 Auto-searching "nginx"...
Found 3 sessions + 4 related entities (nginxconf, cloudron...)

$ bd devlog search "nginx"
Found 3 sessions.
💡 Related: nginxconf, cloudron, mcp-sse, proxy-buffering
🔗 Graph neighbors: cloudron (uses), mcp (affected)

$ bd devlog search "modal"
Found 14 sessions.
💡 Modal entities: managecolumnsmodal, addcolumnmodal (5 total)
🎯 Narrow: bd devlog search "managecolumnsmodal"

$ bd devlog search "database timeout"
No sessions found.
💡 Try: postgres, mcp-connection-timeout, nginx timeout
```

## 🏗️ Solution Architecture

### **4-Tier Unified Search Flow**

```
User query →  [github](https://github.com/agnivade/levenshtein) Exact + Graph →  [github](https://github.com/lithammer/fuzzysearch) Typo →  [ppl-ai-file-upload.s3.amazonaws](https://ppl-ai-file-upload.s3.amazonaws.com/web/direct-files/collection_cf0ddef4-8af9-48da-8003-f9c2b7e7a75c/182b3b3b-d956-4e36-b2b5-fb5151353dc9/how-to-get-youtube-video-title-9TF4pwULR3eJrFbI1TBdDw.md) Fuzzy →  [ppl-ai-file-upload.s3.amazonaws](https://ppl-ai-file-upload.s3.amazonaws.com/web/direct-files/attachments/1600278/9849c7c0-4f76-4796-abaf-bc96318d8666/index.md) Fallback
     ↓              ↓             ↓          ↓           ↓
Sessions +     Related        Did you    Related     Smart
Related     entities (graph)   mean?     entities    suggestions
entities
```

**Tier 1: Exact Match + Graph Context** (Always run)
```
If "nginx" is exact entity:
→ Show sessions mentioning "nginx"
→ Plus: Graph neighbors (cloudron, nginxconf)
```

**Tier 2: Typo Detection** (If no exact match)
```
"nginix" → Levenshtein distance ≤ 3 to "nginx"
→ "Did you mean nginx?" + auto-search
```

**Tier 3: Fuzzy/Substring** (Always run)
```
"modal" → managecolumnsmodal, addcolumnmodal, rowdetailmodal
→ "Related entities: (5 total)"
```

**Tier 4: Smart Fallback** (If no results)
```
"database timeout" → postgres + timeout → mcp-connection-timeout
→ "Try: bd devlog search 'postgres'"
```

### **Libraries**
| Library | Purpose | Size | Speed |
|---------|---------|------|-------|
| `agnivade/levenshtein` | Typo detection | 300 lines | 350ns/op  [github](https://github.com/agnivade/levenshtein) |
| `lithammer/fuzzysearch` | Fuzzy matching | 7KB | ~1µs/op  [github](https://github.com/lithammer/fuzzysearch) |

### **Data Sources**
- **Sessions**: FTS5 index on `_index.md` + devlog files
- **Entities**: SQLite `entities` table (managecolumnsmodal, nginx, etc.)
- **Graph**: SQLite relations (`entities_relationships` table)

***

## 📊 Success Metrics

| Metric | Current | Target |
|--------|---------|--------|
| **Zero-result searches** | 25% | <5% |
| **First-search success** | 65% | 90% |
| **Context discovery** | 0% | 80% (shows related entities) |
| **Search latency** | 50ms | <100ms (with suggestions) |
| **Token efficiency** | Baseline | +20% (fewer follow-up queries) |

***

## 🎨 Output Format

### **Template 1: Results + Context**
```
┌─────────────────────────────────────────────────────────┐
│ 🔍 Search: "nginx"                                      │
├─────────────────────────────────────────────────────────┤
│ 💡 Related entities: nginxconf, cloudron, mcp-sse       │
│ 🔗 Graph neighbors: cloudron (uses), mcp (affected)     │
├─────────────────────────────────────────────────────────┤
│ Found 3 sessions:                                       │
│ 1. [fix] Nginx buffering fix...                         │
└─────────────────────────────────────────────────────────┘
```

### **Template 2: Typo Correction**
```
┌─────────────────────────────────────────────────────────┐
│ 🔍 Search: "nginix"                                     │
├─────────────────────────────────────────────────────────┤
│ ⚠️  No exact matches. Did you mean: nginx ⭐           │
│ 🔄 Auto-searching "nginx"...                            │
│ Found 3 sessions:                                       │
└─────────────────────────────────────────────────────────┘
```

### **Template 3: No Results + Suggestions**
```
┌─────────────────────────────────────────────────────────┐
│ 🔍 Search: "database timeout"                           │
├─────────────────────────────────────────────────────────┤
│ ⚠️  No sessions found.                                  │
│ 💡 Try these:                                           │
│   • postgres                                            │
│   • mcp-connection-timeout                              │
│   • nginx timeout                                       │
└─────────────────────────────────────────────────────────┘
```

***

## 🔧 Technical Implementation

### **Phase 1: Core (1 day)**
```
[ ] Add levenshtein library
[ ] Add fuzzysearch library
[ ] Implement findClosestEntity() → "Did you mean"
[ ] Add --exact flag
[ ] Basic "related entities" (substring match)
```

### **Phase 2: Graph Integration (1 day)**
```
[ ] generateRelatedSuggestions() → Graph neighbors
[ ] findPartialMatches() → Fuzzy + substring
[ ] Ranking system (score by distance + source)
```

### **Phase 3: Smart Fallback (0.5 day)**
```
[ ] findCompoundEntities() → Multi-word query splitting
[ ] CLI formatter (box-drawing UI)
[ ] Actionable hints ("Try bd devlog search X")
```

### **Database Schema** (if needed)
```sql
-- Add to existing entities table
ALTER TABLE entities ADD COLUMN frequency INTEGER DEFAULT 1;

-- Relationships table (if not exists)
CREATE TABLE IF NOT EXISTS entity_relationships (
    source_entity TEXT,
    target_entity TEXT,
    relationship_type TEXT,  -- "uses", "affects", "configures"
    session_id TEXT,
    FOREIGN KEY(source_entity) REFERENCES entities(name),
    FOREIGN KEY(target_entity) REFERENCES entities(name)
);
```

***

## 🎯 User Stories

**Agent:**
```
As an agent, when I type "nginix", I want to see nginx results + learn the correct spelling.
```

**Developer:**
```
As a dev, when I search "modal", I want to see all 5 modal components + graph relationships.
```

**New team member:**
```
As a new hire, when I search "timeout", I want related entities and actionable next steps.
```

**PM:**
```
As a PM, I want every search to surface relevant context, not dead ends.
```

***

## 🚀 Launch Plan

**MVP (Week 1):**
- Typo detection + "Did you mean"
- Basic substring suggestions
- CLI formatter

**Full Release (Week 2):**
- Graph integration
- Fuzzy matching
- Multi-word fallback
- Performance tests

**Success criteria:**
- 90% of searches show suggestions
- <100ms latency on 1000 entities
- 50% reduction in "no results" searches

***

## 📈 Competitive Advantage

| Feature | **BeadsLog** | **grep** | **VSCode Ctrl+T** | **ByteRover** |
|---------|-------------|----------|-------------------|--------------|
| Typo correction | ✅ Levenshtein | ❌ | ✅ | ✅ |
| Fuzzy matching | ✅ Sublime-style | ❌ | ✅ | ✅ |
| Graph context | ✅ Unique | ❌ | ❌ | ❌ |
| Entity ranking | ✅ Multi-source | ❌ | ✅ | ✅ |
| Codebase aware | ✅ Devlog + graph | ❌ | ✅ | ✅ |

**BeadsLog unique:** **Graph-enhanced suggestions** - "nginx" → shows cloudron, mcp-sse, proxy-buffering from architectural relationships.

***

## 🎯 Summary

**Transform search from "keyword lookup" to "context discovery machine".**

Every search becomes:
- **Educational** (typo correction)
- **Contextual** (graph + related entities)
- **Actionable** (copy-paste next commands)
- **Never a dead end** (smart fallbacks)

**Total effort:** ~2 days, ~300 lines of code, 2 tiny libraries.

**Value:** 3x better search experience, 50% fewer follow-up queries, unique graph intelligence.

***

**Approved:** [ ]
**Implemented:** [ ]
**Docs updated:** [ ]

***

***
