#!/bin/bash

# Run all fuzz tests for the go-fsdt library
# This script runs each fuzz test for a specified duration to find potential bugs

set -e

echo "🚀 Starting fuzz tests for go-fsdt library..."
echo "================================================"

# Configuration
FUZZ_TIME="30s"
WORKERS=8

echo "⏱️  Fuzz time per test: $FUZZ_TIME"
echo "👥 Workers per test: $WORKERS"
echo ""

# Function to run a fuzz test
run_fuzz_test() {
    local test_name=$1
    local description=$2
    
    echo "🧪 Running $test_name..."
    echo "📝 $description"
    echo "⏳ Starting fuzz test (this may take a while)..."
    
    if go test -fuzz=$test_name -fuzztime=$FUZZ_TIME -fuzzminimizetime=0s -parallel=$WORKERS; then
        echo "✅ $test_name PASSED"
    else
        echo "❌ $test_name FAILED - Found a bug!"
        return 1
    fi
    
    echo ""
}

# Run all fuzz tests
echo "🔄 Running all fuzz tests..."
echo ""

run_fuzz_test "FuzzFolderCreation" "Tests folder creation with various inputs including edge cases"
run_fuzz_test "FuzzFileOperations" "Tests file operations with various content types and modes"
run_fuzz_test "FuzzLinkOperations" "Tests link operations with various target types"
run_fuzz_test "FuzzFolderOperations" "Tests complex folder operations and cloning"
run_fuzz_test "FuzzDiffOperations" "Tests diff operations with various folder structures"
run_fuzz_test "FuzzEdgeCases" "Tests various edge cases and error conditions"
run_fuzz_test "FuzzSerialization" "Tests serialization and deserialization edge cases"
run_fuzz_test "FuzzMemoryStress" "Tests memory allocation and stress scenarios"

echo "🎉 All fuzz tests completed!"
echo "================================================"
echo "📊 Summary:"
echo "   - Folder creation: ✅"
echo "   - File operations: ✅"
echo "   - Link operations: ✅"
echo "   - Folder operations: ✅"
echo "   - Diff operations: ✅"
echo "   - Edge cases: ✅"
echo "   - Serialization: ✅"
echo "   - Memory stress: ✅"
echo ""
echo "🔍 No bugs found in this run!"
echo "💡 Consider running longer fuzz tests for deeper coverage"
