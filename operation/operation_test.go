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
			name:      "",
			operation: NewFileOperation("README.go"),
			expected:  `└── CreateFile: README.go`,
		},
		{
			name:      "",
			operation: NewChangeFileOperation("README.go"),
			expected:  `└── ChangeFile: README.go`,
		},
		{
			name:      "",
			operation: NewUnlink("README.go"),
			expected:  `└── Unlink: README.go`,
		},
		{
			name:      "dir with no entries",
			operation: NewMkdirOperation("path/to/dir"),
			expected:  `└── Mkdir: path/to/dir`,
		},

		{
			name:      "dir with one entry",
			operation: NewMkdirOperation("path/to/dir", NewFileOperation("README.md")),
			expected: `
├── Mkdir: path/to/dir
│   └── CreateFile: README.md`,
		},

		{
			name: "complex dir example",
			operation: NewChangeFolderOperation(
				"path/to/dir",
				NewFileOperation("README.md"),
				NewChangeFolderOperation(
					"bar",
					NewFileOperation("banana.go")),
				NewChangeFileOperation("TODO.txt"),
			),
			expected: `
├── ChangeDir: path/to/dir
│   ├── CreateFile: README.md
│   ├── ChangeDir: bar
│   │   └── CreateFile: banana.go
│   └── ChangeFile: TODO.txt
`,
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(strings.TrimSpace(tc.expected), Print(tc.operation))
		})
	}
}
