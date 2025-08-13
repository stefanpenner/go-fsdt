package operation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintOperation(t *testing.T) {
	scenarios := []struct {
		name      string
		operation Operation
		expected  string
	}{
		{
			name:      "create file",
			operation: NewCreateFileOperation("README.go", []byte("content"), 0644),
			expected:  `CreateFile: README.go (file (7 bytes, mode: 644)) [New: File does not exist in target]`,
		},
		{
			name: "change file",
			operation: NewChangeFileOperation("README.go",
				FileValue{Content: []byte("old"), Mode: 0644, Size: 3},
				FileValue{Content: []byte("new"), Mode: 0644, Size: 3},
				ReasonContentChanged),
			expected: `ChangeFile: README.go (changed from file (3 bytes, mode: 644) to file (3 bytes, mode: 644)) [ContentChanged: File ContentChanged changed (before: file (3 bytes, mode: 644), after: file (3 bytes, mode: 644))]`,
		},
		{
			name:      "remove file",
			operation: NewRemoveFileOperation("README.go"),
			expected:  `RemoveFile: README.go [Missing: File exists in source but not in target]`,
		},
		{
			name:      "dir with no entries",
			operation: NewCreateDirOperation("path/to/dir"),
			expected:  `CreateDir: path/to/dir (empty directory) [New: Directory does not exist in target]`,
		},

		{
			name:      "dir with one entry",
			operation: NewCreateDirOperation("path/to/dir", NewCreateFileOperation("README.md", []byte("content"), 0644)),
			expected: `CreateDir: path/to/dir (directory with 1 operations) [New: Directory does not exist in target]
└── CreateFile: README.md (file (7 bytes, mode: 644)) [New: File does not exist in target]`,
		},

		{
			name: "complex dir example",
			operation: NewChangeDirOperation(
				"path/to/dir",
				NewCreateFileOperation("README.md", []byte("content"), 0644),
				NewChangeDirOperation(
					"bar",
					NewCreateFileOperation("banana.go", []byte("content"), 0644)),
				NewChangeFileOperation("TODO.txt",
					FileValue{Content: []byte("old"), Mode: 0644, Size: 3},
					FileValue{Content: []byte("new"), Mode: 0644, Size: 3},
					ReasonContentChanged),
			),
			expected: `ChangeDir: path/to/dir (directory with 3 operations) [ContentChanged: Directory contents have changed]
├── CreateFile: README.md (file (7 bytes, mode: 644)) [New: File does not exist in target]
├── ChangeDir: bar (directory with 1 operations) [ContentChanged: Directory contents have changed]
│  └── CreateFile: banana.go (file (7 bytes, mode: 644)) [New: File does not exist in target]
└── ChangeFile: TODO.txt (changed from file (3 bytes, mode: 644) to file (3 bytes, mode: 644)) [ContentChanged: File ContentChanged changed (before: file (3 bytes, mode: 644), after: file (3 bytes, mode: 644))]`,
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			formatter := NewTreeFormatter()
			formatter.ShowReasons = true
			formatter.ShowValues = true
			result := formatter.Format(tc.operation)
			assert.Equal(strings.TrimSpace(tc.expected), strings.TrimSpace(result))
		})
	}
}
