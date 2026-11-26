#!/bin/bash
# Cut a dev-channel pre-release (vX.Y.Z-dev.N) of the current commit.
#
# The version bump is committed in a throwaway worktree and only the tag is
# pushed, so neither the current branch nor main ever carries a dev version
# (a dev version merged into main would be offered to every stable user).
set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
die() { echo -e "${RED}Error: $*${NC}" >&2; exit 1; }

usage() {
    cat <<EOF
Usage: scripts/cut-release.sh --dev [options]

Cut vX.Y.Z-dev.N from the current commit and push only that tag.

  --base X.Y.Z   version being worked towards (default: the open dev series,
                 else the next minor after the latest stable tag)
  --remote NAME  remote to push to (default: origin)
  --no-github    push the tag only, do not create a GitHub pre-release
  --dry-run      build and show everything, push nothing
  --yes          do not ask for confirmation

Install it:   go install github.com/untoldecay/BeadsLog/cmd/bd@vX.Y.Z-dev.N
Follow them:  bd upgrade --channel dev      Go back:  bd upgrade --channel stable

A pushed tag is permanent: the Go module proxy never lets a version be reused.
To fix a bad dev build, cut the next one.
EOF
}

BASE="" REMOTE="origin" GITHUB=1 DRY_RUN=0 YES=0
while [[ $# -gt 0 ]]; do
    case $1 in
        --base)      BASE="${2:-}"; [[ -n "$BASE" ]] || die "--base needs a version"; shift 2 ;;
        --remote)    REMOTE="${2:-}"; [[ -n "$REMOTE" ]] || die "--remote needs a name"; shift 2 ;;
        --no-github) GITHUB=0; shift ;;
        --dry-run)   DRY_RUN=1; shift ;;
        --yes)       YES=1; shift ;;
        -h|--help)   usage; exit 0 ;;
        *)           usage >&2; die "unknown option $1" ;;
    esac
done

ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR"
HEAD_SHA=$(git rev-parse HEAD)
BRANCH=$(git rev-parse --abbrev-ref HEAD)

# The remote's tags are the truth: local ones may be stale.
REMOTE_TAGS=$(git ls-remote --tags --refs "$REMOTE" 'v*') || die "cannot list tags on $REMOTE"
TAGS=$(echo "$REMOTE_TAGS" | sed -n 's#.*refs/tags/v##p')
vsort() { sort -t. -k1,1n -k2,2n -k3,3n; }
STABLE=$(echo "$TAGS" | { grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' || true; } | vsort | tail -1)
[[ -n "$STABLE" ]] || die "no stable tag (vX.Y.Z) on $REMOTE"
newer_than_stable() { [[ "$1" != "$STABLE" && "$(printf '%s\n%s\n' "$1" "$STABLE" | vsort | tail -1)" == "$1" ]]; }

if [[ -z "$BASE" ]]; then
    OPEN=$(echo "$TAGS" | { grep -E '^[0-9]+\.[0-9]+\.[0-9]+-dev\.[0-9]+$' || true; } | sed 's/-dev\..*//' | vsort | tail -1)
    if [[ -n "$OPEN" ]] && newer_than_stable "$OPEN"; then
        BASE=$OPEN
    else
        BASE=$(echo "$STABLE" | awk -F. '{print $1"."$2+1".0"}')
    fi
fi
[[ "$BASE" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "--base must be X.Y.Z, got '$BASE'"
newer_than_stable "$BASE" || die "base $BASE is not newer than the latest stable $STABLE"

LAST_N=$(echo "$TAGS" | { grep -E "^${BASE//./\\.}-dev\.[0-9]+$" || true; } | sed 's/.*-dev\.//' | sort -n | tail -1)
VERSION="$BASE-dev.$(( ${LAST_N:-0} + 1 ))"
TAG="v$VERSION"
git rev-parse -q --verify "refs/tags/$TAG" > /dev/null && die "tag $TAG exists locally but not on $REMOTE; delete it first: git tag -d $TAG"

echo -e "${BLUE}Dev pre-release${NC}"
echo "  Version: $VERSION   (latest stable: $STABLE)"
echo "  Source:  $BRANCH @ ${HEAD_SHA:0:8} $(git log -1 --format=%s)"
echo "  Remote:  $REMOTE ($(git remote get-url "$REMOTE" 2>/dev/null || echo "$REMOTE"))"
if [[ -n "$(git status --porcelain --untracked-files=no)" ]]; then
    echo -e "  ${YELLOW}Uncommitted changes are NOT included.${NC}"
fi

TMP=$(mktemp -d)
TAGGED=0
cleanup() {
    local rc=$?
    git worktree remove --force "$TMP/src" > /dev/null 2>&1 || true
    git worktree prune
    rm -rf "$TMP"
    if [[ $rc -ne 0 && $TAGGED -eq 1 ]]; then git tag -d "$TAG" > /dev/null 2>&1 || true; fi
}
trap cleanup EXIT

# Hooks off: this worktree must not sync devlogs or stage beads files.
NOHOOKS=(-c core.hooksPath=/dev/null)
git "${NOHOOKS[@]}" worktree add -q --detach "$TMP/src" "$HEAD_SHA"
perl -pi -e 's/^(\s*Version = )"[^"]*"/$1"'"$VERSION"'"/' "$TMP/src/cmd/bd/version.go"
[[ "$(git -C "$TMP/src" status --porcelain)" == " M cmd/bd/version.go" ]] || die "version bump changed more or less than cmd/bd/version.go"

echo "  Building..."
(cd "$TMP/src" && go build -o "$TMP/bd" ./cmd/bd)
BUILT=$("$TMP/bd" --version | awk '{print $3}')
[[ "$BUILT" == "$VERSION" ]] || die "built binary reports '$BUILT', expected '$VERSION'"
echo -e "  ${GREEN}Built bd $BUILT${NC}"

if [[ $DRY_RUN -eq 1 ]]; then
    echo -e "${YELLOW}Dry run: nothing tagged or pushed.${NC}"
    exit 0
fi
if [[ $YES -ne 1 ]]; then
    read -r -p "Push $TAG to $REMOTE? This cannot be undone. [y/N] " CONFIRM
    [[ "$CONFIRM" =~ ^[Yy]$ ]] || { echo "Nothing pushed."; exit 0; }
fi

git -C "$TMP/src" "${NOHOOKS[@]}" commit -q -m "chore: dev pre-release $VERSION" -m "Cut from $BRANCH @ $HEAD_SHA. Tag-only commit: not on any branch." -- cmd/bd/version.go
git -C "$TMP/src" tag -a "$TAG" -m "Dev pre-release $TAG"
TAGGED=1
git -C "$TMP/src" "${NOHOOKS[@]}" push -q "$REMOTE" "refs/tags/$TAG"
TAGGED=0 # published: keep the local tag whatever happens next
echo -e "${GREEN}Pushed $TAG${NC}"

if [[ $GITHUB -eq 1 ]]; then
    NOTES="$TMP/notes.md"
    {
        echo "Dev-channel pre-release of \`$BRANCH\` @ ${HEAD_SHA:0:8}. Not a stable release."
        echo
        echo '```bash'
        echo "go install github.com/untoldecay/BeadsLog/cmd/bd@$TAG"
        echo "bd upgrade --channel dev      # follow dev builds"
        echo "bd upgrade --channel stable   # go back"
        echo '```'
        awk '/^## \[Unreleased\]/{on=1; next} /^## \[/{on=0} on' "$TMP/src/CHANGELOG.md"
    } > "$NOTES"
    # A missing release page does not make the tag any less installable.
    if ! command -v gh > /dev/null 2>&1 ||
        ! gh release create "$TAG" --prerelease --latest=false --verify-tag --title "$TAG" --notes-file "$NOTES"; then
        echo -e "${YELLOW}GitHub pre-release not created; the tag is installable anyway.${NC}"
    fi
fi

echo
echo "Install: go install github.com/untoldecay/BeadsLog/cmd/bd@$TAG"
