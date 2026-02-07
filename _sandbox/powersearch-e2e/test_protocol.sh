#!/bin/bash
set -e

# Protocol Update Verification Test
# Focus: Verify Vercel-style "Always-On" protocol injection.

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 Starting Protocol Verification Test...${NC}"

# 1. Setup isolated environment
TEST_DIR="_sandbox/protocol-test/repo"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Path to the freshly built binary
BD="../../../bd"

echo -e "${BLUE}1. Initializing BeadsLog (Manual Mode)...${NC}"
$BD init --prefix test --quiet

# 2. Check Initial Trap
echo -e "${BLUE}2. Verifying Initial Trap in AGENT.md...${NC}"
if [ ! -f "AGENT.md" ]; then
    # bd init might not create AGENT.md if not interactive/selected?
    # Actually init calls configureAgentRules with Candidates.
    # But in non-interactive mode without args, it might default to Candidates.
    # Let's check.
    echo -e "${RED}❌ FAILURE: AGENT.md not created.${NC}"
    # Wait, init logic: "If no candidates specified... assumes forceEnable=true if called from main setup wizard"
    # But quiet mode might skip prompting.
    # Let's check if we need to manually create it or if init does it.
    # Init code: "result.AgentRules = configureAgentRules(quiet, true, targetAgents)"
    # targetAgents comes from "Candidates" if non-interactive default.
    # So it should be there.
    exit 1
fi

if grep -q "BEFORE ANYTHING ELSE: run 'bd onboard'" AGENT.md; then
    echo -e "${GREEN}✅ SUCCESS: Initial trap present.${NC}"
else
    echo -e "${RED}❌ FAILURE: Initial trap missing.${NC}"
    cat AGENT.md
    exit 1
fi

# 3. Run Onboard and Ready
echo -e "${BLUE}3. Running 'bd onboard' and 'bd ready' to unlock protocol...${NC}"
$BD onboard > /dev/null
$BD ready > /dev/null

# 4. Verify Always-On Protocol
echo -e "${BLUE}4. Verifying 'Always-On Protocol' injection...${NC}"

# Check for key phrases from the new protocol
if grep -q "BEADSLOG AGENTS.MD - ALWAYS-ON PROTOCOL" AGENT.md; then
    echo -e "${GREEN}✅ SUCCESS: Header found.${NC}"
else
    echo -e "${RED}❌ FAILURE: Header missing.${NC}"
    exit 1
fi

if grep -q "CORE REASONING LOOP - FOLLOW IN ORDER EVERY TASK" AGENT.md; then
    echo -e "${GREEN}✅ SUCCESS: Core loop found.${NC}"
else
    echo -e "${RED}❌ FAILURE: Core loop missing.${NC}"
    exit 1
fi

if grep -q "Load PROTOCOL.md" AGENT.md; then
    echo -e "${RED}❌ FAILURE: Found legacy instruction 'Load PROTOCOL.md'.${NC}"
    exit 1
else
    echo -e "${GREEN}✅ SUCCESS: Legacy conditional loading removed.${NC}"
fi

# 5. Verify Reference Files
echo -e "${BLUE}5. Verifying Reference Files...${NC}"
if [ -f "_rules/_orchestration/PROTOCOL.md" ]; then
    if grep -q "Protocol: Detailed Reference" "_rules/_orchestration/PROTOCOL.md"; then
        echo -e "${GREEN}✅ SUCCESS: PROTOCOL.md updated to reference mode.${NC}"
    else
        echo -e "${RED}❌ FAILURE: PROTOCOL.md content incorrect.${NC}"
        exit 1
    fi
else
    echo -e "${RED}❌ FAILURE: PROTOCOL.md missing.${NC}"
    exit 1
fi

echo -e "${GREEN}✨ PROTOCOL UPDATE VERIFIED! ✨${NC}"
