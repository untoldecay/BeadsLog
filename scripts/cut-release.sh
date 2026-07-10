#!/bin/bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${BLUE}🚀 BeadsLog Release Orchestrator${NC}"
echo -e "${BLUE}=================================${NC}"

# 1. Get current version
CURRENT_VERSION=$(grep "Version = " "$ROOT_DIR/cmd/bd/version.go" | sed 's/.*"\(.*\)".*/\1/')
echo -e "Current Version: ${YELLOW}$CURRENT_VERSION${NC}"

# 2. Ask for new version
echo -e "\nWhat type of release is this?"
echo "1) Patch (e.g. 0.47.1 -> 0.47.2)"
echo "2) Minor (e.g. 0.47.1 -> 0.48.0)"
echo "3) Major (e.g. 0.47.1 -> 1.0.0)"
echo "4) Keep current ($CURRENT_VERSION)"
echo "5) Custom version"
read -p "Select [1-5]: " VERSION_CHOICE

case $VERSION_CHOICE in
    1) NEW_VERSION=$(echo $CURRENT_VERSION | awk -F. '{$NF = $NF + 1;} 1' OFS=.) ;;
    2) NEW_VERSION=$(echo $CURRENT_VERSION | awk -F. '{$2 = $2 + 1; $3 = 0;} 1' OFS=.) ;;
    3) NEW_VERSION=$(echo $CURRENT_VERSION | awk -F. '{$1 = $1 + 1; $2 = 0; $3 = 0;} 1' OFS=.) ;;
    4) NEW_VERSION=$CURRENT_VERSION ;;
    5) read -p "Enter custom version: " NEW_VERSION ;;
    *) echo "Invalid choice"; exit 1 ;;
esac

echo -e "Target Version: ${GREEN}$NEW_VERSION${NC}"
read -p "Confirm version bump? [y/N] " CONFIRM
if [[ ! $CONFIRM =~ ^[Yy]$ ]]; then
    exit 0
fi

# 3. Handle Release Notes
read -p "Generate/Update release notes? [y/N] " DO_NOTES
if [[ $DO_NOTES =~ ^[Yy]$ ]]; then
    PROMPT_FILE="$ROOT_DIR/_rules/_prompts/_draft-release.md"
    echo "Gathering activity data for AI..."
    
    # Get recent activity
    RAW_ACTIVITY=$(./bd catchup 2>/dev/null || echo "No recent bd activity found")
    GIT_LOG=$(git log --oneline -n 20)
    
    cat > "$PROMPT_FILE" <<EOF
# Prompt: Draft BeadsLog Release Notes (v$NEW_VERSION)

## Objective
Update both the project CHANGELOG.md and the binary tool changelog (internal/changelog/changelog.go) for version v$NEW_VERSION.

## Raw Activity Data
### Git History (Last 20)
$GIT_LOG

### BeadsLog Activity Feed
$RAW_ACTIVITY

## Instructions for Agent:
1.  **Filter Noise**: Ignore typo fixes, minor doc tweaks, and internal refactors.
2.  **Update CHANGELOG.md**: Add an entry under "## [Unreleased]" (or create a new section if needed) following the "Added/Changed/Fixed" format.
3.  **Update internal/changelog/changelog.go**:
    - Add a new entry to the 'entries' slice.
    - Update the 'CurrentVersion' constant to "$NEW_VERSION".
    - Focus on high-signal features and mandatory protocol changes for agents.
4.  **Review**: Ensure the signal-to-noise ratio is high.

**STOP**: When you have updated the files, ask the user to review.
EOF

    echo -e "\n${YELLOW}🤖 ACTION REQUIRED:${NC}"
    echo -e "A drafting prompt has been created at: ${BLUE}_rules/_prompts/_draft-release.md${NC}"
    echo -e "Please ask your agent to read it and update the changelogs."
    read -p "Press [Enter] once the agent has finished and you have reviewed the changes..."
    rm "$PROMPT_FILE"
fi

# 4. Execute Bump and Build
echo -e "\n${YELLOW}Applying version bump...${NC}"
"$SCRIPT_DIR/bump-version.sh" "$NEW_VERSION" --allow-staged --commit

echo -e "${YELLOW}Building tool...${NC}"
go build -o bd ./cmd/bd/

# 5. Publish
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
echo -e "\n${YELLOW}Publishing...${NC}"
read -p "Do you want to publish (push to remote)? [y/N] " DO_PUBLISH
if [[ $DO_PUBLISH =~ ^[Yy]$ ]]; then
    echo "1) Push to current branch ($CURRENT_BRANCH)"
    echo "2) Push to main"
    read -p "Select [1-2]: " PUSH_CHOICE
    
    TARGET_BRANCH=$CURRENT_BRANCH
    if [ "$PUSH_CHOICE" == "2" ]; then
        TARGET_BRANCH="main"
    fi
    
    echo "Pushing to $TARGET_BRANCH..."
    git push origin "$TARGET_BRANCH"
    
    read -p "Create and push tag v$NEW_VERSION? [y/N] " DO_TAG
    if [[ $DO_TAG =~ ^[Yy]$ ]]; then
        git tag -a "v$NEW_VERSION" -m "Release v$NEW_VERSION"
        git push origin "v$NEW_VERSION"
    fi
fi

# 6. Display install commands
SHORT_SHA=$(git rev-parse --short HEAD)
echo -e "\n${GREEN}✨ Release Process Complete!${NC}"
echo -e "Install command for this build (short hash — most reliable):"
echo -e "  ${BLUE}go install github.com/untoldecay/BeadsLog/cmd/bd@$SHORT_SHA${NC}"
echo -e "If the module proxy hasn't seen this commit yet:"
echo -e "  ${BLUE}GOPROXY=direct go install github.com/untoldecay/BeadsLog/cmd/bd@$SHORT_SHA${NC}"
