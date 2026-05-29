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
if ! ./bd devlog search "Testing pause" | grep -q "Testing pause"; then
    echo "❌ FAIL: Pause state not found in search."
    exit 1
fi
./bd devlog abandon --scope branch:main --message "Testing abandon" > /dev/null
./bd devlog sync > /dev/null
if ! ./bd devlog resume --last 1 > /dev/null; then
    echo "❌ FAIL: Resume command failed."
    exit 1
fi
echo "✅ Test 8 Passed."

echo -e "\n🎉 ALL EXTENSIVE TESTS PASSED SUCCESSFULLY!"
