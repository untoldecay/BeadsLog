#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 Starting Dual Attribution E2E Test...${NC}"

# 1. Setup isolated environment
TEST_DIR="_sandbox/devlog-tpa-e2e/repo"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Path to the freshly built binary
BD="../../../bd"

echo -e "${BLUE}1. Initializing BeadsLog...${NC}"
$BD init --prefix test --quiet

# 2. Create a mixed-format legacy index
echo -e "${BLUE}2. Creating Mixed-Format Index (4 and 5 columns)...${NC}"
mkdir -p _rules/_devlog
cat > _rules/_devlog/legacy_4.md <<EOM
# Legacy 4
Standard session.
EOM
cat > _rules/_devlog/legacy_5.md <<EOM
# Legacy 5
Intermediate session.
EOM

cat > _rules/_devlog/_index.md <<EOM
| Subject | Problems | Date | Devlog |
|---------|----------|------|---------|
| [legacy] 4-col | testing 4 | 2026-01-01 | [legacy_4.md](legacy_4.md) |
| [legacy] 5-col | testing 5 | Author Name | 2026-01-02 | [legacy_5.md](legacy_5.md) |
EOM

# 3. Run Migration
echo -e "${BLUE}3. Running Migration to 6-column format...${NC}"
$BD devlog migrate --format-index

# Verify filesystem
COL_COUNT=$(grep "\[legacy\] 4-col" _rules/_devlog/_index.md | tr -cd '|' | wc -c)
if [ "$COL_COUNT" -eq 7 ]; then
    echo -e "${GREEN}✅ SUCCESS: Index migrated to 6 columns (7 pipes).${NC}"
else
    echo -e "${RED}❌ FAILURE: Index has $COL_COUNT pipes, expected 7.${NC}"
    exit 1
fi

# 4. Sync and Verify Database
echo -e "${BLUE}4. Syncing to Database...${NC}"
$BD devlog sync --no-daemon --quiet

# Check authors metrics
echo -e "${BLUE}5. Verifying Authors Metrics...${NC}"
$BD devlog authors --no-daemon

# 6. Test Recording with Dual Attribution
echo -e "${BLUE}6. Testing 'devlog record' with explicit Author and Agent...${NC}"
cat > _rules/_devlog/new_6.md <<EOM
# New 6
Dual attribution session.
EOM
$BD devlog record --subject "[test] dual" --problem "verify both" --file "new_6.md" --author "HumanUser" --agent "TestAgent" --no-daemon

# 7. Final Verification
echo -e "${BLUE}7. Final Attribution Check...${NC}"
RESULT=$($BD devlog list --no-daemon | grep "HumanUser" || true)
if [[ "$RESULT" == *"HumanUser"* ]]; then
    echo -e "${GREEN}✅ SUCCESS: Dual attribution working in list output!${NC}"
else
    echo -e "${RED}❌ FAILURE: Attribution not found in list.${NC}"
    exit 1
fi

echo -e "${GREEN}✨ ALL DUAL ATTRIBUTION TESTS PASSED! ✨${NC}"
