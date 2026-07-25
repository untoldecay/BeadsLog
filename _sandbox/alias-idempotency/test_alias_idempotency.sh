#!/usr/bin/env bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 Alias Writer Idempotency E2E${NC}"

TEST_DIR="_sandbox/alias-idempotency/repo"
BD="$PWD/bd"

if [[ ! -x "$BD" ]]; then
  echo -e "${RED}bd binary not found at $BD — run 'go build -o bd ./cmd/bd' first${NC}" >&2
  exit 1
fi

rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

git init -q
git config user.email test@example.com
git config user.name  Tester

echo -e "${BLUE}1. bd init${NC}"
"$BD" init --prefix ali --quiet

echo -e "${BLUE}2. Sync devlog to materialize seed entities${NC}"
"$BD" devlog sync --quiet 2>/dev/null || true
"$BD" devlog verify --fix 2>&1 | tail -3 || true

echo -e "${BLUE}   Entities present:${NC}"
"$BD" devlog entities 2>&1 | grep -E "^│" | head -5 || true

echo -e "${BLUE}3. Alias 'repo' under 'BeadsLog' (both are extant post-init entities)${NC}"
"$BD" devlog alias BeadsLog repo 2>&1 | tail -5 || true

echo -e "${BLUE}4. First sync — this should create .beads/aliases.jsonl${NC}"
"$BD" sync 2>&1 | tail -3 || true

if [[ ! -f .beads/aliases.jsonl ]]; then
  echo -e "${RED}FAIL: .beads/aliases.jsonl was not created — cannot test idempotency${NC}" >&2
  echo -e "${BLUE}DB alias count:${NC}"
  "$BD" devlog entities 2>&1 | head -10 || true
  exit 2
fi

HASH1=$(shasum -a 256 .beads/aliases.jsonl | awk '{print $1}')
MTIME1=$(stat -f %m .beads/aliases.jsonl)
echo -e "  hash after sync 1: $HASH1"
echo -e "  mtime after sync 1: $MTIME1"

git add -A
git commit -q -m "baseline: alias seed + first sync"

echo -e "${BLUE}5. Second sync — should be a no-op (hash gate must skip rewrite)${NC}"
sleep 1
"$BD" sync 2>&1 | tail -3 || true

HASH2=$(shasum -a 256 .beads/aliases.jsonl | awk '{print $1}')
MTIME2=$(stat -f %m .beads/aliases.jsonl)
echo -e "  hash after sync 2: $HASH2"
echo -e "  mtime after sync 2: $MTIME2"

PORCELAIN=$(git status --porcelain -- .beads/aliases.jsonl 2>&1 || true)

echo ""
echo -e "${BLUE}=== RESULT ===${NC}"
FAIL=0
if [[ "$HASH1" != "$HASH2" ]]; then
  echo -e "${RED}FAIL: aliases.jsonl content hash differs between syncs${NC}"
  FAIL=1
fi
if [[ "$MTIME1" != "$MTIME2" ]]; then
  echo -e "${RED}FAIL: aliases.jsonl mtime changed on second sync (hash gate did not fire)${NC}"
  FAIL=1
fi
if [[ -n "$PORCELAIN" ]]; then
  echo -e "${RED}FAIL: git status is not clean on .beads/aliases.jsonl:${NC}"
  echo "$PORCELAIN"
  FAIL=1
fi

if [[ $FAIL -eq 0 ]]; then
  echo -e "${GREEN}PASS: aliases.jsonl is fully idempotent across syncs${NC}"
  exit 0
fi
exit 1
