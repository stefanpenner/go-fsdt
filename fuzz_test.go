package fsdt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"unicode"

	op "github.com/stefanpenner/go-fsdt/operation"
)

// FuzzFolderCreation tests folder creation with various inputs
func FuzzFolderCreation(f *testing.F) {
	// Seed with some basic test cases
	testCases := []string{
		"",
		"normal",
		"with spaces",
		"with/slashes",
		"with\\backslashes",
		"with\ttabs",
		"with\nnewlines",
		"with\r\r\nmixed",
		"with\x00nulls",
		"with\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0A\x0B\x0C\x0D\x0E\x0F",
		strings.Repeat("a", 1000),
		strings.Repeat("🚀", 100),
		"with\xFF\xFE\xFD\xFC\xFB\xFA\xF9\xF8\xF7\xF6\xF5\xF4\xF3\xF2\xF1\xF0",
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, name string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in folder creation with name '%s': %v", name, r)
			}
		}()

		// Test basic folder creation
		folder := NewFolder()
		if folder == nil {
			t.Fatal("NewFolder() returned nil")
		}

		// Test folder with callback
		folderWithCallback := NewFolder(func(f *Folder) {
			f.File("test.txt", FileOptions{Content: []byte("test")})
		})
		if folderWithCallback == nil {
			t.Fatal("NewFolder with callback returned nil")
		}

		// Test folder creation with various names
		if name != "" {
			subFolder := folder.Folder(name)
			if subFolder == nil {
				t.Fatal("Folder creation with name returned nil")
			}
		}

		// Test file creation with various names
		if name != "" {
			file := folder.File(name)
			if file == nil {
				t.Fatal("File creation with name returned nil")
			}
		}

		// Test string file creation
		if name != "" {
			file := folder.FileString(name, "content")
			if file == nil {
				t.Fatal("FileString creation with name returned nil")
			}
		}
	})
}

// FuzzFileOperations tests file operations with various inputs
func FuzzFileOperations(f *testing.F) {
	// Seed with various content types
	testCases := []struct {
		name    string
		content string
		mode    os.FileMode
	}{
		{"", "", 0644},
		{"normal.txt", "normal content", 0644},
		{"empty.txt", "", 0644},
		{"binary.bin", "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09", 0755},
		{"unicode.txt", "🚀🌟✨💫🎉🎊🎋🎍🎎🎏", 0644},
		{"long.txt", strings.Repeat("a", 10000), 0644},
		{"nulls.txt", "\x00\x00\x00\x00", 0644},
		{"control.txt", "\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0A\x0B\x0C\x0D\x0E\x0F", 0644},
	}

	for _, tc := range testCases {
		f.Add(tc.name, tc.content, uint32(tc.mode))
	}

	f.Fuzz(func(t *testing.T, name string, content string, mode uint32) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in file operations with name '%s', content '%s', mode %d: %v", name, content, mode, r)
			}
		}()

		// Test file creation with various options
		file := NewFile(FileOptions{
			Content: []byte(content),
			Mode:    os.FileMode(mode),
		})
		if file == nil {
			t.Fatal("NewFile returned nil")
		}

		// Test string file creation
		stringFile := NewFileString(content)
		if stringFile == nil {
			t.Fatal("NewFileString returned nil")
		}

		// Test file operations
		if name != "" {
			operation := file.CreateOperation(name, op.Reason{})
			if operation.Operand == "" {
				t.Fatal("CreateOperation returned empty operand")
			}

			operation = file.RemoveOperation(name, op.Reason{})
			if operation.Operand == "" {
				t.Fatal("RemoveOperation returned empty operand")
			}

			operation = file.ChangeOperation(name, op.Reason{})
			if operation.Operand == "" {
				t.Fatal("ChangeOperation returned empty operand")
			}
		}

		// Test file equality
		equal := file.Equal(file)
		if !equal {
			t.Fatal("File should equal itself")
		}

		// Test file cloning
		clone := file.Clone()
		if clone == nil {
			t.Fatal("Clone returned nil")
		}

		// Test content methods
		if !bytes.Equal(file.Content(), []byte(content)) {
			t.Fatal("Content mismatch")
		}

		if file.ContentString() != content {
			t.Fatal("ContentString mismatch")
		}

		// Test strings method
		strings := file.Strings(name)
		if len(strings) != 1 {
			t.Fatal("Strings should return exactly one element")
		}
	})
}

// FuzzLinkOperations tests link operations with various inputs
func FuzzLinkOperations(f *testing.F) {
	// Seed with various target types
	testCases := []struct {
		linkName string
		target   string
	}{
		{"", ""},
		{"normal", "target"},
		{"with spaces", "target with spaces"},
		{"with/slashes", "target/with/slashes"},
		{"with\\backslashes", "target\\with\\backslashes"},
		{"with\ttabs", "target\twith\ttabs"},
		{"with\nnewlines", "target\nwith\nnewlines"},
		{"with\x00nulls", "target\x00with\x00nulls"},
		{"with\xFF\xFE\xFD", "target\xFF\xFE\xFD"},
		{strings.Repeat("a", 100), strings.Repeat("b", 100)},
		{"🚀", "🌟"},
	}

	for _, tc := range testCases {
		f.Add(tc.linkName, tc.target)
	}

	f.Fuzz(func(t *testing.T, linkName string, target string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in link operations with linkName '%s', target '%s': %v", linkName, target, r)
			}
		}()

		// Test link creation
		link := NewLink(target, SYMLINK)
		if link == nil {
			t.Fatal("NewLink returned nil")
		}

		// Test link operations
		if linkName != "" {
			operation := link.CreateOperation(linkName, op.Reason{})
			if operation.Operand == "" {
				t.Fatal("CreateOperation returned empty operand")
			}

			operation = link.RemoveOperation(linkName, op.Reason{})
			if operation.Operand == "" {
				t.Fatal("RemoveOperation returned empty operand")
			}
		}

		// Test link equality
		equal := link.Equal(link)
		if !equal {
			t.Fatal("Link should equal itself")
		}

		// Test link cloning
		clone := link.Clone()
		if clone == nil {
			t.Fatal("Clone returned nil")
		}

		// Test content methods
		if link.ContentString() != target {
			t.Fatal("ContentString mismatch")
		}

		if !bytes.Equal(link.Content(), []byte(target)) {
			t.Fatal("Content mismatch")
		}

		// Test strings method
		strings := link.Strings(linkName)
		if len(strings) != 1 {
			t.Fatal("Strings should return exactly one element")
		}

		// Test target method
		if link.Target() != target {
			t.Fatal("Target mismatch")
		}
	})
}

// FuzzFolderOperations tests complex folder operations
func FuzzFolderOperations(f *testing.F) {
	// Seed with various folder structures
	testCases := []struct {
		depth    int
		width    int
		fileName string
		content  string
	}{
		{1, 1, "test.txt", "content"},
		{2, 2, "test.txt", "content"},
		{3, 3, "test.txt", "content"},
		{5, 5, "test.txt", "content"},
		{10, 2, "test.txt", "content"},
		{2, 10, "test.txt", "content"},
	}

	for _, tc := range testCases {
		f.Add(tc.depth, tc.width, tc.fileName, tc.content)
	}

	f.Fuzz(func(t *testing.T, depth, width int, fileName, content string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in folder operations with depth %d, width %d, fileName '%s', content '%s': %v", depth, width, fileName, content, r)
			}
		}()

		// Ensure we always have a valid folder structure with entries
		if depth <= 0 {
			depth = 1
		}
		if width <= 0 {
			width = 1
		}
		if depth > 10 {
			depth = 10
		}
		if width > 10 {
			width = 10
		}

		// Create a complex folder structure
		root := NewFolder()
		createComplexStructure(t, root, depth, width, fileName, content)

		// Verify the folder has entries before proceeding
		entries := root.Entries()
		if len(entries) == 0 {
			t.Fatal("Folder should have entries after creation")
		}

		// Test various operations
		testFolderOperations(t, root, fileName, content)

		// Test cloning
		clone := root.Clone()
		if clone == nil {
			t.Fatal("Clone returned nil")
		}

		// Test equality
		equal := root.Equal(clone)
		if !equal {
			t.Fatal("Clone should equal original")
		}

		// Test file strings
		fileStrings := root.FileStrings("")
		if len(fileStrings) == 0 {
			t.Fatal("FileStrings should return some files")
		}

		// Test strings
		strings := root.Strings("")
		if len(strings) == 0 {
			t.Fatal("Strings should return some entries")
		}
	})
}

// FuzzDiffOperations tests diff operations with various inputs
func FuzzDiffOperations(f *testing.F) {
	// Seed with various diff scenarios
	testCases := []struct {
		depth    int
		width    int
		fileName string
		content  string
		modify   bool
	}{
		{1, 1, "test.txt", "content", false},
		{2, 2, "test.txt", "content", true},
		{3, 3, "test.txt", "content", false},
		{5, 5, "test.txt", "content", true},
	}

	for _, tc := range testCases {
		f.Add(tc.depth, tc.width, tc.fileName, tc.content, tc.modify)
	}

	f.Fuzz(func(t *testing.T, depth, width int, fileName, content string, modify bool) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in diff operations with depth %d, width %d, fileName '%s', content '%s', modify %t: %v", depth, width, fileName, content, modify, r)
			}
		}()

		// Limit depth and width to prevent excessive memory usage
		if depth < 0 || depth > 5 {
			depth = 1
		}
		if width < 0 || width > 5 {
			width = 1
		}

		// Create two similar folder structures
		folderA := NewFolder()
		folderB := NewFolder()

		createComplexStructure(t, folderA, depth, width, fileName, content)
		createComplexStructure(t, folderB, depth, width, fileName, content)

		// Modify one folder if requested
		if modify {
			modifyFolder(t, folderB, fileName, content+"modified")
		}

		// Test diff operations
		diff := folderA.Diff(folderB)
		if diff.Operand == "" {
			t.Fatal("Diff should return valid operation")
		}

		caseInsensitiveDiff := folderA.CaseInsensitiveDiff(folderB)
		if caseInsensitiveDiff.Operand == "" {
			t.Fatal("CaseInsensitiveDiff should return valid operation")
		}
	})
}

// FuzzEdgeCases tests various edge cases and error conditions
func FuzzEdgeCases(f *testing.F) {
	// Seed with various edge cases
	testCases := []struct {
		operation string
		input     string
	}{
		{"empty", ""},
		{"nulls", "\x00\x00\x00"},
		{"control", "\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0A\x0B\x0C\x0D\x0E\x0F"},
		{"unicode", "🚀🌟✨💫🎉🎊🎋🎍🎎🎏"},
		{"long", strings.Repeat("a", 1000)},
		{"special", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"whitespace", " \t\n\r\f\v"},
		{"path", "/path/to/something/with/many/levels"},
		{"backslash", "\\path\\to\\something\\with\\many\\levels"},
	}

	for _, tc := range testCases {
		f.Add(tc.operation, tc.input)
	}

	f.Fuzz(func(t *testing.T, operation, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in edge case testing with operation '%s', input '%s': %v", operation, input, r)
			}
		}()

		// Test various edge cases
		switch operation {
		case "empty":
			testEmptyInputs(t, input)
		case "nulls":
			testNullInputs(t, input)
		case "control":
			testControlInputs(t, input)
		case "unicode":
			testUnicodeInputs(t, input)
		case "long":
			testLongInputs(t, input)
		case "special":
			testSpecialInputs(t, input)
		case "whitespace":
			testWhitespaceInputs(t, input)
		case "path":
			testPathInputs(t, input)
		case "backslash":
			testBackslashInputs(t, input)
		default:
			testGenericInputs(t, input)
		}
	})
}

// Helper functions for creating complex folder structures
func createComplexStructure(t *testing.T, folder *Folder, depth, width int, fileName, content string) {
	if depth <= 0 {
		return
	}

	for i := 0; i < width; i++ {
		name := fmt.Sprintf("level_%d_item_%d", depth, i)

		if depth == 1 {
			// Create files at the leaf level
			folder.File(name+"_"+fileName, FileOptions{Content: []byte(content)})
			folder.FileString(name+"_string_"+fileName, content)
		} else {
			// Create subfolders
			subFolder := folder.Folder(name)
			createComplexStructure(t, subFolder, depth-1, width, fileName, content)
		}

		// Add some symlinks
		if i%2 == 0 {
			folder.Symlink(name+"_link", name+"_target")
		}
	}
}

func modifyFolder(t *testing.T, folder *Folder, fileName, content string) {
	// Modify some files
	entries := folder.Entries()
	if len(entries) > 0 {
		firstEntry := entries[0]
		if _, ok := folder.Get(firstEntry).(*File); ok {
			// This would require a way to modify file content, which isn't currently supported
			// So we'll just create a new file with the same name
			folder.File(firstEntry+"_"+fileName, FileOptions{Content: []byte(content)})
		}
	}
}

// testFolderOperations tests various folder operations
func testFolderOperations(t *testing.T, folder *Folder, fileName, content string) {
	// Test basic operations
	entries := folder.Entries()
	if len(entries) == 0 {
		t.Fatal("Folder should have entries")
	}

	// Test getting entries
	for _, name := range entries {
		entry := folder.Get(name)
		if entry == nil {
			t.Fatalf("Failed to get entry %s", name)
		}
	}

	// Test operations without modifying the folder structure
	// This ensures equality tests work correctly
	if len(entries) > 0 {
		firstEntry := entries[0]
		// Just verify we can access the entry, don't remove it
		entry := folder.Get(firstEntry)
		if entry == nil {
			t.Fatalf("Failed to get entry %s for verification", firstEntry)
		}

		// Test that we can get the entry type
		entryType := entry.Type()
		if entryType == "" {
			t.Fatalf("Entry %s has empty type", firstEntry)
		}
	}
}

// Helper functions for testing various input types
func testEmptyInputs(t *testing.T, input string) {
	folder := NewFolder()

	// Test with empty strings
	if input == "" {
		file := folder.File("")
		if file == nil {
			t.Fatal("File creation with empty name should not return nil")
		}

		folder.FileString("", "")
		folder.Folder("")
	}
}

func testNullInputs(t *testing.T, input string) {
	folder := NewFolder()

	// Test with null bytes
	if strings.Contains(input, "\x00") {
		// This might cause issues, so we'll test it carefully
		defer func() {
			if r := recover(); r != nil {
				// Expected panic for null bytes
			}
		}()

		folder.File(input, FileOptions{Content: []byte(input)})
	}
}

func testControlInputs(t *testing.T, input string) {
	folder := NewFolder()

	// Test with control characters
	if strings.ContainsAny(input, "\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0A\x0B\x0C\x0D\x0E\x0F") {
		folder.File(input, FileOptions{Content: []byte(input)})
	}
}

func testUnicodeInputs(t *testing.T, input string) {
	folder := NewFolder()

	// Test with unicode characters
	if len(input) > 0 && unicode.IsLetter(rune(input[0])) {
		folder.File(input, FileOptions{Content: []byte(input)})
		folder.FileString(input, input)
	}
}

func testLongInputs(t *testing.T, input string) {
	folder := NewFolder()

	// Test with long inputs
	if len(input) > 100 {
		folder.File(input, FileOptions{Content: []byte(input)})
	}
}

func testSpecialInputs(t *testing.T, input string) {
	folder := NewFolder()

	// Test with special characters
	if strings.ContainsAny(input, "!@#$%^&*()_+-=[]{}|;':\",./<>?") {
		folder.File(input, FileOptions{Content: []byte(input)})
	}
}

func testWhitespaceInputs(t *testing.T, input string) {
	folder := NewFolder()

	// Test with whitespace
	if strings.ContainsAny(input, " \t\n\r\f\v") {
		folder.File(input, FileOptions{Content: []byte(input)})
	}
}

func testPathInputs(t *testing.T, input string) {
	folder := NewFolder()

	// Test with path-like inputs
	if strings.Contains(input, "/") {
		folder.File(input, FileOptions{Content: []byte(input)})
	}
}

func testBackslashInputs(t *testing.T, input string) {
	folder := NewFolder()

	// Test with backslash inputs
	if strings.Contains(input, "\\") {
		folder.File(input, FileOptions{Content: []byte(input)})
	}
}

func testGenericInputs(t *testing.T, input string) {
	folder := NewFolder()

	// Generic test for any input
	folder.File(input, FileOptions{Content: []byte(input)})
}

// FuzzSerialization tests serialization and deserialization edge cases
func FuzzSerialization(f *testing.F) {
	// Seed with various serialization scenarios
	testCases := []struct {
		operation string
		data      string
	}{
		{"json", `{"key": "value"}`},
		{"json_escape", `{"key": "value with \"quotes\""}`},
		{"json_unicode", `{"key": "🚀🌟✨"}`},
		{"json_null", `{"key": null}`},
		{"json_empty", `{}`},
		{"json_nested", `{"key": {"nested": "value"}}`},
		{"json_array", `{"key": ["a", "b", "c"]}`},
	}

	for _, tc := range testCases {
		f.Add(tc.operation, tc.data)
	}

	f.Fuzz(func(t *testing.T, operation, data string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in serialization testing with operation '%s', data '%s': %v", operation, data, r)
			}
		}()

		// Test JSON parsing (if the data looks like JSON)
		if strings.HasPrefix(data, "{") && strings.HasSuffix(data, "}") {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(data), &result); err == nil {
				// Valid JSON, test with our structures
				folder := NewFolder()
				folder.File("test.json", FileOptions{Content: []byte(data)})

				// Test that we can read it back
				content := folder.Get("test.json").Content()
				if !bytes.Equal(content, []byte(data)) {
					t.Fatal("Content mismatch after JSON round-trip")
				}
			}
		}
	})
}

// FuzzMemoryStress tests memory allocation and stress scenarios
func FuzzMemoryStress(f *testing.F) {
	// Seed with various stress scenarios
	testCases := []struct {
		iterations int
		size       int
		operation  string
	}{
		{10, 100, "create"},
		{100, 10, "create"},
		{1000, 1, "create"},
		{1, 10000, "create"},
	}

	for _, tc := range testCases {
		f.Add(tc.iterations, tc.size, tc.operation)
	}

	f.Fuzz(func(t *testing.T, iterations, size int, operation string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in memory stress testing with iterations %d, size %d, operation '%s': %v", iterations, size, operation, r)
			}
		}()

		// Limit to prevent excessive memory usage
		if iterations < 0 || iterations > 1000 {
			iterations = 100
		}
		if size < 0 || size > 10000 {
			size = 100
		}

		switch operation {
		case "create":
			testMemoryStressCreate(t, iterations, size)
		default:
			testMemoryStressCreate(t, iterations, size)
		}
	})
}

func testMemoryStressCreate(t *testing.T, iterations, size int) {
	folder := NewFolder()

	for i := 0; i < iterations; i++ {
		name := fmt.Sprintf("stress_%d", i)
		content := strings.Repeat(fmt.Sprintf("%d", i%10), size)

		folder.File(name, FileOptions{Content: []byte(content)})

		// Verify the file was created
		if entry := folder.Get(name); entry == nil {
			t.Fatalf("Failed to create file %s", name)
		}
	}

	// Verify all files exist
	entries := folder.Entries()
	if len(entries) != iterations {
		t.Fatalf("Expected %d entries, got %d", iterations, len(entries))
	}
}
