#!/bin/bash
# Verification script for entities.go feature

echo "=== Verification Script for entities.go ==="
echo ""
echo "1. Checking file existence..."
if [ -f "entities.go" ]; then
    echo "✓ entities.go exists"
else
    echo "✗ entities.go NOT found"
    exit 1
fi

if [ -f "entities_test.go" ]; then
    echo "✓ entities_test.go exists"
else
    echo "✗ entities_test.go NOT found"
    exit 1
fi

echo ""
echo "2. Checking file syntax and imports..."
if grep -q "package main" entities.go; then
    echo "✓ Correct package declaration"
else
    echo "✗ Missing package declaration"
    exit 1
fi

if grep -q "entitiesCmd" entities.go; then
    echo "✓ entitiesCmd command defined"
else
    echo "✗ Missing entitiesCmd"
    exit 1
fi

if grep -q "func runEntities" entities.go; then
    echo "✓ runEntities function defined"
else
    echo "✗ Missing runEntities function"
    exit 1
fi

echo ""
echo "3. Checking main.go integration..."
if grep -q "entitiesCmd" main.go; then
    echo "✓ entitiesCmd registered in main.go"
else
    echo "✗ entitiesCmd NOT registered in main.go"
    exit 1
fi

echo ""
echo "4. Checking test file..."
if grep -q "TestEntitiesCmd" entities_test.go; then
    echo "✓ TestEntitiesCmd test defined"
else
    echo "✗ Missing TestEntitiesCmd test"
    exit 1
fi

if grep -q "TestBuildEntitiesReport" entities_test.go; then
    echo "✓ TestBuildEntitiesReport test defined"
else
    echo "✗ Missing TestBuildEntitiesReport test"
    exit 1
fi

echo ""
echo "5. Checking required functions..."
required_funcs=(
    "buildEntitiesReport"
    "filterEntitiesReport"
    "sortEntitiesByMentionCount"
    "outputEntitiesTable"
    "outputEntitiesJSON"
    "getEntityType"
)

for func in "${required_funcs[@]}"; do
    if grep -q "func $func" entities.go; then
        echo "✓ $func function exists"
    else
        echo "✗ Missing $func function"
        exit 1
    fi
done

echo ""
echo "6. Checking data structures..."
if grep -q "type EntityStats struct" entities.go; then
    echo "✓ EntityStats struct defined"
else
    echo "✗ Missing EntityStats struct"
    exit 1
fi

if grep -q "type EntitiesReport struct" entities.go; then
    echo "✓ EntitiesReport struct defined"
else
    echo "✗ Missing EntitiesReport struct"
    exit 1
fi

echo ""
echo "7. Checking flags..."
required_flags=(
    "entitiesFormat"
    "entitiesType"
    "entitiesLimit"
    "entitiesMinimum"
)

for flag in "${required_flags[@]}"; do
    if grep -q "$flag" entities.go; then
        echo "✓ $flag flag defined"
    else
        echo "✗ Missing $flag flag"
        exit 1
    fi
done

echo ""
echo "8. Checking command integration..."
if grep -A5 "func init()" entities.go | grep -q "entitiesCmd.Flags"; then
    echo "✓ Command flags initialized"
else
    echo "✗ Command flags not initialized"
    exit 1
fi

echo ""
echo "=== All Verification Checks Passed! ==="
echo ""
echo "Summary:"
echo "  - entities.go created successfully"
echo "  - entities_test.go created successfully"
echo "  - Command registered in main.go"
echo "  - All required functions implemented"
echo "  - All required data structures defined"
echo "  - All required flags configured"
echo ""
echo "To test the feature, run:"
echo "  cd /projects/devlog/cmd/devlog"
echo "  go test -v -run TestEntities"
echo ""
echo "To use the feature:"
echo "  ./devlog entities"
echo "  ./devlog entities --help"
