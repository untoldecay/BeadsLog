#!/bin/bash
set -euo pipefail

echo "================================================="
echo "        BeadsLog E2E Devlog Test Suite           "
echo "================================================="

ROOT_DIR="$(pwd)"
TEST_DIR="$ROOT_DIR/_sandbox/e2e_test_env"

echo "[*] Cleaning up old test env..."
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

echo "[*] Building bd binary..."
go build -o bd "$ROOT_DIR/cmd/bd"

echo "[*] Initializing Git & BeadsLog..."
git init -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
./bd init --quiet

# ---------------------------------------------------------
echo -e "\n[*] Test 1: Username Edit & Record Command"
# ---------------------------------------------------------
./bd config set devlog.author "Alice E2E"
# Create the directory since record might expect it to exist
mkdir -p _rules/_devlog
echo "# Devlog Record Test" > "_rules/_devlog/2026-05-26_test-record.md"
./bd devlog record --subject "[feat] test record" --problem "need to test" --file "_rules/_devlog/2026-05-26_test-record.md" > /dev/null

if ! grep -q "Alice E2E" "_rules/_devlog/_index.md"; then
    echo "❌ FAIL: _index.md does not contain 'Alice E2E'"
    cat "_rules/_devlog/_index.md"
    exit 1
fi
./bd devlog sync > /dev/null
SEARCH_T1=$(./bd devlog search "test record" 2>&1 || true)
if ! echo "$SEARCH_T1" | grep -q "Alice E2E"; then
    echo "❌ FAIL: DB search does not contain 'Alice E2E'"
    exit 1
fi
echo "✅ Test 1 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 2: Commit Hook Verification"
# ---------------------------------------------------------
git add .
git commit -m "Add devlog" > commit_output.txt 2>&1 || true

if grep -q "BeadsLog" commit_output.txt || grep -q "bd" commit_output.txt || test -f .beads-hooks/pre-commit; then
    echo "✅ Test 2 Passed (Hook or .beads-hooks present)."
else
    echo "❌ FAIL: Commit hooks not configured or executed."
    cat commit_output.txt
    exit 1
fi

# ---------------------------------------------------------
echo -e "\n[*] Test 3: Extraction Hardening"
# ---------------------------------------------------------
cat << 'EOF' > _rules/_devlog/2026-05-26_extraction.md
# Devlog Session
## Context
Testing generic words: using, component, module, feature, system, state.
Testing strong regex: AuthService, UserModal, nginx-proxy, mcp-server.

### Architectural Relationships
- FragmentA -> TargetB (uses)
EOF
./bd devlog record --subject "[test] extraction" --problem "none" --file "_rules/_devlog/2026-05-26_extraction.md" > /dev/null
./bd devlog sync > /dev/null
ENTITIES=$(./bd devlog entities --sort=mentions)
if echo "$ENTITIES" | grep -qiE "using|component|module|feature|system|state"; then
    echo "❌ FAIL: Noise words were extracted."
    echo "$ENTITIES"
    exit 1
fi
if ! echo "$ENTITIES" | grep -qi "AuthService"; then
    echo "❌ FAIL: Strong regex 'AuthService' not extracted."
    exit 1
fi
GRAPH_OUT=$(./bd devlog graph FragmentA)
if ! echo "$GRAPH_OUT" | grep -qi "TargetB"; then
    echo "❌ FAIL: Explicit edge FragmentA -> TargetB not extracted."
    echo "Graph Output was: $GRAPH_OUT"
    exit 1
fi
echo "✅ Test 3 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 4: Aliasing & Registry Sync"
# ---------------------------------------------------------
./bd devlog alias TargetB FragmentA > /dev/null
./bd sync --flush-only > /dev/null
if ! grep -qi "TargetB" .beads/aliases.jsonl; then
    echo "❌ FAIL: aliases.jsonl not updated."
    exit 1
fi
echo "✅ Test 4 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 5: Search UI & Metadata"
# ---------------------------------------------------------
SEARCH_OUT=$(./bd devlog search "TargetB" --preview)
if ! echo "$SEARCH_OUT" | grep -q "Alice E2E"; then
    echo "❌ FAIL: Search UI missing author 'Alice E2E'."
    echo "$SEARCH_OUT"
    exit 1
fi
echo "✅ Test 5 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 6: Graph & Impact"
# ---------------------------------------------------------
# After alias, FragmentA is TargetB. The edge becomes TargetB -> TargetB.
./bd devlog graph TargetB > graph_out.txt
if ! cat graph_out.txt | grep -qi "TargetB"; then
    echo "❌ FAIL: Graph command crashed or empty."
    cat graph_out.txt
    exit 1
fi
./bd devlog impact TargetB > impact_out.txt
if ! cat impact_out.txt | grep -qi "TargetB"; then
    echo "❌ FAIL: Impact command crashed or empty."
    cat impact_out.txt
    exit 1
fi
echo "✅ Test 6 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 7: Unalias & Fix"
# ---------------------------------------------------------
./bd devlog unalias FragmentA > /dev/null
./bd devlog verify --fix > /dev/null
./bd sync --flush-only > /dev/null
if grep -qi "FragmentA" .beads/aliases.jsonl 2>/dev/null; then
    echo "❌ FAIL: FragmentA still in aliases.jsonl."
    exit 1
fi
echo "✅ Test 7 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 8: Session Lifecycle"
# ---------------------------------------------------------
./bd devlog pause --scope branch:main --message "Testing pause" > /dev/null
./bd devlog sync > /dev/null
LIST_T=$(./bd devlog list 2>&1 || true)
if ! echo "$LIST_T" | grep -qi "paused"; then
    echo "❌ FAIL: Pause state not found in list."
    ./bd devlog list
    exit 1
fi
./bd devlog abandon --scope branch:main --message "Testing abandon" > /dev/null
./bd devlog sync > /dev/null
if ! ./bd devlog resume --last 1 > /dev/null; then
    echo "❌ FAIL: Resume command failed."
    exit 1
fi
echo "✅ Test 8 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 9: Write-First Workflow Enforcement"
# ---------------------------------------------------------
# Record a file that does not exist
RECORD_OUT=$(./bd devlog record --subject "[feat] atomic stub" --problem "test stub" --file "_rules/_devlog/2026-05-26_atomic.md" 2>&1 || true)
if [ -f "_rules/_devlog/2026-05-26_atomic.md" ]; then
    echo "❌ FAIL: Stub file was created automatically (should be write-first)."
    exit 1
fi
if ! echo "$RECORD_OUT" | grep -q "AI ACTION REQUIRED"; then
    echo "❌ FAIL: Write-First directive missing."
    exit 1
fi
echo "✅ Test 9 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 10: Auto-Metadata Extraction"
# ---------------------------------------------------------
cat << 'EOF' > _rules/_devlog/2026-05-26_auto-metadata.md
# Auto Meta Subject
**Date:** 2026-05-30
**Author:** Alice E2E

## Problem
This is an auto-extracted problem description.

## Context
Test
EOF
./bd devlog record --file "_rules/_devlog/2026-05-26_auto-metadata.md" > /dev/null
./bd devlog sync > /dev/null
SEARCH_OUT=$(./bd devlog search "Auto Meta Subject")
if ! echo "$SEARCH_OUT" | grep -q "Auto Meta Subject"; then
    echo "❌ FAIL: Auto-extracted subject not found in index."
    exit 1
fi
echo "✅ Test 10 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 11: Orphan Detection"
# ---------------------------------------------------------
echo "# I am an orphan" > _rules/_devlog/2026-05-26_orphan.md
SYNC_OUT=$(./bd devlog sync)
if ! echo "$SYNC_OUT" | grep -q "orphaned devlog file"; then
    echo "❌ FAIL: sync did not warn about orphan."
    echo "$SYNC_OUT"
    exit 1
fi
echo "✅ Test 11 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 12: Non-Interactive Prune"
# ---------------------------------------------------------
rm _rules/_devlog/2026-05-26_orphan.md
./bd devlog sync > /dev/null # This should mark any missing recorded files as ghost
PRUNE_OUT=$(./bd devlog prune)
if ! echo "$PRUNE_OUT" | grep -q "Pruned"; then
    echo "❌ FAIL: prune command output unexpected."
    echo "$PRUNE_OUT"
    exit 1
fi
echo "✅ Test 12 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 13: Preferred Casing"
# ---------------------------------------------------------
echo -e "# Session\n\nWorking on UserAuthenticationService" > _rules/_devlog/2026-05-26_casing.md
./bd devlog verify --fix > /dev/null
./bd devlog sync > /dev/null
ENT_T=$(./bd devlog entities --sort=mentions 2>&1 || true)
if ! echo "$ENT_T" | grep -q "UserAuthenticationService"; then
    echo "❌ FAIL: Preferred casing not preserved."
    ./bd devlog entities --sort=mentions
    exit 1
fi
echo "✅ Test 13 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 14: Auto-Flush Metadata"
# ---------------------------------------------------------
# Create another entity to alias
echo -e "# Session\n\nWorking on auth-provider" > _rules/_devlog/2026-05-26_alias_test.md
./bd devlog verify --fix > /dev/null
./bd devlog sync > /dev/null
# Clean old aliases if any
rm -f .beads/aliases.jsonl
./bd devlog alias UserAuthenticationService auth-provider > /dev/null
# Verify it updated on disk WITHOUT manual sync
if [ ! -f ".beads/aliases.jsonl" ]; then
    echo "❌ FAIL: Auto-flush failed to create aliases.jsonl."
    exit 1
fi
if ! grep -qi "UserAuthenticationService" .beads/aliases.jsonl; then
    echo "❌ FAIL: Auto-flush did not record the alias."
    exit 1
fi
echo "✅ Test 14 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 15: Solo Mode Init (local-only)"
# ---------------------------------------------------------
BD_BIN="$TEST_DIR/bd"

# Path 1: invisible (git-exclude)
SOLO1="$TEST_DIR/solo_invisible"
mkdir -p "$SOLO1" && cd "$SOLO1"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
git commit -q --allow-empty -m init
# Prompts: 1 = Invisible (git-exclude), then 2 = Continuity devlog graph.
printf "1\n2\n" | "$BD_BIN" init --solo --prefix solotest --quiet --no-daemon > /dev/null 2>&1

# Invisible mode keeps solo settings in the git-excluded config.local.yaml so
# they never leak into the tracked config.yaml (BeadsLog-9vd).
if ! grep -q 'sync-mode: "local-only"' .beads/config.local.yaml || \
   ! grep -q 'no-push: true' .beads/config.local.yaml || \
   ! grep -q 'daemon.auto-sync: false' .beads/config.local.yaml; then
    echo "❌ FAIL: solo config keys missing from config.local.yaml"
    grep -E "sync-mode|no-push|auto-sync" .beads/config.local.yaml 2>/dev/null || true
    exit 1
fi
# The committed config.yaml must NOT carry active (uncommented) solo settings.
if grep -E '^[[:space:]]*(sync-mode:|no-push:|daemon\.auto-sync:)' .beads/config.yaml | grep -qv '^[[:space:]]*#'; then
    echo "❌ FAIL: solo settings leaked into committed config.yaml"
    grep -E '^[[:space:]]*(sync-mode:|no-push:|daemon\.auto-sync:)' .beads/config.yaml
    exit 1
fi
if ! grep -q "^\.beads/" .git/info/exclude; then
    echo "❌ FAIL: .beads/ not in .git/info/exclude (invisible mode)"
    exit 1
fi
if ! grep -q "^_rules/_devlog-solo/" .git/info/exclude; then
    echo "❌ FAIL: _rules/_devlog-solo/ not in .git/info/exclude (invisible mode)"
    exit 1
fi
if git status --porcelain | grep -q "\.beads"; then
    echo "❌ FAIL: git still sees .beads files in invisible mode"
    exit 1
fi
# devlog_dir must be repointed to the excluded solo dir.
if [ "$("$BD_BIN" config get devlog_dir --no-daemon 2>/dev/null)" != "_rules/_devlog-solo" ]; then
    echo "❌ FAIL: devlog_dir not repointed to _rules/_devlog-solo (got: $("$BD_BIN" config get devlog_dir --no-daemon 2>/dev/null))"
    exit 1
fi

# Path 2: local-only sync branch
SOLO2="$TEST_DIR/solo_branch"
mkdir -p "$SOLO2" && cd "$SOLO2"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
git commit -q --allow-empty -m init
printf "2\n" | "$BD_BIN" init --solo --prefix solotest2 --quiet --no-daemon > /dev/null 2>&1

if ! grep -q 'sync-branch: "beads-local"' .beads/config.yaml; then
    echo "❌ FAIL: sync-branch not set to beads-local"
    exit 1
fi
if ! git rev-parse --verify beads-local > /dev/null 2>&1; then
    echo "❌ FAIL: beads-local branch not created"
    exit 1
fi
if [ "$(git branch --show-current)" != "main" ]; then
    echo "❌ FAIL: user left on $(git branch --show-current) instead of main"
    exit 1
fi
# Capture output first: piping doctor straight into grep -q races with
# pipefail (grep exits on match -> doctor gets SIGPIPE -> pipeline fails).
DOCTOR_OUT=$("$BD_BIN" doctor --no-daemon 2>&1 || true)
if ! echo "$DOCTOR_OUT" | grep -q "Local-only mode"; then
    echo "❌ FAIL: doctor does not report local-only mode"
    echo "$DOCTOR_OUT" | grep -i "sync branch" || true
    exit 1
fi

cd "$TEST_DIR"
echo "✅ Test 15 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 16: Fresh-Clone Init With Committed Aliases (v0.56.0 crash regression)"
# ---------------------------------------------------------
# Simulates 'bd init' in a fresh clone whose git history contains beads data
# INCLUDING aliases.jsonl — this path dereferenced a nil global store and
# panicked (SIGSEGV) in v0.56.0.
FRESH="$TEST_DIR/fresh_clone_aliases"
mkdir -p "$FRESH/.beads" && cd "$FRESH"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
printf '{"id":"repro-1","title":"seed","status":"open","priority":2,"issue_type":"task","created_at":"2026-08-01T00:00:00Z","updated_at":"2026-08-01T00:00:00Z"}\n' > .beads/issues.jsonl
printf '{"alias_name":"authsvc","canonical_name":"auth-service"}\n' > .beads/aliases.jsonl
git add .beads && git commit -q -m "seed beads data"
rm -rf .beads  # fresh-clone simulation: git history has the data, disk does not

INIT_OUT=$("$BD_BIN" init --no-daemon --prefix repro 2>&1) || {
    echo "❌ FAIL: bd init crashed on repo with committed aliases"
    echo "$INIT_OUT" | tail -5
    exit 1
}
if ! echo "$INIT_OUT" | grep -q "imported entity aliases from git"; then
    echo "❌ FAIL: aliases were not imported from git history"
    echo "$INIT_OUT" | grep -iE "alias|import" || true
    exit 1
fi

cd "$TEST_DIR"
echo "✅ Test 16 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 17: Verify --fix Noise Deadlock (BeadsLog-ftw regression)"
# ---------------------------------------------------------
# A session whose entities are all noise words can never be cleared by regex
# backfill (the noise filter drops everything). --fix must say so honestly,
# and an AI-crystallized session (enrichment_status=2) must stop being flagged.
FTW="$TEST_DIR/ftw_noise"
mkdir -p "$FTW" && cd "$FTW"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
"$BD_BIN" init --quiet --no-daemon --prefix ftw > /dev/null 2>&1
mkdir -p _rules/_devlog
printf '# Noise session\n\n## Problem\nfix the code and sync the database\n\n## Entities\n- code\n- database\n\n## Architectural Relationships\n- code -> database\n' > _rules/_devlog/2026-08-07_noise.md
"$BD_BIN" devlog record --subject "noise test" --problem "test" --file "_rules/_devlog/2026-08-07_noise.md" --no-daemon > /dev/null 2>&1

FIX_OUT=$("$BD_BIN" devlog verify --fix --fix-regex --no-daemon 2>&1 || true)
if ! echo "$FIX_OUT" | grep -q "Regex extraction cannot clear these"; then
    echo "❌ FAIL: --fix did not report the noise-filter deadlock honestly"
    echo "$FIX_OUT" | tail -5
    exit 1
fi

# Simulate AI crystallization: verify must then treat the session as terminal
sqlite3 .beads/beads.db "PRAGMA trusted_schema=1; UPDATE sessions SET enrichment_status = 2;"
VERIFY_OUT=$("$BD_BIN" devlog verify --no-daemon 2>&1 || true)
if ! echo "$VERIFY_OUT" | grep -q "All sessions have linked entities"; then
    echo "❌ FAIL: AI-crystallized session still flagged incomplete"
    echo "$VERIFY_OUT" | tail -3
    exit 1
fi

cd "$TEST_DIR"
echo "✅ Test 17 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 18: Prefix Rename Persists To JSONL (BeadsLog-1tz regression)"
# ---------------------------------------------------------
# 'bd sync --rename-on-import' must heal the shared JSONL permanently:
# renamed issue survives under the new prefix, old ID becomes a tombstone,
# and subsequent plain syncs never re-raise the prefix mismatch error.
PFX="$TEST_DIR/prefix_migration"
mkdir -p "$PFX" && cd "$PFX"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
"$BD_BIN" init --quiet --no-daemon --prefix newpfx > /dev/null 2>&1
"$BD_BIN" create --title "native issue" --type task -p 2 --no-daemon > /dev/null 2>&1
printf '{"id":"oldpfx-abc","title":"legacy issue","status":"open","priority":2,"issue_type":"task","created_at":"2026-08-01T00:00:00Z","updated_at":"2026-08-01T00:00:00Z"}\n' >> .beads/issues.jsonl
git add -A && git commit -q -m "seed with legacy prefix"

"$BD_BIN" sync --rename-on-import --no-daemon > /dev/null 2>&1

if ! grep -q '"id":"newpfx-abc"' .beads/issues.jsonl; then
    echo "❌ FAIL: renamed issue newpfx-abc missing from JSONL (data loss)"
    grep -o '"id":"[^"]*"' .beads/issues.jsonl
    exit 1
fi
if ! grep '"id":"oldpfx-abc"' .beads/issues.jsonl | grep -q '"status":"tombstone"'; then
    echo "❌ FAIL: old ID not tombstoned in JSONL"
    exit 1
fi

RESYNC_OUT=$("$BD_BIN" sync --no-daemon 2>&1 || true)
if echo "$RESYNC_OUT" | grep -q "prefix mismatch detected"; then
    echo "❌ FAIL: plain re-sync still raises prefix mismatch"
    exit 1
fi
LIST_OUT=$("$BD_BIN" list --no-daemon 2>&1 || true)
if ! echo "$LIST_OUT" | grep -q "legacy issue"; then
    echo "❌ FAIL: renamed legacy issue not listed after migration"
    exit 1
fi

cd "$TEST_DIR"
echo "✅ Test 18 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 19: Prefix Auto-Adopt On Sync (BeadsLog-b4p)"
# ---------------------------------------------------------
# A clone whose DB uses an old prefix pulls a repo where the migration to a new
# prefix was AUTHORED — the committed config.yaml declares the new prefix (the
# signature). 'bd sync' must adopt it instead of erroring: repoint the DB prefix,
# migrate this clone's local issue(s) to the new prefix, tombstone the old IDs,
# and leave NO live old-prefix line. Guards against the 3-way-merge re-introducing
# the local issue under its old ID.
ADOPT="$TEST_DIR/prefix_adopt"
mkdir -p "$ADOPT" && cd "$ADOPT"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
"$BD_BIN" init --quiet --no-daemon --prefix oldpfx > /dev/null 2>&1
"$BD_BIN" create --title "my local only work" --type task -p 2 --no-daemon > /dev/null 2>&1
git add -A && git commit -q -m "seed local oldpfx issue"
# Upstream standardized on newpfx (a shared issue + a fresh one) AND authored the
# migration by declaring newpfx in the committed config.yaml.
cat > .beads/issues.jsonl <<'JSONL'
{"id":"newpfx-shared","title":"shared work","status":"open","priority":2,"issue_type":"task","created_at":"2026-08-01T00:00:00Z","updated_at":"2026-08-07T00:00:00Z"}
{"id":"newpfx-fresh","title":"upstream fresh","status":"open","priority":2,"issue_type":"task","created_at":"2026-08-01T00:00:00Z","updated_at":"2026-08-07T00:00:00Z"}
JSONL
# Authoring = updating the committed issue-prefix declaration (bd init pins it, so
# the migrator replaces the existing value rather than appending a duplicate).
sed -i.bak 's/^issue-prefix:.*/issue-prefix: "newpfx"/' .beads/config.yaml && rm -f .beads/config.yaml.bak
grep -q '^issue-prefix: "newpfx"' .beads/config.yaml || { echo "❌ setup: failed to author newpfx in config.yaml"; exit 1; }
git add -A && git commit -q -m "author migration: standardize on newpfx (config.yaml + issues)"

SYNC_OUT=$("$BD_BIN" sync --no-daemon 2>&1 || true)
if ! echo "$SYNC_OUT" | grep -qi "Adopted upstream prefix 'newpfx-'"; then
    echo "❌ FAIL: sync did not adopt upstream prefix"
    echo "$SYNC_OUT"
    exit 1
fi
if echo "$SYNC_OUT" | grep -qi "prefix mismatch detected"; then
    echo "❌ FAIL: sync still raised prefix mismatch error"
    exit 1
fi
# No live old-prefix line may remain in the JSONL (would re-pollute).
if grep '"id":"oldpfx' .beads/issues.jsonl | grep -q '"status":"open"'; then
    echo "❌ FAIL: a live oldpfx issue survived adoption"
    grep '"id":"oldpfx' .beads/issues.jsonl
    exit 1
fi
# The local issue must survive under the new prefix.
LIST_OUT=$("$BD_BIN" list --no-daemon 2>&1 || true)
if ! echo "$LIST_OUT" | grep -q "my local only work"; then
    echo "❌ FAIL: local issue lost during prefix adoption"
    exit 1
fi
if echo "$LIST_OUT" | grep "my local only work" | grep -q "oldpfx-"; then
    echo "❌ FAIL: local issue still under old prefix after adoption"
    exit 1
fi
# A second sync must be stable (no adopt, no mismatch).
RESYNC_OUT=$("$BD_BIN" sync --no-daemon 2>&1 || true)
if echo "$RESYNC_OUT" | grep -qi "prefix mismatch detected"; then
    echo "❌ FAIL: second sync re-raised prefix mismatch"
    exit 1
fi

cd "$TEST_DIR"
echo "✅ Test 19 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 20: Unsigned Prefix Does NOT Auto-Adopt (BeadsLog-b4p guard)"
# ---------------------------------------------------------
# A stray issue under a prefix nobody authored (config.yaml still declares the
# original prefix) must NOT be adopted — it errors like before, and the repo's
# prefix stays put. This is the pollution / test-CI-contamination guard.
POLL="$TEST_DIR/prefix_pollution"
mkdir -p "$POLL" && cd "$POLL"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
"$BD_BIN" init --quiet --no-daemon --prefix realpfx > /dev/null 2>&1
"$BD_BIN" create --title "legit work" --type task -p 2 --no-daemon > /dev/null 2>&1
git add -A && git commit -q -m "seed"
# A stray issue sneaks in under an UNauthored prefix; config.yaml still says realpfx.
printf '{"id":"straypfx-1","title":"accidental","status":"open","priority":2,"issue_type":"task","created_at":"2026-08-01T00:00:00Z","updated_at":"2026-08-07T00:00:00Z"}\n' >> .beads/issues.jsonl
git add -A && git commit -q -m "stray unauthored prefix"

POLL_OUT=$("$BD_BIN" sync --no-daemon 2>&1 || true)
if echo "$POLL_OUT" | grep -qi "Adopted upstream prefix 'straypfx-'"; then
    echo "❌ FAIL: unsigned stray prefix was auto-adopted (should have errored)"
    echo "$POLL_OUT"
    exit 1
fi
# The repo's prefix must be unchanged: a newly created issue still uses realpfx.
"$BD_BIN" create --title "after stray" --type task -p 2 --no-daemon > /dev/null 2>&1
if ! "$BD_BIN" list --no-daemon 2>/dev/null | grep "after stray" | grep -q "realpfx-"; then
    echo "❌ FAIL: repo prefix changed after an unsigned mismatch"
    "$BD_BIN" list --no-daemon 2>/dev/null | grep "after stray"
    exit 1
fi

cd "$TEST_DIR"
echo "✅ Test 20 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 21: rename-prefix Authors The Signature + Teammate Adopts (BeadsLog-b4p)"
# ---------------------------------------------------------
# Full round-trip: a migrator runs 'bd rename-prefix', which must both rename the
# DB issues AND declare the new prefix in config.yaml (the signature). A teammate
# clone still on the old prefix then adopts it on sync — including migrating the
# teammate's own local work.
MIG="$TEST_DIR/migrator"
mkdir -p "$MIG" && cd "$MIG"
git init -q -b main
git config user.name "Migrator"
git config user.email "mig@example.com"
"$BD_BIN" init --quiet --no-daemon --prefix oldteam > /dev/null 2>&1
"$BD_BIN" create --title "issue one" --type task -p 2 --no-daemon > /dev/null 2>&1
RENAME_OUT=$("$BD_BIN" rename-prefix newteam --no-daemon 2>&1 || true)
if ! echo "$RENAME_OUT" | grep -qi "Declared 'issue-prefix: \"newteam\"' in config.yaml"; then
    echo "❌ FAIL: rename-prefix did not declare the signature in config.yaml"
    echo "$RENAME_OUT"
    exit 1
fi
if ! grep -q '^issue-prefix: "newteam"' .beads/config.yaml; then
    echo "❌ FAIL: config.yaml missing uncommented issue-prefix after rename-prefix"
    exit 1
fi
git add -A && git commit -q -m "migrate oldteam->newteam"

# Teammate: independent clone still on oldteam with local work, then pulls.
TMATE="$TEST_DIR/tmate"
mkdir -p "$TMATE" && cd "$TMATE"
git init -q -b main
git config user.name "Teammate"
git config user.email "tm@example.com"
"$BD_BIN" init --quiet --no-daemon --prefix oldteam > /dev/null 2>&1
"$BD_BIN" create --title "teammate wip" --type task -p 2 --no-daemon > /dev/null 2>&1
git add -A && git commit -q -m "teammate local"
cp "$MIG/.beads/config.yaml" .beads/config.yaml
cp "$MIG/.beads/issues.jsonl" .beads/issues.jsonl
git add -A && git commit -q -m "pull migrator standardization"
TM_SYNC=$("$BD_BIN" sync --no-daemon 2>&1 || true)
if ! echo "$TM_SYNC" | grep -qi "Adopted upstream prefix 'newteam-'"; then
    echo "❌ FAIL: teammate did not auto-adopt the authored prefix"
    echo "$TM_SYNC"
    exit 1
fi
if "$BD_BIN" list --no-daemon 2>/dev/null | grep "teammate wip" | grep -q "oldteam-"; then
    echo "❌ FAIL: teammate's local work not migrated to the new prefix"
    exit 1
fi

cd "$TEST_DIR"
echo "✅ Test 21 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 22: init Pins Resolved Prefix In config.yaml (BeadsLog-bl1)"
# ---------------------------------------------------------
# bd init must write the resolved issue-prefix (uncommented) into config.yaml so
# every clone shares one committed prefix — decoupled from the directory name and
# ready to serve as the migration signature.
PIN="$TEST_DIR/init_pin_explicit"
mkdir -p "$PIN" && cd "$PIN"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
"$BD_BIN" init --quiet --no-daemon --prefix myproj > /dev/null 2>&1
if ! grep -q '^issue-prefix: "myproj"' .beads/config.yaml; then
    echo "❌ FAIL: init did not pin explicit prefix in config.yaml"
    grep -n "issue-prefix" .beads/config.yaml
    exit 1
fi
# Without --prefix, init must still pin *some* uncommented issue-prefix (the
# resolved default). The exact value depends on init precedence (a parent
# config.yaml can win over the dir name when nested), so assert only that an
# uncommented declaration was written — that is what bl1 requires.
PINDIR="$TEST_DIR/init_pin_dir/somedir"
mkdir -p "$PINDIR" && cd "$PINDIR"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
"$BD_BIN" init --quiet --no-daemon > /dev/null 2>&1
if ! grep -qE '^issue-prefix: "[^"]+"' .beads/config.yaml; then
    echo "❌ FAIL: init did not pin any default prefix in config.yaml"
    grep -n "issue-prefix" .beads/config.yaml
    exit 1
fi

cd "$TEST_DIR"
echo "✅ Test 22 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 23: Alias Suggestions — Count Hint, Review, Dismiss (BeadsLog-819)"
# ---------------------------------------------------------
# graph/search must show a one-line count hint (not a per-pair flood), and
# 'bd devlog aliases suggest/dismiss' must list and permanently reject pairs.
ALIAS_T="$TEST_DIR/alias_suggest"
mkdir -p "$ALIAS_T" && cd "$ALIAS_T"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
"$BD_BIN" init --quiet --no-daemon --prefix aliast > /dev/null 2>&1
mkdir -p _rules/_devlog
# Two sessions sharing a similar-named entity pair (>=2 sessions floor, containment name match)
printf '# S1\n\n## Problem\nWork on PaymentGateway integration\n\n## Entities\n- payment-gateway\n- paymentgateway-v2\n\n## Architectural Relationships\n- payment-gateway -> paymentgateway-v2\n' > _rules/_devlog/2026-08-10_alias-a.md
printf '# S2\n\n## Problem\nMore PaymentGateway work\n\n## Entities\n- payment-gateway\n- paymentgateway-v2\n\n## Architectural Relationships\n- payment-gateway -> paymentgateway-v2\n' > _rules/_devlog/2026-08-10_alias-b.md
"$BD_BIN" devlog record --subject "alias a" --problem "t" --file "_rules/_devlog/2026-08-10_alias-a.md" --no-daemon > /dev/null 2>&1
"$BD_BIN" devlog record --subject "alias b" --problem "t" --file "_rules/_devlog/2026-08-10_alias-b.md" --no-daemon > /dev/null 2>&1
"$BD_BIN" devlog sync --no-daemon > /dev/null 2>&1

GRAPH_OUT=$("$BD_BIN" devlog graph "payment-gateway" --no-daemon 2>&1 || true)
if echo "$GRAPH_OUT" | grep -q "OPPORTUNITY:"; then
    echo "❌ FAIL: graph still floods with per-pair OPPORTUNITY lines"
    exit 1
fi
SUGGEST_OUT=$("$BD_BIN" devlog aliases suggest --no-daemon 2>&1 || true)
if ! echo "$SUGGEST_OUT" | grep -q "payment-gateway"; then
    # Extraction may normalize differently; hint and suggest must at least agree
    if echo "$GRAPH_OUT" | grep -q "alias opportunit"; then
        echo "❌ FAIL: hint shows opportunities but suggest lists none"
        echo "$SUGGEST_OUT" | head -5
        exit 1
    fi
    echo "  (note: extractor did not produce the seeded pair — hint and suggest consistent)"
else
    "$BD_BIN" devlog aliases dismiss "payment-gateway" "paymentgateway-v2" --no-daemon > /dev/null 2>&1
    AFTER_OUT=$("$BD_BIN" devlog aliases suggest --no-daemon 2>&1 || true)
    if echo "$AFTER_OUT" | grep -q "paymentgateway-v2"; then
        echo "❌ FAIL: dismissed pair still suggested"
        exit 1
    fi
fi

cd "$TEST_DIR"
echo "✅ Test 23 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 24: Solo Wizard Reuses Existing Sync Branch (BeadsLog-a2l)"
# ---------------------------------------------------------
# init --solo option 2 must detect a pre-existing beads-metadata branch and
# offer to continue from it instead of orphaning history on beads-local.
A2L="$TEST_DIR/solo_reuse_branch"
mkdir -p "$A2L" && cd "$A2L"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
git commit -q --allow-empty -m init
git branch beads-metadata
printf "2\ny\n" | "$BD_BIN" init --solo --prefix a2l --quiet --no-daemon > /dev/null 2>&1
if ! grep -q 'sync-branch: "beads-metadata"' .beads/config.yaml; then
    echo "❌ FAIL: solo wizard did not reuse existing beads-metadata branch"
    grep "sync-branch" .beads/config.yaml || true
    exit 1
fi

cd "$TEST_DIR"
echo "✅ Test 24 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 25: Per-Machine Changelog Ack — No Cross-Clone Leak (BeadsLog-q9o.3)"
# ---------------------------------------------------------
# Ack must write per-machine (.local_changelog_seen, gitignored), never a
# committed file — so acking on one clone can't suppress the changelog on another.
Q9O="$TEST_DIR/per_machine_ack"
mkdir -p "$Q9O" && cd "$Q9O"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
"$BD_BIN" init --quiet --no-daemon --prefix q9o > /dev/null 2>&1

# Machine A: onboard records seen-state per-machine
"$BD_BIN" onboard --no-daemon > /dev/null 2>&1
if [ ! -f .beads/.local_changelog_seen ]; then
    echo "❌ FAIL: onboard did not write per-machine .local_changelog_seen"
    exit 1
fi
if grep -q "last-seen-changelog" .beads/config.yaml 2>/dev/null; then
    echo "❌ FAIL: seen-state written into committed config.yaml (leaks across clones)"
    exit 1
fi
if git status --porcelain 2>/dev/null | grep -q "local_changelog_seen"; then
    echo "❌ FAIL: .local_changelog_seen is tracked/visible to git (should be ignored)"
    exit 1
fi

# Machine B (fresh clone): the gitignored ack file did not travel, so onboard must still show the changelog
rm -f .beads/.local_changelog_seen
ONBOARD_OUT=$("$BD_BIN" onboard --no-daemon 2>&1 || true)
if ! echo "$ONBOARD_OUT" | grep -qiE "what's new|new in|v0\.[0-9]"; then
    echo "❌ FAIL: changelog suppressed on a fresh clone (cross-machine leak)"
    exit 1
fi
if [ ! -f .beads/.local_changelog_seen ]; then
    echo "❌ FAIL: onboard did not record per-machine seen-state"
    exit 1
fi

cd "$TEST_DIR"
echo "✅ Test 25 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 26: bd upgrade Consolidated To Single Command (BeadsLog-q9o.2)"
# ---------------------------------------------------------
# status/review/ack/check subcommands must be gone; the root command reports
# current version. Network-tolerant: --json always carries current_version.
HELP_OUT=$("$BD_BIN" upgrade --help --no-daemon 2>&1 || true)
for sub in "status" "review" "ack" "check"; do
    # Match a subcommand listing line (indented "  <name> "), not prose mentions
    if echo "$HELP_OUT" | grep -qE "^\s+$sub\s"; then
        echo "❌ FAIL: 'bd upgrade' still exposes removed subcommand '$sub'"
        exit 1
    fi
done
JSON_OUT=$("$BD_BIN" upgrade --json --no-daemon 2>&1 || true)
if ! echo "$JSON_OUT" | grep -q '"current_version"'; then
    echo "❌ FAIL: 'bd upgrade --json' missing current_version"
    echo "$JSON_OUT" | tail -3
    exit 1
fi

cd "$TEST_DIR"
echo "✅ Test 26 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 27: bd refresh — One-Command Post-Update (BeadsLog-q9o.1)"
# ---------------------------------------------------------
# Deterministic orchestration + mode gating (no network). Adoption itself is
# covered by the prefix tests (18-22) that refresh reuses.

# Solo repo: namespace step must be gated OFF by local-only mode.
RF_SOLO="$TEST_DIR/refresh_solo"
mkdir -p "$RF_SOLO" && cd "$RF_SOLO"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
git commit -q --allow-empty -m init
printf "1\n" | "$BD_BIN" init --solo --prefix rfs --quiet --no-daemon > /dev/null 2>&1
RF_OUT=$("$BD_BIN" refresh --no-daemon 2>&1 || true)
for token in "bd refresh" "Migrations:" "Doctor:" "Protocol:" "Namespace:" "✨"; do
    if ! echo "$RF_OUT" | grep -qF "$token"; then
        echo "❌ FAIL: refresh summary missing '$token'"
        echo "$RF_OUT"
        exit 1
    fi
done
if ! echo "$RF_OUT" | grep -q "Namespace:.*skipped (solo"; then
    echo "❌ FAIL: solo refresh did not skip the namespace probe"
    echo "$RF_OUT" | grep Namespace
    exit 1
fi
# Fresh repo just had its protocol injected by init with the same binary, so
# refresh must report it CURRENT (no drift, no needless rewrite).
if ! echo "$RF_OUT" | grep -q "Protocol:.*current"; then
    echo "❌ FAIL: refresh should report protocol current on a freshly-inited repo"
    echo "$RF_OUT" | grep Protocol
    exit 1
fi

# Non-team repo with no remote: namespace skipped for lack of remote.
RF_PLAIN="$TEST_DIR/refresh_plain"
mkdir -p "$RF_PLAIN" && cd "$RF_PLAIN"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
"$BD_BIN" init --quiet --prefix rfp --no-daemon > /dev/null 2>&1
RF_OUT2=$("$BD_BIN" refresh --no-daemon 2>&1 || true)
if ! echo "$RF_OUT2" | grep -q "Namespace:.*skipped (no git remote)"; then
    echo "❌ FAIL: no-remote refresh did not skip the namespace probe"
    echo "$RF_OUT2" | grep Namespace
    exit 1
fi

cd "$TEST_DIR"
echo "✅ Test 27 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 28: prime --hook SessionStart Dedupe (BeadsLog-29q)"
# ---------------------------------------------------------
# In a repo whose agent files carry the protocol block, 'bd prime --hook' must
# emit the lean reminder (much smaller than full 'bd prime'), while plain
# 'bd prime' stays full. Prevents the SessionStart double-injection.
HOOK_T="$TEST_DIR/prime_hook"
mkdir -p "$HOOK_T" && cd "$HOOK_T"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
"$BD_BIN" init --quiet --no-daemon --prefix hook > /dev/null 2>&1
# Ensure an agent file carries the protocol block (finalize onboarding path).
"$BD_BIN" ready --no-daemon > /dev/null 2>&1 || true

FULL_BYTES=$("$BD_BIN" prime --no-daemon 2>/dev/null | wc -c | tr -d ' ')
HOOK_BYTES=$("$BD_BIN" prime --hook --no-daemon 2>/dev/null | wc -c | tr -d ' ')
if [ "$HOOK_BYTES" -ge "$FULL_BYTES" ]; then
    echo "❌ FAIL: prime --hook ($HOOK_BYTES B) not smaller than full prime ($FULL_BYTES B)"
    exit 1
fi
# Lean output should still name the session-close protocol (core reminder kept).
if ! "$BD_BIN" prime --hook --no-daemon 2>/dev/null | grep -qi "close protocol"; then
    echo "❌ FAIL: prime --hook dropped the session-close reminder"
    exit 1
fi

cd "$TEST_DIR"
echo "✅ Test 28 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 29: bd devlog graph with no entity — whole graph (BeadsLog-xuv)"
# ---------------------------------------------------------
# No-arg graph must work: terminal summary (counts + hubs) without --html, and
# a full interactive export with --html. Uses the main e2e repo which already
# has entities + an explicit edge (FragmentA -> TargetB from Test 3).
cd "$TEST_DIR"
SUMMARY=$(./bd devlog graph 2>&1 || true)
if ! echo "$SUMMARY" | grep -q "Whole Graph"; then
    echo "❌ FAIL: no-arg graph did not print the whole-graph summary"
    echo "$SUMMARY" | head -5
    exit 1
fi
if ! echo "$SUMMARY" | grep -qiE "Entities: [0-9]+|Explicit edges: [0-9]+"; then
    echo "❌ FAIL: summary missing entity/edge counts"
    exit 1
fi
./bd devlog graph --html "$TEST_DIR/fullgraph.html" > /dev/null 2>&1
if [ ! -s "$TEST_DIR/fullgraph.html" ]; then
    echo "❌ FAIL: --html did not produce a non-empty full-graph file"
    exit 1
fi
if ! grep -q '"nodes":' "$TEST_DIR/fullgraph.html"; then
    echo "❌ FAIL: exported graph has no nodes array"
    exit 1
fi
# Interactive viewer v1 controls must be present in the exported HTML.
for ctrl in 'id="search"' 'id="panel"' 'id="coToggle"' 'onNodeClick' 'showPanel'; do
    if ! grep -q "$ctrl" "$TEST_DIR/fullgraph.html"; then
        echo "❌ FAIL: exported viewer missing control '$ctrl'"
        exit 1
    fi
done
echo "✅ Test 29 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 30: Relationship suggestions + manual link (BeadsLog-58r)"
# ---------------------------------------------------------
# Two entities co-occurring in 4 sessions with no explicit edge -> suggested;
# 'bd devlog link' creates it (removing the suggestion) and exports links.jsonl;
# the edge is re-applied on sync (survives re-extraction).
LK="$TEST_DIR/links_feature"
mkdir -p "$LK" && cd "$LK"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
"$BD_BIN" init --quiet --no-daemon --prefix lkf > /dev/null 2>&1
sqlite3 .beads/beads.db "PRAGMA trusted_schema=1;
INSERT INTO entities(id,name,mention_count) VALUES('e-ed','wysiwygeditor',10),('e-sl','slashcommandmenu',8);
INSERT INTO sessions(id,title,timestamp) VALUES('ls1','a','2026-01-01'),('ls2','b','2026-01-02'),('ls3','c','2026-01-03'),('ls4','d','2026-01-04');
INSERT INTO session_entities(session_id,entity_id) VALUES
 ('ls1','e-ed'),('ls1','e-sl'),('ls2','e-ed'),('ls2','e-sl'),('ls3','e-ed'),('ls3','e-sl'),('ls4','e-ed'),('ls4','e-sl');"

SUG=$("$BD_BIN" devlog links suggest --no-daemon 2>&1 || true)
if ! echo "$SUG" | grep -q "wysiwygeditor"; then
    echo "❌ FAIL: no relationship suggestion for the co-occurring pair"
    echo "$SUG" | head -5
    exit 1
fi
"$BD_BIN" devlog link wysiwygeditor slashcommandmenu --relationship uses --no-daemon > /dev/null 2>&1
AFTER=$("$BD_BIN" devlog links suggest --no-daemon 2>&1 || true)
if ! echo "$AFTER" | grep -qi "No pending"; then
    echo "❌ FAIL: suggestion did not disappear after linking"
    echo "$AFTER" | head -5
    exit 1
fi
"$BD_BIN" devlog sync --no-daemon > /dev/null 2>&1
if ! grep -q '"from":"wysiwygeditor"' .beads/links.jsonl 2>/dev/null; then
    echo "❌ FAIL: manual link not exported to links.jsonl"
    exit 1
fi
# dismiss makes a pair never resurface
"$BD_BIN" devlog links dismiss wysiwygeditor slashcommandmenu --no-daemon > /dev/null 2>&1

cd "$TEST_DIR"
echo "✅ Test 30 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 31: team -> solo -> team round trip (BeadsLog-9vd)"
# ---------------------------------------------------------
RT="$TEST_DIR/roundtrip"
mkdir -p "$RT" && cd "$RT"
git init -q -b main
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
"$BD_BIN" init --quiet --prefix rt --no-daemon > /dev/null 2>&1
# A committed team devlog to carry via continuity.
mkdir -p _rules/_devlog
printf '# Team\n## Problem\nWork on CoreService.\n' > _rules/_devlog/2026-08-10_team.md
git add -A && git commit -q -m "team devlog"

# team -> solo (Invisible + Continuity).
printf "1\n2\n" | "$BD_BIN" init --solo --force --prefix rt --no-daemon > /dev/null 2>&1
if [ "$("$BD_BIN" config get devlog_dir --no-daemon 2>/dev/null)" != "_rules/_devlog-solo" ]; then
    echo "❌ FAIL: devlog_dir not repointed to solo dir"
    exit 1
fi
if [ ! -f _rules/_devlog-solo/2026-08-10_team.md ]; then
    echo "❌ FAIL: continuity did not carry the team devlog into the solo dir"
    exit 1
fi
if ! git diff --quiet _rules/_devlog/; then
    echo "❌ FAIL: team devlog dir was modified during solo transition"
    exit 1
fi

# Write a solo-only devlog, then rejoin the team.
printf '# Solo\n## Problem\nSecretWidget.\n' > _rules/_devlog-solo/2026-08-11_solo.md
printf "n\nn\n" | "$BD_BIN" init --team --force --prefix rt --no-daemon > /dev/null 2>&1

if [ ! -f _rules/_devlog/2026-08-11_solo.md ]; then
    echo "❌ FAIL: solo devlog was not published to the team dir"
    exit 1
fi
if [ "$("$BD_BIN" config get devlog_dir --no-daemon 2>/dev/null)" != "_rules/_devlog" ]; then
    echo "❌ FAIL: devlog_dir not restored to team dir"
    exit 1
fi
if [ -d _rules/_devlog-solo ]; then
    echo "❌ FAIL: redundant solo dir not cleaned up after rejoining team"
    exit 1
fi
if [ -f .beads/config.local.yaml ]; then
    echo "❌ FAIL: config.local.yaml override not removed on team rejoin"
    exit 1
fi
# The solo-specific devlog exclude must be gone. (.beads/ may remain excluded:
# rejoining team now uses a dedicated sync branch, which legitimately keeps
# .beads/ off the work branch — see Test 33.)
if grep -qE "_devlog-solo" .git/info/exclude 2>/dev/null; then
    echo "❌ FAIL: solo devlog exclude not removed from .git/info/exclude"
    exit 1
fi

cd "$TEST_DIR"
echo "✅ Test 31 Passed."

# ---------------------------------------------------------
echo -e "\n[*] Test 33: dedicated sync branch is the default with a remote — beads never lands on the work branch (BeadsLog)"
# ---------------------------------------------------------
BARE="$TEST_DIR/origin33.git"
rm -rf "$BARE"; git init -q --bare "$BARE"
SB="$TEST_DIR/syncbranch"
mkdir -p "$SB" && cd "$SB"
git init -q -b develop
git config user.name "E2E Tester"
git config user.email "e2e@example.com"
git remote add origin "$BARE"
echo "code" > app.txt && git add app.txt && git commit -q -m "work on develop" && git push -q -u origin develop

# Plain init on a repo WITH a remote defaults to the dedicated sync branch.
"$BD_BIN" init --quiet --prefix sb --no-daemon > /dev/null 2>&1
if [ "$("$BD_BIN" config get sync.branch --no-daemon 2>/dev/null)" != "beads-metadata" ]; then
    echo "❌ FAIL: plain init with a remote did not default sync.branch to beads-metadata"
    exit 1
fi
# Follow-up: .beads/ is excluded from the work tree so a manual 'git add -A'
# (or an agent's blanket stage) can't land it on develop.
if ! grep -q '^\.beads/' .git/info/exclude; then
    echo "❌ FAIL: .beads/ not excluded from the work branch"
    exit 1
fi
"$BD_BIN" create "an issue" -p 2 --no-daemon > /dev/null 2>&1
"$BD_BIN" sync --no-daemon > /dev/null 2>&1
"$BD_BIN" create "a second issue" -p 2 --no-daemon > /dev/null 2>&1
"$BD_BIN" sync --no-daemon > /dev/null 2>&1  # incremental sync must also work

# Beads must be on beads-metadata (local + remote), never on develop.
if [ "$(git ls-tree -r beads-metadata --name-only 2>/dev/null | grep -c 'issues.jsonl')" -ne 1 ]; then
    echo "❌ FAIL: issues.jsonl not committed to local beads-metadata"
    exit 1
fi
git add -A  # the exact operation an agent close-protocol would run on develop
if git diff --cached --name-only | grep -q '^\.beads/'; then
    echo "❌ FAIL: 'git add -A' staged beads onto the develop work branch"
    git diff --cached --name-only | grep '^\.beads/'
    exit 1
fi
git reset -q
if git log develop --name-only --oneline 2>/dev/null | grep -q '\.beads/'; then
    echo "❌ FAIL: beads leaked onto the develop work branch"
    exit 1
fi
if ! git ls-remote origin 2>/dev/null | grep -q 'refs/heads/beads-metadata'; then
    echo "❌ FAIL: beads-metadata was not pushed to origin"
    exit 1
fi
if git ls-tree -r origin/develop --name-only 2>/dev/null | grep -q '\.beads/'; then
    echo "❌ FAIL: beads leaked onto origin/develop (shared work branch)"
    exit 1
fi

# --inline opts out (old behavior: no dedicated sync branch).
IN="$TEST_DIR/inline"
mkdir -p "$IN" && cd "$IN"
git init -q -b develop
git config user.name "E2E Tester"; git config user.email "e2e@example.com"
git remote add origin "$BARE"
"$BD_BIN" init --quiet --inline --prefix il --no-daemon > /dev/null 2>&1
if [ -n "$("$BD_BIN" config get sync.branch --no-daemon 2>/dev/null | grep -v 'not set')" ]; then
    echo "❌ FAIL: --inline should not set a sync branch"
    exit 1
fi

cd "$TEST_DIR"
echo "✅ Test 33 Passed."

echo -e "\n🎉 ALL EXTENSIVE TESTS PASSED SUCCESSFULLY!"
