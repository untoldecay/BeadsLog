# PRD: Devlog Entity Extraction with Ollama + Fallback

**Version:** 1.0  
**Date:** January 21, 2026  
**Owner:** [Your Name]  
**Status:** Planned

## 🎯 Problem Statement

**Current extraction (regex-based):**
```
Devlog session (500 words):
"Fixed nginx timeout in ManageColumnsModal. Changed proxy_buffering from on to off. Updated useSortable hook parameters. Cloudron deployment now stable."

Extracted entities: ["nginx", "modal"]
→ Missing: proxy_buffering, usesortable, cloudron, ManageColumnsModal
→ Graph incomplete, search misses connections
```

**Impact:**
- Poor graph coverage (50% entities missed)
- Weak search relevance (BM25 can't match "proxy buffering")
- Limited architectural awareness

## 🎨 Desired Experience

**Input:** Full devlog markdown session (500-2000 words)  
**Output:** 10-20 entities with types/relationships

```
Before: ["nginx", "modal"]
After: [
  "nginx" (config),
  "proxy_buffering" (nginx_setting), 
  "managecolumnsmodal" (component),
  "usesortable" (hook),
  "cloudron" (deployment),
  "nginx.conf" (file)
]
```

**Search improvement:**
```
bd devlog search "proxy buffering" → 3 sessions (vs 0 before)
bd devlog graph "nginx" → proxy_buffering, cloudron, mcp-sse
```

## 🏗️ Solution Architecture

### **2-Tier Pipeline**

```
Devlog Session (500-2000 words)
         ↓
   [Tier 1: Regex Fallback]
   (Extract known patterns)
         ↓
   [Tier 2: Ollama LLM]
   (Full semantic extraction)
         ↓
   [Deduplication + Ranking]
         ↓
   [SQLite entities table]
```

### **Tier 1: Regex Fallback (Current Codebase)**
- Reuse existing regex patterns
- Instant (1ms)
- Covers 60-70% common entities

**Patterns (extract from your codebase):**
```
nginx[\w-]*          # nginx, nginxconf
(?i)modal[\w]*       # ManageColumnsModal, addcolumnmodal
(use|api)[\w]+Service # authService, schemaCacheService
cloudron[\w]*        # cloudron deployment
mcp[\w]*             # mcp-sse, mcp-connection
proxy_[\w]*          # proxy_buffering, proxy_read_timeout
\d+\w+Modal          # AddColumnModal, RowDetailModal
use[\w]+Hook         # useAuthenticatedApi, useSortable
```

### **Tier 2: Ollama LLM (Primary)**
- Local LLM for semantic extraction
- 500ms on llama3.2:3b
- 85% accuracy vs 70% regex

**Prompt template:**
```
You are an entity extractor for a Go/React/PostgreSQL codebase.

From this devlog session, extract:
1. Components (Modal, Service, Hook)
2. Config files (nginx.conf, env vars)
3. Services (cloudron, mcp, postgres)
4. Technologies (pgvector, redis, jwt)
5. Files/patterns mentioned

Devlog:
```
{paste full devlog markdown}

```

Output JSON only:
```json
{{
  "entities": [
    {{"name": "nginx", "type": "config"}},
    {{"name": "proxy_buffering", "type": "nginx_setting"}},
    {{"name": "managecolumnsmodal", "type": "component"}}
  ]
}}
```
```

## 📊 Success Metrics

| Metric | Current (Regex) | Target (Ollama + Fallback) |
|--------|-----------------|----------------------------|
| **Entities/session** | 2-4 | 8-15 |
| **Extraction time** | 1ms | 500ms |
| **Graph coverage** | 60% | 90% |
| **Search relevance** | Baseline | +40% (BM25 on richer entities) |
| **Cost** | $0 | $0 (local Ollama) |

## 🔧 Technical Implementation

### **Database Schema**
```sql
-- entities table (extend existing)
ALTER TABLE entities ADD COLUMN confidence FLOAT DEFAULT 1.0;
ALTER TABLE entities ADD COLUMN source TEXT DEFAULT 'regex'; -- regex|ollama

-- Extraction log for debugging
CREATE TABLE extraction_log (
    session_id TEXT,
    timestamp DATETIME,
    extractor TEXT,  -- regex|ollama
    input_length INTEGER,
    entities_found INTEGER,
    FOREIGN KEY(session_id) REFERENCES sessions(id)
);
```

### **Extraction Pipeline**

```go
// internal/extractor/pipeline.go
func ExtractFromDevlog(session *Session) []Entity {
    // Tier 1: Regex (instant fallback)
    entities := regex.Extractor.Extract(session.Markdown)
    
    // Tier 2: Ollama (if available)
    if ollama.Available() {
        ollamaEntities := ollama.Extractor.Extract(session.Markdown)
        entities = mergeEntities(entities, ollamaEntities)
    }
    
    // Deduplicate + rank
    entities = deduplicate(entities)
    entities = rankByConfidence(entities)
    
    // Store with metadata
    storeEntities(session.ID, entities)
    
    return entities[:max(20, len(entities))]
}

func mergeEntities(regexEntities, ollamaEntities []Entity) []Entity {
    // Ollama entities get confidence boost
    for _, e := range ollamaEntities {
        e.Confidence *= 1.2
        e.Source = "ollama"
    }
    
    // Regex as fallback
    for _, e := range regexEntities {
        if !entityExists(e) {
            e.Confidence *= 0.8
            e.Source = "regex"
        }
    }
    
    return allEntities
}
```

### **Ollama Integration**
```go
// internal/extractor/ollama.go
type OllamaExtractor struct {
    client *ollama.Client
    model  string
}

func (o *OllamaExtractor) Extract(markdown string) []Entity {
    prompt := fmt.Sprintf(`
Extract entities from this devlog session.
Focus on: components, services, config, files, technologies.

Devlog:
%s

Output JSON only:
`, markdown)
    
    response, err := o.client.Generate(context.Background(), &ollama.GenerateRequest{
        Model:  o.model,
        Prompt: prompt,
        Stream: false,
    })
    
    entities := parseJSON(response.Response)
    return entities
}
```

***

## 🎛️ Configuration

### **config.toml**
```toml
[entity_extraction]
enabled = true
primary_extractor = "ollama"
fallback_extractor = "regex"

[ollama]
model = "llama3.2:3b"
url = "http://localhost:11434"
timeout = "5s"

[logging]
log_extractions = true  # Track extractor performance
```

### **CLI Controls**
```bash
# Test extraction
cat session.md | bd extract-entities --dry-run

# Switch modes
bd config set entity_extraction.primary_extractor ollama

# View extraction log
bd extraction-log --last 10
bd extraction-stats  # Entities/day by extractor
```

***

## 🧪 Integration Points

### **1. Devlog Sync Hook**
```
bd devlog sync
→ Parse new sessions
→ Run extraction pipeline
→ Update entities table + relationships
→ Log performance metrics
```

### **2. Search Enhancement**
```
BM25 query expansion:
bd devlog search "proxy buffering"
→ Expands to: proxy_buffering, nginx, cloudron
→ Returns 3 sessions vs 0
```

### **3. Graph Auto-Links**
```
From extracted entities, infer relationships:
"nginx.conf" + "proxy_buffering" → configures relationship
"ManageColumnsModal" + "useSortable" → uses relationship
```

***

## 📈 Expected Results

### **Sample Devlog → Extracted Entities**

**Input (300 words):**
```
Fixed nginx timeout in ManageColumnsModal. The issue was proxy_buffering being on, causing SSE connections to drop after 60s. Updated nginx.conf with proxy_buffering off and proxy_read_timeout 300s. Fixed useSortable hook parameters in ManageColumnsModal.tsx. Cloudron deployment now stable at v4.1.174. MCP remote SSE transport working end-to-end.
```

**Regex output (current):**
```
["nginx", "modal"]
```

**Ollama + Regex output (proposed):**
```
[
  "nginx" (config, confidence: 1.2),
  "proxy_buffering" (nginx_setting, confidence: 1.1),
  "proxy_read_timeout" (nginx_setting, confidence: 1.0),
  "nginx.conf" (file, confidence: 1.0),
  "managecolumnsmodal" (component, confidence: 1.2),
  "usesortable" (hook, confidence: 1.1),
  "cloudron" (deployment, confidence: 1.2),
  "mcp" (service, confidence: 1.0),
  "sse" (transport, confidence: 0.9)
]
```

**Impact:**
```
bd devlog graph "nginx"
→ proxy_buffering → cloudron → mcp-sse-transport
→ Full debugging chain preserved
```

***

## 🚀 Implementation Plan

### **Phase 1: Regex Fallback (1 day)**
```
[ ] Extract existing regex patterns from codebase
[ ] Add to extraction pipeline
[ ] Test on 10 sample devlogs
[ ] Measure baseline coverage
```

### **Phase 2: Ollama Integration (2 days)**
```
[ ] Ollama Go client integration
[ ] Prompt engineering + JSON parsing
[ ] Confidence scoring + deduplication
[ ] Fallback chain logic
[ ] Integration with bd devlog sync
```

### **Phase 3: Production (1 day)**
```
[ ] Config.toml schema
[ ] CLI controls (bd extract-entities --dry-run)
[ ] Extraction logging/metrics
[ ] Search integration tests
```

**Total:** 4 days, zero cost (Ollama local).

***

## 📊 Success Criteria

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Entities/session** | 8-15 (vs 2-4) | `bd extraction-stats` |
| **Extraction time** | <1s | Extraction log |
| **Graph edges** | 2x increase | `SELECT COUNT(*) FROM entity_relationships` |
| **Search precision** | +30% | Manual A/B testing |
| **Ollama uptime** | 99% | Local monitoring |

***

## 🎯 Competitive Advantage

| Feature | **Current** | **Ollama + Regex** |
|---------|-------------|-------------------|
| **Entity coverage** | Components only | Components + config + relationships |
| **Search quality** | Keyword-only | Semantic + graph-enhanced |
| **Maintenance** | Manual regex | Self-improving (LLM) |
| **Cost** | $0 | $0 (local) |

**Unique:** **Devlog-native extraction** - understands debugging narratives, extracts relationships automatically.

***

## 📜 Dependencies

- ✅ Ollama installed (`curl -fsSL https://ollama.ai/install.sh | sh`)
- ✅ Model pulled (`ollama pull llama3.2:3b`)
- ✅ Existing regex patterns extracted from codebase
- ✅ Devlog sync pipeline identified

**Risks:** None (local, fallback-safe, configurable).

***

**Approved:** [ ]  
**Implemented:** [ ]  
**Docs updated:** [ ]
