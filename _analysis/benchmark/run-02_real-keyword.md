# Run 02 — This repo (104 devlogs), keyword recall

- **Date:** 2026-08-23
- **Corpus:** BeadsLog's own `_rules/_devlog/` (104 files).
- **Model:** sonnet, both arms. Questions: keyword-oriented (exact terms present in devlogs).

## Questions
1. Why does solo mode use a separate `_rules/_devlog-solo` directory instead of the team's committed devlog dir?
2. What branch does `bd init` default to for beads commits when the repo has a git remote, and why?
3. How was the leak of a solo user's private issues fixed (already-tracked `.beads/` files)?
4. Why were relationship link suggestions noisy, and what fixed it?
5. How does solo mode keep per-machine config out of the committed `config.yaml`?

## Result — both 5/5 correct

| Arm | Correct | Tokens | Tool-calls |
|---|---|---|---|
| A — bd | 5/5 | 31,158 | 13 |
| B — grep | 5/5 | **23,839** | 11 |

## Conclusion
**Brute-force still cheaper** (bd +31% tokens). Even at 104 files, a capable agent
greps for guessable keywords (`solo`, `beads-metadata`, `skip-worktree`) and jumps
to the right files without reading all 104. bd's default pretty-printed output is
heavier than terse grep lines. This run motivated three method fixes for Run 03:
compact `--json` bd output, fuzzy (non-keyword) questions, and a blind quality judge.
