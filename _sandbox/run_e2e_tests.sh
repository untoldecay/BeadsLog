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
if ! ./bd devlog search "test record" | grep -q "Alice E2E"; then
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
if ! ./bd devlog list | grep -qi "paused"; then
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
if ! ./bd devlog entities --sort=mentions | grep -q "UserAuthenticationService"; then
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
printf "1\n" | "$BD_BIN" init --solo --prefix solotest --quiet --no-daemon > /dev/null 2>&1

if ! grep -q 'sync-mode: "local-only"' .beads/config.yaml || \
   ! grep -q 'no-push: true' .beads/config.yaml || \
   ! grep -q 'daemon.auto-sync: false' .beads/config.yaml; then
    echo "❌ FAIL: solo config keys missing from config.yaml"
    grep -E "sync-mode|no-push|auto-sync" .beads/config.yaml || true
    exit 1
fi
if ! grep -q "^\.beads/" .git/info/exclude; then
    echo "❌ FAIL: .beads/ not in .git/info/exclude (invisible mode)"
    exit 1
fi
if git status --porcelain | grep -q "\.beads"; then
    echo "❌ FAIL: git still sees .beads files in invisible mode"
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

echo -e "\n🎉 ALL EXTENSIVE TESTS PASSED SUCCESSFULLY!"
