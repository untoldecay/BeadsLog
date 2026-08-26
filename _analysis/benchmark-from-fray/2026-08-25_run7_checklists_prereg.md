# Run 7 — pre-registered coverage checklists (committed BEFORE any answer exists)

Two planning tasks × 5-arm access spectrum (bd + grep full/no-devlog/terse/code-only),
judged on these fixed 10-item checklists (present / partial / missed) + rank.

## Task B — MODIFY touchy: pill → "expression surface"
Task: *drop images on the pill (does NOT open drawer) → enlarge into a mini-UI to add several
images + pre-categorize into spaces; pill morphs to show toast notifications (e.g. on web-extension
media capture). Report subsystems + gotchas, challenges, regressions to avoid, phased plan.*

- **B1** Pill/drawer TWO-WINDOW architecture (contract 34): separate always-interactive pill window + drawer window; pill state machine PILL→PRE_OPEN→OPEN→CLOSING.
- **B2** `GlobalMouseEventAddon` 50ms proximity polling + the macOS Cocoa-Y→screen-Y flip / cross-monitor coordinate bug (drawer-pill-v2 history).
- **B3** Drag-IN-to-open-drawer detection vs the NEW "drop image WITHOUT opening" — must distinguish the gestures.
- **B4** Existing drag stack: native drag-out (33/42), internal `application/fray-item`, drop-to-space (51) + media edge-handoff (52) + the ABANDONED 2-zones drag.
- **B5** Chrome-extension media capture (contract 36) = the event that fires a toast.
- **B6** ~~NO toast/notification pipe exists yet — net-new~~ **[PRE-REG ERROR, corrected post-run]** A toast store DOES exist: `src/stores/notificationStore.js` (`useNotificationStore`, `showNotification`) + `src/components/NotificationToast.vue`, mounted in `Bottombar.vue`/`TopBar.vue`, **drawer/main-window-scoped (Pinia)**. Contract 51's "no toast system yet" is STALE (3rd contract drift found in this bench). Correct answer = **reuse the store, bridge to the pill via IPC** (pill.js has no Pinia). Arms scored: FULL/bd/CODE = ✓ (read code, found it); TERSE/NODEVLOG = ✗ (trusted the stale contract). My own checklist made the same drift error I was testing for.
- **B7** Glass engine (contract 10) for the morphing/enlarged surface — forcePause multi-owner, scoped vs global capture.
- **B8** Pre-categorize into spaces = apply note-scope tags (contract 51 resolver) — reuse drop-to-space filing.
- **B9** Window morph = bounds change + click-through (`setIgnoreMouseEvents`), no `setBounds` during drag (contract 34).
- **B10** Regressions to avoid: breaking the drawer-trigger proximity, re-opening 2-zones fragility, glass pause ownership, multi-monitor coordinate correctness.

## Task A — NEW feature: global command palette / quick-switch
Task: *fuzzy-jump to any space, note, or action from a keyboard shortcut. Report subsystems + gotchas,
challenges, regressions to avoid, phased plan.*

- **A1** Existing shortcut system `useShortcutManager` (cmd+K/P/F, ESC) — **conflict surface** (cmd+K already toggles the main menu).
- **A2** `focusStack` layer model (contracts 20/22) — palette is a new layer; must push/pop cleanly and not swallow ESC.
- **A3** Existing drawer search (contract 28) — reuse for note/space fuzzy find, don't rebuild.
- **A4** Spaces/views store (`setActiveView`/`cycleView`) — palette jumps to a space.
- **A5** Notes store + `getFilteredNotes` — palette finds notes.
- **A6** Multi-window (contract 01) — palette in drawer vs floating window; which window owns it.
- **A7** The ESC unwind chain (contract 22) — palette must slot into ESC priority without breaking drawer close/blur (cf. the `mediaDragActive` defer we just added).
- **A8** Actions registry — are commands centralized or scattered across handlers? (affects "action" listing).
- **A9** Overlay/menu UI (contracts 10 glass, 21 menu positioning, 30 overlay transitions).
- **A10** Regressions to avoid: shortcut conflicts (cmd+K), focus-stack corruption, ESC/blur interaction, dev-vs-dist perf.
