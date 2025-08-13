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
    echo "======================================================"
    
    local failed_tests=0
    
    for test in "${FUZZ_TESTS[@]}"; do
        case $test in
            "FuzzFolderCreation")
                run_fuzz_test "$test" $QUICK_FUZZ_DURATION $QUICK_WORKERS "Tests folder creation with various inputs including edge cases" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzFileOperations")
                run_fuzz_test "$test" $QUICK_FUZZ_DURATION $QUICK_WORKERS "Tests file operations with various content types and modes" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzLinkOperations")
                run_fuzz_test "$test" $QUICK_FUZZ_DURATION $QUICK_WORKERS "Tests link operations with various target types" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzFolderOperations")
                run_fuzz_test "$test" $QUICK_FUZZ_DURATION $QUICK_WORKERS "Tests complex folder operations and cloning" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzDiffOperations")
                run_fuzz_test "$test" $QUICK_FUZZ_DURATION $QUICK_WORKERS "Tests diff operations with various folder structures" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzEdgeCases")
                run_fuzz_test "$test" $QUICK_FUZZ_DURATION $QUICK_WORKERS "Tests various edge cases and error conditions" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzSerialization")
                run_fuzz_test "$test" $QUICK_FUZZ_DURATION $QUICK_WORKERS "Tests serialization and deserialization edge cases" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzMemoryStress")
                run_fuzz_test "$test" $QUICK_FUZZ_DURATION $QUICK_WORKERS "Tests memory allocation and stress scenarios" || failed_tests=$((failed_tests + 1))
                ;;
        esac
        echo ""
    done
    
    return $failed_tests
}

# Function to run default fuzz tests (matching go.yml)
run_default_fuzz_tests() {
    echo "🔄 Running default fuzz tests (matching go.yml)..."
    echo "================================================"
    
    local failed_tests=0
    
    for test in "${FUZZ_TESTS[@]}"; do
        case $test in
            "FuzzFolderCreation")
                run_fuzz_test "$test" $DEFAULT_FUZZ_DURATION $PARALLEL_WORKERS "Tests folder creation with various inputs including edge cases" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzFileOperations")
                run_fuzz_test "$test" $DEFAULT_FUZZ_DURATION $PARALLEL_WORKERS "Tests file operations with various content types and modes" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzLinkOperations")
                run_fuzz_test "$test" $DEFAULT_FUZZ_DURATION $PARALLEL_WORKERS "Tests link operations with various target types" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzFolderOperations")
                run_fuzz_test "$test" $DEFAULT_FUZZ_DURATION $PARALLEL_WORKERS "Tests complex folder operations and cloning" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzDiffOperations")
                run_fuzz_test "$test" $DEFAULT_FUZZ_DURATION $PARALLEL_WORKERS "Tests diff operations with various folder structures" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzEdgeCases")
                run_fuzz_test "$test" $DEFAULT_FUZZ_DURATION $PARALLEL_WORKERS "Tests various edge cases and error conditions" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzSerialization")
                run_fuzz_test "$test" $DEFAULT_FUZZ_DURATION $PARALLEL_WORKERS "Tests serialization and deserialization edge cases" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzMemoryStress")
                run_fuzz_test "$test" $DEFAULT_FUZZ_DURATION $PARALLEL_WORKERS "Tests memory allocation and stress scenarios" || failed_tests=$((failed_tests + 1))
                ;;
        esac
        echo ""
    done
    
    return $failed_tests
}

# Function to run intensive fuzz tests (matching fuzz.yml)
run_intensive_fuzz_tests() {
    echo "🔥 Running intensive fuzz tests (matching fuzz.yml)..."
    echo "==================================================="
    
    local failed_tests=0
    
    for test in "${FUZZ_TESTS[@]}"; do
        case $test in
            "FuzzFolderCreation")
                run_fuzz_test "$test" $INTENSIVE_FUZZ_DURATION $INTENSIVE_WORKERS "Tests folder creation with various inputs including edge cases" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzFileOperations")
                run_fuzz_test "$test" $INTENSIVE_FUZZ_DURATION $INTENSIVE_WORKERS "Tests file operations with various content types and modes" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzLinkOperations")
                run_fuzz_test "$test" $INTENSIVE_FUZZ_DURATION $INTENSIVE_WORKERS "Tests link operations with various target types" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzFolderOperations")
                run_fuzz_test "$test" $INTENSIVE_FUZZ_DURATION $INTENSIVE_WORKERS "Tests complex folder operations and cloning" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzDiffOperations")
                run_fuzz_test "$test" $INTENSIVE_FUZZ_DURATION $INTENSIVE_WORKERS "Tests diff operations with various folder structures" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzEdgeCases")
                run_fuzz_test "$test" $INTENSIVE_FUZZ_DURATION $INTENSIVE_WORKERS "Tests various edge cases and error conditions" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzSerialization")
                run_fuzz_test "$test" $INTENSIVE_FUZZ_DURATION $INTENSIVE_WORKERS "Tests serialization and deserialization edge cases" || failed_tests=$((failed_tests + 1))
                ;;
            "FuzzMemoryStress")
                run_fuzz_test "$test" $INTENSIVE_FUZZ_DURATION $INTENSIVE_WORKERS "Tests memory allocation and stress scenarios" || failed_tests=$((failed_tests + 1))
                ;;
        esac
        echo ""
    done
    
    return $failed_tests
}

# Function to analyze fuzz corpus
analyze_fuzz_corpus() {
    echo "📊 Analyzing fuzz corpus data..."
    echo "=================================="
    
    local total_tests=0
    local total_corpus_files=0
    
    for test in "${FUZZ_TESTS[@]}"; do
        if [ -d "testdata/fuzz/$test" ]; then
            local corpus_count=$(find "testdata/fuzz/$test" -type f | wc -l)
            echo "🧪 $test: $corpus_count corpus files"
            total_tests=$((total_tests + 1))
            total_corpus_files=$((total_corpus_files + corpus_count))
        else
            echo "🧪 $test: No corpus found"
        fi
    done
    
    echo ""
    echo "📈 Summary:"
    echo "   - Total tests: $total_tests"
    echo "   - Total corpus files: $total_corpus_files"
    if [ $total_tests -gt 0 ]; then
        echo "   - Average corpus files per test: $((total_corpus_files / total_tests))"
    fi
    echo ""
}

# Main execution
echo "🚀 Starting fuzz test execution..."
echo ""

# Run quick fuzz tests first (fast feedback)
if ! run_quick_fuzz_tests; then
    echo "❌ Quick fuzz tests failed, stopping execution"
    exit 1
fi

echo "✅ Quick fuzz tests completed successfully"
echo ""

# Run default fuzz tests (matching GitHub Actions)
if ! run_default_fuzz_tests; then
    echo "❌ Default fuzz tests failed, stopping execution"
    exit 1
fi

echo "✅ Default fuzz tests completed successfully"
echo ""

# Run intensive fuzz tests (if not in CI)
if [ -z "$CI" ]; then
    if ! run_intensive_fuzz_tests; then
        echo "❌ Intensive fuzz tests failed, stopping execution"
        exit 1
    fi
    
    echo "✅ Intensive fuzz tests completed successfully"
    echo ""
fi

# Analyze corpus
analyze_fuzz_corpus

echo "🎉 All fuzz tests completed successfully!"
echo "=================================================="
echo "📊 Summary:"
echo "   ✅ Quick fuzz tests: All passed"
echo "   ✅ Default fuzz tests: All passed"
if [ -z "$CI" ]; then
    echo "   ✅ Intensive fuzz tests: All passed"
fi
echo ""
echo "📁 Generated files:"
echo "   - testdata/fuzz/*/ (fuzz corpus files)"
echo ""
echo "💡 Next steps:"
echo "   - Review any corpus files generated"
echo "   - Run './test.sh' for regular testing"
echo "   - Check GitHub Actions for automated fuzz testing"
echo ""
echo "🔧 Configuration used:"
echo "   - Quick tests: ${QUICK_FUZZ_DURATION}m with $QUICK_WORKERS workers"
echo "   - Default tests: ${DEFAULT_FUZZ_DURATION}m with $PARALLEL_WORKERS workers"
if [ -z "$CI" ]; then
    echo "   - Intensive tests: ${INTENSIVE_FUZZ_DURATION}m with $INTENSIVE_WORKERS workers"
fi
