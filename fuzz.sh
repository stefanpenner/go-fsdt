#!/bin/bash

# Run all fuzz tests for go-fsdt library
# This script mirrors the GitHub Actions fuzz testing configuration

set -e

echo "🚀 Running comprehensive fuzz tests for go-fsdt..."
echo "=================================================="

# Configuration (matching GitHub Actions workflows)
DEFAULT_FUZZ_DURATION=2        # minutes (matching go.yml)
INTENSIVE_FUZZ_DURATION=10     # minutes (matching fuzz.yml)
QUICK_FUZZ_DURATION=1          # minutes (matching fuzz-quick.yml)
PARALLEL_WORKERS=8             # workers (matching go.yml)
INTENSIVE_WORKERS=16           # workers (matching fuzz.yml)
QUICK_WORKERS=4                # workers (matching fuzz-quick.yml)

# Fuzz test names (matching all workflows)
FUZZ_TESTS=(
    "FuzzFolderCreation"
    "FuzzFileOperations"
    "FuzzLinkOperations"
    "FuzzFolderOperations"
    "FuzzDiffOperations"
    "FuzzEdgeCases"
    "FuzzSerialization"
    "FuzzMemoryStress"
)

echo "⚙️  Configuration:"
echo "   - Default fuzz duration: ${DEFAULT_FUZZ_DURATION}m"
echo "   - Intensive fuzz duration: ${INTENSIVE_FUZZ_DURATION}m"
echo "   - Quick fuzz duration: ${QUICK_FUZZ_DURATION}m"
echo "   - Default parallel workers: $PARALLEL_WORKERS"
echo "   - Intensive parallel workers: $INTENSIVE_WORKERS"
echo "   - Quick parallel workers: $QUICK_WORKERS"
echo ""

# Check if we're in CI environment
if [ -n "$CI" ]; then
    echo "🔍 CI environment detected, using CI settings"
    FUZZ_DURATION=$DEFAULT_FUZZ_DURATION
    WORKERS=$PARALLEL_WORKERS
else
    echo "💻 Local environment detected, using intensive settings"
    FUZZ_DURATION=$INTENSIVE_FUZZ_DURATION
    WORKERS=$INTENSIVE_WORKERS
fi

echo "🎯 Using fuzz duration: ${FUZZ_DURATION}m with $WORKERS workers"
echo ""

# Install dependencies
echo "📦 Installing Go dependencies..."
go mod download
echo "✅ Dependencies installed"
echo ""

# Create testdata directory structure
echo "📁 Creating testdata directory structure..."
mkdir -p testdata/fuzz
for test in "${FUZZ_TESTS[@]}"; do
    mkdir -p "testdata/fuzz/$test"
done
echo "✅ Testdata directories created"
echo ""

# Function to run a fuzz test
run_fuzz_test() {
    local test_name=$1
    local duration=$2
    local workers=$3
    local description=$4
    
    echo "🧪 Running $test_name..."
    echo "📝 $description"
    echo "⏱️  Duration: ${duration}m"
    echo "👥 Workers: $workers"
    echo "⏳ Starting fuzz test..."
    
    # Calculate timeout (duration + 2 minutes buffer)
    local timeout_seconds=$((duration * 60 + 120))
    
    if timeout ${timeout_seconds}s go test -fuzz=$test_name -fuzztime=${duration}m -parallel=$workers -v; then
        echo "✅ $test_name PASSED"
        return 0
    else
        echo "❌ $test_name FAILED"
        return 1
    fi
}

# Function to run quick fuzz tests (matching fuzz-quick.yml)
run_quick_fuzz_tests() {
    echo "⚡ Running quick fuzz tests (matching fuzz-quick.yml)..."
    echo "⏱️  Duration: ${QUICK_FUZZ_DURATION}m per test"
    echo "👥 Workers: $QUICK_WORKERS per test"
    echo ""
    
    local failed_tests=0
    
    for test in "${FUZZ_TESTS[@]}"; do
        if ! run_fuzz_test "$test" "$QUICK_FUZZ_DURATION" "$QUICK_WORKERS" "Quick fuzz test for $test"; then
            failed_tests=$((failed_tests + 1))
        fi
        echo ""
    done
    
    if [ $failed_tests -eq 0 ]; then
        echo "🎉 All quick fuzz tests passed!"
    else
        echo "⚠️  $failed_tests quick fuzz tests failed"
    fi
    
    return $failed_tests
}

# Function to run default fuzz tests (matching go.yml)
run_default_fuzz_tests() {
    echo "🧪 Running default fuzz tests (matching go.yml)..."
    echo "⏱️  Duration: ${DEFAULT_FUZZ_DURATION}m per test"
    echo "👥 Workers: $PARALLEL_WORKERS per test"
    echo ""
    
    local failed_tests=0
    
    for test in "${FUZZ_TESTS[@]}"; do
        if ! run_fuzz_test "$test" "$DEFAULT_FUZZ_DURATION" "$PARALLEL_WORKERS" "Default fuzz test for $test"; then
            failed_tests=$((failed_tests + 1))
        fi
        echo ""
    done
    
    if [ $failed_tests -eq 0 ]; then
        echo "🎉 All default fuzz tests passed!"
    else
        echo "⚠️  $failed_tests default fuzz tests failed"
    fi
    
    return $failed_tests
}

# Function to run intensive fuzz tests (matching fuzz.yml)
run_intensive_fuzz_tests() {
    echo "🚀 Running intensive fuzz tests (matching fuzz.yml)..."
    echo "⏱️  Duration: ${INTENSIVE_FUZZ_DURATION}m per test"
    echo "👥 Workers: $INTENSIVE_WORKERS per test"
    echo ""
    
    local failed_tests=0
    
    for test in "${FUZZ_TESTS[@]}"; do
        if ! run_fuzz_test "$test" "$INTENSIVE_FUZZ_DURATION" "$INTENSIVE_WORKERS" "Intensive fuzz test for $test"; then
            failed_tests=$((failed_tests + 1))
        fi
        echo ""
    done
    
    if [ $failed_tests -eq 0 ]; then
        echo "🎉 All intensive fuzz tests passed!"
    else
        echo "⚠️  $failed_tests intensive fuzz tests failed"
    fi
    
    return $failed_tests
}

# Main execution
echo "🎯 Starting fuzz test execution..."
echo ""

# Run the appropriate test suite based on environment
if [ -n "$CI" ]; then
    echo "🔍 CI environment: Running default fuzz tests"
    run_default_fuzz_tests
else
    echo "💻 Local environment: Running intensive fuzz tests"
    run_intensive_fuzz_tests
fi

echo ""
echo "🎉 Fuzz testing completed!"
echo "=================================================="
echo "📊 Summary:"
echo "   - Environment: $([ -n "$CI" ] && echo "CI" || echo "Local")"
echo "   - Fuzz duration: ${FUZZ_DURATION}m"
echo "   - Workers: $WORKERS"
echo "   - Tests run: ${#FUZZ_TESTS[@]}"
echo ""
echo "💡 Next steps:"
echo "   - Review any corpus files generated"
echo "   - Run './test.sh' for regular testing (includes performance tests)"
echo "   - Check GitHub Actions for automated fuzz testing"
echo "   - Run 'go test -bench=.' for detailed performance benchmarks"
echo ""
echo "🔧 Configuration used:"
echo "   - Default: ${DEFAULT_FUZZ_DURATION}m with $PARALLEL_WORKERS workers"
echo "   - Intensive: ${INTENSIVE_FUZZ_DURATION}m with $INTENSIVE_WORKERS workers"
echo "   - Quick: ${QUICK_FUZZ_DURATION}m with $QUICK_WORKERS workers"
