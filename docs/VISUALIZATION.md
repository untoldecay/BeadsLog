# Search & Visualization

BeadsLog provides "X-ray vision" into your project's architecture using high-contrast terminal UI (Lipgloss).

## 🔍 Hybrid Search
The `search` command combines traditional text matching (BM25) with graph expansion. If you search for "auth," it doesn't just find the word; it finds the *sessions* related to entities linked to the Auth system.

```bash
bd devlog search "database"
```
**Output Example:**
```text
🔍 Search: "database"

╭────────────────────────────────────────────────╮
│      💡 Related Entities (Matched via FTS)     │
├────────────────────────────────────────────────┤
│ 1. PostgresConnector                           │
│ 2. SchemaCache                                 │
╰────────────────────────────────────────────────╯

╭────────────────────────────────────────────────╮
│           🔗 Graph Neighbors (Impact)          │
├────────────────────────────────────────────────┤
│ 1. UserManager (uses)                          │
│ 2. BillingService (writes_to)                  │
╰────────────────────────────────────────────────╯
```

## 🌳 Architectural Graph
The `graph` command visualizes the tree of dependencies. It's the primary tool for "Mapping" the project.

```bash
bd devlog graph "AuthAPI"
```
**Output Example:**
```text
╭───────────────────┬────────────────────────────╮
│    Root Entity    │ Graph Analysis             │
├───────────────────┼────────────────────────────┤
│                   │ AuthAPI (0)                │
│ AuthAPI           │ ├── Redis (1)              │
│                   │ │   └── SessionStore (2)   │
│                   │ └── OAuthProvider (1)      │
╰───────────────────┴────────────────────────────╯
```

## 💥 Impact Analysis
The `impact` command answers the question: **"What will I break if I change this?"** It shows everything that points *to* your target entity.

```bash
bd devlog impact "SessionManager"
```
**Output Example:**
```text
╭────────────────────────────────────────────────╮
│        💥 Impact Analysis: [SessionManager]    │
├────────────────────────────────────────────────┤
│ - LoginHandler (uses)                          │
│ - LogoutJob (executes)                         │
│ - AdminPortal (monitors)                       │
╰────────────────────────────────────────────────╯
```
