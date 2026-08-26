# bd bench — Run 6 (SYNTHESIS: feasibility report across the access spectrum)

**Date:** 2026-08-25
**Why:** the recall runs (1–5) were "why did X happen" questions on a redundantly-documented repo → quality wash. This run tests a **forward-looking synthesis** task ("feasibility + impact of true real-time co-op on spaces & notes") across the same access spectrum, judged on **coverage** (present/missed subsystems), not prose.

**5 report arms, same task:** BD (bd graph/impact/search + any files) · grep-FULL (code+contracts+devlogs) · grep-NODEVLOG · grep-TERSE (+contracts, subjects only) · grep-CODE-ONLY.

## Efficiency
| Arm | Tokens | Tool-uses | Wall |
|---|---|---|---|
| BD | 36,292 | **20** | 96s |
| grep-FULL | 47,580 | 10 | 77s |
| grep-NODEVLOG | 41,544 | 14 | 64s |
| grep-TERSE | 48,291 | 13 | 69s |
| grep-CODE-ONLY | 36,469 | 13 | 79s |

Note: on this synthesis task **bd used the MOST tool-calls (20)** — its recall-run efficiency edge did NOT hold; graph/impact + code exploration cost more round-trips.

## Coverage (blind judge, 10-item checklist)
| Arm | Coverage | Quality | Rank |
|---|---|---|---|
| grep-FULL | 8.5/10 | 5 | **1** |
| **BD** | 8.5/10 | 5 | **2 (tie-adjacent)** |
| grep-TERSE | 7.5/10 | 4 | 3 |
| grep-NODEVLOG | 7/10 | 4 | 4 |
| grep-CODE-ONLY | **6.5/10** | 4 | 5 |

- **grep-FULL #1, bd #2 (near-tie).** bd uniquely grounded **C6 (the abandoned iCloud-SQLite / WAL-corruption history)** and **C7 (editor is hand-rolled `contenteditable` + `createEditorHistory`, NOT Tiptap)** in exact files/branches — but under-named the `state:updated`/transient bus. grep-FULL had the broadest spec coverage + the sharpest top risk (contract 01 forbids re-rendering a focused contenteditable → remote CRDT patches would jump the cursor).
- **The decisive differentiator = C3 (contract 26 p2p spec) + C6 (abandoned WAL history).** Reports split cleanly: those that read the **design/decision record** (bd + the 3 contract-having grep tiers) vs the one that read **live code only** (code-only) — which consequently **mis-framed co-op as greenfield and omitted the live-writer risk that history already litigated.**

## The finding — memory pays off on SYNTHESIS, not recall
This is the clearest positive signal in the whole benchmark, and it's NOT judge noise:
- **Code-only was last (6.5) with a *consequential* failure**: it treats real-time co-op as greenfield because it never sees **contract 26** (an existing p2p sync design) or the **abandoned iCloud-WAL corruption** — "the two facts that most change the feasibility verdict." An engineer led by that report would re-open a corruption class the team already fought.
- **The memory layers (devlogs + contracts) demonstrably added the coverage that matters** for a real forward decision — preventing repeat-mistakes + surfacing existing designs. That is the tangible use-case, and it's a *decision-quality* win, not a phrasing win.
- **bd's recurring niche held**: it (with full-grep) was the one to surface the **abandoned work** — the parked/reverted history with no code trace, exactly where it won Run 4's Q1.

## Honest caveat
It was the **memory (contracts + devlogs), not bd's tooling**, that drove coverage — **grep-FULL tied/edged bd** reading the same corpus, and bd was the *least* efficient arm here. So the value is **"have the durable memory"** more than **"use bd's retrieval."** bd's own edge (efficiency, semantic recall) is a recall-time property; on broad synthesis, breadth-of-access dominated.

## Combined conclusion (runs 1–6)
- **Recall "why" questions:** answer quality is a wash (rationale redundant across code-comments/contracts/commits/devlogs; ±25% judge noise). bd's edge = efficiency + the occasional devlog-only fact. Curated memory can even go *stale* vs code (the 100ms→50ms drift).
- **Forward synthesis (feasibility):** the memory earns its keep — code-only mis-frames the task and misses litigated risks; having devlogs+contracts is a real decision-quality advantage; bd ties the best coverage and uniquely surfaces abandoned-path risk.
- **Net:** the durable-memory discipline pays off most where you're **deciding what to build next** (surfacing prior designs + abandoned experiments), less where you're **re-deriving why shipped code is the way it is** (which the code itself documents). bd's specific, non-noisy edges are **efficiency** and **recall of the things with no code home** (parked/abandoned work, measured outcomes).

## Daemon hygiene
Baseline `72371` only; zero strays across all six runs.
