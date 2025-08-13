package fsdt

import (
	"fmt"
	"testing"
	"time"

	op "github.com/stefanpenner/go-fsdt/operation"
)

// Performance thresholds - adjust these based on your performance requirements
const (
	// File operations
	MaxFileCreateTime = 100 * time.Microsecond
	MaxFileReadTime   = 100 * time.Microsecond

	// Folder operations
	MaxFolderCreateTime = 400 * time.Microsecond
	MaxFolderReadTime   = 700 * time.Microsecond
	MaxFolderCloneTime  = 500 * time.Microsecond
	MaxFolderEqualTime  = 200 * time.Microsecond

	// Link operations
	MaxLinkCreateTime = 100 * time.Microsecond
	MaxLinkReadTime   = 100 * time.Microsecond

	// Diff operations
	MaxDiffTime = 1 * time.Millisecond
)

// BenchmarkFileOperations benchmarks file creation and reading
func BenchmarkFileOperations(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create file
		start := time.Now()
		file := NewFile(FileOptions{Content: []byte("test content"), Mode: 0644})
		createTime := time.Since(start)

		if createTime > MaxFileCreateTime {
			b.Errorf("File creation took too long: %v (max: %v)", createTime, MaxFileCreateTime)
		}

		// Read file
		start = time.Now()
		_ = file.Content()
		_ = file.Mode()
		_ = file.Type()
		readTime := time.Since(start)

		if readTime > MaxFileReadTime {
			b.Errorf("File read operations took too long: %v (max: %v)", readTime, MaxFileReadTime)
		}
	}
}

// BenchmarkFolderOperations benchmarks folder creation, reading, and operations
func BenchmarkFolderOperations(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create folder
		start := time.Now()
		folder := NewFolder()
		createTime := time.Since(start)

		if createTime > MaxFolderCreateTime {
			b.Errorf("Folder creation took too long: %v (max: %v)", createTime, MaxFolderCreateTime)
		}

		// Add some entries
		start = time.Now()
		folder.File("file1.txt", FileOptions{Content: []byte("file1 content"), Mode: 0644})
		folder.File("file2.txt", FileOptions{Content: []byte("file2 content"), Mode: 0644})
		folder.Folder("subfolder")
		addTime := time.Since(start)

		if addTime > MaxFolderCreateTime {
			b.Errorf("Adding folder entries took too long: %v (max: %v)", addTime, MaxFolderCreateTime)
		}

		// Read folder
		start = time.Now()
		_ = folder.Entries()
		_ = folder.Get("file1.txt")
		_ = folder.Get("file2.txt")
		readTime := time.Since(start)

		if readTime > MaxFolderReadTime {
			b.Errorf("Folder read operations took too long: %v (max: %v)", readTime, MaxFolderReadTime)
		}

		// Clone folder
		start = time.Now()
		clone := folder.Clone()
		cloneTime := time.Since(start)

		if cloneTime > MaxFolderCloneTime {
			b.Errorf("Folder cloning took too long: %v (max: %v)", cloneTime, MaxFolderCloneTime)
		}

		// Test equality
		start = time.Now()
		equal := folder.Equal(clone)
		equalTime := time.Since(start)

		if equalTime > MaxFolderEqualTime {
			b.Errorf("Folder equality check took too long: %v (max: %v)", equalTime, MaxFolderEqualTime)
		}

		if !equal {
			b.Error("Cloned folder should equal original")
		}
	}
}

// BenchmarkLinkOperations benchmarks link creation and reading
func BenchmarkLinkOperations(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create link
		start := time.Now()
		link := NewLink("target.txt", SYMLINK)
		createTime := time.Since(start)

		if createTime > MaxLinkCreateTime {
			b.Errorf("Link creation took too long: %v (max: %v)", createTime, MaxLinkCreateTime)
		}

		// Read link
		start = time.Now()
		_ = link.Target()
		_ = link.Type()
		readTime := time.Since(start)

		if readTime > MaxLinkReadTime {
			b.Errorf("Link read operations took too long: %v (max: %v)", readTime, MaxLinkReadTime)
		}
	}
}

// BenchmarkDiffOperations benchmarks diff operations on folders
func BenchmarkDiffOperations(b *testing.B) {
	// Create two similar folders
	folder1 := createTestFolder("folder1", 3, 2)
	folder2 := createTestFolder("folder2", 3, 2)

	// Modify folder2 slightly
	folder2.File("extra.txt", FileOptions{Content: []byte("extra content"), Mode: 0644})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		start := time.Now()
		diff := Diff(folder1, folder2, true)
		diffTime := time.Since(start)

		if diffTime > MaxDiffTime {
			b.Errorf("Diff operation took too long: %v (max: %v)", diffTime, MaxDiffTime)
		}

		// Check that diff contains operations (folders are different)
		if diff.Operand == op.Noop {
			b.Error("Different folders should have diff operations")
		}
	}
}

// BenchmarkConcurrentOperations benchmarks concurrent access to folder structures
func BenchmarkConcurrentOperations(b *testing.B) {
	folder := createTestFolder("concurrent_test", 3, 2)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Read operations
			_ = folder.Entries()
			_ = folder.Get("file1.txt")

			// Clone operation
			clone := folder.Clone()

			// Equality check
			_ = folder.Equal(clone)
		}
	})
}

// BenchmarkStressTest performs stress testing with many operations
func BenchmarkStressTest(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create a complex structure
		root := NewFolder()

		// Add many files and subfolders
		for j := 0; j < 50; j++ {
			subfolder := root.Folder(fmt.Sprintf("sub_%d", j))
			for k := 0; k < 5; k++ {
				subfolder.File(fmt.Sprintf("file_%d_%d.txt", j, k),
					FileOptions{Content: []byte(fmt.Sprintf("content %d_%d", j, k)), Mode: 0644})
			}
		}

		// Perform operations
		clone := root.Clone()
		cloneFolder := clone.(*Folder)
		diff := Diff(root, cloneFolder, true)

		if diff.Operand == op.Noop {
			b.Error("Diff should contain operations")
		}

		if !root.Equal(clone) {
			b.Error("Clone should equal original")
		}
	}
}

// Helper function to create test folders
func createTestFolder(name string, depth, width int) *Folder {
	if depth <= 0 {
		return NewFolder()
	}

	folder := NewFolder()

	for i := 0; i < width; i++ {
		if depth > 1 {
			// Add subfolder
			subfolder := createTestFolder(fmt.Sprintf("%s_sub_%d", name, i), depth-1, width)
			folder._entries[fmt.Sprintf("sub_%d", i)] = subfolder
		} else {
			// Add file
			file := NewFile(FileOptions{Content: []byte(fmt.Sprintf("content %s_%d", name, i)), Mode: 0644})
			folder._entries[fmt.Sprintf("file_%d.txt", i)] = file
		}
	}

	return folder
}

// TestPerformanceThresholds runs performance tests and checks against thresholds
func TestPerformanceThresholds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	// Test file operations
	t.Run("FileOperations", func(t *testing.T) {
		// Create a test file
		file := NewFile(FileOptions{Content: []byte("test content"), Mode: 0644})

		start := time.Now()
		_ = file.Content()
		readTime := time.Since(start)

		if readTime > MaxFileReadTime {
			t.Errorf("File read operation exceeded threshold: %v (max: %v)", readTime, MaxFileReadTime)
		}

		start = time.Now()
		clone := file.Clone()
		cloneTime := time.Since(start)

		if cloneTime > MaxFileCreateTime {
			t.Errorf("File clone operation exceeded threshold: %v (max: %v)", cloneTime, MaxFileCreateTime)
		}

		// Test equality
		equal := file.Equal(clone)
		if !equal {
			t.Error("Cloned file should equal original")
		}
	})

	// Test folder operations
	t.Run("FolderOperations", func(t *testing.T) {
		folder := createTestFolder("test", 3, 2)

		start := time.Now()
		clone := folder.Clone()
		cloneTime := time.Since(start)

		if cloneTime > MaxFolderCreateTime {
			t.Errorf("Folder clone operation exceeded threshold: %v (max: %v)", cloneTime, MaxFolderCreateTime)
		}

		start = time.Now()
		_ = folder.Entries()
		readTime := time.Since(start)

		if readTime > MaxFolderReadTime {
			t.Errorf("Folder read operation exceeded threshold: %v (max: %v)", readTime, MaxFolderReadTime)
		}

		// Test equality
		equal := folder.Equal(clone)
		if !equal {
			t.Error("Cloned folder should equal original")
		}
	})

	// Test diff operations
	t.Run("DiffOperations", func(t *testing.T) {
		folder1 := createTestFolder("test1", 2, 2)
		folder2 := createTestFolder("test2", 2, 2)
		folder2.File("extra.txt", FileOptions{Content: []byte("extra content"), Mode: 0644})

		start := time.Now()
		diff := Diff(folder1, folder2, true)
		diffTime := time.Since(start)

		if diffTime > MaxDiffTime {
			t.Errorf("Diff operation exceeded threshold: %v (max: %v)", diffTime, MaxDiffTime)
		}

		// Different folders should have diff operations
		if diff.Operand == op.Noop {
			t.Error("Different folders should have diff operations")
		}
	})
}

// TestMemoryLeaks checks for potential memory leaks
func TestMemoryLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak tests in short mode")
	}

	// Create a large structure
	root := createTestFolder("memory_test", 4, 3)

	// Perform many operations
	for i := 0; i < 100; i++ {
		clone := root.Clone()
		cloneFolder := clone.(*Folder)
		diff := Diff(root, cloneFolder, true)

		// Identical folders should have no diff operations
		if diff.Operand != op.Noop {
			t.Error("Identical folders should have no diff operations")
		}

		if !root.Equal(clone) {
			t.Error("Clone should equal original")
		}
	}

	// The test passes if we don't run out of memory
}

// TestConcurrentSafety tests concurrent access safety
func TestConcurrentSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent safety tests in short mode")
	}

	folder := createTestFolder("concurrent_test", 3, 2)
	// Add a specific file that the test will access
	folder.File("file1.txt", FileOptions{Content: []byte("test content"), Mode: 0644})

	// Test concurrent reads
	t.Run("ConcurrentReads", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			// Capture the loop variable explicitly to avoid race conditions
			workerID := i
			go func() {
				defer func() { done <- true }()

				// Perform read operations
				_ = folder.Entries()
				_ = folder.Get("file1.txt")

				// Log which worker is executing (for debugging)
				t.Logf("Worker %d completed read operations", workerID)
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	// Test concurrent clones
	t.Run("ConcurrentClones", func(t *testing.T) {
		done := make(chan bool, 5)

		for i := 0; i < 5; i++ {
			// Capture the loop variable explicitly to avoid race conditions
			workerID := i
			go func() {
				defer func() { done <- true }()

				clone := folder.Clone()
				if !folder.Equal(clone) {
					t.Error("Clone should equal original")
				}

				// Log which worker is executing (for debugging)
				t.Logf("Worker %d completed clone operations", workerID)
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 5; i++ {
			<-done
		}
	})
}
