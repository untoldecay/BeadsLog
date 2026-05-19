# BeadsLog Feature Plan: Tool Changelog & Project Catchup
**Date:** 2026-05-19
**Status:** Planned

## 1. Goal
Improve alignment between the tool, the agent, and the developer by providing:
- **Binary Changelog:** Communicates new CLI features and protocol changes during onboarding.
- **Project Catchup:** A mechanism to summarize team activity since the last time a user/agent checked.

## 2. Feature 1: Tool Changelog (Binary)
- **Source:** Embedded Go struct in `internal/changelog/changelog.go`.
- **Content:** Version ID, Date, New Features, Protocol Updates.
- **State:** `last_seen_changelog_version` stored in `.beads/config.yaml`.
- **Trigger:** `bd onboard` and `bd ready` will compare the binary version with the config and display a "What's New" block if outdated.

## 3. Feature 2: `bd catchup` (Project Activity)
- **Logic:** Identifies `Sessions` and `Closed Issues` where `timestamp > last_catchup_time`.
- **Retention:** `last_catchup_time` stored in the SQLite `metadata` table (enables cross-machine sync).
- **Output:** Raw activity feed (titles, authors, reasons for abandonment).
- **Agent Prompt:** `_rules/_devlog/_generate-catchup.md`. After the raw output, the agent is instructed to use this prompt to generate a high-signal human-readable summary.

## 4. The Catchup Prompt (`_generate-catchup.md`)
Adapted from the exhaustive summarization prompt:
- **Sections:**
    - **Architectural Shifts:** Changes to core components or graph density.
    - **Key Features & Enhancements:** Landed work.
    - **Blocked & Abandoned Paths:** Critical "Tombstone" awareness.
    - **Next Steps:** Open issues from `bd ready`.

## 5. Implementation Roadmap

### Phase 1: Tool Changelog
- Create `internal/changelog/` with embedded data.
- Update `config.yaml` parser to support `last_seen_changelog_version`.
- Update `bd onboard` / `bd ready` UI to inject changelog block.

### Phase 2: `bd catchup` Core
- Add `last_catchup_time` to metadata table (migration).
- Implement `bd catchup` command to fetch raw delta.
- Implement `bd catchup --ack` to update the timestamp.

### Phase 3: Catchup Prompt & Guardrails
- Create `_rules/_devlog/_generate-catchup.md`.
- Update `bd ready` to suggest `bd catchup` if new activity is detected.
