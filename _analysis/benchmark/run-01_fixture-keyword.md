# Run 01 — Tasklet fixture, keyword recall

- **Date:** 2026-08-23
- **Corpus:** "Tasklet" fixture, 8 devlogs (a fictional CLI todo app). *Fixture later dropped as non-demonstrative.*
- **Model:** sonnet, both arms. Questions: keyword-oriented.

## Questions (answer key facts in parens)
1. Why did Tasklet switch storage from JSON files to SQLite? (concurrent writes / file locking-corruption / query)
2. What sync approach did the team abandon, and why? (CRDT; too heavy / tombstone garbage)
3. Which components depend on AuthService? (SyncEngine, NotificationService, CLIParser)
4. What fixed the duplicate reminder bug in NotificationService? (5-second debounce)
5. Is the offline-mode work shippable? What is its status? (PAUSED; needs SyncEngine conflict UI)

## Result — both 5/5 correct

| Arm | Correct | Tokens | Tool-calls |
|---|---|---|---|
| A — bd | 5/5 | 22,043 | 13 |
| B — grep | 5/5 | **17,288** | **2** |

## Conclusion
**Brute-force won** (bd +27% tokens, 13 calls vs 2). Over 8 tiny files, one `grep -r`
+ one read beats issuing a bd query per question. bd's retrieval overhead only pays
off once the corpus is too big to grep cheaply. Honest — confirms the harness isn't
rigged toward bd.
