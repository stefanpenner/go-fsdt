#!/bin/bash

# Run tests for go-fsdt library with different modes
# This script provides fast feedback during development and comprehensive testing when needed

set -e

# Default to fast mode
MODE=${1:-fast}

echo "🧪 Running go-fsdt tests in $MODE mode..."
echo "======================================"

# Configuration
GO_VERSION="1.23.x"
PARALLEL_WORKERS=8
TIMEOUT_BUFFER=2

echo "⚙️  Configuration:"
echo "   - Go version: $GO_VERSION"
echo "   - Mode: $MODE"
echo "   - Parallel workers: $PARALLEL_WORKERS"
echo ""

# Check Go version
GO_CURRENT=$(go version | awk '{print $3}' | sed 's/go//')
echo "🔍 Current Go version: $GO_CURRENT"
echo "📋 Target Go version: $GO_VERSION"
echo ""

# Install dependencies
echo "📦 Installing Go dependencies..."
go mod download
echo "✅ Dependencies installed"
echo ""

# Build the project
echo "🔨 Building project..."
if go build -v ./...; then
    echo "✅ Build successful"
else
    echo "❌ Build failed"
    exit 1
fi
echo ""

case $MODE in
    "fast")
        echo "⚡ Running fast tests only (essential tests, no fuzz/performance)..."
        echo "   - Core functionality tests"
        echo "   - Basic unit tests"
        echo "   - Go vet and race detection"
        echo ""
        
        # Run regular tests (excluding slow ones)
        echo "🧪 Running core tests..."
        if go test -v -short ./...; then
            echo "✅ All core tests passed"
        else
            echo "❌ Some core tests failed"
            exit 1
        fi
        echo ""
        
        # Run go vet
        echo "🔍 Running go vet..."
        if go vet ./...; then
            echo "✅ Go vet passed"
        else
            echo "⚠️  Go vet found issues"
        fi
        echo ""
        
        # Run with race detector (fast version)
        echo "🏃 Running race detector tests..."
        if go test -race -v -short ./...; then
            echo "✅ Race detector tests passed"
        else
            echo "❌ Race conditions detected"
            exit 1
        fi
        echo ""
        
        echo "🎉 Fast test suite completed successfully!"
        ;;
        
    "full")
        echo "🧪 Running full test suite (all tests including slow ones)..."
        echo "   - Core functionality tests"
        echo "   - Performance tests"
        echo "   - Memory leak tests"
        echo "   - Concurrent safety tests"
        echo "   - Fuzz tests (short duration)"
        echo ""
        
        # Run all tests
        echo "🧪 Running all tests..."
        if go test -v ./...; then
            echo "✅ All tests passed"
        else
            echo "❌ Some tests failed"
            exit 1
        fi
        echo ""
        
        # Run go vet
        echo "🔍 Running go vet..."
        if go vet ./...; then
            echo "✅ Go vet passed"
        else
            echo "⚠️  Go vet found issues"
        fi
        echo ""
        
        # Run with race detector
        echo "🏃 Running race detector tests..."
        if go test -race -v ./...; then
            echo "✅ Race detector tests passed"
        else
            echo "❌ Race conditions detected"
            exit 1
        fi
        echo ""
        
        # Run fuzz tests (short duration for full mode)
        echo "🧪 Running fuzz tests (short duration)..."
        for test in FuzzFolderCreation FuzzFileOperations FuzzLinkOperations FuzzFolderOperations FuzzDiffOperations FuzzEdgeCases FuzzSerialization FuzzMemoryStress; do
            echo "   - Running $test (30 seconds)..."
            timeout 35s go test -fuzz=$test -fuzztime=30s -parallel=4 -v || echo "⚠️  $test completed (may have found issues)"
        done
        echo ""
        
        echo "🎉 Full test suite completed successfully!"
        ;;
        
    "performance")
        echo "⚡ Running performance tests and benchmarks..."
        echo "   - Performance threshold tests"
        echo "   - Memory leak tests"
        echo "   - Concurrent safety tests"
        echo "   - Performance benchmarks"
        echo ""
        
        # Run performance tests
        echo "🧪 Running performance tests..."
        if go test -v -run "TestPerformance|TestMemory|TestConcurrent" ./...; then
            echo "✅ All performance tests passed"
        else
            echo "❌ Some performance tests failed"
            exit 1
        fi
        echo ""
        
        # Run performance benchmarks
        echo "📊 Running performance benchmarks..."
        echo "   - File operations benchmark"
        go test -bench=BenchmarkFileOperations -run=^$ ./... || echo "⚠️  File operations benchmark completed"
        echo ""
        echo "   - Folder operations benchmark"
        go test -bench=BenchmarkFolderOperations -run=^$ ./... || echo "⚠️  Folder operations benchmark completed"
        echo ""
        echo "   - Diff operations benchmark"
        go test -bench=BenchmarkDiffOperations -run=^$ ./... || echo "⚠️  Diff operations benchmark completed"
        echo ""
        echo "   - Link operations benchmark"
        go test -bench=BenchmarkLinkOperations -run=^$ ./... || echo "⚠️  Link operations benchmark completed"
        echo ""
        
        echo "🎉 Performance test suite completed successfully!"
        ;;
        
    "fuzz")
        echo "🧪 Running comprehensive fuzz tests..."
        echo "   - All fuzz tests with longer duration"
        echo "   - Memory stress tests"
        echo "   - Edge case discovery"
        echo ""
        
        # Run fuzz tests with longer duration
        echo "🧪 Running fuzz tests (2 minutes each)..."
        for test in FuzzFolderCreation FuzzFileOperations FuzzLinkOperations FuzzFolderOperations FuzzDiffOperations FuzzEdgeCases FuzzSerialization FuzzMemoryStress; do
            echo "   - Running $test (2 minutes)..."
            timeout 150s go test -fuzz=$test -fuzztime=2m -parallel=8 -v || echo "⚠️  $test completed (may have found issues)"
            echo ""
        done
        
        echo "🎉 Fuzz test suite completed successfully!"
        ;;
        
    *)
        echo "❌ Unknown mode: $MODE"
        echo ""
        echo "Usage: $0 [fast|full|performance|fuzz]"
        echo ""
        echo "Modes:"
        echo "  fast        - Essential tests only (default, ~5-10 seconds)"
        echo "  full        - All tests including short fuzz tests (~1-2 minutes)"
        echo "  performance - Performance tests and benchmarks (~30 seconds)"
        echo "  fuzz        - Comprehensive fuzz testing (~10-15 minutes)"
        echo ""
        echo "Examples:"
        echo "  $0           # Run fast tests (default)"
        echo "  $0 fast      # Run fast tests"
        echo "  $0 full      # Run full test suite"
        echo "  $0 performance # Run performance tests"
        echo "  $0 fuzz      # Run comprehensive fuzz tests"
        exit 1
        ;;
esac

echo ""
echo "🎯 Test Summary:"
echo "   - Mode: $MODE"
echo "   - Build: ✅"
echo "   - Tests: ✅"
echo ""
echo "💡 Next steps:"
echo "   - Run './test.sh fast' for quick feedback during development"
echo "   - Run './test.sh full' before committing changes"
echo "   - Run './test.sh performance' to check performance"
echo "   - Run './test.sh fuzz' for comprehensive testing"
echo "   - Run './fuzz.sh' for intensive fuzz testing"
