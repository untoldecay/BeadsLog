#!/bin/bash
set -e

# Powersearch E2E Integration Test
# Focus: Background Enrichment, Crystallization, and Multi-model stability.

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 Starting Powersearch E2E Sandbox Test...${NC}"

# 1. Setup isolated environment
TEST_DIR="_sandbox/powersearch-e2e/repo"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Path to the freshly built binary
BD="../../../bd"

echo -e "${BLUE}1. Initializing BeadsLog (Manual Mode)...${NC}"
$BD init --prefix test --quiet

# 2. Configure for Test (Disable AI first to test fast sync)
echo -e "${BLUE}2. Disabling AI Enrichment...${NC}"
$BD config set entity_extraction.enabled false --quiet
$BD config set entity_extraction.background_enrichment false --quiet

# 3. Create a devlog session
echo -e "${BLUE}3. Creating devlog session...${NC}"
mkdir -p _rules/_devlog
cat > _rules/_devlog/2026-01-27_test.md <<EOF
# Test Session
The frontend calls the backend.
EOF

cat > _rules/_devlog/_index.md <<EOF
| Subject | Problems | Date | Devlog |
|---------|----------|------|---------|
| [test] sync | verification | 2026-01-27 | [2026-01-27_test.md](2026-01-27_test.md) |
EOF

# 4. Sync (Fast)
echo -e "${BLUE}4. Running Sync (Regex Only)...${NC}"
$BD devlog sync --quiet
echo -e "${GREEN}✓ Sync completed instantly.${NC}"

# 5. Enable AI and Background Enrichment
echo -e "${BLUE}5. Enabling Background AI (llama3.2:1b)...${NC}"
$BD config set entity_extraction.enabled true --quiet
$BD config set entity_extraction.background_enrichment true --quiet
$BD config set ollama.model llama3.2:1b --quiet

# 6. Verify Background Status (via RPC if daemon is used, but here we run foreground logic)
# In this sandbox, we'll use 'verify --fix-ai' to simulate the worker's logic foreground
echo -e "${BLUE}6. Running 'verify --fix-ai' to simulate worker crystallization...${NC}"
$BD devlog verify --fix-ai --quiet

# 7. Check if file was crystallized
if grep -q "### Architectural Relationships" _rules/_devlog/2026-01-27_test.md; then
    echo -e "${GREEN}✅ SUCCESS: File was crystallized with architectural relationships!${NC}"
else
    echo -e "${RED}❌ FAILURE: File was NOT crystallized.${NC}"
    exit 1
fi

# 8. Test Onboarding Auto-Correction
echo -e "${BLUE}8. Testing Onboarding Prompt Auto-Correction...${NC}"
# Create a manual-mode prompt
echo "Manual Prompt Header" > _rules/_devlog/_generate-devlog.md
# Run onboard
$BD onboard > /dev/null
# Check if it was updated to AI-Enhanced (since background_enrichment is true)
if grep -q "(AI Enhanced)" _rules/_devlog/_generate-devlog.md; then
    echo -e "${GREEN}✅ SUCCESS: Onboarding updated prompt to match AI config.${NC}"
else
    echo -e "${RED}❌ FAILURE: Onboarding failed to update prompt.${NC}"
    exit 1
fi

echo -e "${GREEN}✨ ALL POWERSEARCH TESTS PASSED! ✨${NC}"
